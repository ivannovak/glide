#!/usr/bin/env bash

set -e

REPO="ivannovak/glide"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="glid"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$OS" in
        darwin)
            PLATFORM="darwin"
            ;;
        linux)
            PLATFORM="linux"
            ;;
        *)
            echo -e "${RED}Unsupported OS: $OS${NC}"
            exit 1
            ;;
    esac
    
    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    BINARY_SUFFIX="${PLATFORM}-${ARCH}"
}

# Get latest release version from GitHub
get_latest_version() {
    echo -e "${YELLOW}Fetching latest version...${NC}"
    VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$VERSION" ]; then
        echo -e "${RED}Failed to fetch latest version${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}Latest version: $VERSION${NC}"
}

# Download binary
download_binary() {
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}-${BINARY_SUFFIX}"
    TEMP_FILE="/tmp/${BINARY_NAME}-${BINARY_SUFFIX}"
    
    echo -e "${YELLOW}Downloading ${BINARY_NAME} for ${PLATFORM}/${ARCH}...${NC}"
    
    if ! curl -L -o "$TEMP_FILE" "$DOWNLOAD_URL" --progress-bar; then
        echo -e "${RED}Failed to download binary${NC}"
        exit 1
    fi
    
    chmod +x "$TEMP_FILE"
}

# Install binary
install_binary() {
    echo -e "${YELLOW}Installing to ${INSTALL_DIR}/${BINARY_NAME}...${NC}"
    
    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TEMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        echo -e "${YELLOW}Sudo required to install to ${INSTALL_DIR}${NC}"
        sudo mv "$TEMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    echo -e "${GREEN}Installation complete!${NC}"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        INSTALLED_VERSION=$("$BINARY_NAME" --version 2>&1 | head -n1)
        echo -e "${GREEN}✓ ${BINARY_NAME} installed successfully${NC}"
        echo -e "${GREEN}  Version: $INSTALLED_VERSION${NC}"
        echo -e "${GREEN}  Location: $(which $BINARY_NAME)${NC}"
    else
        echo -e "${RED}Installation verification failed${NC}"
        echo -e "${YELLOW}You may need to add ${INSTALL_DIR} to your PATH${NC}"
        exit 1
    fi
}

# Main installation flow
main() {
    echo -e "${GREEN}Installing Glide CLI${NC}"
    echo "===================="
    
    detect_platform
    get_latest_version
    download_binary
    install_binary
    verify_installation
    
    echo ""
    echo -e "${GREEN}Get started with: glid help${NC}"
}

# Run main function
main "$@"