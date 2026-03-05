#!/bin/bash

# 生产环境构建脚本

set -e

echo "=== ManboTV 生产构建 ==="

# 1. 构建 Go 后端
echo "[1/4] 构建 Go 后端..."
cd backend
go build -o bin/server ./cmd/server/main.go
cd ..

# 2. 构建前端静态文件
echo "[2/4] 构建前端静态文件..."
export EXPORT_MODE=true
export NEXT_PUBLIC_API_BASE_URL=""
npm run build

# 3. 复制运行时配置
echo "[3/4] 复制运行时配置..."
cp public/runtime-config.js dist/

# 4. 构建 Docker 镜像
echo "[4/4] 构建 Docker 镜像..."
docker-compose build

echo ""
echo "=== 构建完成 ==="
echo "启动服务: docker-compose up -d"
echo ""
