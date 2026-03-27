'use client';

import { usePathname, useSearchParams } from 'next/navigation';
import React, { useEffect } from 'react';

import { endRouteChange } from '@/lib/navigation-feedback';

export default function RouteFeedback() {
  const pathname = usePathname();
  const searchParams = useSearchParams();

  useEffect(() => {
    const frameId = window.requestAnimationFrame(() => {
      endRouteChange();
    });

    return () => window.cancelAnimationFrame(frameId);
  }, [pathname, searchParams]);

  return <div className='route-feedback-bar' aria-hidden='true' />;
}
