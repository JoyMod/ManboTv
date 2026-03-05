'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useRouter } from 'next/navigation';
import {
  LayoutDashboard,
  Users,
  Settings,
  Tv,
  Database,
  LogOut,
  Menu,
  X,
  ChevronRight,
  Plus,
  Trash2,
  Edit3,
  Save,
  RefreshCw,
  AlertCircle,
  CheckCircle,
  Search,
  Film,
  Radio,
  FileJson,
  Shield,
  Download,
  Upload,
  BarChart3,
  TrendingUp,
  Activity,
} from 'lucide-react';

// 类型定义
interface User {
  username: string;
  role: 'owner' | 'admin' | 'user';
  disabled?: boolean;
  tags?: string[];
  enabledApis?: string[];
}

interface VideoSource {
  key: string;
  name: string;
  api: string;
  detail?: string;
  disabled?: boolean;
}

interface LiveSource {
  key: string;
  name: string;
  url: string;
  ua?: string;
  epg?: string;
  disabled?: boolean;
  channelNumber?: number;
}

interface CustomCategory {
  name: string;
  type: 'movie' | 'tv';
  query: string;
  disabled?: boolean;
}

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
}

interface AdminConfig {
  SiteConfig: SiteConfig;
  UserConfig: {
    Users: User[];
    Tags?: { name: string; enabledApis: string[] }[];
  };
  DataSourceConfig: {
    DataSources: VideoSource[];
  };
  LiveConfig: {
    DataSources: LiveSource[];
  };
  CustomCategories: CustomCategory[];
}

// 菜单项
const menuItems = [
  { id: 'dashboard', label: '概览', icon: LayoutDashboard },
  { id: 'users', label: '用户管理', icon: Users },
  { id: 'sources', label: '视频源', icon: Film },
  { id: 'live', label: '直播源', icon: Radio },
  { id: 'categories', label: '分类管理', icon: Database },
  { id: 'data', label: '数据管理', icon: FileJson },
  { id: 'settings', label: '站点设置', icon: Settings },
];

// 简单的 Toast 提示
function useToast() {
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);
  
  const showToast = useCallback((message: string, type: 'success' | 'error' = 'success') => {
    setToast({ message, type });
    setTimeout(() => setToast(null), 3000);
  }, []);
  
  return { toast, showToast };
}

// 确认对话框组件
function ConfirmDialog({
  isOpen,
  title,
  message,
  onConfirm,
  onCancel,
}: {
  isOpen: boolean;
  title: string;
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  if (!isOpen) return null;
  
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm">
      <motion.div
        initial={{ scale: 0.9, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        className="bg-[#1a1a1a] rounded-xl p-6 max-w-md w-full mx-4 border border-gray-800"
      >
        <h3 className="text-xl font-bold text-white mb-2">{title}</h3>
        <p className="text-gray-400 mb-6">{message}</p>
        <div className="flex gap-3 justify-end">
          <button
            onClick={onCancel}
            className="px-4 py-2 text-sm text-gray-300 hover:text-white transition-colors"
          >
            取消
          </button>
          <button
            onClick={onConfirm}
            className="px-4 py-2 bg-[#E50914] text-white rounded-lg hover:bg-red-600 transition-colors"
          >
            确认
          </button>
        </div>
      </motion.div>
    </div>
  );
}

// 加载组件
function LoadingSpinner() {
  return (
    <div className="flex items-center justify-center py-12">
      <RefreshCw className="w-8 h-8 text-[#E50914] animate-spin" />
    </div>
  );
}

// 卡片组件
function Card({ children, className = '' }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={`bg-[#1a1a1a] rounded-xl border border-gray-800 ${className}`}>
      {children}
    </div>
  );
}

// 按钮组件
function Button({
  children,
  onClick,
  variant = 'primary',
  disabled = false,
  className = '',
}: {
  children: React.ReactNode;
  onClick?: () => void;
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
  disabled?: boolean;
  className?: string;
}) {
  const variants = {
    primary: 'bg-[#E50914] hover:bg-red-600 text-white',
    secondary: 'bg-gray-700 hover:bg-gray-600 text-white',
    danger: 'bg-red-600 hover:bg-red-700 text-white',
    ghost: 'text-gray-400 hover:text-white',
  };
  
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`px-4 py-2 rounded-lg font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed ${variants[variant]} ${className}`}
    >
      {children}
    </button>
  );
}

