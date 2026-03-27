'use client';

import Artplayer from 'artplayer';
import Hls from 'hls.js';
import { Loader2 } from 'lucide-react';
import React, { useCallback, useEffect, useRef } from 'react';

import { SkipConfig } from '@/lib/db.client';
import { toImageSrc } from '@/lib/image';

import {
  CustomHlsJsLoader,
  formatTime,
  PlayRecordSaveIntervalMs,
  SkipCheckIntervalMs,
  WakeLockSentinel,
} from '@/components/play/play-utils';

export interface PlayVideoPlayerProps {
  loading: boolean;
  streamUrl: string;
  poster?: string;
  retryToken: number;
  blockAdEnabled: boolean;
  onBlockAdChange: (value: boolean) => void;
  skipConfig: SkipConfig;
  initialResumeTime: number | null;
  desiredQuality: number;
  onPersistProgress: (playTime: number, totalTime: number) => void;
  onSkipConfigChange: (nextConfig: SkipConfig) => Promise<void>;
  onPlayerErrorChange: (message: string | null) => void;
  onQualityListChange: (list: Array<{ index: number; label: string }>) => void;
  onActiveQualityChange: (index: number) => void;
}

export default function PlayVideoPlayer({
  loading,
  streamUrl,
  poster,
  retryToken,
  blockAdEnabled,
  onBlockAdChange,
  skipConfig,
  initialResumeTime,
  desiredQuality,
  onPersistProgress,
  onSkipConfigChange,
  onPlayerErrorChange,
  onQualityListChange,
  onActiveQualityChange,
}: PlayVideoPlayerProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const artRef = useRef<Artplayer | null>(null);
  const hlsRef = useRef<Hls | null>(null);
  const wakeLockRef = useRef<WakeLockSentinel | null>(null);
  const saveIntervalRef = useRef<number | null>(null);
  const skipCheckIntervalRef = useRef<number | null>(null);
  const lastSkipCheckRef = useRef(0);
  const hasResumedRef = useRef(false);
  const skipConfigRef = useRef(skipConfig);

  useEffect(() => {
    skipConfigRef.current = skipConfig;
  }, [skipConfig]);

  const requestWakeLock = useCallback(async () => {
    try {
      if ('wakeLock' in navigator) {
        const wakeLockNavigator = navigator as Navigator & {
          wakeLock?: { request: (type: 'screen') => Promise<WakeLockSentinel> };
        };
        wakeLockRef.current =
          (await wakeLockNavigator.wakeLock?.request('screen')) || null;
      }
    } catch {
      return;
    }
  }, []);

  const releaseWakeLock = useCallback(async () => {
    try {
      if (!wakeLockRef.current) return;
      await wakeLockRef.current.release();
      wakeLockRef.current = null;
    } catch {
      return;
    }
  }, []);

  const persistCurrentProgress = useCallback(() => {
    const art = artRef.current;
    if (!art) return;

    const playTime = Math.floor(art.currentTime || 0);
    const totalTime = Math.floor(art.duration || 0);
    if (playTime <= 0) return;

    onPersistProgress(playTime, totalTime);
  }, [onPersistProgress]);

  const tryResumePlayback = useCallback(
    (video: HTMLVideoElement, notice?: { show?: string }) => {
      if (
        hasResumedRef.current ||
        !initialResumeTime ||
        initialResumeTime <= 0
      ) {
        return;
      }

      video.currentTime = initialResumeTime;
      if (notice) {
        notice.show = `已恢复至 ${formatTime(initialResumeTime)}`;
      }
      hasResumedRef.current = true;
    },
    [initialResumeTime]
  );

  const checkAndSkip = useCallback(() => {
    const art = artRef.current;
    if (!art) return;

    const config = skipConfigRef.current;
    if (!config.enable) return;

    const now = Date.now();
    if (now - lastSkipCheckRef.current < SkipCheckIntervalMs) return;
    lastSkipCheckRef.current = now;

    const currentTime = art.currentTime || 0;
    const duration = art.duration || 0;

    if (config.intro_time > 0 && currentTime < config.intro_time) {
      if (config.intro_time - currentTime < 2) {
        art.currentTime = config.intro_time;
        art.notice.show = `已跳过片头 ${formatTime(config.intro_time)}`;
      }
    }

    if (config.outro_time < 0 && duration > 0) {
      const outroStartTime = duration + config.outro_time;
      if (currentTime >= outroStartTime && currentTime < duration - 2) {
        art.currentTime = duration - 1;
        art.notice.show = '已跳过片尾';
      }
    }
  }, []);

  useEffect(() => {
    if (!containerRef.current || !streamUrl) {
      onPlayerErrorChange(null);
      onQualityListChange([]);
      onActiveQualityChange(-1);
      return;
    }

    hasResumedRef.current = false;
    onPlayerErrorChange(null);
    onQualityListChange([]);
    onActiveQualityChange(-1);

    const art: Artplayer = new Artplayer({
      container: containerRef.current,
      url: streamUrl,
      type: 'm3u8',
      poster: toImageSrc(poster, '/placeholder-backdrop.svg'),
      autoplay: false,
      autoSize: true,
      setting: true,
      playbackRate: true,
      pip: true,
      fullscreen: true,
      fullscreenWeb: true,
      miniProgressBar: true,
      mutex: true,
      theme: '#E50914',
      customType: {
        m3u8: (video: HTMLVideoElement, url: string) => {
          if (video.canPlayType('application/vnd.apple.mpegurl')) {
            video.src = url;
            video.addEventListener(
              'loadedmetadata',
              () => tryResumePlayback(video, art.notice),
              { once: true }
            );
            return;
          }

          if (!Hls.isSupported()) {
            onPlayerErrorChange('当前浏览器不支持播放此视频流');
            art.notice.show = '当前浏览器不支持播放此视频流';
            return;
          }

          const hls = new Hls({
            enableWorker: true,
            ...(blockAdEnabled ? { loader: CustomHlsJsLoader } : {}),
          });
          hlsRef.current = hls;
          hls.loadSource(url);
          hls.attachMedia(video);

          hls.on(Hls.Events.MANIFEST_PARSED, () => {
            const levels = hls.levels || [];
            const qualities = levels.map((level, index) => ({
              index,
              label:
                level.height && level.height > 0
                  ? `${level.height}p`
                  : level.bitrate && level.bitrate > 0
                  ? `${Math.round(level.bitrate / 1000)}kbps`
                  : `线路${index + 1}`,
            }));

            onQualityListChange(qualities);
            onActiveQualityChange(hls.currentLevel);
            tryResumePlayback(video, art.notice);
          });

          hls.on(Hls.Events.LEVEL_SWITCHED, (_event, data) => {
            onActiveQualityChange(
              typeof data?.level === 'number' ? data.level : -1
            );
          });

          hls.on(Hls.Events.ERROR, (_event, data) => {
            if (!data?.fatal) return;
            onPlayerErrorChange('播放流异常，已尝试恢复，可点击“重试播放”');
          });
        },
      },
      settings: [
        {
          name: '跳过片头片尾',
          html: '跳过片头片尾',
          switch: skipConfig.enable,
          onSwitch: (item) => {
            const currentSwitch = Boolean(
              (item as { switch?: boolean }).switch
            );
            const nextConfig = {
              ...skipConfigRef.current,
              enable: !currentSwitch,
            };
            void onSkipConfigChange(nextConfig);
            return !currentSwitch;
          },
        },
        {
          name: '去广告',
          html: '去广告',
          switch: blockAdEnabled,
          onSwitch: (item) => {
            const nextValue = !(item as { switch?: boolean }).switch;
            onBlockAdChange(nextValue);
            return nextValue;
          },
        },
        {
          name: '设置片头',
          html: '设置片头',
          tooltip:
            skipConfig.intro_time === 0
              ? '设置片头时间'
              : formatTime(skipConfig.intro_time),
          onClick: () => {
            const currentTime = artRef.current?.currentTime || 0;
            if (currentTime <= 0) return;
            const nextConfig = {
              ...skipConfigRef.current,
              intro_time: currentTime,
              enable: true,
            };
            void onSkipConfigChange(nextConfig);
            art.notice.show = `片头已设置为 ${formatTime(currentTime)}`;
            return formatTime(currentTime);
          },
        },
        {
          name: '设置片尾',
          html: '设置片尾',
          tooltip:
            skipConfig.outro_time >= 0
              ? '设置片尾时间'
              : `-${formatTime(-skipConfig.outro_time)}`,
          onClick: () => {
            const duration = art.duration || 0;
            const currentTime = art.currentTime || 0;
            const outroTime = -(duration - currentTime);
            if (outroTime >= 0) return;
            const nextConfig = {
              ...skipConfigRef.current,
              outro_time: outroTime,
              enable: true,
            };
            void onSkipConfigChange(nextConfig);
            art.notice.show = `片尾已设置为 ${formatTime(currentTime)}`;
            return `-${formatTime(-outroTime)}`;
          },
        },
      ],
    });

    artRef.current = art;
    void requestWakeLock();

    return () => {
      persistCurrentProgress();
      art.destroy(false);
      artRef.current = null;
      if (hlsRef.current) {
        hlsRef.current.destroy();
        hlsRef.current = null;
      }
      void releaseWakeLock();
    };
  }, [
    blockAdEnabled,
    onBlockAdChange,
    onActiveQualityChange,
    onPlayerErrorChange,
    onPersistProgress,
    onQualityListChange,
    onSkipConfigChange,
    persistCurrentProgress,
    poster,
    releaseWakeLock,
    requestWakeLock,
    retryToken,
    skipConfig,
    streamUrl,
    tryResumePlayback,
  ]);

  useEffect(() => {
    if (!hlsRef.current) return;
    hlsRef.current.currentLevel = desiredQuality;
  }, [desiredQuality]);

  useEffect(() => {
    if (saveIntervalRef.current) {
      window.clearInterval(saveIntervalRef.current);
    }
    saveIntervalRef.current = window.setInterval(
      persistCurrentProgress,
      PlayRecordSaveIntervalMs
    );
    return () => {
      if (saveIntervalRef.current) {
        window.clearInterval(saveIntervalRef.current);
      }
    };
  }, [persistCurrentProgress]);

  useEffect(() => {
    if (skipCheckIntervalRef.current) {
      window.clearInterval(skipCheckIntervalRef.current);
    }
    skipCheckIntervalRef.current = window.setInterval(
      checkAndSkip,
      SkipCheckIntervalMs
    );
    return () => {
      if (skipCheckIntervalRef.current) {
        window.clearInterval(skipCheckIntervalRef.current);
      }
    };
  }, [checkAndSkip]);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      const art = artRef.current;
      if (!art) return;
      if (
        event.target instanceof HTMLInputElement ||
        event.target instanceof HTMLTextAreaElement
      ) {
        return;
      }

      if (event.key === 'm' || event.key === 'M') art.muted = !art.muted;
      if (event.code === 'Space') {
        event.preventDefault();
        if (art.playing) art.pause();
        else art.play();
      }
      if (event.key === 'f' || event.key === 'F') art.fullscreen = !art.fullscreen;
      if (!event.altKey && event.key === 'ArrowRight') {
        art.currentTime = Math.min((art.currentTime || 0) + 10, art.duration || Infinity);
      }
      if (!event.altKey && event.key === 'ArrowLeft') {
        art.currentTime = Math.max((art.currentTime || 0) - 10, 0);
      }
    };

    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, []);

  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        void requestWakeLock();
      } else {
        persistCurrentProgress();
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () =>
      document.removeEventListener('visibilitychange', handleVisibilityChange);
  }, [persistCurrentProgress, requestWakeLock]);

  useEffect(() => {
    const handleBeforeUnload = () => persistCurrentProgress();
    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, [persistCurrentProgress]);

  return (
    <div className='relative aspect-video w-full bg-black'>
      {loading ? (
        <div className='flex h-full items-center justify-center text-netflix-gray-400'>
          <Loader2 className='mr-2 h-5 w-5 animate-spin' />
          正在加载...
        </div>
      ) : streamUrl ? (
        <div ref={containerRef} className='h-full w-full' />
      ) : (
        <div className='flex h-full items-center justify-center text-netflix-gray-500'>
          暂无可用播放源
        </div>
      )}
    </div>
  );
}
