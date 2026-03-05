'use client';

import React, { useEffect, useState, useCallback, useRef } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { motion } from 'framer-motion';
import TopNav from '@/components/layout/TopNav';
import ContentCard from '@/components/home/ContentCard';
import { Loader2, Calendar } from 'lucide-react';

interface AnimeItem {
  id: string;
  title: string;
  cover: string;
  rate: string;
  year: string;
}

const tabs = [
  { label: '番剧', value: 'anime' },
  { label: '每日放送', value: 'calendar' },
];

const categories = [
  { label: '番剧', value: '番剧' },
  { label: '剧场版', value: '剧场版' },
];

const regions = ['全部', '日本', '中国', '欧美'];

const weekdays = [
  { label: '周一', value: 'Mon' },
  { label: '周二', value: 'Tue' },
  { label: '周三', value: 'Wed' },
  { label: '周四', value: 'Thu' },
  { label: '周五', value: 'Fri' },
  { label: '周六', value: 'Sat' },
  { label: '周日', value: 'Sun' },
];

export default function AnimePage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [items, setItems] = useState<AnimeItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [page, setPage] = useState(0);
  const [activeTab, setActiveTab] = useState(searchParams.get('tab') || 'anime');
  const [selectedCategory, setSelectedCategory] = useState(searchParams.get('category') || '番剧');
  const [selectedRegion, setSelectedRegion] = useState(searchParams.get('region') || '全部');
  const [selectedWeekday, setSelectedWeekday] = useState(searchParams.get('weekday') || 'Mon');
  const observerRef = useRef<IntersectionObserver | null>(null);
  const loadingRef = useRef<HTMLDivElement>(null);

  const fetchData = useCallback(async (pageNum: number, isLoadMore = false) => {
    if (!isLoadMore) setLoading(true);
    else setLoadingMore(true);

    try {
      let newItems: AnimeItem[] = [];

      if (activeTab === 'calendar') {
        // 获取每日放送数据
        const response = await fetch('/api/bangumi/calendar');
        if (response.ok) {
          const data = await response.json();
          const dayData = data.find((d: any) => d.weekday.en === selectedWeekday);
          if (dayData) {
            newItems = dayData.items.map((item: any) => ({
              id: item.id?.toString(),
              title: item.name_cn || item.name,
              cover: item.images?.large || item.images?.common || item.images?.medium || '/placeholder-poster.svg',
              rate: item.rating?.score?.toFixed(1) || '',
              year: item.air_date?.split('-')[0] || '',
            }));
          }
        }
        setHasMore(false);
      } else {
        // 获取番剧/剧场版数据
        const params = new URLSearchParams({
          kind: 'tv',
          limit: '25',
          start: (pageNum * 25).toString(),
          category: '动画',
          format: selectedCategory === '番剧' ? '电视剧' : '电影',
        });

        if (selectedRegion !== '全部') {
          params.set('region', selectedRegion);
        }

        const response = await fetch(`/api/douban/recommends?${params}`);
        if (!response.ok) throw new Error('获取数据失败');
        
        const data = await response.json();
        newItems = (data.list || []).map((item: any) => ({
          id: item.id?.toString() || Math.random().toString(),
          title: item.title,
          cover: item.poster || item.cover || '/placeholder-poster.svg',
          rate: item.rate || '',
          year: item.year || '',
        }));
        
        setHasMore(newItems.length === 25);
      }

      if (isLoadMore) {
        setItems(prev => [...prev, ...newItems]);
      } else {
        setItems(newItems);
      }
    } catch (error) {
      console.error('Fetch anime error:', error);
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }, [activeTab, selectedCategory, selectedRegion, selectedWeekday]);

  useEffect(() => {
    setPage(0);
    fetchData(0, false);
    
    const params = new URLSearchParams();
    params.set('tab', activeTab);
    if (activeTab === 'anime') {
      params.set('category', selectedCategory);
      if (selectedRegion !== '全部') params.set('region', selectedRegion);
    } else {
      params.set('weekday', selectedWeekday);
    }
    router.replace(`/anime?${params.toString()}`);
  }, [activeTab, selectedCategory, selectedRegion, selectedWeekday]);

  useEffect(() => {
    if (activeTab === 'calendar' || !loadingRef.current || !hasMore) return;

    observerRef.current = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && !loadingMore && hasMore) {
          setPage(prev => {
            const nextPage = prev + 1;
            fetchData(nextPage, true);
            return nextPage;
          });
        }
      },
      { threshold: 0.1 }
    );

    observerRef.current.observe(loadingRef.current);
    return () => observerRef.current?.disconnect();
  }, [hasMore, loadingMore, fetchData, activeTab]);

  return (
    <main className="min-h-screen bg-[#141414]">
      <TopNav />
      
      {/* Hero Header */}
      <div className="relative h-[50vh] min-h-[400px] overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-b from-purple-600/20 to-[#141414]" />
        <div className="absolute inset-0 bg-gradient-to-t from-[#141414] via-transparent to-black/50" />
        
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="text-center px-4">
            <motion.h1 
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="text-4xl md:text-6xl font-black text-white mb-4"
            >
              动漫
            </motion.h1>
            <motion.p 
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 }}
              className="text-gray-400 text-lg"
            >
              热血番剧 · 精彩剧场 · 追番必备
            </motion.p>
          </div>
        </div>
      </div>

      {/* Tabs & Filters */}
      <div className="sticky top-16 z-40 bg-[#141414]/95 backdrop-blur-md border-b border-gray-800">
        <div className="max-w-[1920px] mx-auto px-4 sm:px-8 py-4">
          {/* Main Tabs */}
          <div className="flex items-center gap-2 mb-4">
            {tabs.map((tab) => (
              <button
                key={tab.value}
                onClick={() => setActiveTab(tab.value)}
                className={`px-6 py-2 rounded-full text-sm font-medium transition-colors ${
                  activeTab === tab.value
                    ? 'bg-[#E50914] text-white'
                    : 'bg-gray-800 text-gray-300 hover:bg-gray-700'
                }`}
              >
                {tab.value === 'calendar' && <Calendar className="w-4 h-4 inline mr-1" />}
                {tab.label}
              </button>
            ))}
          </div>

          {/* Filters based on tab */}
          {activeTab === 'anime' ? (
            <>
              <div className="flex items-center gap-4 overflow-x-auto scrollbar-hide pb-2">
                <span className="text-gray-400 text-sm whitespace-nowrap">类型：</span>
                {categories.map((cat) => (
                  <button
                    key={cat.value}
                    onClick={() => setSelectedCategory(cat.value)}
                    className={`px-4 py-1.5 rounded-full text-sm whitespace-nowrap transition-colors ${
                      selectedCategory === cat.value
                        ? 'bg-white text-black'
                        : 'bg-gray-800 text-gray-300 hover:bg-gray-700'
                    }`}
                  >
                    {cat.label}
                  </button>
                ))}
              </div>

              <div className="flex items-center gap-4 overflow-x-auto scrollbar-hide mt-3">
                <span className="text-gray-400 text-sm whitespace-nowrap">地区：</span>
                {regions.map((region) => (
                  <button
                    key={region}
                    onClick={() => setSelectedRegion(region)}
                    className={`px-3 py-1 rounded-full text-sm whitespace-nowrap transition-colors ${
                      selectedRegion === region
                        ? 'bg-white text-black'
                        : 'bg-gray-800/50 text-gray-400 hover:bg-gray-800'
                    }`}
                  >
                    {region}
                  </button>
                ))}
              </div>
            </>
          ) : (
            <div className="flex items-center gap-2 overflow-x-auto scrollbar-hide">
              {weekdays.map((day) => (
                <button
                  key={day.value}
                  onClick={() => setSelectedWeekday(day.value)}
                  className={`px-4 py-2 rounded-full text-sm whitespace-nowrap transition-colors ${
                    selectedWeekday === day.value
                      ? 'bg-[#E50914] text-white'
                      : 'bg-gray-800 text-gray-300 hover:bg-gray-700'
                  }`}
                >
                  {day.label}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Content Grid */}
      <div className="max-w-[1920px] mx-auto px-4 sm:px-8 py-8">
        {loading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="w-10 h-10 text-[#E50914] animate-spin" />
          </div>
        ) : (
          <>
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 sm:gap-6">
              {items.map((item, index) => (
                <ContentCard
                  key={`${item.id}-${index}`}
                  title={item.title}
                  cover={item.cover}
                  rating={item.rate}
                  year={item.year}
                  type="anime"
                />
              ))}
            </div>

            {items.length === 0 && !loading && (
              <div className="text-center py-20 text-gray-500">
                暂无相关内容
              </div>
            )}

            {activeTab !== 'calendar' && (
              <>
                <div ref={loadingRef} className="flex justify-center py-8">
                  {loadingMore && (
                    <Loader2 className="w-8 h-8 text-[#E50914] animate-spin" />
                  )}
                </div>

                {!hasMore && items.length > 0 && (
                  <div className="text-center py-8 text-gray-500">
                    已加载全部内容
                  </div>
                )}
              </>
            )}
          </>
        )}
      </div>
    </main>
  );
}
