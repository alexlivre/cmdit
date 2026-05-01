#!/usr/bin/env bash
# cmdit installer — one-liner for Linux and macOS
# Usage: curl -sSL https://raw.githubusercontent.com/alexlivre/cmdit/main/install.sh | bash

set -e

REPO="alexlivre/cmdit"
VERSION="${VERSION:-latest}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)       echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

case "$OS" in
    linux)   BINARY="cmdit-linux-$ARCH" ;;
    darwin)  BINARY="cmdit-darwin-$ARCH" ;;
    *)       echo "Unsupported OS: $OS" && exit 1 ;;
esac

# Download URL
if [ "$VERSION" = "latest" ]; then
    URL="https://github.com/$REPO/releases/latest/download/$BINARY"
else
    URL="https://github.com/$REPO/releases/download/$VERSION/$BINARY"
fi

echo "→ Installing cmdit for $OS/$ARCH..."
echo "  Downloading: $URL"

# Download to temp file
TMP=$(mktemp)
curl -sSL "$URL" -o "$TMP"
chmod +x "$TMP"

# Install to /usr/local/bin (requires sudo)
INSTALL_DIR="/usr/local/bin"
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP" "$INSTALL_DIR/cmdit"
else
    sudo mv "$TMP" "$INSTALL_DIR/cmdit"
fi

echo "✅ cmdit installed to $INSTALL_DIR/cmdit"
echo "   Run: cmdit [file]"
