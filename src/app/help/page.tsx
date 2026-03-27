import { Metadata } from 'next';

import StaticInfoPage from '@/components/info/StaticInfoPage';

export const metadata: Metadata = {
  title: '帮助中心 - ManboTV',
  description: 'ManboTV 使用帮助与常见问题',
};

export default function HelpPage() {
  return (
    <StaticInfoPage
      title='帮助中心'
      description='这里集中说明登录、账号、内容模式和播放排查方式。把原来登录页那些点不动的假链接都换成了真实帮助入口。'
      badge='Help Center'
      actions={[
        { label: '返回登录', href: '/login', primary: true },
        { label: '查看隐私政策', href: '/privacy' },
      ]}
      sections={[
        {
          title: '登录与账号',
          items: [
            '默认账号由站点管理员创建；如果你没有账号，请联系管理员分配。',
            '登录失败时优先确认用户名、密码和浏览器 Cookie 未被禁用。',
            '如果切换设备后状态丢失，通常是登录 Cookie 过期或被浏览器清理。',
          ],
        },
        {
          title: '内容与播放',
          items: [
            '儿童模式只显示普通内容，成人模式只显示成人内容，标准模式则两者都显示。',
            '进入播放页后若线路不可用，请尝试切换线路或返回搜索页选择其他资源源站。',
            '如果封面、详情或播放速度异常，刷新一次页面通常会触发最新代理和恢复链路。',
          ],
        },
      ]}
    />
  );
}
