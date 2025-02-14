#!/bin/bash

# 检查P2P密钥
if [ -z "$P2P_KEY_64" ]; then
    echo "Error: P2P_KEY_64 not set"
    exit 1
fi

# 交互式输入密钥文件路径
read -p "Please enter the path to your signing key file: " SIGNING_KEY_PATH

# 验证文件是否存在
if [ ! -f "$SIGNING_KEY_PATH" ]; then
    echo "Error: File not found: $SIGNING_KEY_PATH"
    exit 1
fi

# 如果没有设置密码,请求输入
if [ -z "$KEYSTORE_PASSWORD" ]; then
    read -s -p "Please enter the keystore password: " KEYSTORE_PASSWORD
    echo
fi

# 创建必要的目录
mkdir -p ~/.spotted/config ~/.spotted/keys/signing

# 复制配置文件
cp config/operator.yaml ~/.spotted/config/

# 复制密钥文件
cp "$SIGNING_KEY_PATH" ~/.spotted/keys/signing/operator.key.json

# 设置环境变量
export SIGNING_KEY_PATH="/app/keys/signing/operator.key.json"
export KEYSTORE_PASSWORD="$KEYSTORE_PASSWORD"

# 启动服务
docker-compose up -d

echo "Spotted services started"
echo "Check logs with: docker-compose logs -f operator"
echo "Check service status with: docker-compose ps"