#!/bin/bash

# StorCat Complete Build Script - All compatible platforms
# This script builds all packages that can be reliably built on macOS

set -e  # Exit on any error

echo "🚀 Starting StorCat complete build process..."
echo "============================================"

# Get version from package.json
VERSION=$(node -p "require('./package.json').version")
echo "📋 Building StorCat v$VERSION"
echo ""

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

# Windows builds - Portable (MSI skipped on macOS due to Wine compatibility issues)
echo "🪟 Building Windows x64 Portable..."
npx electron-builder --win portable --x64 --publish=never

echo "🪟 Building Windows ARM64 Portable..."
npx electron-builder --win portable --arm64 --publish=never

# MSI builds - only on Windows (Wine issues on macOS)
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
    echo "🪟 Building Windows x64 MSI..."
    npx electron-builder --win msi --x64 --publish=never

    echo "🪟 Building Windows ARM64 MSI..."
    npx electron-builder --win msi --arm64 --publish=never
    
    MSI_BUILT=true
else
    echo ""
    echo "ℹ️  Skipping MSI builds (Wine compatibility issues on non-Windows)"
    echo "   For MSI installers, build on Windows with:"
    echo "   npx electron-builder --win msi --x64 --publish=never"
    echo "   npx electron-builder --win msi --arm64 --publish=never"
    
    MSI_BUILT=false
fi

# Linux builds - AppImage and DEB (RPM has issues on macOS)
echo "🐧 Building Linux AppImage x64..."
npx electron-builder --linux appimage --x64 --publish=never

echo "🐧 Building Linux AppImage ARM64..."
npx electron-builder --linux appimage --arm64 --publish=never

echo "🐧 Building Linux DEB x64..."
npx electron-builder --linux deb --x64 --publish=never

echo "🐧 Building Linux DEB ARM64..."
npx electron-builder --linux deb --arm64 --publish=never

# Linux-specific builds (RPM and Snap)
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "🐧 Building Linux RPM x64..."
    npx electron-builder --linux rpm --x64 --publish=never

    echo "🐧 Building Linux RPM ARM64..."
    npx electron-builder --linux rpm --arm64 --publish=never
    
    echo "🐧 Building Linux Snap x64..."
    npx electron-builder --linux snap --x64 --publish=never

    echo "🐧 Building Linux Snap ARM64..."
    npx electron-builder --linux snap --arm64 --publish=never
    
    LINUX_EXTRA_BUILT=true
else
    echo ""
    echo "ℹ️  Skipping RPM and Snap builds (not on Linux)"
    echo "   For RPM and Snap packages, build on Linux with:"
    echo "   npx electron-builder --linux rpm --x64 --publish=never"
    echo "   npx electron-builder --linux rpm --arm64 --publish=never"
    echo "   npx electron-builder --linux snap --x64 --publish=never" 
    echo "   npx electron-builder --linux snap --arm64 --publish=never"
    
    LINUX_EXTRA_BUILT=false
fi

# macOS builds - build universal binary last (with notarization)
echo ""
echo "🍎 Building macOS Universal (Intel + Apple Silicon) with notarization..."
echo "⚠️  This may take several minutes due to Apple notarization process..."
npx electron-builder --mac --universal --publish=never

echo ""
echo "✅ Complete build process finished!"
echo "=================================="
echo "📁 Built packages available in 'dist/' directory"
echo ""
echo "📋 Successfully built:"
echo "  - Windows x64 portable"
echo "  - Windows ARM64 portable"

if [ "$MSI_BUILT" = true ]; then
    echo "  - Windows x64 MSI installer"
    echo "  - Windows ARM64 MSI installer"
fi

echo "  - macOS Universal installer (notarized)"
echo "  - Linux x64 AppImage"
echo "  - Linux ARM64 AppImage"
echo "  - Linux x64 DEB package"
echo "  - Linux ARM64 DEB package"

# Count total packages built
TOTAL_BUILT=7  # Base: 2 Windows portable + 1 macOS universal + 4 Linux

if [ "$MSI_BUILT" = true ]; then
    echo "  - Windows x64 MSI installer"
    echo "  - Windows ARM64 MSI installer"
    TOTAL_BUILT=$((TOTAL_BUILT + 2))
fi

if [ "$LINUX_EXTRA_BUILT" = true ]; then
    echo "  - Linux x64 RPM package"
    echo "  - Linux ARM64 RPM package"
    echo "  - Linux x64 Snap package"
    echo "  - Linux ARM64 Snap package"
    TOTAL_BUILT=$((TOTAL_BUILT + 4))
fi

echo ""
if [ "$LINUX_EXTRA_BUILT" = true ] && [ "$MSI_BUILT" = true ]; then
    echo "🎉 All 13 distributions ready for StorCat v$VERSION!"
elif [ "$LINUX_EXTRA_BUILT" = true ] || [ "$MSI_BUILT" = true ]; then
    echo "🎉 $TOTAL_BUILT out of 13 distributions ready for StorCat v$VERSION!"
else
    echo "🎉 7 out of 13 distributions ready for StorCat v$VERSION!"
    echo "💡 Build MSI packages on Windows and RPM/Snap packages on Linux for complete coverage"
fi

echo ""
echo "📦 Version: $VERSION"
echo "📁 Location: dist/"
echo "🔗 Build completed successfully!"