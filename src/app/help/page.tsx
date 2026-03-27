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
        {
          title: '版权与通知处理',
          items: [
            'ManboTV 仅提供程序功能、聚合入口与代理能力，不直接存储、制作或上传影视内容。',
            '若第三方数据源、封面、详情或播放链接涉嫌侵权，请向站点管理员提交权利通知及权属证明，核实后应及时删除、屏蔽或断开相关链接。',
            '使用者应在下载、缓存或访问相关内容后 24 小时内自行删除，并确保其使用行为符合所在地法律法规。',
            '项目明确禁止在中国大陆法律管辖范围内传播、推广、分发或镜像，也禁止发布到抖音、哔哩哔哩、小红书、微博、微信公众号、视频号、快手、知乎、百度贴吧等中国大陆平台。',
          ],
        },
      ]}
    />
  );
}
