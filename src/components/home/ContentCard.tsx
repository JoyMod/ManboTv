'use client';

import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Play, Plus, ThumbsUp, ChevronDown, Check } from 'lucide-react';
import { useRouter } from 'next/navigation';

interface ContentCardProps {
  title: string;
  cover: string;
  rating?: string;
  year?: string;
  type?: 'movie' | 'tv' | 'variety' | 'anime';
  duration?: string;
  episodes?: string;
  overview?: string;
  source?: string;
  id?: string;
  onClick?: () => void;
}

export const ContentCard: React.FC<ContentCardProps> = ({
  title,
  cover,
  rating,
  year,
  type = 'movie',
  duration,
  episodes,
  overview,
  source,
  id,
  onClick,
}) => {
  const router = useRouter();
  const [isHovered, setIsHovered] = useState(false);
  const [isFavorite, setIsFavorite] = useState(false);

  const typeLabels: Record<string, string> = {
    movie: '电影',
    tv: '剧集',
    variety: '综艺',
    anime: '动漫',
  };

  const handleClick = () => {
    if (onClick) {
      onClick();
    } else if (source && id) {
      // 跳转到播放页
      router.push(`/play?source=${source}&id=${id}&title=${encodeURIComponent(title)}&year=${year || ''}`);
    } else {
      // 跳转到搜索页
      router.push(`/search?q=${encodeURIComponent(title)}`);
    }
  };

  const handlePlay = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (source && id) {
      router.push(`/play?source=${source}&id=${id}&title=${encodeURIComponent(title)}&year=${year || ''}`);
    } else {
      router.push(`/search?q=${encodeURIComponent(title)}`);
    }
  };

  const handleAddToList = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      if (!isFavorite) {
        await fetch('/api/favorites', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            source: source || 'douban',
            id: id || title,
            title,
            cover,
            year,
            type,
          }),
        });
        setIsFavorite(true);
      }
    } catch (error) {
      console.error('Add to favorites error:', error);
    }
  };

  return (
    <motion.div
      className="relative flex-none w-full cursor-pointer group"
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      onClick={handleClick}
      whileHover={{ scale: 1.02 }}
      transition={{ duration: 0.2 }}
    >
      {/* 封面图 */}
      <div className="relative aspect-[2/3] rounded-md overflow-hidden bg-gray-800">
        <img
          src={cover}
          alt={title}
          className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-110"
          onError={(e) => {
            (e.target as HTMLImageElement).src = '/placeholder-poster.svg';
          }}
        />
        
        {/* 悬停遮罩 */}
        <div className="absolute inset-0 bg-gradient-to-t from-black/90 via-black/20 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
        
        {/* 类型标签 */}
        <span className="absolute top-2 right-2 px-2 py-0.5 bg-[#E50914] text-white text-xs font-medium rounded">
          {typeLabels[type]}
        </span>

        {/* 播放按钮 - 悬停显示 */}
        <motion.button
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: isHovered ? 1 : 0, scale: isHovered ? 1 : 0.8 }}
          onClick={handlePlay}
          className="absolute inset-0 flex items-center justify-center"
        >
          <div className="w-16 h-16 rounded-full bg-[#E50914] flex items-center justify-center shadow-lg hover:bg-red-600 transition-colors">
            <Play className="w-7 h-7 text-white fill-white ml-1" />
          </div>
        </motion.button>
      </div>

      {/* 信息区域 */}
      <div className="mt-2 space-y-1">
        {/* 标题 */}
        <h3 className="text-white font-medium text-sm truncate group-hover:text-[#E50914] transition-colors">
          {title}
        </h3>
        
        {/* 元信息 */}
        <div className="flex items-center gap-2 text-xs text-gray-400">
          {rating && (
            <span className="text-green-400 font-medium">{rating}</span>
          )}
          {year && <span>{year}</span>}
          {duration && <span>{duration}</span>}
          {episodes && <span>{episodes}</span>}
        </div>

        {/* 悬停展开的操作按钮 */}
        <motion.div
          initial={{ height: 0, opacity: 0 }}
          animate={{ height: isHovered ? 'auto' : 0, opacity: isHovered ? 1 : 0 }}
          className="overflow-hidden"
        >
          <div className="flex items-center gap-2 pt-2">
            <motion.button
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.95 }}
              onClick={handleAddToList}
              className={`w-8 h-8 rounded-full border-2 flex items-center justify-center transition-colors ${
                isFavorite 
                  ? 'bg-[#E50914] border-[#E50914] text-white' 
                  : 'border-gray-500 text-white hover:border-white'
              }`}
            >
              {isFavorite ? <Check className="w-4 h-4" /> : <Plus className="w-4 h-4" />}
            </motion.button>
            <motion.button
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.95 }}
              className="w-8 h-8 rounded-full border-2 border-gray-500 flex items-center justify-center text-white hover:border-white transition-colors"
            >
              <ThumbsUp className="w-4 h-4" />
            </motion.button>
            <motion.button
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.95 }}
              className="w-8 h-8 rounded-full border-2 border-gray-500 flex items-center justify-center text-white hover:border-white transition-colors ml-auto"
            >
              <ChevronDown className="w-4 h-4" />
            </motion.button>
          </div>
        </motion.div>
      </div>

      {/* 简介 - 悬停显示 */}
      {overview && (
        <motion.p
          initial={{ opacity: 0, height: 0 }}
          animate={{ opacity: isHovered ? 1 : 0, height: isHovered ? 'auto' : 0 }}
          className="mt-2 text-xs text-gray-400 line-clamp-2 overflow-hidden"
        >
          {overview}
        </motion.p>
      )}
    </motion.div>
  );
};

export default ContentCard;
