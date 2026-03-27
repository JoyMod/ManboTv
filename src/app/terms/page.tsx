import { Metadata } from 'next';

import StaticInfoPage from '@/components/info/StaticInfoPage';

export const metadata: Metadata = {
  title: '使用条款 - ManboTV',
  description: 'ManboTV 使用条款',
};

export default function TermsPage() {
  return (
    <StaticInfoPage
      title='使用条款'
      description='这些条款用于说明账户使用、内容访问和站点维护边界，避免之前登录页里只有样式没有落地页面。'
      badge='Terms'
      actions={[
        { label: '返回登录', href: '/login', primary: true },
        { label: '查看隐私政策', href: '/privacy' },
      ]}
      sections={[
        {
          title: '使用范围',
          items: [
            '请仅在站点允许的账号范围内访问内容，不要共享账号给未知第三方。',
            '不同内容模式仅用于个人浏览偏好切换，不代表对内容进行合法性背书。',
            '管理员保留关闭源站、调整内容策略和回收账号权限的权利。',
          ],
        },
        {
          title: '责任边界',
          items: [
            '线路质量与上游资源站稳定性相关，站点会尽量代理和恢复，但无法保证所有时间都可用。',
            '用户应自行对浏览习惯、终端环境和未成年人访问场景负责。',
            '如发现明显错误分类、异常内容或敏感资源泄露，应及时反馈管理员处理。',
          ],
        },
      ]}
    />
  );
}
