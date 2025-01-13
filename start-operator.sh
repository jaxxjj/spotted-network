#!/bin/sh

# Load dummy key
KEYSTORE_PATH="/app/dummy.key.json"
KEYSTORE_PASSWORD="testpassword"

echo "Starting operator with keystore: $KEYSTORE_PATH"

# Wait for registry gRPC endpoint to be ready
while ! nc -z registry 8000; do
    echo "Waiting for registry gRPC endpoint..."
    sleep 2
done
echo "Registry gRPC endpoint is ready"

# Get registry host ID using gRPC client
echo "Getting registry host ID..."
REGISTRY_ID=$(./operator -get-registry-id)
if [ -z "$REGISTRY_ID" ]; then
    echo "Failed to get registry host ID"
    exit 1
fi
echo "Got registry ID: $REGISTRY_ID"

# Create join message
JOIN_MESSAGE="join_request_$(date +%s)"
echo "Created join message: $JOIN_MESSAGE"

# Get operator address
echo "Getting operator address..."
ADDRESS=$(./operator -keystore "$KEYSTORE_PATH" -password "$KEYSTORE_PASSWORD")
if [ -z "$ADDRESS" ]; then
    echo "Failed to get operator address"
    exit 1
fi
echo "Operator address: $ADDRESS"

# Sign message
echo "Signing message..."
SIGNATURE=$(./operator -keystore "$KEYSTORE_PATH" -password "$KEYSTORE_PASSWORD" -message "$JOIN_MESSAGE")
if [ -z "$SIGNATURE" ]; then
    echo "Failed to sign message"
    exit 1
fi
echo "Message signature: $SIGNATURE"

# Submit join request
echo "Submitting join request..."
JOIN_RESULT=$(./operator -join -keystore "$KEYSTORE_PATH" -password "$KEYSTORE_PASSWORD" -message "$JOIN_MESSAGE")
if [ "$?" -ne 0 ]; then
    echo "Join request failed: $JOIN_RESULT"
    exit 1
fi
echo "Join request successful"

# Construct the full multiaddr
REGISTRY_ADDR="/ip4/172.20.0.2/tcp/9000/p2p/$REGISTRY_ID"
echo "Connecting to registry at: $REGISTRY_ADDR"

# Start the operator with the correct address
echo "Starting operator node..."
exec ./operator -registry "$REGISTRY_ADDR" -keystore "$KEYSTORE_PATH" -password "$KEYSTORE_PASSWORD" 