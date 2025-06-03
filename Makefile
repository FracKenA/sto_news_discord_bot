# STOBot Makefile
# Star Trek Online Discord Bot - Build & Docker targets for CI/CD

# Variables
APP_NAME := stobot
DOCKER_IMAGE := stobot:latest
GO_VERSION := 1.23.0
CGO_ENABLED := 1

# Build variables
BUILD_DIR := ./bin
CMD_DIR := ./cmd/stobot
MAIN_FILE := $(CMD_DIR)/main.go
BINARY := $(BUILD_DIR)/$(APP_NAME)

# Cross-platform build variables
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)

# Platform-specific binaries
LINUX_AMD64_BINARY := $(BUILD_DIR)/$(APP_NAME)-linux-amd64
LINUX_ARM64_BINARY := $(BUILD_DIR)/$(APP_NAME)-linux-arm64
MACOS_AMD64_BINARY := $(BUILD_DIR)/$(APP_NAME)-darwin-amd64
MACOS_ARM64_BINARY := $(BUILD_DIR)/$(APP_NAME)-darwin-arm64
WINDOWS_AMD64_BINARY := $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe
WINDOWS_ARM64_BINARY := $(BUILD_DIR)/$(APP_NAME)-windows-arm64.exe

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: help build clean test deps fmt lint vet check \
        build-all build-linux build-linux-amd64 build-linux-arm64 \
        build-macos build-macos-amd64 build-macos-arm64 \
        build-windows build-windows-amd64 build-windows-arm64 \
        docker-build docker-build-multiarch docker-build-multiarch-local \
        release package version

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
help:
	@echo "$(BLUE)STOBot - Star Trek Online Discord News Bot$(NC)"
	@echo ""
	@echo "$(YELLOW)Build Commands:$(NC)"
	@echo "  $(GREEN)build$(NC)               Build for current platform"
	@echo "  $(GREEN)build-all$(NC)           Build for all platforms and architectures"
	@echo "  $(GREEN)build-linux$(NC)         Build for Linux (both amd64 and arm64)"
	@echo "  $(GREEN)build-linux-amd64$(NC)   Build for Linux (amd64)"
	@echo "  $(GREEN)build-linux-arm64$(NC)   Build for Linux (arm64)"
	@echo "  $(GREEN)build-macos$(NC)         Build for macOS (both amd64 and arm64)"
	@echo "  $(GREEN)build-macos-amd64$(NC)   Build for macOS (amd64)"
	@echo "  $(GREEN)build-macos-arm64$(NC)   Build for macOS (arm64/M1+)"
	@echo "  $(GREEN)build-windows$(NC)       Build for Windows (both amd64 and arm64)"
	@echo "  $(GREEN)build-windows-amd64$(NC) Build for Windows (amd64)"
	@echo "  $(GREEN)build-windows-arm64$(NC) Build for Windows (arm64)"
	@echo "  $(GREEN)release$(NC)             Create release builds and packages"
	@echo ""
	@echo "$(YELLOW)Docker Commands:$(NC)"
	@echo "  $(GREEN)docker-build$(NC)                 Build Docker image for current platform"
	@echo "  $(GREEN)docker-build-multiarch$(NC)       Build and push multi-arch Docker image"
	@echo "  $(GREEN)docker-build-multiarch-local$(NC) Build multi-arch Docker image locally"
	@echo ""
	@echo "$(YELLOW)Development Commands:$(NC)"
	@echo "  $(GREEN)clean$(NC)               Clean build artifacts"
	@echo "  $(GREEN)test$(NC)                Run tests"
	@echo "  $(GREEN)deps$(NC)                Download and tidy Go dependencies"
	@echo "  $(GREEN)fmt$(NC)                 Format Go code"
	@echo "  $(GREEN)lint$(NC)                Run golangci-lint"
	@echo "  $(GREEN)vet$(NC)                 Run go vet"
	@echo "  $(GREEN)check$(NC)               Run all code quality checks"
	@echo "  $(GREEN)version$(NC)             Show version information"

## build: Build the Go binary for current platform
build: deps
	@echo "$(BLUE)Building $(APP_NAME) for current platform...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(MAIN_FILE)
	@echo "$(GREEN)Build complete: $(BINARY)$(NC)"

## build-all: Build binaries for all platforms and architectures
build-all: build-linux build-macos build-windows
	@echo "$(GREEN)All platform builds complete!$(NC)"
	@echo "$(BLUE)Built binaries:$(NC)"
	@ls -la $(BUILD_DIR)/$(APP_NAME)-*

## build-linux: Build for Linux (both amd64 and arm64)
build-linux: build-linux-amd64 build-linux-arm64

## build-linux-amd64: Build for Linux (amd64)
build-linux-amd64: deps
	@echo "$(BLUE)Building $(APP_NAME) for Linux (amd64)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(LINUX_AMD64_BINARY) $(MAIN_FILE)
	@echo "$(GREEN)Linux amd64 build complete: $(LINUX_AMD64_BINARY)$(NC)"

## build-linux-arm64: Build for Linux (arm64)
build-linux-arm64: deps
	@echo "$(BLUE)Building $(APP_NAME) for Linux (arm64)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(LINUX_ARM64_BINARY) $(MAIN_FILE)
	@echo "$(GREEN)Linux arm64 build complete: $(LINUX_ARM64_BINARY)$(NC)"

## build-macos: Build for macOS (both amd64 and arm64)
build-macos: build-macos-amd64 build-macos-arm64

