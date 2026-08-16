import React from 'react';
import { Layout, Menu } from 'antd';
import {
  DashboardOutlined,
  NodeIndexOutlined,
  UserOutlined,
  SettingOutlined,
  CodeOutlined,
  FileTextOutlined,
  ApiOutlined,
  LogoutOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';

const { Sider } = Layout;

const items = [
  { key: '/', icon: <DashboardOutlined />, label: '仪表盘' },
  { key: '/listeners', icon: <NodeIndexOutlined />, label: '监听器' },
  { key: '/users', icon: <UserOutlined />, label: '用户' },
  { key: '/core', icon: <ApiOutlined />, label: '核心' },
  { key: '/logs', icon: <FileTextOutlined />, label: '日志' },
  { key: '/config', icon: <CodeOutlined />, label: '配置' },
  { key: '/settings', icon: <SettingOutlined />, label: '设置' },
];

const Sidebar: React.FC<{ collapsed: boolean }> = ({ collapsed }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const logout = useAuthStore((s) => s.logout);

  return (
    <Sider trigger={null} collapsible collapsed={collapsed} theme="light">
      <div
        style={{
          height: 64,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontWeight: 700,
          fontSize: 18,
          borderBottom: '1px solid #f0f0f0',
        }}
      >
        3M-UI
      </div>
      <Menu
        mode="inline"
        selectedKeys={[location.pathname]}
        items={items}
        onClick={({ key }) => navigate(key)}
      />
      <Menu
        mode="inline"
        selectable={false}
        items={[
          {
            key: 'logout',
            icon: <LogoutOutlined />,
            label: '退出登录',
            onClick: () => {
              logout();
              navigate('/login');
            },
          },
        ]}
        style={{ position: 'absolute', bottom: 0, width: '100%' }}
      />
    </Sider>
  );
};

export default Sidebar;
