'use client';

import Hls from 'hls.js';
import { AlertCircle, Loader2, RotateCcw, Tv } from 'lucide-react';
import React from 'react';

const HlsMimeType = 'application/vnd.apple.mpegurl';
const NativeVideoReadyEvent = 'loadedmetadata';
const LivePlayerErrorMessage = '当前直播流加载失败，可重试或切换频道';

export interface LiveVideoPlayerProps {
  url?: string;
  title?: string;
}

function isHlsStream(url: string): boolean {
  return url.includes('.m3u8') || url.includes('/api/proxy/m3u8?');
}

export default function LiveVideoPlayer({
  url = '',
  title = '直播频道',
}: LiveVideoPlayerProps) {
  const videoRef = React.useRef<HTMLVideoElement>(null);
  const hlsRef = React.useRef<Hls | null>(null);
  const [loading, setLoading] = React.useState(Boolean(url));
  const [error, setError] = React.useState<string | null>(null);
  const [reloadToken, setReloadToken] = React.useState(0);

  React.useEffect(() => {
    const video = videoRef.current;
    if (!video || !url) {
      setLoading(false);
      setError(null);
      return;
    }

    setLoading(true);
    setError(null);

    const clearPlayer = () => {
      if (hlsRef.current) {
        hlsRef.current.destroy();
        hlsRef.current = null;
      }
      video.pause();
      video.removeAttribute('src');
      video.load();
    };

    const handleReady = () => {
      setLoading(false);
      setError(null);
      void video.play().catch(() => {
        return;
      });
    };

    const handleError = () => {
      setLoading(false);
      setError(LivePlayerErrorMessage);
    };

    video.addEventListener(NativeVideoReadyEvent, handleReady);
    video.addEventListener('error', handleError);

    if (isHlsStream(url)) {
      if (video.canPlayType(HlsMimeType)) {
        video.src = url;
      } else if (Hls.isSupported()) {
        const hls = new Hls({
          enableWorker: true,
          lowLatencyMode: true,
        });
        hlsRef.current = hls;
        hls.loadSource(url);
        hls.attachMedia(video);
        hls.on(Hls.Events.MANIFEST_PARSED, handleReady);
        hls.on(Hls.Events.ERROR, (_event, data) => {
          if (!data?.fatal) {
            return;
          }
          setLoading(false);
          setError(LivePlayerErrorMessage);
        });
      } else {
        setLoading(false);
        setError('当前浏览器不支持直播流播放');
      }
    } else {
      video.src = url;
    }

    return () => {
      video.removeEventListener(NativeVideoReadyEvent, handleReady);
      video.removeEventListener('error', handleError);
      clearPlayer();
    };
  }, [reloadToken, url]);

  if (!url) {
    return (
      <div className='flex h-full w-full items-center justify-center text-netflix-gray-500'>
        <div className='text-center'>
          <Tv className='mx-auto mb-3 h-12 w-12 text-netflix-gray-600' />
          <p>该频道暂无可播放地址</p>
        </div>
      </div>
    );
  }

  return (
    <div className='relative h-full w-full bg-black'>
      <video
        ref={videoRef}
        key={`${url}-${reloadToken}`}
        className='h-full w-full bg-black'
        controls
        autoPlay
        playsInline
      />

      {loading ? (
        <div className='absolute inset-0 flex items-center justify-center bg-black/45 text-zinc-200'>
          <Loader2 className='mr-2 h-5 w-5 animate-spin' />
          正在缓冲 {title}...
        </div>
      ) : null}

      {error ? (
        <div className='absolute inset-x-4 bottom-4 rounded-xl border border-red-500/30 bg-black/75 p-3 text-sm text-red-300 backdrop-blur'>
          <div className='flex items-start gap-2'>
            <AlertCircle className='mt-0.5 h-4 w-4 shrink-0' />
            <div className='min-w-0 flex-1'>
              <p>{error}</p>
            </div>
            <button
              type='button'
              onClick={() => setReloadToken((prev) => prev + 1)}
              className='inline-flex items-center gap-1 rounded-full border border-red-400/30 px-2.5 py-1 text-xs text-white transition-colors hover:border-red-300/60 hover:bg-red-500/15'
            >
              <RotateCcw className='h-3.5 w-3.5' />
              重试
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}
