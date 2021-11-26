package entity

import "path/filepath"

type Artifact struct {
	Directory  string
	File       string
	Version    string
}

func (a Artifact) FullPath() string {
	return filepath.Join(a.Directory, a.File)
}