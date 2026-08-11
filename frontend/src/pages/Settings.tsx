import React from 'react';
import { Typography, Form, Input, Button, Card } from 'antd';

const { Title, Paragraph } = Typography;

const Settings: React.FC = () => {
  return (
    <div>
      <Title level={2}>Settings</Title>
      <Paragraph>
        Configure VPS panel ports, security domains, and JWT tokens.
      </Paragraph>
      <Card style={{ marginTop: 24, maxWidth: 600 }}>
        <Form layout="vertical">
          <Form.Item label="Server Port" name="port">
            <Input defaultValue="8080" />
          </Form.Item>
          <Form.Item label="Mihomo Binary Path" name="mihomoPath">
            <Input defaultValue="/usr/local/bin/mihomo" />
          </Form.Item>
          <Form.Item label="Mihomo Configuration Path" name="configPath">
            <Input defaultValue="/etc/mihomo/config.yaml" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" disabled>
              Update Configuration
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default Settings;
