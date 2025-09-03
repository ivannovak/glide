#!/usr/bin/env bash

set -e

# Version can be passed as argument or defaults to git tag/commit
VERSION=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}

# Branding variables (can be overridden via environment)
COMMAND_NAME=${COMMAND_NAME:-glid}
CONFIG_FILE=${CONFIG_FILE:-.glide.yml}
PROJECT_NAME=${PROJECT_NAME:-Glide}
DESCRIPTION=${DESCRIPTION:-context-aware development CLI}
REPOSITORY_URL=${REPOSITORY_URL:-https://github.com/ivannovak/glide}

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building ${PROJECT_NAME} CLI (${COMMAND_NAME}) v${VERSION}${NC}"
echo "================================"
if [ "$COMMAND_NAME" != "glid" ]; then
    echo -e "${BLUE}Custom branding:${NC}"
    echo "  Command: $COMMAND_NAME"
    echo "  Config:  $CONFIG_FILE"
    echo "  Project: $PROJECT_NAME"
    echo ""
fi

# Ensure dist directory exists
mkdir -p dist

# Build information
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags for static binary
LDFLAGS="-extldflags '-static' -s -w"
LDFLAGS="$LDFLAGS -X github.com/ivannovak/glide/pkg/version.Version=${VERSION}"
LDFLAGS="$LDFLAGS -X github.com/ivannovak/glide/pkg/version.BuildDate=${BUILD_DATE}"
LDFLAGS="$LDFLAGS -X github.com/ivannovak/glide/pkg/version.GitCommit=${GIT_COMMIT}"

# Add branding flags
LDFLAGS="$LDFLAGS -X github.com/ivannovak/glide/pkg/branding.CommandName=${COMMAND_NAME}"
LDFLAGS="$LDFLAGS -X 'github.com/ivannovak/glide/pkg/branding.ConfigFileName=${CONFIG_FILE}'"
LDFLAGS="$LDFLAGS -X 'github.com/ivannovak/glide/pkg/branding.ProjectName=${PROJECT_NAME}'"
LDFLAGS="$LDFLAGS -X 'github.com/ivannovak/glide/pkg/branding.Description=${DESCRIPTION}'"
LDFLAGS="$LDFLAGS -X github.com/ivannovak/glide/pkg/branding.CompletionDir=${COMMAND_NAME}"
LDFLAGS="$LDFLAGS -X github.com/ivannovak/glide/pkg/branding.RepositoryURL=${REPOSITORY_URL}"

# Build for each platform
build_platform() {
    local GOOS=$1
    local GOARCH=$2
    local OUTPUT=$3
    
    echo -e "${YELLOW}Building for ${GOOS}/${GOARCH}...${NC}"
    
    CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
        go build -a -ldflags "${LDFLAGS}" \
        -o "dist/${OUTPUT}" \
        ./cmd/glid
    
    echo -e "${GREEN}✓ Built dist/${OUTPUT}${NC}"
}

# Build for all platforms
build_platform "darwin" "arm64" "${COMMAND_NAME}-darwin-arm64"
build_platform "darwin" "amd64" "${COMMAND_NAME}-darwin-amd64"
build_platform "linux" "amd64" "${COMMAND_NAME}-linux-amd64"
build_platform "linux" "arm64" "${COMMAND_NAME}-linux-arm64"

# Generate checksums
echo -e "${YELLOW}Generating checksums...${NC}"
cd dist
shasum -a 256 ${COMMAND_NAME}-* > checksums.txt
cd ..

echo -e "${GREEN}✓ Checksums generated${NC}"

# Copy install script to dist
cp scripts/install.sh dist/

# Show build artifacts
echo ""
echo -e "${GREEN}Build complete! Artifacts in dist/:${NC}"
ls -lh dist/

# Show file sizes
echo ""
echo -e "${GREEN}Binary sizes:${NC}"
du -h dist/${COMMAND_NAME}-* | sort