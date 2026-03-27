export type ContentMode = 'safe' | 'mixed' | 'adult_only';
export type ContentModePreference = ContentMode | 'inherit';

type ContentModeOption = {
  label: string;
  value: ContentModePreference;
  description: string;
};

const CONTENT_MODE_COOKIE_NAME = 'content_access_mode';
const CONTENT_MODE_COOKIE_MAX_AGE_SECONDS = 180 * 24 * 60 * 60;

export const contentModeOptions: ContentModeOption[] = [
  {
    label: '跟随站点',
    value: 'inherit',
    description: '使用管理员设置的默认内容模式。',
  },
  {
    label: '儿童模式',
    value: 'safe',
    description: '仅显示普通内容，适合未成年人。',
  },
  {
    label: '标准模式',
    value: 'mixed',
    description: '普通内容与成人内容都显示。',
  },
  {
    label: '成人模式',
    value: 'adult_only',
    description: '仅显示成人内容，普通影视不再展示。',
  },
];

export function getContentModePreference(): ContentModePreference {
  if (typeof document === 'undefined') {
    return 'inherit';
  }

  const cookieItems = document.cookie.split(';');
  for (const cookieItem of cookieItems) {
    const [rawKey, ...rawValueParts] = cookieItem.trim().split('=');
    if (rawKey !== CONTENT_MODE_COOKIE_NAME) {
      continue;
    }

    const rawValue = decodeURIComponent(rawValueParts.join('='));
    if (
      rawValue === 'safe' ||
      rawValue === 'mixed' ||
      rawValue === 'adult_only'
    ) {
      return rawValue;
    }
  }

  return 'inherit';
}

export function setContentModePreference(mode: ContentModePreference) {
  if (typeof document === 'undefined') {
    return;
  }

  if (mode === 'inherit') {
    document.cookie = `${CONTENT_MODE_COOKIE_NAME}=; path=/; max-age=0; SameSite=Lax`;
    return;
  }

  document.cookie = `${CONTENT_MODE_COOKIE_NAME}=${encodeURIComponent(
    mode
  )}; path=/; max-age=${CONTENT_MODE_COOKIE_MAX_AGE_SECONDS}; SameSite=Lax`;
}
