#!/bin/bash

# Homebrew Tap Update Script for StorCat
# Updates the tap when a new StorCat release is published

set -e  # Exit on any error

REPO_OWNER="scottkw"
REPO_NAME="storcat"
CASK_FILE="Casks/storcat.rb"

echo "🔄 Updating StorCat Homebrew Tap..."
echo "======================================"

# Function to get latest release info from GitHub API
get_latest_release() {
    curl -s "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
}

# Function to download file and get SHA256
get_sha256() {
    local url="$1"
    local temp_file=$(mktemp)
    
    echo "📥 Downloading: $url" >&2
    curl -L "$url" -o "$temp_file" 2>/dev/null
    local sha256=$(shasum -a 256 "$temp_file" | cut -d' ' -f1)
    rm -f "$temp_file"
    echo "$sha256"
}

# Function to check if version already exists in cask
version_exists() {
    local version="$1"
    grep -q "version \"$version\"" "$CASK_FILE" 2>/dev/null
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

# Check if this version is already in the cask
if version_exists "$VERSION"; then
    echo "ℹ️  Version $VERSION is already up to date in the tap"
    exit 0
fi

# Build download URL for universal binary
UNIVERSAL_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/StorCat-${VERSION}-universal.dmg"

echo ""
echo "🔐 Calculating SHA256 hash..."

# Get SHA256 hash for universal binary
UNIVERSAL_SHA256=$(get_sha256 "$UNIVERSAL_URL")

if [ -z "$UNIVERSAL_SHA256" ]; then
    echo "❌ Error: Failed to calculate SHA256 hash"
    exit 1
fi

echo "✅ Universal SHA256: $UNIVERSAL_SHA256"

# Backup current cask file
cp "$CASK_FILE" "${CASK_FILE}.backup"

# Update the cask file
echo ""
echo "✏️  Updating cask formula..."

cat > "$CASK_FILE" << EOF
cask "storcat" do
  version "$VERSION"
  sha256 "$UNIVERSAL_SHA256"

  url "https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/#{version}/StorCat-#{version}-universal.dmg"
  
  name "StorCat"
  desc "Directory Catalog Manager - Create, search, and browse directory catalogs"
  homepage "https://github.com/${REPO_OWNER}/${REPO_NAME}"
  
  livecheck do
    url :url
    strategy :github_latest
  end
  
  app "StorCat.app"
  
  zap trash: [
    "~/Library/Application Support/StorCat",
    "~/Library/Preferences/com.kenscott.storcat.plist",
    "~/Library/Saved Application State/com.kenscott.storcat.savedState",
    "~/Library/Caches/com.kenscott.storcat",
  ]
end
EOF

echo "✅ Cask formula updated successfully!"

# Validate the cask (optional - requires brew)
if command -v brew >/dev/null 2>&1; then
    echo ""
    echo "🔍 Validating cask syntax..."
    if brew cask audit --strict "$CASK_FILE" 2>/dev/null; then
        echo "✅ Cask validation passed"
    else
        echo "⚠️  Cask validation failed - please check manually"
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
git add "$CASK_FILE"
git commit -m "Update StorCat to version $VERSION

- Updated to StorCat v$VERSION
- Now uses universal binary (Intel x64 + Apple Silicon ARM64)
- SHA256: $UNIVERSAL_SHA256"

echo "✅ Changes committed successfully!"

# Ask if user wants to push
echo ""
read -p "🚀 Push changes to GitHub? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    git push origin main
    echo "✅ Changes pushed to GitHub!"
    echo ""
    echo "🎉 Homebrew tap updated successfully!"
    echo "Users can now install StorCat v$VERSION with:"
    echo "   brew upgrade storcat"
else
    echo "📝 Changes committed locally. Push manually with:"
    echo "   git push origin main"
fi

# Clean up backup
rm -f "${CASK_FILE}.backup"

echo ""
echo "✨ Update complete!"