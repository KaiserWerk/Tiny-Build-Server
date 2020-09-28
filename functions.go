package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KaiserWerk/sessionstore"
	"golang.org/x/crypto/ssh"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	basePath = "data/"
)

func getFlashbag(mgr *sessionstore.SessionManager) func() template.HTML {
	return func() template.HTML {
		if mgr == nil {
			writeToConsole("mgr is nil")
			return template.HTML("")
		}
		var sb strings.Builder
		var source string
		const msgSuccess = `<div class="alert alert-success alert-dismissable"><a href="#" class="close" data-dismiss="alert" aria-label="close">&times;</a><strong>Success!</strong> %%message%%</div>`
		const msgError = `<div class="alert alert-danger alert-dismissable"><a href="#" class="close" data-dismiss="alert" aria-label="close">&times;</a><strong>Error!</strong> %%message%%</div>`
		const msgWarning = `<div class="alert alert-warning alert-dismissable"><a href="#" class="close" data-dismiss="alert" aria-label="close">&times;</a><strong>Warning!</strong> %%message%%</div>`
		const msgInfo = `<div class="alert alert-info alert-dismissable"><a href="#" class="close" data-dismiss="alert" aria-label="close">&times;</a><strong>Info!</strong> %%message%%</div>`

		for _, v := range mgr.GetMessages() {
			if v.MessageType == "success" {
				source = msgSuccess
			} else if v.MessageType == "error" {
				source = msgError
			} else if v.MessageType == "warning" {
				source = msgWarning
			} else if v.MessageType == "info" {
				source = msgInfo
			}

			sb.WriteString(strings.Replace(source, "%%message%%", v.Content, 1))
		}

		return template.HTML(sb.String())
	}
}

func getHeaderIfSet(r *http.Request, key string) (string, error) {
	header := r.Header.Get(key)
	if header == "" {
		return "", errors.New("header is not set or empty")
	}
	return header, nil
}

func checkPayloadRequest(r *http.Request) (*buildDefinition, error) {
	// get id
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		return nil, errors.New("could not determine ID of build definition")
	}
	// convert to integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, errors.New("invalid ID value supplied")
	}
	// get DB connection
	db, err := getDbConnection()
	if err != nil {
		return nil, errors.New("could not get DB connection")
	}
	// fetch the build definition
	var bd buildDefinition
	row := db.QueryRow("SELECT * FROM build_definition WHERE id = ?", id)
	err = row.Scan(&bd.Id, &bd.BuildTargetId, &bd.AlteredBy, &bd.Caption, &bd.Enabled, &bd.DeploymentEnabled,
		&bd.RepoHoster, &bd.RepoHosterUrl, &bd.RepoFullname, &bd.RepoUsername, &bd.RepoSecret, &bd.RepoBranch,
		&bd.AlteredAt, &bd.ApplyMigrations, &bd.DatabaseDSN, &bd.MetaMigrationId, &bd.RunTests,
		&bd.RunBenchmarkTests)
	if err != nil {
		return nil, errors.New("could not scan buildDefinition")
	}

	// check relevant headers and payload values
	switch bd.RepoHoster {
	case "bitbucket":
		headers := []string{"X-Event-Key", "X-Hook-UUID", "X-Request-UUID", "X-Attempt-Number"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = getHeaderIfSet(r, headers[i])
			if err != nil {
				return nil, errors.New("could not get bitbucket header " + headers[i])
			}
		}

		var payload bitBucketPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return nil, errors.New("could not decode json payload")
		}
		if payload.Push.Changes[0].New.Name != bd.RepoBranch {
			return nil, errors.New("branch names do not match (" + payload.Push.Changes[0].New.Name + ")")
		}
		if payload.Repository.FullName != bd.RepoFullname {
			return nil, errors.New("repository names do not match (" + payload.Repository.FullName + ")")
		}
	case "github":
		headers := []string{"X-GitHub-Delivery", "X-GitHub-Event", "X-Hub-Signature"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = getHeaderIfSet(r, headers[i])
			if err != nil {
				return nil, errors.New("could not get github header " + headers[i])
			}
		}

		var payload gitHubPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return nil, errors.New("could not decode json payload")
		}
		if payload.Repository.DefaultBranch != bd.RepoBranch {
			return nil, errors.New("branch names do not match (" + payload.Repository.DefaultBranch + ")")
		}
		if payload.Repository.FullName != bd.RepoFullname {
			return nil, errors.New("repository names do not match (" + payload.Repository.FullName + ")")
		}
	case "gitlab":
		headers := []string{"X-GitLab-Event"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = getHeaderIfSet(r, headers[i])
			if err != nil {
				return nil, errors.New("could not get gitlab header " + headers[i])
			}
		}

		var payload gitLabPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return nil, errors.New("could not decode json payload")
		}
		branch := strings.Split(payload.Ref, "/")[2]
		if branch != bd.RepoBranch {
			return nil, errors.New("branch names do not match (" + branch + ")")
		}
		if payload.Project.PathWithNamespace != bd.RepoFullname {
			return nil, errors.New("repository names do not match (" + payload.Project.PathWithNamespace + ")")
		}
	case "gitea":
		headers := []string{"X-Gitea-Delivery", "X-Gitea-Event"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = getHeaderIfSet(r, headers[i])
			if err != nil {
				return nil, errors.New("could not get gitea header " + headers[i])
			}
		}

		var payload giteaPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return nil, errors.New("could not decode json payload")
		}
		branch := strings.Split(payload.Ref, "/")[2]
		if branch != bd.RepoBranch {
			return nil, errors.New("branch names do not match (" + branch + ")")
		}
		if payload.Repository.FullName != bd.RepoFullname {
			return nil, errors.New("repository names do not match (" + payload.Repository.FullName + ")")
		}
	}

	return &bd, nil
}

