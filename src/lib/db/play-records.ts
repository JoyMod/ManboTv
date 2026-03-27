/* eslint-disable no-console */
'use client';

import {
  type PlayRecord,
  cacheManager,
  dispatchDataUpdate,
  fetchFromApi,
  fetchWithAuth,
  generateStorageKey,
  handleDatabaseOperationFailure,
  LocalStorageMode,
  PLAY_RECORDS_KEY,
  STORAGE_TYPE,
  triggerGlobalError,
} from '@/lib/db/common';

const PlayRecordsEndpoint = '/api/playrecords';
const PlayRecordsUpdatedEvent = 'playRecordsUpdated';

export async function getAllPlayRecords(): Promise<Record<string, PlayRecord>> {
  if (typeof window === 'undefined') {
    return {};
  }

  if (STORAGE_TYPE !== LocalStorageMode) {
    const cachedData = cacheManager.getCachedPlayRecords();
    if (cachedData) {
      fetchFromApi<Record<string, PlayRecord>>(PlayRecordsEndpoint)
        .then((freshData) => {
          if (JSON.stringify(cachedData) === JSON.stringify(freshData)) {
            return;
          }
          cacheManager.cachePlayRecords(freshData);
          dispatchDataUpdate(PlayRecordsUpdatedEvent, freshData);
        })
        .catch((error) => {
          console.warn('后台同步播放记录失败:', error);
          triggerGlobalError('后台同步播放记录失败');
        });

      return cachedData;
    }

    try {
      const freshData = await fetchFromApi<Record<string, PlayRecord>>(
        PlayRecordsEndpoint
      );
      cacheManager.cachePlayRecords(freshData);
      return freshData;
    } catch (error) {
      console.error('获取播放记录失败:', error);
      triggerGlobalError('获取播放记录失败');
      return {};
    }
  }

  try {
    const raw = localStorage.getItem(PLAY_RECORDS_KEY);
    return raw ? (JSON.parse(raw) as Record<string, PlayRecord>) : {};
  } catch (error) {
    console.error('读取播放记录失败:', error);
    triggerGlobalError('读取播放记录失败');
    return {};
  }
}

export async function savePlayRecord(
  source: string,
  id: string,
  record: PlayRecord
): Promise<void> {
  const key = generateStorageKey(source, id);

  if (STORAGE_TYPE !== LocalStorageMode) {
    const cachedRecords = cacheManager.getCachedPlayRecords() || {};
    cachedRecords[key] = record;
    cacheManager.cachePlayRecords(cachedRecords);
    dispatchDataUpdate(PlayRecordsUpdatedEvent, cachedRecords);

    try {
      await fetchWithAuth(PlayRecordsEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ key, record }),
      });
    } catch (error) {
      await handleDatabaseOperationFailure('playRecords', error);
      triggerGlobalError('保存播放记录失败');
      throw error;
    }
    return;
  }

  if (typeof window === 'undefined') {
    console.warn('无法在服务端保存播放记录到 localStorage');
    return;
  }

  try {
    const allRecords = await getAllPlayRecords();
    allRecords[key] = record;
    localStorage.setItem(PLAY_RECORDS_KEY, JSON.stringify(allRecords));
    dispatchDataUpdate(PlayRecordsUpdatedEvent, allRecords);
  } catch (error) {
    console.error('保存播放记录失败:', error);
    triggerGlobalError('保存播放记录失败');
    throw error;
  }
}

export async function deletePlayRecord(
  source: string,
  id: string
): Promise<void> {
  const key = generateStorageKey(source, id);

  if (STORAGE_TYPE !== LocalStorageMode) {
    const cachedRecords = cacheManager.getCachedPlayRecords() || {};
    delete cachedRecords[key];
    cacheManager.cachePlayRecords(cachedRecords);
    dispatchDataUpdate(PlayRecordsUpdatedEvent, cachedRecords);

    try {
      await fetchWithAuth(`${PlayRecordsEndpoint}?key=${encodeURIComponent(key)}`, {
        method: 'DELETE',
      });
    } catch (error) {
      await handleDatabaseOperationFailure('playRecords', error);
      triggerGlobalError('删除播放记录失败');
      throw error;
    }
    return;
  }

  if (typeof window === 'undefined') {
    console.warn('无法在服务端删除播放记录到 localStorage');
    return;
  }

  try {
    const allRecords = await getAllPlayRecords();
    delete allRecords[key];
    localStorage.setItem(PLAY_RECORDS_KEY, JSON.stringify(allRecords));
    dispatchDataUpdate(PlayRecordsUpdatedEvent, allRecords);
  } catch (error) {
    console.error('删除播放记录失败:', error);
    triggerGlobalError('删除播放记录失败');
    throw error;
  }
}

export async function clearAllPlayRecords(): Promise<void> {
  if (STORAGE_TYPE !== LocalStorageMode) {
    cacheManager.cachePlayRecords({});
    dispatchDataUpdate(PlayRecordsUpdatedEvent, {});

    try {
      await fetchWithAuth(PlayRecordsEndpoint, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
      });
    } catch (error) {
      await handleDatabaseOperationFailure('playRecords', error);
      triggerGlobalError('清空播放记录失败');
      throw error;
    }
    return;
  }

  if (typeof window === 'undefined') return;
  localStorage.removeItem(PLAY_RECORDS_KEY);
  dispatchDataUpdate(PlayRecordsUpdatedEvent, {});
}
