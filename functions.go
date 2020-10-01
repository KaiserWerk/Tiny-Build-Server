package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KaiserWerk/sessionstore"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"html/template"
	"io"
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

func checkPayloadRequest(r *http.Request) (buildDefinition, error) {
	// get id
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		return buildDefinition{}, errors.New("could not determine ID of build definition")
	}
	// convert to integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return buildDefinition{}, errors.New("invalid ID value supplied")
	}
	// get DB connection
	db, err := getDbConnection()
	if err != nil {
		return buildDefinition{}, errors.New("could not get DB connection")
	}
	// fetch the build definition
	var bd buildDefinition
	row := db.QueryRow("SELECT id, build_target, build_target_os_arch, build_target_arm, altered_by, caption, "+
		"enabled, deployment_enabled, repo_hoster, repo_hoster_url, repo_fullname, repo_username, repo_secret, "+
		"repo_branch, altered_at, apply_migrations, database_dns, meta_migration_id, run_tests, run_benchmark_tests "+
		"FROM build_definition WHERE id = ?", id)
	err = row.Scan(&bd.Id, &bd.BuildTargetId, &bd.BuildTargetOsArch, &bd.BuildTargetArm, &bd.AlteredBy, &bd.Caption,
		&bd.Enabled, &bd.DeploymentEnabled,
		&bd.RepoHoster, &bd.RepoHosterUrl, &bd.RepoFullname, &bd.RepoUsername, &bd.RepoSecret, &bd.RepoBranch,
		&bd.AlteredAt, &bd.ApplyMigrations, &bd.DatabaseDSN, &bd.MetaMigrationId, &bd.RunTests,
		&bd.RunBenchmarkTests)
	if err != nil {
		return buildDefinition{}, errors.New("could not scan buildDefinition")
	}

	// check relevant headers and payload values
	switch bd.RepoHoster {
	case "bitbucket":
		headers := []string{"X-Event-Key", "X-Hook-UUID", "X-Request-UUID", "X-Attempt-Number"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = getHeaderIfSet(r, headers[i])
			if err != nil {
				return buildDefinition{}, errors.New("could not get bitbucket header " + headers[i])
			}
		}

		var payload bitBucketPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return buildDefinition{}, errors.New("could not decode json payload")
		}
		if payload.Push.Changes[0].New.Name != bd.RepoBranch {
			return buildDefinition{}, errors.New("branch names do not match (" + payload.Push.Changes[0].New.Name + ")")
		}
		if payload.Repository.FullName != bd.RepoFullname {
			return buildDefinition{}, errors.New("repository names do not match (" + payload.Repository.FullName + ")")
		}
	case "github":
		headers := []string{"X-GitHub-Delivery", "X-GitHub-Event", "X-Hub-Signature"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = getHeaderIfSet(r, headers[i])
			if err != nil {
				return buildDefinition{}, errors.New("could not get github header " + headers[i])
			}
		}

		var payload gitHubPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return buildDefinition{}, errors.New("could not decode json payload")
		}
		if payload.Repository.DefaultBranch != bd.RepoBranch {
			return buildDefinition{}, errors.New("branch names do not match (" + payload.Repository.DefaultBranch + ")")
		}
		if payload.Repository.FullName != bd.RepoFullname {
			return buildDefinition{}, errors.New("repository names do not match (" + payload.Repository.FullName + ")")
		}
	case "gitlab":
		headers := []string{"X-GitLab-Event"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = getHeaderIfSet(r, headers[i])
			if err != nil {
				return buildDefinition{}, errors.New("could not get gitlab header " + headers[i])
			}
		}

		var payload gitLabPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return buildDefinition{}, errors.New("could not decode json payload")
		}
		branch := strings.Split(payload.Ref, "/")[2]
		if branch != bd.RepoBranch {
			return buildDefinition{}, errors.New("branch names do not match (" + branch + ")")
		}
		if payload.Project.PathWithNamespace != bd.RepoFullname {
			return buildDefinition{}, errors.New("repository names do not match (" + payload.Project.PathWithNamespace + ")")
		}
	case "gitea":
		headers := []string{"X-Gitea-Delivery", "X-Gitea-Event"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = getHeaderIfSet(r, headers[i])
			if err != nil {
				return buildDefinition{}, errors.New("could not get gitea header " + headers[i])
			}
		}

		var payload giteaPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return buildDefinition{}, errors.New("could not decode json payload")
		}
		branch := strings.Split(payload.Ref, "/")[2]
		if branch != bd.RepoBranch {
			return buildDefinition{}, errors.New("branch names do not match (" + branch + ")")
		}
		if payload.Repository.FullName != bd.RepoFullname {
			return buildDefinition{}, errors.New("repository names do not match (" + payload.Repository.FullName + ")")
		}
	}

	return bd, nil
}

