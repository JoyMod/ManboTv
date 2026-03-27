'use client';

import { Clock, Play, Star } from 'lucide-react';
import React from 'react';

import { getAllPlayRecords, PlayRecord } from '@/lib/db.client';
import { toImageSrc } from '@/lib/image';
import { useFastNavigation } from '@/lib/navigation-feedback';

import SmartImage from '@/components/ui/SmartImage';

interface VideoCardProps {
  id?: string;
  source?: string;
  title: string;
  cover: string;
  year?: string;
  rate?: string;
  desc?: string;
  episodes?: string[];
  className?: string;
  size?: 'sm' | 'md' | 'lg';
  showProgress?: boolean;
  index?: number;
  onClick?: () => void;
}

export default function VideoCard({
  id,
  source,
  title,
  cover,
  year,
  rate,
  desc,
  episodes = [],
  className = '',
  size = 'md',
  showProgress = true,
  onClick,
}: VideoCardProps) {
  const { navigate, prefetchHref } = useFastNavigation();
  const [progress, setProgress] = React.useState(0);
  const [playRecord, setPlayRecord] = React.useState<PlayRecord | null>(null);
  const coverSrc = React.useMemo(
    () => toImageSrc(cover, '/placeholder-poster.svg'),
    [cover]
  );
  const targetHref = React.useMemo(() => {
    if (!source || !id) return '';
    return `/play?source=${encodeURIComponent(source)}&id=${encodeURIComponent(
      id
    )}&title=${encodeURIComponent(title)}&year=${encodeURIComponent(year || '')}`;
  }, [id, source, title, year]);

  // 加载播放进度
  React.useEffect(() => {
    if (!showProgress || !source || !id) return;

    const loadProgress = async () => {
      try {
        const records = await getAllPlayRecords();
        const key = `${source}+${id}`;
        const record = records[key];
        if (record && record.total_time > 0) {
          setPlayRecord(record);
          setProgress(
            Math.min(100, Math.round((record.play_time / record.total_time) * 100))
          );
        }
      } catch {
        setPlayRecord(null);
        setProgress(0);
      }
    };

    void loadProgress();
  }, [source, id, showProgress]);

  React.useEffect(() => {
    prefetchHref(targetHref);
  }, [prefetchHref, targetHref]);

  const handleClick = () => {
    if (onClick) {
      onClick();
      return;
    }

    if (targetHref) {
      navigate(targetHref);
    }
  };

  const sizeClasses = {
    sm: 'aspect-[2/3]',
    md: 'aspect-[2/3]',
    lg: 'aspect-video',
  };

  const titleSizeClasses = {
    sm: 'text-xs',
    md: 'text-sm',
    lg: 'text-base',
  };

  return (
    <div
      onClick={handleClick}
      onPointerEnter={() => prefetchHref(targetHref)}
      className={`group relative cursor-pointer overflow-hidden rounded-lg bg-zinc-800 transition-all duration-300 hover:scale-105 hover:shadow-xl ${className}`}
    >
      {/* 封面 */}
      <div className={`relative ${sizeClasses[size]}`}>
        <SmartImage
          src={coverSrc}
          alt={title}
          fill
          sizes={size === 'lg' ? '50vw' : '(max-width: 768px) 50vw, 20vw'}
          className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-110"
        />

        {/* 渐变遮罩 */}
        <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent opacity-60 transition-opacity group-hover:opacity-80" />

        {/* 播放按钮 */}
        <div className="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity group-hover:opacity-100">
          <div className="rounded-full bg-netflix-red p-3 shadow-lg transition-transform hover:scale-110">
            <Play className="h-6 w-6 fill-white text-white" />
          </div>
        </div>

        {/* 评分 */}
        {rate && rate !== '0' && rate !== '0.0' && (
          <div className="absolute right-2 top-2 flex items-center gap-0.5 rounded bg-yellow-500/90 px-1.5 py-0.5 text-xs font-bold text-black">
            <Star className="h-3 w-3 fill-black" />
            {rate}
          </div>
        )}

        {/* 年份 */}
        {year && (
          <div className="absolute left-2 top-2 rounded bg-black/60 px-1.5 py-0.5 text-xs text-white">
            {year}
          </div>
        )}

        {/* 集数 */}
        {episodes.length > 0 && (
          <div className="absolute bottom-2 right-2 rounded bg-netflix-red px-1.5 py-0.5 text-xs font-medium text-white">
            {episodes.length} 集
          </div>
        )}

        {/* 播放进度 */}
        {progress > 0 && (
          <>
            <div className="absolute bottom-0 left-0 right-0 h-1 bg-zinc-700">
              <div
                className="h-full bg-netflix-red"
                style={{ width: `${progress}%` }}
              />
            </div>
            {playRecord && (
              <div className="absolute bottom-2 left-2 flex items-center gap-1 rounded bg-black/60 px-1.5 py-0.5 text-[10px] text-white">
                <Clock className="h-3 w-3" />
                看到第 {playRecord.index} 集
              </div>
            )}
          </>
        )}

        {/* 继续播放标记 */}
        {progress > 0 && progress < 95 && (
          <div className="absolute right-2 bottom-2 rounded bg-netflix-red px-1.5 py-0.5 text-[10px] font-medium text-white">
            继续播放
          </div>
        )}

        {/* 已看完标记 */}
        {progress >= 95 && (
          <div className="absolute right-2 bottom-2 rounded bg-green-500 px-1.5 py-0.5 text-[10px] font-medium text-white">
            已看完
          </div>
        )}
      </div>

      {/* 信息区域 */}
      <div className="p-2">
        <h3
          className={`${titleSizeClasses[size]} truncate font-medium text-white`}
        >
          {title}
        </h3>
        {desc && size === 'lg' && (
          <p className="mt-1 line-clamp-2 text-xs text-zinc-400">{desc}</p>
        )}
      </div>
    </div>
  );
}
