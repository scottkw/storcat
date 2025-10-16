#!/bin/bash

# StorCat Simple Build Script - Individual platform builds
# This script builds each platform/architecture combination individually for maximum compatibility

set -e  # Exit on any error

echo "🚀 Starting StorCat simple build process..."
echo "==========================================="

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf dist/
mkdir -p dist/

# Build renderer (Vite) - only once
echo "📦 Building renderer with Vite..."
npm run build:renderer

echo ""
echo "🏗️  Building each platform individually..."
echo "========================================"

# Windows builds
echo "🪟 Building Windows x64..."
npx electron-builder --win --x64 --publish=never

echo "🪟 Building Windows ARM64..."
npx electron-builder --win --arm64 --publish=never

# macOS builds  
echo "🍎 Building macOS x64 (Intel)..."
npx electron-builder --mac --x64 --publish=never

echo "🍎 Building macOS ARM64 (Apple Silicon)..."
npx electron-builder --mac --arm64 --publish=never

# Linux builds - AppImage, DEB, and RPM
echo "🐧 Building Linux AppImage x64..."
npx electron-builder --linux appimage --x64 --publish=never

echo "🐧 Building Linux AppImage ARM64..."
npx electron-builder --linux appimage --arm64 --publish=never

echo "🐧 Building Linux DEB x64..."
npx electron-builder --linux deb --x64 --publish=never

echo "🐧 Building Linux DEB ARM64..."
npx electron-builder --linux deb --arm64 --publish=never

echo "🐧 Building Linux RPM x64..."
npx electron-builder --linux rpm --x64 --publish=never

echo "🐧 Building Linux RPM ARM64..."
npx electron-builder --linux rpm --arm64 --publish=never

echo ""
echo "✅ Simple build process completed!"
echo "================================="
echo "📁 Built packages available in 'dist/' directory"
echo ""
echo "📋 Successfully built:"
echo "  - Windows x64 portable"
echo "  - Windows ARM64 portable" 
echo "  - macOS x64 installer"
echo "  - macOS ARM64 installer"
echo "  - Linux x64 AppImage"
echo "  - Linux ARM64 AppImage"
echo "  - Linux x64 DEB package"
echo "  - Linux ARM64 DEB package"
echo "  - Linux x64 RPM package"
echo "  - Linux ARM64 RPM package"
echo ""
echo "🎉 All 10 distributions ready!"