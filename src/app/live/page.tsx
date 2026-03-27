'use client';

import { Loader2, Radio, Search, Tv, X } from 'lucide-react';
import dynamic from 'next/dynamic';
import Link from 'next/link';
import React, { useCallback, useEffect, useMemo, useState } from 'react';

import { toLogoProxyImageSrc } from '@/lib/image';

import TopNav from '@/components/layout/TopNav';
import SmartImage from '@/components/ui/SmartImage';

interface LiveSource {
  key: string;
  name: string;
  disabled?: boolean;
}

interface LiveChannel {
  id: string;
  tvgId?: string;
  name: string;
  logo?: string;
  group?: string;
  url: string;
}

interface EpgItem {
  start: string;
  end: string;
  title: string;
}

const AllGroupsLabel = '全部';
const ProgramDisplayLimit = 30;

type LiveVideoPlayerProps = {
  url?: string;
  title?: string;
};

const LiveVideoPlayer = dynamic<LiveVideoPlayerProps>(
  () =>
    import('../../components/live/LiveVideoPlayer.jsx').then(
      (module) => module.default
    ),
  {
    ssr: false,
    loading: () => (
      <div className='flex h-full w-full items-center justify-center bg-black text-netflix-gray-400'>
        <Loader2 className='mr-2 h-5 w-5 animate-spin' />
        正在加载直播播放器...
      </div>
    ),
  }
);

function ChannelLogo({
  logo,
  name,
}: {
  logo?: string;
  name: string;
}) {
  const [hidden, setHidden] = useState(false);

  if (!logo || hidden) {
    return <Tv className='w-5 h-5' />;
  }

  return (
    <div className='relative h-5 w-5'>
      <SmartImage
        src={toLogoProxyImageSrc(logo)}
        alt={name}
        fill
        sizes='20px'
        className='rounded object-contain'
        onError={() => setHidden(true)}
      />
    </div>
  );
}

function formatProgramTime(value: string): string {
  if (!value) return '--:--';
  const match = value.match(/(\d{2})(\d{2})(\d{2})/);
  if (!match) return value;
  return `${match[2]}:${match[3]}`;
}

