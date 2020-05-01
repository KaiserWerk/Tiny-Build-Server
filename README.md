# Tiny Build Server

This is a functioning, minimalistic build server for Golang and C# projects.
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
    only_custom_actions: true
    repository:
      host: github
      full_name: myuser/test-repo
      username: myuser
      secret: xyz987
      branch: master
    
    actions:
      - "git clone https://github.com/myuser/test-repo clone"
      - "go get"
      - "env GOOS=linux GOARCH=amd64 go build -o %output_dir/example_go %package"
      - "env GOOS=linux GOARCH=arm GOARM=5 go build -o %output_dir/example_go.exe %package"
      - "env GOOS=windows GOARCH=amd64 go build -o %output_dir/example_go %package"

There is no need to restart the server after adding, modifying or removing a build definition.

Field explanations:
  * ``auth_token`` represents an alphanumeric key with an approximate maximum length of 1800
character used to check if the sender of a build request is allowed to do so
  * ``project`` can either be ``golang`` or ``csharp``. Maybe I'll add support for PHP 
or other languages in the future
  * ``only_custom_actions`` is a boolean flag which, if set to true, allows you to use custom commands
under the actions key instead of predefined ones
  * ``repository`` contains information about the repository that is supposed to be built
    * ``host`` can either be github, bitbucket or gitlab
    * ``full_name`` must be the combination of the repository and the owning user, like in a Git URL
    * ``username`` is the username used to authenticate for a ``git pull``.
    * ``secret`` is the associated password/secret token
    * ``branch`` must supply the name of the branch which should be built. Most of the time, this will be 
something like ``master`` or ``release``.
  * ``actions`` contains a list of commands to be execute after each other in order for the project
to be built