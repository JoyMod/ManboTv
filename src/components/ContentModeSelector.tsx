'use client';

import React from 'react';

import {
  contentModeOptions,
  ContentModePreference,
  getContentModePreference,
  setContentModePreference,
} from '@/lib/content-mode';

type ContentModeSelectorProps = {
  compact?: boolean;
};

export default function ContentModeSelector({
  compact = false,
}: ContentModeSelectorProps) {
  const [mode, setMode] = React.useState<ContentModePreference>('inherit');

  React.useEffect(() => {
    setMode(getContentModePreference());
  }, []);

  const handleModeChange = (nextMode: ContentModePreference) => {
    if (nextMode === mode) {
      return;
    }

    setContentModePreference(nextMode);
    setMode(nextMode);
    window.location.reload();
  };

  return (
    <div
      className={
        compact ? 'space-y-2' : 'mb-4 rounded border border-zinc-800 bg-zinc-950/70 p-3'
      }
    >
      <div className='mb-2'>
        <div className='text-sm font-medium text-white'>内容模式</div>
        <div className='text-xs text-zinc-400'>只影响当前浏览器，不会改动全站默认设置。</div>
      </div>
      <div className='space-y-2'>
        {contentModeOptions.map((option) => {
          const active = option.value === mode;
          return (
            <button
              key={option.value}
              type='button'
              onClick={() => handleModeChange(option.value)}
              className={`w-full rounded-lg border px-3 py-2 text-left transition-colors ${
                active
                  ? 'border-red-500/60 bg-red-500/15 text-white'
                  : 'border-zinc-800 bg-zinc-900/80 text-zinc-200 hover:border-zinc-700 hover:bg-zinc-800'
              }`}
            >
              <div className='text-sm font-medium'>{option.label}</div>
              <div className='mt-1 text-xs text-zinc-400'>{option.description}</div>
            </button>
          );
        })}
      </div>
    </div>
  );
}
