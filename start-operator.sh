#!/bin/sh

# Load dummy key
KEYSTORE_PATH="/app/dummy.key.json"
KEYSTORE_PASSWORD="testpassword"

# Wait for registry HTTP endpoint to be ready
while ! wget -q --spider http://registry:8000/p2p/id; do
    echo "Waiting for registry HTTP endpoint..."
    sleep 2
done

# Get registry host ID
REGISTRY_ID=$(curl -s http://registry:8000/p2p/id)
if [ -z "$REGISTRY_ID" ]; then
    echo "Failed to get registry host ID"
    exit 1
fi
echo "Got registry ID: $REGISTRY_ID"

# Wait for registry P2P endpoint to be ready
while ! nc -z registry 9000; do
    echo "Waiting for registry P2P endpoint..."
    sleep 2
done

# Create join message
JOIN_MESSAGE="join_request_$(date +%s)"

# Sign message using operator binary
SIGNATURE=$(./operator -sign -keystore "$KEYSTORE_PATH" -password "$KEYSTORE_PASSWORD" -message "$JOIN_MESSAGE")
ADDRESS=$(./operator -address -keystore "$KEYSTORE_PATH" -password "$KEYSTORE_PASSWORD")

# Create join request
JOIN_REQUEST="{\"address\":\"$ADDRESS\",\"message\":\"$JOIN_MESSAGE\",\"signature\":\"$SIGNATURE\"}"

# Submit join request
echo "Submitting join request..."
RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -d "$JOIN_REQUEST" http://registry:8000/join)

# Check if join was successful
SUCCESS=$(echo $RESPONSE | jq -r '.success')
if [ "$SUCCESS" != "true" ]; then
    echo "Join request failed: $(echo $RESPONSE | jq -r '.error')"
    exit 1
fi

# Construct the full multiaddr
REGISTRY_ADDR="/ip4/172.20.0.2/tcp/9000/p2p/$REGISTRY_ID"
echo "Connecting to registry at: $REGISTRY_ADDR"

# Start the operator with the correct address
exec ./operator -registry "$REGISTRY_ADDR" -keystore "$KEYSTORE_PATH" -password "$KEYSTORE_PASSWORD" 