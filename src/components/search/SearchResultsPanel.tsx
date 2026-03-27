'use client';

import { motion } from 'framer-motion';
import { Filter, Loader2, Power, Zap } from 'lucide-react';
import React, { useEffect, useMemo, useState } from 'react';

import { toProxyImageSrc } from '@/lib/image';
import { useFastNavigation } from '@/lib/navigation-feedback';

import {
  aggregateSearchResults,
  buildAggregateKey,
  qualityScore,
  SearchAggregateGroup,
  SearchResult,
  sortSearchResults,
  SourceTestResult,
} from '@/components/search/search-utils';
import ContentCard from '@/components/ui/ContentCard';
import SmartImage from '@/components/ui/SmartImage';

type ResultViewMode = 'agg' | 'all';
type ContentFilter = 'all' | 'movie' | 'tv' | 'variety' | 'anime';
type SourceFilter = 'all' | 'multi' | 'single';

interface SearchResultsPanelProps {
  query: string;
  loading: boolean;
  results: SearchResult[];
  searchError: string | null;
  sourceStatus: Record<string, 'done' | 'error'>;
  sourceTests: Record<string, SourceTestResult>;
  viewMode: ResultViewMode;
  onViewModeChange: (mode: ResultViewMode) => void;
}

const INITIAL_AGG_PAGE_SIZE = 18;
const INITIAL_ALL_PAGE_SIZE = 24;
const PAGE_SIZE_STEP = 18;
const PREFETCH_RESULT_COUNT = 6;

const contentFilterItems: Array<{ label: string; value: ContentFilter }> = [
  { label: '全部', value: 'all' },
  { label: '电影', value: 'movie' },
  { label: '剧集', value: 'tv' },
  { label: '综艺', value: 'variety' },
  { label: '动漫', value: 'anime' },
];

const sourceFilterItems: Array<{ label: string; value: SourceFilter }> = [
  { label: '全部来源', value: 'all' },
  { label: '多源优先', value: 'multi' },
  { label: '单源补充', value: 'single' },
];

function buildPlayUrl(item: SearchResult, searchTitle: string): string {
  return `/play?source=${encodeURIComponent(item.source || '')}&id=${encodeURIComponent(
    item.id
  )}&title=${encodeURIComponent(item.title)}&ep=${encodeURIComponent(
    item.episodes[0] || ''
  )}&sname=${encodeURIComponent(item.sourceName || item.source || '')}&year=${encodeURIComponent(
    item.year || ''
  )}&stype=${encodeURIComponent(item.type)}&stitle=${encodeURIComponent(
    searchTitle || item.title
  )}`;
}

function buildPreferredPlayUrl(
  group: SearchAggregateGroup,
  searchTitle: string
): string {
  return `/play?title=${encodeURIComponent(group.title)}&year=${encodeURIComponent(
    group.year || ''
  )}&stype=${encodeURIComponent(group.type)}&stitle=${encodeURIComponent(
    searchTitle || group.title
  )}&prefer=1`;
}

function matchSourceFilter(sourceCount: number, sourceFilter: SourceFilter): boolean {
  if (sourceFilter === 'multi') return sourceCount >= 2;
  if (sourceFilter === 'single') return sourceCount < 2;
  return true;
}

