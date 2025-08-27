# ModbusGo Makefile
# Comprehensive build and development automation

# Variables
BINARY_NAME=modbusgo
PACKAGE=github.com/adibhanna/modbusgo
GO=go
GOFLAGS=-v
GOCMD=$(GO)
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint
GOVET=$(GOCMD) vet

# Version information
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)"

# Directories
CMD_DIR=./cmd
EXAMPLES_DIR=./examples
BUILD_DIR=./build
DIST_DIR=./dist
COVERAGE_DIR=./coverage

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

# Default target
.DEFAULT_GOAL := help

# Phony targets
.PHONY: all build clean test coverage bench lint fmt vet security help \
        install uninstall deps update-deps vendor examples docker \
        proto docs serve-docs release integration-test \
        check-tools install-tools clean-cache

## help: Show this help message
help:
	@echo "ModbusGo Makefile Commands:"
	@echo ""
	@grep -E '^##' Makefile | sed -e 's/##//' | column -t -s ':'
	@echo ""
	@echo "Examples:"
	@echo "  make build          # Build all binaries"
	@echo "  make test          # Run all tests"
	@echo "  make coverage      # Generate coverage report"
	@echo "  make lint          # Run linters"
	@echo "  make examples      # Build example programs"

## all: Build and test everything
all: clean deps lint test build examples
	@echo "$(GREEN)✓ All tasks completed successfully$(NC)"

## build: Build the library
build:
	@echo "$(YELLOW)Building ModbusGo library...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/ ./...
	@echo "$(GREEN)✓ Build complete$(NC)"

