#!/bin/bash


SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

export PATH="$PATH:$HOME/bin"

cleanup() {
    echo "Cleaning up..."
    
    if [ -d "$HOME/.spotted" ]; then
        cd "$HOME/.spotted" && docker-compose down -v 2>/dev/null || true
    fi
    

    echo "Removing old Docker networks..."
    docker network rm spotted_spotted-network spotted_spotted-net 2>/dev/null || true
    

    rm -rf "$HOME/.spotted"

    rm -f "$HOME/bin/spotted"
    
    # 删除Docker镜像(可选)
    # docker rmi $(docker images -q 'spotted-*') 2>/dev/null
    
    echo "Cleanup complete"
}

# 设置错误处理
set -e
trap cleanup ERR

# 1. 清理旧环境
cleanup

# 2. 安装
echo "Installing spotted..."
bash "$SCRIPT_DIR/install.sh"

echo "Installation complete. You can now:"
echo "1. Run 'spotted init' to initialize the configuration"
echo "2. Start the services with 'cd ~/.spotted && docker-compose up -d'"
echo "3. Check service status with 'docker-compose ps'"
echo ""
echo "To clean up the environment later, run this script again."