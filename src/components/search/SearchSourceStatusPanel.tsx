'use client';

import { AlertTriangle, Clock3, ServerCrash, ShieldCheck } from 'lucide-react';
import React from 'react';

import {
  SearchExecutionInfo,
  SearchSourceStatusItem,
} from '@/components/search/search-utils';

interface SearchSourceStatusPanelProps {
  execution?: SearchExecutionInfo;
  statuses: SearchSourceStatusItem[];
}

function statusTone(status: SearchSourceStatusItem['status']): string {
  switch (status) {
    case 'done':
      return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-200';
    case 'partial':
      return 'border-amber-500/30 bg-amber-500/10 text-amber-200';
    case 'empty':
      return 'border-netflix-gray-700 bg-black/20 text-netflix-gray-300';
    case 'timeout':
      return 'border-yellow-500/30 bg-yellow-500/10 text-yellow-200';
    default:
      return 'border-rose-500/30 bg-rose-500/10 text-rose-200';
  }
}

function statusLabel(status: SearchSourceStatusItem['status']): string {
  switch (status) {
    case 'done':
      return '完成';
    case 'partial':
      return '部分返回';
    case 'empty':
      return '无结果';
    case 'timeout':
      return '超时';
    default:
      return '失败';
  }
}

export default function SearchSourceStatusPanel({
  execution,
  statuses,
}: SearchSourceStatusPanelProps) {
  if (!statuses.length) return null;

  return (
    <section className='rounded-3xl border border-netflix-gray-800 bg-netflix-surface/60 p-4'>
      <div className='flex flex-wrap items-center justify-between gap-3'>
        <div>
          <p className='flex items-center gap-2 text-sm text-white'>
            <ShieldCheck className='h-4 w-4 text-netflix-red' />
            搜索执行状态
          </p>
          <p className='mt-1 text-xs text-netflix-gray-500'>
            {execution?.completed_sources || statuses.length} /{' '}
            {execution?.total_sources || statuses.length} 个资源站已响应
            {typeof execution?.elapsed_ms === 'number'
              ? ` · ${execution.elapsed_ms}ms`
              : ''}
          </p>
        </div>

        {execution?.degraded && (
          <span className='inline-flex items-center gap-2 rounded-full border border-yellow-500/30 bg-yellow-500/10 px-3 py-1 text-xs text-yellow-200'>
            <AlertTriangle className='h-3.5 w-3.5' />
            当前结果为快速降级返回
          </span>
        )}
      </div>

      <div className='mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-3'>
        {statuses.map((item) => (
          <div
            key={item.source}
            className={`rounded-2xl border px-4 py-3 ${statusTone(item.status)}`}
          >
            <div className='flex items-start justify-between gap-3'>
              <div>
                <p className='text-sm font-medium text-white'>
                  {item.source_name || item.source}
                </p>
                <p className='mt-1 text-xs opacity-80'>{statusLabel(item.status)}</p>
              </div>

              {item.status === 'done' ? (
                <ShieldCheck className='h-4 w-4' />
              ) : item.status === 'timeout' ? (
                <Clock3 className='h-4 w-4' />
              ) : (
                <ServerCrash className='h-4 w-4' />
              )}
            </div>

            <div className='mt-3 flex flex-wrap gap-3 text-[11px] opacity-80'>
              <span>结果 {item.result_count || 0}</span>
              <span>页数 {item.page_count || 0}</span>
              <span>{item.elapsed_ms || 0}ms</span>
            </div>

            {item.error && item.status !== 'done' && (
              <p className='mt-3 line-clamp-2 text-[11px] opacity-75'>
                {item.error}
              </p>
            )}
          </div>
        ))}
      </div>
    </section>
  );
}
