/**
 * API 配置服务
 * 支持 Next.js API (开发模式) 或 Go 后端 (生产/部署模式)
 */

declare global {
  interface Window {
    RUNTIME_CONFIG?: {
      API_BASE_URL?: string;
    };
  }
}

// API 基础 URL 配置
// 默认使用相对路径 (同域) 或从环境变量读取
const getBaseURL = (): string => {
  // 浏览器端从 window 读取运行时配置
  if (typeof window !== 'undefined') {
    const runtimeConfig = window.RUNTIME_CONFIG;
    if (runtimeConfig?.API_BASE_URL) {
      return runtimeConfig.API_BASE_URL;
    }
  }

  // 从环境变量读取
  const envURL = process.env.NEXT_PUBLIC_API_BASE_URL;
  if (envURL) {
    return envURL;
  }

  // 默认: 使用同域 (适用于 Next.js API 或反向代理)
  return '';
};

// API 版本前缀
const API_VERSION = '/api/v1';

// 获取完整 API URL
export const getApiUrl = (path: string): string => {
  const baseURL = getBaseURL();
  
  // 如果 path 已经是完整 URL，直接返回
  if (path.startsWith('http://') || path.startsWith('https://')) {
    return path;
  }

  // 如果 baseURL 为空，使用相对路径
  if (!baseURL) {
    // 开发模式: 使用 Next.js API
    if (path.startsWith('/api/') && !path.startsWith(API_VERSION)) {
      return path; // 保持原有的 /api/ 路径
    }
    return path;
  }

  // 拼接 baseURL 和 path
  // 移除 baseURL 末尾的斜杠
  const cleanBase = baseURL.replace(/\/$/, '');
  
  // 如果 path 以 /api/v1 开头，直接使用
  if (path.startsWith(API_VERSION)) {
    return `${cleanBase}${path}`;
  }
  
  // 如果 path 以 /api/ 开头，替换为 /api/v1
  if (path.startsWith('/api/')) {
    const newPath = path.replace(/^\/api\//, API_VERSION + '/');
    return `${cleanBase}${newPath}`;
  }

  // 其他路径，添加 API 前缀
  return `${cleanBase}${API_VERSION}${path.startsWith('/') ? path : '/' + path}`;
};

// 封装 fetch 请求
export const apiFetch = async (path: string, options?: RequestInit): Promise<Response> => {
  const url = getApiUrl(path);
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    credentials: 'include', // 携带 Cookie
  });

  return response;
};

// 检查是否使用 Go 后端
export const isGoBackend = (): boolean => {
  const baseURL = getBaseURL();
  return baseURL !== '' && !baseURL.includes('localhost:3000');
};

// 健康检查
export const healthCheck = async (): Promise<boolean> => {
  try {
    const response = await fetch(getApiUrl('/health'), {
      method: 'GET',
      credentials: 'include',
    });
    return response.ok;
  } catch {
    return false;
  }
};

// API 路径映射 (Next.js -> Go)
export const apiPaths = {
  // 认证
  login: '/auth/login',
  logout: '/auth/logout',
  changePassword: '/auth/password',
  me: '/auth/me',

  // 搜索
  search: '/search',
  searchBootstrap: '/search/bootstrap',
  searchOne: '/search/one',
  searchSites: '/search/sites',
  searchSuggestions: '/search/suggestions',

  // 详情
  detail: '/detail',
  details: '/details',
  playBootstrap: '/play/bootstrap',

  // 图片
  image: '/image',
  imageHeader: '/image/header',

  // 代理
  proxyM3U8: '/proxy/m3u8',
  proxySegment: '/proxy/segment',
  proxyKey: '/proxy/key',
  proxyLogo: '/proxy/logo',

  // 收藏
  favorites: '/favorites',
  favoritesBootstrap: '/favorites/bootstrap',

  // 播放记录
  playrecords: '/playrecords',

  // 搜索历史
  searchhistory: '/searchhistory',

  // 跳过配置
  skipconfigs: '/skipconfigs',

  // 豆瓣
  douban: '/douban',
  doubanRecommends: '/douban/recommends',
  doubanCategories: '/douban/categories',

  // 直播
  liveSources: '/live/sources',
  liveChannels: '/live/channels',
  liveEpg: '/live/epg',
  livePrecheck: '/live/precheck',

  // 管理后台
  adminConfig: '/admin/config',
  adminUsers: '/admin/users',
  adminSites: '/admin/sites',
  adminDataStatus: '/admin/data-status',
  adminDataExport: '/admin/data/export',
  adminDataImport: '/admin/data/import',
};
