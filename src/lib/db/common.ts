/* eslint-disable no-console, @typescript-eslint/no-explicit-any, @typescript-eslint/no-empty-function */
'use client';

import { getAuthInfoFromBrowserCookie } from '@/lib/auth';
import type { Favorite, PlayRecord, SkipConfig } from '@/lib/types';

export type { Favorite, PlayRecord, SkipConfig } from '@/lib/types';

interface CacheData<T> {
  data: T;
  timestamp: number;
  version: string;
}

interface UserCacheStore {
  playRecords?: CacheData<Record<string, PlayRecord>>;
  favorites?: CacheData<Record<string, Favorite>>;
  searchHistory?: CacheData<string[]>;
  skipConfigs?: CacheData<Record<string, SkipConfig>>;
}

export type CacheUpdateEvent =
  | 'playRecordsUpdated'
  | 'favoritesUpdated'
  | 'searchHistoryUpdated'
  | 'skipConfigsUpdated';

type DatabaseDataType = 'playRecords' | 'favorites' | 'searchHistory';

const PLAY_RECORDS_KEY = 'manbotv_play_records';
const FAVORITES_KEY = 'manbotv_favorites';
const SEARCH_HISTORY_KEY = 'manbotv_search_history';
const SKIP_CONFIGS_KEY = 'manbotv_skip_configs';
const CACHE_PREFIX = 'manbotv_cache_';
const CACHE_VERSION = '1.0.0';
const CACHE_EXPIRE_TIME = 60 * 60 * 1000;
const SEARCH_HISTORY_LIMIT = 20;
const LocalStorageMode = 'localstorage';
const CacheCleanupDelayMs = 1000;
const QuotaLimitBytes = 15 * 1024 * 1024;
const MaxCacheAgeMs = 60 * 24 * 60 * 60 * 1000;

export {
  FAVORITES_KEY,
  LocalStorageMode,
  PLAY_RECORDS_KEY,
  SEARCH_HISTORY_KEY,
  SEARCH_HISTORY_LIMIT,
  SKIP_CONFIGS_KEY,
};

export const STORAGE_TYPE = (() => {
  const raw =
    (typeof window !== 'undefined' &&
      (window as any).RUNTIME_CONFIG?.STORAGE_TYPE) ||
    (process.env.STORAGE_TYPE as
      | 'localstorage'
      | 'redis'
      | 'upstash'
      | undefined) ||
    LocalStorageMode;
  return raw;
})();

export function triggerGlobalError(message: string) {
  if (typeof window === 'undefined') return;
  window.dispatchEvent(
    new CustomEvent('globalError', {
      detail: { message },
    })
  );
}

export function dispatchDataUpdate<T>(
  eventType: CacheUpdateEvent,
  detail: T
): void {
  if (typeof window === 'undefined') return;
  window.dispatchEvent(
    new CustomEvent(eventType, {
      detail,
    })
  );
}

class HybridCacheManager {
  private static instance: HybridCacheManager;

  static getInstance(): HybridCacheManager {
    if (!HybridCacheManager.instance) {
      HybridCacheManager.instance = new HybridCacheManager();
    }
    return HybridCacheManager.instance;
  }

  private getCurrentUsername(): string | null {
    const authInfo = getAuthInfoFromBrowserCookie();
    return authInfo?.username || null;
  }

  private getUserCacheKey(username: string): string {
    return `${CACHE_PREFIX}${username}`;
  }

  private getUserCache(username: string): UserCacheStore {
    if (typeof window === 'undefined') return {};

    try {
      const cacheKey = this.getUserCacheKey(username);
      const cached = localStorage.getItem(cacheKey);
      return cached ? JSON.parse(cached) : {};
    } catch (error) {
      console.warn('获取用户缓存失败:', error);
      return {};
    }
  }

  private saveUserCache(username: string, cache: UserCacheStore): void {
    if (typeof window === 'undefined') return;

    try {
      const cacheSize = JSON.stringify(cache).length;
      if (cacheSize > QuotaLimitBytes) {
        console.warn('缓存过大，清理旧数据');
        this.cleanOldCache(cache);
      }

      const cacheKey = this.getUserCacheKey(username);
      localStorage.setItem(cacheKey, JSON.stringify(cache));
    } catch (error) {
      console.warn('保存用户缓存失败:', error);
      if (
        error instanceof DOMException &&
        error.name === 'QuotaExceededError'
      ) {
        this.clearAllCache();
        try {
          const cacheKey = this.getUserCacheKey(username);
          localStorage.setItem(cacheKey, JSON.stringify(cache));
        } catch (retryError) {
          console.error('重试保存缓存仍然失败:', retryError);
        }
      }
    }
  }

  private cleanOldCache(cache: UserCacheStore): void {
    const now = Date.now();

    if (cache.playRecords && now - cache.playRecords.timestamp > MaxCacheAgeMs) {
      delete cache.playRecords;
    }
    if (cache.favorites && now - cache.favorites.timestamp > MaxCacheAgeMs) {
      delete cache.favorites;
    }
  }

