package buildsteps

import "github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

// The DotnetBuildDefinition
type DotnetBuildDefinition struct {
	CloneDir    string
	ArtifactDir string
	MetaData    entity.BuildDefinition
	Content     entity.BuildDefinitionContent
}

func (bd DotnetBuildDefinition) RunTests(messageCh chan string) error {
	panic("implement me")
}

func (bd DotnetBuildDefinition) RunBenchmarkTests(messageCh chan string) error {
	panic("implement me")
}

func (bd DotnetBuildDefinition) BuildArtifact(messageCh chan string) (string, error) {
	panic("implement me")
}

