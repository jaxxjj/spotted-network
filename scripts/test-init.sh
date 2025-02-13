#!/bin/bash

# Set the current directory to the script's directory
cd "$(dirname "$0")/.."

# Build Linux binary first
echo "Building Linux binary..."
GOOS=linux GOARCH=amd64 go build -o build/spotted-linux-amd64 ./cmd/operator

echo "Building test image..."
docker build -t spotted-init-test -f test/Dockerfile.init .

echo "Starting test container..."
echo "You can test the init command by running: spotted init"
echo "Press Ctrl+D or type 'exit' to exit the container"
echo "----------------------------------------"

docker run -it --rm \
    -v "$(pwd):/app" \
    --name spotted-init-test \
    spotted-init-test