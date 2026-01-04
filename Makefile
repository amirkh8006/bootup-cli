# Bootup CLI Makefile

.PHONY: build clean install uninstall cross-compile fmt lint deps


# Variables
BINARY_NAME = bootup
VERSION ?= $(shell git describe --tags --abbrev=0)
LDFLAGS = -ldflags="-s -w -X github.com/amirkh8006/bootup-cli/cmd.Version=${VERSION}"
BUILD_DIR = build
INSTALL_DIR = /usr/local/bin

# Default target
.DEFAULT_GOAL := build

help:
	@echo "Usage:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS=":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'


build: ## Build the binary for current platform
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o ${BINARY_NAME} ./cmd/bootup

run: build ## Run the application
	@echo "Running ${BINARY_NAME}..."
	./${BINARY_NAME}

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf ${BUILD_DIR}
	rm -f ${BINARY_NAME}
	rm -f ${BINARY_NAME}-*


install: build ## Install binary to system
	@echo "Installing ${BINARY_NAME} to ${INSTALL_DIR}..."
	sudo cp ${BINARY_NAME} ${INSTALL_DIR}/
	@echo "✓ ${BINARY_NAME} installed successfully!"

uninstall: ## Uninstall binary from system
	@echo "Uninstalling ${BINARY_NAME}..."
	@BINARY_PATH=$$(which ${BINARY_NAME} 2>/dev/null); \
	if [ -n "$$BINARY_PATH" ]; then \
		echo "Found ${BINARY_NAME} at $$BINARY_PATH"; \
		sudo rm -f "$$BINARY_PATH"; \
		echo "✓ ${BINARY_NAME} uninstalled successfully!"; \
	else \
		echo "⚠ ${BINARY_NAME} not found in PATH"; \
	fi


cross-compile: clean ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p ${BUILD_DIR}
	
	# Linux
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-linux-amd64 ./cmd/bootup
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-linux-arm64 ./cmd/bootup

	@echo "✓ Cross-compilation complete! Binaries are in ${BUILD_DIR}/"

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

lint: ## Lint the codebase
	@echo "Linting code..."
	golangci-lint run

deps: ## Update Go module dependencies
	@echo "Updating dependencies..."
	go mod download
	go mod tidy