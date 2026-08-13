import React, { useState } from 'react';
import { Layout, Menu, Button, Drawer } from 'antd';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { DashboardOutlined, CloudServerOutlined, CodeOutlined, FileTextOutlined, SettingOutlined, LogoutOutlined, GlobalOutlined, MenuOutlined } from '@ant-design/icons';
import { logout } from '../api/auth';
import { useI18n } from '../i18n';

const { Header, Content, Sider } = Layout;

export default function SidebarLayout({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  const navigate = useNavigate();
  const { t, locale, setLocale } = useI18n();
  const [mobileOpen, setMobileOpen] = useState(false);
  const section = location.pathname.split('/')[1] || 'dashboard';
  const selected = section === 'listeners' || section === 'proxies' || section === 'users' || section === 'client-access' ? '/nodes' : `/${section}`;

  const items = [
    { key: '/dashboard', icon: <DashboardOutlined />, label: <Link to="/dashboard">{t('nav.dashboard')}</Link> },
    { key: '/core', icon: <CloudServerOutlined />, label: <Link to="/core">{t('nav.core')}</Link> },
    { key: '/nodes', icon: <CloudServerOutlined />, label: <Link to="/nodes">{t('nav.nodes')}</Link> },
    { key: '/config', icon: <CodeOutlined />, label: <Link to="/config">{t('nav.config')}</Link> },
    { key: '/logs', icon: <FileTextOutlined />, label: <Link to="/logs">{t('nav.logs')}</Link> },
    { key: '/settings', icon: <SettingOutlined />, label: <Link to="/settings">{t('nav.settings')}</Link> },
  ];

  const go = (path: string) => {
    setMobileOpen(false);
    navigate(path);
  };

  const menu = (
    <Menu
      theme="dark"
      mode="inline"
      selectedKeys={[selected]}
      items={items}
      onClick={({ key }) => setMobileOpen(false)}
      className="app-menu"
    />
  );

  const switchLocale = () => setLocale(locale === 'zh-CN' ? 'en-US' : 'zh-CN');
  const doLogout = () => {
    logout();
    navigate('/login', { replace: true });
  };

  return (
    <Layout className="app-shell">
      <Sider width={240} theme="dark" breakpoint="lg" collapsedWidth="0" className="app-sider">
        <div className="app-brand">
          <span className="app-brand-name">3m-ui</span>
          <span className="app-brand-subtitle">Mihomo Console</span>
        </div>
        {menu}
        <div className="app-sidebar-footer">
          <Button type="text" icon={<GlobalOutlined />} className="app-language" onClick={switchLocale}>
            {locale === 'zh-CN' ? 'English' : '中文'}
          </Button>
        </div>
      </Sider>

      <Layout>
        <Header className="app-header">
          <Button
            type="text"
            icon={<MenuOutlined />}
            className="mobile-menu-button"
            aria-label="Open navigation"
            onClick={() => setMobileOpen(true)}
          />
          <div className="mobile-brand">3m-ui</div>
          <div className="header-actions">
            <Button type="text" icon={<GlobalOutlined />} className="desktop-language" onClick={switchLocale}>
              {locale === 'zh-CN' ? 'English' : '中文'}
            </Button>
            <Button type="text" icon={<LogoutOutlined />} onClick={doLogout}>
              <span className="logout-label">{t('auth.logout')}</span>
            </Button>
          </div>
        </Header>

        <Drawer
          title={<span className="drawer-brand">3m-ui</span>}
          placement="left"
          width="min(82vw, 300px)"
          open={mobileOpen}
          onClose={() => setMobileOpen(false)}
          className="mobile-nav-drawer"
          styles={{ body: { padding: 0 } }}
        >
          {menu}
          <div className="mobile-drawer-footer">
            <Button type="text" icon={<GlobalOutlined />} onClick={switchLocale} block>
              {locale === 'zh-CN' ? 'English' : '中文'}
            </Button>
          </div>
        </Drawer>

        <Content className="app-content">
          <div className="app-content-inner">{children}</div>
        </Content>
      </Layout>
    </Layout>
  );
}
