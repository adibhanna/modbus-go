#!/bin/bash

# Build release artifacts for ModbusGo

set -e

VERSION=${1:-v1.0.0}
DIST_DIR="dist"

echo "Building release artifacts for $VERSION..."

# Clean and create dist directory
rm -rf $DIST_DIR
mkdir -p $DIST_DIR

# Build for Darwin ARM64 (current platform)
echo "Building for darwin/arm64..."
OUTPUT_NAME="modbusgo-$VERSION-darwin-arm64"
mkdir -p "$DIST_DIR/$OUTPUT_NAME"

# Copy documentation
cp README.md DOCUMENTATION.md API_REFERENCE.md "$DIST_DIR/$OUTPUT_NAME/" 2>/dev/null || true

# Build examples
go build -ldflags "-w -s" -o "$DIST_DIR/$OUTPUT_NAME/tcp_server" ./examples/tcp_server/main.go
go build -ldflags "-w -s" -o "$DIST_DIR/$OUTPUT_NAME/tcp_client" ./examples/tcp_client/main.go
go build -ldflags "-w -s" -o "$DIST_DIR/$OUTPUT_NAME/advanced_server" ./examples/advanced_server/main.go

# Create archive
cd $DIST_DIR
tar -czf "$OUTPUT_NAME.tar.gz" "$OUTPUT_NAME"
rm -rf "$OUTPUT_NAME"
cd ..

echo "✓ Built $OUTPUT_NAME.tar.gz"

# Build for Linux AMD64
echo "Building for linux/amd64..."
OUTPUT_NAME="modbusgo-$VERSION-linux-amd64"
mkdir -p "$DIST_DIR/$OUTPUT_NAME"

# Copy documentation
cp README.md DOCUMENTATION.md API_REFERENCE.md "$DIST_DIR/$OUTPUT_NAME/" 2>/dev/null || true

# Build examples
GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o "$DIST_DIR/$OUTPUT_NAME/tcp_server" ./examples/tcp_server/main.go
GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o "$DIST_DIR/$OUTPUT_NAME/tcp_client" ./examples/tcp_client/main.go
GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o "$DIST_DIR/$OUTPUT_NAME/advanced_server" ./examples/advanced_server/main.go

# Create archive
cd $DIST_DIR
tar -czf "$OUTPUT_NAME.tar.gz" "$OUTPUT_NAME"
rm -rf "$OUTPUT_NAME"
cd ..

echo "✓ Built $OUTPUT_NAME.tar.gz"

echo ""
echo "Release artifacts built successfully:"
ls -lh $DIST_DIR/*.tar.gz

echo ""
echo "To upload to GitHub release:"
echo "gh release upload $VERSION $DIST_DIR/*.tar.gz --repo adibhanna/modbus-go"