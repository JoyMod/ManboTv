'use client';

import {
  ChevronDown,
  Film,
  Heart,
  History,
  LogOut,
  Menu,
  Search,
  Settings,
  Sparkles,
  Tv,
  User,
  X,
} from 'lucide-react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import React, { useEffect, useMemo, useRef, useState } from 'react';

import { beginRouteChange, useFastNavigation } from '@/lib/navigation-feedback';

import ContentModeSelector from '@/components/ContentModeSelector';
import Logo from '@/components/ui/Logo';

interface CategorySectionItem {
  label: string;
  value: string;
}

interface CategorySection {
  title: string;
  query: string;
  items: CategorySectionItem[];
}

const navItems = [
  { label: '首页', href: '/' },
  { label: '电影', href: '/movie', key: 'movie' },
  { label: '电视剧', href: '/tv', key: 'tv' },
  { label: '综艺', href: '/variety', key: 'variety' },
  { label: '动漫', href: '/anime', key: 'anime' },
  { label: '热播', href: '/hot' },
  { label: '我的片单', href: '/favorites' },
];

const categoryData: Record<
  string,
  {
    label: string;
    icon: React.ReactNode;
    href: string;
    sections: CategorySection[];
  }
> = {
  movie: {
    label: '电影',
    icon: <Film className='h-4 w-4' />,
    href: '/movie',
    sections: [
      {
        title: '类型',
        query: 'type',
        items: ['动作', '喜剧', '爱情', '科幻', '悬疑', '恐怖', '犯罪', '动画', '剧情', '奇幻', '冒险', '传记'].map(
          (item) => ({ label: item, value: item })
        ),
      },
      {
        title: '地区',
        query: 'region',
        items: ['华语', '美国', '韩国', '日本', '印度', '泰国', '法国', '英国'].map((item) => ({
          label: item,
          value: item,
        })),
      },
      {
        title: '特色',
        query: 'feature',
        items: ['豆瓣高分', '获奖佳作', '新片热映', '经典重温', 'IMAX', '4K'].map((item) => ({
          label: item,
          value: item,
        })),
      },
      {
        title: '年代',
        query: 'year',
        items: ['2026', '2025', '2024', '2023', '2022', '2021', '经典'].map((item) => ({
          label: item,
          value: item,
        })),
      },
    ],
  },
  tv: {
    label: '电视剧',
    icon: <Tv className='h-4 w-4' />,
    href: '/tv',
    sections: [
      {
        title: '剧集类型',
        query: 'type',
        items: ['古装', '都市', '悬疑', '爱情', '武侠', '奇幻', '谍战', '家庭'].map((item) => ({
          label: item,
          value: item,
        })),
      },
      {
        title: '地区频道',
        query: 'region',
        items: ['国产剧', '美剧', '韩剧', '日剧', '港剧', '台剧', '泰剧', '英剧'].map((item) => ({
          label: item,
          value: item,
        })),
      },
      {
        title: '更新状态',
        query: 'status',
        items: ['连载中', '已完结', '即将开播'].map((item) => ({
          label: item,
          value: item,
        })),
      },
      {
        title: '年代',
        query: 'year',
        items: ['2026', '2025', '2024', '2023', '2022', '2021'].map((item) => ({
          label: item,
          value: item,
        })),
      },
    ],
  },
  variety: {
    label: '综艺',
    icon: <Sparkles className='h-4 w-4' />,
    href: '/variety',
    sections: [
      {
        title: '综艺类型',
        query: 'type',
        items: ['真人秀', '脱口秀', '音乐', '选秀', '竞技', '旅行', '美食', '访谈'].map((item) => ({
          label: item,
          value: item,
        })),
      },
      {
        title: '地区频道',
        query: 'region',
        items: ['国内', '韩国', '日本', '港台', '欧美'].map((item) => ({
          label: item,
          value: item,
        })),
      },
      {
        title: '更新状态',
        query: 'status',
        items: ['连载中', '已完结', '即将开播'].map((item) => ({
          label: item,
          value: item,
        })),
      },
      {
        title: '年代',
        query: 'year',
        items: ['2026', '2025', '2024', '2023', '2022', '2021'].map((item) => ({
          label: item,
          value: item,
        })),
      },
    ],
  },
  anime: {
    label: '动漫',
    icon: <Sparkles className='h-4 w-4' />,
    href: '/anime',
    sections: [
      {
        title: '动漫类型',
        query: 'type',
        items: ['热血', '恋爱', '搞笑', '悬疑', '科幻', '冒险', '战斗', '治愈'].map((item) => ({
          label: item,
          value: item,
        })),
      },
      {
        title: '地区频道',
        query: 'region',
        items: [
          { label: '日本动画', value: '日本' },
          { label: '国漫', value: '国产' },
          { label: '欧美动画', value: '欧美' },
        ],
      },
      {
        title: '更新状态',
        query: 'status',
        items: ['连载中', '已完结', '新番', '剧场版', 'OVA'].map((item) => ({
          label: item,
          value: item,
        })),
      },
      {
        title: '年份',
        query: 'year',
        items: ['2024', '2023', '2022', '2021', '2020', '经典'].map((item) => ({
          label: item,
          value: item,
        })),
      },
    ],
  },
};

