'use client';

import { useRouter } from 'next/navigation';
import { startTransition, useCallback } from 'react';

const RouteLoadingAttribute = 'data-route-loading';

export function beginRouteChange() {
  if (typeof document === 'undefined') return;
  document.documentElement.setAttribute(RouteLoadingAttribute, 'true');
}

export function endRouteChange() {
  if (typeof document === 'undefined') return;
  document.documentElement.removeAttribute(RouteLoadingAttribute);
}

export function useFastNavigation() {
  const router = useRouter();

  const prefetchHref = useCallback(
    (href?: string) => {
      if (!href) return;
      router.prefetch(href);
    },
    [router]
  );

  const navigate = useCallback(
    (href: string, options?: { replace?: boolean }) => {
      beginRouteChange();
      startTransition(() => {
        if (options?.replace) {
          router.replace(href);
          return;
        }
        router.push(href);
      });
    },
    [router]
  );

  return { navigate, prefetchHref };
}
