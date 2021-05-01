package buildsteps

import "github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

// The RustBuildDefinition
type RustBuildDefinition struct {
	CloneDir    string
	ArtifactDir string
	MetaData    entity.BuildDefinition
	Content     entity.BuildDefinitionContent
}
