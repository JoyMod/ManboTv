'use client';

import React from 'react';

interface SiteConfig {
  SiteName: string;
  Announcement: string;
  SearchDownstreamMaxPage: number;
  SiteInterfaceCacheTime: number;
  DoubanProxyType: string;
  DoubanProxy: string;
  DoubanImageProxyType: string;
  DoubanImageProxy: string;
  DisableYellowFilter: boolean;
  FluidSearch: boolean;
  ContentAccessMode: 'safe' | 'mixed' | 'adult_only';
  BlockedContentTags: string[];
}

interface ContentAccessModeOption {
  label: string;
  value: SiteConfig['ContentAccessMode'];
  description: string;
}

interface AdminSettingsPanelProps {
  siteConfig: SiteConfig;
  saving: boolean;
  contentAccessModeOptions: ContentAccessModeOption[];
  onConfigChange: (updater: (prev: SiteConfig) => SiteConfig) => void;
  onSave: () => void;
}

export default function AdminSettingsPanel({
  siteConfig,
  saving,
  contentAccessModeOptions,
  onConfigChange,
  onSave,
}: AdminSettingsPanelProps) {
  return (
    <div className='max-w-2xl space-y-6'>
      <div className='bg-netflix-surface rounded-xl p-6'>
        <h3 className='mb-6 text-lg font-bold text-white'>站点设置</h3>

        <div className='space-y-4'>
          <div>
            <label className='mb-2 block text-netflix-gray-300'>站点名称</label>
            <input
              type='text'
              value={siteConfig.SiteName}
              onChange={(e) =>
                onConfigChange((prev) => ({
                  ...prev,
                  SiteName: e.target.value,
                }))
              }
              className='w-full rounded-lg border border-netflix-gray-700 bg-netflix-gray-800 px-4 py-3 text-white focus:border-netflix-red focus:outline-none'
            />
          </div>

          <div>
            <label className='mb-2 block text-netflix-gray-300'>公告</label>
            <textarea
              rows={3}
              value={siteConfig.Announcement}
              onChange={(e) =>
                onConfigChange((prev) => ({
                  ...prev,
                  Announcement: e.target.value,
                }))
              }
              className='w-full resize-none rounded-lg border border-netflix-gray-700 bg-netflix-gray-800 px-4 py-3 text-white focus:border-netflix-red focus:outline-none'
            />
          </div>

          <div className='flex items-center justify-between border-t border-netflix-gray-800 py-4'>
            <span className='text-netflix-gray-300'>启用流式搜索</span>
            <input
              type='checkbox'
              checked={siteConfig.FluidSearch}
              onChange={(e) =>
                onConfigChange((prev) => ({
                  ...prev,
                  FluidSearch: e.target.checked,
                }))
              }
              className='h-5 w-5 accent-netflix-red'
            />
          </div>

          <div className='border-t border-netflix-gray-800 pt-4'>
            <label className='mb-2 block text-netflix-gray-300'>
              内容访问模式
            </label>
            <select
              value={siteConfig.ContentAccessMode}
              onChange={(e) =>
                onConfigChange((prev) => ({
                  ...prev,
                  ContentAccessMode: e.target
                    .value as SiteConfig['ContentAccessMode'],
                  DisableYellowFilter: e.target.value !== 'safe',
                }))
              }
              className='w-full rounded-lg border border-netflix-gray-700 bg-netflix-gray-800 px-4 py-3 text-white focus:border-netflix-red focus:outline-none'
            >
              {contentAccessModeOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            <p className='mt-2 text-sm text-netflix-gray-500'>
              {
                contentAccessModeOptions.find(
                  (option) => option.value === siteConfig.ContentAccessMode
                )?.description
              }
            </p>
          </div>

          <div className='border-t border-netflix-gray-800 pt-4'>
            <label className='mb-2 block text-netflix-gray-300'>
              屏蔽内容标签
            </label>
            <textarea
              rows={4}
              value={siteConfig.BlockedContentTags.join('\n')}
              onChange={(e) =>
                onConfigChange((prev) => ({
                  ...prev,
                  BlockedContentTags: Array.from(
                    new Set(
                      e.target.value
                        .split(/[\n,，]/)
                        .map((item) => item.trim())
                        .filter(Boolean)
                    )
                  ),
                }))
              }
              placeholder={'每行一个标签，例如：\n十八禁\n情色\n电影解说'}
              className='w-full resize-none rounded-lg border border-netflix-gray-700 bg-netflix-gray-800 px-4 py-3 text-white focus:border-netflix-red focus:outline-none'
            />
            <p className='mt-2 text-sm text-netflix-gray-500'>
              这里用于追加自定义屏蔽标签，例如电影解说、花絮、偷拍等。
            </p>
          </div>
        </div>

        <div className='mt-6 border-t border-netflix-gray-800 pt-6'>
          <button
            onClick={onSave}
            disabled={saving}
            className='rounded-lg bg-netflix-red px-6 py-3 font-bold text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-60'
          >
            {saving ? '保存中...' : '保存设置'}
          </button>
        </div>
      </div>
    </div>
  );
}
