$sourcecode = "cmd/tiny-build-server/main.go"
$target = "build/tiny-build-server"
$version = "0.0.4-alpha"
$date = Get-Date -Format "yyyy-MM-dd HH:mm:ss K"

# Linux, 64-bit
$env:GOOS = 'linux';   $env:GOARCH = 'amd64';             go build -o "$($target)-v$($version)-linux-amd64"   -ldflags "-s -w -X 'main.Version=$($version)' -X 'main.VersionDate=$($date)'" $sourcecode
$env:GOOS = 'linux';   $env:GOARCH = '386';               go build -o "$($target)-v$($version)-linux-386"     -ldflags "-s -w -X 'main.Version=$($version)' -X 'main.VersionDate=$($date)'" $sourcecode
# Raspberry Pi
$env:GOOS = 'linux';   $env:GOARCH = 'arm'; $env:GOARM=5; go build -o "$($target)-v$($version)-linux-armv5"   -ldflags "-s -w -X 'main.Version=$($version)' -X 'main.VersionDate=$($date)'" $sourcecode
$env:GOOS = 'linux';   $env:GOARCH = 'arm'; $env:GOARM=6; go build -o "$($target)-v$($version)-linux-armv6"   -ldflags "-s -w -X 'main.Version=$($version)' -X 'main.VersionDate=$($date)'" $sourcecode
$env:GOOS = 'linux';   $env:GOARCH = 'arm'; $env:GOARM=7; go build -o "$($target)-v$($version)-linux-armv7"   -ldflags "-s -w -X 'main.Version=$($version)' -X 'main.VersionDate=$($date)'" $sourcecode
# macOS
$env:GOOS = 'darwin';  $env:GOARCH = 'amd64';             go build -o "$($target)-v$($version)-macos-amd64"   -ldflags "-s -w -X 'main.Version=$($version)' -X 'main.VersionDate=$($date)'" $sourcecode
$env:GOOS = 'darwin';  $env:GOARCH = 'arm64';             go build -o "$($target)-v$($version)-macos-arm"     -ldflags "-s -w -X 'main.Version=$($version)' -X 'main.VersionDate=$($date)'" $sourcecode
# Windows, 64-bit
$env:GOOS = 'windows'; $env:GOARCH = 'amd64';             go build -o "$($target)-v$($version)-win-amd64.exe" -ldflags "-s -w -X 'main.Version=$($version)' -X 'main.VersionDate=$($date)'" $sourcecode