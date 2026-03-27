'use client';

import { Smile } from 'lucide-react';

import BrowsePage from '@/components/BrowsePage';

// 综艺类型选项（18+）
const types = [
  { label: '全部', value: '全部' },
  { label: '真人秀', value: '真人秀' },
  { label: '脱口秀', value: '脱口秀' },
  { label: '音乐', value: '音乐' },
  { label: '情感', value: '情感' },
  { label: '竞技', value: '竞技' },
  { label: '美食', value: '美食' },
  { label: '旅行', value: '旅行' },
  { label: '游戏', value: '游戏' },
  { label: '访谈', value: '访谈' },
  { label: '选秀', value: '选秀' },
  { label: '晚会', value: '晚会' },
  { label: '喜剧', value: '喜剧' },
  { label: '文化', value: '文化' },
  { label: '亲子', value: '亲子' },
  { label: '舞蹈', value: '舞蹈' },
  { label: '时尚', value: '时尚' },
  { label: '明星', value: '明星' },
  { label: '汽车', value: '汽车' },
];

// 地区选项
const regions = [
  { label: '全部', value: '全部' },
  { label: '国内', value: '国内' },
  { label: '韩国', value: '韩国' },
  { label: '日本', value: '日本' },
  { label: '欧美', value: '欧美' },
  { label: '港台', value: '港台' },
];

// 状态选项
const status = [
  { label: '全部', value: '全部' },
  { label: '连载中', value: '连载中' },
  { label: '已完结', value: '已完结' },
  { label: '即将开播', value: '即将开播' },
];

// 年代选项（近5年）
const years = [
  { label: '全部', value: '全部' },
  { label: '2026', value: '2026' },
  { label: '2025', value: '2025' },
  { label: '2024', value: '2024' },
  { label: '2023', value: '2023' },
  { label: '2022', value: '2022' },
  { label: '2021', value: '2021' },
  { label: '2020', value: '2020' },
];

// 排序选项
const sortOptions = [
  { label: '综合排序', value: 'default' },
  { label: '最新上线', value: 'latest' },
  { label: '最热播放', value: 'hot' },
  { label: '最多评论', value: 'comments' },
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

export default function VarietyPage() {
  return (
    <BrowsePage
      title='综艺'
      subtitle='热门综艺 · 欢乐不停 · 精彩不断'
      kind='variety'
      filterGroups={filterGroups}
      sortOptions={sortOptions}
      heroGradient='from-yellow-600/20'
      icon={<Smile className='w-16 h-16 text-yellow-500' />}
    />
  );
}
