package buildsteps

import "github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

// The RustBuildDefinition
type RustBuildDefinition struct {
	CloneDir    string
	ArtifactDir string
	MetaData    entity.BuildDefinition
	Content     entity.BuildDefinitionContent
}

func (r RustBuildDefinition) RunTests(messageCh chan string) error {
	panic("implement me")
}

func (r RustBuildDefinition) RunBenchmarkTests(messageCh chan string) error {
	panic("implement me")
}

func (r RustBuildDefinition) BuildArtifact(messageCh chan string) (string, error) {
	panic("implement me")
}

