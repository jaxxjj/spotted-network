#!/bin/sh
set -e

# Default installation directory
INSTALL_DIR="$HOME/bin"
BINARY_NAME="spotted"
GITHUB_REPO="galxe/spotted-network"

# Process flags
while getopts "b:" opt; do
  case $opt in
    b) INSTALL_DIR="$OPTARG"
    ;;
    \?) echo "Invalid option -$OPTARG" >&2
    exit 1
    ;;
  esac
done

# Ensure installation directory exists
mkdir -p "$INSTALL_DIR"

# Detect operating system and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  armv7l) ARCH="arm" ;;
esac

# Get latest release version
echo "Fetching latest release..."
VERSION=$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$VERSION" ]; then
  echo "Error: Unable to get latest version. Please check your internet connection."
  exit 1
fi

# Construct binary name and URL
BINARY_FILENAME="${BINARY_NAME}-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/${VERSION}/${BINARY_FILENAME}"

# Download binary
echo "Downloading spotted network CLI ${VERSION}..."
curl -L "$DOWNLOAD_URL" -o "$INSTALL_DIR/$BINARY_NAME"

# Make binary executable
chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo "Successfully installed spotted network CLI to $INSTALL_DIR/$BINARY_NAME"

# Check if directory is in PATH
case ":$PATH:" in
  *":$INSTALL_DIR:"*) : ;; # Directory is in PATH
  *)
    echo ""
    echo "To add the binary to your path, run:"
    echo "  export PATH=\$PATH:$INSTALL_DIR"
    echo ""
    echo "To add it permanently, add the above line to your ~/.bashrc or ~/.zshrc"
    ;;
esac

echo ""
echo "To get started, run:"
echo "  $BINARY_NAME --help" 