func applyMigrations(definition buildDefinition, messageCh chan string, projectDir string) error {

	return nil
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
	// instantiate tools for build output
	var sb strings.Builder
	messageCh := make(chan string, 100)
	go func() {
		for {
			select {
			case s, ok := <-messageCh:
				if ok {
					sb.WriteString(s)
				}
			}
		}
	}()

	// determine definition ID as string
	idString := strconv.Itoa(definition.Id)

	// set projectpath
	projectPath := basePath + idString + "/" + strconv.FormatInt(time.Now().Unix(), 10)
	buildPath := projectPath + "/build"
	artifactPath := projectPath + "/artifact"
	clonePath := projectPath + "/clone"
	if fileExists(projectPath) {
		err := os.RemoveAll(projectPath)
		if err != nil {
			messageCh <- "could not remove stale project directory (" + projectPath + "): " + err.Error()
			saveBuildReport(definition, sb.String())
			return
		}
	}

	// create a new build directory
	err := os.MkdirAll(buildPath, 0664)
	if err != nil {
		messageCh <- "could not create build directory (" + buildPath + "): " + err.Error()
		saveBuildReport(definition, sb.String())
		return
	}
	err = os.MkdirAll(artifactPath, 0664)
	if err != nil {
		messageCh <- "could not create artifact directory (" + artifactPath + "): " + err.Error()
		saveBuildReport(definition, sb.String())
		return
	}
	err = os.MkdirAll(clonePath, 0664)
	if err != nil {
		messageCh <- "could not create clone directory (" + clonePath + "): " + err.Error()
		saveBuildReport(definition, sb.String())
		return
	}

	// clone the repository
	var withCredentials bool
	if definition.RepoSecret != "" {
		withCredentials = true
	}
	repositoryUrl := getRepositoryUrl(definition, withCredentials)
	cmd := exec.Command("git", "clone", "--single-branch", "--branch", definition.RepoBranch, repositoryUrl, clonePath)
	messageCh <- cmd.String()
	cmdOutput, err := cmd.Output()
	if err != nil {
		messageCh <- "could not get command output: " + err.Error()
		saveBuildReport(definition, sb.String())
		return
	} else {
		messageCh <- string(cmdOutput)
	}

	//db, err := getDbConnection()
	//if err != nil {
	//
	//}

	// determine project type (build target)
	switch definition.BuildTargetId {
	case 1: // golang
		_, err := handleGolangProject(golangBuildDefinition(definition), messageCh, projectPath)
		if err != nil {
			messageCh <- "could not build golang project (" + projectPath + ")"
		} else {
			// set artifact

			// deploy artifact to any host
			//err = deployArtifact(definition, artifact)
		}
	case 2: // dotnet
		err = handleDotnetProject(dotnetBuildDefinition(definition), messageCh, projectPath)
		if err != nil {
			messageCh <- "could not build dotnet project (" + projectPath + ")"
		}
	case 3: // php
		err = handlePhpProject(phpBuildDefinition(definition), messageCh, projectPath)
		if err != nil {
			messageCh <- "could not build php project (" + projectPath + ")"
		}
	case 4: // rust
		err = handleRustProject(rustBuildDefinition(definition), messageCh, projectPath)
		if err != nil {
			messageCh <- "could not build rust project (" + projectPath + ")"
		}
	}

	saveBuildReport(definition, sb.String())
	fmt.Println("info received!") // set a proper response
}

