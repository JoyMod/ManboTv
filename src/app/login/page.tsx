/* eslint-disable @typescript-eslint/no-explicit-any */

'use client';

import { AlertCircle, CheckCircle } from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Suspense, useEffect, useState } from 'react';

import { CURRENT_VERSION } from '@/lib/version';
import { checkForUpdates, UpdateStatus } from '@/lib/version_check';

import { useSite } from '@/components/SiteProvider';
import Logo from '@/components/ui/Logo';

// 版本显示组件
function VersionDisplay() {
  const [updateStatus, setUpdateStatus] = useState<UpdateStatus | null>(null);
  const [isChecking, setIsChecking] = useState(true);

  useEffect(() => {
    const checkUpdate = async () => {
      try {
        const status = await checkForUpdates();
        setUpdateStatus(status);
      } catch (_) {
        // do nothing
      } finally {
        setIsChecking(false);
      }
    };

    checkUpdate();
  }, []);

  return (
    <button
      onClick={() =>
        window.open('https://github.com/MoonTechLab/LunaTV', '_blank')
      }
      className='fixed bottom-4 left-1/2 transform -translate-x-1/2 flex items-center gap-2 text-xs text-gray-500 transition-colors cursor-pointer hover:text-gray-300'
    >
      <span className='font-mono'>v{CURRENT_VERSION}</span>
      {!isChecking && updateStatus !== UpdateStatus.FETCH_FAILED && (
        <div
          className={`flex items-center gap-1.5 ${updateStatus === UpdateStatus.HAS_UPDATE
            ? 'text-yellow-400'
            : updateStatus === UpdateStatus.NO_UPDATE
              ? 'text-green-400'
              : ''
            }`}
        >
          {updateStatus === UpdateStatus.HAS_UPDATE && (
            <>
              <AlertCircle className='w-3.5 h-3.5' />
              <span className='font-semibold text-xs'>有新版本</span>
            </>
          )}
          {updateStatus === UpdateStatus.NO_UPDATE && (
            <>
              <CheckCircle className='w-3.5 h-3.5' />
              <span className='font-semibold text-xs'>已是最新</span>
            </>
          )}
        </div>
      )}
    </button>
  );
}

function LoginPageClient() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [password, setPassword] = useState('');
  const [username, setUsername] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [shouldAskUsername, setShouldAskUsername] = useState(false);

  // 在客户端挂载后设置配置
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const storageType = (window as any).RUNTIME_CONFIG?.STORAGE_TYPE;
      setShouldAskUsername(storageType && storageType !== 'localstorage');
    }
  }, []);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError(null);

    if (!password || (shouldAskUsername && !username)) return;

    try {
      setLoading(true);
      const res = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          password,
          ...(shouldAskUsername ? { username } : {}),
        }),
      });

      if (res.ok) {
        const redirect = searchParams.get('redirect') || '/';
        router.replace(redirect);
      } else if (res.status === 401) {
        setError('密码错误');
      } else {
        const data = await res.json().catch(() => ({}));
        setError(data.error ?? '服务器错误');
      }
    } catch (error) {
      setError('网络错误，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#141414] flex flex-col">
      {/* 顶部 Logo */}
      <div className="absolute top-0 left-0 right-0 p-6">
        <div className="max-w-[1920px] mx-auto">
          <Logo size="lg" />
        </div>
      </div>

      {/* 背景渐变 */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-b from-black/60 via-[#141414]/80 to-[#141414]" />
        <div className="absolute top-0 left-0 right-0 h-[70vh] bg-gradient-to-b from-black/80 to-transparent" />
      </div>

      {/* 登录框 */}
      <div className="relative flex-1 flex items-center justify-center px-4 py-20">
        <div className="w-full max-w-md bg-black/70 backdrop-blur-sm rounded-lg p-8 md:p-12">
          <h1 className="text-3xl font-bold text-white mb-8">登录</h1>
          
          <form onSubmit={handleSubmit} className="space-y-6">
            {shouldAskUsername && (
              <div>
                <input
                  id="username"
                  type="text"
                  autoComplete="username"
                  className="w-full px-4 py-4 bg-[#333] border border-[#333] rounded text-white placeholder-gray-500 focus:outline-none focus:border-[#E50914] transition-colors"
                  placeholder="用户名"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                />
              </div>
            )}

            <div>
              <input
                id="password"
                type="password"
                autoComplete="current-password"
                className="w-full px-4 py-4 bg-[#333] border border-[#333] rounded text-white placeholder-gray-500 focus:outline-none focus:border-[#E50914] transition-colors"
                placeholder="密码"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>

            {error && (
              <div className="flex items-center gap-2 text-[#E50914] text-sm">
                <AlertCircle className="w-4 h-4" />
                {error}
              </div>
            )}

            {/* 登录按钮 */}
            <button
              type="submit"
              disabled={!password || loading || (shouldAskUsername && !username)}
              className="w-full py-4 bg-[#E50914] text-white font-bold rounded hover:bg-[#f40612] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? '登录中...' : '登录'}
            </button>

            <div className="flex items-center justify-between text-sm text-gray-400">
              <label className="flex items-center gap-2 cursor-pointer hover:text-gray-300">
                <input type="checkbox" className="rounded bg-[#333] border-[#333]" />
                记住我
              </label>
              <a href="#" className="hover:underline">需要帮助?</a>
            </div>
          </form>

          <div className="mt-8 text-gray-400 text-sm">
            <p>
              首次使用? <a href="#" className="text-white hover:underline">免费注册</a>
            </p>
            <p className="mt-4 text-xs">
              登录即表示您同意我们的 <a href="#" className="text-blue-400 hover:underline">使用条款</a> 和 <a href="#" className="text-blue-400 hover:underline">隐私声明</a>。
            </p>
          </div>
        </div>
      </div>

      {/* 页脚 */}
      <footer className="relative py-8 px-4 bg-black/50">
        <div className="max-w-[1920px] mx-auto">
          <p className="text-gray-500 text-sm mb-4">有问题? 请联系我们</p>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-xs text-gray-500">
            <a href="#" className="hover:underline">常见问题</a>
            <a href="#" className="hover:underline">帮助中心</a>
            <a href="#" className="hover:underline">使用条款</a>
            <a href="#" className="hover:underline">隐私政策</a>
          </div>
        </div>
      </footer>

      {/* 版本信息显示 */}
      <VersionDisplay />
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={<div className="min-h-screen bg-[#141414] flex items-center justify-center text-white">加载中...</div>}>
      <LoginPageClient />
    </Suspense>
  );
}
