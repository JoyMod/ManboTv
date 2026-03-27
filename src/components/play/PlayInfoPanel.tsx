'use client';

import {
  ChevronLeft,
  Heart,
  Loader2,
  Power,
} from 'lucide-react';
import React from 'react';

import EpisodeSelector from '@/components/EpisodeSelector';
import { DetailResult, SourceCandidate, SourceTestResult } from '@/components/play/play-utils';

interface PlayInfoPanelProps {
  title: string;
  detail: DetailResult | null;
  source: string;
  id: string;
  skipEnabled: boolean;
  qualityList: Array<{ index: number; label: string }>;
  activeQuality: number;
  onQualityChange: (event: React.ChangeEvent<HTMLSelectElement>) => void;
  onBack: () => void;
  onFavorite: () => void;
  favoriteSaving: boolean;
  isFavorite: boolean;
  sourcePanelLoading: boolean;
  sourceBatchTesting: boolean;
  availableSources: SourceCandidate[];
  sourceTests: Record<string, SourceTestResult>;
  onSourceSwitch: (candidate: SourceCandidate) => void;
  onSourceTest: (candidate: SourceCandidate) => void;
  onTestMoreSources: () => void;
  episodes: string[];
  episodeTitles: string[];
  activeEpisodeIndex: number;
  onEpisodeChange: (index: number) => void;
  onPreviousEpisode: () => void;
  onNextEpisode: () => void;
}

