#!/bin/bash
set -e

# Version information
VERSION="1.0.0"
COMMIT=$(git rev-parse --short HEAD)
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

# Build directory
BUILD_DIR="build"
mkdir -p $BUILD_DIR

# Build function
build() {
    local GOOS=$1
    local GOARCH=$2
    local OUTPUT=$3
    
    echo "Building for $GOOS/$GOARCH..."
    
    # Build flags to include version information
    local BUILD_FLAGS="-X github.com/jaxxjj/spotted-network/pkg/version.Version=$VERSION"
    BUILD_FLAGS="$BUILD_FLAGS -X github.com/jaxxjj/spotted-network/pkg/version.GitCommit=$COMMIT"
    BUILD_FLAGS="$BUILD_FLAGS -X github.com/jaxxjj/spotted-network/pkg/version.BuildTime=$BUILD_TIME"

    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "$BUILD_FLAGS" -o "$BUILD_DIR/$OUTPUT" cmd/operator/main.go
    
    echo "✓ Built $OUTPUT"
}

# Clean build directory
rm -rf $BUILD_DIR/*

# Build for each platform
build darwin amd64 spotted-darwin-amd64      # MacOS Intel
build darwin arm64 spotted-darwin-arm64      # MacOS Apple Silicon
build linux amd64 spotted-linux-amd64        # Linux Intel/AMD
build linux arm64 spotted-linux-arm64        # Linux ARM64
build windows amd64 spotted-windows-amd64.exe # Windows

echo "All builds completed! Binaries are in the $BUILD_DIR directory"
ls -l $BUILD_DIR