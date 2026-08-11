import React, { useState } from 'react';
import { Button, Card, Form, Input, Typography, message } from 'antd';
import { LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { changePassword } from '../api/auth';
import { useI18n } from '../i18n';

const { Title, Paragraph } = Typography;

const ChangePassword: React.FC = () => {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: { currentPassword: string; newPassword: string; confirmPassword: string }) => {
    setLoading(true);
    try {
      await changePassword(values.currentPassword, values.newPassword);
      message.success(t('password.success'));
      navigate('/dashboard', { replace: true });
    } catch (err) {
      message.error((err as { message?: string }).message || t('password.failed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', backgroundColor: '#f0f2f5' }}>
      <Card style={{ width: 400, boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}>
        <Title level={3} style={{ textAlign: 'center', marginBottom: 8 }}>{t('password.title')}</Title>
        <Paragraph style={{ textAlign: 'center', marginBottom: 24 }}>{t('password.subtitle')}</Paragraph>
        <Form layout="vertical" onFinish={onFinish} requiredMark={false}>
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
                validator(_, value) {
                  return !value || getFieldValue('newPassword') === value
                    ? Promise.resolve()
                    : Promise.reject(new Error(t('password.mismatch')));
                },
              }),
            ]}
          >
            <Input.Password prefix={<LockOutlined />} autoComplete="new-password" />
          </Form.Item>
          <Button type="primary" htmlType="submit" block loading={loading}>{t('password.submit')}</Button>
        </Form>
      </Card>
    </div>
  );
};

export default ChangePassword;