export default function PlayInfoPanel({
  title,
  detail,
  source,
  id,
  skipEnabled,
  qualityList,
  activeQuality,
  onQualityChange,
  onBack,
  onFavorite,
  favoriteSaving,
  isFavorite,
  sourcePanelLoading,
  sourceBatchTesting,
  availableSources,
  sourceTests,
  onSourceSwitch,
  onSourceTest,
  onTestMoreSources,
  episodes,
  episodeTitles,
  activeEpisodeIndex,
  onEpisodeChange,
  onPreviousEpisode,
  onNextEpisode,
}: PlayInfoPanelProps) {
  return (
    <div className='space-y-6 lg:col-span-2'>
      <div className='flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between'>
        <div>
          <h1 className='text-3xl font-bold text-white'>{title}</h1>
          <p className='mt-2 text-sm text-netflix-gray-400'>
            {detail?.year || ''}
            {detail?.class ? ` · ${detail.class}` : ''}
            {detail?.source_name ? ` · 来源: ${detail.source_name}` : ''}
            {skipEnabled ? (
              <span className='ml-2 text-green-400'>· 已启用片头片尾跳过</span>
            ) : null}
          </p>
        </div>

        <div className='flex flex-wrap items-center gap-2'>
          {qualityList.length > 0 ? (
            <select
              value={activeQuality}
              onChange={onQualityChange}
              className='rounded bg-zinc-800 px-3 py-2 text-xs text-white outline-none'
            >
              <option value={-1}>自动清晰度</option>
              {qualityList.map((quality) => (
                <option key={quality.index} value={quality.index}>
                  {quality.label}
                </option>
              ))}
            </select>
          ) : null}

          <button
            onClick={onBack}
            className='inline-flex items-center gap-1 rounded bg-zinc-800 px-3 py-2 text-sm text-white hover:bg-zinc-700'
          >
            <ChevronLeft className='h-4 w-4' />
            返回
          </button>
          <button
            onClick={onFavorite}
            disabled={favoriteSaving}
            className={`inline-flex items-center gap-1 rounded px-3 py-2 text-sm transition-colors ${
              isFavorite
                ? 'bg-red-600 text-white'
                : 'bg-zinc-800 text-white hover:bg-zinc-700'
            }`}
          >
            <Heart className={`h-4 w-4 ${isFavorite ? 'fill-white' : ''}`} />
            {isFavorite ? '已收藏' : '收藏'}
          </button>
        </div>
      </div>

      <div className='rounded border border-zinc-800 bg-zinc-900/60 p-4'>
        <div className='mb-3 flex items-center justify-between'>
          <div>
            <h2 className='text-lg font-bold text-white'>资源线路</h2>
            <p className='mt-1 text-xs text-zinc-500'>
              先显示当前线路，其它线路按需检测，避免抢首屏资源。
            </p>
          </div>
          <div className='flex items-center gap-3'>
            {sourcePanelLoading || sourceBatchTesting ? (
              <span className='inline-flex items-center gap-1 text-xs text-zinc-400'>
                <Loader2 className='h-3 w-3 animate-spin' />
                {sourcePanelLoading ? '整理线路中' : '检测更多线路中'}
              </span>
            ) : null}
            <button
              type='button'
              onClick={onTestMoreSources}
              disabled={
                sourcePanelLoading ||
                sourceBatchTesting ||
                availableSources.length === 0
              }
              className='rounded-full border border-zinc-700 px-3 py-1 text-xs text-zinc-300 transition-colors hover:border-zinc-500 hover:text-white disabled:cursor-not-allowed disabled:opacity-50'
            >
              检测更多线路
            </button>
          </div>
        </div>

        {availableSources.length > 0 ? (
          <div className='grid grid-cols-1 gap-2 sm:grid-cols-2'>
            {availableSources.map((candidate) => {
              const key = `${candidate.source}+${candidate.id}`;
              const test = sourceTests[key];
              const isCurrent = candidate.source === source && candidate.id === id;

              return (
                <div
                  key={key}
                  className={`rounded border px-3 py-2 text-left transition-colors ${
                    isCurrent
                      ? 'border-netflix-red bg-netflix-red/10'
                      : 'border-zinc-700 bg-zinc-800/40 hover:bg-zinc-700/50'
                  }`}
                >
                  <button
                    type='button'
                    onClick={() => onSourceSwitch(candidate)}
                    className='w-full text-left'
                  >
                    <div className='flex items-center justify-between'>
                      <span className='truncate text-sm font-semibold text-white'>
                        {candidate.source_name || candidate.source}
                      </span>
                      <Power className='h-4 w-4 text-zinc-300' />
                    </div>
                    <div className='mt-1 flex flex-wrap items-center gap-2 text-[11px] text-zinc-400'>
                      <span>{candidate.episodes.length} 集</span>
                      <span>
                        清晰度：
                        {test?.quality ||
                          (test?.status === 'testing' ? '检测中' : '未知')}
                      </span>
                      <span>
                        速度：
                        {test?.status === 'testing' ? '检测中' : test?.speed || '--'}
                      </span>
                      <span>
                        延迟：
                        {typeof test?.pingMs === 'number' ? `${test.pingMs}ms` : '--'}
                      </span>
                    </div>
                  </button>
                  <div className='mt-3 flex items-center justify-between border-t border-white/5 pt-2'>
                    <span className='text-[11px] text-zinc-500'>
                      {isCurrent ? '当前播放线路' : '可切换候选线路'}
                    </span>
                    <button
                      type='button'
                      onClick={() => onSourceTest(candidate)}
                      disabled={test?.status === 'testing'}
                      className='rounded-full border border-zinc-700 px-2.5 py-1 text-[11px] text-zinc-300 transition-colors hover:border-zinc-500 hover:text-white disabled:cursor-not-allowed disabled:opacity-50'
                    >
                      {test?.status === 'testing'
                        ? '测速中'
                        : test?.status === 'ok'
                        ? '重新测速'
                        : '测速'}
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          <p className='text-sm text-zinc-500'>未找到可用线路</p>
        )}
      </div>

      <div>
        <h2 className='mb-3 text-lg font-bold text-white'>选集</h2>
        <EpisodeSelector
          episodes={episodes}
          episodeTitles={episodeTitles}
          activeIndex={activeEpisodeIndex}
          onEpisodeChange={onEpisodeChange}
          onPrevious={onPreviousEpisode}
          onNext={onNextEpisode}
        />
      </div>

      <div>
        <h2 className='mb-2 text-lg font-bold text-white'>简介</h2>
        <p className='leading-relaxed text-netflix-gray-300'>
          {detail?.desc || '暂无剧情简介'}
        </p>
      </div>

      <div className='rounded border border-zinc-800 bg-zinc-900/60 p-4'>
        <h3 className='mb-2 text-sm font-semibold text-white'>快捷键与功能</h3>
        <div className='grid grid-cols-2 gap-2 text-xs text-zinc-400'>
          <p>空格：播放/暂停</p>
          <p>左右方向键：快退/快进10秒</p>
          <p>M：静音</p>
          <p>F：全屏</p>
          <p>Alt + ←：上一集</p>
          <p>Alt + →：下一集</p>
        </div>
        <p className='mt-2 text-xs text-green-400'>
          播放器设置中可配置片头片尾跳过和去广告
        </p>
      </div>
    </div>
  );
}
