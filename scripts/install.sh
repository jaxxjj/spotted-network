#!/bin/sh
set -e

# Default installation directory
INSTALL_DIR="$HOME/bin"
BINARY_NAME="spotted"
GITHUB_REPO="jaxxjj/spotted-network"
CONFIG_DIR="$HOME/.spotted"

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

# Create config directory
mkdir -p "$CONFIG_DIR"

# Download necessary files
echo "Downloading configuration files..."
FILES_TO_DOWNLOAD=(
    "Dockerfile.operator"
    "docker-compose.yml"
    "otel-collector-config.yaml"
    "prometheus.yml"
    "pkg/repos/blacklist/schema.sql"
    "pkg/repos/consensus_responses/schema.sql"
    "pkg/repos/operators/schema.sql"
    "pkg/repos/tasks/schema.sql"
)

echo "Setting up configuration in $CONFIG_DIR..."
for file in "${FILES_TO_DOWNLOAD[@]}"; do
    echo "Downloading $file..."
    dir=$(dirname "$CONFIG_DIR/$file")
    mkdir -p "$dir"
    
    # Download file with error handling
    if ! curl -L "https://raw.githubusercontent.com/$GITHUB_REPO/${VERSION}/$file" -o "$CONFIG_DIR/$file"; then
        echo "Error: Failed to download $file"
        echo "Please check your internet connection and try again."
        exit 1
    fi
    
    # Make scripts executable
    if [[ "$file" == *".sh" ]]; then
        chmod +x "$CONFIG_DIR/$file"
    fi
done

echo "Successfully installed spotted network CLI to $INSTALL_DIR/$BINARY_NAME"
echo "Configuration files are in $CONFIG_DIR"

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
echo "Next steps:"
echo "1. Initialize the operator:"
echo "   $BINARY_NAME init"
echo ""
echo "2. Start the services:"
echo "   cd $CONFIG_DIR && docker-compose up -d"
echo ""
echo "3. Check service status:"
echo "   docker-compose ps" 