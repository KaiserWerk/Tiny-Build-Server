run:
	go run cmd/tiny-build-server/main.go
test:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
build:
	setx GOOS "linux"
	setx GOARCH "amd64"
	go build -o build/tbs-linux-x64 cmd/tiny-build-server/main.go
	setx GOOS "windows"
	setx GOARCH "amd64"
	go build -o build/tbs-windows-x64.exe cmd/tiny-build-server/main.go
	setx GOOS "linux"
	setx GOARCH "arm"
	setx GOARM "5"
	go build -o build/tbs-raspi cmd/tiny-build-server/main.go
