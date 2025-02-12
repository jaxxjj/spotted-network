#!/bin/bash

echo "Starting operator with signing key: $SIGNING_KEY_PATH"
echo "Using P2P key from environment variable"

# Check if signing key file exists
if [ ! -f "$SIGNING_KEY_PATH" ]; then
    echo "Signing key file not found at: $SIGNING_KEY_PATH"
    exit 1
fi

# Check if P2P key is provided
if [ -z "$P2P_KEY_64" ]; then
    echo "P2P_KEY_64 environment variable is required"
    exit 1
fi

# Check if password is needed for ECDSA signing key
if [[ "$SIGNING_KEY_PATH" == *".key.json" ]] && [ -z "$KEYSTORE_PASSWORD" ]; then
    echo "KEYSTORE_PASSWORD environment variable is required for ECDSA signing key"
    exit 1
fi

# Start the operator node with all required flags
exec ./operator \
    -signing-key "$SIGNING_KEY_PATH" \
    -p2p-key-64 "$P2P_KEY_64" \
    -password "$KEYSTORE_PASSWORD"