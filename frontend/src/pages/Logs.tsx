import React from 'react';
import { Typography, Card, Space, Button } from 'antd';

const { Title, Paragraph } = Typography;

const Logs: React.FC = () => {
  return (
    <div>
      <Title level={2}>System Logs</Title>
      <Paragraph>
        Real-time logging output from the backend and Mihomo core services.
      </Paragraph>
      <Space style={{ marginBottom: 16 }}>
        <Button disabled>Clear Log View</Button>
        <Button type="primary" disabled>Pause Streaming</Button>
      </Space>
      <Card
        bodyStyle={{
          backgroundColor: '#001529',
          color: '#ffffff',
          fontFamily: 'monospace',
          minHeight: '300px',
          borderRadius: '4px',
        }}
      >
        <div style={{ color: '#00ff00' }}>[3m-ui-system] Initialization: Web server loaded.</div>
        <div style={{ color: '#00ff00' }}>[3m-ui-system] SQLite connection established.</div>
        <div>[mihomo-core] Core binary detected: waiting for first service run...</div>
      </Card>
    </div>
  );
};

export default Logs;
