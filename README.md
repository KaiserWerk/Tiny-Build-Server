[![Go Report Card](https://goreportcard.com/badge/github.com/KaiserWerk/Tiny-Build-Server)](https://goreportcard.com/report/github.com/KaiserWerk/Tiny-Build-Server)

# Tiny Build Server

This project is a functioning, minimal build server for Golang and C# projects (possibley PHP and Rust).

It can be used in conjunction with BitBucket, GitHub, GitLab and Gitea and it runs on 
whatever platform you compile it for, be it Windows, Linux or MacOS, even a RaspberryPi.
Releases will deliver standalone binaries for the popular operating systems.

### License

Free to use for any non-commercial purpose; refer to LICENSE.md

### Custom build

* You need Golang version 1.15+ installed
* you need [mjibson/esc](https://github.com/mjibson/esc) to created embedded resources from 
templates and static assets
* Clone the repository
* created the embedded resources with the command ``esc -pkg internal -o internal/embed.go templates/ public/``
* Build the binary (refer to or use the ``build.ps1``)

### Documentation

* [Part I: Installation](docs/installation.md)
* [Part II: Create a build definition](docs/create-a-build-definition.md)
* [Part III: Create a webhook](docs/create-a-webhook.md)


  