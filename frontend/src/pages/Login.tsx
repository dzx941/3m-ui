import React, { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Card, Form, Input, Button, Typography, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { login } from '../api/auth';

const { Title } = Typography;

const Login: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [loading, setLoading] = useState(false);
  const from = (location.state as any)?.from?.pathname || '/';

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      await login(values);
      message.success('Welcome back');
      navigate(from, { replace: true });
    } catch (e: any) {
      message.error(e.message || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f5f5f5' }}>
      <Card style={{ width: 420, borderRadius: 12 }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <Title level={3} style={{ margin: 0 }}>3M-UI</Title>
          <Typography.Text type="secondary">Mihomo Core Management</Typography.Text>
        </div>
        <Form onFinish={onFinish} autoComplete="off">
          <Form.Item name="username" rules={[{ required: true, message: 'Please input username' }]}>
            <Input prefix={<UserOutlined />} placeholder="Username" size="large" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: 'Please input password' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="Password" size="large" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block size="large">
              Sign In
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default Login;
