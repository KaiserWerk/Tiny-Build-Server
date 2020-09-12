$sourcecode = "."
$target = "build/tiny-build-server"
$env:GOOS = 'windows'; $env:GOARCH = 'amd64';               go build -o "$($target)-win64.exe" $sourcecode
$env:GOOS = 'linux';   $env:GOARCH = 'amd64';               go build -o "$($target)-linux64" $sourcecode
$env:GOOS = 'linux';   $env:GOARCH = 'arm'; $env:GOARM=5;   go build -o "$($target)-raspi32" $sourcecode
$env:GOOS = 'darwin';  $env:GOARCH = 'amd64';               go build -o "$($target)-macos64.macos" $sourcecode
# -ldflags "-X main.version=0.0.1,main.commitHash=123abc"