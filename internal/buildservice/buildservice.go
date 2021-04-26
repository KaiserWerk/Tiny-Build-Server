package buildservice

import (
	"encoding/json"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildsteps"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/pkg/sftp"
	"github.com/stvp/slug"
	"golang.org/x/crypto/ssh"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var basePath string = "data/"

func saveBuildReport(definition entity.BuildDefinition, report, result, artifactPath string, executionTime int64, executedAt time.Time) {
	ds := databaseService.New()
	be := entity.BuildExecution{
		BuildDefinitionId: definition.Id,
		ActionLog:         report,
		Result:            result,
		ArtifactPath:      artifactPath,
		ExecutionTime:     float64(executionTime / 60),
		ExecutedAt:        executedAt,
	}

	err := ds.AddBuildExecution(be)
	if err != nil {
		helper.WriteToConsole("saveBuildReport: could not create new build execution: " + err.Error())
	}
}

func StartBuildProcess(definition entity.BuildDefinition, content entity.BuildDefinitionContent) {
	// instantiate tools for build output
	var (
		err error
		sb strings.Builder
		result = "failed"
		executionTime = time.Now().Unix()

		projectPath = fmt.Sprintf("%s%d/%d", basePath, definition.Id, time.Now().Unix())
		buildPath = projectPath + "/build"
		artifactPath = projectPath + "/artifact"
		clonePath = projectPath + "/clone"
	)

	messageCh := make(chan string, 100)
	go func() {
		for {
			s, ok := <-messageCh
			if ok {
				helper.WriteToConsole(s)
				sb.WriteString(strings.TrimSpace(s) + "\n")
			} else {
				return
			}
		}
	}()
	defer func() {
		close(messageCh)
		saveBuildReport(definition, sb.String(), result, artifactPath, time.Now().Unix() - executionTime, time.Now())
	}()
	ds := databaseService.New()
	//defer ds.Quit()

	//if helper.FileExists(projectPath) {
		err = os.RemoveAll(projectPath)
		if err != nil {
			messageCh <- fmt.Sprintf("could not remove stale project directory (%s): %s", projectPath, err.Error())
			return
		}
	//}

	// create a new build directory
	err = os.MkdirAll(buildPath, 0744)
	if err != nil {
		messageCh <- "could not create build directory (" + buildPath + "): " + err.Error()
		return
	}
	err = os.MkdirAll(artifactPath, 0744)
	if err != nil {
		messageCh <- "could not create artifact directory (" + artifactPath + "): " + err.Error()
		return
	}
	err = os.MkdirAll(clonePath, 0744)
	if err != nil {
		messageCh <- "could not create clone directory (" + clonePath + "): " + err.Error()
		return
	}

	// clone the repository
	var withCredentials bool
	if content.Repository.AccessSecret != "" {
		withCredentials = true
	}
	repositoryUrl := getRepositoryUrl(content, withCredentials)
	commandParts := strings.Split(fmt.Sprintf("git clone --single-branch --branch %s %s %s", content.Repository.Branch, repositoryUrl, clonePath), " ")
	cmd := exec.Command(commandParts[0], commandParts[1:]...)
	messageCh <- "clone repository command: " + cmd.String()
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		messageCh <- "could not get command output: " + err.Error()
		return
	} else {
		messageCh <- string(cmdOutput)
	}

	settings, err := ds.GetAllSettings()
	if err != nil {
		messageCh <- "could not obtain setting values: " + err.Error()
		return
	}

	baseDataPath, ok := settings["base_datapath"]
	if !ok || baseDataPath == "" {
		messageCh <- "could not fetch base data path, falling back to default"
	} else {
		baseDataPath = "./"
	}


	switch content.ProjectType {
	case "go":
		fallthrough
	case "golang":
		def := buildsteps.GolangBuildDefinition{
			CloneDir:    strings.ToLower(fmt.Sprintf("%s/%s/%s/clone", baseDataPath, content.Repository.Hoster, slug.Clean(content.Repository.Name))),
			ArtifactDir: strings.ToLower(fmt.Sprintf("%s/%s/%s/artifact", baseDataPath, content.Repository.Hoster, slug.Clean(content.Repository.Name))),
			MetaData:    definition,
			Content:     content,
		}
		_, err = handleGolangProject(def, messageCh, projectPath)
		if err != nil {
			messageCh <- "could not build golang project (" + projectPath + ")"
			return
		} else {
			// set artifact

			// deploy artifact to any host
			//err = deployArtifact(definition, artifact)
		}
	case "dotnet":
		def := buildsteps.DotnetBuildDefinition{
			CloneDir:    strings.ToLower(fmt.Sprintf("%s/%s/%s/clone", baseDataPath, content.Repository.Hoster, slug.Clean(content.Repository.Name))),
			ArtifactDir: strings.ToLower(fmt.Sprintf("%s/%s/%s/artifact", baseDataPath, content.Repository.Hoster, slug.Clean(content.Repository.Name))),
			MetaData:    definition,
			Content:     content,
		}
		err = handleDotnetProject(def, messageCh, projectPath)
		if err != nil {
			messageCh <- "could not build dotnet project (" + projectPath + ")"
			return
		}
	case "php":
		def := buildsteps.PhpBuildDefinition{
			CloneDir:    strings.ToLower(fmt.Sprintf("%s/%s/%s/clone", baseDataPath, content.Repository.Hoster, slug.Clean(content.Repository.Name))),
			ArtifactDir: strings.ToLower(fmt.Sprintf("%s/%s/%s/artifact", baseDataPath, content.Repository.Hoster, slug.Clean(content.Repository.Name))),
			MetaData:    definition,
			Content:     content,
		}
		err = handlePhpProject(def, messageCh, projectPath)
		if err != nil {
			messageCh <- "could not build php project (" + projectPath + ")"
			return
		}
	case "rust":
		def := buildsteps.RustBuildDefinition{
			CloneDir:    strings.ToLower(fmt.Sprintf("%s/%s/%s/clone", baseDataPath, content.Repository.Hoster, slug.Clean(content.Repository.Name))),
			ArtifactDir: strings.ToLower(fmt.Sprintf("%s/%s/%s/artifact", baseDataPath, content.Repository.Hoster, slug.Clean(content.Repository.Name))),
			MetaData:    definition,
			Content:     content,
		}
		err = handleRustProject(def, messageCh, projectPath)
		if err != nil {
			messageCh <- "could not build rust project (" + projectPath + ")"
			return
		}
	}

	fmt.Println("info received!") // set a proper response
}

