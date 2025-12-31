#!/bin/bash
set -e

# Bootup CLI Installation Script
# This script downloads and installs the latest bootup CLI binary

VERSION="${BOOTUP_VERSION:-latest}"
REPO="amirkh8006/bootup-cli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print functions
info() {
    echo -e "${GREEN}ℹ${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1" >&2
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $OS in
        linux) OS="linux" ;;
        darwin) OS="darwin" ;;
        *) error "Unsupported operating system: $OS"; exit 1 ;;
    esac

    case $ARCH in
        x86_64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        armv7l) ARCH="arm" ;;
        *) error "Unsupported architecture: $ARCH"; exit 1 ;;
    esac

    BINARY_NAME="bootup-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="${BINARY_NAME}.exe"
    fi
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Download and install bootup
install_bootup() {
    info "Installing Bootup CLI..."
    info "Platform: ${OS}-${ARCH}"
    
    # Check for required commands
    if ! command_exists curl && ! command_exists wget; then
        error "curl or wget is required but not installed."
        exit 1
    fi

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # Construct download URL
    if [ "$VERSION" = "latest" ]; then
        DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}"
    else
        DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"
    fi

    info "Downloading from: $DOWNLOAD_URL"

    # Download binary
    BINARY_PATH="$TMP_DIR/bootup"
    if command_exists curl; then
        curl -fsSL "$DOWNLOAD_URL" -o "$BINARY_PATH"
    else
        wget -q "$DOWNLOAD_URL" -O "$BINARY_PATH"
    fi

    if [ ! -f "$BINARY_PATH" ]; then
        error "Failed to download bootup binary"
        exit 1
    fi

    # Make executable
    chmod +x "$BINARY_PATH"

    # Install to system
    if [ -w "$INSTALL_DIR" ]; then
        mv "$BINARY_PATH" "$INSTALL_DIR/bootup"
    else
        info "Installing to $INSTALL_DIR (requires sudo)..."
        sudo mv "$BINARY_PATH" "$INSTALL_DIR/bootup"
    fi

    success "Bootup CLI installed successfully!"
}

# Verify installation
verify_installation() {
    if command_exists bootup; then
        VERSION_OUTPUT=$(bootup --version 2>/dev/null || bootup version 2>/dev/null || echo "unknown")
        success "Installation verified: $VERSION_OUTPUT"
        info "Run 'bootup' to get started."
    else
        warn "bootup command not found in PATH. You may need to:"
        warn "  - Add $INSTALL_DIR to your PATH"
        warn "  - Restart your terminal"
        warn "  - Run 'source ~/.bashrc' or 'source ~/.zshrc'"
    fi
}

# Main installation process
main() {
    echo "🚀 Bootup CLI Installation Script"
    echo "=================================="
    
    detect_platform
    install_bootup
    verify_installation
    
    echo ""
    echo "🎉 Installation complete!"
    echo "Visit https://github.com/${REPO} for documentation and support."
}

# Run main function
main "$@"
