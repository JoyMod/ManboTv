'use client';

import React, { useRef, useState } from 'react';
import { motion } from 'framer-motion';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import ContentCard from './ContentCard';

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

interface ContentRowProps {
  title: string;
  items: ContentItem[];
  showRanking?: boolean;
}

export const ContentRow: React.FC<ContentRowProps> = ({
  title,
  items,
  showRanking = false,
}) => {
  const rowRef = useRef<HTMLDivElement>(null);
  const [canScrollLeft, setCanScrollLeft] = useState(false);
  const [canScrollRight, setCanScrollRight] = useState(true);

  const checkScrollButtons = () => {
    if (rowRef.current) {
      const { scrollLeft, scrollWidth, clientWidth } = rowRef.current;
      setCanScrollLeft(scrollLeft > 0);
      setCanScrollRight(scrollLeft < scrollWidth - clientWidth - 10);
    }
  };

  const scroll = (direction: 'left' | 'right') => {
    if (rowRef.current) {
      const scrollAmount = direction === 'left' ? -800 : 800;
      rowRef.current.scrollBy({ left: scrollAmount, behavior: 'smooth' });
      setTimeout(checkScrollButtons, 400);
    }
  };

  return (
    <div className="group relative py-4">
      {/* 标题 */}
      <div className="flex items-center justify-between px-4 sm:px-8 mb-4">
        <h2 className="text-lg sm:text-xl md:text-2xl font-bold text-white flex items-center gap-2">
          {title}
          <ChevronRight className="w-5 h-5 text-gray-500 opacity-0 group-hover:opacity-100 transition-opacity" />
        </h2>
        <span className="text-sm text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity">
          查看全部
        </span>
      </div>

      {/* 滚动容器 */}
      <div className="relative">
        {/* 左箭头 */}
        <motion.button
          initial={{ opacity: 0 }}
          animate={{ opacity: canScrollLeft ? 1 : 0 }}
          whileHover={{ scale: 1.1 }}
          onClick={() => scroll('left')}
          className={`absolute left-0 top-0 bottom-0 z-40 w-12 sm:w-16 bg-black/50 backdrop-blur-sm flex items-center justify-center transition-opacity ${
            canScrollLeft ? 'cursor-pointer' : 'pointer-events-none'
          }`}
        >
          <ChevronLeft className="w-8 h-8 text-white" />
        </motion.button>

        {/* 右箭头 */}
        <motion.button
          initial={{ opacity: 0 }}
          animate={{ opacity: canScrollRight ? 1 : 0 }}
          whileHover={{ scale: 1.1 }}
          onClick={() => scroll('right')}
          className={`absolute right-0 top-0 bottom-0 z-40 w-12 sm:w-16 bg-black/50 backdrop-blur-sm flex items-center justify-center transition-opacity ${
            canScrollRight ? 'cursor-pointer' : 'pointer-events-none'
          }`}
        >
          <ChevronRight className="w-8 h-8 text-white" />
        </motion.button>

        {/* 卡片列表 */}
        <div
          ref={rowRef}
          onScroll={checkScrollButtons}
          className="flex gap-2 sm:gap-4 overflow-x-auto scrollbar-hide px-4 sm:px-8 pb-8"
          style={{ scrollbarWidth: 'none', msOverflowStyle: 'none' }}
        >
          {items.map((item, index) => (
            <div key={item.id} className="flex items-center">
              {/* 排名数字 */}
              {showRanking && index < 10 && (
                <span
                  className="text-6xl sm:text-7xl md:text-8xl font-black text-gray-800 mr-2 sm:mr-4 italic"
                  style={{
                    WebkitTextStroke: '2px #4a4a4a',
                    textShadow: '2px 2px 4px rgba(0,0,0,0.5)',
                  }}
                >
                  {index + 1}
                </span>
              )}
              <ContentCard
                title={item.title}
                cover={item.cover}
                rating={item.rating}
                year={item.year}
                type={item.type}
                duration={item.duration}
                episodes={item.episodes}
                overview={item.overview}
              />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default ContentRow;
