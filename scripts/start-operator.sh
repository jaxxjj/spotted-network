#!/bin/bash

# default values for environment variables
# Replace your keystore path and password here
KEYSTORE_PATH=${KEYSTORE_PATH:-"/app/dummy.key.json"} 
KEYSTORE_PASSWORD=${KEYSTORE_PASSWORD:-"testpassword"}

echo "Starting operator with keystore: $KEYSTORE_PATH"
echo "Using config file: $CONFIG_PATH"

# Wait for registry gRPC endpoint to be ready
while ! nc -z registry 8000; do
  echo "Waiting for registry gRPC endpoint..."
  sleep 1
done
echo "Connection to registry ($(getent hosts registry | awk '{ print $1 }')) 8000 port [tcp/*] succeeded!"
echo "Registry gRPC endpoint is ready"

# Get registry host ID
echo "Getting registry host ID..."
REGISTRY_ID=$(./operator -keystore "$KEYSTORE_PATH" -password "$KEYSTORE_PASSWORD" -registry "registry:8000" -get-registry-id)
echo "Got registry ID: $REGISTRY_ID"

# Create join message
JOIN_MSG="join_request_$(date +%s)"
echo "Created join message: $JOIN_MSG"

# Get operator address
echo "Getting operator address..."
OPERATOR_ADDR=$(./operator -keystore "$KEYSTORE_PATH" -password "$KEYSTORE_PASSWORD")
echo "Operator address: $OPERATOR_ADDR"

# Sign message
echo "Signing message..."
MSG_SIG=$(./operator -keystore "$KEYSTORE_PATH" -password "$KEYSTORE_PASSWORD" -message "$JOIN_MSG")
echo "Message signature: $MSG_SIG"

# Submit join request
echo "Submitting join request..."
./operator -keystore "$KEYSTORE_PATH" -password "$KEYSTORE_PASSWORD" -registry "registry:8000" -join "$JOIN_MSG" "$MSG_SIG"
echo "Join request successful"

# Get registry IP for P2P connection
REGISTRY_IP=$(getent hosts registry | awk '{ print $1 }')
echo "Registry IP: $REGISTRY_IP"

# Construct full multiaddr for registry
REGISTRY_MULTIADDR="/ip4/$REGISTRY_IP/tcp/9000/p2p/$REGISTRY_ID"
echo "Connecting to registry at: $REGISTRY_MULTIADDR"

# Start operator node
echo "Starting operator node..."
./operator -keystore "$KEYSTORE_PATH" -password "$KEYSTORE_PASSWORD" -registry "$REGISTRY_MULTIADDR" 