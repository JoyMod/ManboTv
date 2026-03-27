'use client';

import { Tv } from 'lucide-react';

import BrowsePage from '@/components/BrowsePage';

// 电视剧类型选项（22+）
const types = [
  { label: '全部', value: '全部' },
  { label: '古装', value: '古装' },
  { label: '都市', value: '都市' },
  { label: '悬疑', value: '悬疑' },
  { label: '爱情', value: '爱情' },
  { label: '武侠', value: '武侠' },
  { label: '奇幻', value: '奇幻' },
  { label: '谍战', value: '谍战' },
  { label: '军旅', value: '军旅' },
  { label: '喜剧', value: '喜剧' },
  { label: '家庭', value: '家庭' },
  { label: '科幻', value: '科幻' },
  { label: '青春', value: '青春' },
  { label: '传奇', value: '传奇' },
  { label: '农村', value: '农村' },
  { label: '历史', value: '历史' },
  { label: '宫廷', value: '宫廷' },
  { label: '仙侠', value: '仙侠' },
  { label: '甜宠', value: '甜宠' },
  { label: '职场', value: '职场' },
  { label: '校园', value: '校园' },
  { label: '穿越', value: '穿越' },
  { label: '民国', value: '民国' },
];

// 地区选项
const regions = [
  { label: '全部', value: '全部' },
  { label: '国产剧', value: '国产剧' },
  { label: '美剧', value: '美剧' },
  { label: '韩剧', value: '韩剧' },
  { label: '日剧', value: '日剧' },
  { label: '港剧', value: '港剧' },
  { label: '台剧', value: '台剧' },
  { label: '泰剧', value: '泰剧' },
  { label: '英剧', value: '英剧' },
];

// 状态选项
const status = [
  { label: '全部', value: '全部' },
  { label: '连载中', value: '连载中' },
  { label: '已完结', value: '已完结' },
  { label: '即将开播', value: '即将开播' },
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
];

// 排序选项
const sortOptions = [
  { label: '综合排序', value: 'default' },
  { label: '最新上线', value: 'latest' },
  { label: '最热播放', value: 'hot' },
  { label: '最高评分', value: 'rating' },
  { label: '最多弹幕', value: 'comments' },
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
    label: '年代',
    options: years,
    multiple: false,
  },
];

export default function TVPage() {
  return (
    <BrowsePage
      title='电视剧'
      subtitle='热播剧集 · 精彩连播 · 追剧不停'
      kind='tv'
      filterGroups={filterGroups}
      sortOptions={sortOptions}
      heroGradient='from-blue-600/20'
      icon={<Tv className='w-16 h-16 text-blue-500' />}
    />
  );
}
