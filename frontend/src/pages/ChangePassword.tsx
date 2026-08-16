import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Form, Input, Button, Typography, Alert, message } from 'antd';
import { changePassword } from '../api/auth';

const { Title } = Typography;

const ChangePassword: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: { current_password: string; new_password: string; confirm: string }) => {
    if (values.new_password !== values.confirm) {
      message.error('Passwords do not match');
      return;
    }
    setLoading(true);
    try {
      await changePassword(values.current_password, values.new_password);
      message.success('Password updated');
      navigate('/');
    } catch (e: any) {
      message.error(e.message || 'Failed to change password');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ maxWidth: 480, margin: '40px auto' }}>
      <Card>
        <Title level={4}>Change Password</Title>
        <Alert message="You must change your password before continuing" type="warning" showIcon style={{ marginBottom: 24 }} />
        <Form layout="vertical" onFinish={onFinish}>
          <Form.Item label="Current Password" name="current_password" rules={[{ required: true }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item label="New Password" name="new_password" rules={[{ required: true, min: 8 }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item label="Confirm Password" name="confirm" rules={[{ required: true }]}>
            <Input.Password />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} block>
            Update Password
          </Button>
        </Form>
      </Card>
    </div>
  );
};

export default ChangePassword;
