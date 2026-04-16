'use client';

import {
  BarChart3,
  CheckCircle,
  Edit3,
  Film,
  LayoutDashboard,
  LogOut,
  Plus,
  Search,
  Settings,
  Trash2,
  Users,
  XCircle,
} from 'lucide-react';
import { useRouter } from 'next/navigation';
import React, { useEffect, useMemo, useState } from 'react';

import AdminPageDialogs from '@/components/admin/AdminPageDialogs';
import AdminSettingsPanel from '@/components/admin/AdminSettingsPanel';

type TabId = 'dashboard' | 'users' | 'sources' | 'settings';

type SiteConfig = {
  SiteName: string;
  Announcement: string;
  SearchDownstreamMaxPage: number;
  SearchSourceTimeoutMs: number;
  SearchMaxConcurrent: number;
  SearchDefaultSort: string;
  SearchEnableStream: boolean;
  SiteInterfaceCacheTime: number;
  DoubanProxyType: string;
  DoubanProxy: string;
  DoubanImageProxyType: string;
  DoubanImageProxy: string;
  DisableYellowFilter: boolean;
  FluidSearch: boolean;
  ContentAccessMode: 'safe' | 'mixed' | 'adult_only';
  BlockedContentTags: string[];
};

type AdminUser = {
  username: string;
  role: 'owner' | 'admin' | 'user';
  banned?: boolean;
  tags?: string[];
  enabledApis?: string[];
};

type VideoSite = {
  key: string;
  name: string;
  api: string;
  detail?: string;
  disabled?: boolean;
  from?: 'config' | 'custom';
};

type NoticeState = {
  type: 'success' | 'error';
  message: string;
} | null;

type UserDraft = {
  username: string;
  password: string;
};

type SourceDraft = {
  key: string;
  name: string;
  api: string;
};

const noticeDismissDelayMs = 4000;

const menuItems: Array<{
  id: TabId;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
}> = [
  { id: 'dashboard', label: '概览', icon: LayoutDashboard },
  { id: 'users', label: '用户管理', icon: Users },
  { id: 'sources', label: '视频源', icon: Film },
  { id: 'settings', label: '站点设置', icon: Settings },
];

const defaultSiteConfig: SiteConfig = {
  SiteName: '曼波TV',
  Announcement: '',
  SearchDownstreamMaxPage: 3,
  SearchSourceTimeoutMs: 4500,
  SearchMaxConcurrent: 8,
  SearchDefaultSort: 'smart',
  SearchEnableStream: true,
  SiteInterfaceCacheTime: 30,
  DoubanProxyType: '',
  DoubanProxy: '',
  DoubanImageProxyType: '',
  DoubanImageProxy: '',
  DisableYellowFilter: false,
  FluidSearch: true,
  ContentAccessMode: 'safe',
  BlockedContentTags: [],
};

const contentAccessModeOptions: Array<{
  label: string;
  value: SiteConfig['ContentAccessMode'];
  description: string;
}> = [
  {
    label: '纯净模式',
    value: 'safe',
    description: '仅显示普通内容，屏蔽成人和色情资源。',
  },
  {
    label: '混合模式',
    value: 'mixed',
    description: '普通内容和成人内容都显示。',
  },
  {
    label: '成人专享',
    value: 'adult_only',
    description: '仅显示成人内容，普通影视资源不再展示。',
  },
];

function normalizeContentAccessMode(
  mode?: string,
  disableYellowFilter?: boolean
): SiteConfig['ContentAccessMode'] {
  if (mode === 'safe' || mode === 'mixed' || mode === 'adult_only') {
    return mode;
  }
  return disableYellowFilter ? 'mixed' : 'safe';
}

function createEmptyUserDraft(): UserDraft {
  return {
    username: '',
    password: '',
  };
}

function createEmptySourceDraft(): SourceDraft {
  return {
    key: '',
    name: '',
    api: '',
  };
}

async function requestApi<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options?.headers || {}),
    },
  });

  const json = await res.json().catch(() => ({}));

  if (!res.ok) {
    throw new Error(json?.error || json?.message || `HTTP ${res.status}`);
  }

  if (json && typeof json === 'object' && 'code' in json) {
    if (json.code !== 0) {
      throw new Error(json.message || '请求失败');
    }
    return json.data as T;
  }

  if (json?.error) {
    throw new Error(json.error);
  }

  return json as T;
}