//func loadSysConfig() (sysConfig, error) {
//	cont, err := ioutil.ReadFile("config/app.yaml")
//	if err != nil {
//		return sysConfig{}, errors.New("could not read config/app.yaml file")
//	}
//	var config sysConfig
//	err = yaml.Unmarshal(cont, &config)
//	if err != nil {
//		return sysConfig{}, errors.New("could not parse config/app.yaml file")
//	}
//
//	return config, nil
//}

func readConsoleInput(externalShutdownCh chan os.Signal) {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _, err := reader.ReadLine()
		if err != nil {
			fmt.Printf("  could not process input %v\n", input)
			continue
		}

		switch string(input) {
		case "":
			continue
		case "cluck":
			animal := `   \\
   (o>
\\_//)
 \_/_)
  _|_  
You found the chicken. Hooray!`
			fmt.Println(animal)
		case "shutdown":
			writeToConsole("shutdown via console initiated...")
			time.Sleep(time.Second)
			externalShutdownCh <- os.Interrupt
		case "reload config":
			writeToConsole("reloading configuration...")
			time.Sleep(time.Second)
			// @TODO

			writeToConsole("done")
		case "invalidate sessions":
			writeToConsole("invalidating all sessions...")
			sessMgr.RemoveAllSessions()
			time.Sleep(time.Second)
			writeToConsole("done")
		case "list sessions":
			writeToConsole("all sessions:")
			for _, v := range sessMgr.Sessions {
				writeToConsole("Id: " + v.Id + "\tLifetime:" + v.Lifetime.Format("2006-01-02 15:04:05"))
			}
		default:
			writeToConsole("unrecognized command: " + string(input))
		}
	}
}

//func loadBuildDefinition(id string) (buildDefinition, error) {
//	bdDir := "build_definitions/build_" + id
//	bdFile := bdDir + "/build.yaml"
//	fmt.Println("full build path:", bdFile)
//	if _, err := os.Stat(bdDir); os.IsNotExist(err) {
//		fmt.Printf("build definition with Id %v not found\n", id)
//		return buildDefinition{}, buildDefinitionNotFound{Id: id}
//	}
//
//	if _, err := os.Stat(bdFile); os.IsNotExist(err) {
//		fmt.Printf("config file for build definition with Id %v not found\n", id)
//		return buildDefinition{}, buildDefinitionConfigFileNotFound{Id: id}
//	}
//
//	cont, err := ioutil.ReadFile(bdFile)
//	if err != nil {
//		fmt.Println("could not read build definition config file")
//		return buildDefinition{}, errors.New("could not read build definition config file")
//	}
//	var bd buildDefinition
//	err = yaml.Unmarshal(cont, &bd)
//	if err != nil {
//		fmt.Println("could not unmarshal yaml")
//		return buildDefinition{}, errors.New("could not unmarshal yaml")
//	}
//
//	return bd, nil
//}



