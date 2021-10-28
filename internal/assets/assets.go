package assets

import "embed"

//go:embed templates/*
var templateFS embed.FS

//go:embed webasset
var webAssetFS embed.FS

//go:embed misc
var miscFS embed.FS

func GetTemplate(name string) ([]byte, error) {
	return templateFS.ReadFile("templates/" + name)
}

func GetWebAssetFile(name string) ([]byte, error) {
	return webAssetFS.ReadFile("webasset/" + name)
}

func GetMiscFile(name string) ([]byte, error) {
	return miscFS.ReadFile("misc/" + name)
}