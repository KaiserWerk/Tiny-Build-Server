# Using Variables

Variables in build definitions are usually used to store secrets or values
that can be used in multiple places/build definitions

When being logged in, click *Variables* on the left-hand navigation to
list all **available** Variables.
These are those you created and those others created and marked them *Public*.

If a public Variable from a different user and a Variable from you have the same 
name, your own Variable will have higher precedence.

Example scenario:
You can to create a build pipeline for a private GitHub repository, you need a 
**Personal Access Token**.
You can create a Variable called e.g. **github_token** and use the corresponding
placeholder ``${github_token}`` in your build definition, e.g.
```yaml
name: vendor/repo-name
access_user: 
access_secret: ${github_token}
branch: master
```