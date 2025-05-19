package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"slices"

	"github.com/google/go-github/v58/github"
	"golang.org/x/oauth2"
)

// GithubAuthConfig holds the configuration for GitHub App authentication
type GithubAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AllowedUsers []string
	TokenCache   *TokenCache
}

// TokenCache provides caching for GitHub tokens to reduce the number of API calls
type TokenCache struct {
	mu              sync.RWMutex
	appToken        string
	appTokenExpiry  time.Time
	instToken       string
	instTokenExpiry time.Time
}

// NewTokenCache creates a new token cache
func NewTokenCache() *TokenCache {
	return &TokenCache{}
}

// GetAppToken returns the cached app token if valid, or fetches a new one
func (tc *TokenCache) GetAppToken(getTokenFn func() (string, time.Time, error)) (string, error) {
	tc.mu.RLock()
	if tc.appToken != "" && tc.appTokenExpiry.After(time.Now().Add(time.Minute)) {
		token := tc.appToken
		tc.mu.RUnlock()
		return token, nil
	}
	tc.mu.RUnlock()

	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Double-check after acquiring write lock
	if tc.appToken != "" && tc.appTokenExpiry.After(time.Now().Add(time.Minute)) {
		return tc.appToken, nil
	}

	token, expiry, err := getTokenFn()
	if err != nil {
		return "", err
	}

	tc.appToken = token
	tc.appTokenExpiry = expiry
	return token, nil
}

// GetInstallationToken returns the cached installation token if valid, or fetches a new one
func (tc *TokenCache) GetInstallationToken(getTokenFn func() (string, time.Time, error)) (string, error) {
	tc.mu.RLock()
	if tc.instToken != "" && tc.instTokenExpiry.After(time.Now().Add(time.Minute)) {
		token := tc.instToken
		tc.mu.RUnlock()
		return token, nil
	}
	tc.mu.RUnlock()

	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Double-check after acquiring write lock
	if tc.instToken != "" && tc.instTokenExpiry.After(time.Now().Add(time.Minute)) {
		return tc.instToken, nil
	}

	token, expiry, err := getTokenFn()
	if err != nil {
		return "", err
	}

	tc.instToken = token
	tc.instTokenExpiry = expiry
	return token, nil
}

// GitHubApp represents a GitHub App with authentication capabilities
type GitHubApp struct {
	config     GithubAuthConfig
	httpClient *http.Client
}

// NewGitHubApp creates a new GitHub App with the provided configuration
func NewGitHubApp(config GithubAuthConfig) (*GitHubApp, error) {
	if config.TokenCache == nil {
		config.TokenCache = NewTokenCache()
	}

	app := &GitHubApp{
		config:     config,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	return app, nil
}

// These functions are not needed for the GitHub App user authentication flow
// according to https://docs.github.com/en/apps/creating-github-apps/writing-code-for-a-github-app/building-a-login-with-github-button-with-a-github-app

// GetOAuthConfig returns an OAuth2 configuration for the GitHub App
func (app *GitHubApp) GetOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID: app.config.ClientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
		RedirectURL: app.config.RedirectURL,
		Scopes:      []string{"read:user"},
	}
}

// Middleware creates an HTTP middleware for GitHub App authentication
func (app *GitHubApp) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from cookie
			cookie, err := r.Cookie("github_token")
			if err != nil || cookie.Value == "" {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			// Validate token
			ctx := r.Context()
			ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cookie.Value})
			tc := oauth2.NewClient(ctx, ts)

			// Create GitHub client
			githubClient := github.NewClient(tc)
			user, _, err := githubClient.Users.Get(ctx, "")
			if err != nil {
				http.SetCookie(w, &http.Cookie{
					Name:    "github_token",
					Value:   "",
					Path:    "/",
					Expires: time.Unix(0, 0),
				})
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			// Add user to context using a dedicated type for the key
			type userContextKey struct{}
			ctx = context.WithValue(ctx, userContextKey{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LoginHandler creates an HTTP handler for GitHub login
func (app *GitHubApp) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("[DEBUG] GitHub login handler called")

		// Generate cryptographically secure random state to prevent CSRF attacks
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			http.Error(w, "Failed to generate state token", http.StatusInternalServerError)
			return
		}
		state := base64.URLEncoding.EncodeToString(b)

		// Store state in cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "github_state",
			Value:    state,
			Path:     "/",
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   60 * 5, // 5 minutes
		})

		// Create the GitHub OAuth URL directly
		url := fmt.Sprintf(
			"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=read:user&state=%s",
			app.config.ClientID,
			app.config.RedirectURL,
			state,
		)
		fmt.Printf("[DEBUG] Redirecting to GitHub OAuth URL: %s\n", url)
		http.Redirect(w, r, url, http.StatusFound)
	}
}

