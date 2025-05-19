# WLTrak - Weightlifting Tracking Application

WLTrak is a minimal web application built with the GOTH stack (Go, templ, HTMX) for tracking weightlifting progress. It allows users to record their weightlifting workouts, track exercises, and visualize progress over time.

## Features

- Record daily weightlifting results (sets, reps, weight, percentage of 1RM)
- Store data in a Turso database (or local sqlite)
- Visualize progress with graphs for each exercise
- Responsive design with HTMX for dynamic interactions
- Secure token management
- GitHub App authentication

## Tech Stack

- **Backend**: Go
- **Templates**: [templ](https://github.com/a-h/templ)
- **Frontend Interactivity**: [HTMX](https://htmx.org/)
- **Database**: [Turso](https://turso.tech/)
- **Styling**: Tailwind CSS
- **Containerization**: Docker
- **Deployment**: Azure Container App

## Getting Started

### Prerequisites

- Go 1.24+
- Turso CLI and account
- Docker (for containerization)
- Azure CLI (for deployment)
- Optional: golangci-lint for linting

### Local Development

1. Clone the repository:
   ```
   git clone https://github.com/mahalel/wltrack.git
   cd wltrack
   ```

2. Install dependencies:
   ```
   go mod download
   go install github.com/a-h/templ/cmd/templ@latest
   ```

3. Create a Turso database:
   ```
   turso db create wltrak
   turso db tokens create wltrak
   ```

4. Set environment variables in your OS:
   ```
   # Set these in your shell if you want to use Turso
   export TURSO_URL="libsql://your-database-url.turso.io"
   export TURSO_AUTH_TOKEN="your-token"
   ```

   > Note: If you don't set the TURSO_URL, the application will automatically create and use a local SQLite database in the `data/wltrak.db` file. This is convenient for development and testing. You can delete this file at any time to reset your database.

5. Generate templ files:
   ```
   templ generate
   ```

6. Run the application:
   ```
   go run ./cmd/server/main.go
   ```

7. Open http://localhost:8080 in your browser

### Testing

Run the test suite:
```
just test
```

Run tests with coverage report:
```
just test-coverage
```

The project uses both real SQLite-based tests and mock database implementations for unit testing.

## Development Commands

You can use the provided justfile for common development tasks:

```
just lint        # Format and lint the code
just test        # Run tests
just test-race   # Run tests with race detection
just build       # Build the application
just check       # Run formatting, linting, and tests
just build-docker # Build Docker container
```

## Building and Running with Docker

1. Install just (optional, for convenient commands):
   ```
   # Using Homebrew
   brew install just

   # Or download from https://github.com/casey/just/releases
   ```

2. Use the justfile to build and run:
   ```
   # Generate templ files
   just generate

   # Build locally
   just build

   # Build with Docker (supports cross-compilation for linux/amd64)
   just build-docker

   # Run the container (uses OS environment variables)
   just run-docker

   # Show available commands
   just help
   ```

The project uses a multi-stage Dockerfile for efficient containerization. The static files are copied directly into the container, ensuring assets are available for serving.

## Deployment to Azure Container Apps

1. Log in to Azure:
   ```
   az login
   ```

2. Create a resource group:
   ```
   az group create --name wltrak-group --location eastus
   ```

3. Create a container registry:
   ```
   az acr create --resource-group wltrak-group --name wltrakregistry --sku Basic
   ```

4. Login to your registry:
   ```
   az acr login --name wltrakregistry
   ```

5. Build and push the Docker image:
   ```
   docker build -t wltrakregistry.azurecr.io/wltrak:latest .
   docker push wltrakregistry.azurecr.io/wltrak:latest
   ```

6. Create and update a container app:
   ```
   # Create the initial app
   az containerapp create \
     --name wltrak \
     --resource-group wltrak-group \
     --environment wltrak-env \
     --registry-server wltrakregistry.azurecr.io \
     --image wltrakregistry.azurecr.io/wltrak:latest \
     --target-port 8080 \
     --ingress external \
     --env-vars TURSO_URL="libsql://your-database-url.turso.io" TURSO_AUTH_TOKEN="your-token"
   ```

Note: The project can also be published to GitHub Container Registry (GHCR) using the included GitHub Actions workflow that triggers on version tag pushes. The workflow is configured to run tests, linting, and formatting checks before building and publishing the container.

## Project Structure

```
.
├── cmd
│   └── server          # Main application entry point
├── internal
│   ├── auth            # Authentication functionality
│   ├── config          # Application configuration
│   ├── database        # Database interactions
│   │   └── mock_db.go  # Mock database implementation for testing
│   ├── handlers        # HTTP handlers
│   ├── models          # Data models
│   └── templates       # templ templates
│       └── auth        # Authentication templates
├── static              # Static files for both local development and container
│   ├── css             # Stylesheets
│   └── js              # JavaScript files
├── docs                # Documentation
├── .github/workflows   # GitHub Actions CI/CD workflows
│   ├── ci.yml          # Lint, format, and test workflow
│   └── publish-container.yml # Container publishing workflow
├── Dockerfile          # Docker build configuration 
└── justfile            # Command runner for building, testing, and linting
```

## Environment Variables

The application reads the following environment variables directly from the OS:

- `TURSO_URL`: Your Turso database URL (defaults to local SQLite file in development)
- `TURSO_AUTH_TOKEN`: Your Turso auth token (required only for Turso)
- `PORT`: Server port (defaults to 8080)
- `ENV`: Environment name (defaults to "development")

### Authentication Environment Variables

Optional GitHub App authentication configuration:

- `AUTH_ENABLED`: Set to "true" to enable authentication (default: false)
- `GITHUB_CLIENT_ID`: Your GitHub App client ID
- `GITHUB_CLIENT_SECRET`: Your GitHub App client secret
- `GITHUB_REDIRECT_URL`: OAuth callback URL (e.g., http://localhost:8080/auth/github/callback)
- `ALLOWED_GITHUB_USERS`: Comma-separated list of GitHub usernames allowed to access the app

You can check your current environment setup with:

```bash
just show-env
```

## Development with just

The project includes a `justfile` that serves as a task runner with several useful commands:

```bash
# Show available commands
just help

# Build and run locally
just run

# Build and run in container
just build-docker
just run-docker

# Show current environment variables
just show-env
```

## Authentication with GitHub App

WLTrak supports authentication using GitHub Apps for user login. To set up GitHub App authentication:

1. Create a GitHub App in your GitHub account settings
2. Configure the OAuth settings for the app
3. Set the environment variables for authentication (including allowed GitHub usernames)
4. Run the application with authentication enabled

For detailed setup instructions, see [GitHub App Authentication](docs/github-app-auth.md).

## Continuous Integration

The project uses GitHub Actions for CI/CD:

1. On every push to `main` and PR, the CI workflow runs:
   - Code formatting check with `go fmt`
   - Static analysis with `go vet` and golangci-lint
   - Unit tests with Go's testing framework
   - Test coverage reporting

2. When a version tag (v*) is pushed, the container workflow runs:
   - First ensures all CI checks pass
   - Builds the Docker container
   - Publishes the container to GitHub Container Registry

This ensures that all tests and checks pass before any container is built and published.
