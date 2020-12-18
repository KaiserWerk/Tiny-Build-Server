# Admin Settings

### General Settings

* Data Path - The absolute path of the directory where TBS saves (temporary) working data
* Base URL - The base URL under which the TBS installation is accessible

### Security Settings

* Disable registration - Allow you to manage users manually as an administrator
* Disable password reset - In case you want to manually set new passwords for users who forgot theirs
* Require email activation of new accounts - If checked, locks a new account, sends out a confirmation email
and requires the registering user to click the link
* Two-Factor Authentication - Allows you to restrict logging in by having to complete a secondary authentication step 
  using a code, either via email or SMS

### SMTP Settings

The usage of the sending of emails via SMTP might require you to enable a setting like 
"Enable less secure apps" when using a 3rd party email service like Yahoo or Gmail.

### Executable Paths

These absolute path to the build executable only need to be set if they are not globally
available, that means from any directory.