package buildservice

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildsteps"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/fixtures"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
	"github.com/pkg/sftp"
	"github.com/stvp/slug"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var basePath string = "data/"

func saveBuildReport(definition entity.BuildDefinition, report, result, artifactPath string, executionTime int64, executedAt time.Time) {
	ds := databaseservice.New()
	be := entity.BuildExecution{
		BuildDefinitionId: definition.Id,
		ActionLog:         report,
		Result:            result,
		ArtifactPath:      artifactPath,
		ExecutionTime:     math.Round(float64(executionTime)/(1000*1000*1000)*100) / 100,
		ExecutedAt:        executedAt,
	}

	err := ds.AddBuildExecution(be)
	if err != nil {
		helper.WriteToConsole("saveBuildReport: could not create new build execution: " + err.Error())
	}
}

// StartBuildProcess start the build process for a given build definition
func StartBuildProcess(definition entity.BuildDefinition, content entity.BuildDefinitionContent) {
	// instantiate tools for build output
	var (
		err           error
		sb            strings.Builder
		result        = "failed"
		executionTime = time.Now().UnixNano()

		projectPath  = fmt.Sprintf("%s%d/%d", basePath, definition.Id, time.Now().Unix())
		buildPath    = projectPath + "/build"
		artifactPath = projectPath + "/artifact"
		clonePath    = projectPath + "/clone"
	)

	messageCh := make(chan string)
	go func() {
		for {
			select {
			case s, ok := <-messageCh:
				if ok {
					helper.WriteToConsole(s)
					sb.WriteString(strings.TrimSpace(s) + "\n")
				} else {
					return
				}
			default:
				// waiting
			}

		}
	}()
	defer func() {
		close(messageCh)
		helper.WriteToConsole("writing report")
		//fmt.Println(time.Now().UnixNano(), executionTime, time.Now().UnixNano() - executionTime)
		saveBuildReport(definition, sb.String(), result, artifactPath, time.Now().UnixNano()-executionTime, time.Now())
	}()
	ds := databaseservice.New()
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
	repositoryUrl, err := getRepositoryUrl(content, withCredentials)
	if err != nil {
		messageCh <- fmt.Sprintf("could not determine repository url: %s", err.Error())
		return
	}
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
		baseDataPath = "."
	}
	messageCh <- fmt.Sprintf("setting baseDataPath to %s", baseDataPath)

	switch content.ProjectType {
	case "go":
		fallthrough
	case "golang":
		def := buildsteps.GolangBuildDefinition{
			CloneDir:    clonePath,    // strings.ToLower(fmt.Sprintf("%s/%s/%s/clone", baseDataPath, content.Repository.Hoster, slug.Clean(content.Repository.Name))),
			ArtifactDir: artifactPath, // strings.ToLower(fmt.Sprintf("%s/%s/%s/artifact", baseDataPath, content.Repository.Hoster, slug.Clean(content.Repository.Name))),
			MetaData:    definition,
			Content:     content,
		}
		artifactPath, err = handleGolangProject(def, messageCh, projectPath)
		if err != nil {
			messageCh <- fmt.Sprintf("could not build golang project (%s): %s", projectPath, err.Error())
			return
		}
	case "csharp":
		fallthrough
	case "fsharp":
		fallthrough
	case "vb":
		fallthrough
	case "visualbasic":
		fallthrough
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
	helper.WriteToConsole("build succeeded")
	result = "success"
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

	// TODO gehört eigentlich eine Ebene höher
	//if len(definition.Content.Deployments.EmailDeployments) > 0 || len(definition.Content.Deployments.RemoteDeployments) > 0 {
		err = deployArtifact(definition.Content, messageCh, artifact)
	//}
	// TODO handle err

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
	messageCh <- fmt.Sprintf("artifact to be deployed: %s", artifact)
	localDeploymentCount := len(cont.Deployments.LocalDeployments)
	if localDeploymentCount > 0 {
		messageCh <- fmt.Sprintf("%d local deployment(s) found", localDeploymentCount)
		for _, deployment := range cont.Deployments.LocalDeployments {
			if !deployment.Enabled {
				continue
			}

			fileBytes, err := ioutil.ReadFile(artifact)
			if err != nil {
				messageCh <- fmt.Sprintf("could not read artifact (%s): %s", artifact, err.Error())
				return err
			}
			targetPath := deployment.Path

			if deployment.Path[len(deployment.Path)-1:] == "/" {
				targetPath = filepath.Join(deployment.Path, filepath.Base(artifact))
			}

			err = ioutil.WriteFile(targetPath, fileBytes, 0744)
			if err != nil {
				messageCh <- fmt.Sprintf("could not write artifact (%s) to target (%s): %s", artifact, targetPath, err.Error())
				return err
			}
		}
	} else {
		messageCh <- "no local deployments found"
	}

	emailDeploymentCount := len(cont.Deployments.EmailDeployments)
	if emailDeploymentCount > 0 {
		messageCh <- fmt.Sprintf("%d email deployment(s) found", emailDeploymentCount)

		ds := databaseservice.New()
		settings, err := ds.GetAllSettings()
		if err != nil {
			messageCh <- fmt.Sprintf("email deplyoments: could not read settings: %s", err.Error())
			return err
		}

		artifactContent, err := ioutil.ReadFile(artifact)
		if err != nil {
			messageCh <- "could not read artifact file: " + err.Error()
			return err
		}

		zipBuffer := bytes.Buffer{}
		zipWriter := zip.NewWriter(&zipBuffer)

		zipFile, err := zipWriter.Create(filepath.Base(artifact))
		if err != nil {
			messageCh <- "could not create artifact file in zip archive: " + err.Error()
			return err
		}
		_, err = zipFile.Write(artifactContent)
		if err != nil {
			messageCh <- "could not write artifact file to zip archive: " + err.Error()
			return err
		}
		_ = zipWriter.Close()

		zipArchiveName := artifact + ".zip"
		err = ioutil.WriteFile(zipArchiveName, zipBuffer.Bytes(), 0744)
		if err != nil {
			messageCh <- "could not write zip archive bytes to file: " + err.Error()
			return err
		}

		for _, deployment := range cont.Deployments.EmailDeployments {
			if !deployment.Enabled {
				continue
			}

			data := struct {
				Version string
				Title   string
			}{
				Version: "n/a", // TODO
				Title:   cont.Repository.Name,
			}

			emailBody, err := templateservice.ParseEmailTemplate(string(fixtures.DeploymentEmail), data)
			if err != nil {
				messageCh <- fmt.Sprintf("could not parse deployment email template: %s", err.Error())
				return err
			}
			err = helper.SendEmail(
				settings,
				emailBody,
				fixtures.EmailSubjects[fixtures.DeploymentEmail],
				[]string{deployment.Address},
				[]string{zipArchiveName},
			)
			if err != nil {
				messageCh <- fmt.Sprintf("could not send out deployment email to %s: %s", deployment.Address, err.Error())
				return err
			}
			messageCh <- fmt.Sprintf("deployment email sent to recipient %s", deployment.Address)
		}
	} else {
		messageCh <- "no email deployments found"
	}

	remoteDeploymentCount := len(cont.Deployments.RemoteDeployments)
	if remoteDeploymentCount > 0 {
		messageCh <- fmt.Sprintf("%d remote deployment(s) found", remoteDeploymentCount)
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
	} else {
		messageCh <- "no remote deplyoments found"
	}

	return nil
}

func getRepositoryUrl(cont entity.BuildDefinitionContent, withCredentials bool) (string, error) {
	//var url string
	switch cont.Repository.Hoster {
	case "local":
		return cont.Repository.Url, nil
	default:
		urlParts, err := url.ParseRequestURI(cont.Repository.Url)
		if err != nil {
			return "", err
		}
		if !withCredentials {
			return urlParts.String(), nil
		}
		urlParts.User = url.UserPassword(cont.Repository.AccessUser, cont.Repository.AccessSecret)
		return urlParts.String(), nil
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

// CheckPayloadHeader checks the existence and values taken from HTTP request headers
// from the given HTTP request
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
			return fmt.Errorf("gitea: repository names do not match (from payload: %s, from build definition: %s)"+payload.Repository.FullName, content.Repository.Name)
		}
	default:
		return fmt.Errorf("unrecognized git hoster %s", content.Repository.Hoster)
	}

	return nil
}
