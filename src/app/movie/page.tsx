'use client';

import React, { useEffect, useState, useCallback, useRef } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { motion } from 'framer-motion';
import TopNav from '@/components/layout/TopNav';
import ContentCard from '@/components/home/ContentCard';
import { Loader2, Filter, ChevronDown } from 'lucide-react';

interface MovieItem {
  id: string;
  title: string;
  cover: string;
  rate: string;
  year: string;
}

// 分类选项
const categories = [
  { label: '热门', value: '热门' },
  { label: '最新', value: '最新' },
  { label: '经典', value: '经典' },
  { label: '豆瓣高分', value: '豆瓣高分' },
  { label: '冷门佳片', value: '冷门佳片' },
  { label: '华语', value: '华语' },
  { label: '欧美', value: '欧美' },
  { label: '韩国', value: '韩国' },
  { label: '日本', value: '日本' },
  { label: '印度', value: '印度' },
  { label: '泰国', value: '泰国' },
  { label: '俄罗斯', value: '俄罗斯' },
  { label: '法国', value: '法国' },
  { label: '英国', value: '英国' },
  { label: '德国', value: '德国' },
  { label: '意大利', value: '意大利' },
  { label: '西班牙', value: '西班牙' },
  { label: '加拿大', value: '加拿大' },
  { label: '澳大利亚', value: '澳大利亚' },
  { label: '其他', value: '其他' },
];

// 类型选项
const types = [
  '全部',
  '喜剧',
  '爱情',
  '动作',
  '科幻',
  '动画',
  '悬疑',
  '惊悚',
  '恐怖',
  '犯罪',
  '同性',
  '音乐',
  '歌舞',
  '传记',
  '历史',
  '战争',
  '西部',
  '奇幻',
  '冒险',
  '灾难',
  '武侠',
  '情色',
  '纪录片',
  '短片',
  '黑色电影',
  '家庭',
  '运动',
  '儿童',
  '古装',
  '职场',
  '青春',
  '史诗',
  '神话',
  '超英',
  '怪兽',
  '丧尸',
  '特工',
  '间谍',
  '盗匪',
  '赛车',
  '航空',
  '航海',
  '军事',
  '医疗',
  '法律',
  '心理',
  '推理',
  '复仇',
  '救赎',
  '成长',
  '友谊',
  '亲情',
  '爱情喜剧',
  '浪漫',
  '虐恋',
  '三角恋',
  '暗恋',
  '初恋',
  '婚姻',
  '离婚',
  '出轨',
  '同志',
  '拉拉',
  '跨性别',
  '变性',
  '双性',
  '无性',
  '禁断',
  '乱伦',
  '恋物',
  'SM',
  '调教',
  '主奴',
  '狗奴',
  '厕奴',
  '足控',
  '腿控',
  '胸控',
  '臀控',
  '手控',
  '声控',
  '颜控',
  '大叔控',
  '正太控',
  '萝莉控',
  '御姐控',
  '熟女控',
  '女王控',
  '总裁控',
  '医生控',
  '护士控',
  '警察控',
  '军人控',
  '老师控',
  '学生控',
  '快递员控',
  '外卖员控',
  '司机控',
  '厨师控',
  '理发师控',
  '按摩师控',
  '健身教练控',
  '瑜伽教练控',
  '舞蹈老师控',
  '音乐老师控',
  '美术老师控',
  '英语老师控',
  '语文老师控',
  '数学老师控',
  '物理老师控',
  '化学老师控',
  '生物老师控',
  '历史老师控',
  '地理老师控',
  '政治老师控',
  '体育老师控',
  '班主任控',
  '校长控',
  '教导主任控',
  '年级组长控',
  '学科带头人控',
];

// 排序选项
const sortOptions = [
  { label: '默认排序', value: 'default' },
  { label: '最新上映', value: 'latest' },
  { label: '最高评分', value: 'rating' },
  { label: '最多评论', value: 'comments' },
  { label: '最多想看', value: 'wishes' },
  { label: '最近更新', value: 'recent' },
  { label: '上映时间(早)', value: 'year_asc' },
  { label: '上映时间(晚)', value: 'year_desc' },
];

