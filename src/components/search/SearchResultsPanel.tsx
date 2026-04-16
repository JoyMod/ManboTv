'use client';

import { motion } from 'framer-motion';
import { ChevronRight, Loader2, Power, Sparkles, Zap } from 'lucide-react';
import React, { useEffect, useMemo, useState } from 'react';

import { toProxyImageSrc } from '@/lib/image';
import { useFastNavigation } from '@/lib/navigation-feedback';

import {
  qualityScore,
  SearchAggregateGroup,
  SearchExecutionInfo,
  SearchFacets,
  SearchPageInfo,
  SearchResult,
  SearchSourceStatusItem,
  SourceTestResult,
} from '@/components/search/search-utils';
import SearchFilterPanel from '@/components/search/SearchFilterPanel';
import SearchSourceStatusPanel from '@/components/search/SearchSourceStatusPanel';
import ContentCard from '@/components/ui/ContentCard';
import SmartImage from '@/components/ui/SmartImage';

type SearchViewMode = 'aggregate' | 'lines' | 'sources';
type SearchSourceMode = 'all' | 'multi' | 'single';

interface SearchResultsPanelProps {
  query: string;
  loading: boolean;
  results: SearchResult[];
  aggregates: SearchAggregateGroup[];
  facets: SearchFacets;
  execution?: SearchExecutionInfo;
  pageInfo?: SearchPageInfo;
  searchError: string | null;
  sourceStatusItems: SearchSourceStatusItem[];
  sourceTests: Record<string, SourceTestResult>;
  viewMode: SearchViewMode;
  selectedTypes: string[];
  selectedSources: string[];
  selectedSort: string;
  selectedSourceMode: SearchSourceMode;
  selectedYearFrom?: number;
  selectedYearTo?: number;
  onViewModeChange: (mode: SearchViewMode) => void;
  onSortChange: (value: string) => void;
  onToggleType: (value: string) => void;
  onToggleSource: (value: string) => void;
  onSourceModeChange: (value: SearchSourceMode) => void;
  onYearRangeApply: (yearFrom?: number, yearTo?: number) => void;
  onResetFilters: () => void;
  onLoadMore: () => void;
}

const INITIAL_AGGREGATE_LIMIT = 18;
const INITIAL_SOURCE_LIMIT = 12;
const INITIAL_SOURCE_ITEM_LIMIT = 6;
const PAGE_STEP = 18;
const PREFETCH_RESULT_COUNT = 6;

interface SourceGroup {
  source: string;
  sourceName: string;
  items: SearchResult[];
}

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
  const bestItem = group.items[0];
  if (bestItem?.source && bestItem?.id) {
    return buildPlayUrl(bestItem, searchTitle);
  }

  return `/play?title=${encodeURIComponent(group.title)}&year=${encodeURIComponent(
    group.year || ''
  )}&stype=${encodeURIComponent(group.type)}&stitle=${encodeURIComponent(
    searchTitle || group.title
  )}&prefer=1`;
}

function buildAggregateLaunchUrl(
  group: SearchAggregateGroup,
  searchTitle: string,
  preferredItem?: SearchResult
): string {
  if (preferredItem?.source && preferredItem?.id) {
    return buildPlayUrl(preferredItem, searchTitle);
  }
  return buildPreferredPlayUrl(group, searchTitle);
}

function buildSourceGroups(results: SearchResult[]): SourceGroup[] {
  const grouped = new Map<string, SourceGroup>();

  results.forEach((item) => {
    const source = item.source || 'unknown';
    const existing = grouped.get(source);
    if (existing) {
      existing.items.push(item);
      return;
    }

    grouped.set(source, {
      source,
      sourceName: item.sourceName || source,
      items: [item],
    });
  });

  return Array.from(grouped.values()).sort((left, right) => {
    if (left.items.length !== right.items.length) {
      return right.items.length - left.items.length;
    }
    return left.sourceName.localeCompare(right.sourceName, 'zh-CN');
  });
}

function formatExecutionSummary(execution?: SearchExecutionInfo): string {
  if (!execution) return '服务端聚合搜索已完成。';
  const parts = [
    `${execution.completed_sources || 0}/${execution.total_sources || 0} 个源已响应`,
  ];
  if (typeof execution.elapsed_ms === 'number') {
    parts.push(`${execution.elapsed_ms}ms`);
  }
  if (execution.degraded) {
    parts.push('快速降级返回');
  }
  return parts.join(' · ');
}

