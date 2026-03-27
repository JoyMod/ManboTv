'use client';

import { Loader2 } from 'lucide-react';
import dynamic from 'next/dynamic';
import { useRouter, useSearchParams } from 'next/navigation';
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import {
  deleteSkipConfig,
  getAllPlayRecords,
  getSkipConfig,
  PlayRecord,
  savePlayRecord,
  saveSkipConfig,
  SkipConfig,
} from '@/lib/db.client';

import TopNav from '@/components/layout/TopNav';
import {
  fetchPlayBootstrap,
  testSourceCandidate,
} from '@/components/play/play-data';
import {
  DetailResult,
  InitialSourceTestDelayMs,
  InitialSourceTestLimit,
  PlayBootstrapPayload,
  RelatedResult,
  SourceCandidate,
  SourceTestResult,
} from '@/components/play/play-utils';
import PlayInfoPanel from '@/components/play/PlayInfoPanel';

import type { PlayRelatedPanelProps } from '../../components/play/PlayRelatedPanel';
import type { PlayVideoPlayerProps } from '../../components/play/PlayVideoPlayer';

interface QualityOption {
  index: number;
  label: string;
}

const DefaultQualityIndex = -1;
const DefaultEpisodeIndex = 0;
const EmptySkipBoundarySeconds = 0;

const PlayVideoPlayer = dynamic<PlayVideoPlayerProps>(
  () =>
    import('../../components/play/PlayVideoPlayer.jsx').then(
      (module) => module.default
    ),
  {
    ssr: false,
    loading: () => (
      <div className='relative aspect-video w-full bg-black'>
        <div className='flex h-full items-center justify-center text-netflix-gray-400'>
          <Loader2 className='mr-2 h-5 w-5 animate-spin' />
          正在加载播放器...
        </div>
      </div>
    ),
  }
);

const PlayRelatedPanel = dynamic<PlayRelatedPanelProps>(
  () =>
    import('../../components/play/PlayRelatedPanel.jsx').then(
      (module) => module.default
    ),
  {
    ssr: false,
    loading: () => (
      <div className='rounded border border-zinc-800 bg-zinc-900/60 p-4 text-sm text-zinc-400'>
        正在加载相关推荐...
      </div>
    ),
  }
);

function getSourceKey(candidate: Pick<SourceCandidate, 'source' | 'id'>): string {
  return `${candidate.source}+${candidate.id}`;
}

function prioritizeSourceCandidates(
  sources: SourceCandidate[],
  currentSource: string,
  currentId: string
): SourceCandidate[] {
  return [...sources].sort((left, right) => {
    const leftIsCurrent =
      left.source === currentSource && left.id === currentId ? 1 : 0;
    const rightIsCurrent =
      right.source === currentSource && right.id === currentId ? 1 : 0;
    return rightIsCurrent - leftIsCurrent;
  });
}

