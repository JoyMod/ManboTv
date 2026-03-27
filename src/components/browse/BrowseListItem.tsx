'use client';

import { toProxyImageSrc } from '@/lib/image';
import { useFastNavigation } from '@/lib/navigation-feedback';

import SmartImage from '@/components/ui/SmartImage';

interface BrowseListItemProps {
  title: string;
  cover: string;
  year?: string;
  rate?: string;
}

export default function BrowseListItem({
  title,
  cover,
  year,
  rate,
}: BrowseListItemProps) {
  const { navigate, prefetchHref } = useFastNavigation();
  const searchHref = `/search?q=${encodeURIComponent(title)}`;

  return (
    <div
      onClick={() => navigate(searchHref)}
      onPointerEnter={() => prefetchHref(searchHref)}
      className="flex cursor-pointer gap-4 rounded-lg bg-zinc-900 p-4 transition-colors hover:bg-zinc-800"
    >
      <div className="relative flex h-36 w-24 shrink-0 items-center justify-center overflow-hidden rounded bg-zinc-950">
        <SmartImage
          src={toProxyImageSrc(cover)}
          alt={title}
          fill
          sizes="96px"
          className="object-contain"
        />
      </div>
      <div className="flex-1">
        <h3 className="text-lg font-bold text-white">{title}</h3>
        <div className="mt-2 flex gap-2 text-sm">
          {rate && <span className="text-green-400">{rate}</span>}
          {year && <span className="text-zinc-400">{year}</span>}
        </div>
      </div>
    </div>
  );
}
