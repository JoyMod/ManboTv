# ManBoTV 功能审计报告

## 一、后端 API 与前端路由对照表

### 1.1 公共 API (已对接 ✓)

| 后端路径 | 方法 | 前端路由 | 状态 |
|---------|------|---------|------|
| /api/v1/health | GET | - | ✓ 无需前端路由 |
| /api/v1/auth/login | POST | /api/login | ✓ |
| /api/v1/auth/logout | POST | /api/logout | ✓ |
| /api/v1/search | GET | /api/search | ✓ |
| /api/v1/search/one | GET | /api/search/one | ✓ |
| /api/v1/search/sites | GET | /api/search/resources | ✓ |
| /api/v1/search/suggestions | GET | /api/search/suggestions | ✓ |
| /api/v1/detail | GET | /api/detail | ✓ |
| /api/v1/details | GET | - | ⚠ 未使用 |
| /api/v1/image | GET | - | ✗ **缺失** |
| /api/v1/image/header | GET | - | ✗ **缺失** |
| /api/v1/proxy/m3u8 | GET | /api/proxy/m3u8 | ✓ |
| /api/v1/proxy/segment | GET | /api/proxy/segment | ✓ |
| /api/v1/proxy/key | GET | /api/proxy/key | ✓ |
| /api/v1/proxy/logo | GET | /api/proxy/logo | ✓ |
| /api/v1/douban | GET | /api/douban | ✓ |
| /api/v1/douban/recommends | GET | /api/douban/recommends | ✓ |
| /api/v1/douban/categories | GET | /api/douban/categories | ✓ |

### 1.2 需要认证的 API (部分缺失 ⚠️)

| 后端路径 | 方法 | 前端路由 | 状态 |
|---------|------|---------|------|
| /api/v1/auth/me | GET | - | ⚠ 未使用 |
| /api/v1/auth/password | PUT | - | ⚠ 未使用 |
| /api/v1/favorites | GET | /api/favorites | ✓ |
| /api/v1/favorites | POST | /api/favorites | ✓ |
| /api/v1/favorites/:key | DELETE | - | ✗ **缺失** |
| /api/v1/playrecords | GET | /api/playrecords | ✓ |
| /api/v1/playrecords | POST | /api/playrecords | ✓ |
| /api/v1/playrecords/:key | DELETE | - | ✗ **缺失** |
| /api/v1/searchhistory | GET | /api/searchhistory | ✓ |
| /api/v1/searchhistory | POST | /api/searchhistory | ✓ |
| /api/v1/searchhistory | DELETE | - | ✗ **缺失** |
| /api/v1/skipconfigs | GET | /api/skipconfigs | ✓ |
| /api/v1/skipconfigs | POST | /api/skipconfigs | ✓ |
| /api/v1/skipconfigs | DELETE | - | ✗ **缺失** |
| /api/v1/live/sources | GET | /api/live/sources | ✓ |
| /api/v1/live/channels | GET | /api/live/channels | ✓ |
| /api/v1/live/epg | GET | /api/live/epg | ✓ |
| /api/v1/live/precheck | POST | /api/live/precheck | ✓ |

### 1.3 管理后台 API (新旧版本混乱 ⚠️)

| 后端路径 | 方法 | 前端路由 | 状态 |
|---------|------|---------|------|
| /api/v1/admin/config | GET | /api/admin/config | ✓ |
| /api/v1/admin/config | PUT | - | ✗ **缺失** |
| /api/v1/admin/users | GET | - | ✗ **缺失** |
| /api/v1/admin/users | POST | - | ✗ **缺失** |
| /api/v1/admin/users/:username | PUT | - | ✗ **缺失** |
| /api/v1/admin/users/:username | DELETE | - | ✗ **缺失** |
| /api/v1/admin/sites | GET | - | ✗ **缺失** |
| /api/v1/admin/sites/:key | PUT | - | ✗ **缺失** |
| /api/v1/admin/data-status | GET | - | ⚠ 未使用 |
| /api/v1/admin/data/export | GET | /api/admin/data_migration/export | ✓ |
| /api/v1/admin/data/import | POST | /api/admin/data_migration/import | ✓ |

