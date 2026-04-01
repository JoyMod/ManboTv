'use client';

import { Clapperboard, Flame, Heart, Home, Search } from 'lucide-react';
import { usePathname } from 'next/navigation';
import React from 'react';

import { useFastNavigation } from '@/lib/navigation-feedback';

const navItems = [
  { path: '/', label: '首页', icon: Home },
  { path: '/movie', label: '片库', icon: Clapperboard },
  { path: '/search', label: '搜索', icon: Search },
  { path: '/hot', label: '热播', icon: Flame },
  { path: '/favorites', label: '收藏', icon: Heart },
];

const HiddenPathPrefixes = [
  '/help',
  '/login',
  '/privacy',
  '/register',
  '/terms',
  '/warning',
];

export default function MobileBottomNav() {
  const pathname = usePathname();
  const { navigate, prefetchHref } = useFastNavigation();

  // 只在移动端显示
  const [isMobile, setIsMobile] = React.useState(false);

  React.useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768);
    };

    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  React.useEffect(() => {
    const shouldHideOnPath = HiddenPathPrefixes.some((prefix) =>
      pathname.startsWith(prefix)
    );
    const shouldEnablePrefetch =
      isMobile && !pathname.startsWith('/play') && !shouldHideOnPath;
    if (!shouldEnablePrefetch) {
      return;
    }
    navItems.forEach((item) => prefetchHref(item.path));
  }, [isMobile, pathname, prefetchHref]);

  if (!isMobile) return null;

  const shouldHideOnPath = HiddenPathPrefixes.some((prefix) =>
    pathname.startsWith(prefix)
  );
  if (pathname.startsWith('/play') || shouldHideOnPath) return null;

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 border-t border-zinc-800 bg-zinc-950/95 shadow-[0_-10px_30px_rgba(0,0,0,0.35)] backdrop-blur-md md:hidden">
      <div className="grid grid-cols-5 gap-1 px-2 py-2">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive =
            item.path === '/'
              ? pathname === item.path
              : pathname === item.path || pathname.startsWith(`${item.path}/`);

          return (
            <button
              key={item.path}
              onClick={() => navigate(item.path)}
              onPointerEnter={() => prefetchHref(item.path)}
              className={`flex flex-col items-center gap-1 rounded-2xl px-2 py-2 transition-all ${
                isActive
                  ? 'bg-white text-black shadow-[0_10px_20px_rgba(255,255,255,0.12)]'
                  : 'text-zinc-400'
              }`}
            >
              <Icon className={`h-5 w-5 ${isActive ? 'text-black' : ''}`} />
              <span className="text-[10px] font-medium">{item.label}</span>
            </button>
          );
        })}
      </div>
    </nav>
  );
}
