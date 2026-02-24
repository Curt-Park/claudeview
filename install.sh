#!/usr/bin/env bash
# claudeview installer
# Usage: curl -fsSL https://raw.githubusercontent.com/Curt-Park/claudeview/main/install.sh | bash

set -euo pipefail

REPO="Curt-Park/claudeview"
BINARY="claudeview"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  darwin) OS="darwin" ;;
  linux)  OS="linux"  ;;
  *)
    echo "Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64 | amd64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

PLATFORM="${OS}-${ARCH}"

# Get latest release tag
echo "Fetching latest release..."
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Failed to fetch latest release tag" >&2
  exit 1
fi

echo "Installing ${BINARY} ${LATEST} for ${PLATFORM}..."

# Download binary
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${BINARY}-${PLATFORM}"
TMP=$(mktemp)
trap "rm -f $TMP" EXIT

if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP"; then
  echo "Failed to download from $DOWNLOAD_URL" >&2
  exit 1
fi

# Install
mkdir -p "$INSTALL_DIR"
chmod +x "$TMP"
mv "$TMP" "$INSTALL_DIR/$BINARY"

echo "Installed ${BINARY} ${LATEST} to ${INSTALL_DIR}/${BINARY}"

# Check if INSTALL_DIR is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  echo ""
  echo "NOTE: Add $INSTALL_DIR to your PATH:"
  echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
  echo "  source ~/.bashrc"
fi

echo "Run: ${BINARY} --help"
