#!/bin/bash

# 检查P2P密钥
if [ -z "$P2P_KEY_64" ]; then
    echo "Error: P2P_KEY_64 not set"
    exit 1
fi

# 检查签名方式 - 互斥检查
if [ -n "$SIGNING_KEY_PATH" ] && [ -n "$SIGNING_KEY_PRIV" ]; then
    echo "Error: Cannot use both SIGNING_KEY_PATH and SIGNING_KEY_PRIV"
    exit 1
fi

# 如果都没有设置,报错
if [ -z "$SIGNING_KEY_PATH" ] && [ -z "$SIGNING_KEY_PRIV" ]; then
    echo "Error: Either SIGNING_KEY_PATH or SIGNING_KEY_PRIV must be set"
    exit 1
fi

# 如果使用keystore文件,检查密码和文件
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

# 创建必要的目录
mkdir -p ~/.spotted/config ~/.spotted/keys/signing

# 复制配置文件
cp config/operator.yaml ~/.spotted/config/

# 如果使用keystore文件,复制到指定位置
if [ -n "$SIGNING_KEY_PATH" ]; then
    cp "$SIGNING_KEY_PATH" ~/.spotted/keys/signing/operator.key.json
    export SIGNING_KEY_PATH="/app/keys/signing/operator.key.json"
fi

# 启动服务
docker-compose up -d

echo "Spotted services started"
echo "Check logs with: docker-compose logs -f operator"
echo "Check service status with: docker-compose ps"