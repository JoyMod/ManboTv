# ManboTV Docker 部署指南

## 📋 目录

- [快速开始](#快速开始)
- [环境要求](#环境要求)
- [生产部署](#生产部署)
- [开发环境](#开发环境)
- [常用命令](#常用命令)
- [故障排查](#故障排查)

## 🚀 快速开始

```bash
# 1. 克隆项目
git clone <your-repo-url>
cd ManboTv

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env 文件，修改管理员密码

# 3. 一键启动
./scripts/docker-start.sh

# 4. 访问应用
# 打开浏览器访问 http://localhost:3000
```

## 📦 环境要求

- Docker >= 20.10
- Docker Compose >= 2.0
- 可用端口: 3000 (Web) / 8080 (API) / 6379 (Redis，可选)

## 🏭 生产部署

### 1. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件：

```env
# 修改默认管理员密码 (必须!)
USERNAME=admin
PASSWORD=your_secure_password_here

# 修改 Web 服务端口 (可选，默认 3000)
WEB_PORT=3000
```

### 2. 构建并启动

```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 3. 验证部署

```bash
# 检查服务状态
docker-compose ps

# 测试 API
curl http://localhost:3000/api/v1/health

# 应该返回:
# {"status":"ok","timestamp":1234567890,"version":"1.0.0"}
```

### 4. 访问应用

- Web 界面: http://localhost:3000
- API 文档: http://localhost:3000/api/v1/health

## 🛠️ 开发环境

使用开发配置启动热重载环境：

```bash
# 开发环境包含:
# - 前端: Next.js dev server (热重载)
# - 后端: Air (Go 热重载)
# - Redis: 数据库

docker-compose -f docker-compose.dev.yml up -d

# 前端访问: http://localhost:3000
# 后端 API: http://localhost:8080
```

## 📝 常用命令

```bash
# 查看日志
docker-compose logs -f           # 所有服务
docker-compose logs -f web       # 仅前端
docker-compose logs -f api       # 仅后端
docker-compose logs -f redis     # 仅 Redis

# 重启服务
docker-compose restart

# 停止服务
docker-compose down

# 停止并删除数据卷 (清空数据)
docker-compose down -v

# 更新镜像
docker-compose pull
docker-compose up -d

# 进入容器
docker exec -it manbotv-web sh      # 前端容器
docker exec -it manbotv-api sh      # 后端容器
docker exec -it manbotv-redis sh    # Redis 容器

# 查看资源使用
docker stats
```

## 🔧 故障排查

### 端口被占用

```bash
# 查看端口占用
lsof -i :3000
lsof -i :8080
lsof -i :6379

# 修改端口 (编辑 .env 文件)
WEB_PORT=8080  # 改为其他端口
```

### 容器无法启动

```bash
# 查看详细日志
docker-compose logs --tail=100

# 检查配置
docker-compose config

# 重新构建
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### Redis 连接失败

```bash
# 检查 Redis 状态
docker-compose ps redis
docker-compose logs redis

# 手动连接测试
docker exec -it manbotv-redis redis-cli ping
```

### API 无法访问

```bash
# 检查后端状态
docker-compose ps api
docker-compose logs api

# 测试后端直连
curl http://localhost:8080/api/v1/health
```

## 📁 目录结构

```
ManboTv/
├── backend/              # Go 后端
│   ├── Dockerfile        # 生产构建
│   ├── Dockerfile.dev    # 开发构建
│   ├── .air.toml         # 热重载配置
│   └── ...
├── scripts/
│   └── docker-start.sh   # 一键启动脚本
├── .env                  # 环境变量 (需创建)
├── .env.example          # 环境变量模板
├── docker-compose.yml    # 生产编排
├── docker-compose.dev.yml # 开发编排
├── Dockerfile            # 前端构建
├── nginx.conf            # Nginx 配置
└── DOCKER_DEPLOY.md      # 本文档
```

## 🔒 安全配置

1. **修改默认密码**: 首次部署前务必修改 `.env` 中的 `PASSWORD`
2. **使用 HTTPS**: 生产环境建议配合反向代理 (Nginx/Traefik) 使用 HTTPS
3. **限制访问**: 使用防火墙限制端口访问
4. **定期备份**: 备份 Redis 数据卷

## 🔄 更新升级

```bash
# 1. 拉取最新代码
git pull

# 2. 重新构建
docker-compose down
docker-compose pull
docker-compose build --no-cache

# 3. 启动服务
docker-compose up -d

# 4. 验证
docker-compose ps
```

## 💾 数据备份

```bash
# 备份 Redis 数据
docker exec manbotv-redis redis-cli BGSAVE
docker cp manbotv-redis:/data/dump.rdb ./backup-$(date +%Y%m%d).rdb

# 恢复数据
docker cp ./backup-xxx.rdb manbotv-redis:/data/dump.rdb
docker-compose restart redis
```

## 📖 更多信息

- [README.md](README.md) - 项目说明
- [API 文档](docs/API_MAPPING.md) - API 接口映射
