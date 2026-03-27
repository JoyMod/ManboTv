'use client';

import React, { useEffect, useState } from 'react';

import ContinueWatching from '@/components/ContinueWatching';
import ChannelShowcase from '@/components/home/ChannelShowcase';
import ContentRow from '@/components/home/ContentRow';
import HeroBanner from '@/components/home/HeroBanner';
import TopNav from '@/components/layout/TopNav';

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
  source?: string;
}

interface Section {
  id: string;
  title: string;
  description?: string;
  href?: string;
  items: ContentItem[];
  showRanking?: boolean;
}

const SkeletonRowCount = 3;
const SkeletonCardsPerRow = 6;

function HeroBannerSkeleton() {
  return (
    <section className='relative h-[72vh] min-h-[520px] w-full overflow-hidden bg-[#101010]'>
      <div className='absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(229,9,20,0.25),transparent_42%)]' />
      <div className='absolute inset-0 bg-gradient-to-t from-[#141414] via-[#141414]/40 to-black/30' />
      <div className='relative z-10 mx-auto flex h-full max-w-[1920px] items-end px-4 pb-24 sm:px-8'>
        <div className='w-full max-w-2xl animate-pulse space-y-4'>
          <div className='h-5 w-28 rounded-full bg-white/10' />
          <div className='h-12 w-3/4 rounded-2xl bg-white/10 sm:h-16' />
          <div className='h-5 w-2/3 rounded-full bg-white/10' />
          <div className='space-y-2'>
            <div className='h-4 w-full rounded-full bg-white/10' />
            <div className='h-4 w-11/12 rounded-full bg-white/10' />
            <div className='h-4 w-4/5 rounded-full bg-white/10' />
          </div>
          <div className='flex gap-3 pt-4'>
            <div className='h-11 w-32 rounded bg-white/10' />
            <div className='h-11 w-36 rounded bg-white/10' />
            <div className='h-11 w-11 rounded-full bg-white/10' />
          </div>
        </div>
      </div>
    </section>
  );
}

function HomeRowsSkeleton() {
  return (
    <>
      {Array.from({ length: SkeletonRowCount }).map((_, rowIndex) => (
        <section key={`home-skeleton-${rowIndex}`} className='py-3'>
          <div className='mb-3 px-4 sm:px-8'>
            <div className='h-7 w-32 animate-pulse rounded-full bg-zinc-800' />
            <div className='mt-2 h-4 w-72 animate-pulse rounded-full bg-zinc-900' />
          </div>
          <div className='flex gap-3 overflow-hidden px-4 pb-4 sm:px-8'>
            {Array.from({ length: SkeletonCardsPerRow }).map((_, cardIndex) => (
              <div
                key={`home-skeleton-card-${rowIndex}-${cardIndex}`}
                className='w-[140px] flex-none sm:w-[180px]'
              >
                <div className='aspect-[2/3] animate-pulse rounded-2xl bg-zinc-900' />
                <div className='mt-3 h-4 w-5/6 animate-pulse rounded-full bg-zinc-900' />
                <div className='mt-2 h-3 w-2/5 animate-pulse rounded-full bg-zinc-950' />
              </div>
            ))}
          </div>
        </section>
      ))}
    </>
  );
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

        let nextBanner: BannerItem[] = Array.isArray(data?.banner) ? data.banner : [];
        const nextSections: Section[] = Array.isArray(data?.sections) ? data.sections : [];

        if (nextBanner.length === 0) {
          nextBanner = nextSections
            .flatMap((section) => section.items.slice(0, 1))
            .map((item, index) => ({
              id: item.id || `featured-${index}`,
              title: item.title,
              subtitle: index === 0 ? '今日推荐' : '频道精选',
              description: `${item.title} 已进入首页精选分区，适合直接开始浏览与播放。`,
              backdrop: item.cover,
              rating: item.rating,
              year: item.year,
              tags: ['推荐', item.type || '影视'],
            }));
        }

        setBanner(nextBanner);
        setSections(nextSections);
      } catch {
        setError('加载数据失败，请稍后重试');
      } finally {
        setLoading(false);
      }
    };

    fetchHomeData();
  }, []);

  return (
    <main className='min-h-screen bg-[#141414]'>
      {/* 顶部导航 */}
      <TopNav />

      {/* Hero Banner */}
      {banner.length > 0 ? <HeroBanner items={banner} /> : <HeroBannerSkeleton />}

      {/* 内容区域 */}
      <div className='relative z-10 -mt-12 space-y-10 pb-24'>
        <ChannelShowcase />

        {/* 继续观看 */}
        <div className='px-4 sm:px-8'>
          <ContinueWatching />
        </div>

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
            description={section.description}
            href={section.href}
            items={section.items}
            showRanking={section.showRanking}
          />
        ))}

        {loading && sections.length === 0 ? <HomeRowsSkeleton /> : null}

        {sections.length === 0 && !loading && !error && (
          <div className='px-4 sm:px-8 py-20 text-center'>
            <p className='text-gray-400'>暂无内容</p>
          </div>
        )}
      </div>
    </main>
  );
}
