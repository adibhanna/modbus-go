# Multi-stage build for ModbusGo
# This Dockerfile creates containers for running MODBUS servers and development

# =============================================================================
# Build stage - compiles all binaries
# =============================================================================
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build all examples
RUN mkdir -p /app/bin && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/tcp_server ./examples/tcp_server/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/advanced_server ./examples/advanced_server/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/tcp_client ./examples/tcp_client/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/integration_test ./examples/integration_test/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/config_showcase ./examples/config_showcase/main.go

# =============================================================================
# Runtime stage - minimal image for production
# =============================================================================
FROM alpine:3.19 AS runtime

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 modbus && \
    adduser -D -u 1000 -G modbus modbus

# Set working directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/bin/* /app/bin/

# Copy documentation
COPY --from=builder /build/README.md /app/docs/
COPY --from=builder /build/docs/ /app/docs/

# Copy example configs
COPY --from=builder /build/config-examples/ /app/config-examples/

# Change ownership
RUN chown -R modbus:modbus /app

# Switch to non-root user
USER modbus

# Expose MODBUS TCP ports
EXPOSE 502 5502

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD nc -z localhost 5502 || exit 1

# Default command
CMD ["/app/bin/advanced_server"]

# =============================================================================
# Development stage - full development environment
# =============================================================================
FROM golang:1.25-alpine AS development

# Install development tools
RUN apk add --no-cache git make curl bash netcat-openbsd

# Install golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin v2.7.2

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code (will be overridden by volume mount in docker-compose)
COPY . .

# Expose ports
EXPOSE 502 5502

# Default command for development
CMD ["make", "dev"]

# =============================================================================
# Test stage - runs all tests
# =============================================================================
FROM golang:1.25-alpine AS test

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Run tests with race detector
CMD ["go", "test", "-v", "-race", "-coverprofile=coverage.txt", "./..."]

# =============================================================================
# CI stage - runs full CI pipeline
# =============================================================================
FROM golang:1.25-alpine AS ci

RUN apk add --no-cache git make curl bash

# Install golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin v2.7.2

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Run CI checks
CMD ["make", "ci"]

# Labels
LABEL maintainer="github.com/adibhanna/modbus-go" \
      version="1.0.0" \
      description="ModbusGo - Complete MODBUS Protocol Implementation in Go"