  private clearAllCache(): void {
    const keys = Object.keys(localStorage);
    keys.forEach((key) => {
      if (key.startsWith(CACHE_PREFIX)) {
        localStorage.removeItem(key);
      }
    });
  }

  private isCacheValid<T>(cache: CacheData<T>): boolean {
    const now = Date.now();
    return (
      cache.version === CACHE_VERSION &&
      now - cache.timestamp < CACHE_EXPIRE_TIME
    );
  }

  private createCacheData<T>(data: T): CacheData<T> {
    return {
      data,
      timestamp: Date.now(),
      version: CACHE_VERSION,
    };
  }

  getCachedPlayRecords(): Record<string, PlayRecord> | null {
    const username = this.getCurrentUsername();
    if (!username) return null;
    const cached = this.getUserCache(username).playRecords;
    return cached && this.isCacheValid(cached) ? cached.data : null;
  }

  cachePlayRecords(data: Record<string, PlayRecord>): void {
    const username = this.getCurrentUsername();
    if (!username) return;
    const userCache = this.getUserCache(username);
    userCache.playRecords = this.createCacheData(data);
    this.saveUserCache(username, userCache);
  }

  getCachedFavorites(): Record<string, Favorite> | null {
    const username = this.getCurrentUsername();
    if (!username) return null;
    const cached = this.getUserCache(username).favorites;
    return cached && this.isCacheValid(cached) ? cached.data : null;
  }

  cacheFavorites(data: Record<string, Favorite>): void {
    const username = this.getCurrentUsername();
    if (!username) return;
    const userCache = this.getUserCache(username);
    userCache.favorites = this.createCacheData(data);
    this.saveUserCache(username, userCache);
  }

  getCachedSearchHistory(): string[] | null {
    const username = this.getCurrentUsername();
    if (!username) return null;
    const cached = this.getUserCache(username).searchHistory;
    return cached && this.isCacheValid(cached) ? cached.data : null;
  }

  cacheSearchHistory(data: string[]): void {
    const username = this.getCurrentUsername();
    if (!username) return;
    const userCache = this.getUserCache(username);
    userCache.searchHistory = this.createCacheData(data);
    this.saveUserCache(username, userCache);
  }

  getCachedSkipConfigs(): Record<string, SkipConfig> | null {
    const username = this.getCurrentUsername();
    if (!username) return null;
    const cached = this.getUserCache(username).skipConfigs;
    return cached && this.isCacheValid(cached) ? cached.data : null;
  }

  cacheSkipConfigs(data: Record<string, SkipConfig>): void {
    const username = this.getCurrentUsername();
    if (!username) return;
    const userCache = this.getUserCache(username);
    userCache.skipConfigs = this.createCacheData(data);
    this.saveUserCache(username, userCache);
  }

  clearUserCache(username?: string): void {
    const targetUsername = username || this.getCurrentUsername();
    if (!targetUsername) return;

    try {
      const cacheKey = this.getUserCacheKey(targetUsername);
      localStorage.removeItem(cacheKey);
    } catch (error) {
      console.warn('清除用户缓存失败:', error);
    }
  }

  clearExpiredCaches(): void {
    if (typeof window === 'undefined') return;

    try {
      const keysToRemove: string[] = [];

      for (let index = 0; index < localStorage.length; index += 1) {
        const key = localStorage.key(index);
        if (!key?.startsWith(CACHE_PREFIX)) continue;

        try {
          const cache = JSON.parse(localStorage.getItem(key) || '{}');
          let hasValidData = false;
          for (const [, cacheData] of Object.entries(cache)) {
            if (cacheData && this.isCacheValid(cacheData as CacheData<any>)) {
              hasValidData = true;
              break;
            }
          }
          if (!hasValidData) {
            keysToRemove.push(key);
          }
        } catch {
          keysToRemove.push(key);
        }
      }

      keysToRemove.forEach((key) => localStorage.removeItem(key));
    } catch (error) {
      console.warn('清除过期缓存失败:', error);
    }
  }
}

export const cacheManager = HybridCacheManager.getInstance();

export async function fetchWithAuth(
  url: string,
  options?: RequestInit
): Promise<Response> {
  const response = await fetch(url, options);
  if (response.ok) return response;

  if (response.status === 401) {
    try {
      await fetch('/api/logout', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      });
    } catch (error) {
      console.error('注销请求失败:', error);
    }

    const currentUrl = window.location.pathname + window.location.search;
    const loginUrl = new URL('/login', window.location.origin);
    loginUrl.searchParams.set('redirect', currentUrl);
    window.location.href = loginUrl.toString();
    throw new Error('用户未授权，已跳转到登录页面');
  }

  throw new Error(`请求 ${url} 失败: ${response.status}`);
}

export async function fetchFromApi<T>(path: string): Promise<T> {
  const response = await fetchWithAuth(path);
  return (await response.json()) as T;
}

