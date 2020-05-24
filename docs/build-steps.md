### General process
* receive payload
* read build configuration by ID
* make sure branch in build configuration is the same as in the payload
* remove build definition's clone folder
* clone the repository

### Examplary build actions
* restore dependencies
* build the project
* deploy generated binary to remote server(s)


### Possible build steps
* clone
* restore
* test
* test bench
* build arch

arch = window_amd64, darwin_amd32, raspi_arm5, ...