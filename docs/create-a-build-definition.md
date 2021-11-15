# Create a build definition

First, log into your build server. Now on the left, click *Build Definition*, then
*Add Definition*.

You are presented with a simple form with two elements, the *Caption*, which is a freely 
selectable short description for your new build definition, and the *Content*, the
actual build definition.
The textarea will be filled with a basic skeleton which can be modified 
and extended by you.
The "language" used is called [YAML](https://yaml.org/) and you should be familiar 
with it at a basic level.

Elements like ``${artifact}`` are variables. Some variables, like ``${artifact}``
exist by default and are filled in automatically when a new build execution is triggered.
Others can be created and inserted by you, e.g. ``${myvar}`` (see [Using Variables](using-variables.md)).

There are many sections which can be used, some of them are optional. Here are some examples.

#### Repository (required)
```yaml
repository:
  hoster: github
  hoster_url: https://github.com/KaiserWerk/Tiny-Build-Server
  name: KaiserWerk/Tiny-Build-Server
  access_user: <empty for public repositories>
  access_secret: <required for private repositories>
  branch: release
```

#### Setup, Test and Build sections

There are five sections, *setup*, *test*, *pre_build*, *build* and *post_build*.
You could generally just pack all required steps into one section, these captions are just
for separating concerns. 
That means, in every section, you can set (and unset) environment variables,
use git commands, use to ``go`` command or execute any arbitrary command you like, e.g.
signing the artifact using [minisign](https://jedisct1.github.io/minisign/) or compressing
a binary using [UPX](https://upx.github.io/).
An example might look like this:

```yaml
setup:
  - setenv MYENV release
test:
  - go test ./...
  - go test -bench=.
pre_build:
  - unsetenv MYENV
  - setenv GOOS windows
  - setenv GOARCH amd64
build:
  - go build ${cloneDir}/cmd/myapp/main.go
post_build:
  - minisign -Sm ${artifact} -t 'This comment will be signed as well'
```

By default, the linker flags ``-ldflags "-s -w""`` are set. It is currently not possibly
to modify that behaviour. I'm working on a working solution.

Currently, the following default variables are available:

* ``${artifact}`` contains the internal directory and filename to artifact which is about
to be created (for GOOS=windows, *.exe* is appended automatically)
* ``${cloneDir}`` contains the internal directory which the repository was cloned into

#### Deployments

There are three types of deployments: local deployments, email deployments and remote
deployments.
Local deployments basically just copy the artifact to a different directory, e.g. a 
net drive or an external hard drive.
Email deployments zip the artifact and send out a notification email with the zipped
artifact attached.
Remote deployments copy the artifact to a remote machine using SFTP. Besides the 
usual connection/authentication data you can supply the desired target directory
as well as pre- and post-deployment commands.
All kinds of deployments can be enabled/disabled separately.
Example:

```yaml
deployments:
  local_deployments:
    - enabled: true
      path: /mnt/ext/tbs-artifacts
    - enabled: true
      path: /opt/drive/myapp
  email_deployments:
    - enabled: true
      address: my@address.com
    - enabled: false
      address: other@address.com
    - enabled: true
      address: you@address.com
  remote_deployments:
    - enabled: true
      host: somemachine.org
      port: 22 # Port 22 is the usual default
      connection_type: sftp # currently, only sftp is supported
      username: username
      password: 'pass@word'
      working_directory: /opt/myapp
      pre_deployment_steps:
        - systemctl stop myservice
      post_deployment_steps:
        - systemctl start myservice
```