$sourcecode = "cmd/tiny-build-server/main.go"
$target = "build/tiny-build-server"
$version = "0.0.0-dev"
$date = Get-Date -Format "yyyy-MM-dd HH:mm:ss K"
# Windows, 64-bit
$env:GOOS = 'windows'; $env:GOARCH = 'amd64';             go build -o "$($target)-win-amd64.exe" -ldflags "-s -w -X 'main.version=$($version)' -X 'main.versionDate=$($date)'" $sourcecode
# Linux, 64-bit
$env:GOOS = 'linux';   $env:GOARCH = 'amd64';             go build -o "$($target)-linux-amd64"   -ldflags "-s -w -X 'main.version=$($version)' -X 'main.versionDate=$($date)'" $sourcecode
# Raspberry Pi
$env:GOOS = 'linux';   $env:GOARCH = 'arm'; $env:GOARM=5; go build -o "$($target)-raspi32"       -ldflags "-s -w -X 'main.version=$($version)' -X 'main.versionDate=$($date)'" $sourcecode
# older macOS
$env:GOOS = 'darwin';  $env:GOARCH = 'amd64';             go build -o "$($target)-macos-amd64"   -ldflags "-s -w -X 'main.version=$($version)' -X 'main.versionDate=$($date)'" $sourcecode
# newer macOS
$env:GOOS = 'darwin';  $env:GOARCH = 'arm64';             go build -o "$($target)-macos-arm64"   -ldflags "-s -w -X 'main.version=$($version)' -X 'main.versionDate=$($date)'" $sourcecode
