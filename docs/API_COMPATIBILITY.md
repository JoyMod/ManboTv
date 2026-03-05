# API 兼容性检查清单

## ❌ 路径不匹配

| 前端调用 | 后端实现 | 状态 | 处理方式 |
|---------|---------|------|---------|
| `/api/login` | `/api/v1/auth/login` | ❌ | 需要适配 |
| `/api/logout` | `/api/v1/auth/logout` | ❌ | 需要适配 |
| `/api/change-password` | `/api/v1/auth/password` | ❌ | 需要适配 |
| `/api/admin/user` | `/api/v1/admin/users` | ❌ | 路径+单复数不同 |
| `/api/admin/site` | `/api/v1/admin/sites` | ❌ | 单复数不同 |
| `/api/admin/data_migration/export` | `/api/v1/admin/data/export` | ❌ | 路径不同 |
| `/api/admin/data_migration/import` | `/api/v1/admin/data/import` | ❌ | 路径不同 |

## ❌ 后端缺失的 API

| 前端调用 | 说明 | 优先级 |
|---------|------|--------|
| `/api/admin/source` | 资源站管理 | 🔴 高 |
| `/api/admin/category` | 自定义分类管理 | 🔴 高 |
| `/api/admin/config_file` | 配置文件管理 | 🔴 高 |
| `/api/admin/config_subscription/fetch` | 订阅获取 | 🟡 中 |
| `/api/admin/live` | 直播源管理 | 🟡 中 |
| `/api/admin/live/refresh` | 直播源刷新 | 🟡 中 |
| `/api/search/ws` | WebSocket 搜索 | 🟢 低 |
| `/api/cron` | 定时任务 | 🟢 低 |
| `/api/server-config` | 服务端配置 | 🟡 中 |

## ✅ 已匹配的 API

| 路径 | 说明 |
|------|------|
| `/api/search` | 搜索 |
| `/api/search/one` | 单源搜索 |
| `/api/search/suggestions` | 搜索建议 |
| `/api/detail` | 详情 |
| `/api/details` | 批量详情 |
| `/api/image-proxy` | 图片代理 (注意: 前端是 image-proxy，后端是 /image) |
| `/api/favorites` | 收藏 |
| `/api/playrecords` | 播放记录 |
| `/api/searchhistory` | 搜索历史 |
| `/api/skipconfigs` | 跳过配置 |
| `/api/live/sources` | 直播源 |
| `/api/live/channels` | 直播频道 |
| `/api/live/epg` | 节目单 |
| `/api/live/precheck` | 直播预检 |
| `/api/douban` | 豆瓣 |
| `/api/proxy/m3u8` | M3U8代理 |
| `/api/proxy/segment` | 片段代理 |

## 🔧 解决方案

### 方案 1: 修改前端 API 路径 (推荐)
修改前端代码，统一使用 `/api/v1` 前缀。

### 方案 2: 后端添加兼容路由
在后端添加旧路径的兼容路由，做 301/302 跳转。

### 方案 3: Nginx 反向代理重写
在 Nginx 配置中添加路径重写规则。