func handleGolangProject(definition buildsteps.GolangBuildDefinition, messageCh chan string, projectDir string) (string, error) {
	var err error

	//if definition. {
	//	err = definition.RunTests(messageCh)
	//	if err != nil {
	//		//messageCh <- "process cancelled by test run: " + err.Error()
	//		return "", err
	//	}
	//}
	//
	//if definition.RunBenchmarkTests {
	//	err = definition.RunBenchmarkTests(messageCh)
	//	if err != nil {
	//		//messageCh <- "process cancelled by benchmark test run: " + err.Error()
	//		return "", err
	//	}
	//}

	artifact, err := definition.BuildArtifact(messageCh, projectDir)
	if err != nil {
		//messageCh <- "build failed: " + err.Error()
		return "", err
	}

	//if definition.ApplyMigrations {
	//	//metaMigrationId := definition.MetaMigrationId
	//
	//	//err = definition.applyMigrations(messageCh)
	//	//if err != nil {
	//	//	messageCh <- "process cancelled by migration application: " + err.Error()
	//	//	return err
	//	//}
	//}

	//db := global.GetDbConnection()
	//var depList []entity.DeploymentDefinition
	//var dep entity.DeploymentDefinition
	//rows, err := db.Query("SELECT id, build_definition_id, caption, host, username, password, connection_type, "+
	//	"working_directory, pre_deployment_actions, post_deployment_actions FROM deployment_definition "+
	//	"WHERE build_definition_id = ?", definition.MetaData.Id)
	//if err != nil {
	//	messageCh <- "could not fetch deployment definitions for build definition: " + err.Error()
	//	return "", err
	//}
	//for rows.Next() {
	//	err = rows.Scan(&dep.Id, &dep.BuildDefinitionId, &dep.Caption, &dep.Host, &dep.Username, &dep.Password,
	//		&dep.ConnectionType, &dep.WorkingDirectory, &dep.PreDeploymentActions, &dep.PostDeploymentActions)
	//	if err != nil {
	//		messageCh <- "could not scan row"
	//		continue
	//	}
	//	depList = append(depList, dep)
	//	dep = entity.DeploymentDefinition{}
	//}

	if len(definition.Content.Deployments.EmailDeployments) > 0 || len(definition.Content.Deployments.RemoteDeployments) > 0 {
		err = deployArtifact(definition.Content, messageCh, artifact)
	}

	return artifact, nil
}

