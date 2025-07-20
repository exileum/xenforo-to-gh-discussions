# GoDisc - Go Downloader with Disc Conversion
# Makefile for building, testing, and managing the project

# Project configuration
BINARY_NAME := xenforo-to-gh-discussions
MODULE_NAME := github.com/exileum/xenforo-to-gh-discussions
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go configuration
GO := go
GOFMT := gofmt
GOVET := $(GO) vet
GOLINT := golangci-lint
GO_VERSION := $(shell $(GO) version | cut -d' ' -f3)

# Build configuration
BUILD_DIR := build
DIST_DIR := dist
COVERAGE_DIR := coverage

# Linker flags for version info
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commitHash=$(COMMIT_HASH)"
LDFLAGS_RELEASE := -ldflags "-w -s -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commitHash=$(COMMIT_HASH)"

# Platform detection
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
BINARY_EXT :=

# Colors for output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
MAGENTA := \033[35m
CYAN := \033[36m
WHITE := \033[37m
RESET := \033[0m

# Default target
.DEFAULT_GOAL := help

# Help target - self-documenting Makefile
.PHONY: help
help: ## Show this help message
	@echo "$(CYAN)GoDisc - Available Make Targets$(RESET)"
	@echo
	@awk 'BEGIN {FS = ":.*##"; printf "$(BLUE)Usage:$(RESET)\n  make $(GREEN)<target>$(RESET)\n\n$(BLUE)Targets:$(RESET)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Development
.PHONY: build
build: ## Build the binary
	@echo "$(CYAN)Building $(BINARY_NAME)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS_RELEASE) -o $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) ./cmd/xenforo-to-gh-discussions
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT)$(RESET)"

.PHONY: dev
dev: ## Build for development (with race detector)
	@echo "$(CYAN)Building development version...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build -race $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-dev$(BINARY_EXT) ./cmd/xenforo-to-gh-discussions
	@echo "$(GREEN)Development build complete: $(BUILD_DIR)/$(BINARY_NAME)-dev$(BINARY_EXT)$(RESET)"

.PHONY: install
install: ## Install binary to $GOPATH/bin
	@echo "$(CYAN)Installing $(BINARY_NAME)...$(RESET)"
	$(GO) install $(LDFLAGS) ./cmd/xenforo-to-gh-discussions
	@echo "$(GREEN)Installation complete$(RESET)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(CYAN)Cleaning build artifacts...$(RESET)"
	rm -rf $(BUILD_DIR) $(DIST_DIR) $(COVERAGE_DIR)
	$(GO) clean -cache -testcache -modcache
	@echo "$(GREEN)Clean complete$(RESET)"

##@ Testing
.PHONY: test
test: ## Run all tests
	@echo "$(CYAN)Running all tests...$(RESET)"
	$(GO) test -v ./...
	@echo "$(GREEN)All tests passed$(RESET)"

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "$(CYAN)Running unit tests...$(RESET)"
	$(GO) test -v -short ./...
	@echo "$(GREEN)Unit tests passed$(RESET)"

.PHONY: test-integration
test-integration: ## Run integration tests only
	@echo "$(CYAN)Running integration tests...$(RESET)"
	$(GO) test -v -run Integration ./...
	@echo "$(GREEN)Integration tests passed$(RESET)"

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(CYAN)Running tests with coverage...$(RESET)"
	@mkdir -p $(COVERAGE_DIR)
	$(GO) test -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_DIR)/coverage.html$(RESET)"

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "$(CYAN)Running tests with race detector...$(RESET)"
	$(GO) test -race -short ./...
	@echo "$(GREEN)Race tests passed$(RESET)"

.PHONY: bench
bench: ## Run benchmarks
	@echo "$(CYAN)Running benchmarks...$(RESET)"
	$(GO) test -bench=. -benchmem ./...

##@ Code Quality
.PHONY: lint
lint: ## Run code quality checks
	@echo "$(CYAN)Running code quality checks...$(RESET)"
	@GOFMT_TEMP=$$(mktemp) && \
	$(GOFMT) -d -s . | tee "$$GOFMT_TEMP" && \
	if [ -s "$$GOFMT_TEMP" ]; then \
		echo "$(RED)Code formatting issues found. Run 'make fmt' to fix.$(RESET)"; \
		rm -f "$$GOFMT_TEMP"; \
		exit 1; \
	fi; \
	rm -f "$$GOFMT_TEMP"
	@$(GOVET) ./...
	@$(GO) mod verify
	@echo "$(GREEN)Code quality checks passed$(RESET)"

.PHONY: fmt
fmt: ## Format code
	@echo "$(CYAN)Formatting code...$(RESET)"
	$(GOFMT) -w -s .
	@echo "$(GREEN)Code formatted$(RESET)"

.PHONY: golangci-lint
golangci-lint: ## Run golangci-lint (requires golangci-lint to be installed)
	@echo "$(CYAN)Running golangci-lint...$(RESET)"
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run; \
		echo "$(GREEN)golangci-lint checks passed$(RESET)"; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(RESET)"; \
	fi

.PHONY: cyclo
cyclo: ## Run cyclomatic complexity analysis (requires gocyclo to be installed)
	@echo "$(CYAN)Running cyclomatic complexity analysis...$(RESET)"
	@if command -v gocyclo > /dev/null 2>&1; then \
		CYCLO_OUTPUT=$$(gocyclo -over 10 . 2>&1); \
		if [ -n "$$CYCLO_OUTPUT" ]; then \
			echo "$(YELLOW)Functions with cyclomatic complexity > 10:$(RESET)"; \
			echo "$$CYCLO_OUTPUT"; \
			echo "$(YELLOW)Consider refactoring complex functions$(RESET)"; \
		else \
			echo "$(GREEN)No functions with complexity > 10 found$(RESET)"; \
		fi; \
	else \
		echo "$(YELLOW)gocyclo not installed. Install with: go install github.com/fzipp/gocyclo/cmd/gocyclo@latest$(RESET)"; \
	fi

##@ Dependencies
.PHONY: deps
deps: ## Download and verify dependencies
	@echo "$(CYAN)Downloading dependencies...$(RESET)"
	$(GO) mod download
	$(GO) mod verify
	@echo "$(GREEN)Dependencies updated$(RESET)"

.PHONY: tidy
tidy: ## Clean up go.mod and go.sum
	@echo "$(CYAN)Tidying dependencies...$(RESET)"
	$(GO) mod tidy
	@echo "$(GREEN)Dependencies tidied$(RESET)"

.PHONY: deps-update
deps-update: ## Update all dependencies
	@echo "$(CYAN)Updating dependencies...$(RESET)"
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo "$(GREEN)Dependencies updated$(RESET)"

##@ Release
.PHONY: build-all
build-all: ## Build for all platforms
	@echo "$(CYAN)Building for all platforms...$(RESET)"
	@mkdir -p $(DIST_DIR)
	@for os in linux darwin; do \
		for arch in amd64 arm64; do \
			echo "Building $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch $(GO) build $(LDFLAGS_RELEASE) \
				-o $(DIST_DIR)/$(BINARY_NAME)-$$os-$$arch ./cmd/xenforo-to-gh-discussions; \
		done; \
	done
	@echo "$(GREEN)Cross-platform builds complete in $(DIST_DIR)/$(RESET)"

.PHONY: release
release: clean lint test build-all ## Create a release (clean, lint, test, build-all)
	@echo "$(GREEN)Release build complete$(RESET)"

.PHONY: package
package: build-all ## Create release packages
	@echo "$(CYAN)Creating release packages...$(RESET)"
	@cd $(DIST_DIR) && \
	for binary in $(BINARY_NAME)-*; do \
		if [[ $$binary == *".exe" ]]; then \
			zip $$binary.zip $$binary; \
		else \
			tar -czf $$binary.tar.gz $$binary; \
		fi; \
	done
	@echo "$(GREEN)Release packages created in $(DIST_DIR)/$(RESET)"

##@ Utilities
.PHONY: version
version: ## Show version information
	@echo "$(CYAN)Version Information:$(RESET)"
	@echo "  Binary: $(BINARY_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Commit: $(COMMIT_HASH)"
	@echo "  Go Version: $(GO_VERSION)"
	@echo "  Platform: $(GOOS)/$(GOARCH)"

.PHONY: check
check: lint cyclo test ## Run pre-commit checks (lint + cyclo + test)
	@echo "$(GREEN)All checks passed - ready to commit!$(RESET)"

.PHONY: run
run: build ## Build and run the application (use ARGS="--flag" to pass arguments)
	@echo "$(CYAN)Running $(BINARY_NAME)...$(RESET)"
	./$(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) $(ARGS)

.PHONY: watch
watch: ## Auto-rebuild on file changes (requires entr)
	@if command -v entr > /dev/null 2>&1; then \
		echo "$(CYAN)Watching for changes... (Ctrl+C to stop)$(RESET)"; \
		find . -name "*.go" | entr -r make build; \
	else \
		echo "$(RED)entr not found. Install with: brew install entr (macOS) or apt-get install entr (Linux)$(RESET)"; \
	fi

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(CYAN)Building Docker image...$(RESET)"
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest
	@echo "$(GREEN)Docker image built: $(BINARY_NAME):$(VERSION)$(RESET)"

.PHONY: docker-run
docker-run: docker-build ## Build and run Docker container
	@echo "$(CYAN)Running Docker container...$(RESET)"
	docker run --rm -it $(BINARY_NAME):latest

# Include additional makefiles if they exist
-include Makefile.local
