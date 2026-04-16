'use client';

export interface SearchResult {
  id: string;
  title: string;
  cover: string;
  year: string;
  rating?: string;
  remarks?: string;
  type: 'movie' | 'tv' | 'variety' | 'anime';
  source?: string;
  sourceName?: string;
  episodes: string[];
  episodesTitles: string[];
  tags?: string[];
  isAdult?: boolean;
  matchScore?: number;
  matchReasons?: string[];
}

export interface RawSearchItem {
  id?: string;
  title?: string;
  poster?: string;
  cover?: string;
  year?: string;
  rate?: string;
  remarks?: string;
  source?: string;
  source_name?: string;
  class?: string;
  type_name?: string;
  episodes?: string[];
  episodes_titles?: string[];
  tags?: string[];
  is_adult?: boolean;
  match_score?: number;
  match_reasons?: string[];
}

export interface SourceTestResult {
  status: 'idle' | 'testing' | 'ok' | 'error';
  pingMs?: number;
  quality?: string;
  speed?: string;
}

export interface StreamMessage {
  type?: string;
  source?: string;
  sourceName?: string;
  error?: string;
  results?: RawSearchItem[];
}

export interface SearchBootstrapPayload {
  query?: string;
  normalized_query?: string;
  results?: RawSearchItem[];
  aggregates?: RawSearchAggregateGroup[];
  facets?: SearchFacets;
  history?: string[];
  suggestions?: string[];
  source_status?: Record<string, 'done' | 'error'>;
  source_status_items?: SearchSourceStatusItem[];
  page_info?: SearchPageInfo;
  execution?: SearchExecutionInfo;
  selected_types?: string[];
  selected_sources?: string[];
  selected_sort?: string;
  selected_view?: string;
  selected_year_from?: number;
  selected_year_to?: number;
  selected_source_mode?: string;
}

export interface SuggestionItem {
  text?: string;
}

export interface SearchAggregateGroup {
  key: string;
  title: string;
  year: string;
  type: SearchResult['type'];
  cover: string;
  rating?: string;
  sourceCount: number;
  resultCount: number;
  bestSource?: string;
  bestSourceName?: string;
  tags?: string[];
  items: SearchResult[];
}

export interface RawSearchAggregateGroup {
  key?: string;
  title?: string;
  year?: string;
  type?: SearchResult['type'];
  cover?: string;
  rating?: string;
  source_count?: number;
  result_count?: number;
  best_source?: string;
  best_source_name?: string;
  tags?: string[];
  items?: RawSearchItem[];
}

export interface SearchFacetBucket {
  value: string;
  label: string;
  count: number;
}

export interface SearchFacets {
  types?: SearchFacetBucket[];
  sources?: SearchFacetBucket[];
  years?: SearchFacetBucket[];
}

export interface SearchSourceStatusItem {
  source: string;
  source_name?: string;
  status: 'done' | 'empty' | 'partial' | 'timeout' | 'error';
  result_count?: number;
  page_count?: number;
  elapsed_ms?: number;
  error?: string;
}

export interface SearchPageInfo {
  page?: number;
  page_size?: number;
  total?: number;
  total_pages?: number;
}

export interface SearchExecutionInfo {
  query?: string;
  normalized_query?: string;
  completed_sources?: number;
  total_sources?: number;
  elapsed_ms?: number;
  degraded?: boolean;
  streaming_enabled?: boolean;
}

export const HOT_SEARCHES = [
  '繁花',
  '三大队',
  '年会不能停',
  '周处除三害',
  '沙丘2',
  '热辣滚烫',
  '第二十条',
  '飞驰人生2',
];

const DEFAULT_TIMEOUT_MS = 8000;
const DEFAULT_ID_SUFFIX_LENGTH = 8;
const FOUR_K_SCORE = 100;
const UHD_SCORE = 100;
const FULL_HD_SCORE = 90;
const HD_SCORE = 75;
const SD_SCORE = 55;
const UNKNOWN_QUALITY_SCORE = 40;

export function mapContentType(
  item: RawSearchItem
): 'movie' | 'tv' | 'variety' | 'anime' {
  const text = `${item.type_name || ''} ${item.class || ''}`.toLowerCase();
  if (text.includes('动漫') || text.includes('anime')) return 'anime';
  if (text.includes('综艺') || text.includes('variety')) return 'variety';
  if (
    text.includes('剧') ||
    text.includes('tv') ||
    text.includes('连续') ||
    text.includes('美剧') ||
    text.includes('韩剧')
  ) {
    return 'tv';
  }
  return 'movie';
}

