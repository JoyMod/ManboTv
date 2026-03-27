/* eslint-disable no-console */
'use client';

import {
  type SkipConfig,
  cacheManager,
  dispatchDataUpdate,
  fetchFromApi,
  fetchWithAuth,
  generateStorageKey,
  LocalStorageMode,
  SKIP_CONFIGS_KEY,
  STORAGE_TYPE,
  triggerGlobalError,
} from '@/lib/db/common';

const SkipConfigsEndpoint = '/api/skipconfigs';
const SkipConfigsUpdatedEvent = 'skipConfigsUpdated';

export async function getSkipConfig(
  source: string,
  id: string
): Promise<SkipConfig | null> {
  if (typeof window === 'undefined') {
    return null;
  }

  const key = generateStorageKey(source, id);

  if (STORAGE_TYPE !== LocalStorageMode) {
    const cachedData = cacheManager.getCachedSkipConfigs();
    if (cachedData) {
      fetchFromApi<Record<string, SkipConfig>>(SkipConfigsEndpoint)
        .then((freshData) => {
          if (JSON.stringify(cachedData) === JSON.stringify(freshData)) {
            return;
          }
          cacheManager.cacheSkipConfigs(freshData);
          dispatchDataUpdate(SkipConfigsUpdatedEvent, freshData);
        })
        .catch((error) => {
          console.warn('后台同步跳过片头片尾配置失败:', error);
        });

      return cachedData[key] || null;
    }

    try {
      const freshData = await fetchFromApi<Record<string, SkipConfig>>(
        SkipConfigsEndpoint
      );
      cacheManager.cacheSkipConfigs(freshData);
      return freshData[key] || null;
    } catch (error) {
      console.error('获取跳过片头片尾配置失败:', error);
      triggerGlobalError('获取跳过片头片尾配置失败');
      return null;
    }
  }

  try {
    const raw = localStorage.getItem(SKIP_CONFIGS_KEY);
    if (!raw) return null;
    const configs = JSON.parse(raw) as Record<string, SkipConfig>;
    return configs[key] || null;
  } catch (error) {
    console.error('读取跳过片头片尾配置失败:', error);
    triggerGlobalError('读取跳过片头片尾配置失败');
    return null;
  }
}

export async function saveSkipConfig(
  source: string,
  id: string,
  config: SkipConfig
): Promise<void> {
  const key = generateStorageKey(source, id);

  if (STORAGE_TYPE !== LocalStorageMode) {
    const cachedConfigs = cacheManager.getCachedSkipConfigs() || {};
    cachedConfigs[key] = config;
    cacheManager.cacheSkipConfigs(cachedConfigs);
    dispatchDataUpdate(SkipConfigsUpdatedEvent, cachedConfigs);

    try {
      await fetchWithAuth(SkipConfigsEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ key, config }),
      });
    } catch (error) {
      console.error('保存跳过片头片尾配置失败:', error);
      triggerGlobalError('保存跳过片头片尾配置失败');
    }
    return;
  }

  if (typeof window === 'undefined') {
    console.warn('无法在服务端保存跳过片头片尾配置到 localStorage');
    return;
  }

  try {
    const raw = localStorage.getItem(SKIP_CONFIGS_KEY);
    const configs = raw ? (JSON.parse(raw) as Record<string, SkipConfig>) : {};
    configs[key] = config;
    localStorage.setItem(SKIP_CONFIGS_KEY, JSON.stringify(configs));
    dispatchDataUpdate(SkipConfigsUpdatedEvent, configs);
  } catch (error) {
    console.error('保存跳过片头片尾配置失败:', error);
    triggerGlobalError('保存跳过片头片尾配置失败');
    throw error;
  }
}

export async function getAllSkipConfigs(): Promise<Record<string, SkipConfig>> {
  if (typeof window === 'undefined') {
    return {};
  }

  if (STORAGE_TYPE !== LocalStorageMode) {
    const cachedData = cacheManager.getCachedSkipConfigs();
    if (cachedData) {
      fetchFromApi<Record<string, SkipConfig>>(SkipConfigsEndpoint)
        .then((freshData) => {
          if (JSON.stringify(cachedData) === JSON.stringify(freshData)) {
            return;
          }
          cacheManager.cacheSkipConfigs(freshData);
          dispatchDataUpdate(SkipConfigsUpdatedEvent, freshData);
        })
        .catch((error) => {
          console.warn('后台同步跳过片头片尾配置失败:', error);
          triggerGlobalError('后台同步跳过片头片尾配置失败');
        });

      return cachedData;
    }

    try {
      const freshData = await fetchFromApi<Record<string, SkipConfig>>(
        SkipConfigsEndpoint
      );
      cacheManager.cacheSkipConfigs(freshData);
      return freshData;
    } catch (error) {
      console.error('获取跳过片头片尾配置失败:', error);
      triggerGlobalError('获取跳过片头片尾配置失败');
      return {};
    }
  }

  try {
    const raw = localStorage.getItem(SKIP_CONFIGS_KEY);
    return raw ? (JSON.parse(raw) as Record<string, SkipConfig>) : {};
  } catch (error) {
    console.error('读取跳过片头片尾配置失败:', error);
    triggerGlobalError('读取跳过片头片尾配置失败');
    return {};
  }
}

export async function deleteSkipConfig(
  source: string,
  id: string
): Promise<void> {
  const key = generateStorageKey(source, id);

  if (STORAGE_TYPE !== LocalStorageMode) {
    const cachedConfigs = cacheManager.getCachedSkipConfigs() || {};
    delete cachedConfigs[key];
    cacheManager.cacheSkipConfigs(cachedConfigs);
    dispatchDataUpdate(SkipConfigsUpdatedEvent, cachedConfigs);

    try {
      await fetchWithAuth(
        `${SkipConfigsEndpoint}?key=${encodeURIComponent(key)}`,
        {
          method: 'DELETE',
        }
      );
    } catch (error) {
      console.error('删除跳过片头片尾配置失败:', error);
      triggerGlobalError('删除跳过片头片尾配置失败');
    }
    return;
  }

  if (typeof window === 'undefined') {
    console.warn('无法在服务端删除跳过片头片尾配置到 localStorage');
    return;
  }

  try {
    const raw = localStorage.getItem(SKIP_CONFIGS_KEY);
    if (!raw) return;
    const configs = JSON.parse(raw) as Record<string, SkipConfig>;
    delete configs[key];
    localStorage.setItem(SKIP_CONFIGS_KEY, JSON.stringify(configs));
    dispatchDataUpdate(SkipConfigsUpdatedEvent, configs);
  } catch (error) {
    console.error('删除跳过片头片尾配置失败:', error);
    triggerGlobalError('删除跳过片头片尾配置失败');
    throw error;
  }
}
