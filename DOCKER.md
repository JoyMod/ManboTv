# Docker 部署指南

## 快速开始

### 1. 环境准备

确保已安装 Docker 和 Docker Compose:

```bash
docker --version
docker-compose --version
```

### 2. 配置环境变量

```bash
# 复制配置文件
cp .env.example .env

# 编辑配置 (重要: 修改默认密码!)
nano .env
```

### 3. 构建并启动

```bash
# 构建镜像
docker-compose build

# 启动服务 (后台运行)
docker-compose up -d

# 查看日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f web
docker-compose logs -f backend
```

### 4. 访问应用

- 前端界面: http://localhost:3000
- 后端 API: http://localhost:8080

### 5. 停止服务

```bash
# 停止并删除容器
docker-compose down

# 停止但保留容器
docker-compose stop

# 完全清理 (包括数据卷)
docker-compose down -v
```

## 常用命令

### 查看运行状态

```bash
docker-compose ps
docker-compose top
```

### 重启服务

```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart web
docker-compose restart backend
```

### 更新部署

```bash
# 拉取最新代码后重新构建
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### 进入容器调试

```bash
# 进入前端容器
docker-compose exec web sh

# 进入后端容器
docker-compose exec backend sh

# 进入Redis
docker-compose exec redis redis-cli
```

## 服务架构

```
┌─────────────────────────────────────────────────────────────┐
│                         Docker Network                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │    Web       │  │   Backend    │  │    Redis     │       │
│  │   (Next.js)  │  │    (Go)      │  │   (Cache)    │       │
│  │   Port: 3000 │  │  Port: 8080  │  │  Port: 6379  │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
│         │                 │                 │                │
│         └─────────────────┴─────────────────┘                │
│                    manbotv-network                           │
└─────────────────────────────────────────────────────────────┘
```

## API 代理说明

前端 (`/api/*`) → Next.js API 重写 → Go 后端 (`http://backend:8080`)

配置在 `next.config.js` 中:

```js
async rewrites() {
  return [{
    source: '/api/:path*',
    destination: `${apiProxyTarget}/api/:path*`,
  }];
}
```

## 故障排查

### 1. 容器无法启动

```bash
# 查看详细日志
docker-compose logs --tail=100

# 检查端口占用
sudo lsof -i :3000
sudo lsof -i :8080
```

### 2. 前端无法连接后端

```bash
# 检查网络连接
docker-compose exec web wget -q --spider http://backend:8080/api/health

# 查看后端健康状态
curl http://localhost:8080/api/health
```

### 3. 数据持久化问题

```bash
# 查看数据卷
docker volume ls | grep manbotv

# 备份Redis数据
docker-compose exec redis redis-cli BGSAVE
```

## 生产环境建议

1. **使用反向代理** (Nginx/Traefik) 处理 HTTPS
2. **修改默认密码** 在 `.env` 文件中
3. **定期备份** Redis 数据卷
4. **设置资源限制** 在 `docker-compose.yml` 中

```yaml
services:
  web:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
```
