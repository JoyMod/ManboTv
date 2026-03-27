'use client';

import { X } from 'lucide-react';
import React, { useEffect } from 'react';

interface AdminDialogProps {
  open: boolean;
  title: string;
  description?: string;
  children: React.ReactNode;
  actions: React.ReactNode;
  onClose: () => void;
}

export default function AdminDialog({
  open,
  title,
  description,
  children,
  actions,
  onClose,
}: AdminDialogProps) {
  useEffect(() => {
    if (!open) return;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      document.body.style.overflow = previousOverflow;
    };
  }, [open]);

  if (!open) return null;

  return (
    <div className='fixed inset-0 z-50 flex items-center justify-center p-4'>
      <button
        type='button'
        aria-label='关闭弹窗'
        onClick={onClose}
        className='absolute inset-0 bg-black/70 backdrop-blur-sm'
      />
      <div className='relative z-10 w-full max-w-md rounded-2xl border border-netflix-gray-800 bg-netflix-surface p-6 shadow-2xl'>
        <div className='mb-4 flex items-start justify-between gap-4'>
          <div>
            <h3 className='text-xl font-bold text-white'>{title}</h3>
            {description ? (
              <p className='mt-2 text-sm leading-6 text-netflix-gray-400'>
                {description}
              </p>
            ) : null}
          </div>
          <button
            type='button'
            onClick={onClose}
            className='rounded-full p-2 text-netflix-gray-400 transition-colors hover:bg-netflix-gray-800 hover:text-white'
          >
            <X className='h-4 w-4' />
          </button>
        </div>

        <div className='space-y-4'>{children}</div>

        <div className='mt-6 flex justify-end gap-3'>{actions}</div>
      </div>
    </div>
  );
}
