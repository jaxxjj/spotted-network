#!/bin/bash

# check p2p key
if [ -z "$P2P_KEY_64" ]; then
    echo "Error: P2P_KEY_64 not set"
    exit 1
fi

# check signing method - mutually exclusive check
if [ -n "$SIGNING_KEY_PATH" ] && [ -n "$SIGNING_KEY_PRIV" ]; then
    echo "Error: Cannot use both SIGNING_KEY_PATH and SIGNING_KEY_PRIV"
    exit 1
fi

# if neither is set, report an error
if [ -z "$SIGNING_KEY_PATH" ] && [ -z "$SIGNING_KEY_PRIV" ]; then
    echo "Error: Either SIGNING_KEY_PATH or SIGNING_KEY_PRIV must be set"
    exit 1
fi

# if using keystore file, check password and file
if [ -n "$SIGNING_KEY_PATH" ]; then
    if [ -z "$KEYSTORE_PASSWORD" ]; then
        echo "Error: KEYSTORE_PASSWORD is required when using SIGNING_KEY_PATH"
        exit 1
    fi
    
    if [ ! -f "$SIGNING_KEY_PATH" ]; then
        echo "Error: Keystore file not found: $SIGNING_KEY_PATH"
        exit 1
    fi
fi

# create necessary directories
mkdir -p ~/.spotted/config ~/.spotted/keys/signing

# copy config file
cp config/operator.yaml ~/.spotted/config/

# if using keystore file, copy to specified location
if [ -n "$SIGNING_KEY_PATH" ]; then
    cp "$SIGNING_KEY_PATH" ~/.spotted/keys/signing/operator.key.json
    export SIGNING_KEY_PATH="/app/keys/signing/operator.key.json"
fi

# start services
docker-compose up -d

echo "Spotted services started"
echo "Check logs with: docker-compose logs -f operator"
echo "Check service status with: docker-compose ps"