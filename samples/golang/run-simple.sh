#!/bin/bash
# 运行交互式对话示例

set -e

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# 加载环境变量
if [ -f "$PROJECT_ROOT/.env" ]; then
    export $(grep -v '^#' "$PROJECT_ROOT/.env" | xargs)
else
    echo "❌ 未找到 .env 文件: $PROJECT_ROOT/.env"
    exit 1
fi


cd "$SCRIPT_DIR"
go run ./cmd/chat/
