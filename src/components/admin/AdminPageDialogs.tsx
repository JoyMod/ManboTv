'use client';

import React from 'react';

import AdminDialog from '@/components/admin/AdminDialog';

interface UserDraft {
  username: string;
  password: string;
}

interface SourceDraft {
  key: string;
  name: string;
  api: string;
}

interface AdminPageDialogsProps {
  actionSubmitting: boolean;
  userDialogOpen: boolean;
  userDraft: UserDraft;
  onUserDialogClose: () => void;
  onUserDraftChange: (field: keyof UserDraft, value: string) => void;
  onUserSubmit: () => void;
  sourceDialogOpen: boolean;
  sourceDraft: SourceDraft;
  onSourceDialogClose: () => void;
  onSourceDraftChange: (field: keyof SourceDraft, value: string) => void;
  onSourceSubmit: () => void;
  pendingDeleteUserName: string | null;
  onPendingDeleteUserClose: () => void;
  onPendingDeleteUserConfirm: () => void;
  pendingDeleteSourceName: string | null;
  onPendingDeleteSourceClose: () => void;
  onPendingDeleteSourceConfirm: () => void;
}

export default function AdminPageDialogs({
  actionSubmitting,
  userDialogOpen,
  userDraft,
  onUserDialogClose,
  onUserDraftChange,
  onUserSubmit,
  sourceDialogOpen,
  sourceDraft,
  onSourceDialogClose,
  onSourceDraftChange,
  onSourceSubmit,
  pendingDeleteUserName,
  onPendingDeleteUserClose,
  onPendingDeleteUserConfirm,
  pendingDeleteSourceName,
  onPendingDeleteSourceClose,
  onPendingDeleteSourceConfirm,
}: AdminPageDialogsProps) {
  return (
    <>
      <AdminDialog
        open={userDialogOpen}
        title='添加用户'
        description='创建新的站点账号，默认角色为普通用户。'
        onClose={onUserDialogClose}
        actions={
          <>
            <button
              type='button'
              onClick={onUserDialogClose}
              disabled={actionSubmitting}
              className='rounded-lg border border-netflix-gray-700 px-4 py-2 text-netflix-gray-300 transition-colors hover:bg-netflix-gray-800'
            >
              取消
            </button>
            <button
              type='button'
              onClick={onUserSubmit}
              disabled={actionSubmitting}
              className='rounded-lg bg-netflix-red px-4 py-2 font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-60'
            >
              {actionSubmitting ? '创建中...' : '创建用户'}
            </button>
          </>
        }
      >
        <div>
          <label className='mb-2 block text-sm text-netflix-gray-300'>
            用户名
          </label>
          <input
            type='text'
            value={userDraft.username}
            onChange={(e) => onUserDraftChange('username', e.target.value)}
            className='w-full rounded-lg border border-netflix-gray-700 bg-netflix-gray-800 px-4 py-3 text-white focus:border-netflix-red focus:outline-none'
          />
        </div>
        <div>
          <label className='mb-2 block text-sm text-netflix-gray-300'>
            密码
          </label>
          <input
            type='password'
            value={userDraft.password}
            onChange={(e) => onUserDraftChange('password', e.target.value)}
            className='w-full rounded-lg border border-netflix-gray-700 bg-netflix-gray-800 px-4 py-3 text-white focus:border-netflix-red focus:outline-none'
          />
        </div>
      </AdminDialog>

      <AdminDialog
        open={sourceDialogOpen}
        title='添加视频源'
        description='补充新的资源站 API，用于搜索和播放线路扩展。'
        onClose={onSourceDialogClose}
        actions={
          <>
            <button
              type='button'
              onClick={onSourceDialogClose}
              disabled={actionSubmitting}
              className='rounded-lg border border-netflix-gray-700 px-4 py-2 text-netflix-gray-300 transition-colors hover:bg-netflix-gray-800'
            >
              取消
            </button>
            <button
              type='button'
              onClick={onSourceSubmit}
              disabled={actionSubmitting}
              className='rounded-lg bg-netflix-red px-4 py-2 font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-60'
            >
              {actionSubmitting ? '保存中...' : '添加源'}
            </button>
          </>
        }
      >
        <div>
          <label className='mb-2 block text-sm text-netflix-gray-300'>
            源 Key
          </label>
          <input
            type='text'
            value={sourceDraft.key}
            onChange={(e) => onSourceDraftChange('key', e.target.value)}
            className='w-full rounded-lg border border-netflix-gray-700 bg-netflix-gray-800 px-4 py-3 text-white focus:border-netflix-red focus:outline-none'
          />
        </div>
        <div>
          <label className='mb-2 block text-sm text-netflix-gray-300'>
            源名称
          </label>
          <input
            type='text'
            value={sourceDraft.name}
            onChange={(e) => onSourceDraftChange('name', e.target.value)}
            className='w-full rounded-lg border border-netflix-gray-700 bg-netflix-gray-800 px-4 py-3 text-white focus:border-netflix-red focus:outline-none'
          />
        </div>
        <div>
          <label className='mb-2 block text-sm text-netflix-gray-300'>
            API 地址
          </label>
          <input
            type='url'
            value={sourceDraft.api}
            onChange={(e) => onSourceDraftChange('api', e.target.value)}
            className='w-full rounded-lg border border-netflix-gray-700 bg-netflix-gray-800 px-4 py-3 text-white focus:border-netflix-red focus:outline-none'
          />
        </div>
      </AdminDialog>

      <AdminDialog
        open={Boolean(pendingDeleteUserName)}
        title='确认删除用户'
        description={
          pendingDeleteUserName
            ? `删除后，用户 ${pendingDeleteUserName} 将无法继续登录。`
            : ''
        }
        onClose={onPendingDeleteUserClose}
        actions={
          <>
            <button
              type='button'
              onClick={onPendingDeleteUserClose}
              disabled={actionSubmitting}
              className='rounded-lg border border-netflix-gray-700 px-4 py-2 text-netflix-gray-300 transition-colors hover:bg-netflix-gray-800'
            >
              取消
            </button>
            <button
              type='button'
              onClick={onPendingDeleteUserConfirm}
              disabled={actionSubmitting}
              className='rounded-lg bg-red-600 px-4 py-2 font-semibold text-white transition-colors hover:bg-red-500 disabled:opacity-60'
            >
              {actionSubmitting ? '删除中...' : '确认删除'}
            </button>
          </>
        }
      >
        <p className='text-sm text-netflix-gray-400'>
          此操作不可撤销，请再次确认。
        </p>
      </AdminDialog>

      <AdminDialog
        open={Boolean(pendingDeleteSourceName)}
        title='确认删除视频源'
        description={
          pendingDeleteSourceName
            ? `删除后，${pendingDeleteSourceName} 将不再参与搜索和播放。`
            : ''
        }
        onClose={onPendingDeleteSourceClose}
        actions={
          <>
            <button
              type='button'
              onClick={onPendingDeleteSourceClose}
              disabled={actionSubmitting}
              className='rounded-lg border border-netflix-gray-700 px-4 py-2 text-netflix-gray-300 transition-colors hover:bg-netflix-gray-800'
            >
              取消
            </button>
            <button
              type='button'
              onClick={onPendingDeleteSourceConfirm}
              disabled={actionSubmitting}
              className='rounded-lg bg-red-600 px-4 py-2 font-semibold text-white transition-colors hover:bg-red-500 disabled:opacity-60'
            >
              {actionSubmitting ? '删除中...' : '确认删除'}
            </button>
          </>
        }
      >
        <p className='text-sm text-netflix-gray-400'>
          建议删除前先停用，确认无业务依赖后再执行。
        </p>
      </AdminDialog>
    </>
  );
}
