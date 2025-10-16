#!/bin/bash

# StorCat Build Script - All Platforms and Architectures (Optimized)
# This script builds StorCat for all supported platforms and architectures in parallel

set -e  # Exit on any error

echo "🚀 Starting StorCat optimized build process for all platforms..."
echo "=============================================================="

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf dist/
mkdir -p dist/

# Build renderer (Vite) - only once
echo "📦 Building renderer with Vite..."
npm run build:renderer

echo ""
echo "🏗️  Building for all platforms and architectures in parallel..."
echo "=============================================================="

# Function to run builds in background and track PIDs
build_pids=()

# Windows builds (both architectures together)
echo "🪟 Starting Windows builds (ARM64 + x64)..."
npx electron-builder --win --arm64 --x64 --publish=never &
build_pids+=($!)

# macOS builds (all architectures together) 
echo "🍎 Starting macOS builds (ARM64 + x64 + Universal)..."
npx electron-builder --mac --arm64 --x64 --universal --publish=never &
build_pids+=($!)

# Linux builds (all formats and architectures together)
echo "🐧 Starting Linux builds (AppImage, DEB, RPM for x64 + ARM64)..."
npx electron-builder --linux --x64 --arm64 --publish=never &
build_pids+=($!)

echo ""
echo "⏳ All builds started in parallel. Waiting for completion..."
echo "=============================================================="

# Wait for all background processes and track completion
completed=0
total=${#build_pids[@]}

for pid in "${build_pids[@]}"; do
    echo "⏳ Waiting for build process $pid..."
    if wait $pid; then
        ((completed++))
        echo "✅ Build process $pid completed successfully ($completed/$total)"
    else
        echo "❌ Build process $pid failed"
        # Kill remaining processes
        for remaining_pid in "${build_pids[@]}"; do
            if kill -0 $remaining_pid 2>/dev/null; then
                kill $remaining_pid 2>/dev/null || true
            fi
        done
        exit 1
    fi
done

echo ""
echo "✅ Build process completed!"
echo "=========================="
echo "📁 All builds are available in the 'dist/' directory"
echo ""
echo "📋 Built packages:"
echo "  Windows:"
echo "    - StorCat Setup (x64).exe"
echo "    - StorCat Setup (arm64).exe"
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