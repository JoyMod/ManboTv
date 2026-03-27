export type {
  CacheUpdateEvent,
  Favorite,
  PlayRecord,
  SkipConfig,
} from '@/lib/db/common';
export {
  clearUserCache,
  getCacheStatus,
  preloadUserData,
  refreshAllCache,
  subscribeToDataUpdates,
} from '@/lib/db/common';
export { generateStorageKey } from '@/lib/db/common';
export {
  clearAllFavorites,
  deleteFavorite,
  getAllFavorites,
  isFavorited,
  saveFavorite,
} from '@/lib/db/favorites';
export {
  clearAllPlayRecords,
  deletePlayRecord,
  getAllPlayRecords,
  savePlayRecord,
} from '@/lib/db/play-records';
export {
  addSearchHistory,
  clearSearchHistory,
  deleteSearchHistory,
  getSearchHistory,
} from '@/lib/db/search-history';
export {
  deleteSkipConfig,
  getAllSkipConfigs,
  getSkipConfig,
  saveSkipConfig,
} from '@/lib/db/skip-configs';
