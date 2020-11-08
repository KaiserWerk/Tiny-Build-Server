package buildsteps

import "Tiny-Build-Server/internal/entity"

type DotnetBuildDefinition struct {
	CloneDir    string
	ArtifactDir string
	entity.BuildDefinition
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
