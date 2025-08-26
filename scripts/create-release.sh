#!/bin/bash

# ModbusGo Release Script
# Creates a new release with proper versioning and artifacts

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="ModbusGo"
GITHUB_REPO="adibhanna/modbus-go"

# Function to print colored output
print_info() {
    echo -e "${YELLOW}$1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
    exit 1
}

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    print_error "GitHub CLI (gh) is not installed. Install it from: https://cli.github.com/"
fi

# Get version from user or command line
VERSION=${1:-}
if [ -z "$VERSION" ]; then
    echo "Enter the version number (e.g., v1.0.0):"
    read -r VERSION
fi

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
    print_error "Invalid version format. Use semantic versioning: v1.0.0 or v1.0.0-beta1"
fi

print_info "Creating release $VERSION for $PROJECT_NAME"

# Check if we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ] && [ "$CURRENT_BRANCH" != "master" ]; then
    print_error "You must be on main or master branch to create a release. Current branch: $CURRENT_BRANCH"
fi

# Check for uncommitted changes
if [ -n "$(git status --porcelain)" ]; then
    print_error "You have uncommitted changes. Please commit or stash them first."
fi

# Pull latest changes
print_info "Pulling latest changes..."
git pull origin $CURRENT_BRANCH

# Run tests
print_info "Running tests..."
go test ./... || print_error "Tests failed. Please fix before releasing."
print_success "All tests passed"

# Run linting
print_info "Running linters..."
if command -v golangci-lint &> /dev/null; then
    golangci-lint run ./... || print_error "Linting failed. Please fix before releasing."
    print_success "Linting passed"
else
    print_info "golangci-lint not found, skipping linting"
fi

# Create tag
print_info "Creating git tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"
print_success "Tag created"

# Build release artifacts
print_info "Building release artifacts..."
make clean
make release VERSION=$VERSION || {
    # If make release fails, build manually
    print_info "Make release failed, building manually..."
    
    mkdir -p dist
    
    for GOOS in linux darwin windows; do
        for GOARCH in amd64 arm64; do
            # Skip Windows ARM64
            if [ "$GOOS" = "windows" ] && [ "$GOARCH" = "arm64" ]; then
                continue
            fi
            
            OUTPUT_NAME="modbusgo-$VERSION-$GOOS-$GOARCH"
            print_info "Building $OUTPUT_NAME..."
            
            # Create directory
            mkdir -p "dist/$OUTPUT_NAME"
            
            # Copy documentation
            cp README.md DOCUMENTATION.md API_REFERENCE.md LICENSE "dist/$OUTPUT_NAME/" 2>/dev/null || true
            
            # Build examples
            for example in examples/*/; do
                if [ -f "$example/main.go" ]; then
                    EXAMPLE_NAME=$(basename "$example")
                    EXT=""
                    if [ "$GOOS" = "windows" ]; then
                        EXT=".exe"
                    fi
                    
                    GOOS=$GOOS GOARCH=$GOARCH go build -v \
                        -ldflags "-X main.Version=$VERSION -w -s" \
                        -o "dist/$OUTPUT_NAME/$EXAMPLE_NAME$EXT" \
                        "$example/main.go" || true
                fi
            done
            
            # Create archive
            cd dist
            if [ "$GOOS" = "windows" ]; then
                zip -r "$OUTPUT_NAME.zip" "$OUTPUT_NAME" > /dev/null 2>&1
            else
                tar -czf "$OUTPUT_NAME.tar.gz" "$OUTPUT_NAME" > /dev/null 2>&1
            fi
            cd ..
            
            # Remove directory
            rm -rf "dist/$OUTPUT_NAME"
        done
    done
}

print_success "Artifacts built"

# Generate changelog
print_info "Generating changelog..."
PREVIOUS_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [ -z "$PREVIOUS_TAG" ]; then
    CHANGELOG=$(git log --pretty=format:"* %s (%h)" --no-merges)
else
    CHANGELOG=$(git log --pretty=format:"* %s (%h)" --no-merges ${PREVIOUS_TAG}..HEAD)
fi

# Push tag to GitHub
print_info "Pushing tag to GitHub..."
git push origin "$VERSION"
print_success "Tag pushed"

# Create GitHub release
print_info "Creating GitHub release..."

RELEASE_NOTES="# $PROJECT_NAME $VERSION

## 🚀 Release Highlights

This release includes a complete, production-ready MODBUS implementation with support for all 19 standard function codes.

## ✨ Features

- **Complete Protocol Support**: All standard MODBUS function codes (01-43)
- **Multiple Transports**: TCP/IP, RTU (serial), and ASCII
- **Advanced Features**: File records, FIFO queues, diagnostics, device identification
- **Production Ready**: Comprehensive error handling and recovery
- **Well Tested**: Extensive test coverage (>55%)
- **Zero Dependencies**: Uses only Go standard library

## 📊 Supported Function Codes

| Code | Function | Status |
|------|----------|--------|
| 0x01-0x06 | Basic bit and register operations | ✅ |
| 0x07-0x08 | Exception status and diagnostics | ✅ |
| 0x0B-0x0C | Communication event counters and logs | ✅ |
| 0x0F-0x11 | Multiple coils/registers and server ID | ✅ |
| 0x14-0x18 | File records, mask write, and FIFO | ✅ |
| 0x2B | Encapsulated Interface Transport | ✅ |

## 📝 Changelog

$CHANGELOG

## 📦 Installation

\`\`\`bash
go get github.com/$GITHUB_REPO@$VERSION
\`\`\`

## 🔧 Quick Start

\`\`\`go
// TCP Client
client, err := modbus.NewTCPClient(\"192.168.1.100:502\", 1)
defer client.Close()

values, err := client.ReadHoldingRegisters(100, 10)
\`\`\`

\`\`\`go
// TCP Server
dataStore := modbus.NewDefaultDataStore(10000, 10000, 10000, 10000)
server, err := modbus.NewTCPServer(\":502\", dataStore)
server.Start()
\`\`\`

## 📚 Documentation

- [Complete Documentation](https://github.com/$GITHUB_REPO/blob/main/DOCUMENTATION.md)
- [API Reference](https://github.com/$GITHUB_REPO/blob/main/API_REFERENCE.md)
- [Contributing Guide](https://github.com/$GITHUB_REPO/blob/main/CONTRIBUTING.md)

## 💻 Platform Support

Pre-built binaries are available for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## 🙏 Acknowledgments

Thank you to all contributors and the MODBUS community!

---
*For questions or issues, please visit our [GitHub Issues](https://github.com/$GITHUB_REPO/issues)*"

# Check if this is a pre-release
if [[ $VERSION == *"-"* ]]; then
    PRERELEASE="--prerelease"
else
    PRERELEASE=""
fi

# Create release using GitHub CLI
gh release create "$VERSION" \
    --repo "$GITHUB_REPO" \
    --title "$PROJECT_NAME $VERSION" \
    --notes "$RELEASE_NOTES" \
    $PRERELEASE \
    dist/*.tar.gz \
    dist/*.zip \
    || print_error "Failed to create GitHub release"

print_success "Release $VERSION created successfully!"

# Cleanup
print_info "Cleaning up..."
rm -rf dist/

print_success "Release process complete!"
echo ""
echo "📦 Release URL: https://github.com/$GITHUB_REPO/releases/tag/$VERSION"
echo "📚 Documentation: https://github.com/$GITHUB_REPO/blob/main/DOCUMENTATION.md"
echo ""
echo "Next steps:"
echo "1. Verify the release on GitHub"
echo "2. Update any dependent projects"
echo "3. Announce the release if needed"