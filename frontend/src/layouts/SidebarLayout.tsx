import React from 'react';
import { Layout, Menu } from 'antd';
import { Link, useLocation } from 'react-router-dom';
import {
  DashboardOutlined,
  CloudServerOutlined,
  TeamOutlined,
  LinkOutlined,
  FileTextOutlined,
  CodeOutlined,
  SettingOutlined,
} from '@ant-design/icons';

const { Header, Content, Sider } = Layout;

interface SidebarLayoutProps {
  children: React.ReactNode;
}

const SidebarLayout: React.FC<SidebarLayoutProps> = ({ children }) => {
  const location = useLocation();

  const menuItems = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: <Link to="/dashboard">Dashboard</Link>,
    },
    {
      key: '/listeners',
      icon: <CloudServerOutlined />,
      label: <Link to="/listeners">Listeners</Link>,
    },
    {
      key: '/users',
      icon: <TeamOutlined />,
      label: <Link to="/users">Users</Link>,
    },
    {
      key: '/subscriptions',
      icon: <LinkOutlined />,
      label: <Link to="/subscriptions">Subscriptions</Link>,
    },
    {
      key: '/config',
      icon: <CodeOutlined />,
      label: <Link to="/config">Config</Link>,
    },
    {
      key: '/logs',
      icon: <FileTextOutlined />,
      label: <Link to="/logs">Logs</Link>,
    },
    {
      key: '/settings',
      icon: <SettingOutlined />,
      label: <Link to="/settings">Settings</Link>,
    },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider width={240} theme="dark" breakpoint="lg" collapsedWidth="0">
        <div style={{ height: 32, margin: 16, background: 'rgba(255, 255, 255, 0.2)', borderRadius: 6, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <span style={{ color: '#fff', fontWeight: 'bold', fontSize: '16px' }}>3m-ui Panel</span>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
        />
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', padding: '0 24px', display: 'flex', alignItems: 'center', justifyContent: 'flex-end', boxShadow: '0 1px 4px rgba(0,21,41,.08)' }}>
          <Link to="/login" style={{ color: '#333' }}>Logout</Link>
        </Header>
        <Content style={{ margin: '24px 16px', padding: 24, background: '#fff', minHeight: 280, borderRadius: 8 }}>
          {children}
        </Content>
      </Layout>
    </Layout>
  );
};

export default SidebarLayout;