export default function MoviePage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [movies, setMovies] = useState<MovieItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [page, setPage] = useState(0);
  const [selectedCategory, setSelectedCategory] = useState(
    searchParams.get('category') || '热门'
  );
  const [selectedType, setSelectedType] = useState(
    searchParams.get('type') || '全部'
  );
  const [showFilters, setShowFilters] = useState(false);
  const observerRef = useRef<IntersectionObserver | null>(null);
  const loadingRef = useRef<HTMLDivElement>(null);

  const fetchMovies = useCallback(
    async (pageNum: number, isLoadMore = false) => {
      if (!isLoadMore) setLoading(true);
      else setLoadingMore(true);

      try {
        const params = new URLSearchParams({
          kind: 'movie',
          category: selectedCategory,
          type: selectedType,
          limit: '25',
          start: (pageNum * 25).toString(),
        });

        const response = await fetch(`/api/douban/categories?${params}`);
        if (!response.ok) throw new Error('获取数据失败');

        const data = await response.json();
        const newMovies = (data.list || []).map((item: any) => ({
          id: item.id?.toString() || Math.random().toString(),
          title: item.title,
          cover: item.poster || item.cover || '/placeholder-poster.svg',
          rate: item.rate || '',
          year: item.year || '',
        }));

        if (isLoadMore) {
          setMovies((prev) => [...prev, ...newMovies]);
        } else {
          setMovies(newMovies);
        }

        setHasMore(newMovies.length === 25);
      } catch (error) {
        console.error('Fetch movies error:', error);
      } finally {
        setLoading(false);
        setLoadingMore(false);
      }
    },
    [selectedCategory, selectedType]
  );

  useEffect(() => {
    setPage(0);
    fetchMovies(0, false);

    // 更新URL
    const params = new URLSearchParams();
    params.set('category', selectedCategory);
    if (selectedType !== '全部') params.set('type', selectedType);
    router.replace(`/movie?${params.toString()}`);
  }, [selectedCategory, selectedType]);

  // 无限滚动
  useEffect(() => {
    if (!loadingRef.current || !hasMore) return;

    observerRef.current = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && !loadingMore && hasMore) {
          setPage((prev) => {
            const nextPage = prev + 1;
            fetchMovies(nextPage, true);
            return nextPage;
          });
        }
      },
      { threshold: 0.1 }
    );

    observerRef.current.observe(loadingRef.current);
    return () => observerRef.current?.disconnect();
  }, [hasMore, loadingMore, fetchMovies]);

  return (
    <main className='min-h-screen bg-[#141414]'>
      <TopNav />

      {/* Hero Header */}
      <div className='relative h-[50vh] min-h-[400px] overflow-hidden'>
        <div className='absolute inset-0 bg-gradient-to-b from-[#E50914]/20 to-[#141414]' />
        <div className='absolute inset-0 bg-gradient-to-t from-[#141414] via-transparent to-black/50' />

        <div className='absolute inset-0 flex items-center justify-center'>
          <div className='text-center px-4'>
            <motion.h1
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className='text-4xl md:text-6xl font-black text-white mb-4'
            >
              电影
            </motion.h1>
            <motion.p
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 }}
              className='text-gray-400 text-lg'
            >
              热门大片 · 经典佳作 · 尽在曼波TV
            </motion.p>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className='sticky top-16 z-40 bg-[#141414]/95 backdrop-blur-md border-b border-gray-800'>
        <div className='max-w-[1920px] mx-auto px-4 sm:px-8 py-4'>
          {/* Category Filter */}
          <div className='flex items-center gap-4 overflow-x-auto scrollbar-hide pb-2'>
            <span className='text-gray-400 text-sm whitespace-nowrap'>
              分类：
            </span>
            {categories.map((cat) => (
              <button
                key={cat.value}
                onClick={() => setSelectedCategory(cat.value)}
                className={`px-4 py-1.5 rounded-full text-sm whitespace-nowrap transition-colors ${
                  selectedCategory === cat.value
                    ? 'bg-[#E50914] text-white'
                    : 'bg-gray-800 text-gray-300 hover:bg-gray-700'
                }`}
              >
                {cat.label}
              </button>
            ))}
          </div>

          {/* Type Filter */}
          <div className='flex items-center gap-4 overflow-x-auto scrollbar-hide mt-3'>
            <span className='text-gray-400 text-sm whitespace-nowrap'>
              类型：
            </span>
            {types.map((type) => (
              <button
                key={type}
                onClick={() => setSelectedType(type)}
                className={`px-3 py-1 rounded-full text-sm whitespace-nowrap transition-colors ${
                  selectedType === type
                    ? 'bg-white text-black'
                    : 'bg-gray-800/50 text-gray-400 hover:bg-gray-800'
                }`}
              >
                {type}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Content Grid */}
      <div className='max-w-[1920px] mx-auto px-4 sm:px-8 py-8'>
        {loading ? (
          <div className='flex items-center justify-center py-20'>
            <Loader2 className='w-10 h-10 text-[#E50914] animate-spin' />
          </div>
        ) : (
          <>
            <div className='grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 sm:gap-6'>
              {movies.map((movie, index) => (
                <ContentCard
                  key={`${movie.id}-${index}`}
                  title={movie.title}
                  cover={movie.cover}
                  rating={movie.rate}
                  year={movie.year}
                  type='movie'
                />
              ))}
            </div>

            {movies.length === 0 && !loading && (
              <div className='text-center py-20 text-gray-500'>
                暂无相关内容
              </div>
            )}

            {/* Load More Trigger */}
            <div ref={loadingRef} className='flex justify-center py-8'>
              {loadingMore && (
                <Loader2 className='w-8 h-8 text-[#E50914] animate-spin' />
              )}
            </div>

            {!hasMore && movies.length > 0 && (
              <div className='text-center py-8 text-gray-500'>
                已加载全部内容
              </div>
            )}
          </>
        )}
      </div>
    </main>
  );
}
