import React from 'react';
import { Typography, Card, Input, Button } from 'antd';

const { Title, Paragraph } = Typography;
const { TextArea } = Input;

const Config: React.FC = () => {
  return (
    <div>
      <Title level={2}>Config</Title>
      <Paragraph>
        Global Mihomo / system configuration YAML template.
      </Paragraph>
      <Card style={{ marginTop: 24 }}>
        <Title level={4}>YAML Core Template</Title>
        <TextArea
          rows={10}
          defaultValue={`# Default configuration template
dns:
  enable: true
  listen: 0.0.0.0:1053
  ipv6: false
  enhanced-mode: fake-ip
  nameserver:
    - 119.29.29.29
    - 223.5.5.5
`}
          style={{ fontFamily: 'monospace' }}
        />
        <Button type="primary" style={{ marginTop: 16 }} disabled>
          Save Template
        </Button>
      </Card>
    </div>
  );
};

export default Config;
