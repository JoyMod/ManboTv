'use client';

import {
  ArrowUpDown,
  CalendarRange,
  Radio,
  RotateCcw,
  SlidersHorizontal,
} from 'lucide-react';
import React, { useEffect, useState } from 'react';

import {
  SearchExecutionInfo,
  SearchFacetBucket,
  SearchFacets,
} from '@/components/search/search-utils';

type SearchViewMode = 'aggregate' | 'lines' | 'sources';
type SearchSourceMode = 'all' | 'multi' | 'single';

interface SearchFilterPanelProps {
  facets: SearchFacets;
  execution?: SearchExecutionInfo;
  selectedView: SearchViewMode;
  selectedSort: string;
  selectedTypes: string[];
  selectedSources: string[];
  selectedSourceMode: SearchSourceMode;
  selectedYearFrom?: number;
  selectedYearTo?: number;
  onViewChange: (value: SearchViewMode) => void;
  onSortChange: (value: string) => void;
  onToggleType: (value: string) => void;
  onToggleSource: (value: string) => void;
  onSourceModeChange: (value: SearchSourceMode) => void;
  onYearRangeApply: (yearFrom?: number, yearTo?: number) => void;
  onResetFilters: () => void;
}

const viewOptions: Array<{ label: string; value: SearchViewMode }> = [
  { label: '聚合视图', value: 'aggregate' },
  { label: '线路视图', value: 'lines' },
  { label: '资源站视图', value: 'sources' },
];

const sortOptions: Array<{ label: string; value: string }> = [
  { label: '智能排序', value: 'smart' },
  { label: '年份从新到旧', value: 'year_desc' },
  { label: '年份从旧到新', value: 'year_asc' },
  { label: '标题排序', value: 'title' },
  { label: '可播放优先', value: 'playable' },
];

const sourceModeOptions: Array<{ label: string; value: SearchSourceMode }> = [
  { label: '全部资源', value: 'all' },
  { label: '多源优先', value: 'multi' },
  { label: '单源补充', value: 'single' },
];

function isActive(value: string, selected: string[]): boolean {
  return selected.includes(value);
}

function renderFacetChip(
  item: SearchFacetBucket,
  active: boolean,
  onClick: () => void
) {
  return (
    <button
      key={item.value}
      type='button'
      onClick={onClick}
      className={`rounded-full border px-3 py-1.5 text-xs transition-colors ${
        active
          ? 'border-white bg-white text-black'
          : 'border-netflix-gray-700 text-netflix-gray-300 hover:border-netflix-gray-500 hover:text-white'
      }`}
    >
      {item.label}
      <span className='ml-1 text-[10px] opacity-70'>{item.count}</span>
    </button>
  );
}

