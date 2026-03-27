'use client';

export interface BrowseItemRecord {
  id: string;
  title: string;
  cover: string;
  rate: string;
  year: string;
  type?: 'movie' | 'tv' | 'variety' | 'anime';
}

export type BrowseFilterState = Record<string, string | string[]>;

interface DoubanListItem {
  id?: string | number;
  title?: string;
  poster?: string;
  cover?: string;
  rate?: string;
  year?: string;
}

interface BrowseFetchPlan {
  fetchTags: string[];
  fetchGroups: Record<string, string[]>;
}

const DEFAULT_ID_LENGTH = 9;

function uniqueValues(values: string[]): string[] {
  return Array.from(new Set(values.filter(Boolean)));
}

function toValues(value: string | string[]): string[] {
  return Array.isArray(value) ? value : [value];
}

function isFetchableYearValue(value: string): boolean {
  return /^\d{4}$/.test(value) || /^\d{4}-/.test(value);
}

function normalizeYearTag(value: string): string {
  if (/^\d{4}-/.test(value)) {
    return value.slice(0, 4);
  }
  return value;
}

function parseYear(year: string): number {
  const parsed = Number.parseInt(year, 10);
  return Number.isFinite(parsed) ? parsed : 0;
}

function matchesYearBucket(itemYear: string, filterValue: string): boolean {
  if (!filterValue || filterValue === '全部') return true;

  const year = parseYear(itemYear);
  if (!year) return false;

  if (/^\d{4}$/.test(filterValue)) return year === Number.parseInt(filterValue, 10);
  if (/^\d{4}-/.test(filterValue)) return year === Number.parseInt(filterValue.slice(0, 4), 10);
  if (filterValue === '2010s') return year >= 2010 && year <= 2019;
  if (filterValue === '2000s') return year >= 2000 && year <= 2009;
  if (filterValue === '90s') return year >= 1990 && year <= 1999;
  if (filterValue === 'earlier') return year < 1990;
  if (filterValue === '经典') return year <= 2019;
  return true;
}

export function buildBrowseFetchPlan(
  selectedFilters: BrowseFilterState,
  defaultTag: string
): BrowseFetchPlan {
  const fetchGroups: Record<string, string[]> = {};

  Object.entries(selectedFilters).forEach(([groupId, rawValue]) => {
    const values = toValues(rawValue).filter((value) => value && value !== '全部');
    if (values.length === 0) return;

    if (groupId === 'year') {
      fetchGroups[groupId] = uniqueValues(
        values.filter(isFetchableYearValue).map(normalizeYearTag)
      );
      return;
    }

    fetchGroups[groupId] = uniqueValues(values);
  });

  const fetchTags = uniqueValues(
    Object.values(fetchGroups)
      .flat()
      .filter(Boolean)
  );

  return {
    fetchTags: fetchTags.length > 0 ? fetchTags : [defaultTag],
    fetchGroups,
  };
}

export function mapBrowseItems(
  list: DoubanListItem[],
  type: BrowseItemRecord['type']
): BrowseItemRecord[] {
  return list.map((item) => ({
    id:
      item.id?.toString() ||
      Math.random().toString(36).slice(-DEFAULT_ID_LENGTH),
    title: item.title || '未知标题',
    cover: item.poster || item.cover || '/placeholder-poster.svg',
    rate: item.rate || '',
    year: item.year || '',
    type,
  }));
}

export function mergeBrowseItems(
  tagBuckets: Record<string, BrowseItemRecord[]>,
  fetchPlan: BrowseFetchPlan,
  selectedFilters: BrowseFilterState,
  defaultTag: string,
  sortItems: (list: BrowseItemRecord[]) => BrowseItemRecord[]
): BrowseItemRecord[] {
  const activeFetchGroups = Object.entries(fetchPlan.fetchGroups).filter(
    ([, tags]) => tags.length > 0
  );

  const groupSets = activeFetchGroups.map(([, tags]) => {
    const idSet = new Set<string>();
    tags.forEach((tag) => {
      (tagBuckets[tag] || []).forEach((item) => {
        idSet.add(item.id);
      });
    });
    return idSet;
  });

  const candidateMap = new Map<string, BrowseItemRecord>();
  const sourceTags =
    activeFetchGroups.length > 0 ? fetchPlan.fetchTags : [defaultTag];

  sourceTags.forEach((tag) => {
    (tagBuckets[tag] || []).forEach((item) => {
      candidateMap.set(item.id, item);
    });
  });

  const filtered = Array.from(candidateMap.values()).filter((item) => {
    const matchesFetchGroups = groupSets.every((groupSet) => groupSet.has(item.id));
    if (!matchesFetchGroups) return false;

    const yearFilter = selectedFilters.year;
    if (typeof yearFilter === 'string') {
      return matchesYearBucket(item.year, yearFilter);
    }
    return true;
  });

  return sortItems(filtered);
}
