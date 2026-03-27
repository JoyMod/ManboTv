'use client';

import { AlertCircle } from 'lucide-react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { Suspense, useState } from 'react';

import Logo from '@/components/ui/Logo';

function LoginPageClient() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [password, setPassword] = useState('');
  const [username, setUsername] = useState('admin');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError(null);

    if (!password || !username) return;

    try {
      setLoading(true);
      const res = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          username,
          password,
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
    <div className='min-h-screen bg-netflix-black flex flex-col'>
      {/* 顶部 Logo */}
      <div className='absolute top-0 left-0 right-0 p-6 z-10'>
        <div className='max-w-[1920px] mx-auto'>
          <Logo size='lg' />
        </div>
      </div>

      {/* 背景渐变 */}
      <div className='absolute inset-0 overflow-hidden'>
        <div className='absolute inset-0 bg-gradient-to-b from-black/60 via-netflix-black/80 to-netflix-black' />
        <div className='absolute top-0 left-0 right-0 h-[70vh] bg-gradient-to-b from-black/80 to-transparent' />
      </div>

      {/* 登录框 */}
      <div className='relative flex-1 flex items-center justify-center px-4 py-20'>
        <div className='w-full max-w-md bg-black/70 backdrop-blur-sm rounded-xl p-8 md:p-12 border border-netflix-gray-800'>
          <h1 className='text-3xl font-bold text-white mb-8'>登录</h1>

          <form onSubmit={handleSubmit} className='space-y-6'>
            <div>
              <input
                id='username'
                type='text'
                autoComplete='username'
                className='w-full px-4 py-4 bg-netflix-gray-800 border border-netflix-gray-700 rounded-lg text-white placeholder-netflix-gray-500 focus:outline-none focus:border-netflix-red transition-colors'
                placeholder='用户名'
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
            </div>

            <div>
              <input
                id='password'
                type='password'
                autoComplete='current-password'
                className='w-full px-4 py-4 bg-netflix-gray-800 border border-netflix-gray-700 rounded-lg text-white placeholder-netflix-gray-500 focus:outline-none focus:border-netflix-red transition-colors'
                placeholder='密码'
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>

            {error && (
              <div className='flex items-center gap-2 text-netflix-red text-sm'>
                <AlertCircle className='w-4 h-4' />
                {error}
              </div>
            )}

            {/* 登录按钮 */}
            <button
              type='submit'
              disabled={!password || !username || loading}
              className='w-full py-4 bg-netflix-red text-white font-bold rounded-lg hover:bg-netflix-red-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed'
            >
              {loading ? '登录中...' : '登录'}
            </button>

            <div className='flex items-center justify-between text-sm text-netflix-gray-400'>
              <label className='flex items-center gap-2 cursor-pointer hover:text-white transition-colors'>
                <input
                  type='checkbox'
                  className='rounded bg-netflix-gray-800 border-netflix-gray-700'
                />
                记住我
              </label>
              <Link href='/help' className='hover:underline'>
                需要帮助?
              </Link>
            </div>
          </form>

          <div className='mt-8 text-netflix-gray-400 text-sm'>
            <p>
              首次使用?{' '}
              <Link href='/register' className='text-white hover:underline'>
                免费注册
              </Link>
            </p>
            <p className='mt-4 text-xs'>
              登录即表示您同意我们的{' '}
              <Link href='/terms' className='text-blue-400 hover:underline'>
                使用条款
              </Link>{' '}
              和{' '}
              <Link href='/privacy' className='text-blue-400 hover:underline'>
                隐私声明
              </Link>
              。
            </p>
          </div>
        </div>
      </div>

      {/* 页脚 */}
      <footer className='relative py-8 px-4 bg-black/50'>
        <div className='max-w-[1920px] mx-auto'>
          <p className='text-netflix-gray-500 text-sm mb-4'>
            有问题? 请联系我们
          </p>
          <div className='grid grid-cols-2 md:grid-cols-4 gap-4 text-xs text-netflix-gray-500'>
            <Link href='/help' className='hover:underline'>
              常见问题
            </Link>
            <Link href='/help' className='hover:underline'>
              帮助中心
            </Link>
            <Link href='/terms' className='hover:underline'>
              使用条款
            </Link>
            <Link href='/privacy' className='hover:underline'>
              隐私政策
            </Link>
          </div>
          <div className='mt-6 rounded-xl border border-netflix-gray-800 bg-black/40 p-4 text-xs leading-6 text-netflix-gray-500'>
            <p>
              本站仅提供程序功能与聚合入口，不存储、不制作、不上传影视内容。
            </p>
            <p className='mt-2'>
              相关搜索结果、封面、详情与播放链接可能来自用户配置或第三方公开接口；如权利人认为存在侵权，请联系管理员处理，核实后将尽快删除、屏蔽或断开相关链接。
            </p>
            <p className='mt-2'>
              请在下载、缓存或访问相关内容后 24 小时内自行删除，并确保使用行为符合所在地法律法规。
            </p>
            <p className='mt-2'>
              严禁在中国大陆法律管辖范围内传播本项目或将其发布至抖音、哔哩哔哩、小红书、微博、视频号等中国大陆平台。
            </p>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <div className='min-h-screen bg-[#141414] flex items-center justify-center text-white'>
          加载中...
        </div>
      }
    >
      <LoginPageClient />
    </Suspense>
  );
}
