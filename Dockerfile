# Multi-stage build for ModbusGo
# This Dockerfile creates a minimal container with example servers

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the examples
RUN mkdir -p /app/bin && \
    go build -ldflags="-w -s" -o /app/bin/tcp_server ./examples/tcp_server/main.go && \
    go build -ldflags="-w -s" -o /app/bin/advanced_server ./examples/advanced_server/main.go && \
    go build -ldflags="-w -s" -o /app/bin/tcp_client ./examples/tcp_client/main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN addgroup -g 1000 modbus && \
    adduser -D -u 1000 -G modbus modbus

# Set working directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/bin/* /app/bin/

# Copy documentation
COPY --from=builder /build/README.md /build/DOCUMENTATION.md /build/API_REFERENCE.md /app/docs/

# Change ownership
RUN chown -R modbus:modbus /app

# Switch to non-root user
USER modbus

# Expose MODBUS TCP port
EXPOSE 502 5502

# Default command (can be overridden)
CMD ["/app/bin/advanced_server"]

# Labels
LABEL maintainer="github.com/adibhanna/modbusgo" \
      version="1.0.0" \
      description="ModbusGo - Complete MODBUS Protocol Implementation in Go"