# ManboTV

ManboTV 是一个面向实际部署和长期维护的影视聚合项目。  
它不是旧版 MoonTV/LunaTV 的简单换皮，而是把原先偏前端一体化的实现，重构成了 `Next.js 14 + Go + Gin + Redis` 的前后端分离架构。

这次重构的核心目标很明确：

- 让搜索、详情、播放、图片代理这些高频链路更快
- 让多源聚合在并发场景下更稳
- 让部署后的实例不再是“能打开但不好用的空壳”
- 让后台、内容策略、视频源管理真正适合长期维护

如果你只关心一句话：

> **ManboTV 的价值，不是“把旧项目再跑起来”，而是把旧项目里最影响体验的性能瓶颈、部署痛点和维护成本一起解决掉。**

---

## 为什么这个版本更值得用

### 1. Go 后端天然更适合做影视聚合

旧思路里，搜索聚合、详情解析、代理逻辑、图片处理都容易堆在前端或同一层里，随着视频源增多，问题会越来越明显：

- 多源搜索容易拖慢首屏
- 图片代理容易阻塞
- 播放页一旦串太多请求，点击后会明显发钝
- 遇到慢源、坏源、403 源时，前端兜底成本很高

ManboTV 现在把这部分全部下沉到 Go 后端，收益非常直接：

- Go 天然适合做高并发 I/O 场景
- 聚合搜索可以用 Goroutine 并发拉多个资源站
- Context、超时控制、连接复用、并发限制都更容易做干净
- 图片代理、m3u8 代理、segment 代理这种“高频小请求”更适合放在轻量后端层处理

对于影视站来说，这不是“技术栈偏好”，而是实际体验差异：

- 搜索返回更快
- 播放页首屏更稳
- 图片和视频流代理更扛压
- 多源切换和降级更容易做

### 2. 不再让前端自己拼一堆接口

这个版本已经把多个高频页面改成了 bootstrap 聚合接口，而不是让前端自己一次次扇出请求。

现在已经有这些聚合接口：

- `/api/search/bootstrap`
- `/api/play/bootstrap`
- `/api/browse/bootstrap`
- `/api/favorites/bootstrap`

这意味着：

- 搜索页不是前端自己再拼历史、建议词、结果、源状态
- 播放页不是前端自己串详情、收藏状态、换源候选、相关推荐
- 频道页不是前端自己按筛选条件打多次请求再合并

好处很现实：

- 页面点击响应更直接
- 首屏等待更短
- 状态更一致，不容易出现“页面先出来但数据一块一块补”的撕裂感
- 网络请求数量更少，尤其适合 Docker、自建 NAS、OpenWRT 这种本地部署环境

### 3. 播放链路不再只靠“赌当前源”

很多聚合项目都有一个老问题：

- 搜索能搜到
- 详情能打开
- 但真正播放时，第一条线路挂掉，整个页面就像坏了一样

ManboTV 现在不是只展示线路，而是开始做“线路有效性处理”：

- HLS 播放统一走 `/api/proxy/m3u8`
- 分片走 `/api/proxy/segment`
- 支持 m3u8 内容重写
- 支持分片范围请求
- 播放页会维护候选源和测速结果
- 某些源详情正常但首线 403/EOF 时，可以自动切到已验证可播的候选源

这类改动对真实用户体验的影响，比单纯“页面好不好看”大得多。

### 4. 图片处理不是简单转发，而是做了容错和恢复

旧影视聚合站里另一个非常常见的问题，是“资源有了，但封面不出来”。

ManboTV 当前的图片链路已经做了多层处理：

- 后端图片代理
- 连接复用
- 请求合并和缓存
- 按 host 动态处理 Referer/Origin
- 豆瓣图片分片 host 自动轮换
- 图片失败时的封面恢复接口 `/api/poster/recover`
- Redis 缓存恢复结果，避免反复重复搜索

实际效果就是：

- 黑底占位图明显减少
- 同名资源可以从其他源补海报
- 图片加载稳定性比原来更高
- 浏览页、搜索页、播放页的卡片完整度更高

### 5. 新部署实例不再是空壳

旧 README 最大的问题之一，就是默认部署后本质上是“空应用”：

- 服务能起来
- 页面能打开
- 但没有可用视频源
- 搜索、详情、播放全靠你自己进后台补配置

ManboTV 现在已经解决这个问题：

- 首次启动如果 Redis 里没有管理员配置
- 后端会自动注入一份默认可用视频源配置
- 新部署实例默认就能搜索、看详情、进入播放页

这对现场交付、Docker 部署、路由器/OpenWRT 场景都非常重要。

---

## 相比旧版，优势到底是什么

下面这张表是最核心的差异，不绕弯子：

| 维度 | 旧式实现思路 | ManboTV 当前版本 |
| --- | --- | --- |
| 架构 | 前端偏一体化，接口与页面耦合重 | 前后端分离，Next.js 负责 UI，Go 负责聚合与代理 |
| 多源搜索 | 前端/单层逻辑更容易拖慢首屏 | Go 并发聚合 + 超时控制 + 降级返回 |
| 播放首屏 | 前端自己拼多次请求 | `/api/play/bootstrap` 一次返回首屏核心数据 |
| 频道页筛选 | 前端多次请求再合并 | `/api/browse/bootstrap` 后端聚合结果 |
| 搜索页体验 | 结果、历史、建议词分散拉取 | `/api/search/bootstrap` 统一返回 |
| 图片代理 | 容易慢、容易失效、失败后黑图 | 代理 + 缓存 + 恢复 + 海报换源 |
| 播放容错 | 当前线路失败容易直接黑屏 | 候选源测试 + 自动切线 |
| 首次部署 | 可能是空壳，没有资源 | 空配置自动注入默认视频源 |
| 后台能力 | 配置入口弱、策略控制不足 | 视频源管理、内容模式、标签屏蔽、用户控制 |
| 维护性 | 文件膨胀、职责混乱 | 模块拆分、接口分层、规则约束明确 |

