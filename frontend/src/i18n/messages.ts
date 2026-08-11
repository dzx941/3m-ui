export type Locale = 'zh-CN' | 'en-US';
export type Messages = Record<string, string>;

export const messages: Record<Locale, Messages> = {
  'zh-CN': {
    'login.title': '3m-ui 管理面板',
    'login.subtitle': '登录以管理 Mihomo Core 与代理服务',
    'login.username': '用户名',
    'login.password': '密码',
    'login.submit': '登录',
    'login.success': '登录成功',
    'login.failed': '登录失败，请检查用户名和密码',
    'login.requiredUsername': '请输入用户名',
    'login.requiredPassword': '请输入密码',
    'auth.logout': '退出登录',
  },
  'en-US': {
    'login.title': '3m-ui Panel',
    'login.subtitle': 'Sign in to manage Mihomo Core and proxy services',
    'login.username': 'Username',
    'login.password': 'Password',
    'login.submit': 'Log in',
    'login.success': 'Logged in successfully',
    'login.failed': 'Login failed. Please check your username and password.',
    'login.requiredUsername': 'Please enter your username',
    'login.requiredPassword': 'Please enter your password',
    'auth.logout': 'Log out',
  },
};
