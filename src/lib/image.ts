export function normalizeImageUrl(url?: string | null): string {
  if (!url) return '';
  const value = url.trim();
  if (!value) return '';
  if (value.startsWith('data:image')) return value;
  if (value.startsWith('//')) return `https:${value}`;
  if (value.startsWith('http://') || value.startsWith('https://')) return value;
  if (value.startsWith('/')) return value;
  return '';
}

export function toProxyImageSrc(
  url?: string | null,
  fallback = '/placeholder-poster.svg'
): string {
  const normalized = normalizeImageUrl(url);
  if (!normalized) return fallback;
  if (!normalized.startsWith('http://') && !normalized.startsWith('https://'))
    return normalized;
  return `/api/image?url=${encodeURIComponent(normalized)}`;
}

export function toLogoProxyImageSrc(
  url?: string | null,
  fallback = '/placeholder-poster.svg'
): string {
  const normalized = normalizeImageUrl(url);
  if (!normalized) return fallback;
  if (!normalized.startsWith('http://') && !normalized.startsWith('https://'))
    return normalized;
  return `/api/proxy/logo?url=${encodeURIComponent(normalized)}`;
}

export function toImageSrc(
  url?: string | null,
  fallback = '/placeholder-poster.svg'
): string {
  const normalized = normalizeImageUrl(url);
  return normalized || fallback;
}
