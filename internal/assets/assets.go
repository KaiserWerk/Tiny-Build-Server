package assets

import (
	"embed"
)

//go:embed config
var configFS embed.FS

//go:embed templates/*
var templateFS embed.FS

//go:embed assets
var webAssetFS embed.FS

//go:embed misc
var miscFS embed.FS

func GetConfig(name string) ([]byte, error) {
	return configFS.ReadFile("config/" + name)
}

func GetTemplate(name string) ([]byte, error) {
	return templateFS.ReadFile("templates/" + name)
}

func GetWebAssetFile(name string) ([]byte, error) {
	return webAssetFS.ReadFile("assets/" + name)
}

func GetWebAssetFS() embed.FS {
	return webAssetFS
}

func GetMiscFile(name string) ([]byte, error) {
	return miscFS.ReadFile("misc/" + name)
}
