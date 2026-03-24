#!/bin/bash

# StorCat Linux Build Script
# Builds for Linux AMD64 and ARM64

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}Building StorCat for Linux...${NC}"

cd "$(dirname "$0")/.."

# Build AMD64
echo -e "${YELLOW}Building AMD64...${NC}"
wails build -clean -platform linux/amd64

if [ -f "build/bin/StorCat" ]; then
    echo -e "${GREEN}✓ AMD64 build successful!${NC}"
    mkdir -p dist/linux-amd64
    cp build/bin/StorCat dist/linux-amd64/
    echo "  → dist/linux-amd64/StorCat"
fi

echo ""

# Build ARM64
echo -e "${YELLOW}Building ARM64...${NC}"
wails build -clean -platform linux/arm64

if [ -f "build/bin/StorCat" ]; then
    echo -e "${GREEN}✓ ARM64 build successful!${NC}"
    mkdir -p dist/linux-arm64
    cp build/bin/StorCat dist/linux-arm64/
    echo "  → dist/linux-arm64/StorCat"
fi

echo ""
echo -e "${GREEN}Linux builds complete!${NC}"
