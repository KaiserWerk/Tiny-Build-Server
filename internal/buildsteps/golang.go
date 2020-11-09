package buildsteps

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"errors"
	"fmt"
	"github.com/stvp/slug"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type GolangBuildDefinition struct {
	CloneDir    string
	ArtifactDir string
	entity.BuildDefinition
}

func (bd GolangBuildDefinition) RunTests(messageCh chan string) error {
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = bd.CloneDir

	output, err := cmd.Output()
	if err != nil {
		return errors.New("could not run unit tests: " + err.Error())
	}

	messageCh <- "unit test result:\n" + string(output)
	return nil
}

func (bd GolangBuildDefinition) RunBenchmarkTests(messageCh chan string) error {
	cmd := exec.Command("go", "test", "-bench=.")
	cmd.Dir = bd.CloneDir

	output, err := cmd.Output()
	if err != nil {
		return errors.New("could not run benchmark tests: " + err.Error())
	}

	messageCh <- "benchmark test result:\n" + string(output)
	return nil
}

func (bd GolangBuildDefinition) BuildArtifact(messageCh chan string, projectDir string) (string, error) {
	var err error
	slug.Replacement = '-'
	binaryName := slug.Clean(strings.ToLower(strings.Split(bd.RepoFullname, "/")[1]))
	if strings.Contains(bd.BuildTargetOsArch, "win") {
		binaryName += ".exe"
	}
	messageCh <- "binary name set to " + binaryName

	artifact := bd.ArtifactDir + "/" + binaryName

	buildCommand := fmt.Sprintf(
		`build -o %s -mod=vendor -a -v -work -x -ldflags "-s -w -X main.versionDate=%s" %s`,
		artifact,
		time.Now().Format(time.RFC3339),
		bd.CloneDir,
	)

	// for go, the separator is a forward slash
	buildTargetElements := strings.Split(bd.BuildTargetOsArch, "/")
	err = os.Setenv("GOOS", buildTargetElements[0])
	if err != nil {
		messageCh <- "could not set environment variable GOOS: " + err.Error()
		return "", err
	}
	err = os.Setenv("GOARCH", buildTargetElements[1])
	if err != nil {
		messageCh <- "could not set environment variable GOARCH: " + err.Error()
		return "", err
	}
	if bd.BuildTargetArm > 0 {
		err = os.Setenv("GOARM", strconv.Itoa(bd.BuildTargetArm))
		if err != nil {
			messageCh <- "could not set environment variable GOARM: " + err.Error()
			return "", err
		}
	}

	cmd := exec.Command("go", strings.Split(buildCommand, " ")...)
	messageCh <- "build command to be executed: '" + cmd.String() + "'"
	result, err := cmd.Output()
	if err != nil {
		messageCh <- "build was not successful: " + err.Error()
		return "", err
	}
	messageCh <- "build successful: " + string(result)

	return artifact, nil
}