// 输入框组件
function Input({
  value,
  onChange,
  placeholder,
  type = 'text',
  disabled = false,
}: {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  type?: string;
  disabled?: boolean;
}) {
  return (
    <input
      type={type}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      disabled={disabled}
      className="w-full px-4 py-2 bg-[#2a2a2a] border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-[#E50914] transition-colors disabled:opacity-50"
    />
  );
}

// 开关组件
function Switch({ checked, onChange }: { checked: boolean; onChange: (checked: boolean) => void }) {
  return (
    <button
      onClick={() => onChange(!checked)}
      className={`w-12 h-6 rounded-full transition-colors ${checked ? 'bg-[#E50914]' : 'bg-gray-700'}`}
    >
      <div
        className={`w-5 h-5 rounded-full bg-white transition-transform ${checked ? 'translate-x-6' : 'translate-x-0.5'}`}
      />
    </button>
  );
}

// 统计卡片
function StatCard({ title, value, icon: Icon, trend }: { title: string; value: number; icon: any; trend?: string }) {
  return (
    <Card className="p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-gray-400 text-sm">{title}</p>
          <p className="text-3xl font-bold text-white mt-1">{value}</p>
          {trend && <p className="text-green-400 text-xs mt-1">{trend}</p>}
        </div>
        <div className="w-12 h-12 bg-[#E50914]/10 rounded-lg flex items-center justify-center">
          <Icon className="w-6 h-6 text-[#E50914]" />
        </div>
      </div>
    </Card>
  );
}

