'use client';

import { motion } from 'framer-motion';
import {
  ChevronRight,
  Clock,
  Heart,
  History,
  Loader2,
  LogOut,
  Settings,
  Star,
  Trash2,
  User,
} from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import React, { useEffect, useMemo, useState } from 'react';

import { clearAllFavorites, clearAllPlayRecords } from '@/lib/db.client';
import { toProxyImageSrc } from '@/lib/image';
import { useFastNavigation } from '@/lib/navigation-feedback';

import TopNav from '@/components/layout/TopNav';
import ContentCard from '@/components/ui/ContentCard';
import SmartImage from '@/components/ui/SmartImage';

const tabs = [
  { id: 'favorites', label: '我的片单', icon: Heart },
  { id: 'history', label: '观看历史', icon: History },
  { id: 'settings', label: '账号设置', icon: Settings },
] as const;

const ClearConfirmTimeoutMs = 3000;

interface FavoriteItem {
  id: string;
  source: string;
  source_name?: string;
  title: string;
  cover: string;
  year: string;
}

interface HistoryItem {
  id: string;
  source: string;
  source_name?: string;
  title: string;
  cover: string;
  year: string;
  index?: number;
  play_time?: number;
  total_time?: number;
  last_play_time?: number;
}

interface FavoritesBootstrapResponse {
  username?: string;
  favorites?: FavoriteItem[];
  history?: HistoryItem[];
}