export default function LivePage() {
  const [sources, setSources] = useState<LiveSource[]>([]);
  const [channels, setChannels] = useState<LiveChannel[]>([]);
  const [epg, setEpg] = useState<EpgItem[]>([]);

  const [loadingSources, setLoadingSources] = useState(true);
  const [loadingChannels, setLoadingChannels] = useState(false);
  const [loadingEpg, setLoadingEpg] = useState(false);

  const [selectedSource, setSelectedSource] = useState<string>('');
  const [selectedChannel, setSelectedChannel] = useState<LiveChannel | null>(
    null
  );
  const [selectedGroup, setSelectedGroup] = useState<string>(AllGroupsLabel);
  const [channelQuery, setChannelQuery] = useState('');
  const [error, setError] = useState<string | null>(null);

  const fetchSources = useCallback(async () => {
    setLoadingSources(true);
    setError(null);
    try {
      const response = await fetch('/api/live/sources');
      if (!response.ok) {
        throw new Error('获取直播源失败');
      }
      const data = await response.json();
      const sourceList = Array.isArray(data?.data) ? data.data : [];
      const enabled = (sourceList as LiveSource[]).filter((item) => !item.disabled);
      setSources(enabled);
      setSelectedSource(enabled[0]?.key || '');
    } catch {
      setError('加载直播源失败，请稍后重试');
    } finally {
      setLoadingSources(false);
    }
  }, []);

  useEffect(() => {
    void fetchSources();
  }, [fetchSources]);

  useEffect(() => {
    const fetchChannels = async () => {
      if (!selectedSource) return;

      setLoadingChannels(true);
      setError(null);
      setSelectedChannel(null);
      setSelectedGroup(AllGroupsLabel);
      setChannelQuery('');
      setChannels([]);
      setEpg([]);

      try {
        const response = await fetch(
          `/api/live/channels?source=${encodeURIComponent(selectedSource)}`
        );
        if (!response.ok) {
          throw new Error('获取频道失败');
        }
        const data = await response.json();
        const channelList = Array.isArray(data?.data) ? data.data : [];
        setChannels(channelList);
        if (channelList[0]) {
          setSelectedChannel(channelList[0]);
        }
      } catch {
        setError('加载频道失败，请稍后重试');
      } finally {
        setLoadingChannels(false);
      }
    };

    fetchChannels();
  }, [selectedSource]);

  useEffect(() => {
    const fetchEpg = async () => {
      if (!selectedSource || !selectedChannel?.tvgId) {
        setEpg([]);
        return;
      }

      setLoadingEpg(true);
      try {
        const response = await fetch(
          `/api/live/epg?source=${encodeURIComponent(
            selectedSource
          )}&tvgId=${encodeURIComponent(selectedChannel.tvgId)}`
        );
        if (!response.ok) {
          throw new Error('获取节目单失败');
        }
        const data = await response.json();
        const programs = Array.isArray(data?.data?.programs)
          ? data.data.programs
          : [];
        setEpg(programs);
      } catch {
        setEpg([]);
      } finally {
        setLoadingEpg(false);
      }
    };

    fetchEpg();
  }, [selectedSource, selectedChannel]);

  const groupOptions = useMemo(() => {
    const groupSet = new Set<string>();
    channels.forEach((channel) => {
      if (channel.group) {
        groupSet.add(channel.group);
      }
    });
    return [AllGroupsLabel, ...Array.from(groupSet)];
  }, [channels]);

  const filteredChannels = useMemo(() => {
    const normalizedQuery = channelQuery.trim().toLowerCase();
    return channels.filter((channel) => {
      const matchesGroup =
        selectedGroup === AllGroupsLabel || channel.group === selectedGroup;
      const matchesQuery =
        !normalizedQuery ||
        channel.name.toLowerCase().includes(normalizedQuery) ||
        (channel.group || '').toLowerCase().includes(normalizedQuery);
      return matchesGroup && matchesQuery;
    });
  }, [channelQuery, channels, selectedGroup]);

  const playerUrl = useMemo(() => {
    if (!selectedChannel?.url) return '';
    if (selectedChannel.url.includes('.m3u8')) {
      return `/api/proxy/m3u8?url=${encodeURIComponent(selectedChannel.url)}`;
    }
    return selectedChannel.url;
  }, [selectedChannel]);

  useEffect(() => {
    if (filteredChannels.length === 0) {
      setSelectedChannel(null);
      return;
    }

    const activeStillVisible = filteredChannels.some(
      (channel) => channel.id === selectedChannel?.id
    );
    if (!activeStillVisible) {
      setSelectedChannel(filteredChannels[0]);
    }
  }, [filteredChannels, selectedChannel?.id]);

  return (
    <main className='min-h-screen bg-netflix-black'>
      <TopNav />

      <div className='pt-24 pb-20 px-4 sm:px-8'>
        <div className='max-w-[1920px] mx-auto'>
          <div className='flex items-center gap-3 mb-8'>
            <Radio className='w-8 h-8 text-netflix-red' />
            <h1 className='text-3xl font-bold text-white'>电视直播</h1>
          </div>

          {error && (
            <div className='mb-6 bg-red-500/10 border border-red-500/20 rounded-lg p-4 text-red-400'>
              {error}
            </div>
          )}

          <div className='grid grid-cols-1 lg:grid-cols-3 gap-8'>
            <div className='lg:col-span-2 space-y-4'>
              {!loadingSources && sources.length === 0 ? (
                <div className='aspect-video rounded-xl border border-white/10 bg-netflix-surface p-6'>
                  <div className='flex h-full flex-col items-center justify-center text-center'>
                    <Radio className='mb-4 h-14 w-14 text-netflix-gray-600' />
                    <h2 className='text-xl font-semibold text-white'>
                      还没有可用直播源
                    </h2>
                    <p className='mt-2 max-w-md text-sm text-netflix-gray-400'>
                      当前环境没有配置直播 M3U 源，所以频道列表和节目单不会显示。先去管理后台补直播源，再回来即可直接播放。
                    </p>
                    <div className='mt-5 flex flex-wrap items-center justify-center gap-3'>
                      <Link
                        href='/admin'
                        className='rounded-full bg-netflix-red px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-netflix-red-hover'
                      >
                        去管理后台
                      </Link>
                      <button
                        type='button'
                        onClick={() => {
                          void fetchSources();
                        }}
                        className='rounded-full border border-zinc-700 px-4 py-2 text-sm text-zinc-200 transition-colors hover:border-zinc-500 hover:text-white'
                      >
                        重新检测
                      </button>
                    </div>
                  </div>
                </div>
              ) : selectedChannel ? (
                <div className='aspect-video bg-netflix-surface rounded-xl overflow-hidden'>
                  <LiveVideoPlayer url={playerUrl} title={selectedChannel.name} />
                </div>
              ) : (
                <div className='aspect-video bg-netflix-surface rounded-xl flex items-center justify-center'>
                  <div className='text-center'>
                    <Radio className='w-16 h-16 text-netflix-gray-600 mx-auto mb-4' />
                    <p className='text-netflix-gray-400'>请选择频道</p>
                  </div>
                </div>
              )}

              <div className='bg-netflix-surface rounded-xl p-4'>
                <h2 className='text-lg font-bold text-white mb-3'>节目单</h2>
                {loadingEpg ? (
                  <div className='flex items-center justify-center py-6'>
                    <Loader2 className='w-6 h-6 text-netflix-red animate-spin' />
                  </div>
                ) : epg.length > 0 ? (
                  <div className='space-y-2 max-h-64 overflow-y-auto pr-2'>
                    {epg.slice(0, ProgramDisplayLimit).map((item, index) => (
                      <div
                        key={`${item.start}-${item.title}-${index}`}
                        className='flex items-center gap-3 px-3 py-2 rounded-lg bg-netflix-gray-900/60'
                      >
                        <span className='text-xs text-netflix-gray-400 min-w-20'>
                          {formatProgramTime(item.start)} -{' '}
                          {formatProgramTime(item.end)}
                        </span>
                        <span className='text-sm text-netflix-gray-200'>
                          {item.title}
                        </span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className='text-netflix-gray-500 text-sm'>
                    该频道暂无节目单
                  </p>
                )}
              </div>
            </div>

            <div className='bg-netflix-surface rounded-xl p-4'>
              <h2 className='text-lg font-bold text-white mb-4'>频道列表</h2>

              {loadingSources ? (
                <div className='flex items-center justify-center py-8'>
                  <Loader2 className='w-8 h-8 text-netflix-red animate-spin' />
                </div>
              ) : sources.length === 0 ? (
                <div className='rounded-lg border border-white/10 bg-netflix-gray-900/50 p-4 text-sm text-netflix-gray-400'>
                  暂未配置直播源。前往管理后台添加后，这里会自动显示来源与频道列表。
                </div>
              ) : (
                <>
                  {sources.length > 1 && (
                    <div className='mb-4'>
                      <p className='text-xs text-netflix-gray-500 mb-2'>
                        直播源
                      </p>
                      <div className='flex flex-wrap gap-2'>
                        {sources.map((item) => (
                          <button
                            key={item.key}
                            onClick={() => setSelectedSource(item.key)}
                            className={`px-3 py-1.5 rounded-full text-xs transition-colors ${
                              selectedSource === item.key
                                ? 'bg-netflix-red text-white'
                                : 'bg-netflix-gray-800 text-netflix-gray-300 hover:bg-netflix-gray-700'
                            }`}
                          >
                            {item.name}
                          </button>
                        ))}
                      </div>
                    </div>
                  )}

                  {groupOptions.length > 1 && (
                    <div className='mb-4'>
                      <p className='text-xs text-netflix-gray-500 mb-2'>分组</p>
                      <div className='flex flex-wrap gap-2'>
                        {groupOptions.map((group) => (
                          <button
                            key={group}
                            onClick={() => setSelectedGroup(group)}
                            className={`px-3 py-1.5 rounded-full text-xs transition-colors ${
                              selectedGroup === group
                                ? 'bg-netflix-red text-white'
                                : 'bg-netflix-gray-800 text-netflix-gray-300 hover:bg-netflix-gray-700'
                            }`}
                          >
                            {group}
                          </button>
                        ))}
                      </div>
                    </div>
                  )}

                  <div className='mb-4'>
                    <div className='mb-2 flex items-center justify-between'>
                      <p className='text-xs text-netflix-gray-500'>频道搜索</p>
                      {channelQuery ? (
                        <button
                          type='button'
                          onClick={() => setChannelQuery('')}
                          className='inline-flex items-center gap-1 text-xs text-netflix-gray-400 transition-colors hover:text-white'
                        >
                          <X className='h-3.5 w-3.5' />
                          清空
                        </button>
                      ) : null}
                    </div>
                    <div className='flex items-center rounded-lg border border-netflix-gray-800 bg-netflix-gray-900/70 px-3 py-2'>
                      <Search className='h-4 w-4 text-netflix-gray-500' />
                      <input
                        value={channelQuery}
                        onChange={(event) => setChannelQuery(event.target.value)}
                        placeholder='搜索频道名或分组'
                        className='ml-2 w-full bg-transparent text-sm text-white outline-none placeholder:text-netflix-gray-500'
                      />
                    </div>
                    <p className='mt-2 text-xs text-netflix-gray-500'>
                      当前显示 {filteredChannels.length} / {channels.length} 个频道
                    </p>
                  </div>

                  {selectedChannel ? (
                    <div className='mb-4 rounded-lg border border-white/5 bg-netflix-gray-900/50 p-3'>
                      <div className='flex items-center gap-3'>
                        <ChannelLogo
                          logo={selectedChannel.logo}
                          name={selectedChannel.name}
                        />
                        <div className='min-w-0'>
                          <p className='truncate text-sm font-semibold text-white'>
                            {selectedChannel.name}
                          </p>
                          <p className='text-xs text-netflix-gray-500'>
                            当前来源：{sources.find((item) => item.key === selectedSource)?.name || selectedSource}
                            {selectedChannel.group ? ` · ${selectedChannel.group}` : ''}
                          </p>
                        </div>
                      </div>
                    </div>
                  ) : null}

                  {loadingChannels ? (
                    <div className='flex items-center justify-center py-8'>
                      <Loader2 className='w-8 h-8 text-netflix-red animate-spin' />
                    </div>
                  ) : (
                    <div className='space-y-2 max-h-[55vh] overflow-y-auto pr-1'>
                      {filteredChannels.map((channel) => (
                        <button
                          key={channel.id}
                          onClick={() => setSelectedChannel(channel)}
                          className={`w-full flex items-center gap-3 p-3 rounded-lg transition-colors ${
                            selectedChannel?.id === channel.id
                              ? 'bg-netflix-red text-white'
                              : 'hover:bg-netflix-gray-800 text-netflix-gray-300'
                          }`}
                        >
                          <ChannelLogo logo={channel.logo} name={channel.name} />
                          <span className='truncate'>{channel.name}</span>
                          <span className='ml-auto text-xs opacity-70'>
                            {channel.group || '未分组'}
                          </span>
                        </button>
                      ))}

                      {filteredChannels.length === 0 && (
                        <div className='py-4 text-center'>
                          <p className='text-sm text-netflix-gray-500'>
                            {channelQuery || selectedGroup !== AllGroupsLabel
                              ? '没有匹配当前筛选条件的频道'
                              : '暂无频道'}
                          </p>
                          {channelQuery || selectedGroup !== AllGroupsLabel ? (
                            <button
                              type='button'
                              onClick={() => {
                                setChannelQuery('');
                                setSelectedGroup(AllGroupsLabel);
                              }}
                              className='mt-2 text-xs text-netflix-red hover:underline'
                            >
                              清空筛选
                            </button>
                          ) : null}
                        </div>
                      )}
                    </div>
                  )}
                </>
              )}
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}