**旧版兼容路由 (/api/admin/*):**
- 前端使用的是旧版 POST 路由
- 后端新版 RESTful 路由未对接

---

## 二、前端页面功能检查

### 2.1 已完成的页面 ✓

| 页面 | 路径 | 功能状态 |
|------|------|---------|
| 首页 | / | ✓ Netflix风格，数据聚合 |
| 登录页 | /login | ✓ Netflix风格 |
| 电影分类 | /movie | ✓ 分类筛选，无限滚动 |
| 电视剧分类 | /tv | ✓ 分类筛选，无限滚动 |
| 综艺分类 | /variety | ✓ 分类筛选，无限滚动 |
| 动漫分类 | /anime | ✓ 番剧/剧场版/每日放送 |
| 热播榜 | /hot | ✓ 排行榜 |
| 我的片单 | /favorites | ✓ 收藏和播放历史 |
| 搜索页 | /search | ✓ 已有功能 |
| 播放页 | /play | ✓ 已有功能 |
| 直播页 | /live | ✓ 已有功能 |
| 管理后台 | /admin | ⚠ 基础功能完成，部分API需对齐 |

### 2.2 管理后台功能详细检查

| 功能模块 | 状态 | 说明 |
|---------|------|------|
| 概览统计 | ✓ | 用户/视频源/直播源/分类统计 |
| 用户列表 | ✓ | 显示用户列表 |
| 添加用户 | ⚠ | UI完成，需检查API调用 |
| 删除用户 | ⚠ | UI完成，需检查API调用 |
| 修改密码 | ⚠ | UI完成，需检查API调用 |
| 角色切换 | ⚠ | UI完成，需检查API调用 |
| 视频源列表 | ✓ | 显示视频源 |
| 添加视频源 | ⚠ | UI完成，需检查API调用 |
| 禁用/启用视频源 | ⚠ | UI完成，需检查API调用 |
| 直播源列表 | ✓ | 显示直播源 |
| 分类管理 | ⚠ | UI完成，需检查API调用 |
| 站点设置 | ⚠ | UI完成，需检查API调用 |

---

## 三、发现的问题

### 3.1 高优先级问题 (功能缺失)

1. **图片代理 API 缺失**
   - 后端: /api/v1/image, /api/v1/image/header
   - 前端: 无对应路由
   - 影响: 豆瓣图片可能无法显示

2. **DELETE 路由参数格式不统一**
   - 后端使用 `:key` 路径参数
   - 前端使用 body 传参
   - 影响: 删除操作可能失败

3. **新版 Admin API 未对接**
   - 前端使用旧版 `/api/admin/*` POST 路由
   - 后端新版 RESTful API 未使用
   - 影响: 管理后台功能可能异常

### 3.2 中优先级问题 (功能优化)

4. **缺少 /api/auth/me 调用**
   - 用于获取当前用户信息
   - 影响: 用户状态显示可能不准确

5. **缺少 /api/auth/password 调用**
   - 用于修改密码
   - 影响: 密码修改功能缺失

6. **管理后台缺少数据导出/导入 UI**
   - API已存在但无对应界面

---

## 四、修复计划

### Phase 1: API 路由补全 (高优先级)

1. 创建 `/api/image/route.ts` (GET)
2. 创建 `/api/image/header/route.ts` (GET)
3. 修改 `/api/favorites/route.ts` 添加 DELETE 方法支持 key 参数
4. 修改 `/api/playrecords/route.ts` 添加 DELETE 方法支持 key 参数
5. 修改 `/api/searchhistory/route.ts` 添加 DELETE 方法
6. 修改 `/api/skipconfigs/route.ts` 添加 DELETE 方法

### Phase 2: Admin API 对接 (高优先级)

7. 创建 `/api/admin/users/route.ts` (GET, POST)
8. 创建 `/api/admin/users/[username]/route.ts` (PUT, DELETE)
9. 创建 `/api/admin/sites/route.ts` (GET)
10. 创建 `/api/admin/sites/[key]/route.ts` (PUT)
11. 修改 `/api/admin/config/route.ts` 添加 PUT 方法

### Phase 3: 管理后台功能完善 (中优先级)

12. 修复管理后台 API 调用，对齐后端接口
13. 添加数据导出/导入界面
14. 添加修改密码功能

### Phase 4: 前端功能优化 (低优先级)

15. 添加用户头像上传
16. 添加站点统计分析图表
17. 添加操作日志查看

---

## 五、测试检查清单

- [ ] 图片代理正常工作
- [ ] 收藏删除功能正常
- [ ] 播放记录删除功能正常
- [ ] 搜索历史清空功能正常
- [ ] 跳过配置重置功能正常
- [ ] 管理后台用户CRUD正常
- [ ] 管理后台视频源CRUD正常
- [ ] 管理后台站点设置保存正常
