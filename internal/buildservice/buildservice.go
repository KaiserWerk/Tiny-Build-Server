package buildservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildsteps"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseService"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/pkg/sftp"
	"github.com/stvp/slug"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var basePath string = "data/"

func saveBuildReport(definition entity.BuildDefinition, report string) {
	// TODO
}

func StartBuildProcess(definition entity.BuildDefinition) {
	// instantiate tools for build output
	var sb strings.Builder
	messageCh := make(chan string, 100)
	go func() {
		for {
			s, ok := <-messageCh
			if ok {
				sb.WriteString(s)
			}
		}
	}()
	defer saveBuildReport(definition, sb.String())
	db := databaseService.New()
	defer db.Quit()

	// determine definition ID as string
	idString := strconv.Itoa(definition.Id)

	// set projectpath
	projectPath := basePath + idString + "/" + strconv.FormatInt(time.Now().Unix(), 10)
	buildPath := projectPath + "/build"
	artifactPath := projectPath + "/artifact"
	clonePath := projectPath + "/clone"
	if helper.FileExists(projectPath) {
		err := os.RemoveAll(projectPath)
		if err != nil {
			messageCh <- "could not remove stale project directory (" + projectPath + "): " + err.Error()
			return
		}
	}

	// create a new build directory
	err := os.MkdirAll(buildPath, 0664)
	if err != nil {
		messageCh <- "could not create build directory (" + buildPath + "): " + err.Error()
		return
	}
	err = os.MkdirAll(artifactPath, 0664)
	if err != nil {
		messageCh <- "could not create artifact directory (" + artifactPath + "): " + err.Error()
		return
	}
	err = os.MkdirAll(clonePath, 0664)
	if err != nil {
		messageCh <- "could not create clone directory (" + clonePath + "): " + err.Error()
		return
	}

	// parse the content
	var cont entity.BuildDefinitionContent
	err = yaml.Unmarshal([]byte(definition.Content), &cont)
	if err != nil {
		messageCh <- "could not unmarshal buidl definition content: " + err.Error()
		return
	}

	// clone the repository
	var withCredentials bool
	if cont.Repository.AccessSecret != "" {
		withCredentials = true
	}
	repositoryUrl := getRepositoryUrl(cont, withCredentials)
	cmd := exec.Command("git", "clone", "--single-branch", "--branch", cont.Repository.Branch, repositoryUrl, clonePath)
	messageCh <- cmd.String()
	cmdOutput, err := cmd.CombinedOutput()
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

	settings, err := db.GetAllSettings()
	if err != nil {
		messageCh <- "could not obtain setting values: " + err.Error()
		saveBuildReport(definition, sb.String())
		return
	}

	baseDataPath, ok := settings["base_datapath"]
	if !ok {
		messageCh <- "could not fetch base data path"
		saveBuildReport(definition, sb.String())
		return
	}

	switch cont.ProjectType {
	case "go":
		fallthrough
	case "golang":
		def := buildsteps.GolangBuildDefinition{
			CloneDir:    strings.ToLower(fmt.Sprintf("%s/%s/%s/clone", baseDataPath, cont.Repository.Hoster, slug.Clean(cont.Repository.Name))),
			ArtifactDir: strings.ToLower(fmt.Sprintf("%s/%s/%s/artifact", baseDataPath, cont.Repository.Hoster, slug.Clean(cont.Repository.Name))),
			MetaData:    definition,
			Content:     cont,
		}
		_, err = handleGolangProject(def, messageCh, projectPath)
		if err != nil {
			messageCh <- "could not build golang project (" + projectPath + ")"
			saveBuildReport(definition, sb.String())
			return
		} else {
			// set artifact

			// deploy artifact to any host
			//err = deployArtifact(definition, artifact)
		}
	case "dotnet":
		def := buildsteps.DotnetBuildDefinition{
			CloneDir:    strings.ToLower(fmt.Sprintf("%s/%s/%s/clone", baseDataPath, cont.Repository.Hoster, slug.Clean(cont.Repository.Name))),
			ArtifactDir: strings.ToLower(fmt.Sprintf("%s/%s/%s/artifact", baseDataPath, cont.Repository.Hoster, slug.Clean(cont.Repository.Name))),
			MetaData:    definition,
			Content:     cont,
		}
		err = handleDotnetProject(def, messageCh, projectPath)
		if err != nil {
			messageCh <- "could not build dotnet project (" + projectPath + ")"
			saveBuildReport(definition, sb.String())
			return
		}
	case "php":
		def := buildsteps.PhpBuildDefinition{
			CloneDir:    strings.ToLower(fmt.Sprintf("%s/%s/%s/clone", baseDataPath, cont.Repository.Hoster, slug.Clean(cont.Repository.Name))),
			ArtifactDir: strings.ToLower(fmt.Sprintf("%s/%s/%s/artifact", baseDataPath, cont.Repository.Hoster, slug.Clean(cont.Repository.Name))),
			MetaData:    definition,
			Content:     cont,
		}
		err = handlePhpProject(def, messageCh, projectPath)
		if err != nil {
			messageCh <- "could not build php project (" + projectPath + ")"
			saveBuildReport(definition, sb.String())
			return
		}
	case "rust":
		def := buildsteps.RustBuildDefinition{
			CloneDir:    strings.ToLower(fmt.Sprintf("%s/%s/%s/clone", baseDataPath, cont.Repository.Hoster, slug.Clean(cont.Repository.Name))),
			ArtifactDir: strings.ToLower(fmt.Sprintf("%s/%s/%s/artifact", baseDataPath, cont.Repository.Hoster, slug.Clean(cont.Repository.Name))),
			MetaData:    definition,
			Content:     cont,
		}
		err = handleRustProject(def, messageCh, projectPath)
		if err != nil {
			messageCh <- "could not build rust project (" + projectPath + ")"
			saveBuildReport(definition, sb.String())
			return
		}
	}

	saveBuildReport(definition, sb.String())
	fmt.Println("info received!") // set a proper response
}

