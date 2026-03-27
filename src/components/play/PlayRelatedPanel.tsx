'use client';

import React from 'react';

import { mapType, RelatedResult } from '@/components/play/play-utils';
import ContentCard from '@/components/ui/ContentCard';

export interface PlayRelatedPanelProps {
  relatedVideos: RelatedResult[];
}

export default function PlayRelatedPanel({
  relatedVideos,
}: PlayRelatedPanelProps) {
  return (
    <div>
      <h2 className='mb-4 text-lg font-bold text-white'>相关推荐</h2>
      {relatedVideos.length > 0 ? (
        <div className='grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-2'>
          {relatedVideos.map((video, index) => (
            <ContentCard
              key={`${video.source || 'related'}-${video.id || index}`}
              id={video.id}
              source={video.source}
              title={video.title || '未知标题'}
              cover={video.poster || '/placeholder-poster.svg'}
              year={video.year || ''}
              type={mapType(`${video.type_name || ''} ${video.class || ''}`)}
              size='sm'
            />
          ))}
        </div>
      ) : (
        <p className='text-sm text-netflix-gray-500'>暂无相关推荐</p>
      )}
    </div>
  );
}
