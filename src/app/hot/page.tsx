'use client';

import React, { useEffect, useState, useCallback, useRef } from 'react';
import { motion } from 'framer-motion';
import TopNav from '@/components/layout/TopNav';
import ContentCard from '@/components/home/ContentCard';
import { Loader2, Flame, TrendingUp, Star } from 'lucide-react';

interface HotItem {
  id: string;
  title: string;
  cover: string;
  rate: string;
  year: string;
  type: 'movie' | 'tv' | 'variety' | 'anime';
  hot: number;
}

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
  const [page, setPage] = useState(0);
  const observerRef = useRef<IntersectionObserver | null>(null);
  const loadingRef = useRef<HTMLDivElement>(null);

  const fetchData = useCallback(async (pageNum: number, isLoadMore = false) => {
    if (!isLoadMore) setLoading(true);
    else setLoadingMore(true);

    try {
      // 同时获取电影和电视剧的热门数据
      const [movieData, tvData] = await Promise.all([
        fetch('/api/douban/categories?kind=movie&category=热门&type=全部&limit=20&start=0').then(r => r.json()),
        fetch('/api/douban/categories?kind=tv&category=最近热门&type=tv&limit=20&start=0').then(r => r.json()),
      ]);

      const movies: HotItem[] = (movieData.list || []).map((item: any) => ({
        id: item.id?.toString() || Math.random().toString(),
        title: item.title,
        cover: item.poster || item.cover || '/placeholder-poster.svg',
        rate: item.rate || '',
        year: item.year || '',
        type: 'movie',
        hot: Math.floor(Math.random() * 10000) + 5000,
      }));

      const tvs: HotItem[] = (tvData.list || []).map((item: any) => ({
        id: item.id?.toString() || Math.random().toString(),
        title: item.title,
        cover: item.poster || item.cover || '/placeholder-poster.svg',
        rate: item.rate || '',
        year: item.year || '',
        type: 'tv',
        hot: Math.floor(Math.random() * 10000) + 5000,
      }));

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
          // 综合热门：交替排列
          const maxLen = Math.max(movies.length, tvs.length);
          for (let i = 0; i < maxLen; i++) {
            if (movies[i]) allItems.push(movies[i]);
            if (tvs[i]) allItems.push(tvs[i]);
          }
          allItems = allItems.sort((a, b) => b.hot - a.hot);
        }
      }

      if (isLoadMore) {
        setItems(prev => [...prev, ...allItems]);
      } else {
        setItems(allItems);
      }
      
      setHasMore(false); // 热门数据一次性加载
    } catch (error) {
      console.error('Fetch hot error:', error);
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }, [activeTab]);

  useEffect(() => {
    setPage(0);
    fetchData(0, false);
  }, [activeTab]);

  return (
    <main className="min-h-screen bg-[#141414]">
      <TopNav />
      
      {/* Hero Header */}
      <div className="relative h-[50vh] min-h-[400px] overflow-hidden">
        {/* Animated gradient background */}
        <div className="absolute inset-0 bg-gradient-to-br from-[#E50914]/30 via-orange-500/20 to-[#141414]" />
        <div className="absolute inset-0 bg-gradient-to-t from-[#141414] via-transparent to-black/50" />
        
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="text-center px-4">
            <motion.div
              initial={{ scale: 0.8, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              className="inline-flex items-center gap-2 px-4 py-2 bg-[#E50914] rounded-full text-white text-sm font-bold mb-4"
            >
              <Flame className="w-5 h-5" />
              全网热播榜
            </motion.div>
            <motion.h1 
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="text-4xl md:text-6xl font-black text-white mb-4"
            >
              最新热播
            </motion.h1>
            <motion.p 
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 }}
              className="text-gray-400 text-lg"
            >
              实时更新 · 热门排行 · 不容错过
            </motion.p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="sticky top-16 z-40 bg-[#141414]/95 backdrop-blur-md border-b border-gray-800">
        <div className="max-w-[1920px] mx-auto px-4 sm:px-8 py-4">
          <div className="flex items-center gap-4">
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
                  <Icon className="w-4 h-4" />
                  {tab.label}
                </button>
              );
            })}
          </div>
        </div>
      </div>

      {/* Content Grid with Ranking */}
      <div className="max-w-[1920px] mx-auto px-4 sm:px-8 py-8">
        {loading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="w-10 h-10 text-[#E50914] animate-spin" />
          </div>
        ) : (
          <>
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 sm:gap-6">
              {items.map((item, index) => (
                <div key={`${item.id}-${index}`} className="relative">
                  {/* Ranking Badge */}
                  {index < 10 && (
                    <div className={`absolute -top-2 -left-2 z-20 w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold ${
                      index < 3 
                        ? 'bg-[#E50914] text-white' 
                        : 'bg-gray-700 text-gray-300'
                    }`}>
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
              <div className="text-center py-20 text-gray-500">
                暂无相关内容
              </div>
            )}
          </>
        )}
      </div>
    </main>
  );
}
