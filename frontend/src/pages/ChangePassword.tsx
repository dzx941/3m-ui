import { useState } from 'react';
import { Button, Card, Form, Input, Typography, message } from 'antd';
import { LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { changePassword } from '../api/auth';
import { useI18n } from '../i18n';

const { Title, Paragraph } = Typography;

export default function ChangePassword() {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);

  const submit = async (v: { currentPassword: string; newPassword: string; confirmPassword: string }) => {
    setLoading(true);
    try {
      await changePassword(v.currentPassword, v.newPassword);
      message.success(t('password.success'));
      navigate('/dashboard', { replace: true });
    } catch (e: any) {
      message.error(e.message || t('password.failed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f5f7fa' }}>
      <Card style={{ width: 420 }}>
        <Title level={3}>{t('password.title')}</Title>
        <Paragraph>{t('password.subtitle')}</Paragraph>
        <Form layout="vertical" onFinish={submit} requiredMark={false}>
          <Form.Item name="currentPassword" label={t('password.current')} rules={[{ required: true, message: t('password.required') }]}>
            <Input.Password prefix={<LockOutlined />} autoComplete="current-password" />
          </Form.Item>
          <Form.Item name="newPassword" label={t('password.new')} rules={[{ required: true, message: t('password.required') }, { min: 8, message: t('password.min') }]}>
            <Input.Password prefix={<LockOutlined />} autoComplete="new-password" />
          </Form.Item>
          <Form.Item
            name="confirmPassword"
            label={t('password.confirm')}
            dependencies={['newPassword']}
            rules={[
              { required: true, message: t('password.required') },
              ({ getFieldValue }) => ({
                validator(_, v) {
                  return !v || getFieldValue('newPassword') === v
                    ? Promise.resolve()
                    : Promise.reject(new Error(t('password.mismatch')));
                }
              })
            ]}
          >
            <Input.Password prefix={<LockOutlined />} autoComplete="new-password" />
          </Form.Item>
          <Button type="primary" htmlType="submit" block loading={loading}>{t('password.submit')}</Button>
        </Form>
      </Card>
    </div>
  );
}
