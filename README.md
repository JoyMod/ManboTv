# ManboTV

ManboTV 是一个基于 `Next.js 14 + Go + Gin + Redis` 的前后端分离影视聚合项目。

当前仓库已经不再是旧版 MoonTV/LunaTV 的原始形态，运行方式、部署方式、后台能力和页面结构都已经按 ManboTV 当前实现重构过，旧 README 中大量关于 LunaTV、Zeabur 一键部署、Kvrocks 组合部署的说明与本项目现状冲突，已不再适用。

## 当前架构

- 前端：`Next.js 14`，负责页面渲染、路由、播放器 UI、管理后台 UI
- 后端：`Go + Gin`，负责搜索聚合、详情解析、图片代理、播放代理、后台配置、内容策略
- 存储：`Redis`，用于管理员配置、用户信息、收藏、播放记录、海报恢复缓存等
- 部署方式：标准 `docker compose` 三容器部署

服务组成：

- `manbotv-web`：前端服务，默认端口 `3000`
- `manbotv-api`：后端 API，容器内端口 `8080`
- `manbotv-redis`：Redis 持久化存储

## 当前能力

- 多源聚合搜索
- 详情页与播放页首屏聚合接口
- HLS 播放代理、分片代理、图片代理
- 收藏、继续观看、搜索历史
- 首页内容流、频道页筛选、搜索结果聚合
- 后台视频源管理、内容模式管理、标签屏蔽
- 首次空配置部署时自动注入默认可用视频源，避免新实例变成空壳

## 快速开始

### 1. 准备环境

需要：

- Docker
- Docker Compose Plugin

检查命令：

```bash
docker --version
docker compose version
```

### 2. 配置环境变量

复制示例配置：

```bash
cp .env.example .env
```

最少需要确认这几个值：

```env
USERNAME=admin
PASSWORD=admin888
PORT=3000
```

说明：

- `USERNAME` / `PASSWORD`：站长账号
- `PORT`：前端对外暴露端口

完整可参考：
- [.env.example](/Users/Zhuanz1/Desktop/project/ManboTv/.env.example)

### 3. 启动服务

```bash
docker compose up -d --build
```

查看状态：

```bash
docker compose ps
```

查看日志：

```bash
docker compose logs -f web
docker compose logs -f api
docker compose logs -f redis
```

### 4. 访问项目

- 前端登录页：[http://localhost:3000/login](http://localhost:3000/login)
- 默认后台账号：使用 `.env` 中配置的 `USERNAME` / `PASSWORD`

## 首次部署说明

当前版本和旧 README 不同：

- 新实例首次启动时，如果 Redis 里没有任何管理员配置，后端会自动写入一份默认视频源配置
- 这样新部署实例默认就可以搜索、看详情、进入播放页，不再是完全空壳
- 后续你仍然可以在后台自行替换、增删、禁用视频源

需要说明的是：

- 这些资源全部来自第三方站点
- 不同源的稳定性、速度、是否可播放会波动
- 某些源可能详情可用但首条线路失效，当前前端已支持自动切换到可播候选源

## 常用命令

重建并重启：

```bash
docker compose up -d --build
```

停止服务：

```bash
docker compose down
```

仅重启单个服务：

```bash
docker compose restart web
docker compose restart api
docker compose restart redis
```

进入容器：

```bash
docker compose exec web sh
docker compose exec api sh
docker compose exec redis redis-cli
```

## 本地开发

前端：

```bash
pnpm install
pnpm dev
```

前端检查：

```bash
pnpm typecheck
pnpm lint
```

后端测试：

```bash
cd backend
go test ./...
```

## API 与页面现状

高频页面：

- `/`：首页
- `/movie`：电影频道
- `/tv`：电视剧频道
- `/variety`：综艺频道
- `/anime`：动漫频道
- `/search`：搜索结果页
- `/play`：播放页
- `/favorites`：片单 / 历史
- `/admin`：后台管理

高频接口：

- `/api/home`
- `/api/browse/bootstrap`
- `/api/search`
- `/api/search/bootstrap`
- `/api/detail`
- `/api/play/bootstrap`
- `/api/proxy/m3u8`
- `/api/proxy/segment`

## Docker 相关文档

如果只看 Docker 说明，可直接看：

- [DOCKER.md](/Users/Zhuanz1/Desktop/project/ManboTv/DOCKER.md)

## 维护说明

如果你准备继续二次开发，建议遵守仓库内规则：

- [AGENTS.md](/Users/Zhuanz1/Desktop/project/ManboTv/AGENTS.md)
- [SKILL.md](/Users/Zhuanz1/Desktop/project/ManboTv/.codex/skills/refactor-moontv/SKILL.md)

其中已经明确要求：

- Go / TS / JS / CSS 文件不得超过 800 行
- 任何 bug 修复后必须执行 `docker compose up -d --build`
- 必须用容器内实际运行版本做页面或接口验证

## 免责声明

本项目本身不存储影视资源，页面中的搜索结果、封面、播放地址、详情数据均来自第三方站点。

请自行评估并承担第三方数据源带来的可用性、合规性和稳定性风险。
