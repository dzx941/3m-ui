import React from 'react';
import { Card, Typography } from 'antd';

const { Title, Paragraph } = Typography;

const Login: React.FC = () => {
  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', backgroundColor: '#f0f2f5' }}>
      <Card style={{ width: 400, boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}>
        <Title level={3} style={{ textAlign: 'center', marginBottom: 24 }}>3m-ui Panel</Title>
        <Paragraph style={{ textAlign: 'center' }}>
          This is the Login page placeholder. Standard login layout and authentication logic will be implemented here in subsequent phases.
        </Paragraph>
      </Card>
    </div>
  );
};

export default Login;
