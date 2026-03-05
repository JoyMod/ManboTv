'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useRouter } from 'next/navigation';
import TopNav from '@/components/layout/TopNav';
import ContentCard from '@/components/home/ContentCard';
import { 
  Loader2, 
  Heart, 
  Clock, 
  Trash2, 
  Play,
  AlertCircle,
  X
} from 'lucide-react';

interface FavoriteItem {
  id: string;
  source: string;
  source_name: string;
  title: string;
  cover: string;
  year: string;
  total_episodes: number;
  save_time: number;
}

interface PlayRecord {
  id: string;
  source: string;
  source_name: string;
  title: string;
  cover: string;
  year: string;
  index: number;
  play_time: number;
  total_episodes: number;
  last_play_time: number;
}

const tabs = [
  { label: '我的收藏', value: 'favorites', icon: Heart },
  { label: '播放历史', value: 'history', icon: Clock },
];

export default function FavoritesPage() {
  const router = useRouter();
  const [activeTab, setActiveTab] = useState('favorites');
  const [favorites, setFavorites] = useState<FavoriteItem[]>([]);
  const [history, setHistory] = useState<PlayRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);

  // 获取收藏列表
  const fetchFavorites = useCallback(async () => {
    try {
      const response = await fetch('/api/favorites');
      if (response.status === 401) {
        router.push('/login?redirect=/favorites');
        return;
      }
      if (!response.ok) throw new Error('获取收藏失败');
      
      const data = await response.json();
      // 转换数据格式
      const items = Object.entries(data).map(([key, value]: [string, any]) => ({
        id: key,
        source: value.source || '',
        source_name: value.source_name || '',
        title: value.title || '未知标题',
        cover: value.cover || '/placeholder-poster.svg',
        year: value.year || '',
        total_episodes: value.total_episodes || 0,
        save_time: value.save_time || Date.now(),
      }));
      
      // 按保存时间倒序
      items.sort((a, b) => b.save_time - a.save_time);
      setFavorites(items);
    } catch (err) {
      console.error('Fetch favorites error:', err);
      setError('获取收藏列表失败');
    }
  }, [router]);

  // 获取播放历史
  const fetchHistory = useCallback(async () => {
    try {
      const response = await fetch('/api/playrecords');
      if (response.status === 401) {
        router.push('/login?redirect=/favorites');
        return;
      }
      if (!response.ok) throw new Error('获取播放历史失败');
      
      const data = await response.json();
      // 转换数据格式
      const items = Object.entries(data).map(([key, value]: [string, any]) => ({
        id: key,
        source: value.source || '',
        source_name: value.source_name || '',
        title: value.title || '未知标题',
        cover: value.cover || '/placeholder-poster.svg',
        year: value.year || '',
        index: value.index || 1,
        play_time: value.play_time || 0,
        total_episodes: value.total_episodes || 0,
        last_play_time: value.last_play_time || Date.now(),
      }));
      
      // 按最后播放时间倒序
      items.sort((a, b) => b.last_play_time - a.last_play_time);
      setHistory(items);
    } catch (err) {
      console.error('Fetch history error:', err);
      setError('获取播放历史失败');
    }
  }, [router]);

  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      setError(null);
      
      if (activeTab === 'favorites') {
        await fetchFavorites();
      } else {
        await fetchHistory();
      }
      
      setLoading(false);
    };

    loadData();
  }, [activeTab, fetchFavorites, fetchHistory]);

  // 删除收藏
  const handleDeleteFavorite = async (id: string) => {
    try {
      const item = favorites.find(f => f.id === id);
      if (!item) return;

      const response = await fetch('/api/favorites', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ source: item.source, id: item.id.replace(`${item.source}_`, '') }),
      });

      if (response.ok) {
        setFavorites(prev => prev.filter(f => f.id !== id));
        setDeleteConfirm(null);
      }
    } catch (err) {
      console.error('Delete favorite error:', err);
    }
  };

  // 删除播放记录
  const handleDeleteHistory = async (id: string) => {
    try {
      const item = history.find(h => h.id === id);
      if (!item) return;

      const response = await fetch('/api/playrecords', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ source: item.source, id: item.id.replace(`${item.source}_`, '') }),
      });

      if (response.ok) {
        setHistory(prev => prev.filter(h => h.id !== id));
        setDeleteConfirm(null);
      }
    } catch (err) {
      console.error('Delete history error:', err);
    }
  };

  // 清空全部
  const handleClearAll = async () => {
    try {
      if (activeTab === 'favorites') {
        for (const item of favorites) {
          await fetch('/api/favorites', {
            method: 'DELETE',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ source: item.source, id: item.id.replace(`${item.source}_`, '') }),
          });
        }
        setFavorites([]);
      } else {
        for (const item of history) {
          await fetch('/api/playrecords', {
            method: 'DELETE',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ source: item.source, id: item.id.replace(`${item.source}_`, '') }),
          });
        }
        setHistory([]);
      }
      setDeleteConfirm(null);
    } catch (err) {
      console.error('Clear all error:', err);
    }
  };

  // 继续播放
  const handleContinuePlay = (item: PlayRecord) => {
    router.push(`/play?source=${item.source}&id=${item.id.replace(`${item.source}_`, '')}&title=${encodeURIComponent(item.title)}&year=${item.year}`);
  };

  // 格式化时间
  const formatTime = (timestamp: number) => {
    const date = new Date(timestamp);
    return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' });
  };

  // 格式化播放进度
  const formatProgress = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const displayItems = activeTab === 'favorites' ? favorites : history;

  return (
    <main className="min-h-screen bg-[#141414]">
      <TopNav />
      
      {/* Hero Header */}
      <div className="relative h-[40vh] min-h-[300px] overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-b from-pink-600/20 to-[#141414]" />
        <div className="absolute inset-0 bg-gradient-to-t from-[#141414] via-transparent to-black/50" />
        
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="text-center px-4">
            <motion.h1 
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="text-4xl md:text-5xl font-black text-white mb-4"
            >
              我的片单
            </motion.h1>
            <motion.p 
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 }}
              className="text-gray-400"
            >
              管理你的收藏和播放历史
            </motion.p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="sticky top-16 z-40 bg-[#141414]/95 backdrop-blur-md border-b border-gray-800">
        <div className="max-w-[1920px] mx-auto px-4 sm:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              {tabs.map((tab) => {
                const Icon = tab.icon;
                return (
                  <button
                    key={tab.value}
                    onClick={() => setActiveTab(tab.value)}
                    className={`flex items-center gap-2 px-6 py-3 rounded-full text-sm font-medium transition-all ${
                      activeTab === tab.value
                        ? 'bg-[#E50914] text-white'
                        : 'bg-gray-800 text-gray-300 hover:bg-gray-700'
                    }`}
                  >
                    <Icon className="w-4 h-4" />
                    {tab.label}
                    <span className="ml-1 px-2 py-0.5 bg-black/30 rounded-full text-xs">
                      {activeTab === tab.value 
                        ? displayItems.length 
                        : activeTab === 'favorites' ? favorites.length : history.length}
                    </span>
                  </button>
                );
              })}
            </div>

            {displayItems.length > 0 && (
              <button
                onClick={() => setDeleteConfirm('all')}
                className="flex items-center gap-2 px-4 py-2 text-sm text-gray-400 hover:text-red-400 transition-colors"
              >
                <Trash2 className="w-4 h-4" />
                清空全部
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Content Grid */}
      <div className="max-w-[1920px] mx-auto px-4 sm:px-8 py-8">
        {loading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="w-10 h-10 text-[#E50914] animate-spin" />
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center py-20 text-gray-500">
            <AlertCircle className="w-12 h-12 mb-4" />
            <p>{error}</p>
          </div>
        ) : displayItems.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-gray-500">
            {activeTab === 'favorites' ? (
              <>
                <Heart className="w-16 h-16 mb-4 opacity-30" />
                <p className="text-lg mb-2">暂无收藏</p>
                <p className="text-sm">看到喜欢的影片，点击收藏按钮即可添加到此处</p>
              </>
            ) : (
              <>
                <Clock className="w-16 h-16 mb-4 opacity-30" />
                <p className="text-lg mb-2">暂无播放历史</p>
                <p className="text-sm">开始观看影片，播放记录将自动保存</p>
              </>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 sm:gap-6">
            {displayItems.map((item, index) => (
              <div key={item.id} className="relative group">
                {/* 操作按钮 */}
                <div className="absolute -top-2 -right-2 z-30 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button
                    onClick={() => setDeleteConfirm(item.id)}
                    className="w-8 h-8 bg-red-500 rounded-full flex items-center justify-center text-white hover:bg-red-600 transition-colors"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>

                {/* 继续播放按钮（仅历史记录） */}
                {activeTab === 'history' && 'index' in item && (
                  <div className="absolute inset-x-0 bottom-0 z-20 p-2 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button
                      onClick={() => handleContinuePlay(item as PlayRecord)}
                      className="w-full py-2 bg-[#E50914] rounded text-white text-sm font-medium flex items-center justify-center gap-1"
                    >
                      <Play className="w-4 h-4" />
                      继续播放
                    </button>
                  </div>
                )}

                <ContentCard
                  title={item.title}
                  cover={item.cover}
                  year={item.year}
                  type="movie"
                />

                {/* 额外信息 */}
                <div className="mt-2 text-xs text-gray-500">
                  {activeTab === 'favorites' ? (
                    <span>收藏于 {formatTime((item as FavoriteItem).save_time)}</span>
                  ) : (
                    <div className="space-y-1">
                      <span>观看到第 {(item as PlayRecord).index} 集</span>
                      <div className="flex items-center gap-2">
                        <div className="flex-1 h-1 bg-gray-800 rounded-full overflow-hidden">
                          <div 
                            className="h-full bg-[#E50914]"
                            style={{ width: `${Math.min(100, ((item as PlayRecord).play_time / 3600) * 100)}%` }}
                          />
                        </div>
                        <span>{formatProgress((item as PlayRecord).play_time)}</span>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Delete Confirm Modal */}
      <AnimatePresence>
        {deleteConfirm && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4"
          >
            <motion.div
              initial={{ scale: 0.9, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.9, opacity: 0 }}
              className="bg-[#1a1a1a] rounded-xl p-6 max-w-sm w-full"
            >
              <h3 className="text-xl font-bold text-white mb-2">
                确认删除?
              </h3>
              <p className="text-gray-400 mb-6">
                {deleteConfirm === 'all' 
                  ? `确定要清空所有${activeTab === 'favorites' ? '收藏' : '播放历史'}吗? 此操作不可恢复。`
                  : '确定要删除此项吗?'
                }
              </p>
              <div className="flex gap-3">
                <button
                  onClick={() => setDeleteConfirm(null)}
                  className="flex-1 py-3 bg-gray-800 rounded-lg text-white hover:bg-gray-700 transition-colors"
                >
                  取消
                </button>
                <button
                  onClick={() => {
                    if (deleteConfirm === 'all') {
                      handleClearAll();
                    } else if (activeTab === 'favorites') {
                      handleDeleteFavorite(deleteConfirm);
                    } else {
                      handleDeleteHistory(deleteConfirm);
                    }
                  }}
                  className="flex-1 py-3 bg-[#E50914] rounded-lg text-white hover:bg-red-600 transition-colors"
                >
                  删除
                </button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </main>
  );
}