---

## 重构的性能方向

这个仓库不是只想“看起来更现代”，而是明确奔着性能去的。

在重构文档里，目标写得很清楚：

- 并发能力目标：`10x`
- 内存占用目标：`5x`
- 冷启动目标：`5x`
- 图片代理耗时目标：百毫秒级

这些目标并不表示所有部署环境、所有资源站、所有网络条件下都能稳定跑出同样数字，但方向是明确的：

- **搜索要更并发**
- **播放要更轻首屏**
- **图片代理要更快更稳**
- **服务要更适合低成本自建环境**

对影视聚合项目来说，这些比“多一个炫技 UI 动效”更有价值。

---

## 当前架构

### 技术栈

- 前端：`Next.js 14`
- 后端：`Go + Gin`
- 存储：`Redis`
- 播放：`HLS.js + ArtPlayer`
- 部署：`Docker Compose`

### 服务组成

- `manbotv-web`
  - 前端页面服务
  - 默认端口 `3000`
- `manbotv-api`
  - 搜索、详情、代理、后台 API
  - 容器内端口 `8080`
- `manbotv-redis`
  - 管理配置、用户数据、播放记录、收藏、缓存

### 现在的职责划分

前端主要负责：

- 页面 UI
- 交互状态
- 播放器组件
- 管理后台页面

后端主要负责：

- 多源聚合搜索
- 详情解析
- m3u8 代理
- segment 代理
- 图片代理
- 海报恢复
- 内容标签和分级策略
- 后台配置持久化

---

## 当前版本已经具备的能力

### 内容获取与页面能力

- 多源聚合搜索
- 首页内容流
- 电影 / 电视剧 / 综艺 / 动漫频道页
- 搜索结果聚合展示
- 详情页与播放页联动
- 收藏、继续观看、搜索历史

### 播放相关能力

- HLS 播放代理
- 分片代理与 Range 支持
- m3u8 内容重写
- 换源候选管理
- 当前源测速
- 候选源测速
- 播放失败自动回退到可播线路

### 图片与封面能力

- 统一图片代理
- 海报恢复接口
- Redis 缓存恢复结果
- 占位图与封面换源
- 豆瓣图与第三方图床容错

### 后台与策略能力

- 视频源管理
- 默认源自动初始化
- 内容模式切换
  - `safe`
  - `mixed`
  - `adult_only`
- 标签屏蔽
- 用户与权限配置基础能力

---

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

完整示例见：

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

- 登录页：[http://localhost:3000/login](http://localhost:3000/login)
- 默认后台账号：使用 `.env` 中配置的 `USERNAME` / `PASSWORD`

---

## 首次部署说明

这个版本和旧 README 最大的差别之一，就是首次部署体验已经变了。

现在的行为是：

- 新实例首次启动时，如果 Redis 中没有管理员配置
- 后端会自动写入默认视频源配置
- 页面默认可搜、可进详情、可进入播放页

这意味着：

- 不再需要“先部署，再手工补一大堆配置，最后才能用”
- 更适合现场交付
- 更适合 NAS、自建机、Docker 宿主机、OpenWRT 路由器场景

当然也要说明：

- 数据源来自第三方站点
- 不同源的可用性会波动
- 某些源可能详情可用但首条线路失效
- 当前版本已经支持候选源回退，但第三方站点本身的不稳定仍然无法完全消除

---

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

---

## 本地开发

安装依赖并启动前端：

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

---

## API 与页面现状

### 高频页面

- `/`
- `/movie`
- `/tv`
- `/variety`
- `/anime`
- `/search`
- `/play`
- `/favorites`
- `/admin`

### 高频接口

- `/api/home`
- `/api/browse/bootstrap`
- `/api/search`
- `/api/search/bootstrap`
- `/api/detail`
- `/api/play/bootstrap`
- `/api/proxy/m3u8`
- `/api/proxy/segment`
- `/api/poster/recover`

---

## Docker 相关文档

如果你只想看 Docker 说明，直接看：

- [DOCKER.md](/Users/Zhuanz1/Desktop/project/ManboTv/DOCKER.md)

---

## 维护说明

如果你准备继续二次开发，建议先看仓库规则：

- [AGENTS.md](/Users/Zhuanz1/Desktop/project/ManboTv/AGENTS.md)
- [SKILL.md](/Users/Zhuanz1/Desktop/project/ManboTv/.codex/skills/refactor-moontv/SKILL.md)

当前仓库已经明确要求：

- Go / TS / JS / CSS 文件不得超过 800 行
- 任何 bug 修复后必须执行 `docker compose up -d --build`
- 必须用容器内实际运行版本做页面或接口验证

---

## 免责声明

本项目本身不存储影视资源。  
页面中的搜索结果、封面、播放地址、详情数据均来自第三方站点。

请自行评估并承担第三方数据源带来的可用性、稳定性与合规性风险。