export function generateStorageKey(source: string, id: string): string {
  return `${source}+${id}`;
}

export async function handleDatabaseOperationFailure(
  dataType: DatabaseDataType,
  error: any
): Promise<void> {
  console.error(`数据库操作失败 (${dataType}):`, error);
  triggerGlobalError('数据库操作失败');

  try {
    switch (dataType) {
      case 'playRecords': {
        const freshData = await fetchFromApi<Record<string, PlayRecord>>(
          '/api/playrecords'
        );
        cacheManager.cachePlayRecords(freshData);
        dispatchDataUpdate('playRecordsUpdated', freshData);
        break;
      }
      case 'favorites': {
        const freshData = await fetchFromApi<Record<string, Favorite>>(
          '/api/favorites'
        );
        cacheManager.cacheFavorites(freshData);
        dispatchDataUpdate('favoritesUpdated', freshData);
        break;
      }
      case 'searchHistory': {
        const freshData = await fetchFromApi<string[]>('/api/searchhistory');
        cacheManager.cacheSearchHistory(freshData);
        dispatchDataUpdate('searchHistoryUpdated', freshData);
        break;
      }
    }
  } catch (refreshError) {
    console.error(`刷新${dataType}缓存失败:`, refreshError);
    triggerGlobalError(`刷新${dataType}缓存失败`);
  }
}

export function clearUserCache(): void {
  if (STORAGE_TYPE !== LocalStorageMode) {
    cacheManager.clearUserCache();
  }
}

export async function refreshAllCache(): Promise<void> {
  if (STORAGE_TYPE === LocalStorageMode) return;

  try {
    const [playRecords, favorites, searchHistory, skipConfigs] =
      await Promise.allSettled([
        fetchFromApi<Record<string, PlayRecord>>('/api/playrecords'),
        fetchFromApi<Record<string, Favorite>>('/api/favorites'),
        fetchFromApi<string[]>('/api/searchhistory'),
        fetchFromApi<Record<string, SkipConfig>>('/api/skipconfigs'),
      ]);

    if (playRecords.status === 'fulfilled') {
      cacheManager.cachePlayRecords(playRecords.value);
      dispatchDataUpdate('playRecordsUpdated', playRecords.value);
    }
    if (favorites.status === 'fulfilled') {
      cacheManager.cacheFavorites(favorites.value);
      dispatchDataUpdate('favoritesUpdated', favorites.value);
    }
    if (searchHistory.status === 'fulfilled') {
      cacheManager.cacheSearchHistory(searchHistory.value);
      dispatchDataUpdate('searchHistoryUpdated', searchHistory.value);
    }
    if (skipConfigs.status === 'fulfilled') {
      cacheManager.cacheSkipConfigs(skipConfigs.value);
      dispatchDataUpdate('skipConfigsUpdated', skipConfigs.value);
    }
  } catch (error) {
    console.error('刷新缓存失败:', error);
    triggerGlobalError('刷新缓存失败');
  }
}

export function getCacheStatus(): {
  hasPlayRecords: boolean;
  hasFavorites: boolean;
  hasSearchHistory: boolean;
  hasSkipConfigs: boolean;
  username: string | null;
} {
  if (STORAGE_TYPE === LocalStorageMode) {
    return {
      hasPlayRecords: false,
      hasFavorites: false,
      hasSearchHistory: false,
      hasSkipConfigs: false,
      username: null,
    };
  }

  const authInfo = getAuthInfoFromBrowserCookie();
  return {
    hasPlayRecords: Boolean(cacheManager.getCachedPlayRecords()),
    hasFavorites: Boolean(cacheManager.getCachedFavorites()),
    hasSearchHistory: Boolean(cacheManager.getCachedSearchHistory()),
    hasSkipConfigs: Boolean(cacheManager.getCachedSkipConfigs()),
    username: authInfo?.username || null,
  };
}

export function subscribeToDataUpdates<T>(
  eventType: CacheUpdateEvent,
  callback: (data: T) => void
): () => void {
  if (typeof window === 'undefined') {
    return () => {};
  }

  const handleUpdate = (event: CustomEvent) => {
    callback(event.detail);
  };

  window.addEventListener(eventType, handleUpdate as EventListener);
  return () => {
    window.removeEventListener(eventType, handleUpdate as EventListener);
  };
}

export async function preloadUserData(): Promise<void> {
  if (STORAGE_TYPE === LocalStorageMode) return;

  const status = getCacheStatus();
  if (
    status.hasPlayRecords &&
    status.hasFavorites &&
    status.hasSearchHistory &&
    status.hasSkipConfigs
  ) {
    return;
  }

  refreshAllCache().catch((error) => {
    console.warn('预加载用户数据失败:', error);
    triggerGlobalError('预加载用户数据失败');
  });
}

if (typeof window !== 'undefined') {
  window.setTimeout(() => cacheManager.clearExpiredCaches(), CacheCleanupDelayMs);
}
