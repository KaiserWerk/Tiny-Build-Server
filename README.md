[![Go Report Card](https://goreportcard.com/badge/github.com/KaiserWerk/Tiny-Build-Server)](https://goreportcard.com/report/github.com/KaiserWerk/Tiny-Build-Server)

# Tiny Build Server

This project aims to be a functioning, minimal build server for Golang and .NET projects (possibly PHP and Rust, too).

It can be used in conjunction with *BitBucket*, *GitHub*, *GitLab* and *Gitea* and it runs on 
whatever platform you compile it for, be it *Windows*, *Linux* or *macOS*, even on a RaspberryPi.
Built artifacts can be deployed via SFTP (FTP via SSH).
Releases will deliver standalone binaries for the popular operating systems. If you need
a build for another OS/ARCH, refer to section __Custom Build__ below.

### License

Free to use for any non-commercial purpose; refer to LICENSE.md

### Custom Build

* You need Golang installed (developed on version 1.15) 
* You need [mjibson/esc](https://github.com/mjibson/esc) to create embedded resources from 
templates and static assets.
* Clone the release branch of the repository
* Create the embedded resources with the command found in *docs/embed-command.txt*
* Build the binary (refer to __or__ use the ``build.ps1``)

### Documentation

* [Installation](docs/installation.md)
* [Create a build definition](docs/create-a-build-definition.md)
* [Create a webhook](docs/create-a-webhook.md)
* [Admin Settings](docs/admin-settings.md)

