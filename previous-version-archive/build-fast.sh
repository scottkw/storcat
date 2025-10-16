#!/bin/bash

# StorCat Fast Build Script - Sequential but optimized for speed
# Alternative to parallel builds if system resources are limited

set -e  # Exit on any error

echo "🚀 Starting StorCat fast build process..."
echo "======================================="

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf dist/
mkdir -p dist/

# Build renderer (Vite) - only once
echo "📦 Building renderer with Vite..."
npm run build:renderer

echo ""
echo "🏗️  Building for all platforms (sequential, optimized)..."
echo "======================================================"

# Build each platform with all their architectures at once
echo "🪟 Building Windows (ARM64 + x64)..."
npx electron-builder --win --arm64 --x64 --publish=never

echo "🍎 Building macOS (ARM64 + x64 + Universal)..."
npx electron-builder --mac --arm64 --x64 --universal --publish=never

echo "🐧 Building Linux (AppImage, DEB, RPM for x64 + ARM64)..."
npx electron-builder --linux --x64 --arm64 --publish=never

echo ""
echo "✅ Fast build process completed!"
echo "==============================="
echo "📁 All builds are available in the 'dist/' directory"
echo ""
echo "📋 Built packages:"
echo "  Windows:"
echo "    - StorCat (x64).exe"
echo "    - StorCat (arm64).exe"
echo "  macOS:"
echo "    - StorCat (Apple Silicon).dmg"
echo "    - StorCat (Intel).dmg" 
echo "    - StorCat (Universal).dmg"
echo "  Linux:"
echo "    - StorCat (x64).AppImage"
echo "    - StorCat (arm64).AppImage"
echo "    - storcat_*_amd64.deb"
echo "    - storcat_*_arm64.deb"
echo "    - storcat-*.x86_64.rpm"
echo "    - storcat-*.aarch64.rpm"
echo ""
echo "🎉 Ready for distribution!"