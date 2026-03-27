'use client';

import { AnimatePresence, motion } from 'framer-motion';
import { Clock, Loader2, Search, X } from 'lucide-react';
import { useSearchParams } from 'next/navigation';
import React, { useCallback, useEffect, useRef, useState } from 'react';

import { normalizeImageUrl } from '@/lib/image';
import { useFastNavigation } from '@/lib/navigation-feedback';

import TopNav from '@/components/layout/TopNav';
import {
  fetchWithTimeout,
  HOT_SEARCHES,
  isValidM3U8,
  normalizeItems,
  parseQuality,
  SearchBootstrapPayload,
  SearchResult,
  SourceTestResult,
  SuggestionItem,
} from '@/components/search/search-utils';
import SearchResultsPanel from '@/components/search/SearchResultsPanel';

const PREFETCH_RESULT_THRESHOLD = 1;
const SUGGESTION_DELAY_MS = 180;
const SOURCE_TEST_LIMIT = 24;

export default function SearchPage() {
  const { navigate, prefetchHref } = useFastNavigation();
  const searchParams = useSearchParams();
  const [query, setQuery] = useState(searchParams.get('q') || '');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [history, setHistory] = useState<string[]>([]);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [sourceStatus, setSourceStatus] = useState<
    Record<string, 'done' | 'error'>
  >({});
  const [sourceTests, setSourceTests] = useState<
    Record<string, SourceTestResult>
  >({});
  const [viewMode, setViewMode] = useState<'agg' | 'all'>('agg');
  const [searchError, setSearchError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const testedRef = useRef<Set<string>>(new Set());

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  useEffect(() => {
    if (results.length < PREFETCH_RESULT_THRESHOLD) return;
    prefetchHref('/play');
  }, [prefetchHref, results.length]);

  const loadBootstrap = useCallback(async (searchQuery: string) => {
    const trimmedQuery = searchQuery.trim();
    testedRef.current.clear();
    setSearchError(null);
    setSourceTests({});
    setSourceStatus({});

    if (!trimmedQuery) {
      setLoading(false);
      setResults([]);
    } else {
      setLoading(true);
      setShowSuggestions(false);
      setResults([]);
    }

    try {
      const params = trimmedQuery
        ? `?q=${encodeURIComponent(trimmedQuery)}`
        : '';
      const response = await fetch(`/api/search/bootstrap${params}`);
      if (!response.ok) {
        throw new Error('search bootstrap request failed');
      }

      const payload = (await response.json()) as SearchBootstrapPayload;
      const historyItems = Array.isArray(payload.history)
        ? payload.history.filter((item) => typeof item === 'string')
        : [];
      setHistory(historyItems);

      if (!trimmedQuery) {
        setSuggestions([]);
        setResults([]);
        return;
      }

      const rawItems = Array.isArray(payload.results) ? payload.results : [];
      setResults(normalizeItems(rawItems, normalizeImageUrl));
      setSuggestions(
        Array.isArray(payload.suggestions)
          ? payload.suggestions.filter((item) => typeof item === 'string')
          : []
      );
      setSourceStatus(payload.source_status || {});
    } catch {
      if (trimmedQuery) {
        setSearchError('搜索失败，请稍后重试');
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const nextQuery = searchParams.get('q') || '';
    setQuery(nextQuery);
    void loadBootstrap(nextQuery);
  }, [loadBootstrap, searchParams]);

  useEffect(() => {
    if (!query.trim()) {
      setSuggestions([]);
      return;
    }

    const timer = window.setTimeout(async () => {
      try {
        const response = await fetch(
          `/api/search/suggestions?q=${encodeURIComponent(query.trim())}`
        );
        if (!response.ok) throw new Error('suggestions request failed');
        const data = await response.json();
        const suggestionItems = Array.isArray(data?.suggestions)
          ? data.suggestions
          : Array.isArray(data?.data)
          ? data.data
          : [];
        const list = suggestionItems
          .map((item: SuggestionItem | string) =>
            typeof item === 'string' ? item : item?.text || ''
          )
          .filter(Boolean)
          .slice(0, 8);
        setSuggestions(list);
      } catch {
        const lowerQuery = query.toLowerCase();
        const fallback = [...history, ...HOT_SEARCHES]
          .filter((item, index, list) => {
            return list.indexOf(item) === index && item.toLowerCase().includes(lowerQuery);
          })
          .slice(0, 8);
        setSuggestions(fallback);
      }
    }, SUGGESTION_DELAY_MS);

    return () => window.clearTimeout(timer);
  }, [history, query]);

  const clearHistory = async () => {
    try {
      await fetch('/api/searchhistory', { method: 'DELETE' });
      setHistory([]);
    } catch (_error) {
      setHistory([]);
    }
  };

  const removeFromHistory = async (item: string, event: React.MouseEvent) => {
    event.stopPropagation();
    try {
      await fetch(`/api/searchhistory?keyword=${encodeURIComponent(item)}`, {
        method: 'DELETE',
      });
      setHistory((prev) => prev.filter((current) => current !== item));
    } catch (_error) {
      return;
    }
  };

  const runSourceTest = useCallback(async (item: SearchResult) => {
    if (!item.source || !item.id || item.episodes.length === 0) return;

    const key = `${item.source}+${item.id}`;
    setSourceTests((prev) => ({
      ...prev,
      [key]: { ...(prev[key] || {}), status: 'testing' },
    }));

    try {
      const start = performance.now();
      const response = await fetchWithTimeout(
        `/api/proxy/m3u8?url=${encodeURIComponent(item.episodes[0])}`
      );
      if (!response.ok) throw new Error(`http ${response.status}`);
      const text = await response.text();
      if (!isValidM3U8(text)) throw new Error('invalid m3u8 payload');

      const elapsed = Math.max(1, performance.now() - start);
      const bytes = new Blob([text]).size;
      const kbps = (bytes / 1024 / (elapsed / 1000)).toFixed(0);

      setSourceTests((prev) => ({
        ...prev,
        [key]: {
          status: 'ok',
          pingMs: Math.round(elapsed),
          quality: parseQuality(text),
          speed: `${kbps} KB/s`,
        },
      }));
    } catch {
      setSourceTests((prev) => ({
        ...prev,
        [key]: { status: 'error', quality: '未知', speed: '--' },
      }));
    }
  }, []);

  useEffect(() => {
    if (loading || results.length === 0) return;

    results.slice(0, SOURCE_TEST_LIMIT).forEach((item) => {
      const key = `${item.source || ''}+${item.id || ''}`;
      if (!item.source || !item.id || testedRef.current.has(key)) return;
      testedRef.current.add(key);
      void runSourceTest(item);
    });
  }, [loading, results, runSourceTest]);

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    if (!query.trim()) return;
    navigate(`/search?q=${encodeURIComponent(query.trim())}`);
  };

  const handleSuggestionClick = (value: string) => {
    setQuery(value);
    navigate(`/search?q=${encodeURIComponent(value)}`);
  };

  return (
    <main className='min-h-screen bg-netflix-black'>
      <TopNav />

      <div className='px-4 pb-8 pt-24 sm:px-8'>
        <div className='mx-auto max-w-3xl'>
          <form onSubmit={handleSubmit} className='relative'>
            <div className='relative'>
              <Search className='absolute left-4 top-1/2 h-6 w-6 -translate-y-1/2 text-netflix-gray-500' />
              <input
                ref={inputRef}
                type='text'
                value={query}
                onChange={(event) => {
                  setQuery(event.target.value);
                  setShowSuggestions(event.target.value.length > 0);
                }}
                onFocus={() => setShowSuggestions(query.length > 0)}
                placeholder='搜索电影、电视剧、综艺、动漫...'
                className='h-14 w-full rounded-xl border border-netflix-gray-800 bg-netflix-surface pl-14 pr-12 text-lg text-white placeholder-netflix-gray-500 transition-all focus:border-netflix-red focus:outline-none focus:ring-1 focus:ring-netflix-red/50'
              />
              {query && (
                <button
                  type='button'
                  onClick={() => {
                    setQuery('');
                    setResults([]);
                    setSourceStatus({});
                    setSourceTests({});
                    setShowSuggestions(false);
                    inputRef.current?.focus();
                    navigate('/search');
                  }}
                  className='absolute right-4 top-1/2 -translate-y-1/2 text-netflix-gray-500 hover:text-white'
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
                  className='absolute left-0 right-0 top-full z-50 mt-2 overflow-hidden rounded-xl border border-netflix-gray-800 bg-netflix-surface shadow-netflix-hover'
                >
                  <div className='border-b border-netflix-gray-800 p-4'>
                    <p className='mb-3 flex items-center gap-1 text-xs text-netflix-gray-500'>
                      <Search className='h-3 w-3' />
                      实时建议
                    </p>
                    {suggestions.length > 0 ? (
                      <div className='space-y-2'>
                        {suggestions.map((item) => (
                          <button
                            key={item}
                            type='button'
                            onClick={() => handleSuggestionClick(item)}
                            className='w-full rounded-lg px-3 py-2 text-left text-netflix-gray-300 transition-colors hover:bg-netflix-gray-800/60'
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
                      <p className='flex items-center gap-1 text-xs text-netflix-gray-500'>
                        <Clock className='h-3 w-3' />
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
                            onClick={() => handleSuggestionClick(item)}
                            className='group relative rounded-full bg-netflix-gray-800 px-4 py-2 text-sm text-netflix-gray-300 transition-colors hover:bg-netflix-gray-700'
                          >
                            {item}
                            <span
                              onClick={(event) => removeFromHistory(item, event)}
                              className='absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-netflix-gray-600 text-xs text-white opacity-0 transition-opacity group-hover:opacity-100'
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
          {searchParams.get('q') ? (
            <SearchResultsPanel
              query={searchParams.get('q') || ''}
              loading={loading}
              results={results}
              searchError={searchError}
              sourceStatus={sourceStatus}
              sourceTests={sourceTests}
              viewMode={viewMode}
              onViewModeChange={setViewMode}
            />
          ) : (
            <div className='mx-auto max-w-3xl space-y-8'>
              {history.length > 0 && (
                <section>
                  <div className='mb-4 flex items-center justify-between'>
                    <h2 className='text-lg font-bold text-white'>最近搜索</h2>
                    <button
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
                        onClick={() => handleSuggestionClick(item)}
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
                      onClick={() => handleSuggestionClick(item)}
                      className='rounded-2xl border border-netflix-gray-800 bg-netflix-surface/60 px-4 py-4 text-left text-sm text-netflix-gray-200 transition-colors hover:border-netflix-gray-600 hover:text-white'
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
