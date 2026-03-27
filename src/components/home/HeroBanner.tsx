'use client';

import { motion } from 'framer-motion';
import { Info, Play, Plus } from 'lucide-react';
import React, { useEffect, useMemo, useState } from 'react';

import { toImageSrc, toLogoProxyImageSrc, toProxyImageSrc } from '@/lib/image';
import { useFastNavigation } from '@/lib/navigation-feedback';

import SmartImage from '@/components/ui/SmartImage';

export interface BannerItem {
  id: string;
  title: string;
  subtitle?: string;
  description: string;
  backdrop: string;
  rating?: string;
  year?: string;
  tags?: string[];
}

interface HeroBannerProps {
  items: BannerItem[];
  autoPlayInterval?: number;
}

export const HeroBanner: React.FC<HeroBannerProps> = ({
  items,
  autoPlayInterval = 8000,
}) => {
  const { navigate, prefetchHref } = useFastNavigation();
  const [idx, setIdx] = useState(0);
  const [proxyStage, setProxyStage] = useState(0);

  useEffect(() => {
    if (items.length <= 1) return;
    const timer = window.setInterval(() => {
      setIdx((prev) => (prev + 1) % items.length);
    }, autoPlayInterval);
    return () => window.clearInterval(timer);
  }, [items.length, autoPlayInterval]);

  useEffect(() => {
    setProxyStage(0);
  }, [idx]);

  const current = items[idx];
  const searchHref = useMemo(
    () => `/search?q=${encodeURIComponent(current?.title || '')}`,
    [current?.title]
  );
  const backdrop = useMemo(
    () =>
      proxyStage === 0
        ? toImageSrc(current?.backdrop, '/placeholder-backdrop.svg')
        : proxyStage === 1
        ? toProxyImageSrc(current?.backdrop, '/placeholder-backdrop.svg')
        : toLogoProxyImageSrc(current?.backdrop, '/placeholder-backdrop.svg'),
    [current, proxyStage]
  );

  useEffect(() => {
    if (!current) return;
    prefetchHref(searchHref);
  }, [current, prefetchHref, searchHref]);

  if (!current) return null;

  return (
    <section className='relative h-[72vh] min-h-[520px] w-full overflow-hidden'>
      <SmartImage
        key={`${current.id}-${proxyStage}`}
        src={backdrop}
        alt={current.title}
        fill
        sizes='100vw'
        className='absolute inset-0 h-full w-full object-cover'
        priority
        onError={() => {
          if (proxyStage < 2) {
            setProxyStage((prev) => prev + 1);
          }
        }}
      />

      <div className='absolute inset-0 bg-gradient-to-r from-black via-black/65 to-transparent' />
      <div className='absolute inset-0 bg-gradient-to-t from-[#141414] via-transparent to-transparent' />

      <div className='relative z-10 mx-auto flex h-full max-w-[1920px] items-end px-4 pb-24 sm:px-8'>
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          className='max-w-2xl'
        >
          <h1 className='text-4xl font-extrabold leading-tight text-white sm:text-6xl'>
            {current.title}
          </h1>
          {current.subtitle ? (
            <p className='mt-2 text-lg text-zinc-300'>{current.subtitle}</p>
          ) : null}

          <div className='mt-3 flex items-center gap-3 text-sm text-zinc-300'>
            {current.rating ? (
              <span className='text-green-400'>★ {current.rating}</span>
            ) : null}
            {current.year ? <span>{current.year}</span> : null}
            {Array.isArray(current.tags) && current.tags.length > 0 ? (
              <span>{current.tags.slice(0, 3).join(' · ')}</span>
            ) : null}
          </div>

          <p className='mt-4 line-clamp-3 text-sm text-zinc-200 sm:text-base'>
            {current.description}
          </p>

          <div className='mt-6 flex flex-wrap items-center gap-3'>
            <button
              onClick={() => navigate(searchHref)}
              onPointerEnter={() => prefetchHref(searchHref)}
              className='inline-flex items-center gap-2 rounded bg-white px-6 py-3 font-semibold text-black'
            >
              <Play className='h-5 w-5 fill-black' />
              播放
            </button>
            <button
              onClick={() => navigate(searchHref)}
              onPointerEnter={() => prefetchHref(searchHref)}
              className='inline-flex items-center gap-2 rounded bg-zinc-500/50 px-6 py-3 font-semibold text-white backdrop-blur'
            >
              <Info className='h-5 w-5' />
              更多信息
            </button>
            <button
              onClick={() => navigate(searchHref)}
              onPointerEnter={() => prefetchHref(searchHref)}
              title='打开搜索结果并选择资源后加入片单'
              aria-label={`查看 ${current.title} 并加入片单`}
              className='inline-flex h-11 w-11 items-center justify-center rounded-full border border-white/50 text-white transition-colors hover:bg-white/10'
            >
              <Plus className='h-5 w-5' />
            </button>
          </div>

          <div className='mt-6 flex gap-1'>
            {items.map((_, i) => (
              <button
                key={i}
                onClick={() => setIdx(i)}
                className={`h-1.5 rounded-full transition-all ${
                  i === idx ? 'w-9 bg-red-600' : 'w-5 bg-zinc-500/70'
                }`}
              />
            ))}
          </div>
        </motion.div>
      </div>
    </section>
  );
};

export default HeroBanner;
