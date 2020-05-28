# Create a build definition

In the ``build_definitions`` subfolder, create a folder name ``build_<my-id>`` which contains a
YAML file called ``build.yaml``.
The clone and build folders will be generated when a build is started.

![Examplary folder structure](images/example-build-folder.png)

The ``<my-id>`` is an arbitrary alphanumeric key which identifies a build definition.
You can increment it for every build definition, like ``build_1``, ``build_2``, ``build_3`` and so on or 
give it a name like SUPERMEGADEATHBUILD5000, it's up to you.

The ``build.yaml`` file's content might look like this:

```yaml
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
      - "systemctl --user stop myservice"
    post_deployment_actions:
      - "systemctl --user start myservice"
```

**There is no need to restart the server after adding, modifying or removing a build definition.**

Field explanations:

  * ``auth_token`` represents an alphanumeric key with an approximate maximum length of 1800
character used to check if the sender of a build request is allowed to do so
  * ``project_type`` can either be ``golang`` or ``csharp``. Maybe I'll add support for PHP 
or other languages in the future
  * ``deplyoment_enabled`` allows you to enable or disable deployments for this build definition
  * ``repository`` contains information about the repository that is supposed to be built
    * ``host`` can either be github, bitbucket, gitlab or gitea
    * ``host_url`` is required to be supplied when you use a self-hosted service, like Gitea. If you use a cloud-base service
    this can be left blank
    * ``full_name`` must be the combination of the owner and the repository, like in a Git URL, e.g. ``KaiserWerk/Tiny-Build-Server``
    * ``username`` is the username used to authenticate for a ``git clone``.
    * ``secret`` is the associated password/secret token
    * ``branch`` must supply the name of the branch which should be built. Most of the time, this will be 
something like ``master`` or ``release``.
  * ``actions`` contains a list of commands to be executed sequentially in order for the project to be built. 
  Possible actions are
    * **restore** (to restore dependencies)
    * **test** (runs unit tests)
    * **test bench** (runs benchmark tests; only for Golang)
    * **build os_arch**

The OS and architecture can be any valid combination supported by your Golang installation, like **windows_amd32**. Also you can use **raspi3** and **raspi4** for use with the 
RaspberryPi 3 and 4, respectively.
  * ``deployments`` contains the deployment definitions. You can have multiple deployments defined.
  