import React from 'react';
import { Layout, Menu, Button } from 'antd';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import {
  DashboardOutlined,
  CloudServerOutlined,
  CodeOutlined,
  FileTextOutlined,
  SettingOutlined,
  LogoutOutlined,
  ApiOutlined,
  UserOutlined,
  GlobalOutlined,
} from '@ant-design/icons';
import { logout } from '../api/auth';
import { useI18n } from '../i18n';

const { Header, Content, Sider } = Layout;

export default function SidebarLayout({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  const navigate = useNavigate();
  const { t, locale, setLocale } = useI18n();

  const items = [
    { key: '/dashboard', icon: <DashboardOutlined />, label: <Link to="/dashboard">{t('nav.dashboard')}</Link> },
    { key: '/core', icon: <CloudServerOutlined />, label: <Link to="/core">{t('nav.core')}</Link> },
    { key: '/proxies', icon: <CloudServerOutlined />, label: <Link to="/proxies">{t('nav.nodes')}</Link> },
    { key: '/listeners', icon: <ApiOutlined />, label: <Link to="/listeners">{t('nav.listeners')}</Link> },
    { key: '/users', icon: <UserOutlined />, label: <Link to="/users">{t('nav.users')}</Link> },
    { key: '/client-access', icon: <ApiOutlined />, label: <Link to="/client-access">{t('nav.clientAccess')}</Link> },
    { key: '/config', icon: <CodeOutlined />, label: <Link to="/config">{t('nav.config')}</Link> },
    { key: '/logs', icon: <FileTextOutlined />, label: <Link to="/logs">{t('nav.logs')}</Link> },
    { key: '/settings', icon: <SettingOutlined />, label: <Link to="/settings">{t('nav.settings')}</Link> },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider width={240} theme="dark" breakpoint="lg" collapsedWidth="0" style={{ position: 'relative' }}>
        <div style={{ height: 44, margin: 16, display: 'flex', alignItems: 'center', padding: '0 12px', borderRadius: 8, background: 'rgba(255,255,255,.08)' }}>
          <span style={{ color: '#fff', fontWeight: 700, fontSize: 17 }}>3m-ui</span>
          <span style={{ color: 'rgba(255,255,255,.55)', marginLeft: 8, fontSize: 12 }}>Mihomo Console</span>
        </div>
        <Menu theme="dark" mode="inline" selectedKeys={[`/${location.pathname.split('/')[1] || 'dashboard'}`]} items={items} style={{ marginBottom: 60 }} />
        <div style={{ position: 'absolute', bottom: 24, left: 16, right: 16, display: 'flex', justifyContent: 'center' }}>
          <Button
            type="text"
            icon={<GlobalOutlined />}
            style={{ color: 'rgba(255, 255, 255, 0.65)' }}
            onClick={() => setLocale(locale === 'zh-CN' ? 'en-US' : 'zh-CN')}
          >
            {locale === 'zh-CN' ? 'English' : '中文'}
          </Button>
        </div>
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', padding: '0 24px', display: 'flex', alignItems: 'center', justifyContent: 'flex-end', boxShadow: '0 1px 4px rgba(0,21,41,.08)' }}>
          <a onClick={() => { logout(); navigate('/login', { replace: true }); }}>
            <LogoutOutlined /> {t('auth.logout')}
          </a>
        </Header>
        <Content style={{ margin: 24, padding: 24, background: '#fff', minHeight: 280, borderRadius: 10 }}>
          {children}
        </Content>
      </Layout>
    </Layout>
  );
}
