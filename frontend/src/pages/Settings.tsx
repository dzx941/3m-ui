import React from 'react';
import { Card, Button, Space, Typography, Tag } from 'antd';
import { useNavigate } from 'react-router-dom';
import { useI18n } from '../i18n';
import { useThemeStore, ThemeMode } from '../stores/themeStore';
import { LockOutlined, GlobalOutlined, BgColorsOutlined, InfoCircleOutlined } from '@ant-design/icons';

const { Title, Text } = Typography;

const Settings: React.FC = () => {
  const { t, locale, setLocale } = useI18n();
  const { mode, setMode } = useThemeStore();
  const navigate = useNavigate();

  return (
    <div>
      <h2>{t('settings.title')}</h2>
      <p style={{ color: 'rgba(0,0,0,0.45)' }}>{t('settings.subtitle')}</p>
      <Card title={<><GlobalOutlined /> {t('settings.language')}</>} style={{ marginBottom: 16 }}>
        <Space>
          <Button type={locale === 'zh' ? 'primary' : 'default'} onClick={() => setLocale('zh')}>中文</Button>
          <Button type={locale === 'en' ? 'primary' : 'default'} onClick={() => setLocale('en')}>English</Button>
        </Space>
      </Card>
      <Card title={<><BgColorsOutlined /> {t('settings.theme')}</>} style={{ marginBottom: 16 }}>
        <Space>
          <Button type={mode === 'light' ? 'primary' : 'default'} onClick={() => setMode('light')}>{t('settings.light')}</Button>
          <Button type={mode === 'dark' ? 'primary' : 'default'} onClick={() => setMode('dark')}>{t('settings.dark')}</Button>
          <Button type={mode === 'system' ? 'primary' : 'default'} onClick={() => setMode('system')}>{t('settings.system')}</Button>
        </Space>
      </Card>
      <Card title={<><LockOutlined /> {t('settings.passwordTitle')}</>} style={{ marginBottom: 16 }}>
        <Text type="secondary">{t('settings.passwordDescription')}</Text>
        <div style={{ marginTop: 12 }}>
          <Button type="primary" onClick={() => navigate('/change-password')}>{t('settings.changePassword')}</Button>
        </div>
      </Card>
      <Card title={<><InfoCircleOutlined /> {t('settings.about')}</>}>
        <Space direction="vertical">
          <Text strong>3M-UI</Text>
          <Text type="secondary">{t('app.title')}</Text>
          <Tag color="blue">v1.0.0</Tag>
        </Space>
      </Card>
    </div>
  );
};

export default Settings;
