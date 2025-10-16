#!/bin/bash

# StorCat Build Script for All Platforms
# Builds macOS (Intel + ARM64), Windows, and Linux versions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Version info
VERSION="2.0.0"

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}    StorCat v${VERSION} - Build All Platforms    ${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Change to project root
cd "$(dirname "$0")/.."
PROJECT_ROOT=$(pwd)

echo -e "${GREEN}Project root: ${PROJECT_ROOT}${NC}"
echo ""

# Ensure wails is installed
if ! command -v wails &> /dev/null; then
    echo -e "${RED}Error: wails command not found!${NC}"
    echo "Please install Wails: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    exit 1
fi

# Create dist directory
mkdir -p dist
rm -rf dist/*

echo -e "${YELLOW}Building for all platforms...${NC}"
echo ""

# Build macOS (Universal Binary - Intel + ARM64)
echo -e "${GREEN}[1/4] Building macOS Universal Binary (Intel + ARM64)...${NC}"
wails build -clean -platform darwin/universal

if [ -d "build/bin/StorCat.app" ]; then
    echo -e "${GREEN}✓ macOS build successful${NC}"
    cp -R build/bin/StorCat.app dist/
    echo -e "${GREEN}  → dist/StorCat.app${NC}"
else
    echo -e "${RED}✗ macOS build failed${NC}"
fi
echo ""

# Build Windows AMD64 (no -clean to preserve macOS build)
echo -e "${GREEN}[2/4] Building Windows (AMD64)...${NC}"
wails build -platform windows/amd64

if [ -f "build/bin/StorCat.exe" ]; then
    echo -e "${GREEN}✓ Windows build successful${NC}"
    mkdir -p dist/windows-amd64
    cp build/bin/StorCat.exe dist/windows-amd64/
    echo -e "${GREEN}  → dist/windows-amd64/StorCat.exe${NC}"
else
    echo -e "${RED}✗ Windows build failed${NC}"
fi
echo ""

# Build Linux AMD64 (no -clean to preserve previous builds)
echo -e "${GREEN}[3/4] Building Linux (AMD64)...${NC}"
echo -e "${YELLOW}Note: Cannot cross-compile to Linux from macOS. Use Docker script instead.${NC}"
echo -e "${YELLOW}Run: ./scripts/build-linux-docker.sh${NC}"
echo ""

# Build Linux ARM64
echo -e "${GREEN}[4/4] Building Linux (ARM64)...${NC}"
echo -e "${YELLOW}Note: Cannot cross-compile to Linux from macOS. Use Docker script instead.${NC}"
echo -e "${YELLOW}Run: ./scripts/build-linux-docker.sh${NC}"
echo ""

# List all builds
echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}             Build Summary                      ${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""
echo "Distribution files created in dist/:"
ls -lh dist/
if [ -d "dist/StorCat.app" ]; then
    echo ""
    echo "macOS app size:"
    du -sh dist/StorCat.app
fi
if [ -d "dist/windows-amd64" ]; then
    echo ""
    echo "Windows build:"
    ls -lh dist/windows-amd64/
fi
if [ -d "dist/linux-amd64" ]; then
    echo ""
    echo "Linux AMD64 build:"
    ls -lh dist/linux-amd64/
fi
if [ -d "dist/linux-arm64" ]; then
    echo ""
    echo "Linux ARM64 build:"
    ls -lh dist/linux-arm64/
fi
echo ""
echo -e "${GREEN}✓ All builds complete!${NC}"
