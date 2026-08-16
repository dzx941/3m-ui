import React from 'react';
import { Layout, Menu } from 'antd';
import {
  DashboardOutlined, NodeIndexOutlined, UserOutlined, SettingOutlined,
  CodeOutlined, FileTextOutlined, ApiOutlined, LogoutOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';
import { useI18n } from '../i18n';

const { Sider } = Layout;

const Sidebar: React.FC<{ collapsed: boolean }> = ({ collapsed }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const logout = useAuthStore((s) => s.logout);
  const { t } = useI18n();

  const items = [
    { key: '/', icon: <DashboardOutlined />, label: t('nav.dashboard') },
    { key: '/listeners', icon: <NodeIndexOutlined />, label: t('nav.listeners') },
    { key: '/users', icon: <UserOutlined />, label: t('nav.users') },
    { key: '/core', icon: <ApiOutlined />, label: t('nav.core') },
    { key: '/logs', icon: <FileTextOutlined />, label: t('nav.logs') },
    { key: '/config', icon: <CodeOutlined />, label: t('nav.config') },
    { key: '/settings', icon: <SettingOutlined />, label: t('nav.settings') },
  ];

  return (
    <Sider trigger={null} collapsible collapsed={collapsed} theme="light">
      <div style={{ height: 64, display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 700, fontSize: 18 }}>
        3M-UI
      </div>
      <Menu mode="inline" selectedKeys={[location.pathname]} items={items} onClick={({ key }) => navigate(key)} />
      <Menu mode="inline" selectable={false} items={[
        { key: 'logout', icon: <LogoutOutlined />, label: t('nav.logout'), onClick: () => { logout(); navigate('/login'); } },
      ]} style={{ position: 'absolute', bottom: 0, width: '100%' }} />
    </Sider>
  );
};

export default Sidebar;
