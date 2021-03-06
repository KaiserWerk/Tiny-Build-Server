[![Go Report Card](https://goreportcard.com/badge/github.com/KaiserWerk/Tiny-Build-Server)](https://goreportcard.com/report/github.com/KaiserWerk/Tiny-Build-Server)

# Tiny Build Server

**IMPORTANT: This is work in progress!**

This project aims to be a functioning, minimal build server for Go, .NET, Rust and (probably) PHP projects.

It can be used in conjunction with *BitBucket*, *GitHub*, *GitLab* and *Gitea* and it runs on 
whatever platform you compile it for, be it *Windows*, *Linux* or *macOS*, even on a RaspberryPi.
Built artifacts can be deployed via SFTP (FTP via SSH) and via mail.
Releases will deliver standalone binaries for the popular operating systems. If you need
a build for another OS/ARCH, refer to section __Custom Build__ below.

### License

Free to use for any non-commercial purpose; refer to LICENSE.md

### Dependencies

* Git (for cloning projects)
* [go](https://golang.org/) >= 1.16 (for Golang projects)
* cargo through [rustup](https://www.rust-lang.org/) (for Rust projects)
* [dotnet](https://dotnet.microsoft.com/download) (for .NET projects)

### Documentation

* [Custom build](docs/custom-build.md)
* [Installation](docs/installation.md)
* [Create a build definition](docs/create-a-build-definition.md)
* [Create a webhook](docs/create-a-webhook.md)
* [Admin Settings](docs/admin-settings.md)

