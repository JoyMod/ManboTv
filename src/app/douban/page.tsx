'use client';

import { motion } from 'framer-motion';
import { Loader2 } from 'lucide-react';
import { useSearchParams } from 'next/navigation';
import React, { useEffect,useState } from 'react';

import TopNav from '@/components/layout/TopNav';
import ContentCard from '@/components/ui/ContentCard';

interface DoubanListItem {
  id?: string;
  title?: string;
  poster?: string;
  cover?: string;
  year?: string;
  rate?: string;
}

interface DoubanCardItem {
  id: string;
  title: string;
  cover: string;
  year: string;
  rating: string;
  type: 'movie' | 'tv';
}

export default function DoubanPage() {
  const searchParams = useSearchParams();
  const type = searchParams.get('type') || 'movie';
  const [items, setItems] = useState<DoubanCardItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const apiType = type === 'movie' ? 'movie' : 'tv';
        const tag =
          type === 'anime' ? '日本动画' : type === 'variety' ? '综艺' : '热门';
        const response = await fetch(
          `/api/douban?type=${apiType}&tag=${encodeURIComponent(
            tag
          )}&pageSize=25&pageStart=0`
        );
        if (!response.ok) {
          throw new Error('请求豆瓣数据失败');
        }
        const data = await response.json();
        const rawItems = Array.isArray(data?.list) ? data.list : [];
        const formatted = (rawItems as DoubanListItem[]).map((item, index) => ({
          id: item.id?.toString() || `${type}-${index}`,
          title: item.title || '未知标题',
          cover: item.poster || item.cover || '/placeholder-poster.svg',
          year: item.year || '',
          rating: item.rate || '',
          type: (type === 'movie' ? 'movie' : 'tv') as 'movie' | 'tv',
        }));
        setItems(formatted);
      } catch {
        setItems([]);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [type]);

  const title =
    type === 'movie'
      ? '电影'
      : type === 'tv'
      ? '电视剧'
      : type === 'anime'
      ? '动漫'
      : '综艺';

  return (
    <main className='min-h-screen bg-netflix-black'>
      <TopNav />

      <div className='pt-24 pb-20 px-4 sm:px-8'>
        <div className='max-w-[1920px] mx-auto'>
          <h1 className='text-3xl font-bold text-white mb-8'>{title}</h1>

          {loading ? (
            <div className='flex items-center justify-center py-20'>
              <Loader2 className='w-12 h-12 text-netflix-red animate-spin' />
            </div>
          ) : (
            <div className='grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 sm:gap-6'>
              {items.map((item, index) => (
                <motion.div
                  key={item.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: index * 0.05 }}
                >
                  <ContentCard
                    id={item.id}
                    title={item.title}
                    cover={item.cover}
                    rating={item.rating}
                    year={item.year}
                    type={item.type}
                  />
                </motion.div>
              ))}
            </div>
          )}
        </div>
      </div>
    </main>
  );
}
