import React from 'react';
import { Card, Tabs } from 'antd';
import { ApiOutlined, CloudServerOutlined, LinkOutlined } from '@ant-design/icons';
import ListenersPage from './Listeners';
import ProxyNodes from './ProxyNodes';
import ClientAccessPage from './ClientAccess';

export default function ConnectionsPage() {
  return (
    <Card
      bordered={false}
      styles={{ body: { padding: 0 } }}
    >
      <Tabs
        defaultActiveKey="listeners"
        size="large"
        items={[
          {
            key: 'listeners',
            label: (
              <span>
                <ApiOutlined />
                监听器
              </span>
            ),
            children: <ListenersPage />,
          },
          {
            key: 'proxies',
            label: (
              <span>
                <CloudServerOutlined />
                代理节点
              </span>
            ),
            children: <ProxyNodes />,
          },
          {
            key: 'client-access',
            label: (
              <span>
                <LinkOutlined />
                客户端接入
              </span>
            ),
            children: <ClientAccessPage />,
          },
        ]}
      />
    </Card>
  );
}