func startBuildProcess(definition buildDefinition) {

	/*
		* clone
		* restore
		* test
		* test bench
		* build arch

		arch = window_amd64, darwin_amd32, raspi3, ...
	*/

	// tools for build output instantiated hehe
	var sb strings.Builder
	messageCh := make(chan string, 100)
	// definition ID as string determined
	idString := strconv.Itoa(definition.Id)
	// basepath festlegen
	projectPath := basePath + idString + "/" + strconv.FormatInt(time.Now().Unix(), 10)
	if fileExists(projectPath) {
		err := os.RemoveAll(projectPath)
		if err != nil {
			sb.WriteString("could not remove stale build directory ("+ projectPath +"): " + err.Error())
		}
	}
	// create a new build directory
	err := os.MkdirAll(projectPath + "/", 0664)
	if err != nil {
		sb.WriteString("could not create build directory ("+ projectPath +"): " + err.Error())
	}


	//baseDir := "build_definitions/build_" + id
	//cloneDir := baseDir + "/clone"
	////buildDir := baseDir + "/build"
	//// remove the clone directory possibly remaining
	//// from previous build processes
	//os.RemoveAll(cloneDir)
	//
	//// clone the repository
	//repositoryUrl := getRepositoryUrl(definition, true)
	//cmd := exec.Command("git", "clone", repositoryUrl, cloneDir)
	//err := cmd.Run()
	//if err != nil {
	//	fmt.Println("could not clone repository; aborting: " + err.Error())
	//	return
	//}
	//
	//sysConf, err := loadSysConfig()
	//if err != nil {
	//	fmt.Println("could not load system config")
	//	return
	//}
	//// change dir
	//err = os.Chdir(cloneDir)
	//if err != nil {
	//	fmt.Println("could not change dir to clone: " + err.Error())
	//	return
	//}
	//
	//switch definition.ProjectType {
	//case "go":
	//case "golang":
	//	outputFile := ""
	//	for _, v := range definition.Actions {
	//		switch true {
	//		case strings.Contains(v, "restore"):
	//			// restore dependencies
	//			err = exec.Command(sysConf.GolangExecutable, "get", "-u").Run()
	//			if err != nil {
	//				fmt.Println("could not restore dependencies: " + err.Error())
	//				return
	//			}
	//		case strings.Contains(v, "test"):
	//			// tests and bench tests don't really matter for now
	//			err = exec.Command(sysConf.GolangExecutable, "test").Run()
	//			if err != nil {
	//				fmt.Println("could not restore dependencies: " + err.Error())
	//				return
	//			}
	//		case strings.Contains(v, "test bench"):
	//			err = exec.Command(sysConf.GolangExecutable, "test", "-bench=.").Run()
	//			if err != nil {
	//				fmt.Println("could not restore dependencies: " + err.Error())
	//				return
	//			}
	//		case strings.Contains(v, "build"):
	//			var (
	//				targetOS   string
	//				targetArch string
	//				targetArm  string
	//			)
	//
	//			osArch := strings.Split(v, " ")[1]
	//
	//			// its sth like raspi
	//			if !strings.Contains(osArch, "_") {
	//				switch osArch {
	//				case "raspi3":
	//					targetOS = "linux"
	//					targetArch = "arm"
	//					targetArm = "5"
	//				case "raspi4":
	//					targetOS = "linux"
	//					targetArch = "arm"
	//					targetArm = "6"
	//				}
	//			} else {
	//				parts := strings.Split(osArch, "_")
	//				targetOS = parts[0]
	//				targetArch = parts[1]
	//			}
	//
	//			_ = os.Setenv("GOOS", targetOS)
	//			_ = os.Setenv("GOARCH", targetArch)
	//			_ = os.Setenv("GOARM", targetArm)
	//			outputFile = fmt.Sprintf("../build/%s%s", repoName, fileExt[targetOS])
	//			err = exec.Command(sysConf.GolangExecutable, "build", "-o", outputFile).Run()
	//			if err != nil {
	//				fmt.Println("could not build: " + err.Error())
	//				return
	//			}
	//		}
	//	}
	//
	//	if definition.DeploymentEnabled {
	//		deployToHost(outputFile, definition)
	//	}
	//	// @TODO reset cwd?
	//
	//case "cs":
	//case "csharp":
	//	// dotnet publish MyProject\Presentation\Presentation.csproj -o C:\MyProject -p:PublishSingleFile=true -p:PublishTrimmed=true -r win-x64
	//	// sysconfig!
	//	// @TODO
	//}

	fmt.Println("build completed!")
}

