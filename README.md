# WLTrak - Weightlifting Tracking Application

WLTrak is a minimal web application built with the GOTH stack (Go, templ, HTMX) for tracking weightlifting progress. It allows users to record their weightlifting workouts, track exercises, and visualize progress over time.

## Features

- Record daily weightlifting results (sets, reps, weight, percentage of 1RM)
- Store data in a Turso database
- Visualize progress with graphs for each exercise
- Responsive design with HTMX for dynamic interactions
- Secure token management

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

### Local Development

1. Clone the repository:
   ```
   git clone https://github.com/your-username/wltrak.git
   cd wltrak
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
   # Set these in your shell
   export TURSO_URL="libsql://your-database-url.turso.io"
   export TURSO_AUTH_TOKEN="your-token"
   ```

5. Generate templ files:
   ```
   templ generate
   ```

6. Run the application:
   ```
   go run ./cmd/server/main.go
   ```

7. Open http://localhost:8080 in your browser

# Building and Running with Docker

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

Note: The project can also be published to GitHub Container Registry (GHCR) using the included GitHub Actions workflow that triggers on version tag pushes.

## Project Structure

```
.
├── cmd
│   └── server          # Main application entry point
├── internal
│   ├── config          # Application configuration
│   ├── database        # Database interactions
│   ├── handlers        # HTTP handlers
│   ├── models          # Data models
│   └── templates       # templ templates
├── static              # Static files for both local development and container
│   ├── css             # Stylesheets
│   └── js              # JavaScript files
├── .github/workflows   # GitHub Actions CI/CD workflows
├── Dockerfile          # Docker build configuration
└── justfile            # Command runner for development tasks
```

## Environment Variables

The application reads the following environment variables directly from the OS:

- `TURSO_URL`: Your Turso database URL (required)
- `TURSO_AUTH_TOKEN`: Your Turso auth token (required)
- `PORT`: Server port (defaults to 8080)
- `ENV`: Environment name (defaults to "development")

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
