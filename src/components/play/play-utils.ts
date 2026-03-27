'use client';

import Hls, {
  type HlsConfig,
  type LoaderCallbacks,
  type LoaderConfiguration,
  type LoaderContext,
} from 'hls.js';

export interface DetailResult {
  id?: string;
  title?: string;
  poster?: string;
  episodes?: string[];
  episodes_titles?: string[];
  source?: string;
  source_name?: string;
  class?: string;
  year?: string;
  desc?: string;
}

export interface RelatedResult {
  id?: string;
  title?: string;
  poster?: string;
  year?: string;
  source?: string;
  class?: string;
  type_name?: string;
}

export interface SearchPlayableItem {
  id?: string;
  source?: string;
  title?: string;
}

export interface SourceCandidate {
  id: string;
  source: string;
  source_name?: string;
  title: string;
  year?: string;
  episodes: string[];
  episodes_titles?: string[];
}

export interface SourceTestResult {
  pingMs?: number;
  quality?: string;
  speed?: string;
  status: 'idle' | 'testing' | 'ok' | 'error';
}

export interface PlayBootstrapRedirect {
  id: string;
  source: string;
  title?: string;
}

export interface PlayBootstrapPayload {
  detail?: DetailResult | null;
  redirect?: PlayBootstrapRedirect | null;
  is_favorite?: boolean;
  available_sources?: SourceCandidate[];
  related_videos?: RelatedResult[];
}

export interface WakeLockSentinel {
  released: boolean;
  release(): Promise<void>;
  addEventListener(type: 'release', listener: () => void): void;
  removeEventListener(type: 'release', listener: () => void): void;
}

export const SourcePanelLoadDelayMs = 1800;
export const RelatedVideosLoadDelayMs = 3200;
export const InitialSourceTestLimit = 3;
export const InitialSourceTestDelayMs = 1200;
export const PlayRecordSaveIntervalMs = 15000;
export const SkipCheckIntervalMs = 1000;
export const DefaultFetchTimeoutMs = 8000;

export function mapType(text: string): 'movie' | 'tv' | 'variety' | 'anime' {
  const lower = text.toLowerCase();
  if (lower.includes('动漫') || lower.includes('anime')) return 'anime';
  if (lower.includes('综艺') || lower.includes('variety')) return 'variety';
  if (lower.includes('剧') || lower.includes('tv')) return 'tv';
  return 'movie';
}

export function normalizeTitle(title: string): string {
  return (title || '')
    .toLowerCase()
    .replace(/[\s\-_.:：!！?？,，。·'""''']/g, '')
    .trim();
}

export function parseQuality(manifest: string): string {
  const resolutionMatch = manifest.match(/RESOLUTION=(\d+)x(\d+)/i);
  if (resolutionMatch?.[2]) {
    return `${resolutionMatch[2]}p`;
  }

  const upper = manifest.toUpperCase();
  if (upper.includes('4K') || upper.includes('2160')) return '4K';
  if (upper.includes('1080')) return '1080p';
  if (upper.includes('720')) return '720p';
  if (upper.includes('480')) return '480p';
  return '未知';
}

export function qualityScore(quality: string): number {
  if (quality.includes('4K')) return 100;
  if (quality.includes('2160')) return 100;
  if (quality.includes('1080')) return 90;
  if (quality.includes('720')) return 75;
  if (quality.includes('480')) return 55;
  return 40;
}

export function isValidM3U8(content: string): boolean {
  return /#EXTM3U/i.test(content || '');
}

export async function fetchWithTimeout(
  input: RequestInfo | URL,
  timeoutMs = DefaultFetchTimeoutMs,
  init?: RequestInit
): Promise<Response> {
  const controller = new AbortController();
  const timer = window.setTimeout(() => controller.abort(), timeoutMs);
  try {
    return await fetch(input, { ...(init || {}), signal: controller.signal });
  } finally {
    window.clearTimeout(timer);
  }
}

export function formatTime(seconds: number): string {
  if (seconds === 0) return '00:00';
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const remainingSeconds = Math.round(seconds % 60);

  if (hours === 0) {
    return `${minutes.toString().padStart(2, '0')}:${remainingSeconds
      .toString()
      .padStart(2, '0')}`;
  }

  return `${hours.toString().padStart(2, '0')}:${minutes
    .toString()
    .padStart(2, '0')}:${remainingSeconds.toString().padStart(2, '0')}`;
}

export function filterAdsFromM3U8(m3u8Content: string): string {
  if (!m3u8Content) return '';
  return m3u8Content
    .split('\n')
    .filter((line) => !line.includes('#EXT-X-DISCONTINUITY'))
    .join('\n');
}

export class CustomHlsJsLoader extends (Hls.DefaultConfig.loader as typeof Hls.DefaultConfig.loader) {
  constructor(config: HlsConfig) {
    super(config);
    const originalLoad = this.load.bind(this);

    this.load = (
      context: LoaderContext,
      loaderConfig: LoaderConfiguration,
      callbacks: LoaderCallbacks<LoaderContext>
    ) => {
      const loaderType = (context as LoaderContext & { type?: string }).type;
      if (loaderType === 'manifest' || loaderType === 'level') {
        const originalSuccess = callbacks.onSuccess;
        callbacks.onSuccess = (...args) => {
          const [response] = args as Parameters<
            NonNullable<LoaderCallbacks<LoaderContext>['onSuccess']>
          >;
          if (response.data && typeof response.data === 'string') {
            response.data = filterAdsFromM3U8(response.data);
          }
          if (!originalSuccess) {
            return;
          }
          return originalSuccess(...args);
        };
      }

      return originalLoad(context, loaderConfig, callbacks);
    };
  }
}