func handleGolangProject(definition buildsteps.GolangBuildDefinition, messageCh chan string, projectDir string) (string, error) {
	var err error

	//if definition. {
	//	err = definition.RunTests(messageCh)
	//	if err != nil {
	//		//messageCh <- "process cancelled by test.html run: " + err.Error()
	//		return "", err
	//	}
	//}
	//
	//if definition.RunBenchmarkTests {
	//	err = definition.RunBenchmarkTests(messageCh)
	//	if err != nil {
	//		//messageCh <- "process cancelled by benchmark test.html run: " + err.Error()
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
	// TODO email deployments!
	for _, deployment := range cont.Deployments.RemoteDeployments {
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

func getRepositoryUrl(cont entity.BuildDefinitionContent, withCredentials bool) string {
	var url string

	switch cont.Repository.Hoster {
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
		url = cont.Repository.HosterURL + "/" + cont.Repository.Name
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", cont.Repository.AccessUser, cont.Repository.AccessSecret, url)
		}
		return "https://" + url

	default:
		url = cont.Repository.HosterURL + "/" + cont.Repository.Name
		if withCredentials {
			url = fmt.Sprintf("%s:%s@%s", cont.Repository.AccessUser, cont.Repository.AccessSecret, url)
		}
		return "http://" + url // just http
	}
}

func CheckPayloadRequest(r *http.Request) (entity.BuildDefinition, error) {
	buildDefinition := entity.BuildDefinition{}

	// get id
	token := r.URL.Query().Get("token")
	if token == "" {
		return buildDefinition, fmt.Errorf("could not determine token")
	}

	// get DB connection
	ds := databaseService.New()
	buildDefinition, err := ds.FindBuildDefinition("token = ?", token)
	if err != nil {
		return buildDefinition, fmt.Errorf("build definition cannot be found in database for token %s", token)
	}

	var cont entity.BuildDefinitionContent
	err = yaml.Unmarshal([]byte(buildDefinition.Content), &cont)
	if err != nil {
		return entity.BuildDefinition{}, fmt.Errorf("could not unmarshal build definition content yaml for token %s", token)
	}

	// check relevant headers and payload values
	switch cont.Repository.Hoster {
	case "bitbucket":
		headers := []string{"X-Event-RegToken", "X-Hook-UUID", "X-Request-UUID", "X-Attempt-Number"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = helper.GetHeaderIfSet(r, headers[i])
			if err != nil {
				return entity.BuildDefinition{}, errors.New("bitbucket: could not get bitbucket header " + headers[i])
			}
		}

		var payload entity.BitBucketPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		_ = r.Body.Close()
		if err != nil {
			return entity.BuildDefinition{}, errors.New("bitbucket: could not decode json payload")
		}
		if payload.Push.Changes[0].New.Name != cont.Repository.Branch {
			return entity.BuildDefinition{}, errors.New("bitbucket: branch names do not match (" + payload.Push.Changes[0].New.Name + ")")
		}
		if payload.Repository.FullName != cont.Repository.Name {
			return entity.BuildDefinition{}, errors.New("bitbucket: repository names do not match (" + payload.Repository.FullName + ")")
		}
	case "github":
		headers := []string{"X-GitHub-Delivery", "X-GitHub-Event", "X-Hub-Signature"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = helper.GetHeaderIfSet(r, headers[i])
			if err != nil {
				return entity.BuildDefinition{}, errors.New("github: could not get github header " + headers[i])
			}
		}

		var payload entity.GitHubPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		_ = r.Body.Close()
		if err != nil {
			return entity.BuildDefinition{}, errors.New("github: could not decode json payload")
		}
		if payload.Repository.DefaultBranch != cont.Repository.Branch {
			return entity.BuildDefinition{}, errors.New("github: branch names do not match (" + payload.Repository.DefaultBranch + ")")
		}
		if payload.Repository.FullName != cont.Repository.Name {
			return entity.BuildDefinition{}, errors.New("github: repository names do not match (" + payload.Repository.FullName + ")")
		}
	case "gitlab":
		headers := []string{"X-GitLab-Event"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = helper.GetHeaderIfSet(r, headers[i])
			if err != nil {
				return entity.BuildDefinition{}, errors.New("gitlab: could not get gitlab header " + headers[i])
			}
		}

		var payload entity.GitLabPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		_ = r.Body.Close()
		if err != nil {
			return entity.BuildDefinition{}, errors.New("gitlab: could not decode json payload")
		}
		branch := strings.Split(payload.Ref, "/")[2]
		if branch != cont.Repository.Branch {
			return entity.BuildDefinition{}, errors.New("gitlab: branch names do not match (" + branch + ")")
		}
		if payload.Project.PathWithNamespace != cont.Repository.Name {
			return entity.BuildDefinition{}, errors.New("gitlab: repository names do not match (" + payload.Project.PathWithNamespace + ")")
		}
	case "gitea":
		headers := []string{"X-Gitea-Delivery", "X-Gitea-Event"}
		headerValues := make([]string, len(headers))
		for i := range headers {
			headerValues[i], err = helper.GetHeaderIfSet(r, headers[i])
			if err != nil {
				return entity.BuildDefinition{}, errors.New("gitea: could not get gitea header " + headers[i])
			}
		}

		var payload entity.GiteaPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		_ = r.Body.Close()
		if err != nil {
			return entity.BuildDefinition{}, errors.New("gitea: could not decode json payload")
		}
		branch := strings.Split(payload.Ref, "/")[2]
		if branch != cont.Repository.Branch {
			return entity.BuildDefinition{}, errors.New("gitea: branch names do not match (" + branch + ")")
		}
		if payload.Repository.FullName != cont.Repository.Name {
			return entity.BuildDefinition{}, errors.New("gitea: repository names do not match (" + payload.Repository.FullName + ")")
		}
	}

	return buildDefinition, nil
}