export default function SearchFilterPanel({
  facets,
  execution,
  selectedView,
  selectedSort,
  selectedTypes,
  selectedSources,
  selectedSourceMode,
  selectedYearFrom,
  selectedYearTo,
  onViewChange,
  onSortChange,
  onToggleType,
  onToggleSource,
  onSourceModeChange,
  onYearRangeApply,
  onResetFilters,
}: SearchFilterPanelProps) {
  const [yearFromInput, setYearFromInput] = useState(
    selectedYearFrom ? String(selectedYearFrom) : ''
  );
  const [yearToInput, setYearToInput] = useState(
    selectedYearTo ? String(selectedYearTo) : ''
  );

  useEffect(() => {
    setYearFromInput(selectedYearFrom ? String(selectedYearFrom) : '');
    setYearToInput(selectedYearTo ? String(selectedYearTo) : '');
  }, [selectedYearFrom, selectedYearTo]);

  const applyYearRange = () => {
    const nextYearFrom = Number.parseInt(yearFromInput, 10);
    const nextYearTo = Number.parseInt(yearToInput, 10);
    onYearRangeApply(
      Number.isFinite(nextYearFrom) ? nextYearFrom : undefined,
      Number.isFinite(nextYearTo) ? nextYearTo : undefined
    );
  };

  return (
    <section className='rounded-3xl border border-netflix-gray-800 bg-netflix-surface/70 p-4 md:p-5'>
      <div className='flex flex-wrap items-center justify-between gap-3'>
        <div>
          <p className='flex items-center gap-2 text-sm text-netflix-gray-300'>
            <SlidersHorizontal className='h-4 w-4' />
            高级搜索
          </p>
          <p className='mt-1 text-xs text-netflix-gray-500'>
            {execution?.degraded
              ? '部分资源站已提前返回，当前结果可先浏览。'
              : '服务端已完成聚合、筛选和排序。'}
          </p>
        </div>

        <button
          type='button'
          onClick={onResetFilters}
          className='inline-flex items-center gap-2 rounded-full border border-netflix-gray-700 px-4 py-2 text-xs text-netflix-gray-300 transition-colors hover:border-netflix-gray-500 hover:text-white'
        >
          <RotateCcw className='h-3.5 w-3.5' />
          重置筛选
        </button>
      </div>

      <div className='mt-4 flex flex-wrap gap-2'>
        {viewOptions.map((option) => (
          <button
            key={option.value}
            type='button'
            onClick={() => onViewChange(option.value)}
            className={`rounded-full px-3 py-1.5 text-xs transition-colors ${
              selectedView === option.value
                ? 'bg-netflix-red text-white'
                : 'border border-netflix-gray-700 text-netflix-gray-300 hover:border-netflix-gray-500 hover:text-white'
            }`}
          >
            {option.label}
          </button>
        ))}
      </div>

      <div className='mt-5 grid gap-4 lg:grid-cols-[1.1fr_1fr]'>
        <div className='space-y-4'>
          <div>
            <p className='mb-2 flex items-center gap-2 text-xs uppercase tracking-[0.2em] text-netflix-gray-500'>
              <ArrowUpDown className='h-3.5 w-3.5' />
              排序
            </p>
            <div className='flex flex-wrap gap-2'>
              {sortOptions.map((option) => (
                <button
                  key={option.value}
                  type='button'
                  onClick={() => onSortChange(option.value)}
                  className={`rounded-full px-3 py-1.5 text-xs transition-colors ${
                    selectedSort === option.value
                      ? 'bg-white text-black'
                      : 'border border-netflix-gray-700 text-netflix-gray-300 hover:border-netflix-gray-500 hover:text-white'
                  }`}
                >
                  {option.label}
                </button>
              ))}
            </div>
          </div>

          <div>
            <p className='mb-2 flex items-center gap-2 text-xs uppercase tracking-[0.2em] text-netflix-gray-500'>
              <Radio className='h-3.5 w-3.5' />
              来源模式
            </p>
            <div className='flex flex-wrap gap-2'>
              {sourceModeOptions.map((option) => (
                <button
                  key={option.value}
                  type='button'
                  onClick={() => onSourceModeChange(option.value)}
                  className={`rounded-full px-3 py-1.5 text-xs transition-colors ${
                    selectedSourceMode === option.value
                      ? 'bg-netflix-red text-white'
                      : 'border border-netflix-gray-700 text-netflix-gray-300 hover:border-netflix-gray-500 hover:text-white'
                  }`}
                >
                  {option.label}
                </button>
              ))}
            </div>
          </div>

          <div>
            <p className='mb-2 text-xs uppercase tracking-[0.2em] text-netflix-gray-500'>
              类型
            </p>
            <div className='flex flex-wrap gap-2'>
              {(facets.types || []).map((item) =>
                renderFacetChip(item, isActive(item.value, selectedTypes), () =>
                  onToggleType(item.value)
                )
              )}
            </div>
          </div>
        </div>

        <div className='space-y-4'>
          <div>
            <p className='mb-2 text-xs uppercase tracking-[0.2em] text-netflix-gray-500'>
              资源站
            </p>
            <div className='flex flex-wrap gap-2'>
              {(facets.sources || []).map((item) =>
                renderFacetChip(
                  item,
                  isActive(item.value, selectedSources),
                  () => onToggleSource(item.value)
                )
              )}
            </div>
          </div>

          <div>
            <p className='mb-2 flex items-center gap-2 text-xs uppercase tracking-[0.2em] text-netflix-gray-500'>
              <CalendarRange className='h-3.5 w-3.5' />
              年份区间
            </p>
            <div className='flex flex-wrap items-center gap-2'>
              <input
                type='number'
                inputMode='numeric'
                value={yearFromInput}
                onChange={(event) => setYearFromInput(event.target.value)}
                placeholder='起始年份'
                className='w-28 rounded-full border border-netflix-gray-700 bg-black/30 px-4 py-2 text-sm text-white outline-none focus:border-netflix-red'
              />
              <span className='text-sm text-netflix-gray-500'>到</span>
              <input
                type='number'
                inputMode='numeric'
                value={yearToInput}
                onChange={(event) => setYearToInput(event.target.value)}
                placeholder='结束年份'
                className='w-28 rounded-full border border-netflix-gray-700 bg-black/30 px-4 py-2 text-sm text-white outline-none focus:border-netflix-red'
              />
              <button
                type='button'
                onClick={applyYearRange}
                className='rounded-full bg-netflix-red px-4 py-2 text-xs font-medium text-white transition-colors hover:bg-netflix-red-hover'
              >
                应用
              </button>
            </div>

            {(facets.years || []).length > 0 && (
              <div className='mt-3 flex flex-wrap gap-2'>
                {(facets.years || []).slice(0, 10).map((item) =>
                  renderFacetChip(
                    item,
                    String(selectedYearFrom || '') === item.value &&
                      String(selectedYearTo || '') === item.value,
                    () => {
                      const exactYear = Number.parseInt(item.value, 10);
                      onYearRangeApply(
                        Number.isFinite(exactYear) ? exactYear : undefined,
                        Number.isFinite(exactYear) ? exactYear : undefined
                      );
                    }
                  )
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </section>
  );
}
