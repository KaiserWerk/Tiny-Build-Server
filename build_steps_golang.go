package main

import (
	"github.com/stvp/slug"
	"os/exec"
	"runtime"
	"strings"
)

type golangBuildDefinition buildDefinition

func (bd golangBuildDefinition) runTests(messageCh chan string) error {

	return nil
}

func (bd golangBuildDefinition) runBenchmarkTests(messageCh chan string) error {

	return nil
}

func (bd golangBuildDefinition) buildArtifact(messageCh chan string, projectDir string) (string, error) {
	cloneDir := projectDir + "/clone"
	artifactDir := projectDir + "/artifact"

	slug.Replacement = '-'
	binaryName := slug.Clean(strings.ToLower(strings.Split(bd.RepoFullname, "/")[1]))
	if strings.Contains(bd.BuildTargetOsArch, "win") {
		binaryName += ".exe"
	}
	messageCh <- "binary name set to " + binaryName

	// format all go files?

	var dateStr string
	if runtime.GOOS == "windows" {
		dateStr = `Get-Date -Format "yyyy-MM-dd HH:mm:ss K"`
	} else {
		dateStr = `date -u +"%Y-%m-%d %H:%M:%S %:z"`
	}
	artifact := artifactDir + "/" + binaryName
	// set env vars!!!!!!!!!!!!

	// !!!
	buildCommand := `build -o "<output>" -mod=vendor -a -v -work -x -ldflags "-s -w -X main.versionDate=` + dateStr + `" <input>`
	buildCommand = strings.Replace(buildCommand, "<output>", artifact, 1)
	buildCommand = strings.Replace(buildCommand, "<input>", cloneDir, 1)

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
