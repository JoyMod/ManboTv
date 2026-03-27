'use client';

import { AnimatePresence, motion } from 'framer-motion';
import {
  ChevronDown,
  ChevronUp,
  Film,
  LayoutGrid,
  List,
  Loader2,
  SlidersHorizontal,
  X,
} from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import React, {
  memo,
  ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';

import BrowseListItem from '@/components/browse/BrowseListItem';
import TopNav from '@/components/layout/TopNav';
import ContentCard from '@/components/ui/ContentCard';

export interface BrowseItem {
  id: string;
  title: string;
  cover: string;
  rate: string;
  year: string;
  type?: 'movie' | 'tv' | 'variety' | 'anime';
}

interface BrowseBootstrapResponse {
  items?: BrowseItem[];
  total_results?: number;
  has_more?: boolean;
}

interface FilterOption {
  label: string;
  value: string;
}

export interface FilterGroup {
  id: string;
  label: string;
  options: FilterOption[];
  multiple?: boolean;
}

export interface SortOption {
  label: string;
  value: string;
}

export interface BrowsePageProps {
  title: string;
  subtitle: string;
  kind: string;
  filterGroups: FilterGroup[];
  sortOptions: SortOption[];
  heroGradient?: string;
  icon?: ReactNode;
  children?: ReactNode;
}

const PAGE_SIZE = 25;
const SCROLL_THRESHOLD = 200;
const FILTER_CARD_VISIBLE_COUNT = 10;
const FILTER_PANEL_ANIMATION_DURATION = 0.2;
const CARD_STAGGER_DELAY = 0.02;
const CARD_STAGGER_MAX_DELAY = 0.5;
const BROWSE_BOOTSTRAP_ENDPOINT = '/api/browse/bootstrap';
const STICKY_NAV_TOP = '64px';

const FilterCard = memo(function FilterCard({
  group,
  selectedValue,
  onChange,
}: {
  group: FilterGroup;
  selectedValue: string | string[];
  onChange: (value: string | string[]) => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const values = Array.isArray(selectedValue) ? selectedValue : [selectedValue];
  const displayOptions = expanded
    ? group.options
    : group.options.slice(0, FILTER_CARD_VISIBLE_COUNT);
  const hasMore = group.options.length > FILTER_CARD_VISIBLE_COUNT;

  return (
    <div className='rounded-xl border border-zinc-800 bg-zinc-900/80 p-4 backdrop-blur-sm'>
      <div className='mb-3 flex items-center justify-between'>
        <h3 className='text-sm font-medium text-zinc-500'>{group.label}</h3>
        {values.some((value) => value !== '全部') && (
          <button
            onClick={() => onChange(group.multiple ? ['全部'] : '全部')}
            className='text-xs text-zinc-600 transition-colors hover:text-white'
          >
            重置
          </button>
        )}
      </div>

      <div className='flex flex-wrap gap-2'>
        {displayOptions.map((option) => {
          const isSelected = values.includes(option.value);
          const isAll = option.value === '全部';

          return (
            <button
              key={option.value}
              onClick={() => {
                if (group.multiple) {
                  if (isAll) {
                    onChange(['全部']);
                    return;
                  }

                  const nextValues = isSelected
                    ? values.filter((value) => value !== option.value)
                    : [...values.filter((value) => value !== '全部'), option.value];

                  onChange(nextValues.length > 0 ? nextValues : ['全部']);
                  return;
                }

                onChange(option.value);
              }}
              className={`rounded-full px-3 py-1.5 text-xs font-medium transition-all duration-200 ${
                isSelected && !isAll
                  ? 'bg-white text-black shadow-lg shadow-white/10'
                  : 'border border-zinc-700 bg-transparent text-zinc-400 hover:border-zinc-500 hover:text-white'
              }`}
            >
              {option.label}
            </button>
          );
        })}

        {!expanded && hasMore && (
          <button
            onClick={() => setExpanded(true)}
            className='rounded-full border border-zinc-700 px-3 py-1.5 text-xs font-medium text-zinc-500 transition-colors hover:border-zinc-500 hover:text-white'
          >
            +{group.options.length - FILTER_CARD_VISIBLE_COUNT}
          </button>
        )}
      </div>

      {expanded && hasMore && (
        <button
          onClick={() => setExpanded(false)}
          className='mt-3 flex items-center gap-1 text-xs text-zinc-600 transition-colors hover:text-white'
        >
          <ChevronUp className='h-3 w-3' />
          收起
        </button>
      )}
    </div>
  );
});

const LoadingSpinner = memo(function LoadingSpinner({
  size = 'md',
}: {
  size?: 'sm' | 'md' | 'lg';
}) {
  const sizeClasses = {
    sm: 'h-4 w-4',
    md: 'h-8 w-8',
    lg: 'h-12 w-12',
  };

  return <Loader2 className={`${sizeClasses[size]} animate-spin text-netflix-red`} />;
});

function buildFilterQueryParams(selectedFilters: Record<string, string | string[]>) {
  const params = new URLSearchParams();

  Object.entries(selectedFilters).forEach(([key, value]) => {
    if (Array.isArray(value)) {
      const filtered = value.filter((item) => item && item !== '全部');
      if (filtered.length > 0) {
        params.set(key, filtered.join(','));
      }
      return;
    }

    if (value && value !== '全部') {
      params.set(key, value);
    }
  });

  return params;
}

export default function BrowsePage({
  title,
  subtitle,
  kind,
  filterGroups,
  sortOptions,
  heroGradient = 'from-netflix-red/20',
  icon = <Film className='h-16 w-16 text-netflix-red' />,
  children,
}: BrowsePageProps) {
  const router = useRouter();
  const searchParams = useSearchParams();

  const [items, setItems] = useState<BrowseItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [showFilterPanel, setShowFilterPanel] = useState(false);
  const [totalResults, setTotalResults] = useState(0);
  const [selectedFilters, setSelectedFilters] = useState<Record<string, string | string[]>>(() => {
    const initial: Record<string, string | string[]> = {};
    filterGroups.forEach((group) => {
      const value = searchParams.get(group.id);
      if (value) {
        initial[group.id] = group.multiple ? value.split(',') : value;
      } else {
        initial[group.id] = group.multiple ? ['全部'] : '全部';
      }
    });
    return initial;
  });
  const [selectedSort, setSelectedSort] = useState(
    searchParams.get('sort') || sortOptions[0]?.value || 'default'
  );

  const pageRef = useRef(0);
  const hasMoreRef = useRef(true);
  const loadingMoreRef = useRef(false);
  const observerRef = useRef<IntersectionObserver | null>(null);
  const loadMoreTriggerRef = useRef<HTMLDivElement>(null);

  const activeFilterCount = useMemo(
    () =>
      Object.values(selectedFilters).reduce((count, value) => {
        if (Array.isArray(value)) {
          return count + value.filter((item) => item !== '全部').length;
        }
        return value !== '全部' ? count + 1 : count;
      }, 0),
    [selectedFilters]
  );

  const activeFilterChips = useMemo(
    () =>
      filterGroups.flatMap((group) => {
        const rawValue = selectedFilters[group.id];
        const values = Array.isArray(rawValue) ? rawValue : [rawValue];

        return values
          .filter((value) => value && value !== '全部')
          .map((value) => ({
            groupId: group.id,
            label: `${group.label} · ${value}`,
            value,
            multiple: Boolean(group.multiple),
          }));
      }),
    [filterGroups, selectedFilters]
  );

  const fetchItems = useCallback(
    async (isLoadMore = false) => {
      if (isLoadMore && (loadingMoreRef.current || !hasMoreRef.current)) {
        return;
      }

      if (isLoadMore) {
        loadingMoreRef.current = true;
        setLoadingMore(true);
      } else {
        setLoading(true);
        setError(null);
      }

      try {
        const currentPage = isLoadMore ? pageRef.current + 1 : 0;
        const params = buildFilterQueryParams(selectedFilters);
        params.set('kind', kind);
        params.set('page', String(currentPage));
        params.set('pageSize', String(PAGE_SIZE));
        params.set('sort', selectedSort);

        const response = await fetch(`${BROWSE_BOOTSTRAP_ENDPOINT}?${params.toString()}`);
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const payload = (await response.json()) as BrowseBootstrapResponse;
        const nextItems = Array.isArray(payload.items) ? payload.items : [];

        setItems(nextItems);
        setTotalResults(payload.total_results ?? nextItems.length);
        hasMoreRef.current = Boolean(payload.has_more);
        setHasMore(hasMoreRef.current);
        pageRef.current = currentPage;
      } catch (fetchError) {
        setError(fetchError instanceof Error ? fetchError.message : '获取数据失败');
      } finally {
        setLoading(false);
        loadingMoreRef.current = false;
        setLoadingMore(false);
      }
    },
    [kind, selectedFilters, selectedSort]
  );

  useEffect(() => {
    pageRef.current = 0;
    hasMoreRef.current = true;
    loadingMoreRef.current = false;
    setHasMore(true);
    setLoadingMore(false);
    void fetchItems(false);

    const params = buildFilterQueryParams(selectedFilters);
    if (selectedSort !== 'default') {
      params.set('sort', selectedSort);
    }

    const query = params.toString();
    router.replace(query ? `/${kind}?${query}` : `/${kind}`, { scroll: false });
  }, [fetchItems, kind, router, selectedFilters, selectedSort]);

  useEffect(() => {
    const trigger = loadMoreTriggerRef.current;
    if (!trigger) {
      return;
    }

    observerRef.current?.disconnect();
    observerRef.current = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMoreRef.current && !loadingMoreRef.current) {
          void fetchItems(true);
        }
      },
      {
        root: null,
        rootMargin: `${SCROLL_THRESHOLD}px`,
        threshold: 0,
      }
    );

    observerRef.current.observe(trigger);

    return () => {
      observerRef.current?.disconnect();
    };
  }, [fetchItems, items.length, viewMode]);

  const handleFilterChange = useCallback((groupId: string, value: string | string[]) => {
    pageRef.current = 0;
    hasMoreRef.current = true;
    setHasMore(true);
    setSelectedFilters((previous) => ({ ...previous, [groupId]: value }));
  }, []);

  const handleClearFilters = useCallback(() => {
    const cleared: Record<string, string | string[]> = {};
    filterGroups.forEach((group) => {
      cleared[group.id] = group.multiple ? ['全部'] : '全部';
    });
    pageRef.current = 0;
    hasMoreRef.current = true;
    setHasMore(true);
    setSelectedFilters(cleared);
  }, [filterGroups]);

  const handleRemoveFilterChip = useCallback((
    groupId: string,
    value: string,
    multiple: boolean
  ) => {
    pageRef.current = 0;
    hasMoreRef.current = true;
    setHasMore(true);
    setSelectedFilters((previous) => {
      const current = previous[groupId];
      if (multiple && Array.isArray(current)) {
        const nextValues = current.filter((item) => item !== value && item !== '全部');
        return {
          ...previous,
          [groupId]: nextValues.length > 0 ? nextValues : ['全部'],
        };
      }

      return {
        ...previous,
        [groupId]: '全部',
      };
    });
  }, []);

  const handleToggleFilterPanel = useCallback(() => {
    setShowFilterPanel((previous) => {
      const next = !previous;
      if (next) {
        window.scrollTo({ top: 0, behavior: 'smooth' });
      }
      return next;
    });
  }, []);

  return (
    <main className='min-h-screen bg-[#0a0a0a]'>
      <TopNav />

      <div className='relative pt-16'>
        <div className={`absolute inset-0 bg-gradient-to-b ${heroGradient} to-[#0a0a0a]`} />
        <div className='absolute inset-0 bg-gradient-to-t from-[#0a0a0a] via-transparent to-black/30' />

        <div className='relative px-4 pb-6 pt-8 sm:px-8'>
          <div className='flex items-center gap-4'>
            <div className='rounded-2xl bg-white/5 p-4 backdrop-blur-sm'>{icon}</div>
            <div>
              <h1 className='text-3xl font-black text-white md:text-4xl'>{title}</h1>
              <p className='mt-2 text-zinc-400'>{subtitle}</p>
            </div>
          </div>
        </div>
      </div>

      <div className='relative z-10 px-4 pb-4 sm:px-8'>
        <button
          onClick={handleToggleFilterPanel}
          className='mb-4 flex items-center gap-2 text-sm text-zinc-500 transition-colors hover:text-white'
        >
          {showFilterPanel ? (
            <ChevronUp className='h-4 w-4' />
          ) : (
            <ChevronDown className='h-4 w-4' />
          )}
          {showFilterPanel ? '收起筛选' : '展开筛选'}
          {!showFilterPanel && activeFilterCount > 0 && (
            <span className='rounded-full bg-netflix-red px-2 py-0.5 text-xs text-white'>
              {activeFilterCount}
            </span>
          )}
        </button>

        <AnimatePresence>
          {showFilterPanel && (
            <motion.div
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: 'auto', opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              transition={{ duration: FILTER_PANEL_ANIMATION_DURATION }}
              className='overflow-hidden'
            >
              <div className='mb-6 grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4'>
                {filterGroups.map((group) => (
                  <FilterCard
                    key={group.id}
                    group={group}
                    selectedValue={selectedFilters[group.id]}
                    onChange={(value) => handleFilterChange(group.id, value)}
                  />
                ))}
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      <div
        className='sticky top-16 z-40 border-y border-zinc-800 bg-zinc-900/95 backdrop-blur-md'
        style={{ position: 'sticky', top: STICKY_NAV_TOP }}
      >
        <div className='mx-auto flex max-w-[1920px] items-center justify-between px-4 py-3 sm:px-8'>
          <div className='flex items-center gap-3'>
            <button
              onClick={handleToggleFilterPanel}
              className='flex items-center gap-2 rounded-full bg-zinc-800 px-4 py-2 text-sm text-white transition-colors hover:bg-zinc-700'
            >
              <SlidersHorizontal className='h-4 w-4' />
              <span className='hidden sm:inline'>筛选</span>
              {activeFilterCount > 0 && (
                <span className='ml-1 flex h-5 w-5 items-center justify-center rounded-full bg-netflix-red text-xs font-bold'>
                  {activeFilterCount}
                </span>
              )}
            </button>

            {activeFilterCount > 0 && (
              <button
                onClick={handleClearFilters}
                className='flex items-center gap-1 text-sm text-zinc-500 transition-colors hover:text-white'
              >
                <X className='h-4 w-4' />
                <span className='hidden sm:inline'>清除</span>
              </button>
            )}
          </div>

          <div className='hidden text-sm text-zinc-400 md:block'>
            找到 <span className='font-bold text-white'>{totalResults}</span> 部影片
          </div>

          <div className='flex items-center gap-2'>
            <select
              value={selectedSort}
              onChange={(event) => setSelectedSort(event.target.value)}
              className='rounded-full bg-zinc-800 px-3 py-2 text-sm text-white outline-none hover:bg-zinc-700'
            >
              {sortOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>

            <div className='hidden rounded-full bg-zinc-800 p-1 md:flex'>
              <button
                onClick={() => setViewMode('grid')}
                className={`rounded-full p-1.5 transition-colors ${
                  viewMode === 'grid' ? 'bg-zinc-600 text-white' : 'text-zinc-400 hover:text-white'
                }`}
              >
                <LayoutGrid className='h-4 w-4' />
              </button>
              <button
                onClick={() => setViewMode('list')}
                className={`rounded-full p-1.5 transition-colors ${
                  viewMode === 'list' ? 'bg-zinc-600 text-white' : 'text-zinc-400 hover:text-white'
                }`}
              >
                <List className='h-4 w-4' />
              </button>
            </div>
          </div>
        </div>
      </div>

      <div className='px-4 py-8 sm:px-8'>
        <div className='mx-auto max-w-[1920px]'>
          {activeFilterChips.length > 0 && (
            <div className='mb-6 flex flex-wrap items-center gap-2'>
              {activeFilterChips.map((chip) => (
                <button
                  key={`${chip.groupId}-${chip.value}`}
                  onClick={() =>
                    handleRemoveFilterChip(chip.groupId, chip.value, chip.multiple)
                  }
                  className='inline-flex items-center gap-2 rounded-full border border-zinc-700 bg-zinc-900/80 px-3 py-1.5 text-xs text-zinc-200 transition-colors hover:border-zinc-500 hover:text-white'
                >
                  {chip.label}
                  <X className='h-3 w-3' />
                </button>
              ))}
            </div>
          )}

          {error && (
            <div className='mb-6 rounded-lg border border-red-500/30 bg-red-500/10 p-4 text-center text-red-400'>
              {error}
              <button
                onClick={() => void fetchItems(false)}
                className='ml-2 text-sm underline hover:text-red-300'
              >
                重试
              </button>
            </div>
          )}

          {loading && items.length === 0 ? (
            <div className='flex items-center justify-center py-20'>
              <LoadingSpinner size='lg' />
            </div>
          ) : (
            <>
              {viewMode === 'grid' && (
                <div className='grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-7'>
                  {items.map((item, index) => (
                    <motion.div
                      key={`${item.id}-${index}`}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{
                        duration: 0.3,
                        delay: Math.min(index * CARD_STAGGER_DELAY, CARD_STAGGER_MAX_DELAY),
                      }}
                    >
                      <ContentCard
                        id={item.id}
                        title={item.title}
                        cover={item.cover}
                        rating={item.rate}
                        year={item.year}
                        type={item.type}
                      />
                    </motion.div>
                  ))}
                </div>
              )}

              {viewMode === 'list' && (
                <div className='space-y-4'>
                  {items.map((item) => (
                    <BrowseListItem
                      key={item.id}
                      title={item.title}
                      cover={item.cover}
                      rate={item.rate}
                      year={item.year}
                    />
                  ))}
                </div>
              )}

              {items.length === 0 && !loading && !error && (
                <div className='py-20 text-center'>
                  <p className='text-lg text-zinc-500'>暂无相关内容</p>
                  <p className='mt-2 text-sm text-zinc-600'>尝试调整筛选条件</p>
                </div>
              )}

              <div ref={loadMoreTriggerRef} className='flex h-20 items-center justify-center'>
                {loadingMore && <LoadingSpinner size='md' />}
                {!loadingMore && hasMore && items.length > 0 && (
                  <button
                    onClick={() => void fetchItems(true)}
                    className='rounded-full border border-zinc-700 px-4 py-2 text-sm text-zinc-300 transition-colors hover:border-zinc-500 hover:text-white'
                  >
                    加载更多
                  </button>
                )}
                {!hasMore && items.length > 0 && <p className='text-zinc-500'>已加载全部内容</p>}
              </div>
            </>
          )}

          {children}
        </div>
      </div>
    </main>
  );
}