## build-macos-amd64: Build for macOS (amd64)
build-macos-amd64: deps
	@echo "$(BLUE)Building $(APP_NAME) for macOS (amd64)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@if [ "$$(uname)" = "Darwin" ]; then \
		CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(MACOS_AMD64_BINARY) $(MAIN_FILE); \
		echo "$(GREEN)macOS amd64 build complete: $(MACOS_AMD64_BINARY)$(NC)"; \
	else \
		echo "$(YELLOW)Skipping macOS amd64 build on non-macOS system$(NC)"; \
	fi

## build-macos-arm64: Build for macOS (arm64/M1+)
build-macos-arm64: deps
	@echo "$(BLUE)Building $(APP_NAME) for macOS (arm64)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@if [ "$$(uname)" = "Darwin" ]; then \
		CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(MACOS_ARM64_BINARY) $(MAIN_FILE); \
		echo "$(GREEN)macOS arm64 build complete: $(MACOS_ARM64_BINARY)$(NC)"; \
	else \
		echo "$(YELLOW)Skipping macOS arm64 build on non-macOS system$(NC)"; \
	fi

## build-windows: Build for Windows (both amd64 and arm64)
build-windows: build-windows-amd64 build-windows-arm64

## build-windows-amd64: Build for Windows (amd64)
build-windows-amd64: deps
	@echo "$(BLUE)Building $(APP_NAME) for Windows (amd64)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@if command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1; then \
		CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -ldflags "$(LDFLAGS)" -o $(WINDOWS_AMD64_BINARY) $(MAIN_FILE); \
		echo "$(GREEN)Windows amd64 build complete: $(WINDOWS_AMD64_BINARY)$(NC)"; \
	else \
		echo "$(YELLOW)MinGW-w64 not found. Skipping Windows amd64 build.$(NC)"; \
	fi

## build-windows-arm64: Build for Windows (arm64)
build-windows-arm64: deps
	@echo "$(BLUE)Building $(APP_NAME) for Windows (arm64)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@if command -v aarch64-w64-mingw32-gcc >/dev/null 2>&1; then \
		CGO_ENABLED=1 GOOS=windows GOARCH=arm64 CC=aarch64-w64-mingw32-gcc go build -ldflags "$(LDFLAGS)" -o $(WINDOWS_ARM64_BINARY) $(MAIN_FILE); \
		echo "$(GREEN)Windows arm64 build complete: $(WINDOWS_ARM64_BINARY)$(NC)"; \
	else \
		echo "$(YELLOW)MinGW-w64 for arm64 not found. Skipping Windows arm64 build.$(NC)"; \
	fi

## release: Create release builds with checksums and packages
release: clean build-all
	@echo "$(BLUE)Creating release packages...$(NC)"
	@mkdir -p $(BUILD_DIR)/release
	@cd $(BUILD_DIR) && \
		for binary in $(APP_NAME)-*; do \
			if [ -f "$$binary" ]; then \
				case "$$binary" in \
					*.exe) \
						zip -q release/$${binary%.exe}-$(VERSION).zip $$binary ;; \
					*) \
						tar -czf release/$$binary-$(VERSION).tar.gz $$binary ;; \
				esac; \
			fi; \
		done
	@cd $(BUILD_DIR)/release && sha256sum * > checksums.sha256
	@echo "$(GREEN)Release packages created in $(BUILD_DIR)/release/$(NC)"
	@echo "$(YELLOW)Files:$(NC)"
	@ls -la $(BUILD_DIR)/release/

## package: Create distribution packages (alias for release)
package: release

## clean: Clean build artifacts and temporary files
clean:
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "$(GREEN)Clean complete$(NC)"

## test: Run all tests
test:
	@echo "$(BLUE)Running tests...$(NC)"
	@go test -v ./...
	@echo "$(GREEN)Tests complete$(NC)"

## deps: Download and tidy Go dependencies
deps:
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

## fmt: Format Go code
fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)Code formatted$(NC)"

## lint: Run golangci-lint (requires golangci-lint to be installed)
lint:
	@echo "$(BLUE)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint not installed, skipping...$(NC)"; \
	fi

## vet: Run go vet
vet:
	@echo "$(BLUE)Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)Vet complete$(NC)"

## check: Run all code quality checks
check: fmt vet lint test

## version: Show version information
version:
	@echo "$(BLUE)Version Information:$(NC)"
	@echo "  Version:    $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Go Version: $(shell go version)"

## docker-build: Build Docker image for current platform
docker-build:
	@echo "$(BLUE)Building Docker image for current platform...$(NC)"
	@docker build -t $(DOCKER_IMAGE) .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE)$(NC)"

## docker-build-multiarch: Build multi-architecture Docker image using buildx
docker-build-multiarch:
	@echo "$(BLUE)Building multi-architecture Docker image...$(NC)"
	@echo "$(YELLOW)Creating buildx builder if it doesn't exist...$(NC)"
	@docker buildx create --use --name multiarch-builder 2>/dev/null || true
	@echo "$(BLUE)Building for platforms: linux/amd64, linux/arm64...$(NC)"
	@docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--file Dockerfile \
		--tag $(DOCKER_IMAGE) \
		--push .
	@echo "$(GREEN)Multi-architecture Docker image built and pushed: $(DOCKER_IMAGE)$(NC)"

## docker-build-multiarch-local: Build multi-architecture Docker image locally (no push)
docker-build-multiarch-local:
	@echo "$(BLUE)Building Docker image for current platform locally...$(NC)"
	@echo "$(YELLOW)Note: Building for current platform only (multi-arch requires push)$(NC)"
	@docker buildx build \
		--file Dockerfile \
		--tag $(DOCKER_IMAGE) \
		--load .
	@echo "$(GREEN)Docker image built locally: $(DOCKER_IMAGE)$(NC)"
