#!/bin/bash

# winget Source Update Script for StorCat
# Updates the winget manifests when a new StorCat release is published

set -e  # Exit on any error

REPO_OWNER="scottkw"
REPO_NAME="storcat"
MANIFEST_DIR="manifests/s/scottkw/StorCat"

echo "🔄 Updating StorCat winget Source..."
echo "==================================="

# Function to get latest release info from GitHub API
get_latest_release() {
    curl -s "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
}

# Function to download file and get SHA256
get_sha256() {
    local url="$1"
    local temp_file=$(mktemp)
    
    echo "📥 Downloading: $url"
    curl -L "$url" -o "$temp_file" 2>/dev/null
    local sha256=$(shasum -a 256 "$temp_file" | cut -d' ' -f1)
    rm -f "$temp_file"
    echo "$sha256"
}

# Function to check if version already exists
version_exists() {
    local version="$1"
    [ -d "${MANIFEST_DIR}/${version}" ]
}

# Get latest release information
echo "🔍 Fetching latest release information..."
RELEASE_JSON=$(get_latest_release)

# Extract version (remove 'v' prefix if present)
VERSION=$(echo "$RELEASE_JSON" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/' | sed 's/^v//')
if [ -z "$VERSION" ]; then
    echo "❌ Error: Could not determine latest version"
    exit 1
fi

echo "📋 Latest release: $VERSION"

# Check if this version is already in the manifests
if version_exists "$VERSION"; then
    echo "ℹ️  Version $VERSION is already up to date in the source"
    exit 0
fi

# Build download URL based on your naming convention
WINDOWS_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/StorCat.${VERSION}.exe"

echo ""
echo "🔐 Calculating SHA256 hash..."

# Get SHA256 hash for Windows executable
WINDOWS_SHA256=$(get_sha256 "$WINDOWS_URL")

if [ -z "$WINDOWS_SHA256" ]; then
    echo "❌ Error: Failed to calculate SHA256 hash"
    exit 1
fi

echo "✅ Windows SHA256: $WINDOWS_SHA256"

# Create directory for new version
VERSION_DIR="${MANIFEST_DIR}/${VERSION}"
mkdir -p "$VERSION_DIR"

echo ""
echo "✏️  Creating manifest files..."

# Create version manifest
cat > "${VERSION_DIR}/scottkw.StorCat.yaml" << EOF
# Created using winget-create 1.6.0.0
# yaml-language-server: \$schema=https://aka.ms/winget-manifest.version.1.6.0.schema.json

PackageIdentifier: scottkw.StorCat
PackageVersion: $VERSION
DefaultLocale: en-US
ManifestType: version
ManifestVersion: 1.6.0
EOF

# Create installer manifest
cat > "${VERSION_DIR}/scottkw.StorCat.installer.yaml" << EOF
# Created using winget-create 1.6.0.0
# yaml-language-server: \$schema=https://aka.ms/winget-manifest.installer.1.6.0.schema.json

PackageIdentifier: scottkw.StorCat
PackageVersion: $VERSION
InstallerType: portable
Commands:
- storcat
Installers:
- Architecture: x64
  InstallerUrl: https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/StorCat.${VERSION}.exe
  InstallerSha256: $WINDOWS_SHA256
ManifestType: installer
ManifestVersion: 1.6.0
EOF

# Get release notes from GitHub API
RELEASE_NOTES=$(echo "$RELEASE_JSON" | grep '"body"' | sed 's/.*"body": *"\([^"]*\)".*/\1/' | sed 's/\\n/\n/g' | sed 's/\\"/"/g')

# Create locale manifest
cat > "${VERSION_DIR}/scottkw.StorCat.locale.en-US.yaml" << EOF
# Created using winget-create 1.6.0.0
# yaml-language-server: \$schema=https://aka.ms/winget-manifest.defaultLocale.1.6.0.schema.json

PackageIdentifier: scottkw.StorCat
PackageVersion: $VERSION
PackageLocale: en-US
Publisher: Ken Scott
PublisherUrl: https://github.com/${REPO_OWNER}
PublisherSupportUrl: https://github.com/${REPO_OWNER}/${REPO_NAME}/issues
PackageName: StorCat
PackageUrl: https://github.com/${REPO_OWNER}/${REPO_NAME}
License: MIT
LicenseUrl: https://github.com/${REPO_OWNER}/${REPO_NAME}/blob/main/LICENSE
ShortDescription: Directory Catalog Manager - Create, search, and browse directory catalogs
Description: |-
  StorCat is a modern Electron-based directory catalog management application that provides a clean, intuitive GUI interface for creating, searching, and browsing directory catalogs. Generate both JSON and HTML representations of directory structures for easy documentation and sharing.
  
  Features:
  - Create comprehensive directory catalogs with detailed file information
  - Fast full-text search across multiple catalog files
  - Interactive catalog browser with modern table interface
  - Dark/Light mode with complete theme support
  - Professional table components with filtering, sorting, and pagination
  - Cross-platform compatibility (Windows, macOS, Linux)
  - 100% compatible with existing bash script catalog files
Moniker: storcat
Tags:
- catalog
- directory
- file-manager
- electron
- search
- documentation
- json
- html
ReleaseNotes: |-
  StorCat v$VERSION
  
  $RELEASE_NOTES
ReleaseNotesUrl: https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/tag/${VERSION}
Documentations:
- DocumentLabel: User Guide
  DocumentUrl: https://github.com/${REPO_OWNER}/${REPO_NAME}#readme
- DocumentLabel: Build Instructions
  DocumentUrl: https://github.com/${REPO_OWNER}/${REPO_NAME}#building-the-application
ManifestType: defaultLocale
ManifestVersion: 1.6.0
EOF

echo "✅ Manifest files created successfully!"

# Validate manifests (if winget is available)
if command -v winget >/dev/null 2>&1; then
    echo ""
    echo "🔍 Validating manifests..."
    if winget validate "$VERSION_DIR" 2>/dev/null; then
        echo "✅ Manifest validation passed"
    else
        echo "⚠️  Manifest validation failed - please check manually"
    fi
fi

# Git operations
echo ""
echo "📝 Committing changes..."

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo "❌ Error: Not in a git repository"
    exit 1
fi

# Add and commit changes
git add "$VERSION_DIR"
git commit -m "Add StorCat version $VERSION

- Added winget manifests for StorCat v$VERSION
- Windows x64 portable executable
- SHA256: $WINDOWS_SHA256
- Release URL: https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/tag/${VERSION}"

echo "✅ Changes committed successfully!"

# Ask if user wants to push
echo ""
read -p "🚀 Push changes to GitHub? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    git push origin main
    echo "✅ Changes pushed to GitHub!"
    echo ""
    echo "🎉 winget source updated successfully!"
    echo "Users can now install StorCat v$VERSION with:"
    echo "   winget upgrade scottkw.StorCat"
else
    echo "📝 Changes committed locally. Push manually with:"
    echo "   git push origin main"
fi

echo ""
echo "✨ Update complete!"