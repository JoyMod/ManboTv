'use client';

import { Sparkles } from 'lucide-react';

import BrowsePage from '@/components/BrowsePage';

// 动漫类型选项（20+）
const types = [
  { label: '全部', value: '全部' },
  { label: '热血', value: '热血' },
  { label: '恋爱', value: '恋爱' },
  { label: '搞笑', value: '搞笑' },
  { label: '悬疑', value: '悬疑' },
  { label: '科幻', value: '科幻' },
  { label: '机战', value: '机战' },
  { label: '运动', value: '运动' },
  { label: '校园', value: '校园' },
  { label: '魔法', value: '魔法' },
  { label: '冒险', value: '冒险' },
  { label: '战斗', value: '战斗' },
  { label: '日常', value: '日常' },
  { label: '治愈', value: '治愈' },
  { label: '奇幻', value: '奇幻' },
  { label: '后宫', value: '后宫' },
  { label: '百合', value: '百合' },
  { label: '耽美', value: '耽美' },
  { label: '神魔', value: '神魔' },
  { label: '推理', value: '推理' },
  { label: '音乐', value: '音乐' },
];

// 地区选项
const regions = [
  { label: '全部', value: '全部' },
  { label: '日本', value: '日本' },
  { label: '国产', value: '国产' },
  { label: '欧美', value: '欧美' },
];

// 状态选项
const status = [
  { label: '全部', value: '全部' },
  { label: '连载中', value: '连载中' },
  { label: '已完结', value: '已完结' },
  { label: '新番', value: '新番' },
  { label: '剧场版', value: '剧场版' },
  { label: 'OVA', value: 'OVA' },
];

// 年份选项（按季度划分）
const years = [
  { label: '全部', value: '全部' },
  { label: '2026冬', value: '2026-冬' },
  { label: '2025秋', value: '2025-秋' },
  { label: '2025夏', value: '2025-夏' },
  { label: '2025春', value: '2025-春' },
  { label: '2024冬', value: '2024-冬' },
  { label: '2024', value: '2024' },
  { label: '2023', value: '2023' },
  { label: '2022', value: '2022' },
  { label: '2021', value: '2021' },
  { label: '2020', value: '2020' },
  { label: '经典', value: '经典' },
];

// 排序选项
const sortOptions = [
  { label: '综合排序', value: 'default' },
  { label: '新番上线', value: 'latest' },
  { label: '最高评分', value: 'rating' },
  { label: '最多追番', value: 'follow' },
  { label: '最多播放', value: 'hot' },
];

// 筛选组配置
const filterGroups = [
  {
    id: 'type',
    label: '类型',
    options: types,
    multiple: true,
  },
  {
    id: 'region',
    label: '地区',
    options: regions,
    multiple: false,
  },
  {
    id: 'status',
    label: '状态',
    options: status,
    multiple: false,
  },
  {
    id: 'year',
    label: '年份',
    options: years,
    multiple: false,
  },
];

export default function AnimePage() {
  return (
    <BrowsePage
      title='动漫'
      subtitle='热血番剧 · 精彩剧场 · 追番必备'
      kind='anime'
      filterGroups={filterGroups}
      sortOptions={sortOptions}
      heroGradient='from-purple-600/20'
      icon={<Sparkles className='w-16 h-16 text-purple-500' />}
    />
  );
}
