package buildsteps

import (
	"errors"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/stvp/slug"
	"os"
	"os/exec"
	"strings"
)

type GolangBuildDefinition struct {
	CloneDir    string
	ArtifactDir string
	MetaData entity.BuildDefinition
	Content entity.BuildDefinitionContent
}

func (bd GolangBuildDefinition) RunTests(messageCh chan string) error {
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = bd.CloneDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New("could not run unit tests: " + err.Error())
	}

	messageCh <- "unit test result:\n" + string(output)
	return nil
}

func (bd GolangBuildDefinition) RunBenchmarkTests(messageCh chan string) error {
	cmd := exec.Command("go", "test", "-bench=.")
	cmd.Dir = bd.CloneDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New("could not run benchmark tests: " + err.Error())
	}

	messageCh <- "benchmark test result:\n" + string(output)
	return nil
}

func (bd GolangBuildDefinition) BuildArtifact(messageCh chan string, projectDir string) (string, error) {
	var err error
	slug.Replacement = '-' // TODO: move elsewhere
	binaryName := slug.Clean(strings.ToLower(strings.Split(bd.Content.Repository.Name, "/")[1]))

	// TODO: remove clone directory?

	// pre build steps
	for _, preBuildStep := range bd.Content.PreBuild {
		// set an environment variable
		if strings.Contains(preBuildStep, "setenv") && strings.Contains(preBuildStep, "=") {
			setenv := strings.Replace(preBuildStep, "setenv ", "", 1)
			parts := strings.Split(setenv, "=")
			if len(parts) != 2 {
				continue
			}
			err = os.Setenv(parts[0], parts[1])
			if err != nil {
				messageCh <- fmt.Sprintf("preBuildStep failed: could not set envvar %s to value %s", parts[0], parts[1])
				continue
			}
			messageCh <- fmt.Sprintf("preBuildStep executed: setting envvar '%s' to value '%s'", parts[0], parts[1])
		} else {
			parts := strings.Split(preBuildStep, " ")
			cmd := exec.Command(parts[0], parts[1:]...)
			cmd.Dir = bd.CloneDir
			output, err := cmd.CombinedOutput()
			if err != nil {
				messageCh <- fmt.Sprintf("preBuildStep failed: %s", err.Error())
				continue
			}
			messageCh <- fmt.Sprintf("preBuildStep executed: %s", string(output))
		}
	}

	// if it's windows, append .exe to the binary file
	if strings.Contains(strings.ToLower(os.Getenv("GOOS")), "win") {
		binaryName += ".exe"
	}
	messageCh <- "binary name set to " + binaryName
	artifact := bd.ArtifactDir + "/" + binaryName

	// actual build steps
	for _, buildStep := range bd.Content.Build {
		if buildStep == "go build" {
			buildCommand := fmt.Sprintf(
				`go build -o %s -a -v -work -x -ldflags "-s -w -X" %s`,
				artifact,
				bd.CloneDir,
			)
			parts := strings.Split(buildCommand, " ")
			cmd := exec.Command(parts[0], parts[1:]...)
			messageCh <- fmt.Sprintf("build command to be executed: '%s'", cmd.String())
			result, err := cmd.CombinedOutput()
			if err != nil {
				messageCh <- "build failed: " + err.Error()
				return "", err
			}
			messageCh <- "build successful: " + string(result)
		} else {
			parts := strings.Split(buildStep, " ")
			cmd := exec.Command(parts[0], parts[1:]...)
			result, err := cmd.CombinedOutput()
			if err != nil {
				messageCh <- "build with custom command failed: " + err.Error()
				return "", err
			}
			messageCh <- "build with custom command successful: " + string(result)
		}
	}

	// post build steps
	for _, postBuildStep := range bd.Content.PostBuild {
		// set an environment variable
		if strings.Contains(postBuildStep, "setenv") && strings.Contains(postBuildStep, "=") {
			setenv := strings.Replace(postBuildStep, "setenv ", "", 1)
			parts := strings.Split(setenv, "=")
			if len(parts) != 2 {
				continue
			}
			err = os.Setenv(parts[0], parts[1])
			if err != nil {
				messageCh <- fmt.Sprintf("postBuildStep failed: could not set envvar %s to value %s", parts[0], parts[1])
				continue
			}
			messageCh <- fmt.Sprintf("postBuildStep executed: setting envvar '%s' to value '%s'", parts[0], parts[1])
		} else {
			parts := strings.Split(postBuildStep, " ")
			cmd := exec.Command(parts[0], parts[1:]...)
			cmd.Dir = bd.CloneDir
			output, err := cmd.CombinedOutput()
			if err != nil {
				messageCh <- fmt.Sprintf("postBuildStep failed: %s", err.Error())
				continue
			}
			messageCh <- fmt.Sprintf("postBuildStep executed: %s", string(output))
		}
	}

	return artifact, nil
}