export default function AdminPage() {
  const router = useRouter();
  const { toast, showToast } = useToast();
  const [activeTab, setActiveTab] = useState('dashboard');
  const [isSidebarOpen, setIsSidebarOpen] = useState(true);
  const [config, setConfig] = useState<AdminConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [confirmDialog, setConfirmDialog] = useState<{
    isOpen: boolean;
    title: string;
    message: string;
    onConfirm: () => void;
  }>({ isOpen: false, title: '', message: '', onConfirm: () => undefined });

  // 获取配置
  const fetchConfig = useCallback(async () => {
    try {
      const res = await fetch('/api/admin/config');
      if (!res.ok) throw new Error('获取配置失败');
      const data = await res.json();
      setConfig(data);
    } catch (err) {
      showToast('获取配置失败', 'error');
    } finally {
      setLoading(false);
    }
  }, [showToast]);

  useEffect(() => {
    fetchConfig();
  }, [fetchConfig]);

  // 保存站点配置
  const saveSiteConfig = async (siteConfig: SiteConfig) => {
    try {
      const res = await fetch('/api/admin/config', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...config, SiteConfig: siteConfig }),
      });
      if (!res.ok) throw new Error('保存失败');
      showToast('保存成功');
      fetchConfig();
    } catch (err) {
      showToast('保存失败', 'error');
    }
  };

  // 退出登录
  const handleLogout = async () => {
    await fetch('/api/logout', { method: 'POST' });
    router.push('/login');
  };

  // 显示确认对话框
  const showConfirm = (title: string, message: string, onConfirm: () => void) => {
    setConfirmDialog({ isOpen: true, title, message, onConfirm });
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-[#141414] flex items-center justify-center">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#141414] flex">
      {/* Toast 提示 */}
      <AnimatePresence>
        {toast && (
          <motion.div
            initial={{ opacity: 0, y: -50 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -50 }}
            className={`fixed top-4 right-4 z-50 px-6 py-3 rounded-lg flex items-center gap-2 ${
              toast.type === 'success' ? 'bg-green-600' : 'bg-red-600'
            }`}
          >
            {toast.type === 'success' ? <CheckCircle className="w-5 h-5" /> : <AlertCircle className="w-5 h-5" />}
            <span className="text-white">{toast.message}</span>
          </motion.div>
        )}
      </AnimatePresence>

      {/* 确认对话框 */}
      <ConfirmDialog
        isOpen={confirmDialog.isOpen}
        title={confirmDialog.title}
        message={confirmDialog.message}
        onConfirm={() => {
          confirmDialog.onConfirm();
          setConfirmDialog(prev => ({ ...prev, isOpen: false }));
        }}
        onCancel={() => setConfirmDialog(prev => ({ ...prev, isOpen: false }))}
      />

      {/* 侧边栏 */}
      <motion.aside
        initial={false}
        animate={{ width: isSidebarOpen ? 260 : 80 }}
        className="bg-[#0a0a0a] border-r border-gray-800 flex flex-col"
      >
        {/* Logo */}
        <div className="h-16 flex items-center px-6 border-b border-gray-800">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-[#E50914] rounded flex items-center justify-center flex-shrink-0">
              <span className="text-white font-bold text-lg">M</span>
            </div>
            {isSidebarOpen && (
              <motion.span
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="text-white font-bold text-lg whitespace-nowrap"
              >
                ManBoTV
              </motion.span>
            )}
          </div>
        </div>

        {/* 菜单 */}
        <nav className="flex-1 py-4 px-3 space-y-1">
          {menuItems.map((item) => {
            const Icon = item.icon;
            const isActive = activeTab === item.id;
            return (
              <button
                key={item.id}
                onClick={() => setActiveTab(item.id)}
                className={`w-full flex items-center gap-3 px-3 py-3 rounded-lg transition-all ${
                  isActive
                    ? 'bg-[#E50914] text-white'
                    : 'text-gray-400 hover:bg-gray-800 hover:text-white'
                }`}
              >
                <Icon className="w-5 h-5 flex-shrink-0" />
                {isSidebarOpen && (
                  <span className="font-medium whitespace-nowrap">{item.label}</span>
                )}
              </button>
            );
          })}
        </nav>

        {/* 底部操作 */}
        <div className="p-3 border-t border-gray-800 space-y-1">
          <button
            onClick={() => setIsSidebarOpen(!isSidebarOpen)}
            className="w-full flex items-center gap-3 px-3 py-3 rounded-lg text-gray-400 hover:bg-gray-800 hover:text-white transition-all"
          >
            {isSidebarOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            {isSidebarOpen && <span className="font-medium whitespace-nowrap">收起菜单</span>}
          </button>
          <button
            onClick={handleLogout}
            className="w-full flex items-center gap-3 px-3 py-3 rounded-lg text-gray-400 hover:bg-gray-800 hover:text-white transition-all"
          >
            <LogOut className="w-5 h-5" />
            {isSidebarOpen && <span className="font-medium whitespace-nowrap">退出登录</span>}
          </button>
        </div>
      </motion.aside>

      {/* 主内容区 */}
      <main className="flex-1 overflow-auto">
        {/* 顶部栏 */}
        <header className="h-16 bg-[#0a0a0a]/80 backdrop-blur-md border-b border-gray-800 sticky top-0 z-10 px-8 flex items-center justify-between">
          <h1 className="text-xl font-bold text-white">
            {menuItems.find((m) => m.id === activeTab)?.label}
          </h1>
          <div className="flex items-center gap-4">
            <span className="text-gray-400 text-sm">管理员</span>
            <div className="w-8 h-8 bg-gradient-to-br from-[#E50914] to-[#FF6B00] rounded-full flex items-center justify-center">
              <span className="text-white font-bold text-sm">A</span>
            </div>
          </div>
        </header>

        {/* 内容 */}
        <div className="p-8">
          {activeTab === 'dashboard' && (
            <DashboardTab config={config} />
          )}
          {activeTab === 'users' && (
            <UsersTab
              users={config?.UserConfig?.Users || []}
              onRefresh={fetchConfig}
              showToast={showToast}
              showConfirm={showConfirm}
            />
          )}
          {activeTab === 'sources' && (
            <SourcesTab
              sources={config?.DataSourceConfig?.DataSources || []}
              onRefresh={fetchConfig}
              showToast={showToast}
              showConfirm={showConfirm}
            />
          )}
          {activeTab === 'live' && (
            <LiveTab
              sources={config?.LiveConfig?.DataSources || []}
              onRefresh={fetchConfig}
              showToast={showToast}
              showConfirm={showConfirm}
            />
          )}
          {activeTab === 'categories' && (
            <CategoriesTab
              categories={config?.CustomCategories || []}
              onRefresh={fetchConfig}
              showToast={showToast}
              showConfirm={showConfirm}
            />
          )}
          {activeTab === 'data' && (
            <DataManagementTab
              showToast={showToast}
            />
          )}
          {activeTab === 'settings' && (
            <SettingsTab
              config={config?.SiteConfig}
              onSave={saveSiteConfig}
            />
          )}
        </div>
      </main>
    </div>
  );
}

// 概览页
function DashboardTab({ config }: { config: AdminConfig | null }) {
  const stats = [
    { title: '总用户数', value: config?.UserConfig?.Users?.length || 0, icon: Users, trend: '+5% 本月新增' },
    { title: '视频源', value: config?.DataSourceConfig?.DataSources?.length || 0, icon: Film, trend: '2 个禁用' },
    { title: '直播源', value: config?.LiveConfig?.DataSources?.length || 0, icon: Radio, trend: '正常' },
    { title: '自定义分类', value: config?.CustomCategories?.length || 0, icon: Database },
  ];

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {stats.map((stat) => (
          <StatCard key={stat.title} {...stat} />
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="p-6">
          <h3 className="text-lg font-bold text-white mb-4">站点信息</h3>
          <div className="space-y-4">
            <div className="flex justify-between py-3 border-b border-gray-800">
              <span className="text-gray-400">站点名称</span>
              <span className="text-white">{config?.SiteConfig?.SiteName || 'ManBoTV'}</span>
            </div>
            <div className="flex justify-between py-3 border-b border-gray-800">
              <span className="text-gray-400">豆瓣代理</span>
              <span className="text-white">{config?.SiteConfig?.DoubanProxyType || '默认'}</span>
            </div>
            <div className="flex justify-between py-3 border-b border-gray-800">
              <span className="text-gray-400">流式搜索</span>
              <span className="text-white">{config?.SiteConfig?.FluidSearch ? '启用' : '禁用'}</span>
            </div>
            <div className="flex justify-between py-3">
              <span className="text-gray-400">黄暴过滤</span>
              <span className="text-white">{config?.SiteConfig?.DisableYellowFilter ? '禁用' : '启用'}</span>
            </div>
          </div>
        </Card>

        <Card className="p-6">
          <h3 className="text-lg font-bold text-white mb-4">系统状态</h3>
          <div className="space-y-4">
            <div className="flex items-center gap-3 py-3 border-b border-gray-800">
              <div className="w-2 h-2 rounded-full bg-green-500" />
              <span className="text-gray-400">API 服务</span>
              <span className="text-green-400 ml-auto">正常运行</span>
            </div>
            <div className="flex items-center gap-3 py-3 border-b border-gray-800">
              <div className="w-2 h-2 rounded-full bg-green-500" />
              <span className="text-gray-400">Redis 缓存</span>
              <span className="text-green-400 ml-auto">已连接</span>
            </div>
            <div className="flex items-center gap-3 py-3 border-b border-gray-800">
              <div className="w-2 h-2 rounded-full bg-green-500" />
              <span className="text-gray-400">图片代理</span>
              <span className="text-green-400 ml-auto">正常</span>
            </div>
            <div className="flex items-center gap-3 py-3">
              <div className="w-2 h-2 rounded-full bg-blue-500" />
              <span className="text-gray-400">版本</span>
              <span className="text-white ml-auto">v1.0.0</span>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}

// 用户管理页
function UsersTab({
  users,
  onRefresh,
  showToast,
  showConfirm,
}: {
  users: User[];
  onRefresh: () => void;
  showToast: (msg: string, type?: 'success' | 'error') => void;
  showConfirm: (title: string, msg: string, onConfirm: () => void) => void;
}) {
  const [showAddForm, setShowAddForm] = useState(false);
  const [newUser, setNewUser] = useState<{ username: string; password: string; role: 'user' | 'admin' }>({ username: '', password: '', role: 'user' });
  const [editingUser, setEditingUser] = useState<string | null>(null);
  const [editPassword, setEditPassword] = useState('');

  const handleAddUser = async () => {
    if (!newUser.username || !newUser.password) {
      showToast('请填写完整信息', 'error');
      return;
    }
    try {
      const res = await fetch('/api/admin/users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newUser),
      });
      if (!res.ok) throw new Error('添加失败');
      showToast('添加用户成功');
      setNewUser({ username: '', password: '', role: 'user' });
      setShowAddForm(false);
      onRefresh();
    } catch (err) {
      showToast('添加用户失败', 'error');
    }
  };

  const handleDeleteUser = (username: string) => {
    showConfirm('删除用户', `确定要删除用户 "${username}" 吗？`, async () => {
      try {
        const res = await fetch(`/api/admin/users/${username}`, { method: 'DELETE' });
        if (!res.ok) throw new Error('删除失败');
        showToast('删除成功');
        onRefresh();
      } catch (err) {
        showToast('删除失败', 'error');
      }
    });
  };

  const handleToggleRole = async (user: User) => {
    const newRole = user.role === 'user' ? 'admin' : 'user';
    try {
      const res = await fetch(`/api/admin/users/${user.username}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role: newRole }),
      });
      if (!res.ok) throw new Error('更新失败');
      showToast('权限更新成功');
      onRefresh();
    } catch (err) {
      showToast('更新失败', 'error');
    }
  };

  const handleChangePassword = async (username: string) => {
    if (!editPassword) {
      showToast('请输入新密码', 'error');
      return;
    }
    try {
      const res = await fetch(`/api/admin/users/${username}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password: editPassword }),
      });
      if (!res.ok) throw new Error('修改失败');
      showToast('密码修改成功');
      setEditingUser(null);
      setEditPassword('');
      onRefresh();
    } catch (err) {
      showToast('修改失败', 'error');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-lg font-bold text-white">用户列表</h2>
        <Button onClick={() => setShowAddForm(!showAddForm)}>
          <Plus className="w-4 h-4 mr-2" />
          添加用户
        </Button>
      </div>

      {showAddForm && (
        <Card className="p-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Input
              value={newUser.username}
              onChange={(v) => setNewUser({ ...newUser, username: v })}
              placeholder="用户名"
            />
            <Input
              value={newUser.password}
              onChange={(v) => setNewUser({ ...newUser, password: v })}
              placeholder="密码"
              type="password"
            />
            <div className="flex gap-2">
              <select
                value={newUser.role}
                onChange={(e) => setNewUser({ ...newUser, role: e.target.value as 'user' | 'admin' })}
                className="flex-1 px-4 py-2 bg-[#2a2a2a] border border-gray-700 rounded-lg text-white"
              >
                <option value="user">普通用户</option>
                <option value="admin">管理员</option>
              </select>
              <Button onClick={handleAddUser}>添加</Button>
            </div>
          </div>
        </Card>
      )}

      <Card>
        <table className="w-full">
          <thead className="bg-[#2a2a2a]">
            <tr>
              <th className="px-6 py-4 text-left text-sm font-medium text-gray-400">用户名</th>
              <th className="px-6 py-4 text-left text-sm font-medium text-gray-400">角色</th>
              <th className="px-6 py-4 text-left text-sm font-medium text-gray-400">状态</th>
              <th className="px-6 py-4 text-right text-sm font-medium text-gray-400">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-800">
            {users.map((user) => (
              <tr key={user.username} className="hover:bg-[#2a2a2a]/50">
                <td className="px-6 py-4 text-white">{user.username}</td>
                <td className="px-6 py-4">
                  <span
                    className={`px-2 py-1 rounded text-xs font-medium ${
                      user.role === 'owner'
                        ? 'bg-purple-600 text-white'
                        : user.role === 'admin'
                        ? 'bg-blue-600 text-white'
                        : 'bg-gray-600 text-white'
                    }`}
                  >
                    {user.role === 'owner' ? '所有者' : user.role === 'admin' ? '管理员' : '用户'}
                  </span>
                </td>
                <td className="px-6 py-4">
                  <span className={`text-sm ${user.disabled ? 'text-red-400' : 'text-green-400'}`}>
                    {user.disabled ? '禁用' : '正常'}
                  </span>
                </td>
                <td className="px-6 py-4 text-right space-x-2">
                  {user.role !== 'owner' && (
                    <>
                      {editingUser === user.username ? (
                        <>
                          <Input
                            value={editPassword}
                            onChange={setEditPassword}
                            placeholder="新密码"
                            type="password"
                          />
                          <Button variant="ghost" onClick={() => handleChangePassword(user.username)}>
                            <Save className="w-4 h-4" />
                          </Button>
                        </>
                      ) : (
                        <Button variant="ghost" onClick={() => setEditingUser(user.username)}>
                          <Edit3 className="w-4 h-4" />
                        </Button>
                      )}
                      <Button variant="ghost" onClick={() => handleToggleRole(user)}>
                        <Shield className="w-4 h-4" />
                      </Button>
                      <Button variant="ghost" onClick={() => handleDeleteUser(user.username)}>
                        <Trash2 className="w-4 h-4 text-red-400" />
                      </Button>
                    </>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </div>
  );
}

// 视频源管理页
function SourcesTab({
  sources,
  onRefresh,
  showToast,
  showConfirm,
}: {
  sources: VideoSource[];
  onRefresh: () => void;
  showToast: (msg: string, type?: 'success' | 'error') => void;
  showConfirm: (title: string, msg: string, onConfirm: () => void) => void;
}) {
  const [showAddForm, setShowAddForm] = useState(false);
  const [newSource, setNewSource] = useState({ key: '', name: '', api: '', detail: '' });

  const handleAdd = async () => {
    if (!newSource.key || !newSource.name || !newSource.api) {
      showToast('请填写完整信息', 'error');
      return;
    }
    try {
      const res = await fetch('/api/admin/source', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          action: 'add',
          key: newSource.key,
          name: newSource.name,
          api: newSource.api,
          detail: newSource.detail,
        }),
      });
      if (!res.ok) throw new Error('添加失败');
      showToast('添加成功');
      setNewSource({ key: '', name: '', api: '', detail: '' });
      setShowAddForm(false);
      onRefresh();
    } catch (err) {
      showToast('添加失败', 'error');
    }
  };

  const handleToggle = async (source: VideoSource) => {
    try {
      const res = await fetch('/api/admin/source', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          action: 'toggle',
          key: source.key,
        }),
      });
      if (!res.ok) throw new Error('更新失败');
      showToast('更新成功');
      onRefresh();
    } catch (err) {
      showToast('更新失败', 'error');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-lg font-bold text-white">视频源列表</h2>
        <Button onClick={() => setShowAddForm(!showAddForm)}>
          <Plus className="w-4 h-4 mr-2" />
          添加源
        </Button>
      </div>

      {showAddForm && (
        <Card className="p-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input value={newSource.key} onChange={(v) => setNewSource({ ...newSource, key: v })} placeholder="唯一标识" />
            <Input value={newSource.name} onChange={(v) => setNewSource({ ...newSource, name: v })} placeholder="名称" />
            <Input value={newSource.api} onChange={(v) => setNewSource({ ...newSource, api: v })} placeholder="API地址" />
            <Input value={newSource.detail} onChange={(v) => setNewSource({ ...newSource, detail: v })} placeholder="详情地址（可选）" />
          </div>
          <div className="mt-4 flex justify-end">
            <Button onClick={handleAdd}>添加</Button>
          </div>
        </Card>
      )}

      <div className="grid gap-4">
        {sources.map((source) => (
          <Card key={source.key} className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center gap-3">
                  <h3 className="text-white font-medium">{source.name}</h3>
                  {source.disabled && (
                    <span className="px-2 py-0.5 bg-red-600 text-white text-xs rounded">已禁用</span>
                  )}
                </div>
                <p className="text-gray-500 text-sm mt-1">{source.api}</p>
              </div>
              <Switch checked={!source.disabled} onChange={() => handleToggle(source)} />
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}

// 直播源管理页
function LiveTab({
  sources,
  onRefresh,
  showToast,
  showConfirm,
}: {
  sources: LiveSource[];
  onRefresh: () => void;
  showToast: (msg: string, type?: 'success' | 'error') => void;
  showConfirm: (title: string, msg: string, onConfirm: () => void) => void;
}) {
  return (
    <div className="space-y-6">
      <Card className="p-6">
        <div className="flex items-center gap-4">
          <RefreshCw className="w-8 h-8 text-[#E50914]" />
          <div>
            <h3 className="text-white font-bold">直播源管理</h3>
            <p className="text-gray-400 text-sm">当前共有 {sources.length} 个直播源</p>
          </div>
        </div>
      </Card>

      <div className="grid gap-4">
        {sources.map((source) => (
          <Card key={source.key} className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-white font-medium">{source.name}</h3>
                <p className="text-gray-500 text-sm mt-1 truncate max-w-md">{source.url}</p>
              </div>
              {source.channelNumber && (
                <span className="text-gray-400 text-sm">{source.channelNumber} 个频道</span>
              )}
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}

// 分类管理页
function CategoriesTab({
  categories,
  onRefresh,
  showToast,
  showConfirm,
}: {
  categories: CustomCategory[];
  onRefresh: () => void;
  showToast: (msg: string, type?: 'success' | 'error') => void;
  showConfirm: (title: string, msg: string, onConfirm: () => void) => void;
}) {
  const [showAddForm, setShowAddForm] = useState(false);
  const [newCategory, setNewCategory] = useState<{ name: string; type: 'movie' | 'tv'; query: string }>({ name: '', type: 'movie', query: '' });

  const handleAdd = async () => {
    if (!newCategory.name || !newCategory.query) {
      showToast('请填写完整信息', 'error');
      return;
    }
    try {
      const res = await fetch('/api/admin/category', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newCategory),
      });
      if (!res.ok) throw new Error('添加失败');
      showToast('添加成功');
      setNewCategory({ name: '', type: 'movie', query: '' });
      setShowAddForm(false);
      onRefresh();
    } catch (err) {
      showToast('添加失败', 'error');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-lg font-bold text-white">自定义分类</h2>
        <Button onClick={() => setShowAddForm(!showAddForm)}>
          <Plus className="w-4 h-4 mr-2" />
          添加分类
        </Button>
      </div>

      {showAddForm && (
        <Card className="p-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Input value={newCategory.name} onChange={(v) => setNewCategory({ ...newCategory, name: v })} placeholder="分类名称" />
            <select
              value={newCategory.type}
              onChange={(e) => setNewCategory({ ...newCategory, type: e.target.value as 'movie' | 'tv' })}
              className="px-4 py-2 bg-[#2a2a2a] border border-gray-700 rounded-lg text-white"
            >
              <option value="movie">电影</option>
              <option value="tv">电视剧</option>
            </select>
            <Input value={newCategory.query} onChange={(v) => setNewCategory({ ...newCategory, query: v })} placeholder="搜索关键词" />
          </div>
          <div className="mt-4 flex justify-end">
            <Button onClick={handleAdd}>添加</Button>
          </div>
        </Card>
      )}

      <div className="grid gap-4">
        {categories.map((cat, idx) => (
          <Card key={idx} className="p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <span className={`px-2 py-1 rounded text-xs ${cat.type === 'movie' ? 'bg-blue-600' : 'bg-purple-600'} text-white`}>
                  {cat.type === 'movie' ? '电影' : '电视剧'}
                </span>
                <span className="text-white font-medium">{cat.name}</span>
                <span className="text-gray-500 text-sm">{cat.query}</span>
              </div>
              {cat.disabled && <span className="text-red-400 text-sm">已禁用</span>}
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}

// 数据管理页
function DataManagementTab({
  showToast,
}: {
  showToast: (msg: string, type?: 'success' | 'error') => void;
}) {
  const [isExporting, setIsExporting] = useState(false);
  const [isImporting, setIsImporting] = useState(false);

  const handleExport = async () => {
    setIsExporting(true);
    try {
      const res = await fetch('/api/admin/data_migration/export');
      if (!res.ok) throw new Error('导出失败');
      
      const data = await res.json();
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      
      const a = document.createElement('a');
      a.href = url;
      a.download = `manbotv-backup-${new Date().toISOString().split('T')[0]}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      
      showToast('数据导出成功');
    } catch (err) {
      showToast('数据导出失败', 'error');
    } finally {
      setIsExporting(false);
    }
  };

  const handleImport = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setIsImporting(true);
    try {
      const text = await file.text();
      const data = JSON.parse(text);

      const res = await fetch('/api/admin/data_migration/import', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });

      if (!res.ok) throw new Error('导入失败');
      showToast('数据导入成功');
    } catch (err) {
      showToast('数据导入失败', 'error');
    } finally {
      setIsImporting(false);
      e.target.value = '';
    }
  };

  return (
    <div className="space-y-6">
      <Card className="p-6">
        <h3 className="text-lg font-bold text-white mb-4">数据备份与恢复</h3>
        <p className="text-gray-400 mb-6">导出所有配置数据用于备份，或从备份文件恢复数据。</p>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* 导出 */}
          <div className="p-6 bg-[#2a2a2a] rounded-lg">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 bg-blue-600/20 rounded-lg flex items-center justify-center">
                <Download className="w-5 h-5 text-blue-500" />
              </div>
              <div>
                <h4 className="text-white font-medium">导出数据</h4>
                <p className="text-gray-500 text-sm">备份所有配置</p>
              </div>
            </div>
            <Button onClick={handleExport} disabled={isExporting}>
              {isExporting ? <RefreshCw className="w-4 h-4 mr-2 animate-spin" /> : <Download className="w-4 h-4 mr-2" />}
              {isExporting ? '导出中...' : '导出数据'}
            </Button>
          </div>

          {/* 导入 */}
          <div className="p-6 bg-[#2a2a2a] rounded-lg">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 bg-green-600/20 rounded-lg flex items-center justify-center">
                <Upload className="w-5 h-5 text-green-500" />
              </div>
              <div>
                <h4 className="text-white font-medium">导入数据</h4>
                <p className="text-gray-500 text-sm">从备份恢复</p>
              </div>
            </div>
            <label className="block">
              <input
                type="file"
                accept=".json"
                onChange={handleImport}
                disabled={isImporting}
                className="hidden"
              />
              <span className={`inline-flex items-center px-4 py-2 rounded-lg font-medium cursor-pointer transition-all ${
                isImporting ? 'opacity-50' : ''
              } bg-[#E50914] hover:bg-red-600 text-white`}>
                {isImporting ? <RefreshCw className="w-4 h-4 mr-2 animate-spin" /> : <Upload className="w-4 h-4 mr-2" />}
                {isImporting ? '导入中...' : '选择文件'}
              </span>
            </label>
          </div>
        </div>
      </Card>

      <Card className="p-6">
        <h3 className="text-lg font-bold text-white mb-4">数据清理</h3>
        <div className="space-y-4">
          <div className="flex items-center justify-between py-3 border-b border-gray-800">
            <div>
              <span className="text-white font-medium">清理搜索历史</span>
              <p className="text-gray-500 text-sm">删除所有用户的搜索历史记录</p>
            </div>
            <Button variant="danger">清理</Button>
          </div>
          <div className="flex items-center justify-between py-3 border-b border-gray-800">
            <div>
              <span className="text-white font-medium">清理播放记录</span>
              <p className="text-gray-500 text-sm">删除所有用户的播放历史记录</p>
            </div>
            <Button variant="danger">清理</Button>
          </div>
          <div className="flex items-center justify-between py-3">
            <div>
              <span className="text-white font-medium">重置所有配置</span>
              <p className="text-gray-500 text-sm">恢复默认配置（不可逆）</p>
            </div>
            <Button variant="danger">重置</Button>
          </div>
        </div>
      </Card>
    </div>
  );
}

// 站点设置页
function SettingsTab({
  config,
  onSave,
}: {
  config?: SiteConfig;
  onSave: (config: SiteConfig) => void;
}) {
  const [form, setForm] = useState<SiteConfig>(
    config || {
      SiteName: 'ManBoTV',
      Announcement: '',
      SearchDownstreamMaxPage: 5,
      SiteInterfaceCacheTime: 3600,
      DoubanProxyType: 'cmliussss-cdn-tencent',
      DoubanProxy: '',
      DoubanImageProxyType: 'cmliussss-cdn-tencent',
      DoubanImageProxy: '',
      DisableYellowFilter: false,
      FluidSearch: true,
    }
  );

  useEffect(() => {
    if (config) setForm(config);
  }, [config]);

  return (
    <div className="max-w-2xl">
      <Card className="p-6">
        <h2 className="text-lg font-bold text-white mb-6">基本设置</h2>
        
        <div className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-gray-400 mb-2">站点名称</label>
            <Input value={form.SiteName} onChange={(v) => setForm({ ...form, SiteName: v })} />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-400 mb-2">公告</label>
            <textarea
              value={form.Announcement}
              onChange={(e) => setForm({ ...form, Announcement: e.target.value })}
              rows={3}
              className="w-full px-4 py-2 bg-[#2a2a2a] border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-[#E50914]"
            />
          </div>

          <div className="flex items-center justify-between py-4 border-t border-gray-800">
            <div>
              <span className="text-white font-medium">流式搜索</span>
              <p className="text-gray-500 text-sm">实时显示搜索结果</p>
            </div>
            <Switch checked={form.FluidSearch} onChange={(v) => setForm({ ...form, FluidSearch: v })} />
          </div>

          <div className="flex items-center justify-between py-4 border-t border-gray-800">
            <div>
              <span className="text-white font-medium">禁用黄暴过滤</span>
              <p className="text-gray-500 text-sm">关闭敏感内容过滤</p>
            </div>
            <Switch checked={form.DisableYellowFilter} onChange={(v) => setForm({ ...form, DisableYellowFilter: v })} />
          </div>

          <div className="pt-4">
            <Button onClick={() => onSave(form)}>
              <Save className="w-4 h-4 mr-2" />
              保存设置
            </Button>
          </div>
        </div>
      </Card>
    </div>
  );
}
