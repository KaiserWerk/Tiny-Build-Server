package buildsteps

import "github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

// The PhpBuildDefinition
type PhpBuildDefinition struct {
	CloneDir    string
	ArtifactDir string
	MetaData    entity.BuildDefinition
	Content     entity.BuildDefinitionContent
}
