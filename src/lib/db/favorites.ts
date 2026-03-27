/* eslint-disable no-console */
'use client';

import {
  type Favorite,
  cacheManager,
  dispatchDataUpdate,
  FAVORITES_KEY,
  fetchFromApi,
  fetchWithAuth,
  generateStorageKey,
  handleDatabaseOperationFailure,
  LocalStorageMode,
  STORAGE_TYPE,
  triggerGlobalError,
} from '@/lib/db/common';

const FavoritesEndpoint = '/api/favorites';
const FavoritesUpdatedEvent = 'favoritesUpdated';

export async function getAllFavorites(): Promise<Record<string, Favorite>> {
  if (typeof window === 'undefined') {
    return {};
  }

  if (STORAGE_TYPE !== LocalStorageMode) {
    const cachedData = cacheManager.getCachedFavorites();
    if (cachedData) {
      fetchFromApi<Record<string, Favorite>>(FavoritesEndpoint)
        .then((freshData) => {
          if (JSON.stringify(cachedData) === JSON.stringify(freshData)) {
            return;
          }
          cacheManager.cacheFavorites(freshData);
          dispatchDataUpdate(FavoritesUpdatedEvent, freshData);
        })
        .catch((error) => {
          console.warn('后台同步收藏失败:', error);
          triggerGlobalError('后台同步收藏失败');
        });

      return cachedData;
    }

    try {
      const freshData = await fetchFromApi<Record<string, Favorite>>(
        FavoritesEndpoint
      );
      cacheManager.cacheFavorites(freshData);
      return freshData;
    } catch (error) {
      console.error('获取收藏失败:', error);
      triggerGlobalError('获取收藏失败');
      return {};
    }
  }

  try {
    const raw = localStorage.getItem(FAVORITES_KEY);
    return raw ? (JSON.parse(raw) as Record<string, Favorite>) : {};
  } catch (error) {
    console.error('读取收藏失败:', error);
    triggerGlobalError('读取收藏失败');
    return {};
  }
}

export async function saveFavorite(
  source: string,
  id: string,
  favorite: Favorite
): Promise<void> {
  const key = generateStorageKey(source, id);

  if (STORAGE_TYPE !== LocalStorageMode) {
    const cachedFavorites = cacheManager.getCachedFavorites() || {};
    cachedFavorites[key] = favorite;
    cacheManager.cacheFavorites(cachedFavorites);
    dispatchDataUpdate(FavoritesUpdatedEvent, cachedFavorites);

    try {
      await fetchWithAuth(FavoritesEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ key, favorite }),
      });
    } catch (error) {
      await handleDatabaseOperationFailure('favorites', error);
      triggerGlobalError('保存收藏失败');
      throw error;
    }
    return;
  }

  if (typeof window === 'undefined') {
    console.warn('无法在服务端保存收藏到 localStorage');
    return;
  }

  try {
    const allFavorites = await getAllFavorites();
    allFavorites[key] = favorite;
    localStorage.setItem(FAVORITES_KEY, JSON.stringify(allFavorites));
    dispatchDataUpdate(FavoritesUpdatedEvent, allFavorites);
  } catch (error) {
    console.error('保存收藏失败:', error);
    triggerGlobalError('保存收藏失败');
    throw error;
  }
}

export async function deleteFavorite(
  source: string,
  id: string
): Promise<void> {
  const key = generateStorageKey(source, id);

  if (STORAGE_TYPE !== LocalStorageMode) {
    const cachedFavorites = cacheManager.getCachedFavorites() || {};
    delete cachedFavorites[key];
    cacheManager.cacheFavorites(cachedFavorites);
    dispatchDataUpdate(FavoritesUpdatedEvent, cachedFavorites);

    try {
      await fetchWithAuth(`${FavoritesEndpoint}?key=${encodeURIComponent(key)}`, {
        method: 'DELETE',
      });
    } catch (error) {
      await handleDatabaseOperationFailure('favorites', error);
      triggerGlobalError('删除收藏失败');
      throw error;
    }
    return;
  }

  if (typeof window === 'undefined') {
    console.warn('无法在服务端删除收藏到 localStorage');
    return;
  }

  try {
    const allFavorites = await getAllFavorites();
    delete allFavorites[key];
    localStorage.setItem(FAVORITES_KEY, JSON.stringify(allFavorites));
    dispatchDataUpdate(FavoritesUpdatedEvent, allFavorites);
  } catch (error) {
    console.error('删除收藏失败:', error);
    triggerGlobalError('删除收藏失败');
    throw error;
  }
}

export async function isFavorited(
  source: string,
  id: string
): Promise<boolean> {
  const key = generateStorageKey(source, id);

  if (STORAGE_TYPE !== LocalStorageMode) {
    const cachedFavorites = cacheManager.getCachedFavorites();
    if (cachedFavorites) {
      fetchFromApi<Record<string, Favorite>>(FavoritesEndpoint)
        .then((freshData) => {
          if (JSON.stringify(cachedFavorites) === JSON.stringify(freshData)) {
            return;
          }
          cacheManager.cacheFavorites(freshData);
          dispatchDataUpdate(FavoritesUpdatedEvent, freshData);
        })
        .catch((error) => {
          console.warn('后台同步收藏失败:', error);
          triggerGlobalError('后台同步收藏失败');
        });

      return Boolean(cachedFavorites[key]);
    }

    try {
      const freshData = await fetchFromApi<Record<string, Favorite>>(
        FavoritesEndpoint
      );
      cacheManager.cacheFavorites(freshData);
      return Boolean(freshData[key]);
    } catch (error) {
      console.error('检查收藏状态失败:', error);
      triggerGlobalError('检查收藏状态失败');
      return false;
    }
  }

  const allFavorites = await getAllFavorites();
  return Boolean(allFavorites[key]);
}

export async function clearAllFavorites(): Promise<void> {
  if (STORAGE_TYPE !== LocalStorageMode) {
    cacheManager.cacheFavorites({});
    dispatchDataUpdate(FavoritesUpdatedEvent, {});

    try {
      await fetchWithAuth(FavoritesEndpoint, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
      });
    } catch (error) {
      await handleDatabaseOperationFailure('favorites', error);
      triggerGlobalError('清空收藏失败');
      throw error;
    }
    return;
  }

  if (typeof window === 'undefined') return;
  localStorage.removeItem(FAVORITES_KEY);
  dispatchDataUpdate(FavoritesUpdatedEvent, {});
}
