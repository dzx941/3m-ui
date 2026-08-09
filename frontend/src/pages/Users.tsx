import React from 'react';
import { Typography, Table, Button } from 'antd';
import { UserAddOutlined } from '@ant-design/icons';

const { Title, Paragraph } = Typography;

const Users: React.FC = () => {
  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>Users</Title>
        <Button type="primary" icon={<UserAddOutlined />} disabled>
          Create User
        </Button>
      </div>
      <Paragraph>
        Manage administrator and sub-users of the VPS.
      </Paragraph>
      <Table
        dataSource={[]}
        columns={[
          { title: 'ID', dataIndex: 'id', key: 'id' },
          { title: 'Username', dataIndex: 'username', key: 'username' },
          { title: 'Role', dataIndex: 'role', key: 'role' },
          { title: 'Created At', dataIndex: 'createdAt', key: 'createdAt' },
        ]}
        locale={{ emptyText: 'No Users' }}
      />
    </div>
  );
};

export default Users;
