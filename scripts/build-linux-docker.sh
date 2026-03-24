#!/bin/bash

# StorCat Linux Build Script using Docker
# Builds Linux AMD64 and ARM64 using Docker containers

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}Building StorCat for Linux using Docker...${NC}"
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed!${NC}"
    echo "Please install Docker Desktop for Mac: https://www.docker.com/products/docker-desktop"
    exit 1
fi

# Check if Docker is running
if ! docker info &> /dev/null; then
    echo -e "${RED}Error: Docker is not running!${NC}"
    echo "Please start Docker Desktop and try again."
    exit 1
fi

cd "$(dirname "$0")/.."
PROJECT_ROOT=$(pwd)

echo -e "${GREEN}Project root: ${PROJECT_ROOT}${NC}"
echo ""

# Create dist directories
mkdir -p dist/linux-amd64
mkdir -p dist/linux-arm64

# Build AMD64
echo -e "${YELLOW}Building Linux AMD64 in Docker...${NC}"
echo -e "${BLUE}This may take 10-15 minutes on first run (downloading and installing dependencies)${NC}"
docker run --rm \
    -v "${PROJECT_ROOT}:/app" \
    -w /app \
    --platform linux/amd64 \
    golang:1.23 \
    bash -c "
        set -e
        export PATH=\$PATH:/root/go/bin
        echo 'Installing system packages...'
        apt-get update -qq
        DEBIAN_FRONTEND=noninteractive apt-get install -y -qq libgtk-3-dev libwebkit2gtk-4.0-dev npm
        echo 'Installing Wails CLI...'
        go install github.com/wailsapp/wails/v2/cmd/wails@latest
        echo 'Cleaning frontend node_modules to avoid platform conflicts...'
        cd /app/frontend
        rm -rf node_modules package-lock.json
        echo 'Building StorCat for Linux AMD64...'
        cd /app
        wails build -platform linux/amd64
        cp build/bin/StorCat /app/dist/linux-amd64/
        echo 'AMD64 build complete!'
    "

if [ -f "dist/linux-amd64/StorCat" ]; then
    echo -e "${GREEN}✓ Linux AMD64 build successful!${NC}"
    ls -lh dist/linux-amd64/StorCat
else
    echo -e "${RED}✗ Linux AMD64 build failed${NC}"
    exit 1
fi

echo ""

# Build ARM64
echo -e "${YELLOW}Building Linux ARM64 in Docker...${NC}"
echo -e "${BLUE}This may take 10-15 minutes on first run (downloading and installing dependencies)${NC}"
docker run --rm \
    -v "${PROJECT_ROOT}:/app" \
    -w /app \
    --platform linux/arm64 \
    golang:1.23 \
    bash -c "
        set -e
        export PATH=\$PATH:/root/go/bin
        echo 'Installing system packages...'
        apt-get update -qq
        DEBIAN_FRONTEND=noninteractive apt-get install -y -qq libgtk-3-dev libwebkit2gtk-4.0-dev npm
        echo 'Installing Wails CLI...'
        go install github.com/wailsapp/wails/v2/cmd/wails@latest
        echo 'Cleaning frontend node_modules to avoid platform conflicts...'
        cd /app/frontend
        rm -rf node_modules package-lock.json
        echo 'Building StorCat for Linux ARM64...'
        cd /app
        wails build -platform linux/arm64
        cp build/bin/StorCat /app/dist/linux-arm64/
        echo 'ARM64 build complete!'
    "

if [ -f "dist/linux-arm64/StorCat" ]; then
    echo -e "${GREEN}✓ Linux ARM64 build successful!${NC}"
    ls -lh dist/linux-arm64/StorCat
else
    echo -e "${RED}✗ Linux ARM64 build failed${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}All Linux builds complete!${NC}"
echo ""
echo "Build sizes:"
if [ -f "dist/linux-amd64/StorCat" ]; then
    du -sh dist/linux-amd64/
fi
if [ -f "dist/linux-arm64/StorCat" ]; then
    du -sh dist/linux-arm64/
fi
