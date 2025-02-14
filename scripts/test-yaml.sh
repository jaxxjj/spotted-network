#!/bin/bash

# 检查yq是否安装
if ! command -v yq &> /dev/null; then
    echo "yq is required. Please install it first."
    echo "Installation instructions: https://github.com/mikefarah/yq#install"
    exit 1
fi

# 检查参数
if [ -z "$1" ]; then
    echo "Usage: $0 <config.yaml>"
    exit 1
fi

CONFIG_FILE="$1"

# 检查文件是否存在
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Config file not found: $CONFIG_FILE"
    exit 1
fi

echo "Reading config from: $CONFIG_FILE"
echo "------------------------"

# 读取并打印配置
echo "Database Configuration:"
echo "Host: $(yq e '.database.host' "$CONFIG_FILE")"
echo "Port: $(yq e '.database.port' "$CONFIG_FILE")"
echo "User: $(yq e '.database.username' "$CONFIG_FILE")"
echo "DB Name: $(yq e '.database.dbname' "$CONFIG_FILE")"

echo -e "\nRedis Configuration:"
echo "Host: $(yq e '.redis.host' "$CONFIG_FILE")"
echo "Port: $(yq e '.redis.port' "$CONFIG_FILE")"

echo -e "\nOther Settings:"
echo "HTTP Port: $(yq e '.http.port' "$CONFIG_FILE")"
echo "P2P Port: $(yq e '.p2p.port' "$CONFIG_FILE")"
echo "Metrics Port: $(yq e '.metric.port' "$CONFIG_FILE")"