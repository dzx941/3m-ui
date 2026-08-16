import React from 'react';
import { Layout, Button, Space, Tag, Dropdown } from 'antd';
import { MenuFoldOutlined, MenuUnfoldOutlined, UserOutlined, GlobalOutlined, BgColorsOutlined } from '@ant-design/icons';
import { useAuthStore } from '../stores/authStore';
import { useThemeStore, ThemeMode } from '../stores/themeStore';
import { useI18n } from '../i18n';

const { Header } = Layout;

const HeaderBar: React.FC<{ collapsed: boolean; setCollapsed: (v: boolean) => void; }> = ({ collapsed, setCollapsed }) => {
  const { t, locale, setLocale } = useI18n();
  const { mode, setMode } = useThemeStore();
  const username = useAuthStore((s) => s.username);

  const langItems = [
    { key: 'zh', label: '中文' },
    { key: 'en', label: 'English' },
  ];

  const themeItems = [
    { key: 'light', label: t('settings.light') },
    { key: 'dark', label: t('settings.dark') },
    { key: 'system', label: t('settings.system') },
  ];

  return (
    <Header style={{ padding: '0 24px', background: 'transparent', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
      <Button type="text" icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />} onClick={() => setCollapsed(!collapsed)} />
      <Space>
        <Dropdown menu={{ items: themeItems, selectedKeys: [mode], onClick: (e) => setMode(e.key as ThemeMode) }}>
          <Button type="text" icon={<BgColorsOutlined />}>{t('settings.theme')}</Button>
        </Dropdown>
        <Dropdown menu={{ items: langItems, selectedKeys: [locale], onClick: (e) => setLocale(e.key as 'en' | 'zh') }}>
          <Button type="text" icon={<GlobalOutlined />}>{locale === 'zh' ? '中文' : 'English'}</Button>
        </Dropdown>
        <Tag icon={<UserOutlined />}>{username || 'Admin'}</Tag>
      </Space>
    </Header>
  );
};

export default HeaderBar;
