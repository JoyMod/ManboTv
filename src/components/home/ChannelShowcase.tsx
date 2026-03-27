'use client';

import { ChevronRight, Clapperboard, Film, Flame, Sparkles, Tv } from 'lucide-react';
import Link from 'next/link';
import React from 'react';

interface ChannelItem {
  title: string;
  subtitle: string;
  href: string;
  accent: string;
  icon: React.ReactNode;
  tags: string[];
}

const channelItems: ChannelItem[] = [
  {
    title: '电影片库',
    subtitle: '新片热映、高分佳作、经典重温',
    href: '/movie',
    accent: 'from-red-500/30 via-red-500/10 to-transparent',
    icon: <Film className='h-5 w-5' />,
    tags: ['高分', '热映', '经典'],
  },
  {
    title: '剧集频道',
    subtitle: '国产剧、美剧、韩剧、日剧快速进入',
    href: '/tv',
    accent: 'from-sky-500/30 via-sky-500/10 to-transparent',
    icon: <Tv className='h-5 w-5' />,
    tags: ['国产剧', '悬疑', '都市'],
  },
  {
    title: '综艺精选',
    subtitle: '真人秀、脱口秀、音乐现场持续更新',
    href: '/variety',
    accent: 'from-amber-500/30 via-amber-500/10 to-transparent',
    icon: <Sparkles className='h-5 w-5' />,
    tags: ['真人秀', '脱口秀', '音乐'],
  },
  {
    title: '动漫专区',
    subtitle: '热血新番、剧场版、国漫精选',
    href: '/anime',
    accent: 'from-emerald-500/30 via-emerald-500/10 to-transparent',
    icon: <Clapperboard className='h-5 w-5' />,
    tags: ['新番', '热血', '治愈'],
  },
  {
    title: '热播榜单',
    subtitle: '适合快速找片，直接看当下最热内容',
    href: '/hot',
    accent: 'from-fuchsia-500/30 via-fuchsia-500/10 to-transparent',
    icon: <Flame className='h-5 w-5' />,
    tags: ['热播', '追更', '趋势'],
  },
];

export default function ChannelShowcase() {
  return (
    <section className='px-4 sm:px-8'>
      <div className='mb-5 flex items-end justify-between'>
        <div>
          <p className='text-xs uppercase tracking-[0.28em] text-zinc-500'>
            Browse Faster
          </p>
          <h2 className='mt-2 text-2xl font-bold text-white'>频道入口</h2>
          <p className='mt-1 text-sm text-zinc-400'>
            先进入频道，再做筛选，路径更短，信息也更稳定。
          </p>
        </div>
      </div>

      <div className='grid gap-4 md:grid-cols-2 xl:grid-cols-5'>
        {channelItems.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className='group relative overflow-hidden rounded-3xl border border-white/10 bg-zinc-950/80 p-5 transition-all duration-300 hover:-translate-y-1 hover:border-white/20'
          >
            <div
              className={`absolute inset-0 bg-gradient-to-br ${item.accent} opacity-90 transition-opacity group-hover:opacity-100`}
            />
            <div className='absolute inset-0 bg-[radial-gradient(circle_at_top_right,rgba(255,255,255,0.12),transparent_40%)]' />

            <div className='relative z-10 flex h-full flex-col'>
              <div className='flex items-center justify-between'>
                <div className='inline-flex rounded-2xl border border-white/10 bg-white/5 p-3 text-white'>
                  {item.icon}
                </div>
                <ChevronRight className='h-4 w-4 text-zinc-500 transition-transform group-hover:translate-x-1 group-hover:text-white' />
              </div>

              <div className='mt-8'>
                <h3 className='text-lg font-semibold text-white'>{item.title}</h3>
                <p className='mt-2 text-sm leading-6 text-zinc-300'>
                  {item.subtitle}
                </p>
              </div>

              <div className='mt-5 flex flex-wrap gap-2'>
                {item.tags.map((tag) => (
                  <span
                    key={tag}
                    className='rounded-full border border-white/10 bg-black/20 px-2.5 py-1 text-xs text-zinc-200'
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          </Link>
        ))}
      </div>
    </section>
  );
}
