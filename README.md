[![Go Report Card](https://goreportcard.com/badge/github.com/KaiserWerk/Tiny-Build-Server)](https://goreportcard.com/report/github.com/KaiserWerk/Tiny-Build-Server)

# Tiny Build Server

This project is, or rather will be, a functioning, minimalistic build server for Golang and C# projects.
It is not meant to be used in production since it is a proof-of-concept.
It can be used in conjunction with BitBucket, GitHub and GitLab and it runs on 
whatever platform you compile it for, be it Windows, Linux or MacOS, even a RaspberryPi.
You need to make some changes to your build definitions, though.

Build information is stored in files and folders which are placed in the ``build_definitions``
subfolder. A build definition is a folder containing a YAML file.

Example for Golang:
In the ``build_definitions`` subfolder, create a folder name ``build_1`` which contains a
YAML file called ``build.yaml``.
The YAML file's content might look like this:

    auth_token: 123abc
    project_type: golang
    deployment_enabled: true
    repository:
      host: github
      host_url:   # empty when using a cloud-based service
      full_name: myuser/test-repo
      username: myuser
      secret: xyz987
      branch: master
    actions:
      - "restore"
      - "test"
      - "test bench"
      - "build linux_amd64"
    deployments:
      - host: "mydomain.com:22"
        username: myuser
        password: mysecretpassword
        connection_type: sftp  # currently, only sftp (SSH File Transfer Protocol) is supported
        working_directory: /usr/local/bin/my_binary
        pre_deployment_actions:
          - "sudo service myservice stop"
        post_deployment_actions:
          - "sudo service myservice start"

There is no need to restart the server after adding, modifying or removing a build definition.

Field explanations:
  * ``auth_token`` represents an alphanumeric key with an approximate maximum length of 1800
character used to check if the sender of a build request is allowed to do so
  * ``project`` can either be ``golang`` or ``csharp``. Maybe I'll add support for PHP 
or other languages in the future
  * ``deplyoment_enabled`` allows you to enable or disable deployments for this build definition
  * ``repository`` contains information about the repository that is supposed to be built
    * ``host`` can either be github, bitbucket, gitlab or gitea
    * ``host_url`` is required to be supplied when you use a self-hosted service, like Gitea. If you use a cloud-base service
    this can be left blank
    * ``full_name`` must be the combination of the repository and the owning user, like in a Git URL, e.g. KaiserWerk/Tiny-Build-Server
    * ``username`` is the username used to authenticate for a ``git pull``.
    * ``secret`` is the associated password/secret token
    * ``branch`` must supply the name of the branch which should be built. Most of the time, this will be 
something like ``master`` or ``release``.
  * ``actions`` contains a list of commands to be execute after each other in order for the project
to be built. This can be either **restore** (to restore dependencies), **test** (runs unit tests), **test bench** (runs benchmark tests; only for Golang) and **build os_arch**.
The OS and architecture can be any valid combination supported by your Golang installation, like **windows_amd32**. Also you can use **raspi3** and **raspi4** for use with the 
RaspberryPi 3 and 4, respectively.
  * ``deployments`` contains the deployment definitions. You can have multiple 