# GitHub App Authentication Setup

This document explains how to set up GitHub App authentication for user login with WLTrack, following the [official GitHub documentation](https://docs.github.com/en/apps/creating-github-apps/writing-code-for-a-github-app/building-a-login-with-github-button-with-a-github-app).

## Creating a GitHub App

1. Go to your GitHub account settings
2. Select "Developer settings" > "GitHub Apps" > "New GitHub App"
3. Fill in the required information:
   - **GitHub App name**: A unique name for your app (e.g., "WLTrack-Auth")
   - **Homepage URL**: Your application's URL
   - **Callback URL**: `http://your-app-url/auth/github/callback` (replace with your actual URL)
   - **Request user authorization (OAuth) during installation**: Check this box
   - Under "Webhook", uncheck "Active" if you don't need webhooks
   - Under "Permissions", no additional permissions are required for basic authentication

2. Create the app and make note of the following information:
   - Client ID (shown in the app settings page)
   - Client Secret (shown in the app settings page)

3. Decide which GitHub users should be allowed to access your application

## Configuration Options

Set the following environment variables to enable GitHub App authentication:

```sh
# Enable authentication
export AUTH_ENABLED=true

# GitHub App credentials
export GITHUB_CLIENT_ID="your_client_id_here"
export GITHUB_CLIENT_SECRET="your_client_secret_here"
export GITHUB_REDIRECT_URL="http://your-app-url/auth/github/callback"
export ALLOWED_GITHUB_USERS="your_username,another_username"
```

## Running the Application with Authentication

Once you've set up the environment variables, run the application:

```sh
go run ./cmd/server/main.go
```

## How the Authentication Works

1. User clicks the "Login with GitHub" button
2. They're redirected to GitHub's authorization page
3. After granting access, GitHub redirects back to your callback URL with a code
4. Your application exchanges the code for a user access token
5. The access token is used to identify the user
6. The application verifies if the user's GitHub username is in the allowed list
7. If authorized, a session is created; otherwise, access is denied

## Security Considerations

- Use HTTPS in production to protect authentication data
- Store session data securely
- Consider implementing token refresh mechanisms for long-lived sessions
- Set appropriate cookie security flags (HttpOnly, SameSite, Secure)
- Restrict access to specific GitHub users using the ALLOWED_GITHUB_USERS environment variable

## Troubleshooting

If you encounter issues with GitHub App authentication:

1. Check that your environment variables are set correctly
2. Verify that your callback URL exactly matches what's configured in the GitHub App
3. Ensure your client secret is correct and hasn't been regenerated
4. Examine application logs for detailed error messages
4. Ensure your GitHub App is properly configured with OAuth enabled
5. Check if the user's GitHub username is in your ALLOWED_GITHUB_USERS list
6. Test with GitHub's OAuth flow debugging tools if available
