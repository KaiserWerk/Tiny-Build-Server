package buildsteps

import "Tiny-Build-Server/internal/entity"

type PhpBuildDefinition struct {
	CloneDir    string
	ArtifactDir string
	entity.BuildDefinition
}