export default function SearchResultsPanel({
  query,
  loading,
  results,
  searchError,
  sourceStatus,
  sourceTests,
  viewMode,
  onViewModeChange,
}: SearchResultsPanelProps) {
  const { navigate, prefetchHref } = useFastNavigation();
  const [contentFilter, setContentFilter] = useState<ContentFilter>('all');
  const [sourceFilter, setSourceFilter] = useState<SourceFilter>('all');
  const [hiddenAggregateKeys, setHiddenAggregateKeys] = useState<string[]>([]);
  const [aggregateLimit, setAggregateLimit] = useState(INITIAL_AGG_PAGE_SIZE);
  const [allLimit, setAllLimit] = useState(INITIAL_ALL_PAGE_SIZE);

  useEffect(() => {
    setHiddenAggregateKeys([]);
    setAggregateLimit(INITIAL_AGG_PAGE_SIZE);
    setAllLimit(INITIAL_ALL_PAGE_SIZE);
    setContentFilter('all');
    setSourceFilter('all');
  }, [query]);

  const aggregatedResults = useMemo(
    () => aggregateSearchResults(results),
    [results]
  );

  const groupSourceCountMap = useMemo(() => {
    const map = new Map<string, number>();
    aggregatedResults.forEach((group) => {
      map.set(group.key, group.sourceCount);
    });
    return map;
  }, [aggregatedResults]);

  const sortedResults = useMemo(() => sortSearchResults(results), [results]);

  const filteredAggregatedResults = useMemo(
    () =>
      aggregatedResults.filter((group) => {
        const contentMatches =
          contentFilter === 'all' || group.type === contentFilter;
        return contentMatches && matchSourceFilter(group.sourceCount, sourceFilter);
      }),
    [aggregatedResults, contentFilter, sourceFilter]
  );

  const visibleAggregatedResults = useMemo(
    () =>
      filteredAggregatedResults.filter(
        (group) => !!group.cover && !hiddenAggregateKeys.includes(group.key)
      ),
    [filteredAggregatedResults, hiddenAggregateKeys]
  );

  const filteredAllResults = useMemo(
    () =>
      sortedResults.filter((item) => {
        const contentMatches =
          contentFilter === 'all' || item.type === contentFilter;
        const sourceCount = groupSourceCountMap.get(buildAggregateKey(item));
        return (
          contentMatches &&
          matchSourceFilter(sourceCount || 0, sourceFilter)
        );
      }),
    [contentFilter, groupSourceCountMap, sortedResults, sourceFilter]
  );

  const displayedAggregatedResults = visibleAggregatedResults.slice(0, aggregateLimit);
  const displayedAllResults = filteredAllResults.slice(0, allLimit);
  const totalSources = Object.keys(sourceStatus).length;
  const failedSources = Object.values(sourceStatus).filter(
    (value) => value === 'error'
  ).length;
  const showBlockingLoading = loading && results.length === 0;

  useEffect(() => {
    if (viewMode === 'agg') {
      displayedAggregatedResults.slice(0, PREFETCH_RESULT_COUNT).forEach((group) => {
        prefetchHref(buildPreferredPlayUrl(group, query));
      });
      return;
    }

    displayedAllResults.slice(0, PREFETCH_RESULT_COUNT).forEach((item) => {
      prefetchHref(buildPlayUrl(item, query));
    });
  }, [displayedAggregatedResults, displayedAllResults, prefetchHref, query, viewMode]);

  if (!query) return null;

  return (
    <>
      <div className='mb-6 flex flex-wrap items-start justify-between gap-4'>
        <div>
          <h1 className='text-xl font-bold text-white'>
            "{query}" 的搜索结果
            {!loading && results.length > 0 && (
              <span className='ml-2 text-base font-normal text-netflix-gray-500'>
                ({viewMode === 'agg' ? visibleAggregatedResults.length : filteredAllResults.length})
              </span>
            )}
          </h1>
          <p className='mt-2 text-sm text-netflix-gray-500'>
            先看聚合结果，再按线路展开，能明显减少误点和等待。
          </p>
        </div>

        <div className='flex flex-wrap items-center gap-2'>
          {results.length > 0 && (
            <div className='rounded-full border border-netflix-gray-700 p-1'>
              <button
                onClick={() => onViewModeChange('agg')}
                className={`rounded-full px-3 py-1 text-xs ${
                  viewMode === 'agg'
                    ? 'bg-netflix-red text-white'
                    : 'text-netflix-gray-300'
                }`}
              >
                聚合
              </button>
              <button
                onClick={() => onViewModeChange('all')}
                className={`rounded-full px-3 py-1 text-xs ${
                  viewMode === 'all'
                    ? 'bg-netflix-red text-white'
                    : 'text-netflix-gray-300'
                }`}
              >
                全部
              </button>
            </div>
          )}

          {totalSources > 0 && (
            <span className='text-xs text-netflix-gray-400'>
              已返回 {totalSources} 个资源站
              {failedSources > 0 ? `（失败 ${failedSources}）` : ''}
            </span>
          )}
        </div>
      </div>

      {results.length > 0 && (
        <div className='mb-6 rounded-2xl border border-netflix-gray-800 bg-netflix-surface/60 p-4'>
          <div className='mb-3 flex items-center gap-2 text-sm text-netflix-gray-400'>
            <Filter className='h-4 w-4' />
            结果筛选
          </div>

          <div className='flex flex-wrap gap-2'>
            {contentFilterItems.map((item) => (
              <button
                key={item.value}
                onClick={() => setContentFilter(item.value)}
                className={`rounded-full px-3 py-1.5 text-xs transition-colors ${
                  contentFilter === item.value
                    ? 'bg-white text-black'
                    : 'border border-netflix-gray-700 text-netflix-gray-300 hover:border-netflix-gray-500 hover:text-white'
                }`}
              >
                {item.label}
              </button>
            ))}
          </div>

          <div className='mt-3 flex flex-wrap gap-2'>
            {sourceFilterItems.map((item) => (
              <button
                key={item.value}
                onClick={() => setSourceFilter(item.value)}
                className={`rounded-full px-3 py-1.5 text-xs transition-colors ${
                  sourceFilter === item.value
                    ? 'bg-netflix-red text-white'
                    : 'border border-netflix-gray-700 text-netflix-gray-300 hover:border-netflix-gray-500 hover:text-white'
                }`}
              >
                {item.label}
              </button>
            ))}
          </div>
        </div>
      )}

      {searchError && (
        <div className='mb-6 rounded-lg border border-yellow-500/20 bg-yellow-500/10 px-4 py-3 text-sm text-yellow-300'>
          {searchError}
        </div>
      )}

      {loading && results.length > 0 && (
        <div className='mb-4 rounded-lg border border-netflix-gray-800 bg-netflix-surface/40 px-4 py-3 text-sm text-netflix-gray-400'>
          正在继续聚合其它资源站，当前结果已可先行浏览。
        </div>
      )}

      {showBlockingLoading ? (
        <div className='flex items-center justify-center py-20'>
          <Loader2 className='h-12 w-12 animate-spin text-netflix-red' />
        </div>
      ) : results.length === 0 ? (
        <div className='py-20 text-center'>
          <p className='text-lg text-netflix-gray-500'>未找到相关结果</p>
          <p className='mt-2 text-sm text-netflix-gray-600'>尝试更换关键词搜索</p>
        </div>
      ) : viewMode === 'agg' ? (
        <>
          <div className='grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 sm:gap-6'>
            {displayedAggregatedResults.map((group, index) => {
              const first = group.items[0];
              const sourceNames = Array.from(
                new Set(
                  group.items
                    .map((item) => item.sourceName || item.source || '')
                    .filter(Boolean)
                )
              );
              const bestTested = group.items
                .map((item) => {
                  const key = `${item.source || ''}+${item.id || ''}`;
                  const test = sourceTests[key];
                  if (!test || test.status !== 'ok') return null;
                  return {
                    item,
                    score:
                      qualityScore(test.quality || '未知') * 0.7 +
                      (typeof test.pingMs === 'number'
                        ? Math.max(0, 300 - test.pingMs) * 0.3
                        : 0),
                    test,
                  };
                })
                .filter(Boolean)
                .sort((a, b) => (b?.score || 0) - (a?.score || 0))[0];
              const launchItem = bestTested?.item || first;

              return (
                <motion.div
                  key={group.key}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: Math.min(index * 0.03, 0.4) }}
                  className='rounded-lg border border-netflix-gray-800 bg-netflix-surface/70 p-2'
                >
                  <div className='relative overflow-hidden rounded'>
                    <SmartImage
                      src={toProxyImageSrc(group.cover)}
                      alt={group.title}
                      fill
                      sizes='(max-width: 768px) 50vw, 16vw'
                      className='aspect-[2/3] w-full object-cover'
                      onError={() => {
                        setHiddenAggregateKeys((prev) =>
                          prev.includes(group.key) ? prev : [...prev, group.key]
                        );
                      }}
                    />
                    <button
                      onClick={() => navigate(buildPreferredPlayUrl(group, query))}
                      onPointerEnter={() =>
                        prefetchHref(buildPreferredPlayUrl(group, query))
                      }
                      className='absolute right-2 top-2 inline-flex items-center gap-1 rounded bg-netflix-red px-2 py-1 text-[11px] text-white'
                    >
                      <Zap className='h-3 w-3' />
                      优选
                    </button>
                  </div>

                  <p className='mt-2 truncate text-sm font-semibold text-white'>
                    {group.title}
                  </p>
                  <p className='mt-1 text-[11px] text-netflix-gray-400'>
                    共 {group.items.length} 条线路 · {group.sourceCount} 个源
                  </p>
                  <p className='mt-1 text-[11px] text-netflix-gray-500'>
                    {bestTested
                      ? `优选参考: ${bestTested.test.quality || '未知'} · ${
                          typeof bestTested.test.pingMs === 'number'
                            ? `${bestTested.test.pingMs}ms`
                            : '--'
                        }`
                      : '优选参考: 待测速'}
                  </p>
                  <div className='mt-2 flex flex-wrap gap-1'>
                    {sourceNames.slice(0, 3).map((name) => (
                      <span
                        key={name}
                        className='rounded bg-netflix-gray-800 px-2 py-0.5 text-[10px] text-netflix-gray-300'
                      >
                        {name}
                      </span>
                    ))}
                  </div>

                  {launchItem?.source && launchItem?.id ? (
                    <button
                      onClick={() => navigate(buildPlayUrl(launchItem, query))}
                      onPointerEnter={() =>
                        prefetchHref(buildPlayUrl(launchItem, query))
                      }
                      className='mt-2 inline-flex items-center gap-1 rounded bg-black/70 px-2 py-1 text-[11px] text-white hover:bg-netflix-red'
                    >
                      <Power className='h-3 w-3' />
                      查看线路
                    </button>
                  ) : null}
                </motion.div>
              );
            })}
          </div>

          {visibleAggregatedResults.length > aggregateLimit && (
            <div className='mt-8 flex justify-center'>
              <button
                onClick={() => setAggregateLimit((prev) => prev + PAGE_SIZE_STEP)}
                className='rounded-full border border-netflix-gray-700 px-5 py-2 text-sm text-netflix-gray-300 transition-colors hover:border-netflix-gray-500 hover:text-white'
              >
                加载更多聚合结果
              </button>
            </div>
          )}
        </>
      ) : (
        <>
          <div className='grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 sm:gap-6'>
            {displayedAllResults.map((item, index) => {
              const key = `${item.source || 'unknown'}+${item.id}`;
              const test = sourceTests[key];

              return (
                <motion.div
                  key={`${item.source || 'unknown'}-${item.id}-${index}`}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: Math.min(index * 0.03, 0.4) }}
                  className='relative'
                >
                  <ContentCard
                    id={item.id}
                    title={item.title}
                    cover={item.cover}
                    firstEpisode={item.episodes[0]}
                    rating={item.rating}
                    searchTitle={query || item.title}
                    year={item.year}
                    type={item.type}
                    source={item.source}
                    sourceName={item.sourceName}
                  />
                  <button
                    onClick={(event) => {
                      event.stopPropagation();
                      navigate(buildPlayUrl(item, query));
                    }}
                    onPointerEnter={() => prefetchHref(buildPlayUrl(item, query))}
                    className='absolute right-2 top-2 z-20 inline-flex items-center gap-1 rounded bg-black/70 px-2 py-1 text-[11px] text-white hover:bg-netflix-red'
                    title='线路选择与测速'
                  >
                    <Power className='h-3 w-3' />
                    线路
                  </button>

                  <div className='mt-1 space-y-1'>
                    <p className='truncate text-[11px] text-netflix-gray-400'>
                      {item.sourceName || item.source || '未知来源'}
                    </p>
                    <p className='text-[11px] text-netflix-gray-500'>
                      清晰度:{' '}
                      {test?.quality || (test?.status === 'testing' ? '检测中' : '未知')} · 延迟:{' '}
                      {typeof test?.pingMs === 'number' ? `${test.pingMs}ms` : '--'}
                    </p>
                  </div>
                </motion.div>
              );
            })}
          </div>

          {filteredAllResults.length > allLimit && (
            <div className='mt-8 flex justify-center'>
              <button
                onClick={() => setAllLimit((prev) => prev + PAGE_SIZE_STEP)}
                className='rounded-full border border-netflix-gray-700 px-5 py-2 text-sm text-netflix-gray-300 transition-colors hover:border-netflix-gray-500 hover:text-white'
              >
                加载更多线路结果
              </button>
            </div>
          )}
        </>
      )}
    </>
  );
}