const buildCategoryHref = (baseHref: string, query: string, value: string) => {
  const params = new URLSearchParams({ [query]: value });
  return `${baseHref}?${params.toString()}`;
};

export function TopNav() {
  const pathname = usePathname();
  const { navigate, prefetchHref } = useFastNavigation();
  const userMenuRef = useRef<HTMLDivElement | null>(null);
  const [scrolled, setScrolled] = useState(false);
  const [search, setSearch] = useState('');
  const [mobileOpen, setMobileOpen] = useState(false);
  const [userOpen, setUserOpen] = useState(false);
  const [hoverKey, setHoverKey] = useState<string | null>(null);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 24);
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  const pathnameCategoryKey = useMemo(() => {
    const matchedItem = navItems.find(
      (item) =>
        item.key &&
        (item.href === '/'
          ? pathname === '/'
          : pathname === item.href || pathname.startsWith(`${item.href}/`))
    );
    return matchedItem?.key || null;
  }, [pathname]);

  const activeCategoryKey = hoverKey || pathnameCategoryKey;

  const activeCategory = useMemo(() => {
    if (!activeCategoryKey) return null;
    return categoryData[activeCategoryKey] || null;
  }, [activeCategoryKey]);

  useEffect(() => {
    navItems.forEach((item) => prefetchHref(item.href));
  }, [prefetchHref]);

  useEffect(() => {
    setUserOpen(false);
    setMobileOpen(false);
    setHoverKey(null);
  }, [pathname]);

  useEffect(() => {
    const handlePointerDown = (event: MouseEvent) => {
      const target = event.target;
      if (!(target instanceof Node)) {
        return;
      }

      if (userOpen && userMenuRef.current && !userMenuRef.current.contains(target)) {
        setUserOpen(false);
      }
    };

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key !== 'Escape') {
        return;
      }
      setUserOpen(false);
      setMobileOpen(false);
      setHoverKey(null);
    };

    document.addEventListener('mousedown', handlePointerDown);
    window.addEventListener('keydown', handleEscape);
    return () => {
      document.removeEventListener('mousedown', handlePointerDown);
      window.removeEventListener('keydown', handleEscape);
    };
  }, [userOpen]);

  const logout = async () => {
    await fetch('/api/logout', { method: 'POST' });
    setUserOpen(false);
    setMobileOpen(false);
    navigate('/login', { replace: true });
  };

  const searchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!search.trim()) return;
    setUserOpen(false);
    setMobileOpen(false);
    navigate(`/search?q=${encodeURIComponent(search.trim())}`);
    setSearch('');
  };

  return (
    <>
      <header
        className={`fixed inset-x-0 top-0 z-50 transition-all ${
          scrolled
            ? 'bg-[#141414]/95 shadow-lg backdrop-blur'
            : 'bg-gradient-to-b from-black/85 to-transparent'
        }`}
      >
        <div className='mx-auto flex h-16 max-w-[1920px] items-center justify-between px-4 sm:px-8'>
          <div className='flex items-center gap-6'>
            <Link
              href='/'
              className='shrink-0'
              onClick={() => beginRouteChange()}
              onMouseEnter={() => prefetchHref('/')}
            >
              <Logo size='md' />
            </Link>

            <nav className='hidden items-center gap-4 lg:flex'>
              {navItems.map((item) => {
                const active =
                  item.href === '/'
                    ? pathname === '/'
                    : pathname.startsWith(item.href);
                const menuActive = item.key && activeCategoryKey === item.key;
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    onMouseEnter={() => setHoverKey(item.key || null)}
                    onFocus={() => setHoverKey(item.key || null)}
                    onClick={() => beginRouteChange()}
                    className={`rounded-full px-3 py-2 text-sm transition-all ${
                      menuActive
                        ? 'bg-white/10 text-white'
                        : 'bg-transparent'
                    } ${
                      active
                        ? 'font-semibold text-white'
                        : 'text-zinc-300 hover:text-white'
                    }`}
                  >
                    <span className='inline-flex items-center gap-1.5'>
                      {item.label}
                      {item.key ? <ChevronDown className='h-3.5 w-3.5' /> : null}
                    </span>
                  </Link>
                );
              })}
            </nav>
          </div>

          <div className='flex items-center gap-3'>
            <form
              onSubmit={searchSubmit}
              className='hidden items-center rounded bg-black/55 px-3 py-1.5 sm:flex'
            >
              <Search className='h-4 w-4 text-zinc-300' />
              <input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder='搜索影视'
                className='ml-2 w-32 bg-transparent text-sm text-white outline-none placeholder:text-zinc-400 md:w-52'
              />
            </form>

            <div ref={userMenuRef} className='relative hidden sm:block'>
              <button
                onClick={() => setUserOpen((v) => !v)}
                aria-expanded={userOpen}
                className='flex h-8 w-8 items-center justify-center rounded bg-red-700'
              >
                <User className='h-4 w-4 text-white' />
              </button>
              {userOpen ? (
                <div className='absolute right-0 mt-2 w-72 rounded border border-zinc-800 bg-[#1a1a1a] p-2'>
                  <div className='mb-2'>
                    <ContentModeSelector compact />
                  </div>
                  <Link
                    href='/favorites'
                    onClick={() => {
                      setUserOpen(false);
                      beginRouteChange();
                    }}
                    onMouseEnter={() => prefetchHref('/favorites')}
                    className='flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-200 hover:bg-zinc-800'
                  >
                    <Heart className='h-4 w-4' /> 我的片单
                  </Link>
                  <Link
                    href='/favorites?tab=history'
                    onClick={() => {
                      setUserOpen(false);
                      beginRouteChange();
                    }}
                    onMouseEnter={() => prefetchHref('/favorites?tab=history')}
                    className='flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-200 hover:bg-zinc-800'
                  >
                    <History className='h-4 w-4' /> 观看历史
                  </Link>
                  <Link
                    href='/admin'
                    onClick={() => {
                      setUserOpen(false);
                      beginRouteChange();
                    }}
                    onMouseEnter={() => prefetchHref('/admin')}
                    className='flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-200 hover:bg-zinc-800'
                  >
                    <Settings className='h-4 w-4' /> 管理后台
                  </Link>
                  <button
                    onClick={logout}
                    className='mt-1 flex w-full items-center gap-2 rounded px-3 py-2 text-left text-sm text-zinc-200 hover:bg-zinc-800'
                  >
                    <LogOut className='h-4 w-4' /> 退出登录
                  </button>
                </div>
              ) : null}
            </div>

            <button
              onClick={() => setMobileOpen(true)}
              className='rounded p-2 text-white lg:hidden'
            >
              <Menu className='h-5 w-5' />
            </button>
          </div>
        </div>
      </header>

      {activeCategory ? (
        <div
          onMouseLeave={() => setHoverKey(null)}
          className='hidden border-b border-white/10 bg-[linear-gradient(180deg,rgba(16,16,16,0.98)_0%,rgba(12,12,12,0.94)_100%)] lg:block'
        >
          <div className='mx-auto flex max-w-[1920px] flex-wrap items-start gap-6 px-8 pb-4 pt-20'>
            <div className='min-w-[160px] text-white'>
              <div className='mb-2 flex items-center gap-3'>
                <div className='flex h-10 w-10 items-center justify-center rounded-2xl bg-netflix-red/15 text-netflix-red'>
                  {activeCategory.icon}
                </div>
                <div className='text-lg font-semibold tracking-wide'>{activeCategory.label}</div>
              </div>
              <div className='text-sm text-zinc-400'>当前频道的快捷筛选入口</div>
            </div>

            <div className='grid min-w-0 flex-1 gap-4 xl:grid-cols-4'>
              {activeCategory.sections.map((section) => (
                <div key={section.title} className='min-w-0'>
                  <h4 className='mb-2 text-xs font-medium tracking-[0.24em] text-zinc-500'>
                    {section.title}
                  </h4>
                  <div className='flex flex-wrap gap-2'>
                    {section.items.map((item) => (
                      <Link
                        key={`${section.query}-${item.value}`}
                        href={buildCategoryHref(activeCategory.href, section.query, item.value)}
                        onClick={() => beginRouteChange()}
                        onMouseEnter={() =>
                          prefetchHref(buildCategoryHref(activeCategory.href, section.query, item.value))
                        }
                        className='rounded-full border border-white/10 bg-white/[0.04] px-3 py-1.5 text-sm font-medium text-zinc-200 transition-all hover:border-netflix-red/40 hover:bg-netflix-red/12 hover:text-white'
                      >
                        {item.label}
                      </Link>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      ) : null}

      {mobileOpen ? (
        <div
          className='fixed inset-0 z-[60] bg-black/80 lg:hidden'
          onClick={() => setMobileOpen(false)}
        >
          <div
            className='absolute right-0 top-0 h-full w-[82%] max-w-sm bg-[#151515] p-4'
            onClick={(event) => event.stopPropagation()}
          >
            <div className='mb-4 flex items-center justify-between'>
              <Logo size='sm' />
              <button onClick={() => setMobileOpen(false)}>
                <X className='h-5 w-5 text-white' />
              </button>
            </div>

            <form
              onSubmit={searchSubmit}
              className='mb-4 flex items-center rounded bg-zinc-900 px-3 py-2'
            >
              <Search className='h-4 w-4 text-zinc-300' />
              <input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder='搜索影视'
                className='ml-2 w-full bg-transparent text-sm text-white outline-none placeholder:text-zinc-500'
              />
            </form>

            <div className='space-y-1'>
              <ContentModeSelector />
              {navItems.map((item) => (
                <Link
                  key={item.href}
                  href={item.href}
                  onClick={() => {
                    beginRouteChange();
                    setMobileOpen(false);
                  }}
                  onMouseEnter={() => prefetchHref(item.href)}
                  className='block rounded px-3 py-2 text-sm text-zinc-200 hover:bg-zinc-800'
                >
                  {item.label}
                </Link>
              ))}
            </div>

            <div className='mt-4 border-t border-white/10 pt-4'>
              <p className='mb-2 text-xs tracking-[0.2em] text-zinc-500'>账户</p>
              <div className='space-y-1'>
                <Link
                  href='/favorites'
                  onClick={() => {
                    beginRouteChange();
                    setMobileOpen(false);
                  }}
                  onMouseEnter={() => prefetchHref('/favorites')}
                  className='flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-200 hover:bg-zinc-800'
                >
                  <Heart className='h-4 w-4' /> 我的片单
                </Link>
                <Link
                  href='/favorites?tab=history'
                  onClick={() => {
                    beginRouteChange();
                    setMobileOpen(false);
                  }}
                  onMouseEnter={() => prefetchHref('/favorites?tab=history')}
                  className='flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-200 hover:bg-zinc-800'
                >
                  <History className='h-4 w-4' /> 观看历史
                </Link>
                <Link
                  href='/admin'
                  onClick={() => {
                    beginRouteChange();
                    setMobileOpen(false);
                  }}
                  onMouseEnter={() => prefetchHref('/admin')}
                  className='flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-200 hover:bg-zinc-800'
                >
                  <Settings className='h-4 w-4' /> 管理后台
                </Link>
                <button
                  type='button'
                  onClick={() => {
                    void logout();
                  }}
                  className='flex w-full items-center gap-2 rounded px-3 py-2 text-left text-sm text-zinc-200 hover:bg-zinc-800'
                >
                  <LogOut className='h-4 w-4' /> 退出登录
                </button>
              </div>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}

export default TopNav;
