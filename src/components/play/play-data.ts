'use client';

import {
  fetchWithTimeout,
  isValidM3U8,
  parseQuality,
  PlayBootstrapPayload,
  qualityScore,
  SourceCandidate,
  SourceTestResult,
} from '@/components/play/play-utils';

interface PlayBootstrapRequest {
  directEpisode?: string;
  fallbackTitle: string;
  id: string;
  preferLine: boolean;
  searchTitle: string;
  searchType: string;
  searchYear: string;
  source: string;
}

export async function fetchPlayBootstrap(
  request: PlayBootstrapRequest
): Promise<PlayBootstrapPayload> {
  const params = new URLSearchParams();

  if (request.source) params.set('source', request.source);
  if (request.id) params.set('id', request.id);
  if (request.fallbackTitle) params.set('title', request.fallbackTitle);
  if (request.searchTitle) params.set('stitle', request.searchTitle);
  if (request.searchYear) params.set('year', request.searchYear);
  if (request.searchType) params.set('stype', request.searchType);
  if (request.directEpisode) params.set('ep', request.directEpisode);
  if (request.preferLine) params.set('prefer', '1');

  const response = await fetch(`/api/play/bootstrap?${params.toString()}`);
  if (!response.ok) {
    throw new Error(`bootstrap request failed: ${response.status}`);
  }

  return (await response.json()) as PlayBootstrapPayload;
}

export async function testSourceCandidate(
  candidate: SourceCandidate
): Promise<SourceTestResult> {
  const firstUrl = candidate.episodes[0] || '';
  if (!firstUrl) {
    return { status: 'error', quality: '未知', speed: '--' };
  }

  try {
    const start = performance.now();
    const response = await fetchWithTimeout(
      `/api/proxy/m3u8?url=${encodeURIComponent(firstUrl)}`
    );
    if (!response.ok) throw new Error(`http ${response.status}`);

    const text = await response.text();
    if (!isValidM3U8(text)) throw new Error('invalid m3u8 payload');

    const elapsed = Math.max(1, performance.now() - start);
    const bytes = new Blob([text]).size;
    const kbps = (bytes / 1024 / (elapsed / 1000)).toFixed(0);

    return {
      status: 'ok',
      pingMs: Math.round(elapsed),
      quality: parseQuality(text),
      speed: `${kbps} KB/s`,
    };
  } catch {
    return { status: 'error', quality: '未知', speed: '--' };
  }
}

export function scoreTestedSource(result: SourceTestResult): number {
  const speedScore = result.speed
    ? Math.min(100, Number.parseFloat(result.speed) / 15)
    : 0;
  const latencyScore = result.pingMs ? 1000 / result.pingMs : 0;

  return (
    qualityScore(result.quality || '未知') * 0.5 +
    speedScore * 0.3 +
    latencyScore * 0.2
  );
}
