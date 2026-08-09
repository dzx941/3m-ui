import React from 'react';
import { Typography, Table, Button } from 'antd';
import { PlusOutlined } from '@ant-design/icons';

const { Title, Paragraph } = Typography;

const Listeners: React.FC = () => {
  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>Listeners</Title>
        <Button type="primary" icon={<PlusOutlined />} disabled>
          Add Listener
        </Button>
      </div>
      <Paragraph>
        Manage your Mihomo inbound/listener nodes.
      </Paragraph>
      <Table
        dataSource={[]}
        columns={[
          { title: 'ID', dataIndex: 'id', key: 'id' },
          { title: 'Name', dataIndex: 'name', key: 'name' },
          { title: 'Type', dataIndex: 'type', key: 'type' },
          { title: 'Listen IP', dataIndex: 'listen', key: 'listen' },
          { title: 'Port', dataIndex: 'port', key: 'port' },
          { title: 'Status', dataIndex: 'enabled', key: 'enabled' },
        ]}
        locale={{ emptyText: 'No Listeners Created Yet' }}
      />
    </div>
  );
};

export default Listeners;