function getErrorMessage(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

function toHours(seconds: number): number {
  return Math.round((seconds / 3600) * 10) / 10;
}

export default function FavoritesPage() {
  const router = useRouter();
  const { navigate, prefetchHref } = useFastNavigation();
  const searchParams = useSearchParams();
  const defaultTab = (searchParams.get('tab') ||
    'favorites') as (typeof tabs)[number]['id'];
  const [activeTab, setActiveTab] = useState<(typeof tabs)[number]['id']>(
    tabs.some((tab) => tab.id === defaultTab) ? defaultTab : 'favorites'
  );
  const [loading, setLoading] = useState(true);
  const [username, setUsername] = useState('当前用户');
  const [favorites, setFavorites] = useState<FavoriteItem[]>([]);
  const [history, setHistory] = useState<HistoryItem[]>([]);
  const [error, setError] = useState<string | null>(null);

  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [savingPassword, setSavingPassword] = useState(false);
  const [passwordMessage, setPasswordMessage] = useState<string | null>(null);
  const [pendingClearSection, setPendingClearSection] = useState<
    'favorites' | 'history' | null
  >(null);

  const loadData = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch('/api/favorites/bootstrap');
      if (!response.ok) {
        throw new Error('favorites bootstrap request failed');
      }

      const data = (await response.json()) as FavoritesBootstrapResponse;
      if (data.username) {
        setUsername(data.username);
      }
      setFavorites(
        Array.isArray(data.favorites)
          ? data.favorites.map((item) => ({
              id: item.id,
              source: item.source,
              source_name: item.source_name || '',
              title: item.title || '未知标题',
              cover: item.cover || '/placeholder-poster.svg',
              year: item.year || '',
            }))
          : []
      );
      setHistory(
        Array.isArray(data.history)
          ? data.history.map((item) => ({
              id: item.id,
              source: item.source,
              source_name: item.source_name || '',
              title: item.title || '未知标题',
              cover: item.cover || '/placeholder-poster.svg',
              year: item.year || '',
              index: item.index || 0,
              play_time: item.play_time || 0,
              total_time: item.total_time || 0,
              last_play_time: item.last_play_time || 0,
            }))
          : []
      );
    } catch {
      setError('加载数据失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadData();
  }, []);

  useEffect(() => {
    const tab = searchParams.get('tab');
    if (tab && tabs.some((item) => item.id === tab)) {
      setActiveTab(tab as (typeof tabs)[number]['id']);
    }
  }, [searchParams]);

  useEffect(() => {
    if (!pendingClearSection) {
      return;
    }

    const timer = window.setTimeout(() => {
      setPendingClearSection(null);
    }, ClearConfirmTimeoutMs);

    return () => window.clearTimeout(timer);
  }, [pendingClearSection]);

  const handleLogout = async () => {
    await fetch('/api/logout', { method: 'POST' });
    router.push('/login');
  };

  const removeFavorite = async (key: string) => {
    await fetch(`/api/favorites?key=${encodeURIComponent(key)}`, {
      method: 'DELETE',
    });
    setFavorites((prev) => prev.filter((item) => item.id !== key));
  };

  const removeHistory = async (key: string) => {
    await fetch(`/api/playrecords?key=${encodeURIComponent(key)}`, {
      method: 'DELETE',
    });
    setHistory((prev) => prev.filter((item) => item.id !== key));
  };

  const buildHistoryHref = (item: HistoryItem) => {
    const [derivedSource, derivedId] = item.id.split('+');
    const source = item.source || derivedSource;
    const id = derivedId || item.id;
    if (!source || !id) {
      return '';
    }

    const params = new URLSearchParams({
      source,
      id,
      title: item.title,
    });
    if (typeof item.index === 'number' && item.index > 0) {
      params.set('ep', String(item.index));
    }
    return `/play?${params.toString()}`;
  };

  const handleOpenHistoryItem = (item: HistoryItem) => {
    const href = buildHistoryHref(item);
    if (!href) {
      return;
    }
    navigate(href);
  };

  const handleClearSection = async (section: 'favorites' | 'history') => {
    if (pendingClearSection !== section) {
      setPendingClearSection(section);
      return;
    }

    setError(null);
    try {
      if (section === 'favorites') {
        await clearAllFavorites();
        setFavorites([]);
      } else {
        await clearAllPlayRecords();
        setHistory([]);
      }
      setPendingClearSection(null);
    } catch {
      setError(section === 'favorites' ? '清空收藏失败' : '清空观看记录失败');
    }
  };

  const handleChangePassword = async () => {
    setPasswordMessage(null);
    if (!newPassword.trim()) {
      setPasswordMessage('请输入新密码');
      return;
    }

    setSavingPassword(true);
    try {
      const response = await fetch('/api/change-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          oldPassword,
          newPassword,
        }),
      });
      const data = await response.json();
      if (!response.ok || data?.error) {
        throw new Error(data?.error || '修改失败');
      }
      setPasswordMessage('密码修改成功');
      setOldPassword('');
      setNewPassword('');
    } catch (changeError) {
      setPasswordMessage(getErrorMessage(changeError, '密码修改失败'));
    } finally {
      setSavingPassword(false);
    }
  };

  const totalWatchHours = useMemo(() => {
    const totalSeconds = history.reduce(
      (sum, item) => sum + (item.play_time || 0),
      0
    );
    return toHours(totalSeconds);
  }, [history]);

  const completedCount = useMemo(
    () =>
      history.filter(
        (item) =>
          (item.total_time || 0) > 0 &&
          (item.play_time || 0) >= (item.total_time || 0)
      ).length,
    [history]
  );

  return (
    <main className='min-h-screen bg-netflix-black'>
      <TopNav />

      <div className='pt-24 pb-20 px-4 sm:px-8'>
        <div className='max-w-6xl mx-auto'>
          {error && (
            <div className='mb-6 bg-red-500/10 border border-red-500/20 rounded-lg p-4 text-red-400'>
              {error}
            </div>
          )}

          <div className='bg-gradient-to-br from-netflix-surface to-netflix-gray-900 rounded-2xl p-6 sm:p-8 mb-8'>
            <div className='flex flex-col sm:flex-row items-center gap-6'>
              <div className='w-24 h-24 rounded-full bg-gradient-to-br from-netflix-red to-orange-500 flex items-center justify-center'>
                <User className='w-12 h-12 text-white' />
              </div>

              <div className='text-center sm:text-left'>
                <h1 className='text-2xl font-bold text-white mb-1'>
                  {username}
                </h1>
                <p className='text-netflix-gray-400'>已登录用户</p>

                <div className='flex items-center gap-6 mt-4 text-sm'>
                  <div className='flex items-center gap-2 text-netflix-gray-300'>
                    <Clock className='w-4 h-4' />
                    观影时长：{totalWatchHours}小时
                  </div>
                  <div className='flex items-center gap-2 text-netflix-gray-300'>
                    <Heart className='w-4 h-4' />
                    收藏：{favorites.length}部
                  </div>
                  <div className='flex items-center gap-2 text-netflix-gray-300'>
                    <Star className='w-4 h-4' />
                    看完：{completedCount}部
                  </div>
                </div>
              </div>

              <button
                onClick={handleLogout}
                className='sm:ml-auto flex items-center gap-2 px-6 py-3 bg-netflix-gray-800 text-white rounded-lg hover:bg-netflix-red transition-colors'
              >
                <LogOut className='w-5 h-5' />
                退出登录
              </button>
            </div>
          </div>

          <div className='flex items-center gap-2 mb-8 border-b border-netflix-gray-800 pb-4'>
            {tabs.map((tab) => {
              const Icon = tab.icon;
              return (
                <button
                  key={tab.id}
                  onClick={() => {
                    setActiveTab(tab.id);
                    router.replace(`/favorites?tab=${tab.id}`);
                  }}
                  className={`flex items-center gap-2 px-6 py-3 rounded-lg font-medium transition-colors ${
                    activeTab === tab.id
                      ? 'bg-netflix-red text-white'
                      : 'text-netflix-gray-300 hover:bg-netflix-gray-800'
                  }`}
                >
                  <Icon className='w-5 h-5' />
                  {tab.label}
                </button>
              );
            })}
          </div>

          {loading ? (
            <div className='flex items-center justify-center py-20'>
              <Loader2 className='w-10 h-10 text-netflix-red animate-spin' />
            </div>
          ) : (
            <div>
              {activeTab === 'favorites' && (
                <div>
                  <div className='mb-6 flex items-center justify-between gap-3'>
                    <h2 className='text-xl font-bold text-white'>我的收藏</h2>
                    {favorites.length > 0 ? (
                      <button
                        onClick={() => handleClearSection('favorites')}
                        className={`rounded-lg px-4 py-2 text-sm transition-colors ${
                          pendingClearSection === 'favorites'
                            ? 'bg-red-600 text-white hover:bg-red-500'
                            : 'bg-netflix-gray-800 text-netflix-gray-300 hover:bg-netflix-gray-700 hover:text-white'
                        }`}
                      >
                        {pendingClearSection === 'favorites'
                          ? '确认清空收藏'
                          : '清空收藏'}
                      </button>
                    ) : null}
                  </div>
                  {favorites.length > 0 ? (
                    <div className='grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4 sm:gap-6'>
                      {favorites.map((item, index) => (
                        <motion.div
                          key={item.id}
                          initial={{ opacity: 0, y: 20 }}
                          animate={{ opacity: 1, y: 0 }}
                          transition={{ delay: index * 0.04 }}
                          className='relative'
                        >
                          <button
                            onClick={() => removeFavorite(item.id)}
                            className='absolute z-20 top-2 right-2 w-7 h-7 rounded-full bg-black/70 text-white flex items-center justify-center hover:bg-netflix-red transition-colors'
                          >
                            <Trash2 className='w-4 h-4' />
                          </button>
                          <ContentCard
                            id={item.id.split('+')[1] || item.id}
                            source={item.source}
                            title={item.title}
                            cover={item.cover}
                            year={item.year}
                          />
                        </motion.div>
                      ))}
                    </div>
                  ) : (
                    <div className='text-center py-20'>
                      <p className='text-netflix-gray-500'>暂无收藏影片</p>
                    </div>
                  )}
                </div>
              )}

              {activeTab === 'history' && (
                <div>
                  <div className='mb-6 flex items-center justify-between gap-3'>
                    <h2 className='text-xl font-bold text-white'>观看历史</h2>
                    {history.length > 0 ? (
                      <button
                        onClick={() => handleClearSection('history')}
                        className={`rounded-lg px-4 py-2 text-sm transition-colors ${
                          pendingClearSection === 'history'
                            ? 'bg-red-600 text-white hover:bg-red-500'
                            : 'bg-netflix-gray-800 text-netflix-gray-300 hover:bg-netflix-gray-700 hover:text-white'
                        }`}
                      >
                        {pendingClearSection === 'history'
                          ? '确认清空记录'
                          : '清空记录'}
                      </button>
                    ) : null}
                  </div>
                  {history.length > 0 ? (
                    <div className='space-y-4'>
                      {history.map((item) => {
                        const progressPercent =
                          item.total_time && item.total_time > 0
                            ? Math.min(
                                100,
                                Math.round(
                                  ((item.play_time || 0) / item.total_time) *
                                    100
                                )
                              )
                            : 0;
                        return (
                          <div
                            key={item.id}
                            onClick={() => handleOpenHistoryItem(item)}
                            onPointerEnter={() => {
                              const href = buildHistoryHref(item);
                              if (href) {
                                prefetchHref(href);
                              }
                            }}
                            className='group flex cursor-pointer gap-4 rounded-lg bg-netflix-surface p-4 transition-colors hover:bg-netflix-surface-hover'
                          >
                            <div className='relative w-32 sm:w-40 aspect-video rounded overflow-hidden'>
                              <SmartImage
                                src={toProxyImageSrc(item.cover)}
                                alt={item.title}
                                fill
                                sizes='(max-width: 640px) 128px, 160px'
                                className='w-full h-full object-cover'
                              />
                              <div className='absolute bottom-0 left-0 right-0 h-1 bg-netflix-gray-800'>
                                <div
                                  className='h-full bg-netflix-red'
                                  style={{ width: `${progressPercent}%` }}
                                />
                              </div>
                            </div>

                            <div className='flex-1 min-w-0'>
                              <h3 className='text-white font-bold mb-1 group-hover:text-netflix-red transition-colors truncate'>
                                {item.title}
                              </h3>
                              <p className='text-netflix-gray-400 text-sm mb-2'>
                                {item.year}{' '}
                                {item.index ? `· 第${item.index}集` : ''}
                              </p>
                              <p className='text-netflix-gray-500 text-sm'>
                                观看到 {progressPercent}%
                              </p>
                            </div>

                            <div className='flex flex-col items-end justify-center gap-3'>
                              <button
                                onClick={(event) => {
                                  event.stopPropagation();
                                  handleOpenHistoryItem(item);
                                }}
                                className='rounded-lg bg-netflix-gray-800 px-3 py-2 text-sm text-white transition-colors hover:bg-netflix-red'
                              >
                                继续播放
                              </button>
                              <button
                                onClick={(event) => {
                                  event.stopPropagation();
                                  void removeHistory(item.id);
                                }}
                                className='self-center text-netflix-gray-500 transition-colors hover:text-netflix-red'
                              >
                                <Trash2 className='w-5 h-5' />
                              </button>
                            </div>
                            <ChevronRight className='self-center w-5 h-5 text-netflix-gray-600' />
                          </div>
                        );
                      })}
                    </div>
                  ) : (
                    <div className='text-center py-20'>
                      <p className='text-netflix-gray-500'>暂无观看记录</p>
                    </div>
                  )}
                </div>
              )}

              {activeTab === 'settings' && (
                <div className='max-w-2xl'>
                  <h2 className='text-xl font-bold text-white mb-6'>
                    账号设置
                  </h2>

                  <div className='space-y-6'>
                    <div className='p-4 bg-netflix-surface rounded-lg'>
                      <h3 className='text-white font-bold mb-4'>修改密码</h3>
                      <div className='space-y-4'>
                        <input
                          type='password'
                          placeholder='当前密码'
                          value={oldPassword}
                          onChange={(e) => setOldPassword(e.target.value)}
                          className='w-full px-4 py-3 bg-netflix-gray-800 border border-netflix-gray-700 rounded-lg text-white placeholder-netflix-gray-500 focus:outline-none focus:border-netflix-red'
                        />
                        <input
                          type='password'
                          placeholder='新密码'
                          value={newPassword}
                          onChange={(e) => setNewPassword(e.target.value)}
                          className='w-full px-4 py-3 bg-netflix-gray-800 border border-netflix-gray-700 rounded-lg text-white placeholder-netflix-gray-500 focus:outline-none focus:border-netflix-red'
                        />
                        <button
                          onClick={handleChangePassword}
                          disabled={savingPassword}
                          className='px-6 py-3 bg-netflix-red text-white font-bold rounded-lg hover:bg-netflix-red-hover transition-colors disabled:opacity-60'
                        >
                          {savingPassword ? '保存中...' : '保存修改'}
                        </button>
                        {passwordMessage && (
                          <p className='text-sm text-netflix-gray-300'>
                            {passwordMessage}
                          </p>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </main>
  );
}
