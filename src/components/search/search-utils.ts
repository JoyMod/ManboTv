'use client';

export interface SearchResult {
  id: string;
  title: string;
  cover: string;
  year: string;
  rating?: string;
  type: 'movie' | 'tv' | 'variety' | 'anime';
  source?: string;
  sourceName?: string;
  episodes: string[];
  episodesTitles: string[];
  tags?: string[];
  isAdult?: boolean;
}

export interface RawSearchItem {
  id?: string;
  title?: string;
  poster?: string;
  cover?: string;
  year?: string;
  rate?: string;
  source?: string;
  source_name?: string;
  class?: string;
  type_name?: string;
  episodes?: string[];
  episodes_titles?: string[];
  tags?: string[];
  is_adult?: boolean;
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
  results?: RawSearchItem[];
  history?: string[];
  suggestions?: string[];
  source_status?: Record<string, 'done' | 'error'>;
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
  items: SearchResult[];
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
      rating: item.rate || '',
      type: mapContentType(item),
      source: item.source,
      sourceName: item.source_name || '',
      episodes: Array.isArray(item.episodes) ? item.episodes : [],
      episodesTitles: Array.isArray(item.episodes_titles)
        ? item.episodes_titles
        : [],
      tags: Array.isArray(item.tags) ? item.tags : [],
      isAdult: Boolean(item.is_adult),
    });
  });

  return list;
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
      items: [item],
    });
  });

  return Array.from(grouped.values())
    .map((group) => ({
      ...group,
      sourceCount: new Set(
        group.items.map((item) => item.source).filter(Boolean)
      ).size,
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
