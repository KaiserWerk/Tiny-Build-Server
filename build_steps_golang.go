package main

import (
	"github.com/stvp/slug"
	"os/exec"
	"strings"
)

type golangBuildDefinition buildDefinition


func (bd golangBuildDefinition) runTests(messageCh chan string) error {

	return nil
}

func (bd golangBuildDefinition) runBenchmarkTests(messageCh chan string) error {

	return nil
}

func (bd golangBuildDefinition) buildArtifact(messageCh chan string, projectDir string) error {
	cloneDir := projectDir + "/clone"
	buildDir := projectDir + "/build"
	artifactDir := projectDir + "/artifact"

	slug.Replacement = '-'
	binaryName := slug.Clean(strings.ToLower(strings.Split(bd.RepoFullname, "/")[1]))
	if strings.Contains(bd.BuildTargetOsArch, "win") {
		binaryName += ".exe"
	}
	messageCh <- "binary name set to " + binaryName
	cmd := exec.Command("")

	// Command
	// go build -o <output> -a -v -work -x -ldflags "-s -w -X main.<Var>=<Val>" <input>


	return nil
}

func (bd golangBuildDefinition) applyMigrations(messageCh chan string) error {

	return nil
}