func handleGolangProject(definition golangBuildDefinition, messageCh chan string, projectDir string) (string, error) {
	var err error
	if definition.RunTests {
		err = definition.runTests(messageCh)
		if err != nil {
			//messageCh <- "process cancelled by test run: " + err.Error()
			return "", err
		}
	}

	if definition.RunBenchmarkTests {
		err = definition.runBenchmarkTests(messageCh)
		if err != nil {
			//messageCh <- "process cancelled by benchmark test run: " + err.Error()
			return "", err
		}
	}

	artifact, err := definition.buildArtifact(messageCh, projectDir)
	if err != nil {
		//messageCh <- "build failed: " + err.Error()
		return "", err
	}

	if definition.ApplyMigrations {
		//metaMigrationId := definition.MetaMigrationId

		//err = definition.applyMigrations(messageCh)
		//if err != nil {
		//	messageCh <- "process cancelled by migration application: " + err.Error()
		//	return err
		//}
	}

	db, err := getDbConnection()
	if err != nil {
		messageCh <- "could not establish database connection: " + err.Error()
		return "", err
	}

	var depList []deploymentDefinition
	var dep deploymentDefinition
	rows, err := db.Query("SELECT id, build_definition_id, caption, host, username, password, connection_type, " +
		"working_directory, pre_deployment_actions, post_deployment_actions FROM deployment_definition " +
		"WHERE build_definition_id = ?", definition.Id)
	if err != nil {
		messageCh <- "could not fetch deployment definitions for build definition: " + err.Error()
		return "", err
	}
	for rows.Next() {
		err = rows.Scan(&dep.Id, &dep.BuildDefinitionId, &dep.Caption, &dep.Host, &dep.Username, &dep.Password,
			&dep.ConnectionType, &dep.WorkingDirectory, &dep.PreDeploymentActions, &dep.PostDeploymentActions)
		if err != nil {
			messageCh <- "could not scan row"
			continue
		}
		depList = append(depList, dep)
		dep = deploymentDefinition{}
	}

	if len(depList) > 0 && definition.DeploymentEnabled {
		err = deployArtifact(depList, messageCh, artifact)
	}


	return artifact, nil
}

func handleDotnetProject(definition dotnetBuildDefinition, messageCh chan string, projectDir string) error {

	return nil
}

func handlePhpProject(definition phpBuildDefinition, messageCh chan string, projectDir string) error {

	return nil
}

func handleRustProject(definition rustBuildDefinition, messageCh chan string, projectDir string) error {

	return nil
}

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

func saveBuildReport(definition buildDefinition, report string) {

}

func deployArtifact(deploymentDefinitions []deploymentDefinition, messageCh chan string, artifact string) error {
	for _, deployment := range deploymentDefinitions {
		// first, the pre deployment actions
		sshConfig := &ssh.ClientConfig{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			User:            deployment.Username,
			Auth: []ssh.AuthMethod{
				ssh.Password(deployment.Password),
			},
		}
		sshClient, err := ssh.Dial("tcp", deployment.Host, sshConfig)
		if err != nil {
			return err
		}

		var preDeploymentActions []string
		if deployment.PreDeploymentActions != "" {
			preDeploymentActions = strings.Split(deployment.PreDeploymentActions,"\n")
		}

		if len(preDeploymentActions) > 0 {
			for _, action := range preDeploymentActions {
				session, err := sshClient.NewSession()
				if err != nil {
					return err
				}
				outp, err := session.Output(action)
				if err != nil {
					return err
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
			return err
		}

		// then, the actual deployment
		sftpClient, err := sftp.NewClient(sshClient)
		if err != nil {
			return err
		}

		// create destination file
		// @TODO really necessary?
		dstFile, err := sftpClient.Create(deployment.WorkingDirectory)
		if err != nil {
			return err
		}

		// create source file
		srcFile, err := os.Open(artifact)
		if err != nil {
			return err
		}

		// copy source file to destination file
		bytes, err := io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}
		_ = dstFile.Close()
		_ = srcFile.Close()
		fmt.Printf("%d bytes copied\n", bytes)
		_ = session.Close()

		// then, the post deployment actions

		var postDeploymentActions []string
		if deployment.PostDeploymentActions != "" {
			postDeploymentActions = strings.Split(deployment.PostDeploymentActions,"\n")
		}

		if len(postDeploymentActions) > 0 {
			for _, action := range postDeploymentActions {
				session, err := sshClient.NewSession()
				if err != nil {
					return err
				}
				outp, err := session.Output(action)
				if err != nil {
					return err
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
				return err
			}
		}

		_ = sftpClient.Close()
	}

	return nil
}

func getRepositoryUrl(bd buildDefinition, withCredentials bool) string {
	var url string

	switch bd.RepoHoster {
	case "bitbucket":
		url = "bitbucket.org/" + bd.RepoFullname
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", bd.RepoUsername, bd.RepoSecret, url)
		}
		return "https://" + url
	case "github":
		url = "github.com/" + bd.RepoFullname
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", bd.RepoUsername, bd.RepoSecret, url)
		}
		return "https://" + url
	case "gitlab":
		url = "gitlab.com/" + bd.RepoFullname
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", bd.RepoUsername, bd.RepoSecret, url)
		}
		return "https://" + url
	case "gitea":
		url = bd.RepoHosterUrl + "/" + bd.RepoFullname
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", bd.RepoUsername, bd.RepoSecret, url)
		}
		return "https://" + url

	default:
		url = bd.RepoHosterUrl + "/" + bd.RepoFullname
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", bd.RepoUsername, bd.RepoSecret, url)
		}
		return "http://" + url // just http
	}
}
