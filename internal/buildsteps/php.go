package buildsteps

import "github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

// The PhpBuildDefinition
type PhpBuildDefinition struct {
	CloneDir    string
	ArtifactDir string
	MetaData    entity.BuildDefinition
	Content     entity.BuildDefinitionContent
}

func (p PhpBuildDefinition) RunTests(messageCh chan string) error {
	panic("implement me")
}

func (p PhpBuildDefinition) RunBenchmarkTests(messageCh chan string) error {
	panic("implement me")
}

func (p PhpBuildDefinition) BuildArtifact(messageCh chan string) (string, error) {
	panic("implement me")
}

