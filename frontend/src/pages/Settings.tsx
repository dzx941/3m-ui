import { Card, Form, Input, Typography, Button, Alert, Space } from 'antd';
import { LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useI18n } from '../i18n';

const { Title, Paragraph } = Typography;

export default function Settings() {
  const { t } = useI18n();
  const navigate = useNavigate();

  return (
    <>
      <Title level={2}>{t('settings.title')}</Title>
      <Paragraph>{t('settings.subtitle')}</Paragraph>

      <Alert
        type="info"
        showIcon
        message={t('settings.serviceConfigTitle')}
        description={t('settings.serviceConfigDescription')}
      />

      <Card style={{ marginTop: 24, maxWidth: 700 }} title={t('settings.pathsTitle')}>
        <Form layout="vertical">
          <Form.Item label={t('settings.appConfig')}>
            <Input value="/etc/3m-ui/config.yaml" readOnly />
          </Form.Item>
          <Form.Item label={t('settings.mihomoBinary')}>
            <Input value="/usr/local/bin/mihomo" readOnly />
          </Form.Item>
          <Form.Item label={t('settings.mihomoConfig')}>
            <Input value="/var/lib/3m-ui/mihomo/config.yaml" readOnly />
          </Form.Item>
        </Form>
      </Card>

      <Card style={{ marginTop: 24, maxWidth: 700 }} title={t('settings.securityTitle')}>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <div>
            <Typography.Text strong>{t('settings.passwordTitle')}</Typography.Text>
            <Paragraph type="secondary" style={{ marginTop: 4, marginBottom: 0 }}>
              {t('settings.passwordDescription')}
            </Paragraph>
          </div>
          <Button type="primary" icon={<LockOutlined />} onClick={() => navigate('/change-password')}>
            {t('settings.changePassword')}
          </Button>
        </Space>
      </Card>
    </>
  );
}
