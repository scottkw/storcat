#!/bin/sh
# install-cli.sh -- Create /usr/local/bin/storcat symlink for CLI access
# Run after installing StorCat.app to /Applications.

set -e

APP_BINARY="/Applications/StorCat.app/Contents/MacOS/StorCat"
LINK_PATH="/usr/local/bin/storcat"

if [ ! -f "$APP_BINARY" ]; then
    echo "Error: StorCat.app not found in /Applications" >&2
    echo "Install StorCat.app to /Applications first." >&2
    exit 1
fi

if [ ! -d "/usr/local/bin" ]; then
    echo "Creating /usr/local/bin (may require sudo)..." >&2
    sudo mkdir -p /usr/local/bin
fi

ln -sf "$APP_BINARY" "$LINK_PATH"
echo "Installed: $LINK_PATH -> $APP_BINARY"
echo "Run 'storcat --help' to verify."
