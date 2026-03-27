'use client';

import React from 'react';

interface LogoProps {
  className?: string;
  size?: 'sm' | 'md' | 'lg';
}

export const Logo: React.FC<LogoProps> = ({ className = '', size = 'md' }) => {
  const sizes = {
    sm: { width: 80, height: 28 },
    md: { width: 120, height: 40 },
    lg: { width: 160, height: 54 },
  };

  const { width, height } = sizes[size];

  return (
    <svg
      viewBox='0 0 120 40'
      width={width}
      height={height}
      className={className}
      xmlns='http://www.w3.org/2000/svg'
    >
      {/* 播放按钮融合 M 字母 */}
      <g transform='translate(0, 5)'>
        {/* M 的左半部分 */}
        <path
          d='M5 30 L15 5 L25 30'
          stroke='#E50914'
          strokeWidth='5'
          fill='none'
          strokeLinecap='round'
          strokeLinejoin='round'
        />
        {/* 播放三角形 */}
        <polygon points='20,12 32,20 20,28' fill='#E50914' />
      </g>

      {/* 文字部分 */}
      <text
        x='38'
        y='28'
        fontSize='16'
        fill='white'
        fontWeight='800'
        fontFamily="'Noto Sans SC', system-ui, -apple-system, sans-serif"
        letterSpacing='0.5'
      >
        ManboTV
      </text>

      {/* 下划线装饰 */}
      <rect x='38' y='32' width='74' height='3' rx='1.5' fill='#E50914' />
    </svg>
  );
};

export default Logo;
