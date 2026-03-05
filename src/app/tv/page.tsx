'use client';

import React, { useEffect, useState, useCallback, useRef } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { motion } from 'framer-motion';
import TopNav from '@/components/layout/TopNav';
import ContentCard from '@/components/home/ContentCard';
import { Loader2 } from 'lucide-react';

interface TVItem {
  id: string;
  title: string;
  cover: string;
  rate: string;
  year: string;
}

const categories = [
  { label: '最近热门', value: '最近热门' },
  { label: '国产剧', value: '国产剧' },
  { label: '美剧', value: '美剧' },
  { label: '日剧', value: '日剧' },
  { label: '韩剧', value: '韩剧' },
  { label: '港剧', value: '港剧' },
  { label: '台剧', value: '台剧' },
  { label: '泰剧', value: '泰剧' },
];

const types = ['全部', '剧情', '爱情', '喜剧', '悬疑', '古装', '动作', '家庭', '犯罪', '奇幻'];

export default function TVPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [tvs, setTvs] = useState<TVItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [page, setPage] = useState(0);
  const [selectedCategory, setSelectedCategory] = useState(searchParams.get('category') || '最近热门');
  const [selectedType, setSelectedType] = useState(searchParams.get('type') || '全部');
  const observerRef = useRef<IntersectionObserver | null>(null);
  const loadingRef = useRef<HTMLDivElement>(null);

  const fetchTVs = useCallback(async (pageNum: number, isLoadMore = false) => {
    if (!isLoadMore) setLoading(true);
    else setLoadingMore(true);

    try {
      const params = new URLSearchParams({
        kind: 'tv',
        category: selectedCategory,
        type: selectedType === '全部' ? 'tv' : selectedType,
        limit: '25',
        start: (pageNum * 25).toString(),
      });

      const response = await fetch(`/api/douban/categories?${params}`);
      if (!response.ok) throw new Error('获取数据失败');
      
      const data = await response.json();
      const newItems = (data.list || []).map((item: any) => ({
        id: item.id?.toString() || Math.random().toString(),
        title: item.title,
        cover: item.poster || item.cover || '/placeholder-poster.svg',
        rate: item.rate || '',
        year: item.year || '',
      }));

      if (isLoadMore) {
        setTvs(prev => [...prev, ...newItems]);
      } else {
        setTvs(newItems);
      }
      
      setHasMore(newItems.length === 25);
    } catch (error) {
      console.error('Fetch TV error:', error);
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }, [selectedCategory, selectedType]);

  useEffect(() => {
    setPage(0);
    fetchTVs(0, false);
    
    const params = new URLSearchParams();
    params.set('category', selectedCategory);
    if (selectedType !== '全部') params.set('type', selectedType);
    router.replace(`/tv?${params.toString()}`);
  }, [selectedCategory, selectedType]);

  useEffect(() => {
    if (!loadingRef.current || !hasMore) return;

    observerRef.current = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && !loadingMore && hasMore) {
          setPage(prev => {
            const nextPage = prev + 1;
            fetchTVs(nextPage, true);
            return nextPage;
          });
        }
      },
      { threshold: 0.1 }
    );

    observerRef.current.observe(loadingRef.current);
    return () => observerRef.current?.disconnect();
  }, [hasMore, loadingMore, fetchTVs]);

  return (
    <main className="min-h-screen bg-[#141414]">
      <TopNav />
      
      {/* Hero Header */}
      <div className="relative h-[50vh] min-h-[400px] overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-b from-blue-600/20 to-[#141414]" />
        <div className="absolute inset-0 bg-gradient-to-t from-[#141414] via-transparent to-black/50" />
        
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="text-center px-4">
            <motion.h1 
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="text-4xl md:text-6xl font-black text-white mb-4"
            >
              电视剧
            </motion.h1>
            <motion.p 
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 }}
              className="text-gray-400 text-lg"
            >
              热播剧集 · 精彩连播 · 追剧不停
            </motion.p>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="sticky top-16 z-40 bg-[#141414]/95 backdrop-blur-md border-b border-gray-800">
        <div className="max-w-[1920px] mx-auto px-4 sm:px-8 py-4">
          <div className="flex items-center gap-4 overflow-x-auto scrollbar-hide pb-2">
            <span className="text-gray-400 text-sm whitespace-nowrap">分类：</span>
            {categories.map((cat) => (
              <button
                key={cat.value}
                onClick={() => setSelectedCategory(cat.value)}
                className={`px-4 py-1.5 rounded-full text-sm whitespace-nowrap transition-colors ${
                  selectedCategory === cat.value
                    ? 'bg-[#E50914] text-white'
                    : 'bg-gray-800 text-gray-300 hover:bg-gray-700'
                }`}
              >
                {cat.label}
              </button>
            ))}
          </div>

          <div className="flex items-center gap-4 overflow-x-auto scrollbar-hide mt-3">
            <span className="text-gray-400 text-sm whitespace-nowrap">类型：</span>
            {types.map((type) => (
              <button
                key={type}
                onClick={() => setSelectedType(type)}
                className={`px-3 py-1 rounded-full text-sm whitespace-nowrap transition-colors ${
                  selectedType === type
                    ? 'bg-white text-black'
                    : 'bg-gray-800/50 text-gray-400 hover:bg-gray-800'
                }`}
              >
                {type}
              </button>
            ))}
          </div>
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
              {tvs.map((tv, index) => (
                <ContentCard
                  key={`${tv.id}-${index}`}
                  title={tv.title}
                  cover={tv.cover}
                  rating={tv.rate}
                  year={tv.year}
                  type="tv"
                />
              ))}
            </div>

            {tvs.length === 0 && !loading && (
              <div className="text-center py-20 text-gray-500">
                暂无相关内容
              </div>
            )}

            <div ref={loadingRef} className="flex justify-center py-8">
              {loadingMore && (
                <Loader2 className="w-8 h-8 text-[#E50914] animate-spin" />
              )}
            </div>

            {!hasMore && tvs.length > 0 && (
              <div className="text-center py-8 text-gray-500">
                已加载全部内容
              </div>
            )}
          </>
        )}
      </div>
    </main>
  );
}
