package buildsteps

import "github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

type RustBuildDefinition struct {
	CloneDir    string
	ArtifactDir string
	entity.BuildDefinition
}
