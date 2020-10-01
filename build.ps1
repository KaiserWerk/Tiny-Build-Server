$sourcecode = "."
$target = "build/tiny-build-server"
$version = "1.0.0"
$date = Get-Date -Format "yyyy-MM-dd HH:mm:ss K"
$env:GOOS = 'windows'; $env:GOARCH = 'amd64';             go build -o "$($target)-win64.exe" -ldflags "-s -w -X 'main.version=$($version)' -X 'main.versionDate=$($date)'" $sourcecode
$env:GOOS = 'linux';   $env:GOARCH = 'amd64';             go build -o "$($target)-linux64"   -ldflags "-s -w -X 'main.version=$($version)' -X 'main.versionDate=$($date)'" $sourcecode
$env:GOOS = 'linux';   $env:GOARCH = 'arm'; $env:GOARM=5; go build -o "$($target)-raspi32"   -ldflags "-s -w -X 'main.version=$($version)' -X 'main.versionDate=$($date)'" $sourcecode
$env:GOOS = 'darwin';  $env:GOARCH = 'amd64';             go build -o "$($target)-macos64"   -ldflags "-s -w -X 'main.version=$($version)' -X 'main.versionDate=$($date)'" $sourcecode
# git rev-parse --verify HEAD