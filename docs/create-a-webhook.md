# Creating a webhook

When you push new code into your repository, you have to notify your build server that new code
is available. This is what webhooks are for. Please read the paragraph corresponding to your used
Git system. For every webhook, you will need your build URL. As URL, you need to 
supply the address of your build server, the build ID and the auth token,
e.g. http://my-build-server.com:5000/bitbucket-receive?id=<my-id>&token=<my-auth-token>.

### BitBucket

1. Go to [bitbucket.org sign in](https://bitbucket.org/account/signin/) and log in using your credentials
2. Click the repository you want to create a webhook for.
3. On the left, click __Repository settings__.
4. In the Repository settings, click on __Webhooks__.
5. Click on __Add webhook__.
    * As a title, you can enter whatever you want, e.g. Build Webhook.
    * Enter the build URL.
    * For debugging, enable request history collection.
    * Click __Save__.

More Info: https://confluence.atlassian.com/bitbucket/manage-webhooks-735643732.html

### GitHub

1. Go to [GitHub Login](https://github.com/login) and log in with your credentials.
2. Click on the repository you want to create a webhook for.
3. Click __Settings__ on the right side.
4. Click __Webhooks__.
5. On the right side, click __Add webhook__ and re-enter your password, if required.
6. Enter the build URL, set the content type to __application/json__ and click __Add Webhook__.

### GitLab

coming soon

### Gitea

coming soon