// CallbackHandler creates an HTTP handler for GitHub OAuth callback
func (app *GitHubApp) CallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[DEBUG] OAuth Callback called. Raw URL: %s\n", r.URL.String())

		// Verify state parameter to prevent CSRF
		stateCookie, err := r.Cookie("github_state")
		if err != nil {
			http.Error(w, "State verification failed: No state cookie", http.StatusBadRequest)
			return
		}

		stateParam := r.URL.Query().Get("state")
		if stateParam == "" || stateParam != stateCookie.Value {
			http.Error(w, "State verification failed: Invalid state", http.StatusBadRequest)
			return
		}

		// Clear state cookie
		http.SetCookie(w, &http.Cookie{
			Name:    "github_state",
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0),
		})

		code := r.URL.Query().Get("code")
		if code == "" {
			fmt.Println("[ERROR] Code parameter missing in callback")
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}
		fmt.Printf("[DEBUG] Received authorization code: %s...\n", code[:10])

		// Create a direct request to exchange the code for an access token
		// Following GitHub's recommended approach for GitHub App user authentication
		// https://docs.github.com/en/apps/creating-github-apps/writing-code-for-a-github-app/building-a-login-with-github-button-with-a-github-app

		fmt.Println("[DEBUG] Starting token exchange with GitHub")

		// Setup HTTP client with timeout
		httpClient := &http.Client{Timeout: 10 * time.Second}

		// Create request body
		data := map[string]string{
			"client_id":     app.config.ClientID,
			"client_secret": app.config.ClientSecret,
			"code":          code,
			"redirect_uri":  app.config.RedirectURL,
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("[ERROR] Failed to marshal request: %v\n", err)
			http.Error(w, "Failed to create token request: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println("[DEBUG] Prepared token exchange request data")

		// Create token exchange request
		req, err := http.NewRequest(
			"POST",
			"https://github.com/login/oauth/access_token",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			fmt.Printf("[ERROR] Failed to create request: %v\n", err)
			http.Error(w, "Failed to create token request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Set headers for JSON request and response
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Execute the request
		fmt.Println("[DEBUG] Sending token exchange request to GitHub")
		resp, err := httpClient.Do(req)
		if err != nil {
			fmt.Printf("[ERROR] Token exchange request failed: %v\n", err)
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		fmt.Printf("[DEBUG] Token exchange response status: %s\n", resp.Status)

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("[ERROR] Failed to read response body: %v\n", err)
			http.Error(w, "Failed to read response: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("[DEBUG] Raw response: %s\n", string(respBody))

		// Parse the response
		// Parse the token response
		var tokenResp struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			Scope       string `json:"scope"`
			Error       string `json:"error"`
			ErrorDesc   string `json:"error_description"`
		}

		if err := json.Unmarshal(respBody, &tokenResp); err != nil {
			fmt.Printf("[ERROR] Failed to parse response: %v, raw response: %s\n", err, string(respBody))
			http.Error(w, "Failed to parse token response: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if tokenResp.Error != "" {
			errMsg := fmt.Sprintf("OAuth error: %s: %s", tokenResp.Error, tokenResp.ErrorDesc)
			fmt.Printf("[ERROR] %s\n", errMsg)
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}

		if tokenResp.AccessToken == "" {
			fmt.Println("[ERROR] No access token received from GitHub")
			http.Error(w, "No access token in response", http.StatusInternalServerError)
			return
		}

		fmt.Println("[DEBUG] Successfully received GitHub access token")

		// We'll validate the token with a direct API call below

		// Verify the token works by making a test request
		userClient := &http.Client{Timeout: 10 * time.Second}
		userReq, err := http.NewRequest("GET", "https://api.github.com/user", nil)
		if err != nil {
			fmt.Printf("[ERROR] Failed to create user validation request: %v\n", err)
			http.Error(w, "Failed to validate token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
		userReq.Header.Set("Accept", "application/json")

		fmt.Println("[DEBUG] Validating token with GitHub API")
		userResp, err := userClient.Do(userReq)
		if err != nil {
			fmt.Printf("[ERROR] Failed to validate token: %v\n", err)
			http.Error(w, "Failed to validate token: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer userResp.Body.Close()

		if userResp.StatusCode != http.StatusOK {
			userErrBody, _ := io.ReadAll(userResp.Body)
			fmt.Printf("[ERROR] Token validation failed with status %d: %s\n", userResp.StatusCode, string(userErrBody))
			http.Error(w, fmt.Sprintf("Token validation failed with status %d", userResp.StatusCode), http.StatusInternalServerError)
			return
		}

		var userData struct {
			Login string `json:"login"`
			Name  string `json:"name"`
		}

		if err := json.NewDecoder(userResp.Body).Decode(&userData); err != nil {
			fmt.Printf("[ERROR] Failed to parse user info: %v\n", err)
			http.Error(w, "Failed to parse user info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("[DEBUG] Successfully authenticated GitHub user: %s (%s)\n", userData.Login, userData.Name)

		// Check if the user is in the allowed list
		if len(app.config.AllowedUsers) > 0 {
			userAllowed := slices.Contains(app.config.AllowedUsers, userData.Login)

			if !userAllowed {
				fmt.Printf("[ERROR] User %s is not in the allowed users list\n", userData.Login)
				http.Error(w, "You are not authorized to access this application", http.StatusForbidden)
				return
			}
			fmt.Printf("[DEBUG] User %s is in the allowed users list\n", userData.Login)
		}

		// Set token in cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "github_token",
			Value:    tokenResp.AccessToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   60 * 60 * 24 * 7, // 1 week
		})

		// Redirect to home page
		fmt.Println("[DEBUG] Authentication successful, redirecting to home page")
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// LogoutHandler creates an HTTP handler for logout
func (app *GitHubApp) LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Clear cookie
		http.SetCookie(w, &http.Cookie{
			Name:    "github_token",
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0),
		})

		http.Redirect(w, r, "/", http.StatusFound)
	}
}
