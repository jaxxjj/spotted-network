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

# prepare docker run command
DOCKER_CMD="docker run -d \
    --name spotted-operator \
    --network host \
    -v ~/.spotted/config:/app/config"

# create necessary directories
mkdir -p ~/.spotted/config ~/.spotted/keys/signing

# copy config file
cp config/operator.yaml ~/.spotted/config/

# set different parameters based on signing method
if [ -n "$SIGNING_KEY_PATH" ]; then
    # copy key file
    cp "$SIGNING_KEY_PATH" ~/.spotted/keys/signing/operator.key.json
    
    # add key directory mount
    DOCKER_CMD+=" -v ~/.spotted/keys:/app/keys"
    
    # set startup parameters
    SIGNING_ARG="--signing-key-path /app/keys/signing/operator.key.json --password $KEYSTORE_PASSWORD"
else
    # use private key
    SIGNING_ARG="--signing-key-priv $SIGNING_KEY_PRIV"
fi

# complete docker command
DOCKER_CMD+=" -p 4014:4014 \
    -p 8080:8080 \
    -p 10000:10000 \
    spotted-operator:latest \
    spotted start \
    --docker-mode \
    $SIGNING_ARG \
    --p2p-key-64 \"$P2P_KEY_64\" \
    --config /app/config/operator.yaml"

# execute docker command
eval "$DOCKER_CMD"

echo "Spotted operator container started"
echo "Check logs with: docker logs spotted-operator"