func deployToHost(outputFile string, definition buildDefinition) {
	for _, v := range definition.Deployments {
		// first, the pre deployment actions
		sshConfig := &ssh.ClientConfig{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			User:            v.Username,
			Auth: []ssh.AuthMethod{
				ssh.Password(v.Password),
			},
		}
		sshClient, err := ssh.Dial("tcp", v.Host, sshConfig)
		if err != nil {
			fmt.Println("could not establish ssh connection:", err.Error())
			return
		}

		if len(v.PreDeploymentActions) > 0 {
			for _, action := range v.PreDeploymentActions {
				session, err := sshClient.NewSession()
				if err != nil {
					fmt.Printf("Failed to create session: %s\n", err.Error())
					return
				}
				outp, err := session.Output(action)
				if err != nil {
					fmt.Println("could not execute pre deployment command: " + action)
					//return
				} else {
					outpDisplay := string(outp)
					if outpDisplay != "" {
						fmt.Println("output from pre remote command:", outpDisplay)
					}
				}
				_ = session.Close()
			}
		}

		session, err := sshClient.NewSession()
		if err != nil {
			fmt.Printf("Failed to create session: %s\n", err.Error())
			return
		}

		// then, the actual deployment
		sftpClient, err := sftp.NewClient(sshClient)
		if err != nil {
			fmt.Println("could not create sftp client instance:", err.Error())
			return
		}

		// create destination file
		// @TODO really necessary?
		dstFile, err := sftpClient.Create(v.WorkingDirectory)
		if err != nil {
			fmt.Println("failed to create remote file:", err.Error())
			return
		}

		// create source file
		srcFile, err := os.Open(outputFile)
		if err != nil {
			fmt.Println("failed to open source file:", err.Error())
			return
		}

		// copy source file to destination file
		bytes, err := io.Copy(dstFile, srcFile)
		if err != nil {
			fmt.Println("failed to copy file:", err.Error())
			return
		}
		_ = dstFile.Close()
		_ = srcFile.Close()
		fmt.Printf("%d bytes copied\n", bytes)
		_ = session.Close()

		// then, the post deployment actions
		if len(v.PostDeploymentActions) > 0 {
			for _, action := range v.PostDeploymentActions {
				session, err := sshClient.NewSession()
				if err != nil {
					fmt.Printf("Failed to create session: %s\n", err.Error())
					return
				}
				outp, err := session.Output(action)
				if err != nil {
					fmt.Println("could not execute post deployment command:", action, "cause:", err.Error())
					//return
				} else {
					outpDisplay := string(outp)
					if outpDisplay != "" {
						fmt.Println("output from post remote command:", outpDisplay)
					}
				}
				_ = session.Close()
			}
		}

		if sshClient != nil {
			err = sshClient.Close()
			if err != nil {
				fmt.Println("could not close ssh client connection")
				return
			}
		}

		_ = sftpClient.Close()
	}
}

func getRepositoryUrl(d buildDefinition, withCredentials bool) string {
	var url string

	switch d.RepoHoster {
	case "bitbucket":
		url = "bitbucket.org/" + d.RepoFullname
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", d.RepoUsername, d.RepoSecret, url)
		}
		return "https://" + url
	case "github":
		url = "github.com/" + d.RepoFullname
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", d.RepoUsername, d.RepoSecret, url)
		}
		return "https://" + url
	case "gitlab":
		url = "gitlab.com/" + d.RepoFullname
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", d.RepoUsername, d.RepoSecret, url)
		}
		return "https://" + url
	case "gitea":
		url = d.RepoHosterUrl + "/" + d.RepoFullname
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", d.RepoUsername, d.RepoSecret, url)
		}
		return "https://" + url

	default:
		url = d.RepoHosterUrl + "/" + d.RepoFullname
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", d.RepoUsername, d.RepoSecret, url)
		}
		return "http://" + url // just http
	}
}
