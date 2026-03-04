# ManboTV 产品需求文档 (PRD)

## 项目概述

**项目名称**: ManboTV (原 MoonTV 重构版)  
**定位**: 高性能影视聚合播放器  
**目标用户**: 影视爱好者、追剧用户  

### 核心目标
1. **性能提升**: 解决原架构并发差、加载慢的问题
2. **用户体验**: 流畅的搜索、播放、收藏体验
3. **可维护性**: 清晰的代码结构，完善的规范

---

## 系统架构

```
Frontend (Next.js 14 + React 18 + TypeScript)
                    │
                    │ HTTP / WebSocket
                    ▼
              API Gateway
                    │
                    ▼
Backend (Go 1.21+ + Gin + Redis)
```

---

## 功能模块

### 用户端
- 搜索 (多源聚合)
- 播放 (在线播放、播放记录)
- 收藏 (添加/删除)
- 详情 (影片信息)
- 直播

### 管理端
- 源管理
- 直播管理
- 分类管理
- 用户管理

---

## 性能指标

| 指标 | 目标值 |
|-----|--------|
| 首屏加载 | < 2s |
| 搜索响应 | < 1s |
| 并发用户 | 10000+ |
| 服务内存 | < 50MB |

---

## 文档结构

```
docs/requirements/
├── PRD_OVERVIEW.md
├── PRD_FRONTEND.md
├── PRD_BACKEND.md
├── PRD_API.md
└── PRD_DATABASE.md
```
