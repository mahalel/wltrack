package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mahalel/wltrack/internal/auth"
	"github.com/mahalel/wltrack/internal/config"
	authtmpl "github.com/mahalel/wltrack/internal/templates/auth"
)

// Secret masking for logging has been moved to the auth package

// GitHubApp holds the GitHub App authentication instance
var GitHubApp *auth.GitHubApp

// InitGitHubAuth initializes GitHub App authentication
func InitGitHubAuth(cfg config.Config) error {
	if !cfg.AuthEnabled {
		log.Println("GitHub authentication is disabled")
		return nil
	}

	// Skip if no GitHub client ID is provided
	if cfg.GithubClientID == "" {
		log.Println("No GitHub Client ID provided, skipping authentication setup")
		return nil
	}

	// Configure the GitHub App for user login
	authConfig := auth.GithubAuthConfig{
		ClientID:     cfg.GithubClientID,
		ClientSecret: cfg.GithubClientSecret,
		RedirectURL:  cfg.GithubRedirectURL,
		AllowedUsers: cfg.AllowedGithubUsers,
	}

	if cfg.GithubClientID == "" {
		log.Println("No GitHub Client ID provided")
	}
	if cfg.GithubClientSecret == "" {
		log.Println("No GitHub Client Secret provided")
	}

	var err error
	GitHubApp, err = auth.NewGitHubApp(authConfig)
	if err != nil {
		fmt.Printf("[ERROR] Failed to initialize GitHub App: %v\n", err)
		return err
	}

	log.Println("GitHub App authentication initialized successfully")
	return nil
}

// GithubAuthMiddleware returns the GitHub authentication middleware if enabled
func GithubAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication if GitHub App is not initialized
			if GitHubApp == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Skip authentication for login and callback routes
			switch r.URL.Path {
			case "/auth/github/login", "/auth/github/callback", "/login":
				next.ServeHTTP(w, r)
				return
			}

			// Skip authentication for static assets and public paths
			if r.URL.Path == "/favicon.ico" ||
				r.URL.Path == "/robots.txt" ||
				r.URL.Path == "/health/ready" ||
				r.URL.Path == "/health/live" ||
				strings.HasPrefix(r.URL.Path, "/static/") {
				next.ServeHTTP(w, r)
				return
			}

			// Special handling for the root path when not authenticated
			if r.URL.Path == "/" {
				// Check if user has a valid session before allowing access to root
				// Get token from cookie
				cookie, err := r.Cookie("github_token")
				if err != nil || cookie.Value == "" {
					http.Redirect(w, r, "/login", http.StatusFound)
					return
				}
			}

			if r.URL.Path == "/logout" {
				GitHubApp.LogoutHandler().ServeHTTP(w, r)
				return
			}

			// Authentication middleware
			GitHubApp.Middleware()(next).ServeHTTP(w, r)
		})
	}
}

// LoginPageHandler serves the login page
func LoginPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip showing login page if GitHub App is not initialized
		if GitHubApp == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Check if user is already authenticated
		cookie, err := r.Cookie("github_token")
		if err == nil && cookie.Value != "" {
			// If already logged in, redirect to home
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		component := authtmpl.LoginPage()
		_ = component.Render(r.Context(), w)
	}
}

// GitHubLoginHandler handles GitHub login
func GitHubLoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if GitHubApp == nil {
			http.Error(w, "GitHub authentication is not configured", http.StatusInternalServerError)
			return
		}

		GitHubApp.LoginHandler().ServeHTTP(w, r)
	}
}

// GitHubCallbackHandler handles GitHub OAuth callback
func GitHubCallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if GitHubApp == nil {
			http.Error(w, "GitHub authentication is not configured", http.StatusInternalServerError)
			return
		}

		GitHubApp.CallbackHandler().ServeHTTP(w, r)
	}
}

// LogoutHandler handles user logout
func LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if GitHubApp == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		GitHubApp.LogoutHandler().ServeHTTP(w, r)
	}
}
