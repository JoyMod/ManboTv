'use client';

import { X } from 'lucide-react';
import React, { useEffect } from 'react';

interface ActionSheetOption {
  label: string;
  value: string;
  icon?: React.ReactNode;
  danger?: boolean;
  disabled?: boolean;
}

interface MobileActionSheetProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  options: ActionSheetOption[];
  onSelect: (value: string) => void;
  cancelText?: string;
}

export default function MobileActionSheet({
  isOpen,
  onClose,
  title,
  options,
  onSelect,
  cancelText = '取消',
}: MobileActionSheetProps) {
  // 阻止背景滚动
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 md:hidden">
      {/* 遮罩 */}
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />

      {/* 内容 */}
      <div className="absolute bottom-0 left-0 right-0 animate-in slide-in-from-bottom duration-200">
        <div className="rounded-t-2xl bg-zinc-900 p-2">
          {/* 标题 */}
          {title && (
            <div className="flex items-center justify-between border-b border-zinc-800 px-4 py-3">
              <span className="text-sm font-medium text-zinc-400">{title}</span>
              <button
                onClick={onClose}
                className="rounded-full p-1 text-zinc-400 hover:bg-zinc-800"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          )}

          {/* 选项 */}
          <div className="py-1">
            {options.map((option, index) => (
              <button
                key={option.value}
                onClick={() => {
                  if (!option.disabled) {
                    onSelect(option.value);
                    onClose();
                  }
                }}
                disabled={option.disabled}
                className={`flex w-full items-center gap-3 px-4 py-3 text-left transition-colors ${
                  option.danger
                    ? 'text-red-500'
                    : option.disabled
                    ? 'text-zinc-600'
                    : 'text-white hover:bg-zinc-800'
                } ${index !== options.length - 1 ? 'border-b border-zinc-800/50' : ''}`}
              >
                {option.icon && <span className="text-zinc-400">{option.icon}</span>}
                <span className="flex-1">{option.label}</span>
              </button>
            ))}
          </div>

          {/* 取消按钮 */}
          <div className="mt-2 border-t border-zinc-800 pt-2">
            <button
              onClick={onClose}
              className="w-full rounded-lg bg-zinc-800 py-3 text-center font-medium text-white transition-colors hover:bg-zinc-700"
            >
              {cancelText}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
