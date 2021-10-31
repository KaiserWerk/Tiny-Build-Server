#!/bin/bash
source="cmd/tiny-build-server/main.go"
target="build/tiny-build-server"
version="0.0.0-dev-linux"
versionDate=$(date +"%Y-%m-%d %T")
export GOOS=windows; export GOARCH=amd64; go build -o "$target-win64.exe" -ldflags "-s -w -X 'main.Version=$version' -X 'main.VersionDate=$versionDate'" $source