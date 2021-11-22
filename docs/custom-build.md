# Custom Build

* You need to have [Golang](https://golang.org/) installed 
* Clone the release branch of the repository
* Navigate into the root directory of the project (usually ``Tiny-Build-Server``)
* Build the binary (refer to **or** use the ``build.ps1`` or ``build.sh``) using the 
command
``go build -o <output-filename> -ldflags "-s -w" cmd/tiny-build-server/main.go``
  * Using the linker flag ``-X``, you can set ``Version`` and ``VersionDate`` to values
  you desire, as well. Example:
    ``go build -o tiny-build-server -ldflags "-s -w -X 'main.Version=1.13.4' -X 'main.VersionDate=2021-11-22T23:57.42.086'" cmd/tiny-build-server/main.go``