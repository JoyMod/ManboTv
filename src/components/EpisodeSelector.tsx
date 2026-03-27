'use client';

import { Check, ChevronLeft, ChevronRight, Play } from 'lucide-react';
import React, { useMemo, useState } from 'react';

interface EpisodeSelectorProps {
  episodes: string[];
  episodeTitles?: string[];
  activeIndex: number;
  watchedEpisodes?: Set<number>;
  onEpisodeChange: (index: number) => void;
  onPrevious?: () => void;
  onNext?: () => void;
  className?: string;
  collapsed?: boolean;
  onToggleCollapse?: () => void;
  sourceTests?: Record<string, { quality?: string; speed?: string; status?: string }>;
  currentSourceKey?: string;
}

const EpisodesPerPage = 30;
const GridCols = 5;

export default function EpisodeSelector({
  episodes,
  episodeTitles = [],
  activeIndex,
  watchedEpisodes = new Set(),
  onEpisodeChange,
  onPrevious,
  onNext,
  className = '',
  collapsed = false,
  onToggleCollapse,
}: EpisodeSelectorProps) {
  const [currentPage, setCurrentPage] = useState(0);

  const totalPages = useMemo(
    () => Math.ceil(episodes.length / EpisodesPerPage),
    [episodes.length]
  );

  const currentPageEpisodes = useMemo(() => {
    const start = currentPage * EpisodesPerPage;
    const end = Math.min(start + EpisodesPerPage, episodes.length);
    return episodes.slice(start, end).map((url, idx) => ({
      url,
      index: start + idx,
      title: episodeTitles[start + idx] || `第 ${start + idx + 1} 集`,
    }));
  }, [episodes, episodeTitles, currentPage]);

  const goToPage = (page: number) => {
    if (page >= 0 && page < totalPages) {
      setCurrentPage(page);
    }
  };

  const goToActivePage = () => {
    const activePage = Math.floor(activeIndex / EpisodesPerPage);
    setCurrentPage(activePage);
  };

  // 当 activeIndex 变化时，确保它在当前可视范围内
  React.useEffect(() => {
    const activePage = Math.floor(activeIndex / EpisodesPerPage);
    if (activePage !== currentPage) {
      setCurrentPage(activePage);
    }
  }, [activeIndex, currentPage]);

  if (episodes.length === 0) {
    return (
      <div className={`rounded-lg border border-zinc-800 bg-zinc-900/60 p-4 ${className}`}>
        <p className="text-sm text-zinc-500">暂无选集数据</p>
      </div>
    );
  }

  if (collapsed) {
    return (
      <div
        className={`rounded-lg border border-zinc-800 bg-zinc-900/60 ${className}`}
      >
        <button
          onClick={onToggleCollapse}
          className="flex w-full items-center justify-between p-3 text-left"
        >
          <div>
            <h3 className="font-semibold text-white">选集</h3>
            <p className="text-xs text-zinc-400">
              共 {episodes.length} 集 · 当前第 {activeIndex + 1} 集
            </p>
          </div>
          <ChevronLeft className="h-5 w-5 text-zinc-400" />
        </button>
      </div>
    );
  }

  return (
    <div
      className={`rounded-lg border border-zinc-800 bg-zinc-900/60 ${className}`}
    >
      {/* 头部 */}
      <div className="flex items-center justify-between border-b border-zinc-800 p-3">
        <div className="flex items-center gap-3">
          <h3 className="font-semibold text-white">选集</h3>
          <span className="text-xs text-zinc-400">
            共 {episodes.length} 集
          </span>
        </div>
        <div className="flex items-center gap-2">
          {/* 上一集/下一集按钮 */}
          <button
            onClick={onPrevious}
            disabled={activeIndex <= 0}
            className="rounded bg-zinc-800 p-1.5 text-zinc-300 transition-colors hover:bg-zinc-700 disabled:opacity-50 disabled:cursor-not-allowed"
            title="上一集 (Alt + ←)"
          >
            <ChevronLeft className="h-4 w-4" />
          </button>
          <button
            onClick={onNext}
            disabled={activeIndex >= episodes.length - 1}
            className="rounded bg-zinc-800 p-1.5 text-zinc-300 transition-colors hover:bg-zinc-700 disabled:opacity-50 disabled:cursor-not-allowed"
            title="下一集 (Alt + →)"
          >
            <ChevronRight className="h-4 w-4" />
          </button>
          {onToggleCollapse && (
            <button
              onClick={onToggleCollapse}
              className="rounded bg-zinc-800 p-1.5 text-zinc-300 transition-colors hover:bg-zinc-700 lg:hidden"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          )}
        </div>
      </div>

      {/* 分页控制 */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between border-b border-zinc-800 px-3 py-2">
          <button
            onClick={() => goToPage(currentPage - 1)}
            disabled={currentPage === 0}
            className="rounded p-1 text-zinc-400 transition-colors hover:bg-zinc-800 hover:text-white disabled:opacity-50"
          >
            <ChevronLeft className="h-4 w-4" />
          </button>
          <div className="flex items-center gap-1">
            {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
              // 显示当前页附近的页码
              let pageNum: number;
              if (totalPages <= 5) {
                pageNum = i;
              } else if (currentPage <= 2) {
                pageNum = i;
              } else if (currentPage >= totalPages - 3) {
                pageNum = totalPages - 5 + i;
              } else {
                pageNum = currentPage - 2 + i;
              }

              return (
                <button
                  key={pageNum}
                  onClick={() => goToPage(pageNum)}
                  className={`min-w-[28px] rounded px-2 py-1 text-xs transition-colors ${
                    pageNum === currentPage
                      ? 'bg-netflix-red text-white'
                      : 'text-zinc-400 hover:bg-zinc-800 hover:text-white'
                  }`}
                >
                  {pageNum + 1}
                </button>
              );
            })}
          </div>
          <button
            onClick={() => goToPage(currentPage + 1)}
            disabled={currentPage === totalPages - 1}
            className="rounded p-1 text-zinc-400 transition-colors hover:bg-zinc-800 hover:text-white disabled:opacity-50"
          >
            <ChevronRight className="h-4 w-4" />
          </button>
        </div>
      )}

      {/* 集数网格 */}
      <div className="max-h-[400px] overflow-y-auto p-3">
        <div
          className="grid gap-2"
          style={{ gridTemplateColumns: `repeat(${GridCols}, minmax(0, 1fr))` }}
        >
          {currentPageEpisodes.map(({ url, index, title }) => {
            const isActive = index === activeIndex;
            const isWatched = watchedEpisodes.has(index);

            return (
              <button
                key={`${url}-${index}`}
                onClick={() => onEpisodeChange(index)}
                className={`relative flex flex-col items-center justify-center rounded border p-2 text-center transition-all ${
                  isActive
                    ? 'border-netflix-red bg-netflix-red/10 text-white'
                    : isWatched
                    ? 'border-zinc-700 bg-zinc-800/50 text-zinc-300 hover:bg-zinc-800'
                    : 'border-zinc-800 bg-zinc-900/50 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200'
                }`}
                title={title}
              >
                <span className="text-xs font-medium truncate w-full">
                  {title}
                </span>
                {isActive && (
                  <Play className="absolute right-1 top-1 h-3 w-3 text-netflix-red" />
                )}
                {isWatched && !isActive && (
                  <Check className="absolute right-1 top-1 h-3 w-3 text-green-500" />
                )}
              </button>
            );
          })}
        </div>
      </div>

      {/* 底部：跳转当前 */}
      <div className="border-t border-zinc-800 p-2">
        <button
          onClick={goToActivePage}
          className="w-full rounded bg-zinc-800 py-1.5 text-xs text-zinc-300 transition-colors hover:bg-zinc-700"
        >
          跳转至当前播放 (第 {activeIndex + 1} 集)
        </button>
      </div>
    </div>
  );
}
