#!/bin/bash

# default values for environment variables

echo "Starting operator with operator key: $OPERATOR_KEY_PATH"
echo "Using signing key: $SIGNING_KEY_PATH"
echo "Using config file: $CONFIG_PATH"

# Check if REGISTRY_PEER_ID is set
if [ -z "$REGISTRY_PEER_ID" ]; then
    echo "REGISTRY_PEER_ID environment variable is not set"
    exit 1
fi
echo "Using registry peer ID: $REGISTRY_PEER_ID"

echo "Waiting for registry to be ready..."

# Wait for registry P2P endpoint to be ready
while ! nc -z registry 9000; do
    echo "Waiting for registry P2P endpoint..."
    sleep 1
done
echo "Registry P2P endpoint is ready"

# Get registry IP for P2P connection
REGISTRY_IP=$(getent hosts registry | awk '{ print $1 }')
echo "Registry IP: $REGISTRY_IP"

# Start the operator node
exec ./operator \
  -operator-key "$OPERATOR_KEY_PATH" \
  -signing-key "$SIGNING_KEY_PATH" \
  -password "$KEYSTORE_PASSWORD" \
  -registry "/ip4/$REGISTRY_IP/tcp/9000/p2p/$REGISTRY_PEER_ID"