import { Metadata } from 'next';

import StaticInfoPage from '@/components/info/StaticInfoPage';

export const metadata: Metadata = {
  title: '账号申请 - ManboTV',
  description: 'ManboTV 账号申请说明',
};

export default function RegisterPage() {
  return (
    <StaticInfoPage
      title='账号申请说明'
      description='当前站点不开放匿名自助注册。为了控制访问权限、内容模式和线路范围，账号由管理员统一创建与分组。'
      badge='Access Request'
      actions={[
        { label: '返回登录', href: '/login', primary: true },
        { label: '查看帮助中心', href: '/help' },
      ]}
      sections={[
        {
          title: '为什么不开放自助注册',
          items: [
            '站点包含不同内容模式和可访问源站，需要统一做权限控制。',
            '管理员可以按用户组控制可见视频源、是否允许访问后台以及可用内容范围。',
            '这比开放匿名注册更适合私有部署和家庭共享场景。',
          ],
        },
        {
          title: '如何获得账号',
          items: [
            '联系站点管理员创建账号。',
            '管理员创建后，你可以直接使用分配的用户名和初始密码登录。',
            '首次登录成功后，建议进入收藏页修改密码。',
          ],
        },
      ]}
    />
  );
}