export default function AdminPage() {
  const router = useRouter();
  const [activeTab, setActiveTab] = useState<TabId>('dashboard');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<NoticeState>(null);
  const [actionSubmitting, setActionSubmitting] = useState(false);

  const [siteConfig, setSiteConfig] = useState<SiteConfig>(defaultSiteConfig);
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [sites, setSites] = useState<VideoSite[]>([]);
  const [userKeyword, setUserKeyword] = useState('');
  const [userDraft, setUserDraft] = useState<UserDraft>(createEmptyUserDraft);
  const [sourceDraft, setSourceDraft] = useState<SourceDraft>(
    createEmptySourceDraft
  );
  const [userDialogOpen, setUserDialogOpen] = useState(false);
  const [sourceDialogOpen, setSourceDialogOpen] = useState(false);
  const [pendingDeleteUser, setPendingDeleteUser] = useState<AdminUser | null>(
    null
  );
  const [pendingDeleteSource, setPendingDeleteSource] =
    useState<VideoSite | null>(null);

  const loadAll = async () => {
    try {
      setLoading(true);
      setError(null);

      const configRaw = await requestApi<{
        SiteConfig?: Partial<SiteConfig>;
        Config?: { SiteConfig?: Partial<SiteConfig> };
      }>('/api/admin/config');
      const mergedSiteConfig =
        configRaw?.SiteConfig || configRaw?.Config?.SiteConfig || {};
      setSiteConfig({
        ...defaultSiteConfig,
        ...mergedSiteConfig,
        ContentAccessMode: normalizeContentAccessMode(
          mergedSiteConfig?.ContentAccessMode,
          mergedSiteConfig?.DisableYellowFilter
        ),
      });

      const usersData = await requestApi<AdminUser[]>('/api/admin/users');
      setUsers(Array.isArray(usersData) ? usersData : []);

      const sitesData = await requestApi<VideoSite[]>('/api/admin/sites');
      setSites(Array.isArray(sitesData) ? sitesData : []);
    } catch (e) {
      setError(e instanceof Error ? e.message : '加载后台数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadAll();
  }, []);

  useEffect(() => {
    if (!notice) return;
    const timer = window.setTimeout(() => {
      setNotice(null);
    }, noticeDismissDelayMs);
    return () => {
      window.clearTimeout(timer);
    };
  }, [notice]);

  const filteredUsers = useMemo(() => {
    const keyword = userKeyword.trim().toLowerCase();
    if (!keyword) return users;
    return users.filter((u) => u.username.toLowerCase().includes(keyword));
  }, [users, userKeyword]);

  const stats = useMemo(
    () => ({
      totalUsers: users.length,
      activeUsers: users.filter((u) => !u.banned).length,
      totalSources: sites.length,
      activeSources: sites.filter((s) => !s.disabled).length,
    }),
    [users, sites]
  );

  const handleLogout = async () => {
    await fetch('/api/logout', { method: 'POST' });
    router.push('/login');
  };

  const showNotice = (message: string, type: 'success' | 'error') => {
    setNotice({ message, type });
  };

  const handleCreateUser = () => {
    setUserDraft(createEmptyUserDraft());
    setUserDialogOpen(true);
  };

  const handleUserDraftChange = (field: keyof UserDraft, value: string) => {
    setUserDraft((prev) => ({ ...prev, [field]: value }));
  };

  const submitCreateUser = async () => {
    if (!userDraft.username.trim() || !userDraft.password.trim()) {
      showNotice('请填写完整的用户名和密码', 'error');
      return;
    }
    try {
      setActionSubmitting(true);
      await requestApi('/api/admin/users', {
        method: 'POST',
        body: JSON.stringify({
          username: userDraft.username.trim(),
          password: userDraft.password,
          role: 'user',
        }),
      });
      await loadAll();
      setUserDialogOpen(false);
      setUserDraft(createEmptyUserDraft());
      showNotice('用户已创建', 'success');
    } catch (e) {
      showNotice(e instanceof Error ? e.message : '创建用户失败', 'error');
    } finally {
      setActionSubmitting(false);
    }
  };

  const handleToggleBan = async (user: AdminUser) => {
    if (user.role === 'owner') return;
    try {
      await requestApi(
        `/api/admin/users/${encodeURIComponent(user.username)}`,
        {
          method: 'PUT',
          body: JSON.stringify({ banned: !user.banned }),
        }
      );
      await loadAll();
      showNotice(
        `${user.username} 已${user.banned ? '恢复' : '禁用'}`,
        'success'
      );
    } catch (e) {
      showNotice(e instanceof Error ? e.message : '更新用户状态失败', 'error');
    }
  };

  const handleDeleteUser = (user: AdminUser) => {
    if (user.role === 'owner') return;
    setPendingDeleteUser(user);
  };

  const confirmDeleteUser = async () => {
    if (!pendingDeleteUser) return;
    try {
      setActionSubmitting(true);
      await requestApi(
        `/api/admin/users/${encodeURIComponent(pendingDeleteUser.username)}`,
        {
          method: 'DELETE',
        }
      );
      await loadAll();
      showNotice(`用户 ${pendingDeleteUser.username} 已删除`, 'success');
      setPendingDeleteUser(null);
    } catch (e) {
      showNotice(e instanceof Error ? e.message : '删除用户失败', 'error');
    } finally {
      setActionSubmitting(false);
    }
  };

  const handleAddSource = () => {
    setSourceDraft(createEmptySourceDraft());
    setSourceDialogOpen(true);
  };

  const handleSourceDraftChange = (
    field: keyof SourceDraft,
    value: string
  ) => {
    setSourceDraft((prev) => ({ ...prev, [field]: value }));
  };

  const submitAddSource = async () => {
    if (
      !sourceDraft.key.trim() ||
      !sourceDraft.name.trim() ||
      !sourceDraft.api.trim()
    ) {
      showNotice('请填写完整的源信息', 'error');
      return;
    }
    try {
      setActionSubmitting(true);
      await requestApi('/api/admin/source', {
        method: 'POST',
        body: JSON.stringify({
          action: 'add',
          key: sourceDraft.key.trim(),
          name: sourceDraft.name.trim(),
          api: sourceDraft.api.trim(),
        }),
      });
      await loadAll();
      setSourceDialogOpen(false);
      setSourceDraft(createEmptySourceDraft());
      showNotice('视频源已添加', 'success');
    } catch (e) {
      showNotice(e instanceof Error ? e.message : '添加视频源失败', 'error');
    } finally {
      setActionSubmitting(false);
    }
  };

  const handleToggleSource = async (site: VideoSite) => {
    try {
      await requestApi(`/api/admin/sites/${encodeURIComponent(site.key)}`, {
        method: 'PUT',
        body: JSON.stringify({ disabled: !site.disabled }),
      });
      await loadAll();
      showNotice(`${site.name} 已${site.disabled ? '启用' : '禁用'}`, 'success');
    } catch (e) {
      showNotice(e instanceof Error ? e.message : '更新视频源状态失败', 'error');
    }
  };

  const handleDeleteSource = (site: VideoSite) => {
    setPendingDeleteSource(site);
  };

  const confirmDeleteSource = async () => {
    if (!pendingDeleteSource) return;
    try {
      setActionSubmitting(true);
      await requestApi(
        `/api/admin/sites/${encodeURIComponent(pendingDeleteSource.key)}`,
        {
          method: 'DELETE',
        }
      );
      await loadAll();
      showNotice(`视频源 ${pendingDeleteSource.name} 已删除`, 'success');
      setPendingDeleteSource(null);
    } catch (e) {
      showNotice(e instanceof Error ? e.message : '删除视频源失败', 'error');
    } finally {
      setActionSubmitting(false);
    }
  };

  const handleSaveSettings = async () => {
    try {
      setSaving(true);
      await requestApi('/api/admin/config', {
        method: 'PUT',
        body: JSON.stringify({
          SiteConfig: {
            ...siteConfig,
            DisableYellowFilter: siteConfig.ContentAccessMode !== 'safe',
          },
        }),
      });
      showNotice('站点设置已保存', 'success');
    } catch (e) {
      showNotice(e instanceof Error ? e.message : '保存站点设置失败', 'error');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className='min-h-screen bg-netflix-black flex'>
      <aside className='w-72 bg-netflix-surface border-r border-netflix-gray-800 flex flex-col'>
        <div className='p-6 border-b border-netflix-gray-800'>
          <h1 className='text-xl font-bold text-white'>管理后台</h1>
        </div>

        <nav className='flex-1 p-4 space-y-2'>
          {menuItems.map((item) => {
            const Icon = item.icon;
            return (
              <button
                key={item.id}
                onClick={() => setActiveTab(item.id)}
                className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg transition-colors ${
                  activeTab === item.id
                    ? 'bg-netflix-red text-white'
                    : 'text-netflix-gray-300 hover:bg-netflix-gray-800'
                }`}
              >
                <Icon className='w-5 h-5 shrink-0' />
                <span>{item.label}</span>
              </button>
            );
          })}
        </nav>

        <div className='p-4 border-t border-netflix-gray-800'>
          <button
            onClick={handleLogout}
            className='w-full flex items-center gap-3 px-4 py-3 text-netflix-gray-300 hover:bg-netflix-gray-800 rounded-lg transition-colors'
          >
            <LogOut className='w-5 h-5 shrink-0' />
            <span>退出登录</span>
          </button>
        </div>
      </aside>

      <main className='flex-1 overflow-auto'>
        <header className='sticky top-0 z-10 bg-netflix-black/95 backdrop-blur border-b border-netflix-gray-800 px-8 py-4'>
          <div className='flex items-center justify-between'>
            <h2 className='text-xl font-bold text-white'>
              {menuItems.find((i) => i.id === activeTab)?.label}
            </h2>
            <span className='text-netflix-gray-400'>admin</span>
          </div>
        </header>

        <div className='p-8'>
          {notice ? (
            <div
              className={`mb-6 flex items-center justify-between rounded-lg border p-4 ${
                notice.type === 'success'
                  ? 'border-green-500/30 bg-green-500/10 text-green-300'
                  : 'border-red-500/30 bg-red-500/10 text-red-300'
              }`}
            >
              <span>{notice.message}</span>
              <button
                type='button'
                onClick={() => setNotice(null)}
                className='text-current/80 transition-opacity hover:opacity-100'
              >
                关闭
              </button>
            </div>
          ) : null}

          {loading && <p className='text-netflix-gray-400'>加载中...</p>}
          {error && (
            <div className='mb-6 rounded-lg border border-red-500/30 bg-red-500/10 p-4 text-red-300'>
              {error}
            </div>
          )}

          {!loading && activeTab === 'dashboard' && (
            <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6'>
              {Object.entries(stats).map(([key, value]) => (
                <div key={key} className='bg-netflix-surface rounded-xl p-6'>
                  <div className='flex items-center justify-between mb-4'>
                    <BarChart3 className='w-8 h-8 text-netflix-red' />
                    <span className='text-2xl font-bold text-white'>
                      {value}
                    </span>
                  </div>
                  <p className='text-netflix-gray-400 capitalize'>
                    {key.replace(/([A-Z])/g, ' $1').trim()}
                  </p>
                </div>
              ))}
            </div>
          )}

          {!loading && activeTab === 'users' && (
            <div className='space-y-6'>
              <div className='flex items-center justify-between'>
                <div className='relative'>
                  <Search className='absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-netflix-gray-500' />
                  <input
                    type='text'
                    value={userKeyword}
                    onChange={(e) => setUserKeyword(e.target.value)}
                    placeholder='搜索用户...'
                    className='pl-10 pr-4 py-2 bg-netflix-surface border border-netflix-gray-800 rounded-lg text-white placeholder-netflix-gray-500 focus:outline-none focus:border-netflix-red'
                  />
                </div>
                <button
                  onClick={handleCreateUser}
                  className='flex items-center gap-2 px-4 py-2 bg-netflix-red text-white rounded-lg hover:bg-netflix-red-hover transition-colors'
                >
                  <Plus className='w-5 h-5' />
                  添加用户
                </button>
              </div>

              <div className='bg-netflix-surface rounded-xl overflow-hidden'>
                <table className='w-full'>
                  <thead className='bg-netflix-gray-900'>
                    <tr>
                      <th className='px-6 py-4 text-left text-netflix-gray-400'>
                        用户名
                      </th>
                      <th className='px-6 py-4 text-left text-netflix-gray-400'>
                        角色
                      </th>
                      <th className='px-6 py-4 text-left text-netflix-gray-400'>
                        状态
                      </th>
                      <th className='px-6 py-4 text-right text-netflix-gray-400'>
                        操作
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredUsers.map((user) => {
                      const active = !user.banned;
                      return (
                        <tr
                          key={user.username}
                          className='border-t border-netflix-gray-800'
                        >
                          <td className='px-6 py-4 text-white'>
                            {user.username}
                          </td>
                          <td className='px-6 py-4'>
                            <span className='px-2 py-1 bg-netflix-gray-800 rounded text-sm text-netflix-gray-300'>
                              {user.role}
                            </span>
                          </td>
                          <td className='px-6 py-4'>
                            <span
                              className={`flex items-center gap-1 ${
                                active ? 'text-green-400' : 'text-red-400'
                              }`}
                            >
                              {active ? (
                                <CheckCircle className='w-4 h-4' />
                              ) : (
                                <XCircle className='w-4 h-4' />
                              )}
                              {active ? '正常' : '禁用'}
                            </span>
                          </td>
                          <td className='px-6 py-4'>
                            <div className='flex justify-end gap-2'>
                              <button
                                onClick={() => handleToggleBan(user)}
                                className='p-2 text-netflix-gray-400 hover:text-white transition-colors'
                                title={active ? '禁用' : '启用'}
                                disabled={user.role === 'owner'}
                              >
                                <Edit3 className='w-4 h-4' />
                              </button>
                              <button
                                onClick={() => handleDeleteUser(user)}
                                className='p-2 text-netflix-gray-400 hover:text-red-400 transition-colors'
                                disabled={user.role === 'owner'}
                              >
                                <Trash2 className='w-4 h-4' />
                              </button>
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {!loading && activeTab === 'sources' && (
            <div className='space-y-6'>
              <div className='flex items-center justify-between'>
                <h3 className='text-lg font-bold text-white'>视频源管理</h3>
                <button
                  onClick={handleAddSource}
                  className='flex items-center gap-2 px-4 py-2 bg-netflix-red text-white rounded-lg hover:bg-netflix-red-hover transition-colors'
                >
                  <Plus className='w-5 h-5' />
                  添加源
                </button>
              </div>

              <div className='grid gap-4'>
                {sites.map((site) => {
                  const active = !site.disabled;
                  return (
                    <div
                      key={site.key}
                      className='bg-netflix-surface rounded-xl p-6 flex items-center justify-between'
                    >
                      <div>
                        <h4 className='text-white font-bold mb-1'>
                          {site.name}
                        </h4>
                        <p className='text-netflix-gray-400 text-sm'>
                          {site.api}
                        </p>
                      </div>

                      <div className='flex items-center gap-4'>
                        <span
                          className={`px-3 py-1 rounded-full text-sm ${
                            active
                              ? 'bg-green-500/20 text-green-400'
                              : 'bg-red-500/20 text-red-400'
                          }`}
                        >
                          {active ? '正常' : '禁用'}
                        </span>

                        <div className='flex items-center gap-2'>
                          <button
                            onClick={() => handleToggleSource(site)}
                            className='p-2 text-netflix-gray-400 hover:text-white transition-colors'
                          >
                            <Edit3 className='w-5 h-5' />
                          </button>
                          <button
                            onClick={() => handleDeleteSource(site)}
                            className='p-2 text-netflix-gray-400 hover:text-red-400 transition-colors'
                          >
                            <Trash2 className='w-5 h-5' />
                          </button>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {!loading && activeTab === 'settings' && (
            <AdminSettingsPanel
              siteConfig={siteConfig}
              saving={saving}
              contentAccessModeOptions={contentAccessModeOptions}
              onConfigChange={setSiteConfig}
              onSave={handleSaveSettings}
            />
          )}
        </div>
      </main>
      <AdminPageDialogs
        actionSubmitting={actionSubmitting}
        userDialogOpen={userDialogOpen}
        userDraft={userDraft}
        onUserDialogClose={() => {
          if (!actionSubmitting) setUserDialogOpen(false);
        }}
        onUserDraftChange={handleUserDraftChange}
        onUserSubmit={submitCreateUser}
        sourceDialogOpen={sourceDialogOpen}
        sourceDraft={sourceDraft}
        onSourceDialogClose={() => {
          if (!actionSubmitting) setSourceDialogOpen(false);
        }}
        onSourceDraftChange={handleSourceDraftChange}
        onSourceSubmit={submitAddSource}
        pendingDeleteUserName={pendingDeleteUser?.username || null}
        onPendingDeleteUserClose={() => {
          if (!actionSubmitting) setPendingDeleteUser(null);
        }}
        onPendingDeleteUserConfirm={confirmDeleteUser}
        pendingDeleteSourceName={pendingDeleteSource?.name || null}
        onPendingDeleteSourceClose={() => {
          if (!actionSubmitting) setPendingDeleteSource(null);
        }}
        onPendingDeleteSourceConfirm={confirmDeleteSource}
      />
    </div>
  );
}
