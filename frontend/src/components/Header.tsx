import React from 'react';
import { Layout, Button, theme, Space, Tag } from 'antd';
import { MenuFoldOutlined, MenuUnfoldOutlined, UserOutlined } from '@ant-design/icons';
import { useAuthStore } from '../stores/authStore';

const { Header } = Layout;

const HeaderBar: React.FC<{
  collapsed: boolean;
  setCollapsed: (v: boolean) => void;
}> = ({ collapsed, setCollapsed }) => {
  const { token } = theme.useToken();
  const username = useAuthStore((s) => s.username);

  return (
    <Header style={{ padding: '0 24px', background: token.colorBgContainer, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
      <Button
        type="text"
        icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
        onClick={() => setCollapsed(!collapsed)}
      />
      <Space>
        <Tag icon={<UserOutlined />}>{username || 'Admin'}</Tag>
      </Space>
    </Header>
  );
};

export default HeaderBar;
