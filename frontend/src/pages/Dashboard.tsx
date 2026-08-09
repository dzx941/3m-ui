import React, { useState, useEffect } from 'react';
import { Card, Typography, Row, Col, Statistic, Button, Space, Tag, message, Descriptions } from 'antd';
import {
  DesktopOutlined,
  TeamOutlined,
  CloudServerOutlined,
  PlayCircleOutlined,
  StopOutlined,
  ReloadOutlined,
} from '@ant-design/icons';

const { Title, Paragraph } = Typography;

interface MihomoStatus {
  running: boolean;
  version: string;
  pid: number;
  uptime: string;
}

const Dashboard: React.FC = () => {
  const [status, setStatus] = useState<MihomoStatus | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [errorMsg, setErrorMsg] = useState<string>('');

  // Base URL of backend API
  const API_BASE = 'http://localhost:8080/api/v1';

  const fetchStatus = async () => {
    try {
      const res = await fetch(`${API_BASE}/mihomo/status`);
      if (!res.ok) {
        throw new Error(`HTTP error! status: ${res.status}`);
      }
      const data: MihomoStatus = await res.json();
      setStatus(data);
      setErrorMsg('');
    } catch (err: any) {
      setErrorMsg('Backend service unreachable');
      setStatus(null);
    }
  };

  useEffect(() => {
    fetchStatus();
    // Poll every 3 seconds
    const interval = setInterval(fetchStatus, 3000);
    return () => clearInterval(interval);
  }, []);

  const handleAction = async (action: 'start' | 'stop' | 'restart') => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/mihomo/${action}`, {
        method: 'POST',
      });
      const data = await res.json();
      if (res.ok) {
        message.success(`Mihomo Core ${action}ed successfully!`);
        await fetchStatus();
      } else {
        message.error(data.error || `Failed to ${action} Mihomo Core.`);
      }
    } catch (err) {
      message.error(`Connection failed while trying to ${action} Mihomo.`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Title level={2}>Dashboard</Title>
      <Paragraph>
        Welcome to the 3m-ui management panel dashboard. View and control the status of your Mihomo Core service.
      </Paragraph>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        {/* Core Status Card */}
        <Col xs={24} lg={16}>
          <Card
            title={
              <Space>
                <DesktopOutlined />
                <span>Mihomo Core Service Control</span>
              </Space>
            }
            bordered={false}
            extra={
              errorMsg ? (
                <Tag color="warning">{errorMsg}</Tag>
              ) : status?.running ? (
                <Tag color="success">Running</Tag>
              ) : (
                <Tag color="error">Stopped</Tag>
              )
            }
          >
            <div style={{ marginBottom: 24 }}>
              <Descriptions bordered column={{ xs: 1, sm: 2 }}>
                <Descriptions.Item label="Service Status">
                  {status?.running ? (
                    <span style={{ color: '#52c41a', fontWeight: 'bold' }}>Active</span>
                  ) : (
                    <span style={{ color: '#ff4d4f', fontWeight: 'bold' }}>Inactive</span>
                  )}
                </Descriptions.Item>
                <Descriptions.Item label="PID">
                  {status?.running ? status.pid : 'N/A'}
                </Descriptions.Item>
                <Descriptions.Item label="Uptime">
                  {status?.running ? status.uptime : '0s'}
                </Descriptions.Item>
                <Descriptions.Item label="Core Version">
                  {status ? status.version : 'Unknown'}
                </Descriptions.Item>
              </Descriptions>
            </div>

            <Space size="middle">
              <Button
                type="primary"
                icon={<PlayCircleOutlined />}
                disabled={status?.running || loading}
                onClick={() => handleAction('start')}
              >
                Start
              </Button>
              <Button
                type="primary"
                danger
                icon={<StopOutlined />}
                disabled={!status?.running || loading}
                onClick={() => handleAction('stop')}
              >
                Stop
              </Button>
              <Button
                icon={<ReloadOutlined />}
                disabled={loading}
                onClick={() => handleAction('restart')}
              >
                Restart
              </Button>
            </Space>
          </Card>
        </Col>

        {/* Info Cards Side panel */}
        <Col xs={24} lg={8}>
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <Card bordered={false}>
              <Statistic
                title="Active Listeners"
                value={0}
                prefix={<CloudServerOutlined />}
              />
            </Card>
            <Card bordered={false}>
              <Statistic
                title="Total Users"
                value={0}
                prefix={<TeamOutlined />}
              />
            </Card>
          </Space>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