func handleDotnetProject(definition buildsteps.DotnetBuildDefinition, messageCh chan string, projectDir string) error {

	return nil
}

func handlePhpProject(definition buildsteps.PhpBuildDefinition, messageCh chan string, projectDir string) error {

	return nil
}

func handleRustProject(definition buildsteps.RustBuildDefinition, messageCh chan string, projectDir string) error {

	return nil
}

func deployArtifact(cont entity.BuildDefinitionContent, messageCh chan string, artifact string) error {
	if len(cont.Deployments.LocalDeployments) > 0 {
		for _, deployment := range cont.Deployments.LocalDeployments {
			if !deployment.Enabled {
				continue
			}
			// TODO implement
		}
	}

	if len(cont.Deployments.EmailDeployments) > 0 {
		for _, deployment := range cont.Deployments.EmailDeployments {
			if !deployment.Enabled {
				continue
			}
			// TODO implement
		}
	}

	if len(cont.Deployments.RemoteDeployments) > 0 {
		for _, deployment := range cont.Deployments.RemoteDeployments {
			if !deployment.Enabled {
				continue
			}
			// first, the pre deployment actions
			sshConfig := &ssh.ClientConfig{
				//HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				User: deployment.Username,
				Auth: []ssh.AuthMethod{
					ssh.Password(deployment.Password),
				},
			}
			sshClient, err := ssh.Dial("tcp", deployment.Host, sshConfig)
			if err != nil {
				return err
			}

			if len(deployment.PreDeploymentSteps) > 0 {
				for _, action := range deployment.PreDeploymentSteps {
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
							messageCh <- "output from pre remote command: " + outpDisplay
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
			// TODO really necessary?
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
			if len(deployment.PostDeploymentSteps) > 0 {
				for _, action := range deployment.PostDeploymentSteps {
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
							messageCh <- "output from post remote command: " + outpDisplay
						}
					}
					_ = session.Close()
				}
			}

			_ = sftpClient.Close()
			_ = sshClient.Close()
		}
	}

	return nil
}

func getRepositoryUrl(cont entity.BuildDefinitionContent, withCredentials bool) string {
	var url string
	switch cont.Repository.Hoster {
	case "local":
		return cont.Repository.Url
	case "bitbucket":
		url = "bitbucket.org/" + cont.Repository.Name
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", cont.Repository.AccessUser, cont.Repository.AccessSecret, url)
		}
		return "https://" + url
	case "github":
		url = "github.com/" + cont.Repository.Name
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", cont.Repository.AccessUser, cont.Repository.AccessSecret, url)
		}
		return "https://" + url
	case "gitlab":
		url = "gitlab.com/" + cont.Repository.Name
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", cont.Repository.AccessUser, cont.Repository.AccessSecret, url)
		}
		return "https://" + url
	case "gitea":
		url = cont.Repository.Url + "/" + cont.Repository.Name
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", cont.Repository.AccessUser, cont.Repository.AccessSecret, url)
		}
		return "https://" + url

	default:
		url = cont.Repository.Url + "/" + cont.Repository.Name
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", cont.Repository.AccessUser, cont.Repository.AccessSecret, url)
		}
		return "http://" + url // just http
	}
}

//func GetBuilDefinitionFromRequest(r *http.Request) (entity.BuildDefinition, error) {
//	buildDefinition := entity.BuildDefinition{}
//
//	// get DB connection
//	ds := databaseService.New()
//	buildDefinition, err := ds.FindBuildDefinition("token = ?", token)
//	if err != nil {
//		return buildDefinition, fmt.Errorf("build definition cannot be found in database for token %s", token)
//	}
//
//	var cont entity.BuildDefinitionContent
//	err = yaml.Unmarshal([]byte(buildDefinition.Content), &cont)
//	if err != nil {
//		return entity.BuildDefinition{}, fmt.Errorf("could not unmarshal build definition content yaml for token %s", token)
//	}
//}

