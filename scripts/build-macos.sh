#!/bin/bash

# StorCat macOS Build Script
# Builds universal binary (Intel + ARM64)

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Building StorCat for macOS (Universal)...${NC}"

cd "$(dirname "$0")/.."

# Build universal binary
wails build -clean -platform darwin/universal

if [ -d "build/bin/StorCat.app" ]; then
    echo -e "${GREEN}✓ Build successful!${NC}"
    echo ""
    echo "Application created at: build/bin/StorCat.app"
    du -sh build/bin/StorCat.app
    echo ""
    echo "To run: open build/bin/StorCat.app"
else
    echo "Build failed!"
    exit 1
fi
