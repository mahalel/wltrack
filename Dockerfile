# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies in a single layer
RUN apk add --no-cache gcc musl-dev git nodejs npm

# Set working directory
WORKDIR /app

# Copy go.mod, go.sum, and package.json files first for better caching
COPY go.mod go.sum package.json package-lock.json ./
RUN go mod download
RUN npm install

# Install templ before copying the rest of the files for better caching
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy all necessary code files
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY src/ ./src/
COPY tailwind.config.js ./
COPY scripts/ ./scripts/

# Generate templates after all source code is copied
RUN templ generate

# No need to copy code files again as they're already copied above

# Build JavaScript bundle first
RUN npm run build:js

# Run Tailwind CSS build after template generation to ensure it picks up all classes
RUN npm run build:css

# Build the application with CGO enabled and optimized flags
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -tags=sqlite_omit_load_extension \
    -ldflags="-s -w" \
    -trimpath \
    -o /app/bin/server ./cmd/server

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies in a single layer
RUN apk add --no-cache ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/bin/server /server

# Copy static files from host
COPY static/ /static/

# Copy generated files from builder stage
COPY --from=builder /app/static/css/tailwind.css /static/css/tailwind.css
COPY --from=builder /app/static/js/main.js /static/js/main.js

# Set environment variables
ENV PORT=8080

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["/server"]
