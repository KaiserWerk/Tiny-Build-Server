# Installation

Installing the **TinyBuildServer** is quite easy. 

* Start your MySQL server, create a new database (and you would usually set up a MySQL user). 
* Place the binary at an appropriate location, e.g. upload to any server.
* Start the binary once to create the ``app.yaml`` configuration file, 
set the configuration values according to your needs (mainly the database driver and DSN).
* Once the changes are applied, the database schema will be automatically applied at startup,
  if you use the `--automigrate` flag.

### Startup

Start the server with the following exemplary command:

``./tiny-build-server --port=1337 --config="/etc/tiny-build-server/app.yaml"``

For Windows it would be

``.\tiny-build-server.exe --port=1337 --config="C:\\TBS\\app.yaml"``

The default port is 8271. The default configuration file location is just ``app.yaml`` 
relative to the executable.
If you want to use default values, you can omit the parameters.

### Setup

By default, there are no registered users; the first account registration creates an 
administrative user you can use to manage your build server.

At first startup of the Docker image, an administrative user with the name 'admin', the email 'test@mail.org' and the 
password 'test' is created.