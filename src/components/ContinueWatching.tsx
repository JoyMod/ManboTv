'use client';

import { ChevronRight, Clock, Play, Trash2 } from 'lucide-react';
import React, { useEffect, useState } from 'react';

import {
  deletePlayRecord,
  getAllPlayRecords,
  PlayRecord,
} from '@/lib/db.client';
import { toProxyImageSrc } from '@/lib/image';
import { useFastNavigation } from '@/lib/navigation-feedback';

import SmartImage from '@/components/ui/SmartImage';

interface ContinueWatchingItem {
  key: string;
  record: PlayRecord;
  progress: number;
}

interface ContinueWatchingProps {
  className?: string;
  limit?: number;
  onDataChange?: () => void;
}

function formatDuration(seconds: number): string {
  if (!seconds || seconds <= 0) return '00:00';
  const hours = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  if (hours > 0) {
    return `${hours}小时${mins}分`;
  }
  return `${mins}分钟`;
}

function formatTimeAgo(timestamp: number): string {
  const now = Date.now();
  const diff = now - timestamp;
  const minutes = Math.floor(diff / (1000 * 60));
  const hours = Math.floor(diff / (1000 * 60 * 60));
  const days = Math.floor(diff / (1000 * 60 * 60 * 24));

  if (minutes < 1) return '刚刚';
  if (minutes < 60) return `${minutes}分钟前`;
  if (hours < 24) return `${hours}小时前`;
  if (days < 7) return `${days}天前`;
  return new Date(timestamp).toLocaleDateString('zh-CN');
}

export default function ContinueWatching({
  className = '',
  limit = 6,
  onDataChange,
}: ContinueWatchingProps) {
  const { navigate, prefetchHref } = useFastNavigation();
  const [items, setItems] = useState<ContinueWatchingItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAll, setShowAll] = useState(false);

  const loadRecords = async () => {
    try {
      const records = await getAllPlayRecords();
      const list: ContinueWatchingItem[] = Object.entries(records)
        .map(([key, record]) => ({
          key,
          record,
          progress:
            record.total_time > 0
              ? Math.min(100, Math.round((record.play_time / record.total_time) * 100))
              : 0,
        }))
        .sort((a, b) => b.record.save_time - a.record.save_time);
      setItems(list);
    } catch {
      setItems([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadRecords();

    // 监听播放记录更新事件
    const handleUpdate = () => {
      void loadRecords();
    };

    window.addEventListener('playRecordsUpdated', handleUpdate);
    return () => {
      window.removeEventListener('playRecordsUpdated', handleUpdate);
    };
  }, []);

  const handleDelete = async (key: string, e: React.MouseEvent) => {
    e.stopPropagation();
    const [source, id] = key.split('+');
    if (source && id) {
      try {
        await deletePlayRecord(source, id);
        setItems((prev) => prev.filter((item) => item.key !== key));
        onDataChange?.();
      } catch {
        return;
      }
    }
  };

  const buildPlayHref = (item: ContinueWatchingItem) => {
    const { record, key } = item;
    const [source, id] = key.split('+');
    if (!source || !id) return '';
    return `/play?source=${encodeURIComponent(source)}&id=${encodeURIComponent(
      id
    )}&title=${encodeURIComponent(record.title)}&ep=${encodeURIComponent(record.index)}`;
  };

  const handleClick = (item: ContinueWatchingItem) => {
    const href = buildPlayHref(item);
    if (href) navigate(href);
  };

  const displayedItems = showAll ? items : items.slice(0, limit);

  if (loading) {
    return (
      <div className={`${className}`}>
        <h2 className="mb-4 text-xl font-bold text-white">继续观看</h2>
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
          {Array.from({ length: limit }).map((_, i) => (
            <div
              key={i}
              className="aspect-[2/3] animate-pulse rounded-lg bg-zinc-800"
            />
          ))}
        </div>
      </div>
    );
  }

  if (items.length === 0) {
    return null;
  }

  return (
    <div className={`${className}`}>
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-xl font-bold text-white">继续观看</h2>
        {items.length > limit && (
          <button
            onClick={() => setShowAll(!showAll)}
            className="flex items-center gap-1 text-sm text-netflix-red hover:underline"
          >
            {showAll ? '收起' : '查看全部'}
            <ChevronRight
              className={`h-4 w-4 transition-transform ${showAll ? 'rotate-90' : ''}`}
            />
          </button>
        )}
      </div>

      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
        {displayedItems.map((item) => {
          const { record, key, progress } = item;
          return (
            <div
              key={key}
              onClick={() => handleClick(item)}
              onPointerEnter={() => prefetchHref(buildPlayHref(item))}
              className="group relative cursor-pointer overflow-hidden rounded-lg bg-zinc-800 transition-transform hover:scale-105"
            >
              {/* 封面 */}
              <div className="relative aspect-[2/3]">
                <SmartImage
                  src={toProxyImageSrc(record.cover)}
                  alt={record.title}
                  fill
                  sizes="(max-width: 768px) 50vw, 16vw"
                  className="h-full w-full object-cover"
                />
                {/* 进度条 */}
                <div className="absolute bottom-0 left-0 right-0 h-1 bg-zinc-700">
                  <div
                    className="h-full bg-netflix-red"
                    style={{ width: `${progress}%` }}
                  />
                </div>
                {/* 播放按钮 */}
                <div className="absolute inset-0 flex items-center justify-center bg-black/40 opacity-0 transition-opacity group-hover:opacity-100">
                  <div className="rounded-full bg-netflix-red p-3">
                    <Play className="h-6 w-6 fill-white text-white" />
                  </div>
                </div>
                {/* 删除按钮 */}
                <button
                  onClick={(e) => handleDelete(key, e)}
                  className="absolute right-2 top-2 rounded-full bg-black/60 p-1.5 text-white opacity-0 transition-opacity hover:bg-red-600 group-hover:opacity-100"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
                {/* 集数标记 */}
                <div className="absolute left-2 top-2 rounded bg-black/60 px-1.5 py-0.5 text-xs text-white">
                  第 {record.index} 集
                </div>
              </div>

              {/* 信息 */}
              <div className="p-2">
                <h3 className="truncate text-sm font-medium text-white">
                  {record.title}
                </h3>
                <div className="mt-1 flex items-center gap-2 text-xs text-zinc-400">
                  <span className="flex items-center gap-0.5">
                    <Clock className="h-3 w-3" />
                    {formatTimeAgo(record.save_time)}
                  </span>
                </div>
                <div className="mt-1 text-xs text-zinc-500">
                  观看到 {formatDuration(record.play_time)}
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
