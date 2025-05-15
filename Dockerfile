# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies in a single layer
RUN apk add --no-cache gcc musl-dev git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Install templ before copying the rest of the files for better caching
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy only necessary files for template generation
COPY internal/templates/ ./internal/templates/
RUN templ generate

# Copy the rest of the application source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

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

# Copy static files
COPY static/ /static/

# Set environment variables
ENV PORT=8080

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["/server"]