func CheckPayloadHeader(content entity.BuildDefinitionContent, r *http.Request) error {
	var err error
	// check relevant headers and payload values
	switch content.Repository.Hoster {
	case "bitbucket":
		headers := []string{"X-Event-RegToken", "X-Hook-UUID", "X-Request-UUID", "X-Attempt-Number"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = helper.GetHeaderIfSet(r, headers[i])
			if err != nil {
				return fmt.Errorf("bitbucket: could not get bitbucket header %s", headers[i])
			}
		}

		var payload entity.BitBucketPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return fmt.Errorf("bitbucket: could not decode json payload: %s", err.Error())
		}
		_ = r.Body.Close()
		if payload.Push.Changes[0].New.Name != content.Repository.Branch {
			return fmt.Errorf("bitbucket: branch names do not match (from payload: %s, from build definition: %s)", payload.Push.Changes[0].New.Name, content.Repository.Branch)
		}
		if payload.Repository.FullName != content.Repository.Name {
			return fmt.Errorf("bitbucket: repository names do not match (from payload: %s, from build definition: %s)", payload.Repository.FullName, content.Repository.Name)
		}
	case "github":
		headers := []string{"X-GitHub-Delivery", "X-GitHub-Event", "X-Hub-Signature"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = helper.GetHeaderIfSet(r, headers[i])
			if err != nil {
				return fmt.Errorf("github: could not get github header %s", headers[i])
			}
		}

		var payload entity.GitHubPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return fmt.Errorf("github: could not decode json payload")
		}
		_ = r.Body.Close()
		if payload.Repository.DefaultBranch != content.Repository.Branch {
			return fmt.Errorf("github: branch names do not match (from payload: %s, from build definition: %s)", payload.Repository.DefaultBranch, content.Repository.Branch)
		}
		if payload.Repository.FullName != content.Repository.Name {
			return fmt.Errorf("github: repository names do not match (from payload: %s, from build definition: %s)", payload.Repository.FullName, content.Repository.Name)
		}
	case "gitlab":
		headers := []string{"X-GitLab-Event"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = helper.GetHeaderIfSet(r, headers[i])
			if err != nil {
				return fmt.Errorf("gitlab: could not get gitlab header " + headers[i])
			}
		}

		var payload entity.GitLabPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		_ = r.Body.Close()
		if err != nil {
			return fmt.Errorf("gitlab: could not decode json payload: %s", err.Error())
		}
		branch := strings.Split(payload.Ref, "/")[2]
		if branch != content.Repository.Branch {
			return fmt.Errorf("gitlab: branch names do not match (from payload: %s, from build definition: %s)", branch, content.Repository.Branch)
		}
		if payload.Project.PathWithNamespace != content.Repository.Name {
			return fmt.Errorf("gitlab: repository names do not match (from payload: %s, from build definition: %s)", payload.Project.PathWithNamespace, content.Repository.Name)
		}
	case "gitea":
		headers := []string{"X-Gitea-Delivery", "X-Gitea-Event"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = helper.GetHeaderIfSet(r, headers[i])
			if err != nil {
				return fmt.Errorf("gitea: could not get gitea header %s", headers[i])
			}
		}

		var payload entity.GiteaPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		_ = r.Body.Close()
		if err != nil {
			return fmt.Errorf("gitea: could not decode json payload: %s", err.Error())
		}
		branch := strings.Split(payload.Ref, "/")[2]
		if branch != content.Repository.Branch {
			return fmt.Errorf("gitea: branch names do not match (from payload: %s, from build definition: %s)", branch, content.Repository.Branch)
		}
		if payload.Repository.FullName != content.Repository.Name {
			return fmt.Errorf("gitea: repository names do not match (from payload: %s, from build definition: %s)" + payload.Repository.FullName, content.Repository.Name)
		}
	default:
		return fmt.Errorf("unrecognized git hoster %s", content.Repository.Hoster)
	}

	return nil
}
