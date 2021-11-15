package entity

type Pipeline struct {
	CloneDir    string
	ArtifactDir string
	MetaData    BuildDefinition
	Content     BuildDefinitionContent
}
