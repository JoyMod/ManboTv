/* eslint-disable no-console */
'use client';

import {
  cacheManager,
  dispatchDataUpdate,
  fetchFromApi,
  fetchWithAuth,
  handleDatabaseOperationFailure,
  LocalStorageMode,
  SEARCH_HISTORY_KEY,
  SEARCH_HISTORY_LIMIT,
  STORAGE_TYPE,
  triggerGlobalError,
} from '@/lib/db/common';

const SearchHistoryEndpoint = '/api/searchhistory';
const SearchHistoryUpdatedEvent = 'searchHistoryUpdated';

function buildNextHistory(history: string[], keyword: string): string[] {
  const nextHistory = [keyword, ...history.filter((item) => item !== keyword)];
  if (nextHistory.length > SEARCH_HISTORY_LIMIT) {
    nextHistory.length = SEARCH_HISTORY_LIMIT;
  }
  return nextHistory;
}

export async function getSearchHistory(): Promise<string[]> {
  if (typeof window === 'undefined') {
    return [];
  }

  if (STORAGE_TYPE !== LocalStorageMode) {
    const cachedData = cacheManager.getCachedSearchHistory();
    if (cachedData) {
      fetchFromApi<string[]>(SearchHistoryEndpoint)
        .then((freshData) => {
          if (JSON.stringify(cachedData) === JSON.stringify(freshData)) {
            return;
          }
          cacheManager.cacheSearchHistory(freshData);
          dispatchDataUpdate(SearchHistoryUpdatedEvent, freshData);
        })
        .catch((error) => {
          console.warn('后台同步搜索历史失败:', error);
          triggerGlobalError('后台同步搜索历史失败');
        });

      return cachedData;
    }

    try {
      const freshData = await fetchFromApi<string[]>(SearchHistoryEndpoint);
      cacheManager.cacheSearchHistory(freshData);
      return freshData;
    } catch (error) {
      console.error('获取搜索历史失败:', error);
      triggerGlobalError('获取搜索历史失败');
      return [];
    }
  }

  try {
    const raw = localStorage.getItem(SEARCH_HISTORY_KEY);
    if (!raw) return [];
    const history = JSON.parse(raw) as string[];
    return Array.isArray(history) ? history : [];
  } catch (error) {
    console.error('读取搜索历史失败:', error);
    triggerGlobalError('读取搜索历史失败');
    return [];
  }
}

export async function addSearchHistory(keyword: string): Promise<void> {
  const trimmedKeyword = keyword.trim();
  if (!trimmedKeyword) return;

  if (STORAGE_TYPE !== LocalStorageMode) {
    const cachedHistory = cacheManager.getCachedSearchHistory() || [];
    const nextHistory = buildNextHistory(cachedHistory, trimmedKeyword);
    cacheManager.cacheSearchHistory(nextHistory);
    dispatchDataUpdate(SearchHistoryUpdatedEvent, nextHistory);

    try {
      await fetchWithAuth(SearchHistoryEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ keyword: trimmedKeyword }),
      });
    } catch (error) {
      await handleDatabaseOperationFailure('searchHistory', error);
    }
    return;
  }

  if (typeof window === 'undefined') return;

  try {
    const nextHistory = buildNextHistory(
      await getSearchHistory(),
      trimmedKeyword
    );
    localStorage.setItem(SEARCH_HISTORY_KEY, JSON.stringify(nextHistory));
    dispatchDataUpdate(SearchHistoryUpdatedEvent, nextHistory);
  } catch (error) {
    console.error('保存搜索历史失败:', error);
    triggerGlobalError('保存搜索历史失败');
  }
}

export async function clearSearchHistory(): Promise<void> {
  if (STORAGE_TYPE !== LocalStorageMode) {
    cacheManager.cacheSearchHistory([]);
    dispatchDataUpdate(SearchHistoryUpdatedEvent, []);

    try {
      await fetchWithAuth(SearchHistoryEndpoint, {
        method: 'DELETE',
      });
    } catch (error) {
      await handleDatabaseOperationFailure('searchHistory', error);
    }
    return;
  }

  if (typeof window === 'undefined') return;
  localStorage.removeItem(SEARCH_HISTORY_KEY);
  dispatchDataUpdate(SearchHistoryUpdatedEvent, []);
}

export async function deleteSearchHistory(keyword: string): Promise<void> {
  const trimmedKeyword = keyword.trim();
  if (!trimmedKeyword) return;

  if (STORAGE_TYPE !== LocalStorageMode) {
    const cachedHistory = cacheManager.getCachedSearchHistory() || [];
    const nextHistory = cachedHistory.filter(
      (item) => item !== trimmedKeyword
    );
    cacheManager.cacheSearchHistory(nextHistory);
    dispatchDataUpdate(SearchHistoryUpdatedEvent, nextHistory);

    try {
      await fetchWithAuth(
        `${SearchHistoryEndpoint}?keyword=${encodeURIComponent(trimmedKeyword)}`,
        {
          method: 'DELETE',
        }
      );
    } catch (error) {
      await handleDatabaseOperationFailure('searchHistory', error);
    }
    return;
  }

  if (typeof window === 'undefined') return;

  try {
    const nextHistory = (await getSearchHistory()).filter(
      (item) => item !== trimmedKeyword
    );
    localStorage.setItem(SEARCH_HISTORY_KEY, JSON.stringify(nextHistory));
    dispatchDataUpdate(SearchHistoryUpdatedEvent, nextHistory);
  } catch (error) {
    console.error('删除搜索历史失败:', error);
    triggerGlobalError('删除搜索历史失败');
  }
}
