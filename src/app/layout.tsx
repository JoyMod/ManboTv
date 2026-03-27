/* eslint-disable @typescript-eslint/no-explicit-any */

import type { Metadata, Viewport } from 'next';

import './globals.css';

import MobileBottomNav from '@/components/MobileBottomNav';
import RouteFeedback from '@/components/RouteFeedback';

export const dynamic = 'force-dynamic';

// 动态生成 metadata，支持配置更新后的标题变化
export async function generateMetadata(): Promise<Metadata> {
  const siteName = process.env.NEXT_PUBLIC_SITE_NAME || 'ManboTV';

  return {
    title: {
      default: siteName,
      template: `%s | ${siteName}`,
    },
    description: '全网影视聚合平台 - 海量高清影视资源在线观看',
    manifest: '/manifest.json',
    icons: {
      icon: '/favicon.svg',
      shortcut: '/favicon.svg',
      apple: '/favicon.svg',
    },
    appleWebApp: {
      capable: true,
      statusBarStyle: 'black',
      title: siteName,
    },
  };
}

export const viewport: Viewport = {
  viewportFit: 'cover',
  themeColor: '#141414',
};

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang='zh-CN'>
      <head>
        <meta
          name='viewport'
          content='width=device-width, initial-scale=1.0, viewport-fit=cover'
        />
        <link rel='icon' type='image/svg+xml' href='/favicon.svg' />
        <link rel='alternate icon' href='/favicon.ico' />
        <link rel='apple-touch-icon' href='/favicon.svg' />
      </head>
      <body className='min-h-screen bg-netflix-black pb-20 text-white md:pb-0'>
        <RouteFeedback />
        {children}
        <MobileBottomNav />
      </body>
    </html>
  );
}
