package buildservice

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/logging"
	"github.com/sirupsen/logrus"
	"github.com/stvp/slug"
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

	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/fixtures"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var basePath string = "data"

func init() {
	ds := databaseservice.Get()
	settings, err := ds.GetAllSettings()
	if err != nil {
		panic("could not initate buildservice: " + err.Error())
	}

	if path, ok := settings["base_datapath"]; ok && path != "" {
		basePath = path
	}

}

func saveBuildReport(definition entity.BuildDefinition, report, result, artifactPath string, executionTime int64, executedAt time.Time) {
	logger := logging.New(logrus.DebugLevel, "saveBuildReport", true)
	ds := databaseservice.Get()
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
		logger.WithField("error", err.Error()).Error("could not insert new build execution")
	}
}

// StartBuildProcess start the build process for a given build definition
func StartBuildProcess(definition entity.BuildDefinition) {
	// instantiate tools for build output
	var (
		ds            = databaseservice.Get()
		err           error
		binaryName    string
		sb            strings.Builder
		result        = "failed"
		logger        = logging.New(logrus.DebugLevel, "buildProc", true)
		executionTime = time.Now().UnixNano()
		projectPath   = fmt.Sprintf("%s/%d/%d", basePath, definition.Id, executionTime)
		artifactPath  = projectPath + "/artifact"
		clonePath     = projectPath + "/clone"
	)

	messageCh := make(chan string)
	go func() {
		for {
			select {
			case s, ok := <-messageCh:
				if ok {
					m := strings.TrimSpace(s)
					if m != "" {
						sb.WriteString(m + "\n")
					}
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
		logger.Trace("writing report")
		//fmt.Println(time.Now().UnixNano(), executionTime, time.Now().UnixNano() - executionTime)
		saveBuildReport(definition, sb.String(), result, artifactPath, time.Now().UnixNano()-executionTime, time.Now())
	}()

	messageCh <- fmt.Sprintf("setting basePath to %s", basePath)

	err = os.RemoveAll(projectPath)
	if err != nil {
		messageCh <- fmt.Sprintf("could not remove stale project directory (%s): %s", projectPath, err.Error())
		return
	}

	err = os.MkdirAll(artifactPath, 0700)
	if err != nil {
		messageCh <- "could not create artifact directory (" + artifactPath + "): " + err.Error()
		return
	}
	err = os.MkdirAll(clonePath, 0700)
	if err != nil {
		messageCh <- "could not create clone directory (" + clonePath + "): " + err.Error()
		return
	}

	variables, err := ds.GetAvailableVariablesForUser(definition.CreatedBy)
	if err != nil {
		messageCh <- fmt.Sprintf("could not determine variables for user: '%s'", err.Error())
		return
	}
	helper.ReplaceVariables(&definition.Content, variables)

	var content entity.BuildDefinitionContent
	if err = helper.UnmarshalBuildDefinitionContent(definition.Content, &content); err != nil {
		messageCh <- fmt.Sprintf("could not unmarshal build definition: '%s'", err.Error())
		return
	}

	// clone the repository
	withCredentials := content.Repository.AccessSecret != ""
	if withCredentials && content.Repository.AccessUser == "" {
		content.Repository.AccessUser = "nobody"
	}

	// if no branch is set, use master as the default
	if content.Repository.Branch == "" {
		content.Repository.Branch = "master"
	}

	repositoryUrl, err := getRepositoryUrl(content, withCredentials)
	if err != nil {
		messageCh <- fmt.Sprintf("could not determine repository url: %s", err.Error())
		return
	}

	err = cloneRepository(messageCh, content.Repository.Branch, repositoryUrl, clonePath)
	if err != nil {
		messageCh <- fmt.Sprintf("could not clone repository: '%s'", err.Error())
		return
	}

	binaryName = slug.Clean(strings.ToLower(strings.Split(content.Repository.Name, "/")[1]))
	artifactPath = filepath.Join(artifactPath, binaryName)
	helper.ReplaceVariables(&definition.Content, []entity.UserVariable{{
		Variable: "artifact",
		Value:    artifactPath,
	}, {
		Variable: "cloneDir",
		Value:    clonePath,
	}})

	// do the unmarshal again with updated variables
	if err = helper.UnmarshalBuildDefinitionContent(definition.Content, &content); err != nil {
		messageCh <- fmt.Sprintf("could not unmarshal build definition: '%s'", err.Error())
		return
	}

	allSteps := make([]string, 0)
	allSteps = append(allSteps, content.Setup...)
	allSteps = append(allSteps, content.Test...)
	allSteps = append(allSteps, content.PreBuild...)
	allSteps = append(allSteps, content.Build...)
	allSteps = append(allSteps, content.PostBuild...)

	for _, step := range allSteps {
		switch true {
		case strings.HasPrefix(step, "setenv"):
			parts := strings.Split(step, " ")
			if len(parts) != 3 {
				messageCh <- fmt.Sprintf("step '%s' has an invalid format", step)
				return
			}
			err = os.Setenv(parts[1], parts[2])
			if err != nil {
				messageCh <- fmt.Sprintf("step '%s' was not successful: '%s'", step, err.Error())
				return
			}
		case strings.HasPrefix(step, "unsetenv"):
			parts := strings.Split(step, " ")
			if len(parts) != 2 {
				messageCh <- fmt.Sprintf("step '%s' has an invalid format", step)
				return
			}
			err = os.Unsetenv(parts[1])
			if err != nil {
				messageCh <- fmt.Sprintf("step '%s' was not successful: '%s'", step, err.Error())
				return
			}
		case strings.HasPrefix(step, "go build"):
			if strings.ToLower(os.Getenv("GOOS")) == "windows" {
				artifactPath += ".exe"
			}
			parts := strings.Split(step, " ")
			if len(parts) != 3 {
				messageCh <- fmt.Sprintf("step '%s' has an invalid format", step)
				return
			}
			cmd := exec.Command("go", "build", "-o", artifactPath, "-ldflags", "-s -w", parts[2])
			b, err := cmd.CombinedOutput()
			if err != nil {
				messageCh <- fmt.Sprintf("could not execute command '%s': '%s' -> (%s)", cmd.String(), err.Error(), string(b))
				return
			}
			messageCh <- string(b)
		default:
			parts := strings.Split(step, " ")
			var cmd *exec.Cmd
			if len(parts) <= 0 { // :)
				messageCh <- "empty step"
				return
			} else if len(parts) == 1 {
				cmd = exec.Command(parts[0])
			} else {
				cmd = exec.Command(parts[0], strings.Join(parts[1:], " "))
			}

			messageCh <- fmt.Sprintf("step: %s", cmd.String())

			b, err := cmd.CombinedOutput()
			if err != nil {
				messageCh <- fmt.Sprintf("could not execute command '%s': '%s' -> (%s)", cmd.String(), err.Error(), string(b))
				return
			}

			messageCh <- string(b)
		}

	}

	err = deployArtifact(messageCh, content, artifactPath)
	if err != nil {
		messageCh <- fmt.Sprintf("could not deploy artifact")
		return
	}

	logger.Trace("build succeeded")
	result = "success"
}

func cloneRepository(messageCh chan string, branch string, repositoryUrl string, path string) error {
	//commandParts := strings.Split(fmt.Sprintf("git clone --single-branch --branch %s %s %s", content.Repository.Branch, repositoryUrl, clonePath), " ")
	cmd := exec.Command("git", "clone", "--single-branch", "--branch", branch, repositoryUrl, path)
	messageCh <- "clone repository command: " + cmd.String()
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	messageCh <- string(cmdOutput)
	return nil
}

func deployArtifact(messageCh chan string, cont entity.BuildDefinitionContent, artifact string) error {
	messageCh <- fmt.Sprintf("artifact to be deployed: %s", artifact)
	localDeploymentCount := len(cont.Deployments.LocalDeployments)
	if localDeploymentCount > 0 {
		messageCh <- fmt.Sprintf("%d local deployment(s) found", localDeploymentCount)
		for _, deployment := range cont.Deployments.LocalDeployments {
			if !deployment.Enabled {
				messageCh <- "skipping disabled deployment"
				continue
			}

			fileBytes, err := ioutil.ReadFile(artifact)
			if err != nil {
				messageCh <- fmt.Sprintf("could not read artifact (%s): %s", artifact, err.Error())
				return err
			}

			_ = os.MkdirAll(filepath.Dir(deployment.Path), 0744)

			err = os.WriteFile(deployment.Path, fileBytes, 0744)
			if err != nil {
				messageCh <- fmt.Sprintf("could not write artifact (%s) to target (%s): %s", artifact, deployment.Path, err.Error())
				return err
			}
		}
	} else {
		messageCh <- "no local deployments found"
	}

	emailDeploymentCount := len(cont.Deployments.EmailDeployments)
	if emailDeploymentCount > 0 {
		messageCh <- fmt.Sprintf("%d email deployment(s) found", emailDeploymentCount)

		ds := databaseservice.Get()
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
				messageCh <- "skipping disabled deployment"
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
				messageCh <- "skipping disabled deployment"
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
//	ds := databaseService.Get()
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
				return fmt.Errorf("bitbucket: could not get header %s", headers[i])
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
		if err != nil {
			return fmt.Errorf("gitea: could not decode json payload: %s", err.Error())
		}
		_ = r.Body.Close()

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

func getCurrentVersionTag() string {
	cmd := exec.Command("git", "tag", "-l", "--sort=-version:refname")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	versions := strings.Split(strings.ReplaceAll(strings.TrimSpace(string(output)), "\r\n", "\n"), "\n")
	if len(versions) > 0 {
		return versions[0]
	}

	return ""
}
