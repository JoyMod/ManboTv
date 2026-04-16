'use client';

import { AnimatePresence, motion } from 'framer-motion';
import { Clock, Loader2, Search, X } from 'lucide-react';
import { useSearchParams } from 'next/navigation';
import React, { useEffect, useRef, useState } from 'react';

import { normalizeImageUrl } from '@/lib/image';
import { useFastNavigation } from '@/lib/navigation-feedback';

import TopNav from '@/components/layout/TopNav';
import {
  fetchWithTimeout,
  HOT_SEARCHES,
  normalizeAggregateGroups,
  normalizeItems,
  normalizeSearchSourceStatuses,
  parseQuality,
  SearchAggregateGroup,
  SearchBootstrapPayload,
  SearchExecutionInfo,
  SearchFacets,
  SearchPageInfo,
  SearchResult,
  SearchSourceStatusItem,
  SourceTestResult,
  SuggestionItem,
} from '@/components/search/search-utils';
import SearchResultsPanel from '@/components/search/SearchResultsPanel';

const DEFAULT_PAGE_SIZE = 60;
const SUGGESTION_DELAY_MS = 180;
const SOURCE_TEST_LIMIT = 24;

type SearchViewMode = 'aggregate' | 'lines' | 'sources';
type SearchSourceMode = 'all' | 'multi' | 'single';

interface SearchSelections {
  sort?: string;
  view?: SearchViewMode;
  sourceMode?: SearchSourceMode;
  yearFrom?: number;
  yearTo?: number;
}

function parseListParam(value: string | null): string[] {
  if (!value) return [];
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);
}

