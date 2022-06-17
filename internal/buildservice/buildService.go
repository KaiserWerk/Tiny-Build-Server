package buildservice

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/deploymentservice"
	"github.com/KaiserWerk/sessionstore/v2"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/sirupsen/logrus"
	"github.com/stvp/slug"
	"golang.org/x/sync/errgroup"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
)

var basePath string = "data"

type BuildService struct {
	BasePath  string
	SessMgr   *sessionstore.SessionManager
	Logger    *logrus.Entry
	Ds        *databaseservice.DatabaseService
	MessageCh chan string
}

func New(basePath string, sessMgr *sessionstore.SessionManager, logger *logrus.Entry, ds *databaseservice.DatabaseService) (*BuildService, error) {
	bs := BuildService{
		BasePath:  basePath,
		SessMgr:   sessMgr,
		Logger:    logger.WithField("context", "buildSvc"),
		Ds:        ds,
		MessageCh: make(chan string, 100),
	}
	settings, err := ds.GetAllSettings()
	if err != nil {
		return nil, fmt.Errorf("could not initate buildservice: %s", err.Error())
	}

	if path, ok := settings["base_datapath"]; ok && path != "" {
		bs.BasePath = path
	}

	return &bs, nil
}

func (bs *BuildService) SaveBuildReport(definition entity.BuildDefinition, report, result, artifactPath string, executionTime int64, executedAt time.Time, userId uint) {
	be := entity.BuildExecution{
		BuildDefinitionID: definition.ID,
		ManuallyRunBy:     userId,
		ActionLog:         report,
		Result:            result,
		ArtifactPath:      artifactPath,
		ExecutionTime:     math.Round(float64(executionTime)/(1000*1000*1000)*100) / 100,
		ExecutedAt:        executedAt,
	}

	err := bs.Ds.AddBuildExecution(be)
	if err != nil {
		bs.Logger.WithField("error", err.Error()).Error("could not insert new build execution")
	}
}

