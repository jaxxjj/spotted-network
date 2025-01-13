#!/bin/sh

# Wait for registry to be up and get its host ID
while true; do
    # Try to get the registry's host ID from its logs
    REGISTRY_ID=$(wget -qO- http://registry:8000/p2p/id 2>/dev/null)
    if [ ! -z "$REGISTRY_ID" ]; then
        break
    fi
    echo "Waiting for registry host ID..."
    sleep 2
done

echo "Found registry ID: $REGISTRY_ID"

# Construct the full multiaddr
REGISTRY_ADDR="/ip4/172.20.0.2/tcp/9000/p2p/$REGISTRY_ID"
echo "Connecting to registry at: $REGISTRY_ADDR"

# Start the operator with the correct address
exec ./operator -registry "$REGISTRY_ADDR" 