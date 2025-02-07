#!/bin/bash

echo "Starting operator with signing key: $SIGNING_KEY_PATH"
echo "Using P2P key: $P2P_KEY_PATH"

# Check if both key files exist
if [ ! -f "$SIGNING_KEY_PATH" ]; then
    echo "Signing key file not found at: $SIGNING_KEY_PATH"
    exit 1
fi

if [ ! -f "$P2P_KEY_PATH" ]; then
    echo "P2P key file not found at: $P2P_KEY_PATH"
    exit 1
fi

# Check if password is needed for ECDSA signing key
if [[ "$SIGNING_KEY_PATH" == *".key.json" ]] && [ -z "$KEYSTORE_PASSWORD" ]; then
    echo "KEYSTORE_PASSWORD environment variable is required for ECDSA signing key"
    exit 1
fi

# Wait for dependencies to be ready
echo "Waiting for dependencies..."

# Wait for postgres to be ready
until PGPASSWORD=$POSTGRES_PASSWORD psql -h "$POSTGRES_HOST" -U "$POSTGRES_USERNAME" -d "$POSTGRES_DBNAME" -c '\q'; do
    echo "Waiting for postgres..."
    sleep 1
done
echo "Postgres is ready"

# Wait for redis to be ready
until redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping > /dev/null 2>&1; do
    echo "Waiting for redis..."
    sleep 1
done
echo "Redis is ready"

# Start the operator node with all required flags
exec ./operator \
    -signing-key "$SIGNING_KEY_PATH" \
    -p2p-key "$P2P_KEY_PATH" \
    -password "$KEYSTORE_PASSWORD"