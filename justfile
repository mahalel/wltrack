# WLTrack commands
#
# Directory Structure Notes:
# - Static assets are stored in the 'static' directory

# Default command runs generation and build
default: clean generate build

# Generate templ files
generate:
    @echo "Generating templ files..."
    templ generate

# Build the Go application locally
build: generate
    @echo "Building Tailwind CSS..."
    npm run minify:css
    @echo "Building Go binary..."
    mkdir -p bin
    CGO_ENABLED=1 go build -o bin/server ./cmd/server

# Run the application locally
run: build
    @echo "Running application..."
    ENV=development ./bin/server

# Run in development mode with Tailwind CSS watcher
dev: generate
    @echo "Running in development mode with Tailwind CSS watcher..."
    npm run dev

# Build container using Docker (for cross-compilation with CGO support)
build-docker: clean generate
    @echo "Building container with Docker for cross-compilation with CGO..."
    docker buildx build --platform linux/amd64 -t wltrack:latest .
    @echo "Container built successfully as 'wltrack:latest'"


# Run the application in a Docker container
# This passes the host OS environment variables to the container
run-docker: build-docker
    @echo "Running container built with Docker..."
    docker run -p 8080:8080 \
        -e TURSO_URL \
        -e TURSO_AUTH_TOKEN \
        -e PORT="8080" \
        -e ENV="development" \
        wltrack:latest

# Run tests
test: generate
    @echo "Running tests..."
    go test -v ./...

# Run tests with race detection
test-race: generate
    @echo "Running tests with race detection..."
    go test -race -v ./...

# Run tests with coverage report
test-coverage: generate
    @echo "Running tests with coverage reporting..."
    go test -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    rm -rf bin/*
    find . -type f -name '*_templ.go' -delete
    @echo "Note: This doesn't remove the static files - they're shared between dev and container"

# Format and lint the code
format:
    @echo "Formatting code..."
    go fmt ./...
    @echo "Vetting code..."
    go vet ./...

lint:
    @echo "Linting code..."
    golangci-lint run

# Setup development environment
setup:
    @echo "Setting up development environment..."
    go mod download
    go install github.com/a-h/templ/cmd/templ@latest
    npm install
    mkdir -p bin

# Show environment variables loaded from the OS
show-env:
    @echo "Current Environment Variables:"
    @echo "TURSO_URL is $(if [ -n "${TURSO_URL:-}" ]; then echo "set"; else echo "not set"; fi)"
    @echo "TURSO_AUTH_TOKEN is $(if [ -n "${TURSO_AUTH_TOKEN:-}" ]; then echo "set"; else echo "not set"; fi)"


# Reset the local database (development only)
reset-db:
    @echo "Resetting local database..."
    ./scripts/dev/reset-db.sh

# Run all checks (format, lint, test)
check: lint test

# Help message (default if no command specified)
help:
    @echo "WLTrack Justfile commands:"
    @echo "  just              - Clean, generate, and build"
    @echo "  just generate     - Generate templ files"
    @echo "  just build        - Build the Go application and Tailwind CSS (with CGO enabled for SQLite)"
    @echo "  just build-docker - Build container with Docker (for cross-compilation)"
    @echo "  just run          - Run the application locally in development mode"
    @echo "  just dev          - Run in dev mode with live Tailwind CSS reloading"
    @echo "  just run-docker   - Run the application in Docker container (uses OS environment variables)"
    @echo "  just deploy       - Deploy to Azure Container App"
    @echo "  just test         - Run tests"
    @echo "  just test-race    - Run tests with race detection"
    @echo "  just test-coverage - Run tests with coverage report"
    @echo "  just clean        - Clean build artifacts"
    @echo "  just format       - Format the code"
    @echo "  just lint         - Lint the code"
    @echo "  just check        - Run all checks (format, lint, test)"
    @echo "  just setup        - Setup development environment"
    @echo "  just show-env     - Show current OS environment variables"
    @echo "  just reset-db     - Reset the local development database"
    @echo "  just help         - Show this help"