export default function PlayPage() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const source = searchParams.get('source') || '';
  const id = searchParams.get('id') || '';
  const fallbackTitle = searchParams.get('title') || '未知影片';
  const directEpisode = searchParams.get('ep') || '';
  const sourceDisplayName = searchParams.get('sname') || source;
  const searchTitle = searchParams.get('stitle') || fallbackTitle;
  const searchYear = searchParams.get('year') || '';
  const searchType = searchParams.get('stype') || '';
  const preferLine = ['1', 'true', 'yes'].includes(
    (searchParams.get('prefer') || '').toLowerCase()
  );

  const [detail, setDetail] = useState<DetailResult | null>(null);
  const [relatedVideos, setRelatedVideos] = useState<RelatedResult[]>([]);
  const [relatedLoading, setRelatedLoading] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeEpisodeIndex, setActiveEpisodeIndex] = useState(0);
  const [isFavorite, setIsFavorite] = useState(false);
  const [favoriteSaving, setFavoriteSaving] = useState(false);
  const [playerError, setPlayerError] = useState<string | null>(null);
  const [retryToken, setRetryToken] = useState(0);
  const [qualityList, setQualityList] = useState<QualityOption[]>([]);
  const [activeQuality, setActiveQuality] = useState(DefaultQualityIndex);
  const [availableSources, setAvailableSources] = useState<SourceCandidate[]>([]);
  const [sourceTests, setSourceTests] = useState<Record<string, SourceTestResult>>({});
  const [sourcePanelLoading, setSourcePanelLoading] = useState(false);
  const [sourceBatchTesting, setSourceBatchTesting] = useState(false);
  const [resumeTime, setResumeTime] = useState<number | null>(null);
  const [skipConfig, setSkipConfig] = useState<SkipConfig>({
    enable: false,
    intro_time: EmptySkipBoundarySeconds,
    outro_time: EmptySkipBoundarySeconds,
  });
  const [blockAdEnabled, setBlockAdEnabled] = useState<boolean>(() => {
    if (typeof window === 'undefined') return true;
    const value = localStorage.getItem('enable_blockad');
    return value !== null ? value === 'true' : true;
  });
  const autoFallbackAttemptRef = useRef('');

  const episodes = detail?.episodes || [];
  const episodeTitles = detail?.episodes_titles || [];
  const currentEpisodeUrl =
    episodes[activeEpisodeIndex] ||
    (activeEpisodeIndex === DefaultEpisodeIndex ? directEpisode : '');
  const title = detail?.title || fallbackTitle;

  const streamUrl = useMemo(() => {
    if (!currentEpisodeUrl) return '';
    const separator = currentEpisodeUrl.includes('?') ? '&' : '?';
    return `/api/proxy/m3u8?url=${encodeURIComponent(
      currentEpisodeUrl
    )}${separator}blockAd=${blockAdEnabled}`;
  }, [blockAdEnabled, currentEpisodeUrl]);

  const buildPlayHref = useCallback(
    (nextSource: string, nextId: string, nextTitle: string) => {
      const params = new URLSearchParams({
        source: nextSource,
        id: nextId,
        title: nextTitle || fallbackTitle,
      });

      if (searchTitle) params.set('stitle', searchTitle);
      if (searchYear) params.set('year', searchYear);
      if (searchType) params.set('stype', searchType);

      return `/play?${params.toString()}`;
    },
    [fallbackTitle, searchTitle, searchType, searchYear]
  );

  const updateSourceTest = useCallback((key: string, value: SourceTestResult) => {
    setSourceTests((prev) => ({ ...prev, [key]: value }));
  }, []);

  const runSourceTest = useCallback(
    async (candidate: SourceCandidate) => {
      const key = getSourceKey(candidate);
      updateSourceTest(key, { status: 'testing' });
      const result = await testSourceCandidate(candidate);
      updateSourceTest(key, result);
      return result;
    },
    [updateSourceTest]
  );

  const handleSkipConfigChange = useCallback(
    async (nextConfig: SkipConfig) => {
      if (!source || !id) return;

      setSkipConfig(nextConfig);

      if (
        !nextConfig.enable &&
        nextConfig.intro_time === EmptySkipBoundarySeconds &&
        nextConfig.outro_time === EmptySkipBoundarySeconds
      ) {
        await deleteSkipConfig(source, id);
        return;
      }

      await saveSkipConfig(source, id, nextConfig);
    },
    [id, source]
  );

  const handleFavorite = useCallback(async () => {
    if (!source || !id || favoriteSaving) return;
    setFavoriteSaving(true);

    try {
      if (!isFavorite) {
        await fetch('/api/favorites', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            source,
            id,
            title,
            cover: detail?.poster || '',
            year: detail?.year || '',
          }),
        });
        setIsFavorite(true);
      } else {
        await fetch(`/api/favorites?key=${encodeURIComponent(`${source}+${id}`)}`, {
          method: 'DELETE',
        });
        setIsFavorite(false);
      }
    } finally {
      setFavoriteSaving(false);
    }
  }, [detail, favoriteSaving, id, isFavorite, source, title]);

  const handleRetryPlay = useCallback(() => {
    setPlayerError(null);
    setRetryToken((prev) => prev + 1);
  }, []);

  const handleQualityChange = useCallback(
    (event: React.ChangeEvent<HTMLSelectElement>) => {
      setActiveQuality(Number(event.target.value));
    },
    []
  );

  const handleSourceSwitch = useCallback(
    (candidate: SourceCandidate) => {
      if (!candidate.source || !candidate.id) return;
      if (candidate.source === source && candidate.id === id) return;

      router.replace(
        buildPlayHref(candidate.source, candidate.id, candidate.title || fallbackTitle)
      );
    },
    [buildPlayHref, fallbackTitle, id, router, source]
  );

  const handleSourceTest = useCallback(
    (candidate: SourceCandidate) => {
      void runSourceTest(candidate);
    },
    [runSourceTest]
  );

  const handleTestMoreSources = useCallback(() => {
    if (sourcePanelLoading || sourceBatchTesting || availableSources.length === 0) {
      return;
    }

    void (async () => {
      setSourceBatchTesting(true);
      try {
        const pendingSources = prioritizeSourceCandidates(
          availableSources,
          source,
          id
        ).filter((candidate) => {
          const key = getSourceKey(candidate);
          const test = sourceTests[key];
          const isCurrent = candidate.source === source && candidate.id === id;
          if (isCurrent) return test?.status !== 'ok';
          return test?.status !== 'ok' && test?.status !== 'testing';
        });

        for (const candidate of pendingSources) {
          await runSourceTest(candidate);
        }
      } finally {
        setSourceBatchTesting(false);
      }
    })();
  }, [
    availableSources,
    id,
    runSourceTest,
    source,
    sourceBatchTesting,
    sourcePanelLoading,
    sourceTests,
  ]);

  const handleEpisodeChange = useCallback(
    (index: number) => {
      if (index === activeEpisodeIndex) return;
      setActiveEpisodeIndex(index);
      setResumeTime(EmptySkipBoundarySeconds);
    },
    [activeEpisodeIndex]
  );

  const handlePreviousEpisode = useCallback(() => {
    if (activeEpisodeIndex <= 0) return;
    handleEpisodeChange(activeEpisodeIndex - 1);
  }, [activeEpisodeIndex, handleEpisodeChange]);

  const handleNextEpisode = useCallback(() => {
    if (activeEpisodeIndex >= episodes.length - 1) return;
    handleEpisodeChange(activeEpisodeIndex + 1);
  }, [activeEpisodeIndex, episodes.length, handleEpisodeChange]);

  useEffect(() => {
    autoFallbackAttemptRef.current = '';
  }, [id, source]);

  useEffect(() => {
    let cancelled = false;
    const loadDetail = async () => {
      setDetail(null);
      setAvailableSources([]);
      setSourceTests({});
      setRelatedVideos([]);
      setRelatedLoading(false);
      setIsFavorite(false);
      setSourcePanelLoading(true);
      setSourceBatchTesting(false);
      setPlayerError(null);
      setQualityList([]);
      setActiveQuality(DefaultQualityIndex);
      setResumeTime(null);

      setError(null);
      setRelatedLoading(true);
      setLoading(true);

      try {
        const data = (await fetchPlayBootstrap({
          source,
          id,
          fallbackTitle,
          searchTitle,
          searchYear,
          searchType,
          directEpisode,
          preferLine,
        })) as PlayBootstrapPayload;
        if (cancelled) return;

        if (data.redirect?.source && data.redirect?.id) {
          router.replace(
            buildPlayHref(
              data.redirect.source,
              data.redirect.id,
              data.redirect.title || fallbackTitle
            )
          );
          return;
        }

        const detailData = data.detail;
        if (!detailData) {
          throw new Error('missing bootstrap detail');
        }

        const resolvedEpisodes =
          Array.isArray(detailData.episodes) && detailData.episodes.length > 0
            ? detailData.episodes
            : directEpisode
            ? [directEpisode]
            : [];
        const resolvedEpisodeTitles =
          Array.isArray(detailData.episodes_titles) &&
            detailData.episodes_titles.length > 0
            ? detailData.episodes_titles
            : directEpisode
            ? ['1']
            : [];

        if (cancelled) return;
        setDetail({
          ...detailData,
          title: detailData.title || fallbackTitle,
          episodes: resolvedEpisodes,
          episodes_titles: resolvedEpisodeTitles,
        });
        setIsFavorite(Boolean(data.is_favorite));
        setAvailableSources(
          prioritizeSourceCandidates(
            (Array.isArray(data.available_sources) ? data.available_sources : []).map(
              (candidate) => ({
                id: String(candidate.id || ''),
                source: String(candidate.source || ''),
                source_name: String(candidate.source_name || ''),
                title: String(candidate.title || ''),
                year: String(candidate.year || ''),
                episodes: Array.isArray(candidate.episodes)
                  ? candidate.episodes
                  : [],
                episodes_titles: Array.isArray(candidate.episodes_titles)
                  ? candidate.episodes_titles
                  : [],
              })
            ),
            source,
            id
          )
        );
        setRelatedVideos(
          Array.isArray(data.related_videos) ? data.related_videos : []
        );
        setActiveEpisodeIndex(DefaultEpisodeIndex);
        setLoading(false);
        setSourcePanelLoading(false);
        setRelatedLoading(false);
      } catch (_error) {
        if (cancelled) return;
        setError(
          !source || !id || source === 'douban'
            ? '未找到可播放资源，请尝试切换关键词或资源站'
            : '加载播放信息失败，请稍后重试'
        );
        setLoading(false);
        setSourcePanelLoading(false);
        setRelatedLoading(false);
      }
    };

    void loadDetail();

    return () => {
      cancelled = true;
    };
  }, [
    buildPlayHref,
    directEpisode,
    fallbackTitle,
    id,
    preferLine,
    router,
    searchTitle,
    searchType,
    searchYear,
    source,
  ]);

  useEffect(() => {
    if (sourcePanelLoading || availableSources.length === 0 || !source || !id) {
      return;
    }

    const currentKey = getSourceKey({ source, id });
    const currentTest = sourceTests[currentKey];
    if (currentTest?.status !== 'error') {
      return;
    }
    if (autoFallbackAttemptRef.current === currentKey) {
      return;
    }

    const fallbackCandidate = prioritizeSourceCandidates(
      availableSources,
      source,
      id
    ).find((candidate) => {
      if (candidate.source === source && candidate.id === id) {
        return false;
      }

      return sourceTests[getSourceKey(candidate)]?.status === 'ok';
    });

    if (!fallbackCandidate) {
      return;
    }

    autoFallbackAttemptRef.current = currentKey;
    router.replace(
      buildPlayHref(
        fallbackCandidate.source,
        fallbackCandidate.id,
        fallbackCandidate.title || fallbackTitle
      )
    );
  }, [
    availableSources,
    buildPlayHref,
    fallbackTitle,
    id,
    router,
    source,
    sourcePanelLoading,
    sourceTests,
  ]);

  useEffect(() => {
    if (sourcePanelLoading || availableSources.length === 0) {
      return;
    }

    const initialSources = prioritizeSourceCandidates(
      availableSources,
      source,
      id
    ).slice(0, InitialSourceTestLimit);
    if (initialSources.length === 0) {
      return;
    }

    const timer = window.setTimeout(() => {
      initialSources.forEach((candidate) => {
        const key = getSourceKey(candidate);
        if (sourceTests[key]?.status === 'testing' || sourceTests[key]?.status === 'ok') {
          return;
        }
        void runSourceTest(candidate);
      });
    }, InitialSourceTestDelayMs);

    return () => window.clearTimeout(timer);
  }, [
    availableSources,
    id,
    runSourceTest,
    source,
    sourcePanelLoading,
    sourceTests,
  ]);

  useEffect(() => {
    const loadPlaybackState = async () => {
      if (!source || !id) return;

      try {
        const allRecords = await getAllPlayRecords();
        const record = allRecords[`${source}+${id}`];
        if (record) {
          if (record.index > 0 && record.index <= episodes.length) {
            setActiveEpisodeIndex(record.index - 1);
          }
          if (record.play_time > 0) {
            setResumeTime(record.play_time);
          }
        }
      } catch (_error) {
        return;
      }

      try {
        const config = await getSkipConfig(source, id);
        if (config) {
          setSkipConfig(config);
        }
      } catch (_error) {
        return;
      }
    };

    void loadPlaybackState();
  }, [episodes.length, id, source]);

  const handlePersistProgress = useCallback(
    (playTime: number, totalTime: number) => {
      if (!source || !id || playTime <= 0) return;

      const record: PlayRecord = {
        title,
        source_name: detail?.source_name || sourceDisplayName,
        cover: detail?.poster || '',
        year: detail?.year || searchYear || '',
        index: activeEpisodeIndex + 1,
        total_episodes: episodes.length,
        play_time: playTime,
        total_time: totalTime,
        save_time: Date.now(),
        search_title: title,
      };

      void savePlayRecord(source, id, record);
    },
    [
      activeEpisodeIndex,
      detail,
      episodes.length,
      id,
      searchYear,
      source,
      sourceDisplayName,
      title,
    ]
  );

  return (
    <main className='min-h-screen bg-netflix-black'>
      <TopNav />

      <div className='pt-16'>
        <PlayVideoPlayer
          loading={loading}
          streamUrl={streamUrl}
          poster={detail?.poster}
          retryToken={retryToken}
          blockAdEnabled={blockAdEnabled}
          onBlockAdChange={setBlockAdEnabled}
          skipConfig={skipConfig}
          initialResumeTime={resumeTime}
          desiredQuality={activeQuality}
          onPersistProgress={handlePersistProgress}
          onSkipConfigChange={handleSkipConfigChange}
          onPlayerErrorChange={setPlayerError}
          onQualityListChange={setQualityList}
          onActiveQualityChange={setActiveQuality}
        />

        <div className='mx-auto max-w-[1920px] px-4 py-8 sm:px-8'>
          {playerError ? (
            <div className='mb-6 flex flex-wrap items-center gap-3 rounded border border-yellow-500/30 bg-yellow-500/10 p-3 text-yellow-300'>
              <span className='text-sm'>{playerError}</span>
              <button
                onClick={handleRetryPlay}
                className='rounded bg-yellow-600 px-3 py-1 text-xs font-semibold text-white hover:bg-yellow-500'
              >
                重试播放
              </button>
            </div>
          ) : null}

          {error ? (
            <div className='mb-6 rounded border border-red-500/30 bg-red-500/10 p-3 text-red-400'>
              {error}
            </div>
          ) : null}

          <div className='grid grid-cols-1 gap-8 lg:grid-cols-3'>
            <PlayInfoPanel
              title={title}
              detail={detail}
              source={source}
              id={id}
              skipEnabled={skipConfig.enable}
              qualityList={qualityList}
              activeQuality={activeQuality}
              onQualityChange={handleQualityChange}
              onBack={() => router.back()}
              onFavorite={handleFavorite}
              favoriteSaving={favoriteSaving}
              isFavorite={isFavorite}
              sourcePanelLoading={sourcePanelLoading}
              sourceBatchTesting={sourceBatchTesting}
              availableSources={availableSources}
              sourceTests={sourceTests}
              onSourceSwitch={handleSourceSwitch}
              onSourceTest={handleSourceTest}
              onTestMoreSources={handleTestMoreSources}
              episodes={episodes}
              episodeTitles={episodeTitles}
              activeEpisodeIndex={activeEpisodeIndex}
              onEpisodeChange={handleEpisodeChange}
              onPreviousEpisode={handlePreviousEpisode}
              onNextEpisode={handleNextEpisode}
            />

            <div>
              {relatedLoading && relatedVideos.length === 0 ? (
                <div className='rounded border border-zinc-800 bg-zinc-900/60 p-4 text-sm text-zinc-400'>
                  正在加载相关推荐，不影响当前播放...
                </div>
              ) : null}
              <PlayRelatedPanel relatedVideos={relatedVideos} />
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}
