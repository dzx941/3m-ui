import React from 'react';
import { Card, Typography, Row, Col, Statistic } from 'antd';
import { DesktopOutlined, TeamOutlined, CloudServerOutlined } from '@ant-design/icons';

const { Title, Paragraph } = Typography;

const Dashboard: React.FC = () => {
  return (
    <div>
      <Title level={2}>Dashboard</Title>
      <Paragraph>
        Welcome to the 3m-ui management panel dashboard placeholder.
      </Paragraph>
      <Row gutter={16} style={{ marginTop: 24 }}>
        <Col span={8}>
          <Card bordered={false}>
            <Statistic
              title="Active Listeners"
              value={0}
              prefix={<CloudServerOutlined />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card bordered={false}>
            <Statistic
              title="Total Users"
              value={0}
              prefix={<TeamOutlined />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card bordered={false}>
            <Statistic
              title="System Status"
              value="Running"
              valueStyle={{ color: '#3f8600' }}
              prefix={<DesktopOutlined />}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
