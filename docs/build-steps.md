### General process
* receive payload
* read build configuration by ID
* make sure branch in build configuration is the same as in the payload
* remove build definition's clone and build folders
* clone the repository

### Examplary build actions
* restore dependencies
* build the project
* deploy generated binary to remote server(s) via SFTP


### Possible build steps
* clone (will always be done as a first step; automagically)
* restore
* test
* test bench
* build arch

arch = window_amd64, darwin_amd32, raspi3, ...