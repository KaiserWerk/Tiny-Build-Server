# Custom Build

* You need [Golang](https://golang.org/) installed (developed on version 1.16)
* Clone the release branch of the repository
* Navigate into the root directory of the project (usually ``Tiny-Build-Server``)
* Build the binary (refer to **or** use the ``build.ps1`` or ``build.sh``) using the
``go build -o <output-filename> -ldflags "-s -w" cmd/tiny-build-server/main.go``
* Using the linker flag ``-X``, you can set ``Version`` and ``VersionDate`` to values
you desire, as well.