// StartBuildProcess start the build process for a given build definition
func (bs *BuildService) StartBuildProcess(definition entity.BuildDefinition, userId uint) {
	// instantiate tools for build output
	var (
		err           error
		binaryName    string
		sb            strings.Builder
		result        = "failed"
		executionTime = time.Now().UnixNano()
		projectPath   = fmt.Sprintf("%s/%d/%d", basePath, definition.ID, executionTime)
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
		bs.Logger.Trace("writing report")
		//fmt.Println(time.Now().UnixNano(), executionTime, time.Now().UnixNano() - executionTime)
		bs.SaveBuildReport(definition, sb.String(), result, artifactPath, time.Now().UnixNano()-executionTime, time.Now(), userId)
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

	variables, err := bs.Ds.GetAvailableVariablesForUser(definition.CreatedBy)
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

	repositoryUrl, err := bs.getRepositoryUrl(content, withCredentials)
	if err != nil {
		messageCh <- fmt.Sprintf("could not determine repository url: %s", err.Error())
		return
	}

	err = bs.CloneRepository(content.Repository.Branch, repositoryUrl, clonePath)
	if err != nil {
		messageCh <- fmt.Sprintf("could not clone repository: '%s'", err.Error())
		return
	}

	binaryName = slug.Clean(strings.ToLower(strings.Split(content.Repository.Name, "/")[1]))

	artifact := entity.Artifact{
		Directory: artifactPath,
		File:      binaryName,
		Version:   "",
	}

	helper.ReplaceVariables(&definition.Content, []entity.UserVariable{{
		Variable: "artifact",
		Value:    artifact.FullPath(),
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
		messageCh <- fmt.Sprintf("step: %s", step)
		switch true {
		case strings.HasPrefix(step, "setenv"):
			parts, err := splitCommand(step)
			if err != nil {
				messageCh <- fmt.Sprintf("could not prepare step command '%s': %s", step, err.Error())
			}
			if len(parts) != 3 {
				messageCh <- fmt.Sprintf("step '%s' has an invalid format", step)
				continue
			}

			if err = os.Setenv(parts[1], parts[2]); err != nil {
				messageCh <- fmt.Sprintf("step '%s' was not successful: '%s'", step, err.Error())
				continue
			}
		case strings.HasPrefix(step, "unsetenv"):
			parts, err := splitCommand(step)
			if err != nil {
				messageCh <- fmt.Sprintf("could not prepare step command '%s': %s", step, err.Error())
			}
			if len(parts) != 2 {
				messageCh <- fmt.Sprintf("step '%s' has an invalid format", step)
				continue
			}
			err = os.Unsetenv(parts[1])
			if err != nil {
				messageCh <- fmt.Sprintf("step '%s' was not successful: '%s'", step, err.Error())
				continue
			}
		//case strings.HasPrefix(step, "go build"):
		//	if strings.ToLower(os.Getenv("GOOS")) == "windows" {
		//		artifact.File += ".exe"
		//	}
		//	parts, err := splitCommand(step)
		//	if err != nil {
		//		messageCh <- fmt.Sprintf("could not prepare step command '%s': %s", step, err.Error())
		//	}
		//	if len(parts) != 3 {
		//		messageCh <- fmt.Sprintf("step '%s' has an invalid format", step)
		//		return
		//	}
		//	cmd := exec.Command("go", "build", "-v", "-o", artifact.FullPath(), "-ldflags", "-s -w", parts[2])
		//	b, err := cmd.CombinedOutput()
		//	if err != nil {
		//		messageCh <- fmt.Sprintf("could not execute command '%s': '%s' -> (%s)", cmd.String(), err.Error(), string(b))
		//		return
		//	}
		//	messageCh <- string(b)
		default:
			parts, err := splitCommand(step)
			if err != nil {
				messageCh <- fmt.Sprintf("could not prepare step command '%s': %s", step, err.Error())
			}
			var cmd *exec.Cmd
			if len(parts) <= 0 { // :)
				messageCh <- "empty step"
				continue
			} else if len(parts) == 1 {
				cmd = exec.Command(parts[0])
			} else {
				cmd = exec.Command(parts[0], parts[1:]...)
			}

			cmd.Dir = clonePath

			b, err := cmd.CombinedOutput()
			if err != nil {
				messageCh <- fmt.Sprintf("could not execute command '%s': '%s' -> (%s)", cmd.String(), err.Error(), string(b))
				continue
			}

			messageCh <- string(b)
		}

	}

	err = bs.DeployArtifact(content, artifact)
	if err != nil {
		messageCh <- fmt.Sprintf("could not deploy artifact: " + err.Error())
		return
	}

	bs.Logger.Trace("build succeeded")
	result = "success"
}

func (bs *BuildService) CloneRepository(branch string, repositoryUrl string, path string) error {
	//commandParts := strings.Split(fmt.Sprintf("git clone --single-branch --branch %s %s %s", content.Repository.Branch, repositoryUrl, clonePath), " ")
	cmd := exec.Command("git", "clone", "--single-branch", "--branch", branch, repositoryUrl, path)
	bs.MessageCh <- "clone repository command: " + cmd.String()
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	bs.MessageCh <- string(cmdOutput)
	return nil
}

func (bs *BuildService) DeployArtifact(cont entity.BuildDefinitionContent, artifact entity.Artifact) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	eg, _ := errgroup.WithContext(ctx)
	bs.MessageCh <- fmt.Sprintf("artifact to be deployed: %s", artifact)
	localDeploymentCount := len(cont.Deployments.LocalDeployments)
	if localDeploymentCount > 0 {
		bs.MessageCh <- fmt.Sprintf("%d local deployment(s) found", localDeploymentCount)
		for _, deployment := range cont.Deployments.LocalDeployments {
			eg.Go(func() error {
				if err := deploymentservice.DoLocalDeployment(&deployment, artifact); err != nil {
					return err
				}
				return nil
			})
		}
	} else {
		bs.MessageCh <- "no local deployments found"
	}

	emailDeploymentCount := len(cont.Deployments.EmailDeployments)
	if emailDeploymentCount > 0 {
		bs.MessageCh <- fmt.Sprintf("%d email deployment(s) found", emailDeploymentCount)

		settings, err := bs.Ds.GetAllSettings()
		if err != nil {
			bs.MessageCh <- fmt.Sprintf("email deplyoments: could not read settings: %s", err.Error())
			return err
		}

		artifactContent, err := ioutil.ReadFile(artifact.FullPath())
		if err != nil {
			bs.MessageCh <- "could not read artifact file: " + err.Error()
			return err
		}

		zipBuffer := bytes.Buffer{}
		zipWriter := zip.NewWriter(&zipBuffer)

		zipFile, err := zipWriter.Create(artifact.File)
		if err != nil {
			bs.MessageCh <- "could not create artifact file in zip archive: " + err.Error()
			return err
		}
		_, err = zipFile.Write(artifactContent)
		if err != nil {
			bs.MessageCh <- "could not write artifact file to zip archive: " + err.Error()
			return err
		}
		_ = zipWriter.Close()

		zipArchiveName := artifact.File + ".zip"
		err = ioutil.WriteFile(zipArchiveName, zipBuffer.Bytes(), 0744)
		if err != nil {
			bs.MessageCh <- "could not write zip archive bytes to file: " + err.Error()
			return err
		}

		for _, deployment := range cont.Deployments.EmailDeployments {
			eg.Go(func() error {
				if err := deploymentservice.DoEmailDeployment(&deployment, "artifact", settings, zipArchiveName); err != nil {
					return err
				}
				return nil
			})
		}
	} else {
		bs.MessageCh <- "no email deployments found"
	}

	remoteDeploymentCount := len(cont.Deployments.RemoteDeployments)
	if remoteDeploymentCount > 0 {
		bs.MessageCh <- fmt.Sprintf("%d remote deployment(s) found", remoteDeploymentCount)
		for _, deployment := range cont.Deployments.RemoteDeployments {
			eg.Go(func() error {
				if err := deploymentservice.DoRemoteDeployment(&deployment, artifact); err != nil {
					return err
				}
				return nil
			})
		}
	} else {
		bs.MessageCh <- "no remote deplyoments found"
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (bs *BuildService) getRepositoryUrl(cont entity.BuildDefinitionContent, withCredentials bool) (string, error) {
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

// CheckPayloadHeader checks the existence and values taken from HTTP request headers
// from the given HTTP request
func CheckPayloadHeader(content entity.BuildDefinitionContent, r *http.Request) error {
	var err error
	// check relevant headers and payload values
	switch content.Repository.Hoster {
	case "bitbucket":
		headers := []string{"X-Event-Key", "X-Hook-UUID", "X-Request-UUID", "X-Attempt-Number"}
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

func splitCommand(input string) ([]string, error) {
	return shellquote.Split(input)
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
