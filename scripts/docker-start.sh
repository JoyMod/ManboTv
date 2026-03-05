#!/bin/bash

# ManboTV Docker 一键启动脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== ManboTV Docker 部署脚本 ===${NC}"
echo ""

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}错误: Docker 未安装${NC}"
    echo "请先安装 Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}错误: Docker Compose 未安装${NC}"
    echo "请先安装 Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
fi

# 检查 .env 文件
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}警告: .env 文件不存在${NC}"
    echo "正在从 .env.example 创建..."
    cp .env.example .env
    echo -e "${YELLOW}请编辑 .env 文件配置管理员密码，然后重新运行此脚本${NC}"
    exit 1
fi

# 加载环境变量
export $(grep -v '^#' .env | xargs)

# 检查密码是否已修改
if [ "$PASSWORD" = "your_secure_password_here" ] || [ "$PASSWORD" = "admin123" ]; then
    echo -e "${YELLOW}警告: 您使用的是默认密码${NC}"
    echo "为了安全起见，请在 .env 文件中修改 PASSWORD"
    echo ""
    read -p "是否继续? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo -e "${GREEN}步骤 1/4: 拉取最新镜像...${NC}"
docker-compose pull

echo ""
echo -e "${GREEN}步骤 2/4: 构建镜像...${NC}"
docker-compose build --no-cache

echo ""
echo -e "${GREEN}步骤 3/4: 启动服务...${NC}"
docker-compose up -d

echo ""
echo -e "${GREEN}步骤 4/4: 等待服务就绪...${NC}"
sleep 5

# 检查服务状态
echo ""
echo "服务状态:"
docker-compose ps

# 获取访问地址
PORT=${WEB_PORT:-3000}
IP=$(hostname -I | awk '{print $1}')

echo ""
echo -e "${GREEN}=== 部署完成! ===${NC}"
echo ""
echo "访问地址:"
echo "  - 本机: http://localhost:$PORT"
echo "  - 局域网: http://$IP:$PORT"
echo ""
echo "默认管理员账号:"
echo "  用户名: ${USERNAME:-admin}"
echo "  密码: ${PASSWORD:-admin123}"
echo ""
echo "常用命令:"
echo "  查看日志: docker-compose logs -f"
echo "  停止服务: docker-compose down"
echo "  重启服务: docker-compose restart"
echo "  更新镜像: docker-compose pull && docker-compose up -d"
