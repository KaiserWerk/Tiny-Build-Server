package service

import (
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildsteps"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/pkg/sftp"
	"github.com/stvp/slug"
	"golang.org/x/crypto/ssh"
	"io"
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
	if helper.FileExists(projectPath) {
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

	settings, err := helper.GetAllSettings()
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

	switch definition.BuildTarget {
	case "golang": // golang
		def := buildsteps.GolangBuildDefinition{
			CloneDir:        strings.ToLower(fmt.Sprintf("%s/%s/%s/clone", baseDataPath, definition.RepoHoster, slug.Clean(definition.RepoFullname))),
			ArtifactDir:     strings.ToLower(fmt.Sprintf("%s/%s/%s/artifact", baseDataPath, definition.RepoHoster, slug.Clean(definition.RepoFullname))),
			BuildDefinition: definition,
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
	case "dotnet": // dotnet
		def := buildsteps.DotnetBuildDefinition{
			CloneDir:        strings.ToLower(fmt.Sprintf("%s/%s/%s/clone", baseDataPath, definition.RepoHoster, slug.Clean(definition.RepoFullname))),
			ArtifactDir:     strings.ToLower(fmt.Sprintf("%s/%s/%s/artifact", baseDataPath, definition.RepoHoster, slug.Clean(definition.RepoFullname))),
			BuildDefinition: definition,
		}
		err = handleDotnetProject(def, messageCh, projectPath)
		if err != nil {
			messageCh <- "could not build dotnet project (" + projectPath + ")"
			saveBuildReport(definition, sb.String())
			return
		}
	case "php": // php
		def := buildsteps.PhpBuildDefinition{
			CloneDir:        strings.ToLower(fmt.Sprintf("%s/%s/%s/clone", baseDataPath, definition.RepoHoster, slug.Clean(definition.RepoFullname))),
			ArtifactDir:     strings.ToLower(fmt.Sprintf("%s/%s/%s/artifact", baseDataPath, definition.RepoHoster, slug.Clean(definition.RepoFullname))),
			BuildDefinition: definition,
		}
		err = handlePhpProject(def, messageCh, projectPath)
		if err != nil {
			messageCh <- "could not build php project (" + projectPath + ")"
			saveBuildReport(definition, sb.String())
			return
		}
	case "rust": // rust
		def := buildsteps.RustBuildDefinition{
			CloneDir:        strings.ToLower(fmt.Sprintf("%s/%s/%s/clone", baseDataPath, definition.RepoHoster, slug.Clean(definition.RepoFullname))),
			ArtifactDir:     strings.ToLower(fmt.Sprintf("%s/%s/%s/artifact", baseDataPath, definition.RepoHoster, slug.Clean(definition.RepoFullname))),
			BuildDefinition: definition,
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

	if definition.ApplyMigrations {
		//metaMigrationId := definition.MetaMigrationId

		//err = definition.applyMigrations(messageCh)
		//if err != nil {
		//	messageCh <- "process cancelled by migration application: " + err.Error()
		//	return err
		//}
	}

	db := helper.GetDbConnection()
	var depList []entity.DeploymentDefinition
	var dep entity.DeploymentDefinition
	rows, err := db.Query("SELECT id, build_definition_id, caption, host, username, password, connection_type, "+
		"working_directory, pre_deployment_actions, post_deployment_actions FROM deployment_definition "+
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
		dep = entity.DeploymentDefinition{}
	}

	if len(depList) > 0 && definition.DeploymentEnabled {
		err = deployArtifact(depList, messageCh, artifact)
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

func deployArtifact(deploymentDefinitions []entity.DeploymentDefinition, messageCh chan string, artifact string) error {
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
			preDeploymentActions = strings.Split(deployment.PreDeploymentActions, "\n")
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
			postDeploymentActions = strings.Split(deployment.PostDeploymentActions, "\n")
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

func getRepositoryUrl(bd entity.BuildDefinition, withCredentials bool) string {
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
