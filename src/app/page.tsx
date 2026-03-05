'use client';

import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import TopNav from '@/components/layout/TopNav';
import HeroBanner from '@/components/home/HeroBanner';
import ContentRow from '@/components/home/ContentRow';
import { Loader2 } from 'lucide-react';

interface BannerItem {
  id: string;
  title: string;
  subtitle?: string;
  description: string;
  backdrop: string;
  logo?: string;
  rating?: string;
  year?: string;
  duration?: string;
  tags?: string[];
}

interface ContentItem {
  id: string;
  title: string;
  cover: string;
  rating?: string;
  year?: string;
  type?: 'movie' | 'tv' | 'variety' | 'anime';
  duration?: string;
  episodes?: string;
  overview?: string;
}

interface Section {
  id: string;
  title: string;
  items: ContentItem[];
  showRanking?: boolean;
}

export default function HomePage() {
  const [banner, setBanner] = useState<BannerItem[]>([]);
  const [sections, setSections] = useState<Section[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchHomeData = async () => {
      try {
        const response = await fetch('/api/home');
        if (!response.ok) throw new Error('获取首页数据失败');
        const data = await response.json();
        setBanner(data.banner || []);
        setSections(data.sections || []);
      } catch (err) {
        console.error('Fetch home data error:', err);
        setError('加载数据失败，请稍后重试');
      } finally {
        setLoading(false);
      }
    };

    fetchHomeData();
  }, []);

  if (loading) {
    return (
      <main className='min-h-screen bg-[#141414] flex items-center justify-center'>
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className='flex flex-col items-center gap-4'
        >
          <Loader2 className='w-12 h-12 text-[#E50914] animate-spin' />
          <p className='text-gray-400'>正在加载精彩内容...</p>
        </motion.div>
      </main>
    );
  }

  return (
    <main className='min-h-screen bg-[#141414]'>
      {/* 顶部导航 */}
      <TopNav />

      {/* Hero Banner */}
      {banner.length > 0 && <HeroBanner items={banner} />}

      {/* 内容区域 */}
      <div className='relative z-10 -mt-20 pb-20 space-y-2'>
        {error && (
          <div className='px-4 sm:px-8 py-4'>
            <div className='bg-red-500/10 border border-red-500/20 rounded-lg p-4 text-red-400 text-center'>
              {error}
            </div>
          </div>
        )}

        {sections.map((section) => (
          <ContentRow
            key={section.id}
            title={section.title}
            items={section.items}
            showRanking={section.showRanking}
          />
        ))}

        {sections.length === 0 && !error && (
          <div className='px-4 sm:px-8 py-20 text-center'>
            <p className='text-gray-400'>暂无内容</p>
          </div>
        )}
      </div>
    </main>
  );
}
