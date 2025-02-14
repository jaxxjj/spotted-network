#!/bin/bash

# 检查必要的参数
if [ -z "$P2P_KEY_64" ]; then
    echo "Error: P2P_KEY_64 not set"
    exit 1
fi

# 检查签名密钥参数(必须提供其中一个)
if [ -z "$SIGNING_KEY_PATH" ] && [ -z "$SIGNING_KEY_PRIV" ]; then
    echo "Error: Either SIGNING_KEY_PATH or SIGNING_KEY_PRIV must be set"
    exit 1
fi

# 如果同时提供了两种签名方式,报错
if [ -n "$SIGNING_KEY_PATH" ] && [ -n "$SIGNING_KEY_PRIV" ]; then
    echo "Error: Cannot use both SIGNING_KEY_PATH and SIGNING_KEY_PRIV"
    exit 1
fi

# 准备docker运行命令
DOCKER_CMD="docker run -d \
    --name spotted-operator \
    --network host \
    -v ~/.spotted/config:/app/config"

# 根据签名方式设置不同的参数
if [ -n "$SIGNING_KEY_PATH" ]; then
    # 使用keystore文件
    if [ -z "$KEYSTORE_PASSWORD" ]; then
        echo "Error: KEYSTORE_PASSWORD is required when using SIGNING_KEY_PATH"
        exit 1
    fi
    
    # 创建密钥目录并复制文件
    mkdir -p ~/.spotted/keys/signing
    cp "$SIGNING_KEY_PATH" ~/.spotted/keys/signing/
    
    # 添加密钥目录挂载
    DOCKER_CMD+=" -v ~/.spotted/keys:/app/keys"
    
    # 设置启动参数
    SIGNING_ARG="--signing-key-path /app/keys/signing/$(basename "$SIGNING_KEY_PATH") --password $KEYSTORE_PASSWORD"
else
    # 使用私钥
    SIGNING_ARG="--signing-key-priv $SIGNING_KEY_PRIV"
fi

# 创建配置目录并复制配置文件
mkdir -p ~/.spotted/config
cp config/operator.yaml ~/.spotted/config/

# 完成docker命令
DOCKER_CMD+=" -p 4014:4014 \
    -p 8080:8080 \
    -p 10000:10000 \
    spotted-operator:latest \
    spotted start \
    --docker-mode \
    $SIGNING_ARG \
    --p2p-key-64 \"$P2P_KEY_64\" \
    --config /app/config/operator.yaml"

# 执行docker命令
eval "$DOCKER_CMD"

echo "Spotted operator container started"
echo "Check logs with: docker logs spotted-operator"