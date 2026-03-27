'use client';

import { ChevronLeft, ChevronRight } from 'lucide-react';
import Link from 'next/link';
import React, { useRef, useState } from 'react';

import ContentCard from '@/components/ui/ContentCard';

interface ContentItem {
  id: string;
  title: string;
  cover: string;
  rating?: string;
  year?: string;
  type?: 'movie' | 'tv' | 'variety' | 'anime';
  source?: string;
}

interface ContentRowProps {
  title: string;
  items: ContentItem[];
  showRanking?: boolean;
  description?: string;
  href?: string;
}

export const ContentRow: React.FC<ContentRowProps> = ({
  title,
  items,
  showRanking = false,
  description,
  href,
}) => {
  const ref = useRef<HTMLDivElement>(null);
  const [left, setLeft] = useState(false);
  const [right, setRight] = useState(true);

  const check = () => {
    if (!ref.current) return;
    const { scrollLeft, scrollWidth, clientWidth } = ref.current;
    setLeft(scrollLeft > 0);
    setRight(scrollLeft + clientWidth < scrollWidth - 8);
  };

  const slide = (dir: 'l' | 'r') => {
    if (!ref.current) return;
    ref.current.scrollBy({
      left: dir === 'l' ? -900 : 900,
      behavior: 'smooth',
    });
    window.setTimeout(check, 260);
  };

  return (
    <section className='group relative py-3'>
      <div className='mb-3 flex items-center justify-between px-4 sm:px-8'>
        <div>
          <h2 className='text-lg font-bold text-white sm:text-2xl'>{title}</h2>
          {description ? (
            <p className='mt-1 text-sm text-zinc-400'>{description}</p>
          ) : null}
        </div>
        {href ? (
          <Link
            href={href}
            className='hidden rounded-full border border-zinc-700 px-4 py-2 text-sm text-zinc-300 transition-colors hover:border-zinc-500 hover:text-white sm:inline-flex'
          >
            查看更多
          </Link>
        ) : null}
      </div>

      <div className='relative'>
        <button
          onClick={() => slide('l')}
          className={`absolute left-0 top-0 z-20 h-full w-12 bg-black/55 text-white transition-opacity ${
            left ? 'opacity-100' : 'pointer-events-none opacity-0'
          }`}
        >
          <ChevronLeft className='mx-auto h-6 w-6' />
        </button>
        <button
          onClick={() => slide('r')}
          className={`absolute right-0 top-0 z-20 h-full w-12 bg-black/55 text-white transition-opacity ${
            right ? 'opacity-100' : 'pointer-events-none opacity-0'
          }`}
        >
          <ChevronRight className='mx-auto h-6 w-6' />
        </button>

        <div
          ref={ref}
          onScroll={check}
          className='scrollbar-hide flex gap-3 overflow-x-auto px-4 pb-4 sm:px-8'
        >
          {items.map((item, idx) => (
            <div
              key={`${item.id}-${idx}`}
              className='w-[140px] flex-none sm:w-[180px]'
            >
              {showRanking && idx < 10 && (
                <div className='mb-1 text-xs font-semibold text-red-500'>
                  TOP {idx + 1}
                </div>
              )}
              <ContentCard
                id={item.id}
                source={item.source}
                title={item.title}
                cover={item.cover}
                rating={item.rating}
                year={item.year}
                type={item.type}
              />
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default ContentRow;
