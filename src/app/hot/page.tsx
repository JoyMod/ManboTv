'use client';

import { motion } from 'framer-motion';
import { Flame, Loader2, Star,TrendingUp } from 'lucide-react';
import React, { useCallback, useEffect, useRef,useState } from 'react';

import ContentCard from '@/components/home/ContentCard';
import TopNav from '@/components/layout/TopNav';

interface HotItem {
  id: string;
  title: string;
  cover: string;
  rate: string;
  year: string;
  type: 'movie' | 'tv' | 'variety' | 'anime';
  hot: number;
}

interface DoubanResponseItem {
  id?: string;
  title?: string;
  poster?: string;
  cover?: string;
  rate?: string;
  year?: string;
}

const PAGE_SIZE = 20;
const HOT_TAG = '热门';

const tabs = [
  { label: '综合热门', value: 'all', icon: Flame },
  { label: '电影榜', value: 'movie', icon: TrendingUp },
  { label: '剧集榜', value: 'tv', icon: Star },
];

export default function HotPage() {
  const [items, setItems] = useState<HotItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('all');
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(0);
  const observerRef = useRef<IntersectionObserver | null>(null);
  const loadingRef = useRef<HTMLDivElement>(null);

  const fetchDoubanList = useCallback(
    async (kind: 'movie' | 'tv', pageNum: number) => {
      const start = pageNum * PAGE_SIZE;
      const params = new URLSearchParams({
        type: kind,
        tag: HOT_TAG,
        pageSize: PAGE_SIZE.toString(),
        pageStart: start.toString(),
      });

      const response = await fetch(`/api/douban?${params.toString()}`);
      if (!response.ok) {
        throw new Error('获取热门内容失败');
      }

      const raw = await response.json();
      const list = Array.isArray(raw?.list)
        ? raw.list
        : Array.isArray(raw?.data?.list)
        ? raw.data.list
        : [];

      const normalized: HotItem[] = (list as DoubanResponseItem[]).map(
        (item, index) => {
          const rating = Number.parseFloat(item.rate || '0');
          const rankScore = PAGE_SIZE - index;
          const hotScore = Math.round(
            (Number.isFinite(rating) ? rating : 0) * 1000 + rankScore
          );

          return {
            id: item.id?.toString() || `${kind}-${start + index}`,
            title: item.title || '未知标题',
            cover: item.poster || item.cover || '/placeholder-poster.svg',
            rate: item.rate || '',
            year: item.year || '',
            type: kind,
            hot: hotScore,
          };
        }
      );

      return normalized;
    },
    []
  );

  const fetchData = useCallback(
    async (pageNum: number, isLoadMore = false) => {
      if (!isLoadMore) setLoading(true);
      else setLoadingMore(true);
      setError(null);

      try {
        const [movies, tvs] = await Promise.all([
          activeTab === 'tv'
            ? Promise.resolve([])
            : fetchDoubanList('movie', pageNum),
          activeTab === 'movie'
            ? Promise.resolve([])
            : fetchDoubanList('tv', pageNum),
        ]);

        let allItems: HotItem[] = [];

        switch (activeTab) {
          case 'movie': {
            allItems = movies.sort((a, b) => b.hot - a.hot);
            break;
          }
          case 'tv': {
            allItems = tvs.sort((a, b) => b.hot - a.hot);
            break;
          }
          default: {
            const maxLen = Math.max(movies.length, tvs.length);
            for (let i = 0; i < maxLen; i++) {
              if (movies[i]) allItems.push(movies[i]);
              if (tvs[i]) allItems.push(tvs[i]);
            }
            allItems = allItems.sort((a, b) => b.hot - a.hot);
          }
        }

        if (isLoadMore) {
          setItems((prev) => {
            const deduped = new Map<string, HotItem>();
            [...prev, ...allItems].forEach((item) => {
              deduped.set(`${item.type}-${item.id}`, item);
            });
            return Array.from(deduped.values());
          });
        } else {
          setItems(allItems);
        }

        const movieHasMore = movies.length === PAGE_SIZE;
        const tvHasMore = tvs.length === PAGE_SIZE;
        const nextHasMore =
          activeTab === 'movie'
            ? movieHasMore
            : activeTab === 'tv'
            ? tvHasMore
            : movieHasMore || tvHasMore;
        setHasMore(nextHasMore);
      } catch {
        setError('加载热门内容失败，请稍后重试');
      } finally {
        setLoading(false);
        setLoadingMore(false);
      }
    },
    [activeTab, fetchDoubanList]
  );

  useEffect(() => {
    setItems([]);
    setHasMore(true);
    setPage(0);
    fetchData(0, false);
  }, [activeTab, fetchData]);

  useEffect(() => {
    if (page === 0 || loadingMore || !hasMore) return;
    fetchData(page, true);
  }, [page, fetchData, hasMore, loadingMore]);

  useEffect(() => {
    if (loading || loadingMore || !hasMore) return;

    if (observerRef.current) {
      observerRef.current.disconnect();
    }

    observerRef.current = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) {
          setPage((prev) => prev + 1);
        }
      },
      { rootMargin: '240px' }
    );

    if (loadingRef.current) {
      observerRef.current.observe(loadingRef.current);
    }

    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect();
      }
    };
  }, [loading, loadingMore, hasMore]);

  return (
    <main className='min-h-screen bg-[#141414]'>
      <TopNav />

      <div className='relative h-[50vh] min-h-[400px] overflow-hidden'>
        <div className='absolute inset-0 bg-gradient-to-br from-[#E50914]/30 via-orange-500/20 to-[#141414]' />
        <div className='absolute inset-0 bg-gradient-to-t from-[#141414] via-transparent to-black/50' />

        <div className='absolute inset-0 flex items-center justify-center'>
          <div className='text-center px-4'>
            <motion.div
              initial={{ scale: 0.8, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              className='inline-flex items-center gap-2 px-4 py-2 bg-[#E50914] rounded-full text-white text-sm font-bold mb-4'
            >
              <Flame className='w-5 h-5' />
              全网热播榜
            </motion.div>
            <motion.h1
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className='text-4xl md:text-6xl font-black text-white mb-4'
            >
              最新热播
            </motion.h1>
            <motion.p
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 }}
              className='text-gray-400 text-lg'
            >
              实时更新 · 热门排行 · 不容错过
            </motion.p>
          </div>
        </div>
      </div>

      <div className='sticky top-16 z-40 bg-[#141414]/95 backdrop-blur-md border-b border-gray-800'>
        <div className='max-w-[1920px] mx-auto px-4 sm:px-8 py-4'>
          <div className='flex items-center gap-4'>
            {tabs.map((tab) => {
              const Icon = tab.icon;
              return (
                <button
                  key={tab.value}
                  onClick={() => setActiveTab(tab.value)}
                  className={`flex items-center gap-2 px-6 py-3 rounded-full text-sm font-medium transition-all ${
                    activeTab === tab.value
                      ? 'bg-[#E50914] text-white'
                      : 'bg-gray-800 text-gray-300 hover:bg-gray-700'
                  }`}
                >
                  <Icon className='w-4 h-4' />
                  {tab.label}
                </button>
              );
            })}
          </div>
        </div>
      </div>

      <div className='max-w-[1920px] mx-auto px-4 sm:px-8 py-8'>
        {loading ? (
          <div className='flex items-center justify-center py-20'>
            <Loader2 className='w-10 h-10 text-[#E50914] animate-spin' />
          </div>
        ) : (
          <>
            <div className='grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 sm:gap-6'>
              {items.map((item, index) => (
                <div key={`${item.id}-${index}`} className='relative'>
                  {index < 10 && (
                    <div
                      className={`absolute -top-2 -left-2 z-20 w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold ${
                        index < 3
                          ? 'bg-[#E50914] text-white'
                          : 'bg-gray-700 text-gray-300'
                      }`}
                    >
                      {index + 1}
                    </div>
                  )}
                  <ContentCard
                    title={item.title}
                    cover={item.cover}
                    rating={item.rate}
                    year={item.year}
                    type={item.type}
                  />
                </div>
              ))}
            </div>

            {items.length === 0 && !loading && (
              <div className='text-center py-20 text-gray-500'>
                暂无相关内容
              </div>
            )}

            {error && (
              <div className='text-center py-8 text-red-400 text-sm'>
                {error}
              </div>
            )}

            {hasMore && (
              <div
                ref={loadingRef}
                className='flex items-center justify-center py-8'
              >
                {loadingMore ? (
                  <Loader2 className='w-6 h-6 text-[#E50914] animate-spin' />
                ) : (
                  <span className='text-xs text-gray-500'>
                    向下滚动加载更多
                  </span>
                )}
              </div>
            )}
          </>
        )}
      </div>
    </main>
  );
}