export function normalizeTitle(title: string): string {
  return (title || '')
    .toLowerCase()
    .replace(/[\s\-_.:：!！?？,，。·'"“”‘’]/g, '')
    .trim();
}

export function buildAggregateKey(
  item: Pick<SearchResult, 'title' | 'year' | 'type'>
): string {
  return `${normalizeTitle(item.title)}-${item.year || 'unknown'}-${item.type}`;
}

export function parseYearValue(year: string): number {
  const parsed = Number.parseInt(year, 10);
  return Number.isFinite(parsed) ? parsed : 0;
}

export function parseQuality(manifest: string): string {
  const match = manifest.match(/RESOLUTION=(\d+)x(\d+)/i);
  if (match?.[2]) return `${match[2]}p`;
  const upper = manifest.toUpperCase();
  if (upper.includes('4K') || upper.includes('2160')) return '4K';
  if (upper.includes('1080')) return '1080p';
  if (upper.includes('720')) return '720p';
  if (upper.includes('480')) return '480p';
  return '未知';
}

export function isValidM3U8(content: string): boolean {
  return /#EXTM3U/i.test(content || '');
}

export function qualityScore(quality: string): number {
  if (quality.includes('4K')) return FOUR_K_SCORE;
  if (quality.includes('2160')) return UHD_SCORE;
  if (quality.includes('1080')) return FULL_HD_SCORE;
  if (quality.includes('720')) return HD_SCORE;
  if (quality.includes('480')) return SD_SCORE;
  return UNKNOWN_QUALITY_SCORE;
}

export function normalizeItems(
  rawItems: RawSearchItem[],
  normalizeImageUrl: (url?: string | null) => string
): SearchResult[] {
  const list: SearchResult[] = [];

  rawItems.forEach((item, index) => {
    const normalizedCover = normalizeImageUrl(item.poster || item.cover || '');
    if (!normalizedCover) return;

    list.push({
      id:
        item.id?.toString() ||
        `search-${index}-${Math.random().toString(36).slice(2, DEFAULT_ID_SUFFIX_LENGTH)}`,
      title: item.title || '未知标题',
      cover: normalizedCover,
      year: item.year || '',
      rating: item.rate || item.remarks || '',
      remarks: item.remarks || '',
      type: mapContentType(item),
      source: item.source,
      sourceName: item.source_name || '',
      episodes: Array.isArray(item.episodes) ? item.episodes : [],
      episodesTitles: Array.isArray(item.episodes_titles)
        ? item.episodes_titles
        : [],
      tags: Array.isArray(item.tags) ? item.tags : [],
      isAdult: Boolean(item.is_adult),
      matchScore: Number.isFinite(item.match_score) ? item.match_score : 0,
      matchReasons: Array.isArray(item.match_reasons)
        ? item.match_reasons.filter((reason) => typeof reason === 'string')
        : [],
    });
  });

  return list;
}

export function normalizeAggregateGroups(
  rawGroups: RawSearchAggregateGroup[],
  normalizeImageUrl: (url?: string | null) => string
): SearchAggregateGroup[] {
  return rawGroups
    .map((group) => ({
      key: group.key || '',
      title: group.title || '未知标题',
      year: group.year || '',
      type: (group.type || 'movie') as SearchResult['type'],
      cover: normalizeImageUrl(group.cover || ''),
      rating: group.rating || '',
      sourceCount: Number(group.source_count || 0),
      resultCount: Number(group.result_count || 0),
      bestSource: group.best_source || '',
      bestSourceName: group.best_source_name || '',
      tags: Array.isArray(group.tags) ? group.tags : [],
      items: normalizeItems(
        Array.isArray(group.items) ? group.items : [],
        normalizeImageUrl
      ),
    }))
    .filter((group) => Boolean(group.key) && Boolean(group.cover));
}

export function normalizeSearchSourceStatuses(
  items: SearchSourceStatusItem[] | undefined
): SearchSourceStatusItem[] {
  if (!Array.isArray(items)) return [];
  return items
    .filter((item) => item && typeof item.source === 'string')
    .map((item) => ({
      source: item.source,
      source_name: item.source_name || item.source,
      status: item.status || 'error',
      result_count: Number(item.result_count || 0),
      page_count: Number(item.page_count || 0),
      elapsed_ms: Number(item.elapsed_ms || 0),
      error: item.error || '',
    }));
}

export function aggregateSearchResults(
  results: SearchResult[]
): SearchAggregateGroup[] {
  const grouped = new Map<string, SearchAggregateGroup>();

  results.forEach((item) => {
    const key = buildAggregateKey(item);
    const existing = grouped.get(key);
    if (existing) {
      existing.items.push(item);
      return;
    }

    grouped.set(key, {
      key,
      title: item.title,
      year: item.year,
      type: item.type,
      cover: item.cover,
      rating: item.rating,
      sourceCount: 0,
      resultCount: 0,
      items: [item],
    });
  });

  return Array.from(grouped.values())
    .map((group) => ({
      ...group,
      sourceCount: new Set(
        group.items.map((item) => item.source).filter(Boolean)
      ).size,
      resultCount: group.items.length,
    }))
    .sort((a, b) => {
      if (b.sourceCount !== a.sourceCount) return b.sourceCount - a.sourceCount;
      if (b.items.length !== a.items.length) return b.items.length - a.items.length;
      return parseYearValue(b.year) - parseYearValue(a.year);
    });
}

export function sortSearchResults(results: SearchResult[]): SearchResult[] {
  const sourceCountByGroup = new Map<string, number>();
  aggregateSearchResults(results).forEach((group) => {
    sourceCountByGroup.set(group.key, group.sourceCount);
  });

  return [...results].sort((a, b) => {
    const aCount = sourceCountByGroup.get(buildAggregateKey(a)) || 0;
    const bCount = sourceCountByGroup.get(buildAggregateKey(b)) || 0;
    if (bCount !== aCount) return bCount - aCount;
    return parseYearValue(b.year) - parseYearValue(a.year);
  });
}

export async function fetchWithTimeout(
  input: RequestInfo | URL,
  init?: RequestInit,
  timeoutMs = DEFAULT_TIMEOUT_MS
): Promise<Response> {
  const controller = new AbortController();
  const timer = window.setTimeout(() => controller.abort(), timeoutMs);

  try {
    return await fetch(input, { ...(init || {}), signal: controller.signal });
  } finally {
    window.clearTimeout(timer);
  }
}