export default function SearchResultsPanel({
  query,
  loading,
  results,
  aggregates,
  facets,
  execution,
  pageInfo,
  searchError,
  sourceStatusItems,
  sourceTests,
  viewMode,
  selectedTypes,
  selectedSources,
  selectedSort,
  selectedSourceMode,
  selectedYearFrom,
  selectedYearTo,
  onViewModeChange,
  onSortChange,
  onToggleType,
  onToggleSource,
  onSourceModeChange,
  onYearRangeApply,
  onResetFilters,
  onLoadMore,
}: SearchResultsPanelProps) {
  const { navigate, prefetchHref } = useFastNavigation();
  const [aggregateLimit, setAggregateLimit] = useState(INITIAL_AGGREGATE_LIMIT);
  const [sourceLimit, setSourceLimit] = useState(INITIAL_SOURCE_LIMIT);
  const [sourceItemLimits, setSourceItemLimits] = useState<Record<string, number>>(
    {}
  );

  useEffect(() => {
    setAggregateLimit(INITIAL_AGGREGATE_LIMIT);
    setSourceLimit(INITIAL_SOURCE_LIMIT);
    setSourceItemLimits({});
  }, [query, selectedSort, selectedTypes, selectedSources, selectedSourceMode, selectedYearFrom, selectedYearTo]);

  const sourceGroups = useMemo(() => buildSourceGroups(results), [results]);
  const displayedAggregates = aggregates.slice(0, aggregateLimit);
  const displayedSourceGroups = sourceGroups.slice(0, sourceLimit);
  const visibleTotal =
    viewMode === 'aggregate'
      ? aggregates.length
      : viewMode === 'sources'
      ? sourceGroups.length
      : pageInfo?.total || results.length;
  const hasMoreLines =
    Number(pageInfo?.total || 0) > results.length &&
    Number(pageInfo?.page_size || 0) > 0;

  useEffect(() => {
    const targets =
      viewMode === 'aggregate'
        ? displayedAggregates.slice(0, PREFETCH_RESULT_COUNT).map((item) =>
            buildPreferredPlayUrl(item, query)
          )
        : results.slice(0, PREFETCH_RESULT_COUNT).map((item) =>
            buildPlayUrl(item, query)
          );
    targets.forEach((href) => prefetchHref(href));
  }, [displayedAggregates, prefetchHref, query, results, viewMode]);

  if (!query) return null;

  return (
    <div className='space-y-6'>
      <section className='rounded-[32px] border border-netflix-gray-800 bg-[radial-gradient(circle_at_top_left,rgba(229,9,20,0.22),transparent_38%),linear-gradient(180deg,rgba(18,18,18,0.95),rgba(9,9,9,0.96))] p-5 md:p-7'>
        <div className='flex flex-wrap items-start justify-between gap-5'>
          <div className='space-y-2'>
            <p className='text-xs uppercase tracking-[0.35em] text-netflix-gray-500'>
              Search Workbench
            </p>
            <h1 className='text-2xl font-black text-white md:text-3xl'>
              {query}
            </h1>
            <p className='text-sm text-netflix-gray-400'>
              {formatExecutionSummary(execution)}
            </p>
          </div>

          <div className='rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-right'>
            <p className='text-xs uppercase tracking-[0.3em] text-netflix-gray-500'>
              当前命中
            </p>
            <p className='mt-1 text-3xl font-black text-white'>{visibleTotal}</p>
            <p className='mt-1 text-xs text-netflix-gray-400'>
              {viewMode === 'aggregate'
                ? '聚合结果'
                : viewMode === 'sources'
                ? '资源站分组'
                : '线路结果'}
            </p>
          </div>
        </div>
      </section>

      <SearchFilterPanel
        facets={facets}
        execution={execution}
        selectedView={viewMode}
        selectedSort={selectedSort}
        selectedTypes={selectedTypes}
        selectedSources={selectedSources}
        selectedSourceMode={selectedSourceMode}
        selectedYearFrom={selectedYearFrom}
        selectedYearTo={selectedYearTo}
        onViewChange={onViewModeChange}
        onSortChange={onSortChange}
        onToggleType={onToggleType}
        onToggleSource={onToggleSource}
        onSourceModeChange={onSourceModeChange}
        onYearRangeApply={onYearRangeApply}
        onResetFilters={onResetFilters}
      />

      <SearchSourceStatusPanel
        execution={execution}
        statuses={sourceStatusItems}
      />

      {searchError && (
        <div className='rounded-2xl border border-yellow-500/20 bg-yellow-500/10 px-4 py-3 text-sm text-yellow-200'>
          {searchError}
        </div>
      )}

      {loading && results.length > 0 && (
        <div className='rounded-2xl border border-netflix-gray-800 bg-netflix-surface/60 px-4 py-3 text-sm text-netflix-gray-400'>
          正在更新筛选结果，当前内容可先浏览。
        </div>
      )}

      {loading && results.length === 0 ? (
        <div className='flex items-center justify-center py-16'>
          <Loader2 className='h-10 w-10 animate-spin text-netflix-red' />
        </div>
      ) : visibleTotal === 0 ? (
        <div className='rounded-3xl border border-netflix-gray-800 bg-netflix-surface/50 py-16 text-center'>
          <p className='text-lg text-netflix-gray-300'>没有找到匹配结果</p>
          <p className='mt-2 text-sm text-netflix-gray-500'>
            可以尝试更换关键词、放宽年份范围，或者切换来源模式。
          </p>
        </div>
      ) : viewMode === 'aggregate' ? (
        <>
          <div className='grid grid-cols-2 gap-4 md:grid-cols-4 xl:grid-cols-6'>
            {displayedAggregates.map((group, index) => {
              const testedItems = group.items
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
                .sort((left, right) => (right?.score || 0) - (left?.score || 0));
              const bestCandidate = testedItems[0]?.item || group.items[0];

              return (
                <motion.article
                  key={group.key}
                  initial={{ opacity: 0, y: 16 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: Math.min(index * 0.03, 0.35) }}
                  className='rounded-3xl border border-netflix-gray-800 bg-netflix-surface/60 p-3'
                >
                  <div className='relative overflow-hidden rounded-2xl'>
                    <SmartImage
                      src={toProxyImageSrc(group.cover)}
                      alt={group.title}
                      fill
                      sizes='(max-width: 768px) 50vw, 16vw'
                      className='aspect-[2/3] w-full object-cover'
                    />
                    <button
                      type='button'
                      onClick={() =>
                        navigate(buildAggregateLaunchUrl(group, query, bestCandidate))
                      }
                      className='absolute right-2 top-2 inline-flex items-center gap-1 rounded-full bg-netflix-red px-2.5 py-1 text-[11px] text-white shadow-lg'
                    >
                      <Zap className='h-3 w-3' />
                      优选
                    </button>
                  </div>

                  <div className='mt-3 space-y-2'>
                    <p className='line-clamp-2 text-sm font-semibold text-white'>
                      {group.title}
                    </p>
                    <p className='text-[11px] text-netflix-gray-400'>
                      {group.year || '年份未知'} · {group.sourceCount} 个源 ·{' '}
                      {group.resultCount} 条线路
                    </p>

                    {group.tags && group.tags.length > 0 && (
                      <div className='flex flex-wrap gap-1'>
                        {group.tags.slice(0, 3).map((tag) => (
                          <span
                            key={tag}
                            className='rounded-full border border-netflix-gray-700 px-2 py-0.5 text-[10px] text-netflix-gray-300'
                          >
                            {tag}
                          </span>
                        ))}
                      </div>
                    )}

                    {bestCandidate?.matchReasons && bestCandidate.matchReasons.length > 0 && (
                      <p className='text-[11px] text-netflix-gray-500'>
                        <Sparkles className='mr-1 inline h-3 w-3' />
                        {bestCandidate.matchReasons.slice(0, 2).join(' · ')}
                      </p>
                    )}

                    <button
                      type='button'
                      onClick={() =>
                        navigate(buildAggregateLaunchUrl(group, query, bestCandidate))
                      }
                      className='inline-flex items-center gap-2 rounded-full bg-white/10 px-3 py-1.5 text-xs text-white transition-colors hover:bg-netflix-red'
                    >
                      <Power className='h-3.5 w-3.5' />
                      立即查看
                    </button>
                  </div>
                </motion.article>
              );
            })}
          </div>

          {aggregates.length > aggregateLimit && (
            <div className='flex justify-center'>
              <button
                type='button'
                onClick={() => setAggregateLimit((value) => value + PAGE_STEP)}
                className='rounded-full border border-netflix-gray-700 px-5 py-2 text-sm text-netflix-gray-300 transition-colors hover:border-netflix-gray-500 hover:text-white'
              >
                加载更多聚合结果
              </button>
            </div>
          )}
        </>
      ) : viewMode === 'sources' ? (
        <>
          <div className='space-y-4'>
            {displayedSourceGroups.map((group) => (
              (() => {
                const visibleItems =
                  sourceItemLimits[group.source] || INITIAL_SOURCE_ITEM_LIMIT;
                return (
                  <section
                    key={group.source}
                    className='rounded-3xl border border-netflix-gray-800 bg-netflix-surface/50 p-4'
                  >
                    <div className='mb-4 flex items-center justify-between gap-3'>
                      <div>
                        <p className='text-lg font-semibold text-white'>
                          {group.sourceName}
                        </p>
                        <p className='text-sm text-netflix-gray-500'>
                          当前已加载 {group.items.length} 条线路
                        </p>
                      </div>
                      <span className='rounded-full border border-netflix-gray-700 px-3 py-1 text-xs text-netflix-gray-300'>
                        {group.source}
                      </span>
                    </div>

                    <div className='grid grid-cols-2 gap-4 md:grid-cols-4 xl:grid-cols-6'>
                      {group.items.slice(0, visibleItems).map((item) => (
                        <div key={`${item.source}-${item.id}`}>
                          <ContentCard
                            id={item.id}
                            title={item.title}
                            cover={item.cover}
                            firstEpisode={item.episodes[0]}
                            rating={item.rating}
                            year={item.year}
                            type={item.type}
                            source={item.source}
                            sourceName={item.sourceName}
                            searchTitle={query || item.title}
                          />
                        </div>
                      ))}
                    </div>

                    {group.items.length > visibleItems && (
                      <div className='mt-4 flex justify-center'>
                        <button
                          type='button'
                          onClick={() =>
                            setSourceItemLimits((previous) => ({
                              ...previous,
                              [group.source]: visibleItems + INITIAL_SOURCE_ITEM_LIMIT,
                            }))
                          }
                          className='rounded-full border border-netflix-gray-700 px-4 py-2 text-xs text-netflix-gray-300 transition-colors hover:border-netflix-gray-500 hover:text-white'
                        >
                          展开该源更多线路
                        </button>
                      </div>
                    )}
                  </section>
                );
              })()
            ))}
          </div>

          {sourceGroups.length > sourceLimit && (
            <div className='flex justify-center'>
              <button
                type='button'
                onClick={() => setSourceLimit((value) => value + 6)}
                className='rounded-full border border-netflix-gray-700 px-5 py-2 text-sm text-netflix-gray-300 transition-colors hover:border-netflix-gray-500 hover:text-white'
              >
                展开更多资源站
              </button>
            </div>
          )}

          {hasMoreLines && (
            <div className='flex justify-center'>
              <button
                type='button'
                onClick={onLoadMore}
                className='rounded-full border border-netflix-gray-700 px-5 py-2 text-sm text-netflix-gray-300 transition-colors hover:border-netflix-gray-500 hover:text-white'
              >
                继续加载更多资源站结果
              </button>
            </div>
          )}
        </>
      ) : (
        <>
          <div className='grid grid-cols-2 gap-4 md:grid-cols-4 xl:grid-cols-6'>
            {results.map((item, index) => (
              <motion.div
                key={`${item.source || 'unknown'}-${item.id}-${index}`}
                initial={{ opacity: 0, y: 16 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: Math.min(index * 0.02, 0.3) }}
                className='space-y-2'
              >
                <div className='relative'>
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
                    type='button'
                    onClick={() => navigate(buildPlayUrl(item, query))}
                    className='absolute right-2 top-2 z-20 inline-flex items-center gap-1 rounded-full bg-black/70 px-2.5 py-1 text-[11px] text-white transition-colors hover:bg-netflix-red'
                  >
                    <ChevronRight className='h-3 w-3' />
                    播放
                  </button>
                </div>

                <div className='rounded-2xl border border-netflix-gray-800 bg-black/20 px-3 py-2'>
                  <p className='text-[11px] text-netflix-gray-400'>
                    {item.sourceName || item.source || '未知来源'} · 分数{' '}
                    {Math.round(item.matchScore || 0)}
                  </p>
                  {item.matchReasons && item.matchReasons.length > 0 && (
                    <p className='mt-1 line-clamp-2 text-[11px] text-netflix-gray-500'>
                      {item.matchReasons.join(' · ')}
                    </p>
                  )}
                </div>
              </motion.div>
            ))}
          </div>

          {hasMoreLines && (
            <div className='flex justify-center'>
              <button
                type='button'
                onClick={onLoadMore}
                className='rounded-full border border-netflix-gray-700 px-5 py-2 text-sm text-netflix-gray-300 transition-colors hover:border-netflix-gray-500 hover:text-white'
              >
                加载更多线路结果
              </button>
            </div>
          )}
        </>
      )}

      {viewMode === 'lines' && (
        <div className='rounded-2xl border border-netflix-gray-800 bg-netflix-surface/40 px-4 py-3 text-xs text-netflix-gray-500'>
          当前线路结果按服务端智能分排序，综合考虑标题命中、年份、类型和可播放性。
        </div>
      )}
    </div>
  );
}