function parseOptionalInt(value: string | null): number | undefined {
  if (!value) return undefined;
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function normalizeView(value?: string): SearchViewMode {
  switch (value) {
    case 'lines':
    case 'sources':
      return value;
    default:
      return 'aggregate';
  }
}

function normalizeSourceMode(value?: string): SearchSourceMode {
  switch (value) {
    case 'multi':
    case 'single':
      return value;
    default:
      return 'all';
  }
}

function unwrapApiData<T>(payload: unknown): T {
  if (
    payload &&
    typeof payload === 'object' &&
    'code' in payload &&
    'data' in payload
  ) {
    return (payload as { data: T }).data;
  }
  return payload as T;
}

function toggleSelection(current: string[], value: string): string[] {
  return current.includes(value)
    ? current.filter((item) => item !== value)
    : [...current, value];
}

export default function SearchPage() {
  const { navigate, prefetchHref } = useFastNavigation();
  const searchParams = useSearchParams();
  const searchParamsKey = searchParams.toString();
  const searchQuery = searchParams.get('q') || '';
  const selectedTypes = parseListParam(searchParams.get('types'));
  const selectedSources = parseListParam(searchParams.get('sources'));

  const [query, setQuery] = useState(searchQuery);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [aggregates, setAggregates] = useState<SearchAggregateGroup[]>([]);
  const [facets, setFacets] = useState<SearchFacets>({});
  const [pageInfo, setPageInfo] = useState<SearchPageInfo | undefined>();
  const [execution, setExecution] = useState<SearchExecutionInfo | undefined>();
  const [loading, setLoading] = useState(false);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [history, setHistory] = useState<string[]>([]);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [sourceStatusItems, setSourceStatusItems] = useState<
    SearchSourceStatusItem[]
  >([]);
  const [sourceTests, setSourceTests] = useState<
    Record<string, SourceTestResult>
  >({});
  const [searchError, setSearchError] = useState<string | null>(null);
  const [serverSelections, setServerSelections] = useState<SearchSelections>({});
  const inputRef = useRef<HTMLInputElement>(null);
  const testedRef = useRef<Set<string>>(new Set());

  const currentView = normalizeView(
    searchParams.get('view') || serverSelections.view
  );
  const currentSort = searchParams.get('sort') || serverSelections.sort || 'smart';
  const currentSourceMode = normalizeSourceMode(
    searchParams.get('source_mode') || serverSelections.sourceMode
  );
  const currentYearFrom =
    parseOptionalInt(searchParams.get('year_from')) || serverSelections.yearFrom;
  const currentYearTo =
    parseOptionalInt(searchParams.get('year_to')) || serverSelections.yearTo;
  const currentPageSize =
    parseOptionalInt(searchParams.get('page_size')) || DEFAULT_PAGE_SIZE;

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  useEffect(() => {
    setQuery(searchQuery);
  }, [searchQuery]);

  useEffect(() => {
    if (!searchQuery) {
      prefetchHref('/play');
    }
  }, [prefetchHref, searchQuery]);

  const buildNavigationParams = (
    patch: Record<string, string | number | string[] | undefined | null>
  ) => {
    const nextParams = new URLSearchParams(searchParams.toString());
    Object.entries(patch).forEach(([key, value]) => {
      if (
        value === undefined ||
        value === null ||
        value === '' ||
        (Array.isArray(value) && value.length === 0)
      ) {
        nextParams.delete(key);
        return;
      }

      if (Array.isArray(value)) {
        nextParams.set(key, value.join(','));
        return;
      }

      nextParams.set(key, String(value));
    });
    return nextParams;
  };

  const navigateWithPatch = (
    patch: Record<string, string | number | string[] | undefined | null>
  ) => {
    const nextParams = buildNavigationParams(patch);
    const queryString = nextParams.toString();
    navigate(queryString ? `/search?${queryString}` : '/search');
  };

  useEffect(() => {
    const controller = new AbortController();

    const loadBootstrap = async () => {
      testedRef.current.clear();
      setSourceTests({});
      setSearchError(null);

      const params = new URLSearchParams(searchParams.toString());
      if (searchQuery && !params.get('page_size')) {
        params.set('page_size', String(DEFAULT_PAGE_SIZE));
      }

      const hasQuery = Boolean(searchQuery.trim());
      setLoading(hasQuery);
      if (hasQuery) {
        setShowSuggestions(false);
      }

      try {
        const queryString = params.toString();
        const response = await fetch(
          `/api/v1/search/bootstrap${queryString ? `?${queryString}` : ''}`,
          {
            signal: controller.signal,
            credentials: 'include',
          }
        );
        if (!response.ok) {
          throw new Error(`search bootstrap request failed: ${response.status}`);
        }

        const rawPayload = await response.json();
        const payload = unwrapApiData<SearchBootstrapPayload>(rawPayload);

        setHistory(
          Array.isArray(payload.history)
            ? payload.history.filter((item) => typeof item === 'string')
            : []
        );
        setSuggestions(
          Array.isArray(payload.suggestions)
            ? payload.suggestions.filter((item) => typeof item === 'string')
            : []
        );
        setExecution(payload.execution);
        setFacets(payload.facets || {});
        setPageInfo(payload.page_info);
        setSourceStatusItems(
          normalizeSearchSourceStatuses(payload.source_status_items)
        );
        setServerSelections({
          sort: payload.selected_sort || 'smart',
          view: normalizeView(payload.selected_view),
          sourceMode: normalizeSourceMode(payload.selected_source_mode),
          yearFrom: payload.selected_year_from,
          yearTo: payload.selected_year_to,
        });

        if (!hasQuery) {
          setResults([]);
          setAggregates([]);
          return;
        }

        setResults(
          normalizeItems(
            Array.isArray(payload.results) ? payload.results : [],
            normalizeImageUrl
          )
        );
        setAggregates(
          normalizeAggregateGroups(
            Array.isArray(payload.aggregates) ? payload.aggregates : [],
            normalizeImageUrl
          )
        );
      } catch (error) {
        if ((error as Error).name === 'AbortError') return;
        if (searchQuery.trim()) {
          setSearchError('搜索失败，请稍后重试');
        }
      } finally {
        setLoading(false);
      }
    };

    void loadBootstrap();

    return () => controller.abort();
  }, [searchParams, searchParamsKey, searchQuery]);

  useEffect(() => {
    if (!query.trim()) {
      setSuggestions([]);
      return;
    }

    const timer = window.setTimeout(async () => {
      try {
        const response = await fetch(
          `/api/v1/search/suggestions?q=${encodeURIComponent(query.trim())}`,
          {
            credentials: 'include',
          }
        );
        if (!response.ok) throw new Error('suggestions request failed');

        const rawData = await response.json();
        const data = unwrapApiData<SuggestionItem[] | string[]>(rawData);
        const list = (Array.isArray(data) ? data : [])
          .map((item) =>
            typeof item === 'string' ? item : item?.text ? item.text : ''
          )
          .filter(Boolean)
          .slice(0, 10);

        setSuggestions(list);
      } catch {
        const lowerQuery = query.toLowerCase();
        const fallback = [...history, ...HOT_SEARCHES]
          .filter((item, index, list) => {
            return (
              list.indexOf(item) === index &&
              item.toLowerCase().includes(lowerQuery)
            );
          })
          .slice(0, 10);
        setSuggestions(fallback);
      }
    }, SUGGESTION_DELAY_MS);

    return () => window.clearTimeout(timer);
  }, [history, query]);

  const clearHistory = async () => {
    try {
      await fetch('/api/v1/searchhistory', {
        method: 'DELETE',
        credentials: 'include',
      });
      setHistory([]);
    } catch {
      setHistory([]);
    }
  };

  const removeFromHistory = async (item: string, event: React.MouseEvent) => {
    event.stopPropagation();
    try {
      await fetch(`/api/v1/searchhistory?keyword=${encodeURIComponent(item)}`, {
        method: 'DELETE',
        credentials: 'include',
      });
      setHistory((previous) => previous.filter((current) => current !== item));
    } catch {
      return;
    }
  };

  const runSourceTest = async (item: SearchResult) => {
    if (!item.source || !item.id || item.episodes.length === 0) return;
    const key = `${item.source}+${item.id}`;

    setSourceTests((previous) => ({
      ...previous,
      [key]: { ...(previous[key] || {}), status: 'testing' },
    }));

    try {
      const startedAt = performance.now();
      const response = await fetchWithTimeout(
        `/api/proxy/m3u8?url=${encodeURIComponent(item.episodes[0])}`
      );
      if (!response.ok) throw new Error(`http ${response.status}`);
      const text = await response.text();
      if (!/#EXTM3U/i.test(text)) throw new Error('invalid m3u8 payload');

      const elapsed = Math.max(1, performance.now() - startedAt);
      const bytes = new Blob([text]).size;
      const kbps = (bytes / 1024 / (elapsed / 1000)).toFixed(0);

      setSourceTests((previous) => ({
        ...previous,
        [key]: {
          status: 'ok',
          pingMs: Math.round(elapsed),
          quality: parseQuality(text),
          speed: `${kbps} KB/s`,
        },
      }));
    } catch {
      setSourceTests((previous) => ({
        ...previous,
        [key]: { status: 'error', quality: '未知', speed: '--' },
      }));
    }
  };

  useEffect(() => {
    if (loading || results.length === 0) return;
    results.slice(0, SOURCE_TEST_LIMIT).forEach((item) => {
      const key = `${item.source || ''}+${item.id || ''}`;
      if (!item.source || !item.id || testedRef.current.has(key)) return;
      testedRef.current.add(key);
      void runSourceTest(item);
    });
  }, [loading, results]);

  const submitSearch = (searchValue: string) => {
    const trimmed = searchValue.trim();
    if (!trimmed) {
      navigate('/search');
      return;
    }

    navigateWithPatch({
      q: trimmed,
      page: null,
      page_size: currentPageSize || DEFAULT_PAGE_SIZE,
    });
  };

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    submitSearch(query);
  };

  return (
    <main className='min-h-screen bg-netflix-black'>
      <TopNav />

      <div className='px-4 pb-8 pt-24 sm:px-8'>
        <div className='mx-auto max-w-5xl'>
          <form onSubmit={handleSubmit} className='relative'>
            <div className='relative overflow-hidden rounded-[28px] border border-netflix-gray-800 bg-[radial-gradient(circle_at_top_left,rgba(229,9,20,0.18),transparent_35%),linear-gradient(180deg,rgba(19,19,19,0.95),rgba(9,9,9,0.96))] p-3 shadow-[0_30px_80px_rgba(0,0,0,0.35)]'>
              <Search className='absolute left-8 top-1/2 h-6 w-6 -translate-y-1/2 text-netflix-gray-500' />
              <input
                ref={inputRef}
                type='text'
                value={query}
                onChange={(event) => {
                  setQuery(event.target.value);
                  setShowSuggestions(event.target.value.length > 0);
                }}
                onFocus={() => setShowSuggestions(query.length > 0)}
                placeholder='输入片名、年份、类型或资源偏好，例如：沙丘 2024 电影 多源'
                className='h-14 w-full rounded-2xl border border-white/10 bg-black/20 pl-14 pr-12 text-lg text-white placeholder-netflix-gray-500 outline-none transition-colors focus:border-netflix-red'
              />
              {query && (
                <button
                  type='button'
                  onClick={() => {
                    setQuery('');
                    setShowSuggestions(false);
                    navigate('/search');
                    inputRef.current?.focus();
                  }}
                  className='absolute right-8 top-1/2 -translate-y-1/2 text-netflix-gray-500 transition-colors hover:text-white'
                >
                  <X className='h-5 w-5' />
                </button>
              )}
            </div>

            <AnimatePresence>
              {showSuggestions && !loading && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  className='absolute left-0 right-0 top-full z-50 mt-3 overflow-hidden rounded-3xl border border-netflix-gray-800 bg-netflix-surface shadow-netflix-hover'
                >
                  <div className='border-b border-netflix-gray-800 p-4'>
                    <p className='mb-3 flex items-center gap-2 text-xs text-netflix-gray-500'>
                      <Search className='h-3.5 w-3.5' />
                      联想建议
                    </p>
                    {suggestions.length > 0 ? (
                      <div className='space-y-2'>
                        {suggestions.map((item) => (
                          <button
                            key={item}
                            type='button'
                            onClick={() => submitSearch(item)}
                            className='w-full rounded-2xl px-3 py-2 text-left text-netflix-gray-300 transition-colors hover:bg-netflix-gray-800/60 hover:text-white'
                          >
                            {item}
                          </button>
                        ))}
                      </div>
                    ) : (
                      <p className='text-sm text-netflix-gray-500'>暂无建议</p>
                    )}
                  </div>

                  <div className='p-4'>
                    <div className='mb-3 flex items-center justify-between'>
                      <p className='flex items-center gap-2 text-xs text-netflix-gray-500'>
                        <Clock className='h-3.5 w-3.5' />
                        搜索历史
                      </p>
                      {history.length > 0 && (
                        <button
                          type='button'
                          onClick={clearHistory}
                          className='text-xs text-netflix-gray-500 transition-colors hover:text-netflix-red'
                        >
                          清空
                        </button>
                      )}
                    </div>

                    {history.length > 0 ? (
                      <div className='flex flex-wrap gap-2'>
                        {history.map((item) => (
                          <button
                            key={item}
                            type='button'
                            onClick={() => submitSearch(item)}
                            className='group relative rounded-full border border-netflix-gray-700 px-4 py-2 text-sm text-netflix-gray-300 transition-colors hover:border-netflix-gray-500 hover:text-white'
                          >
                            {item}
                            <span
                              onClick={(event) => removeFromHistory(item, event)}
                              className='absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-netflix-gray-600 text-[10px] text-white opacity-0 transition-opacity group-hover:opacity-100'
                            >
                              ×
                            </span>
                          </button>
                        ))}
                      </div>
                    ) : (
                      <p className='text-sm text-netflix-gray-500'>暂无搜索历史</p>
                    )}
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </form>
        </div>
      </div>

      <div className='px-4 pb-20 sm:px-8'>
        <div className='mx-auto max-w-[1920px]'>
          {searchQuery ? (
            <SearchResultsPanel
              query={searchQuery}
              loading={loading}
              results={results}
              aggregates={aggregates}
              facets={facets}
              execution={execution}
              pageInfo={pageInfo}
              searchError={searchError}
              sourceStatusItems={sourceStatusItems}
              sourceTests={sourceTests}
              viewMode={currentView}
              selectedTypes={selectedTypes}
              selectedSources={selectedSources}
              selectedSort={currentSort}
              selectedSourceMode={currentSourceMode}
              selectedYearFrom={currentYearFrom}
              selectedYearTo={currentYearTo}
              onViewModeChange={(value) =>
                navigateWithPatch({ view: value, page: null })
              }
              onSortChange={(value) =>
                navigateWithPatch({ sort: value, page: null })
              }
              onToggleType={(value) =>
                navigateWithPatch({
                  types: toggleSelection(selectedTypes, value),
                  page: null,
                })
              }
              onToggleSource={(value) =>
                navigateWithPatch({
                  sources: toggleSelection(selectedSources, value),
                  page: null,
                })
              }
              onSourceModeChange={(value) =>
                navigateWithPatch({ source_mode: value, page: null })
              }
              onYearRangeApply={(yearFrom, yearTo) =>
                navigateWithPatch({
                  year_from: yearFrom,
                  year_to: yearTo,
                  page: null,
                })
              }
              onResetFilters={() =>
                navigateWithPatch({
                  view: 'aggregate',
                  sort: undefined,
                  types: [],
                  sources: [],
                  source_mode: undefined,
                  year_from: undefined,
                  year_to: undefined,
                  page: undefined,
                })
              }
              onLoadMore={() =>
                navigateWithPatch({
                  page_size: currentPageSize + DEFAULT_PAGE_SIZE,
                  page: undefined,
                })
              }
            />
          ) : (
            <div className='mx-auto max-w-4xl space-y-8'>
              {history.length > 0 && (
                <section>
                  <div className='mb-4 flex items-center justify-between'>
                    <h2 className='text-lg font-bold text-white'>最近搜索</h2>
                    <button
                      type='button'
                      onClick={clearHistory}
                      className='text-sm text-netflix-gray-500 transition-colors hover:text-netflix-red'
                    >
                      清空
                    </button>
                  </div>
                  <div className='flex flex-wrap gap-3'>
                    {history.map((item) => (
                      <button
                        key={item}
                        type='button'
                        onClick={() => submitSearch(item)}
                        className='rounded-full border border-netflix-gray-700 px-4 py-2 text-sm text-netflix-gray-300 transition-colors hover:border-netflix-gray-500 hover:text-white'
                      >
                        {item}
                      </button>
                    ))}
                  </div>
                </section>
              )}

              <section>
                <h2 className='mb-4 text-lg font-bold text-white'>热门搜索</h2>
                <div className='grid grid-cols-2 gap-3 sm:grid-cols-4'>
                  {HOT_SEARCHES.map((item) => (
                    <button
                      key={item}
                      type='button'
                      onClick={() => submitSearch(item)}
                      className='rounded-3xl border border-netflix-gray-800 bg-netflix-surface/60 px-4 py-5 text-left text-sm text-netflix-gray-200 transition-colors hover:border-netflix-gray-600 hover:text-white'
                    >
                      {item}
                    </button>
                  ))}
                </div>
              </section>

              {loading && (
                <div className='flex items-center justify-center py-16'>
                  <Loader2 className='h-10 w-10 animate-spin text-netflix-red' />
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </main>
  );
}
