import { Metadata } from 'next';

import StaticInfoPage from '@/components/info/StaticInfoPage';

export const metadata: Metadata = {
  title: '隐私政策 - ManboTV',
  description: 'ManboTV 隐私政策',
};

export default function PrivacyPage() {
  return (
    <StaticInfoPage
      title='隐私政策'
      description='这里说明站点会记录哪些基础数据，以及这些数据如何用于收藏、播放记录和个性化功能。'
      badge='Privacy'
      actions={[
        { label: '返回登录', href: '/login', primary: true },
        { label: '查看帮助中心', href: '/help' },
      ]}
      sections={[
        {
          title: '会记录的数据',
          items: [
            '登录后会保存认证 Cookie，用于维持会话状态。',
            '收藏、播放记录、搜索历史会保存在当前站点的数据存储中，用于恢复你的使用进度。',
            '内容模式偏好会保存在浏览器 Cookie 中，只影响当前浏览器实例。',
          ],
        },
        {
          title: '不会做的事',
          items: [
            '不会把你的内容模式偏好自动同步到其他设备。',
            '不会在未登录状态下伪造你的收藏或播放记录。',
            '不会因为切换儿童模式或成人模式而直接修改站点默认配置。',
          ],
        },
      ]}
    />
  );
}
