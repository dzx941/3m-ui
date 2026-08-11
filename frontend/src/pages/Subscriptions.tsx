import React from 'react';
import { Typography, Table, Button } from 'antd';
import { LinkOutlined } from '@ant-design/icons';

const { Title, Paragraph } = Typography;

const Subscriptions: React.FC = () => {
  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>Subscriptions</Title>
        <Button type="primary" icon={<LinkOutlined />} disabled>
          Generate Subscription
        </Button>
      </div>
      <Paragraph>
        Generate and view sub-user subscriptions for Mihomo/Clash clients.
      </Paragraph>
      <Table
        dataSource={[]}
        columns={[
          { title: 'ID', dataIndex: 'id', key: 'id' },
          { title: 'User ID', dataIndex: 'userId', key: 'userId' },
          { title: 'Token', dataIndex: 'token', key: 'token' },
          { title: 'Format', dataIndex: 'format', key: 'format' },
          { title: 'Expire Time', dataIndex: 'expireTime', key: 'expireTime' },
        ]}
        locale={{ emptyText: 'No Subscriptions Generated' }}
      />
    </div>
  );
};

export default Subscriptions;
