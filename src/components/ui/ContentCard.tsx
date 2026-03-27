'use client';

import { motion } from 'framer-motion';
import { Check, Play, Plus } from 'lucide-react';
import React, { useCallback, useEffect, useMemo, useState } from 'react';

import {
  normalizeImageUrl,
  toImageSrc,
  toLogoProxyImageSrc,
  toProxyImageSrc,
} from '@/lib/image';
import { useFastNavigation } from '@/lib/navigation-feedback';

import SmartImage from '@/components/ui/SmartImage';

interface PosterSearchCandidate {
  title?: string;
  poster?: string;
  cover?: string;
  source?: string;
  source_name?: string;
}

const posterRecoveryCache = new Map<string, string | null>();
const posterRecoveryRequests = new Map<string, Promise<string | null>>();

function normalizePosterLookupTitle(value: string): string {
  return (value || '')
    .toLowerCase()
    .replace(/[\s\-_.:：!！?？,，。·'"“”‘’/]/g, '')
    .trim();
}

export type ContentType = 'movie' | 'tv' | 'variety' | 'anime';

export interface ContentCardProps {
  id?: string;
  title: string;
  cover: string;
  firstEpisode?: string;
  rating?: string;
  year?: string;
  type?: ContentType;
  source?: string;
  sourceName?: string;
  searchTitle?: string;
  className?: string;
  size?: 'sm' | 'md' | 'lg';
  desc?: string;
}

const typeLabel: Record<ContentType, string> = {
  movie: '电影',
  tv: '剧集',
  variety: '综艺',
  anime: '动漫',
};

export const ContentCard: React.FC<ContentCardProps> = ({
  id,
  title,
  cover,
  firstEpisode,
  rating,
  year,
  type = 'movie',
  source,
  sourceName,
  searchTitle,
  className = '',
  size = 'md',
}) => {
  const { navigate, prefetchHref } = useFastNavigation();
  const [favorite, setFavorite] = useState(false);
  const [proxyStage, setProxyStage] = useState(0);
  const [resolvedCover, setResolvedCover] = useState(cover);
  const [recoverTried, setRecoverTried] = useState(false);
  const [searchRecoverTried, setSearchRecoverTried] = useState(false);
  const normalizedResolvedCover = useMemo(
    () => normalizeImageUrl(resolvedCover),
    [resolvedCover]
  );
  const isRemoteCover = normalizedResolvedCover.startsWith('http://') || normalizedResolvedCover.startsWith('https://');

  const imageSrc = useMemo(
    () =>
      proxyStage === 0
        ? isRemoteCover
          ? toProxyImageSrc(normalizedResolvedCover)
          : toImageSrc(normalizedResolvedCover)
        : proxyStage === 1
        ? toLogoProxyImageSrc(normalizedResolvedCover)
        : toImageSrc(normalizedResolvedCover),
    [isRemoteCover, normalizedResolvedCover, proxyStage]
  );

  useEffect(() => {
    setResolvedCover(cover);
    setProxyStage(0);
    setRecoverTried(false);
    setSearchRecoverTried(false);
  }, [cover, source, id]);

  const recoverPosterByDetail = useCallback(async () => {
    if (recoverTried || !source || !id) return;
    setRecoverTried(true);

    try {
      const response = await fetch(
        `/api/detail?source=${encodeURIComponent(
          source
        )}&id=${encodeURIComponent(id)}`
      );
      if (!response.ok) return;

      const detail = await response.json();
      const poster = normalizeImageUrl(detail?.poster || detail?.cover || '');
      if (poster) {
        setResolvedCover(poster);
        setProxyStage(0);
      }
    } catch {
      // ignore recover failure
    }
  }, [id, recoverTried, source]);

  const recoverPosterBySearch = useCallback(async () => {
    if (searchRecoverTried) return;

    const lookupTitle = (searchTitle || title).trim();
    if (!lookupTitle) {
      setSearchRecoverTried(true);
      return;
    }

    const cacheKey = normalizePosterLookupTitle(lookupTitle);
    if (!cacheKey) {
      setSearchRecoverTried(true);
      return;
    }

    const applyRecoveredCover = (nextCover: string | null) => {
      if (!nextCover) return;
      setResolvedCover(nextCover);
      setProxyStage(0);
    };

    if (posterRecoveryCache.has(cacheKey)) {
      setSearchRecoverTried(true);
      applyRecoveredCover(posterRecoveryCache.get(cacheKey) || null);
      return;
    }

    const inflightRequest = posterRecoveryRequests.get(cacheKey);
    if (inflightRequest) {
      setSearchRecoverTried(true);
      applyRecoveredCover(await inflightRequest);
      return;
    }

    const request = (async () => {
      try {
        const params = new URLSearchParams({
          title: lookupTitle,
          cover: normalizedResolvedCover,
        });
        if (year) params.set('year', year);
        if (type) params.set('type', type);
        if (source) params.set('source', source);

        const response = await fetch(`/api/poster/recover?${params.toString()}`);
        if (!response.ok) return null;

        const payload = await response.json();
        const legacyPoster = normalizeImageUrl(payload?.poster || '');
        if (legacyPoster) return legacyPoster;

        const nestedPoster = normalizeImageUrl(payload?.data?.poster || '');
        if (nestedPoster) return nestedPoster;

        const candidates = Array.isArray(payload?.results) ? payload.results : [];
        if (!candidates.length) return null;

        const matchedCandidate = candidates.find((candidate: PosterSearchCandidate) =>
          normalizeImageUrl(candidate.poster || candidate.cover || '')
        );
        return normalizeImageUrl(matchedCandidate?.poster || matchedCandidate?.cover || '');
      } catch {
        return null;
      }
    })();

    posterRecoveryRequests.set(cacheKey, request);

    try {
      const recoveredCover = await request;
      posterRecoveryCache.set(cacheKey, recoveredCover);
      setSearchRecoverTried(true);
      applyRecoveredCover(recoveredCover);
    } finally {
      posterRecoveryRequests.delete(cacheKey);
    }
  }, [normalizedResolvedCover, searchRecoverTried, searchTitle, source, title, type, year]);

  useEffect(() => {
    if (!normalizedResolvedCover && source && id && !recoverTried) {
      void recoverPosterByDetail();
    }
  }, [id, normalizedResolvedCover, recoverPosterByDetail, recoverTried, source]);

  useEffect(() => {
    if (!normalizedResolvedCover && !searchRecoverTried) {
      void recoverPosterBySearch();
    }
  }, [normalizedResolvedCover, recoverPosterBySearch, searchRecoverTried]);
  const ratio = size === 'lg' ? 'aspect-video' : 'aspect-[2/3]';

  const playHref = useMemo(() => {
    if (!source || source === 'douban' || !id) return '';

    const params = new URLSearchParams({
      source,
      id,
      title,
    });
    if (firstEpisode) params.set('ep', firstEpisode);
    if (year) params.set('year', year);
    if (type) params.set('stype', type);
    if (sourceName) params.set('sname', sourceName);
    if (searchTitle) params.set('stitle', searchTitle);
    return `/play?${params.toString()}`;
  }, [firstEpisode, id, searchTitle, source, sourceName, title, type, year]);

  const searchHref = useMemo(
    () => `/search?q=${encodeURIComponent(title)}`,
    [title]
  );
  const destinationHref = playHref || searchHref;

  useEffect(() => {
    prefetchHref(destinationHref);
  }, [destinationHref, prefetchHref]);

  const jump = () => {
    navigate(destinationHref);
  };

  const addFavorite = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (favorite) return;
    try {
      await fetch('/api/favorites', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          source: source || 'douban',
          id: id || title,
          title,
          cover,
          year,
        }),
      });
      setFavorite(true);
    } catch {
      setFavorite(false);
    }
  };

  return (
    <motion.article
      onClick={jump}
      onPointerEnter={() => prefetchHref(destinationHref)}
      onFocus={() => prefetchHref(destinationHref)}
      whileHover={{ y: -6, scale: 1.03 }}
      whileTap={{ scale: 0.985 }}
      transition={{ duration: 0.22 }}
      className={`group relative cursor-pointer will-change-transform ${className}`}
    >
      <div
        className={`relative overflow-hidden rounded-md bg-[#1f1f1f] ${ratio}`}
      >
        <SmartImage
          key={`${imageSrc}-${proxyStage}-${recoverTried}-${searchRecoverTried}`}
          src={imageSrc}
          alt={title}
          fill
          sizes={
            size === 'lg'
              ? '(max-width: 1024px) 100vw, 50vw'
              : '(max-width: 768px) 50vw, 20vw'
          }
          className='h-full w-full object-cover transition-transform duration-500 group-hover:scale-110'
          onError={() => {
            if (proxyStage < 2) {
              setProxyStage((prev) => prev + 1);
              return;
            }
            if (!recoverTried && source && id) {
              void recoverPosterByDetail();
              return;
            }
            if (!searchRecoverTried) {
              void recoverPosterBySearch();
            }
          }}
        />
        <div className='absolute inset-0 bg-gradient-to-t from-black/80 via-black/10 to-transparent' />

        <div className='absolute left-2 top-2 rounded bg-black/60 px-1.5 py-0.5 text-[10px] text-white'>
          {typeLabel[type]}
        </div>

        <div className='absolute inset-x-0 bottom-0 p-2 opacity-0 transition-opacity group-hover:opacity-100'>
          <div className='flex items-center gap-2'>
            <button
              onClick={(e) => {
                e.stopPropagation();
                jump();
              }}
              className='flex h-8 w-8 items-center justify-center rounded-full bg-white text-black'
            >
              <Play className='h-4 w-4 fill-black' />
            </button>
            <button
              onClick={addFavorite}
              className='flex h-8 w-8 items-center justify-center rounded-full border border-white/60 text-white'
            >
              {favorite ? (
                <Check className='h-4 w-4' />
              ) : (
                <Plus className='h-4 w-4' />
              )}
            </button>
          </div>
        </div>
      </div>

      <div className='mt-2'>
        <h3 className='truncate text-sm font-semibold text-white'>{title}</h3>
        <p className='mt-1 text-xs text-zinc-400'>
          {rating ? `★ ${rating}` : ''}
          {rating && year ? ' · ' : ''}
          {year || ''}
        </p>
      </div>
    </motion.article>
  );
};

export default ContentCard;
