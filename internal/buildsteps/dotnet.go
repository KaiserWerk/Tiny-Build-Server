package buildsteps

import "github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

// The DotnetBuildDefinition
type DotnetBuildDefinition struct {
	CloneDir    string
	ArtifactDir string
	MetaData    entity.BuildDefinition
	Content     entity.BuildDefinitionContent
}

func (bd DotnetBuildDefinition) Initialize() {

}

func (bd DotnetBuildDefinition) FetchSource() {

}

func (bd DotnetBuildDefinition) RunTests() {

}

func (bd DotnetBuildDefinition) BuildArtifact() {

}

func (bd DotnetBuildDefinition) ApplyMigrations() {

}

func (bd DotnetBuildDefinition) CreateReport() {

}

func (bd DotnetBuildDefinition) CleanUp() {

}
