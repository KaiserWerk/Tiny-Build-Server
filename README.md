[![Go Report Card](https://goreportcard.com/badge/github.com/KaiserWerk/Tiny-Build-Server)](https://goreportcard.com/report/github.com/KaiserWerk/Tiny-Build-Server)

# Tiny Build Server

This project is, or rather will be, a functioning, minimalistic build server for Golang and C# projects.
It is not meant to be used in production since it is a proof-of-concept, but this is up to you.
It can be used in conjunction with BitBucket, GitHub and GitLab and it runs on 
whatever platform you compile it for, be it Windows, Linux or MacOS, even a RaspberryPi.

Build information are stored in files and folders which are placed in the ``build_definitions``
subfolder. A build definition is a folder containing a YAML file.

Prerequisites
* Basic knowledge of the Go programming language

### Documentation

* [Part I: Installation](docs/installation.md)
* [Part II: Create a build definition](docs/create-a-build-definition.md)
* [Part III: Create a webhook](docs/create-a-webhook.md)


  