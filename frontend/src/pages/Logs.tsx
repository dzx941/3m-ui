import React from 'react';
import { Typography, Card, Space, Button } from 'antd';

const { Title, Paragraph } = Typography;

const Logs: React.FC = () => {
  return (
    <div>
      <Title level={2}>系统日志</Title>
      <Paragraph>
        后端服务与 Mihomo 内核实时日志输出。
      </Paragraph>
      <Space style={{ marginBottom: 16 }}>
        <Button disabled>清空日志</Button>
        <Button type="primary" disabled>暂停流式输出</Button>
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
