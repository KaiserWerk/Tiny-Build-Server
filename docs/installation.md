# Installation

Installing the build server is quite easy. 

* Start your MySQL server, import the schema dump from ``docs/schema.sql`` and setup a MySQL
user. 
* Place the binary at an appropriate location, e.g. upload to any server.
* Create a configuration file (you can copy the ``config/app.dist.yaml`` as a starting point), 
set the configuration values according to your needs, mainly the MySQL DSN.
The default location is ``config/app.yaml`` relative to the executable. For more info,
refer to ``config/app.dist.yaml``.

### Startup

Start the server with the following command:

``./tiny-build-server -port 8271 -config config/app.yaml``

The default port is 8271. If you want to use default values, you can omit the parameters.

### Setup

By default, there are no registered users; the first account registration creates an 
administrative user you can use to manage your build server.