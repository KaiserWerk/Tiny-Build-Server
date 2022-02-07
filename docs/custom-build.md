# Custom Build

* You need to have [Golang](https://golang.org/) v1.16+ installed 
* Clone the repository and check out the tag you would like to build
* Build the binary from the root directory (refer to **or** use the ``build.ps1`` or ``build.sh``) using the 
command
``go build -o <output-filename> -ldflags "-s -w" cmd/tiny-build-server/main.go``
  * Using the linker flag ``-X``, you can set ``Version`` and ``VersionDate`` to values
  you desire, as well. Example:
    ``go build -o tiny-build-server-binary -ldflags "-s -w -X 'main.Version=1.13.4' -X 'main.VersionDate=2021-11-22T23:57.42.086'" cmd/tiny-build-server/main.go``