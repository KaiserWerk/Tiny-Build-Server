# Creating a webhook

When you push new code into your repository, you have to notify your build server that new code
is available. This is what webhooks are for. Please read the paragraph corresponding to your used
Git system. For every webhook, you will need your build URL. As URL, you need to 
supply the address of your build server, the build ID and the auth token,
e.g. http://my-build-server.com:8271/api/v1/receive?id=my-id&token=my-auth-token.

With most Git services, the process is largely the same. Select the repository, go to Settings/Webhooks 
and supply the build URL and a token. Please read the specifics below.

### BitBucket

1. Go to [BitBucket.org](https://bitbucket.org/account/signin/) and log in using your credentials
2. Click the repository you want to create a webhook for.
3. On the left, click __Repository settings__.
4. In the Repository settings, click __Webhooks__.
5. Click __Add webhook__.
    * As a title, you can enter whatever you want, e.g. Build Webhook.
    * Enter the build URL.
    * For debugging, enable request history collection.
    * Click __Save__.

More Info: https://confluence.atlassian.com/bitbucket/manage-webhooks-735643732.html

### GitHub

1. Go to [GitHub.com](https://github.com/login) and log in with your credentials.
2. Click the repository you want to create a webhook for.
3. Click __Settings__ on the right side.
4. Click __Webhooks__.
5. On the right side, click __Add webhook__ and re-enter your password, if required.
6. Enter the build URL, set the content type to __application/json__ and click __Add Webhook__.

More Info: https://developer.github.com/webhooks/

### GitLab

* Go to [GitLab.com](https://gitlab.com/users/sign_in) and log in with your credentials.
* Click the repository you want to create a webhook for.
* On the left, Click Settings -> Webhooks.
* Enter the build URL with the ID and the token. As a trigger, check _Push events_. Click __Add webhook__.

More Info: https://gitlab.com/help/user/project/integrations/webhooks

### Gitea

* Go to your Gitea installation URL and log in.
* Click the repository in question.
* On the top right, click __Settings__, then __Webhooks__, then __Add Webhook__ and __Gitea__.
* Supply the build URL with ID and auth token. Set Method to POST and content type to application/json, if
not already set. As triggering event, set the __Push-Event__. To finish, click __Add Webhook__.

More Info: https://docs.gitea.io/en-us/webhooks/
