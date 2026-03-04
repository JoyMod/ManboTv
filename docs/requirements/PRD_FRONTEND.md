# 前端需求规格文档

## 技术栈

- **框架**: Next.js 14 (App Router)
- **语言**: TypeScript 5
- **样式**: Tailwind CSS 3
- **UI组件**: Headless UI + Radix UI
- **状态管理**: Zustand + React Query (TanStack Query)
- **播放器**: Vidstack + HLS.js
- **HTTP客户端**: Axios

---

## 页面结构

```
/                    - 首页 (推荐/分类)
/search              - 搜索结果页
/detail/:id          - 影片详情页
/play/:id            - 播放页
/favorites           - 我的收藏
/history             - 播放历史
/live                - 电视直播
/admin/*             - 管理后台
/login               - 登录页
```

---

## 状态管理设计

### Zustand Store 划分

```typescript
// stores/
├── authStore.ts      # 认证状态
├── searchStore.ts    # 搜索状态
├── playerStore.ts    # 播放器状态
└── configStore.ts    # 全局配置
```

### React Query Keys

```typescript
['search', query, page]           # 搜索结果
['detail', id]                    # 影片详情
['favorites']                     # 收藏列表
['history']                       # 播放历史
['live', channelId]               # 直播源
```

---

## 组件规范

### 目录结构
```
components/
├── ui/               # 基础UI组件
├── layout/           # 布局组件
├── player/           # 播放器相关
├── search/           # 搜索相关
├── detail/           # 详情页组件
└── admin/            # 后台组件
```

### 命名规范
- 组件: PascalCase (e.g., `SearchBox.tsx`)
- Hooks: camelCase with `use` prefix (e.g., `useSearch.ts`)
- Utils: camelCase (e.g., `formatTime.ts`)

---

## API 调用规范

```typescript
// 统一封装
const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  timeout: 10000,
});

// React Query Hook
export function useSearch(query: string) {
  return useQuery({
    queryKey: ['search', query],
    queryFn: () => api.get(`/search?q=${query}`).then(r => r.data),
    staleTime: 1000 * 60 * 5, // 5分钟
  });
}
```

---

## 性能优化

- 图片懒加载 (next/image)
- 虚拟滚动 (长列表)
- 路由预加载
- 组件懒加载
- Debounce/Throttle (搜索输入)
