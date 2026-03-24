#!/bin/bash

# StorCat Windows Build Script
# Builds for Windows AMD64

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Building StorCat for Windows (AMD64)...${NC}"

cd "$(dirname "$0")/.."

# Build Windows binary
wails build -clean -platform windows/amd64

if [ -f "build/bin/StorCat.exe" ]; then
    echo -e "${GREEN}✓ Build successful!${NC}"
    echo ""
    echo "Executable created at: build/bin/StorCat.exe"
    ls -lh build/bin/StorCat.exe
else
    echo "Build failed!"
    exit 1
fi
