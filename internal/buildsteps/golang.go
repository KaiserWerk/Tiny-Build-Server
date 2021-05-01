package buildsteps

import (
	"errors"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/stvp/slug"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// The GolangBuildDefinition
type GolangBuildDefinition struct {
	CloneDir    string
	ArtifactDir string
	MetaData    entity.BuildDefinition
	Content     entity.BuildDefinitionContent
}

// RunTests runs the unit tests
func (bd GolangBuildDefinition) RunTests(messageCh chan string) error {
	cmd := exec.Command("go", "test", "-race", "./...")
	cmd.Dir = bd.CloneDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New("could not run unit tests: " + err.Error())
	}

	messageCh <- "unit test result:\n" + string(output)
	return nil
}

// RunBenchmarkTests runs the benchmark tests
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

// BuildArtifact builds an artifact
func (bd GolangBuildDefinition) BuildArtifact(messageCh chan string, projectDir string) (string, error) {
	var err error
	slug.Replacement = '-' // TODO: move elsewhere
	binaryName := slug.Clean(strings.ToLower(strings.Split(bd.Content.Repository.Name, "/")[1]))

	// pre build steps
	for _, preBuildStep := range bd.Content.PreBuild {
		// set an environment variable
		if strings.Contains(preBuildStep, "setenv") && strings.Contains(preBuildStep, "=") {
			setenv := strings.Replace(preBuildStep, "setenv ", "", 1)
			parts := strings.Split(setenv, "=")
			if len(parts) != 2 {
				messageCh <- "incorrect setenv syntax (" + preBuildStep + ")"
				continue
			}
			err = os.Setenv(parts[0], parts[1])
			if err != nil {
				messageCh <- fmt.Sprintf("preBuildStep failed: could not set env var %s to value %s", parts[0], parts[1])
				continue
			}
			messageCh <- fmt.Sprintf("preBuildStep executed: setting env var '%s' to value '%s'", parts[0], parts[1])
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
	goos := os.Getenv("GOOS")
	if goos == "" {
		goos = runtime.GOOS
	}
	if strings.Contains(strings.ToLower(goos), "win") {
		binaryName += ".exe"
	}
	messageCh <- "binary name set to " + binaryName
	artifact := bd.ArtifactDir + "/" + binaryName

	// actual build steps
	for _, buildStep := range bd.Content.Build {
		if strings.Contains(buildStep, "setenv") && strings.Contains(buildStep, "=") {
			setenv := strings.Replace(buildStep, "setenv ", "", 1)
			parts := strings.Split(setenv, "=")
			if len(parts) != 2 {
				messageCh <- "incorrect setenv syntax (" + buildStep + ")"
				continue
			}
			err = os.Setenv(parts[0], parts[1])
			if err != nil {
				messageCh <- fmt.Sprintf("preBuildStep failed: could not set env var %s to value %s", parts[0], parts[1])
				continue
			}
			messageCh <- fmt.Sprintf("preBuildStep executed: setting env var '%s' to value '%s'", parts[0], parts[1])
		} else if strings.Contains(buildStep, "build") {
			if strings.Contains(buildStep, " ") {
				stepParts := strings.Split(buildStep, " ")
				if len(stepParts) != 2 {
					messageCh <- fmt.Sprintf("build step '%s' has incorrect build syntax; skipped", buildStep)
					continue
				}
				bd.CloneDir = filepath.Join(bd.CloneDir, strings.TrimLeft(stepParts[1], "./"))
			} else {
				bd.CloneDir = filepath.Join(bd.CloneDir, "main.go")
			}

			buildCommand := []string{
				"go",
				"build",
				"-o",
				artifact,
				//"-a",
				//"-v",
				//"-work",
				//"-x",
				`-ldflags`,
				`-s -w`,
				bd.CloneDir,
			}

			cmd := exec.Command(buildCommand[0], buildCommand[1:]...)
			messageCh <- fmt.Sprintf("build command to be executed: %s", cmd.String())
			err := cmd.Run()
			if err != nil {
				messageCh <- fmt.Sprintf("build step '%s': failed because %s", buildStep, err.Error())
				return "", err
			}
			messageCh <- fmt.Sprintf("build step '%s' successful", buildStep)
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
