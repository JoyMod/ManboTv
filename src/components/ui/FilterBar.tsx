'use client';

import { AnimatePresence,motion } from 'framer-motion';
import {
  Check,
  ChevronDown,
  LayoutGrid,
  List,
  SlidersHorizontal,
  X,
} from 'lucide-react';
import React, { useEffect,useState } from 'react';

export interface FilterOption {
  label: string;
  value: string;
}

export interface FilterGroup {
  id: string;
  label: string;
  options: FilterOption[];
  multiple?: boolean;
}

export interface SortOption {
  label: string;
  value: string;
}

interface FilterBarProps {
  filterGroups: FilterGroup[];
  sortOptions: SortOption[];
  selectedFilters: Record<string, string | string[]>;
  selectedSort: string;
  onFilterChange: (groupId: string, value: string | string[]) => void;
  onSortChange: (value: string) => void;
  onClearFilters: () => void;
  totalResults: number;
  viewMode?: 'grid' | 'list';
  onViewModeChange?: (mode: 'grid' | 'list') => void;
}

// 底部固定筛选栏
function StickyFilterBar({
  filterGroups,
  sortOptions,
  selectedFilters,
  selectedSort,
  onFilterChange,
  onSortChange,
  onClearFilters,
  totalResults,
  viewMode,
  onViewModeChange,
  activeDropdown,
  setActiveDropdown,
}: FilterBarProps & {
  activeDropdown: string | null;
  setActiveDropdown: (id: string | null) => void;
}) {
  // 计算激活的筛选器数量
  const activeFilterCount = Object.entries(selectedFilters).reduce(
    (count, [, value]) => {
      if (Array.isArray(value)) {
        return count + value.filter((v) => v !== '全部').length;
      }
      return value !== '全部' ? count + 1 : count;
    },
    0
  );

  const selectedSortLabel =
    sortOptions.find((opt) => opt.value === selectedSort)?.label || '综合排序';

  return (
    <div className="fixed left-0 right-0 top-16 z-40 border-b border-zinc-800 bg-zinc-900/95 backdrop-blur-md">
      <div className="mx-auto flex max-w-[1920px] items-center justify-between px-4 py-3 sm:px-8">
        {/* 左侧：筛选按钮和计数 */}
        <div className="flex items-center gap-4">
          {/* 筛选按钮 */}
          <button
            onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
            className="flex items-center gap-2 rounded-full bg-zinc-800 px-4 py-2 text-sm text-white transition-colors hover:bg-zinc-700"
          >
            <SlidersHorizontal className="h-4 w-4" />
            筛选
            {activeFilterCount > 0 && (
              <span className="ml-1 flex h-5 w-5 items-center justify-center rounded-full bg-netflix-red text-xs font-bold">
                {activeFilterCount}
              </span>
            )}
          </button>

          {/* 筛选下拉菜单 */}
          <div className="hidden items-center gap-2 md:flex">
            {filterGroups.map((group) => {
              const value = selectedFilters[group.id];
              const label = Array.isArray(value)
                ? value.filter((v) => v !== '全部').join(', ') || '全部'
                : value || '全部';
              const isActive = label !== '全部';

              return (
                <div key={group.id} className="relative">
                  <button
                    onClick={() =>
                      setActiveDropdown(
                        activeDropdown === group.id ? null : group.id
                      )
                    }
                    className={`flex items-center gap-1 rounded-full px-3 py-2 text-sm transition-colors ${
                      isActive
                        ? 'bg-netflix-red/20 text-netflix-red'
                        : 'bg-zinc-800 text-zinc-300 hover:bg-zinc-700'
                    }`}
                  >
                    {group.label} {label}
                    <ChevronDown
                      className={`h-4 w-4 transition-transform ${
                        activeDropdown === group.id ? 'rotate-180' : ''
                      }`}
                    />
                  </button>

                  {/* 下拉菜单 */}
                  <AnimatePresence>
                    {activeDropdown === group.id && (
                      <motion.div
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: 10 }}
                        className="absolute bottom-full left-0 z-50 mb-2 w-48 rounded-lg border border-zinc-700 bg-zinc-900 p-2 shadow-xl"
                      >
                        {group.options.map((option) => {
                          const isSelected = Array.isArray(value)
                            ? value.includes(option.value)
                            : value === option.value;
                          return (
                            <button
                              key={option.value}
                              onClick={() => {
                                if (group.multiple) {
                                  if (option.value === '全部') {
                                    onFilterChange(group.id, ['全部']);
                                  } else {
                                    const currentValues = Array.isArray(value)
                                      ? value
                                      : [value];
                                    const newValues = isSelected
                                      ? currentValues.filter(
                                          (v) => v !== option.value
                                        )
                                      : [
                                          ...currentValues.filter(
                                            (v) => v !== '全部'
                                          ),
                                          option.value,
                                        ];
                                    onFilterChange(
                                      group.id,
                                      newValues.length > 0
                                        ? newValues
                                        : ['全部']
                                    );
                                  }
                                } else {
                                  onFilterChange(group.id, option.value);
                                  setActiveDropdown(null);
                                }
                              }}
                              className={`flex w-full items-center justify-between rounded px-3 py-2 text-left text-sm transition-colors ${
                                isSelected
                                  ? 'bg-netflix-red/20 text-netflix-red'
                                  : 'text-zinc-300 hover:bg-zinc-800'
                              }`}
                            >
                              {option.label}
                              {isSelected && <Check className="h-4 w-4" />}
                            </button>
                          );
                        })}
                      </motion.div>
                    )}
                  </AnimatePresence>
                </div>
              );
            })}
          </div>

          {/* 清除筛选 */}
          {activeFilterCount > 0 && (
            <button
              onClick={onClearFilters}
              className="flex items-center gap-1 text-sm text-zinc-500 transition-colors hover:text-white"
            >
              <X className="h-4 w-4" />
              清除
            </button>
          )}
        </div>

        {/* 中间：结果计数 */}
        <div className="hidden text-sm text-zinc-400 md:block">
          找到 <span className="font-bold text-white">{totalResults}</span>{' '}
          部影片
        </div>

        {/* 右侧：排序和视图切换 */}
        <div className="flex items-center gap-2">
          {/* 排序下拉 */}
          <div className="relative">
            <button
              onClick={() =>
                setActiveDropdown(activeDropdown === 'sort' ? null : 'sort')
              }
              className="flex items-center gap-1 rounded-full bg-zinc-800 px-3 py-2 text-sm text-zinc-300 transition-colors hover:bg-zinc-700"
            >
              {selectedSortLabel}
              <ChevronDown
                className={`h-4 w-4 transition-transform ${
                  activeDropdown === 'sort' ? 'rotate-180' : ''
                }`}
              />
            </button>

            <AnimatePresence>
              {activeDropdown === 'sort' && (
                <motion.div
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: 10 }}
                  className="absolute bottom-full right-0 z-50 mb-2 w-32 rounded-lg border border-zinc-700 bg-zinc-900 p-2 shadow-xl"
                >
                  {sortOptions.map((option) => (
                    <button
                      key={option.value}
                      onClick={() => {
                        onSortChange(option.value);
                        setActiveDropdown(null);
                      }}
                      className={`flex w-full items-center justify-between rounded px-3 py-2 text-left text-sm transition-colors ${
                        selectedSort === option.value
                          ? 'bg-netflix-red/20 text-netflix-red'
                          : 'text-zinc-300 hover:bg-zinc-800'
                      }`}
                    >
                      {option.label}
                      {selectedSort === option.value && (
                        <Check className="h-4 w-4" />
                      )}
                    </button>
                  ))}
                </motion.div>
              )}
            </AnimatePresence>
          </div>

          {/* 视图切换 */}
          {onViewModeChange && (
            <div className="hidden rounded-full bg-zinc-800 p-1 md:flex">
              <button
                onClick={() => onViewModeChange('grid')}
                className={`rounded-full p-1.5 transition-colors ${
                  viewMode === 'grid'
                    ? 'bg-zinc-600 text-white'
                    : 'text-zinc-400 hover:text-white'
                }`}
              >
                <LayoutGrid className="h-4 w-4" />
              </button>
              <button
                onClick={() => onViewModeChange('list')}
                className={`rounded-full p-1.5 transition-colors ${
                  viewMode === 'list'
                    ? 'bg-zinc-600 text-white'
                    : 'text-zinc-400 hover:text-white'
                }`}
              >
                <List className="h-4 w-4" />
              </button>
            </div>
          )}
        </div>
      </div>

      {/* 点击外部关闭下拉 */}
      {activeDropdown && (
        <div
          className="fixed inset-0 z-[-1]"
          onClick={() => setActiveDropdown(null)}
        />
      )}
    </div>
  );
}

export default function FilterBar(props: FilterBarProps) {
  const [activeDropdown, setActiveDropdown] = useState<string | null>(null);

  // 点击外部关闭下拉菜单
  useEffect(() => {
    const handleClickOutside = () => setActiveDropdown(null);
    if (activeDropdown) {
      document.addEventListener('click', handleClickOutside);
      return () => document.removeEventListener('click', handleClickOutside);
    }
  }, [activeDropdown]);

  return (
    <StickyFilterBar
      {...props}
      activeDropdown={activeDropdown}
      setActiveDropdown={setActiveDropdown}
    />
  );
}
