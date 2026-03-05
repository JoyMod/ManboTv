'use client';

import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Play, Plus, Info, Volume2, VolumeX } from 'lucide-react';

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

interface HeroBannerProps {
  items: BannerItem[];
  autoPlayInterval?: number;
}

export const HeroBanner: React.FC<HeroBannerProps> = ({
  items,
  autoPlayInterval = 8000,
}) => {
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isMuted, setIsMuted] = useState(true);
  const [isLoading, setIsLoading] = useState(true);

  const currentItem = items[currentIndex];

  useEffect(() => {
    if (items.length <= 1) return;

    const interval = setInterval(() => {
      setCurrentIndex((prev) => (prev + 1) % items.length);
    }, autoPlayInterval);

    return () => clearInterval(interval);
  }, [items.length, autoPlayInterval]);

  const handlePrev = () => {
    setCurrentIndex((prev) => (prev - 1 + items.length) % items.length);
  };

  const handleNext = () => {
    setCurrentIndex((prev) => (prev + 1) % items.length);
  };

  return (
    <div className="relative w-full h-[70vh] min-h-[500px] max-h-[900px] overflow-hidden">
      {/* 背景图 */}
      <AnimatePresence mode="wait">
        <motion.div
          key={currentItem.id}
          initial={{ opacity: 0, scale: 1.1 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.8 }}
          className="absolute inset-0"
        >
          <img
            src={currentItem.backdrop}
            alt={currentItem.title}
            className="w-full h-full object-cover"
            onLoad={() => setIsLoading(false)}
            onError={(e) => {
              (e.target as HTMLImageElement).src = '/placeholder-backdrop.svg';
            }}
          />
          {/* 渐变遮罩 */}
          <div className="absolute inset-0 bg-gradient-to-r from-black/80 via-black/40 to-transparent" />
          <div className="absolute inset-0 bg-gradient-to-t from-[#141414] via-transparent to-black/30" />
        </motion.div>
      </AnimatePresence>

      {/* 内容 */}
      <div className="absolute inset-0 flex items-center">
        <div className="max-w-[1920px] mx-auto px-4 sm:px-8 lg:px-16 w-full">
          <AnimatePresence mode="wait">
            <motion.div
              key={currentItem.id}
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              transition={{ duration: 0.6, delay: 0.2 }}
              className="max-w-2xl"
            >
              {/* 标签 */}
              {currentItem.tags && (
                <div className="flex flex-wrap gap-2 mb-4">
                  {currentItem.tags.map((tag) => (
                    <span
                      key={tag}
                      className="px-3 py-1 bg-[#E50914] text-white text-xs font-bold uppercase tracking-wider rounded"
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              )}

              {/* Logo/标题 */}
              {currentItem.logo ? (
                <img
                  src={currentItem.logo}
                  alt={currentItem.title}
                  className="w-auto h-24 sm:h-32 md:h-40 object-contain mb-6"
                />
              ) : (
                <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-black text-white mb-4 drop-shadow-2xl">
                  {currentItem.title}
                </h1>
              )}

              {/* 副标题 */}
              {currentItem.subtitle && (
                <p className="text-xl sm:text-2xl text-gray-300 mb-4">
                  {currentItem.subtitle}
                </p>
              )}

              {/* 元信息 */}
              <div className="flex items-center gap-4 mb-4 text-sm text-gray-300">
                {currentItem.rating && (
                  <span className="text-green-400 font-bold">{currentItem.rating}</span>
                )}
                {currentItem.year && <span>{currentItem.year}</span>}
                {currentItem.duration && <span>{currentItem.duration}</span>}
              </div>

              {/* 描述 */}
              <p className="text-base sm:text-lg text-gray-200 line-clamp-3 mb-8 drop-shadow-lg">
                {currentItem.description}
              </p>

              {/* 按钮组 */}
              <div className="flex items-center gap-4">
                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.98 }}
                  className="flex items-center gap-2 px-6 sm:px-8 py-3 bg-white text-black font-bold rounded hover:bg-gray-200 transition-colors"
                >
                  <Play className="w-5 h-5 fill-black" />
                  立即播放
                </motion.button>

                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.98 }}
                  className="flex items-center gap-2 px-6 sm:px-8 py-3 bg-gray-500/70 backdrop-blur-sm text-white font-bold rounded hover:bg-gray-500/50 transition-colors"
                >
                  <Plus className="w-5 h-5" />
                  加入片单
                </motion.button>

                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.98 }}
                  className="hidden sm:flex items-center gap-2 px-6 py-3 bg-gray-500/70 backdrop-blur-sm text-white font-bold rounded hover:bg-gray-500/50 transition-colors"
                >
                  <Info className="w-5 h-5" />
                  更多信息
                </motion.button>
              </div>
            </motion.div>
          </AnimatePresence>
        </div>
      </div>

      {/* 音量控制 */}
      <button
        onClick={() => setIsMuted(!isMuted)}
        className="absolute bottom-32 right-8 sm:right-16 w-12 h-12 rounded-full border-2 border-gray-500 flex items-center justify-center hover:border-white hover:bg-white/10 transition-all z-20"
      >
        {isMuted ? (
          <VolumeX className="w-5 h-5 text-white" />
        ) : (
          <Volume2 className="w-5 h-5 text-white" />
        )}
      </button>

      {/* 底部指示器 */}
      {items.length > 1 && (
        <div className="absolute bottom-20 right-8 sm:right-16 flex gap-2 z-20">
          {items.map((_, index) => (
            <button
              key={index}
              onClick={() => setCurrentIndex(index)}
              className={`h-1 rounded-full transition-all duration-300 ${
                index === currentIndex
                  ? 'w-8 bg-[#E50914]'
                  : 'w-4 bg-gray-500 hover:bg-gray-400'
              }`}
            />
          ))}
        </div>
      )}

      {/* 侧边导航 */}
      {items.length > 1 && (
        <>
          <button
            onClick={handlePrev}
            className="absolute left-4 top-1/2 -translate-y-1/2 w-12 h-12 rounded-full bg-black/30 flex items-center justify-center hover:bg-black/50 transition-colors z-20"
          >
            <span className="sr-only">上一个</span>
            <svg className="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <button
            onClick={handleNext}
            className="absolute right-4 top-1/2 -translate-y-1/2 w-12 h-12 rounded-full bg-black/30 flex items-center justify-center hover:bg-black/50 transition-colors z-20"
          >
            <span className="sr-only">下一个</span>
            <svg className="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </button>
        </>
      )}
    </div>
  );
};

export default HeroBanner;
