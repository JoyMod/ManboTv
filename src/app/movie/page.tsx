'use client';

import { Film } from 'lucide-react';

import { MOVIE_TYPE_OPTIONS } from '@/lib/content-taxonomy';

import BrowsePage from '@/components/BrowsePage';

// 地区选项
const regions = [
  { label: '全部', value: '全部' },
  { label: '华语', value: '华语' },
  { label: '美国', value: '美国' },
  { label: '韩国', value: '韩国' },
  { label: '日本', value: '日本' },
  { label: '印度', value: '印度' },
  { label: '泰国', value: '泰国' },
  { label: '法国', value: '法国' },
  { label: '英国', value: '英国' },
  { label: '德国', value: '德国' },
  { label: '俄罗斯', value: '俄罗斯' },
  { label: '西班牙', value: '西班牙' },
  { label: '意大利', value: '意大利' },
  { label: '其他', value: '其他' },
];

// 年代选项（单年精确到2026）
const years = [
  { label: '全部', value: '全部' },
  { label: '2026', value: '2026' },
  { label: '2025', value: '2025' },
  { label: '2024', value: '2024' },
  { label: '2023', value: '2023' },
  { label: '2022', value: '2022' },
  { label: '2021', value: '2021' },
  { label: '2020', value: '2020' },
  { label: '2019', value: '2019' },
  { label: '2010s', value: '2010s' },
  { label: '2000s', value: '2000s' },
  { label: '90年代', value: '90s' },
  { label: '更早', value: 'earlier' },
];

// 特色选项
const features = [
  { label: '全部', value: '全部' },
  { label: '豆瓣高分', value: '豆瓣高分' },
  { label: '获奖佳作', value: '获奖佳作' },
  { label: '新片热映', value: '新片热映' },
  { label: '经典重温', value: '经典重温' },
  { label: 'IMAX', value: 'IMAX' },
  { label: '4K', value: '4K' },
];

// 排序选项
const sortOptions = [
  { label: '综合排序', value: 'default' },
  { label: '最新上线', value: 'latest' },
  { label: '最热播放', value: 'hot' },
  { label: '最高评分', value: 'rating' },
  { label: '最多评论', value: 'comments' },
];

// 筛选组配置
const filterGroups = [
  {
    id: 'type',
    label: '类型',
    options: MOVIE_TYPE_OPTIONS,
    multiple: true,
  },
  {
    id: 'region',
    label: '地区',
    options: regions,
    multiple: false,
  },
  {
    id: 'feature',
    label: '特色',
    options: features,
    multiple: false,
  },
  {
    id: 'year',
    label: '年代',
    options: years,
    multiple: false,
  },
];

export default function MoviePage() {
  return (
    <BrowsePage
      title='电影'
      subtitle='热门大片 · 经典佳作 · 尽在曼波TV'
      kind='movie'
      filterGroups={filterGroups}
      sortOptions={sortOptions}
      heroGradient='from-netflix-red/20'
      icon={<Film className='w-16 h-16 text-netflix-red' />}
    />
  );
}