## examples: Build all example programs
examples:
	@echo "$(YELLOW)Building examples...$(NC)"
	@mkdir -p $(BUILD_DIR)/examples
	@for dir in $(EXAMPLES_DIR)/*/; do \
		if [ -f $$dir/main.go ]; then \
			name=$$(basename $$dir); \
			echo "  Building $$name..."; \
			$(GOBUILD) $(GOFLAGS) -o $(BUILD_DIR)/examples/$$name $$dir/main.go || exit 1; \
		fi; \
	done
	@echo "$(GREEN)✓ Examples built successfully$(NC)"

## test: Run all tests
test:
	@echo "$(YELLOW)Running tests...$(NC)"
	@$(GOTEST) $(GOFLAGS) -race -short $$($(GOCMD) list ./... | grep -v /examples/)
	@echo "$(GREEN)✓ Tests passed$(NC)"

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "$(YELLOW)Running tests (verbose)...$(NC)"
	@$(GOTEST) -v -race $$($(GOCMD) list ./... | grep -v /examples/)

## integration-test: Run integration tests
integration-test:
	@echo "$(YELLOW)Running integration tests...$(NC)"
	@if [ -f ./test_integration.sh ]; then \
		./test_integration.sh; \
	else \
		$(GOTEST) $(GOFLAGS) -tags=integration -race ./...; \
	fi
	@echo "$(GREEN)✓ Integration tests passed$(NC)"

## coverage: Generate test coverage report
coverage:
	@echo "$(YELLOW)Generating coverage report...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@$(GOTEST) -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic $$($(GOCMD) list ./... | grep -v /examples/)
	@$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@$(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out
	@echo "$(GREEN)✓ Coverage report generated at $(COVERAGE_DIR)/coverage.html$(NC)"

## coverage-view: View coverage report in browser
coverage-view: coverage
	@echo "$(YELLOW)Opening coverage report...$(NC)"
	@if command -v xdg-open > /dev/null; then \
		xdg-open $(COVERAGE_DIR)/coverage.html; \
	elif command -v open > /dev/null; then \
		open $(COVERAGE_DIR)/coverage.html; \
	else \
		echo "Please open $(COVERAGE_DIR)/coverage.html manually"; \
	fi

## bench: Run benchmarks
bench:
	@echo "$(YELLOW)Running benchmarks...$(NC)"
	@$(GOTEST) -bench=. -benchmem -benchtime=10s ./...
	@echo "$(GREEN)✓ Benchmarks complete$(NC)"

## bench-compare: Run benchmarks and compare with previous results
bench-compare:
	@echo "$(YELLOW)Running benchmark comparison...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(GOTEST) -bench=. -benchmem -benchtime=10s ./... > $(BUILD_DIR)/bench-new.txt
	@if [ -f $(BUILD_DIR)/bench-old.txt ]; then \
		benchcmp $(BUILD_DIR)/bench-old.txt $(BUILD_DIR)/bench-new.txt; \
	else \
		echo "No previous benchmark results found"; \
	fi
	@mv $(BUILD_DIR)/bench-new.txt $(BUILD_DIR)/bench-old.txt

## lint: Run linters
lint: check-tools
	@echo "$(YELLOW)Running linters...$(NC)"
	@if command -v $(GOLINT) > /dev/null; then \
		$(GOLINT) run . ./config ./modbus ./pdu ./transport; \
	else \
		echo "$(YELLOW)golangci-lint not found, using go vet$(NC)"; \
		$(GOVET) . ./config ./modbus ./pdu ./transport; \
	fi
	@echo "$(GREEN)✓ Lint checks passed$(NC)"

## fmt: Format code
fmt:
	@echo "$(YELLOW)Formatting code...$(NC)"
	@$(GOFMT) -s -w .
	@$(GOCMD) fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	fi
	@echo "$(GREEN)✓ Code formatted$(NC)"

## fmt-check: Check code formatting
fmt-check:
	@echo "$(YELLOW)Checking code formatting...$(NC)"
	@if [ -n "$$($(GOFMT) -l .)" ]; then \
		echo "$(RED)✗ Code needs formatting. Run 'make fmt'$(NC)"; \
		$(GOFMT) -l .; \
		exit 1; \
	fi
	@echo "$(GREEN)✓ Code formatting is correct$(NC)"

## vet: Run go vet
vet:
	@echo "$(YELLOW)Running go vet...$(NC)"
	@$(GOVET) . ./config ./modbus ./pdu ./transport
	@echo "$(GREEN)✓ Vet checks passed$(NC)"

## security: Run security checks
security:
	@echo "$(YELLOW)Running security checks...$(NC)"
	@if command -v gosec > /dev/null; then \
		gosec -quiet ./...; \
	else \
		echo "$(YELLOW)gosec not installed, install with: go install github.com/securego/gosec/v2/cmd/gosec@latest$(NC)"; \
	fi
	@if command -v nancy > /dev/null; then \
		$(GOCMD) list -json -deps ./... | nancy sleuth; \
	else \
		echo "$(YELLOW)nancy not installed, install with: go install github.com/sonatype-nexus-community/nancy@latest$(NC)"; \
	fi
	@echo "$(GREEN)✓ Security checks complete$(NC)"

## deps: Download dependencies
deps:
	@echo "$(YELLOW)Downloading dependencies...$(NC)"
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "$(GREEN)✓ Dependencies ready$(NC)"

## update-deps: Update dependencies to latest versions
update-deps:
	@echo "$(YELLOW)Updating dependencies...$(NC)"
	@$(GOCMD) get -u ./...
	@$(GOMOD) tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

## vendor: Create vendor directory
vendor:
	@echo "$(YELLOW)Creating vendor directory...$(NC)"
	@$(GOMOD) vendor
	@echo "$(GREEN)✓ Vendor directory created$(NC)"

## clean: Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR) $(DIST_DIR) $(COVERAGE_DIR)
	@rm -f *.out *.prof *.test
	@echo "$(GREEN)✓ Clean complete$(NC)"

## clean-cache: Clean Go build and test cache
clean-cache: clean
	@echo "$(YELLOW)Cleaning Go cache...$(NC)"
	@$(GOCLEAN) -cache
	@$(GOCLEAN) -testcache
	@$(GOCLEAN) -modcache
	@echo "$(GREEN)✓ Cache cleaned$(NC)"

## install: Install the library
install:
	@echo "$(YELLOW)Installing ModbusGo...$(NC)"
	@$(GOCMD) install $(GOFLAGS) $(LDFLAGS) ./...
	@echo "$(GREEN)✓ Installation complete$(NC)"

## uninstall: Uninstall the library
uninstall:
	@echo "$(YELLOW)Uninstalling ModbusGo...$(NC)"
	@$(GOCLEAN) -i ./...
	@echo "$(GREEN)✓ Uninstallation complete$(NC)"

## docs: Generate documentation
docs:
	@echo "$(YELLOW)Generating documentation...$(NC)"
	@if command -v godoc > /dev/null; then \
		echo "Documentation available at http://localhost:6060/pkg/$(PACKAGE)/"; \
		echo "Run 'make serve-docs' to start the documentation server"; \
	else \
		echo "$(YELLOW)godoc not installed, install with: go install golang.org/x/tools/cmd/godoc@latest$(NC)"; \
		$(GOCMD) doc -all ./...; \
	fi

## serve-docs: Serve documentation locally
serve-docs:
	@echo "$(YELLOW)Starting documentation server...$(NC)"
	@if command -v godoc > /dev/null; then \
		echo "$(GREEN)Documentation server starting at http://localhost:6060/pkg/$(PACKAGE)/$(NC)"; \
		godoc -http=:6060; \
	else \
		echo "$(RED)godoc not installed, install with: go install golang.org/x/tools/cmd/godoc@latest$(NC)"; \
		exit 1; \
	fi

## proto: Generate protobuf code (if applicable)
proto:
	@echo "$(YELLOW)Generating protobuf code...$(NC)"
	@if [ -d ./proto ]; then \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			proto/*.proto; \
		echo "$(GREEN)✓ Protobuf generation complete$(NC)"; \
	else \
		echo "No proto files found"; \
	fi

## docker: Build Docker image
docker:
	@echo "$(YELLOW)Building Docker image...$(NC)"
	@if [ -f Dockerfile ]; then \
		docker build -t $(BINARY_NAME):$(VERSION) .; \
		docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest; \
		echo "$(GREEN)✓ Docker image built: $(BINARY_NAME):$(VERSION)$(NC)"; \
	else \
		echo "$(RED)Dockerfile not found$(NC)"; \
		exit 1; \
	fi

## docker-run: Run Docker container
docker-run:
	@echo "$(YELLOW)Running Docker container...$(NC)"
	@docker run -it --rm -p 502:502 $(BINARY_NAME):latest

## release: Create release artifacts
release: clean test
	@echo "$(YELLOW)Creating release artifacts...$(NC)"
	@mkdir -p $(DIST_DIR)
	
	@# Build for multiple platforms
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			echo "  Building for $$os/$$arch..."; \
			output_name=$(BINARY_NAME); \
			if [ $$os = "windows" ]; then output_name=$(BINARY_NAME).exe; fi; \
			GOOS=$$os GOARCH=$$arch $(GOBUILD) $(LDFLAGS) \
				-o $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-$$os-$$arch/$$output_name ./...; \
		done; \
	done
	
	@# Create archives
	@for dir in $(DIST_DIR)/*; do \
		if [ -d $$dir ]; then \
			base=$$(basename $$dir); \
			echo "  Creating archive $$base.tar.gz..."; \
			tar -czf $(DIST_DIR)/$$base.tar.gz -C $(DIST_DIR) $$base; \
			rm -rf $$dir; \
		fi; \
	done
	
	@echo "$(GREEN)✓ Release artifacts created in $(DIST_DIR)$(NC)"

## check-tools: Check if required tools are installed
check-tools:
	@echo "$(YELLOW)Checking required tools...$(NC)"
	@command -v $(GO) >/dev/null 2>&1 || { echo "$(RED)go is not installed$(NC)"; exit 1; }
	@command -v git >/dev/null 2>&1 || { echo "$(RED)git is not installed$(NC)"; exit 1; }
	@echo "$(GREEN)✓ Required tools are installed$(NC)"

## install-tools: Install development tools
install-tools:
	@echo "$(YELLOW)Installing development tools...$(NC)"
	
	@# Install golangci-lint
	@if ! command -v golangci-lint > /dev/null; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
	fi
	
	@# Install other tools
	@echo "Installing Go tools..."
	@$(GOCMD) install golang.org/x/tools/cmd/goimports@latest
	@$(GOCMD) install golang.org/x/tools/cmd/godoc@latest
	@$(GOCMD) install github.com/securego/gosec/v2/cmd/gosec@latest
	@$(GOCMD) install github.com/sonatype-nexus-community/nancy@latest
	@$(GOCMD) install golang.org/x/perf/cmd/benchstat@latest
	@$(GOCMD) install github.com/kisielk/errcheck@latest
	@$(GOCMD) install honnef.co/go/tools/cmd/staticcheck@latest
	
	@echo "$(GREEN)✓ Development tools installed$(NC)"

## info: Show environment information
info:
	@echo "$(YELLOW)Environment Information:$(NC)"
	@echo "  Go Version:    $$($(GOCMD) version)"
	@echo "  Git Version:   $$(git --version)"
	@echo "  Project:       $(PACKAGE)"
	@echo "  Version:       $(VERSION)"
	@echo "  Commit:        $(COMMIT)"
	@echo "  Build Date:    $(BUILD_DATE)"
	@echo "  OS/Arch:       $$($(GOCMD) env GOOS)/$$($(GOCMD) env GOARCH)"
	@echo "  GOPATH:        $$($(GOCMD) env GOPATH)"
	@echo "  GOROOT:        $$($(GOCMD) env GOROOT)"

## run-tcp-server: Run TCP server example
run-tcp-server: build examples
	@echo "$(YELLOW)Starting TCP Server on port 5502...$(NC)"
	@$(BUILD_DIR)/examples/tcp_server

## run-tcp-client: Run TCP client example
run-tcp-client: build examples
	@echo "$(YELLOW)Starting TCP Client...$(NC)"
	@$(BUILD_DIR)/examples/tcp_client

## run-advanced-server: Run advanced server example
run-advanced-server: build examples
	@echo "$(YELLOW)Starting Advanced Server on port 5502...$(NC)"
	@$(BUILD_DIR)/examples/advanced_server

## watch: Watch for changes and run tests
watch:
	@echo "$(YELLOW)Watching for changes...$(NC)"
	@if command -v fswatch > /dev/null; then \
		fswatch -o . -e ".*" -i "\\.go$$" | xargs -n1 -I{} sh -c 'clear; make test'; \
	elif command -v inotifywait > /dev/null; then \
		while true; do \
			inotifywait -r -e modify --include '\.go$$' .; \
			clear; \
			make test; \
		done; \
	else \
		echo "$(RED)No file watcher found. Install fswatch or inotify-tools$(NC)"; \
		exit 1; \
	fi

## ci: Run continuous integration checks
ci: clean deps fmt-check lint vet test coverage
	@echo "$(GREEN)✓ CI checks passed$(NC)"

## pre-commit: Run pre-commit checks
pre-commit: fmt lint vet test
	@echo "$(GREEN)✓ Pre-commit checks passed$(NC)"

## stats: Show code statistics
stats:
	@echo "$(YELLOW)Code Statistics:$(NC)"
	@echo ""
	@echo "Lines of Code:"
	@find . -name '*.go' -not -path "./vendor/*" -not -path "./.git/*" | xargs wc -l | tail -1
	@echo ""
	@echo "File Count:"
	@find . -name '*.go' -not -path "./vendor/*" -not -path "./.git/*" | wc -l
	@echo ""
	@echo "Package Count:"
	@$(GOCMD) list ./... | wc -l
	@echo ""
	@echo "Test Coverage:"
	@$(GOTEST) -cover ./... | grep -E "^ok|^FAIL" | awk '{print $$2, $$4, $$5}' | column -t

# Performance profiling targets
## profile-cpu: Generate CPU profile
profile-cpu:
	@echo "$(YELLOW)Generating CPU profile...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(GOTEST) -cpuprofile=$(BUILD_DIR)/cpu.prof -bench=. ./...
	@echo "$(GREEN)✓ CPU profile saved to $(BUILD_DIR)/cpu.prof$(NC)"
	@echo "View with: go tool pprof $(BUILD_DIR)/cpu.prof"

## profile-mem: Generate memory profile
profile-mem:
	@echo "$(YELLOW)Generating memory profile...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(GOTEST) -memprofile=$(BUILD_DIR)/mem.prof -bench=. ./...
	@echo "$(GREEN)✓ Memory profile saved to $(BUILD_DIR)/mem.prof$(NC)"
	@echo "View with: go tool pprof $(BUILD_DIR)/mem.prof"

## profile-view-cpu: View CPU profile
profile-view-cpu: profile-cpu
	@$(GOCMD) tool pprof -http=:8080 $(BUILD_DIR)/cpu.prof

## profile-view-mem: View memory profile  
profile-view-mem: profile-mem
	@$(GOCMD) tool pprof -http=:8080 $(BUILD_DIR)/mem.prof

# Quick command aliases
t: test
b: build
c: clean
l: lint
f: fmt
cov: coverage