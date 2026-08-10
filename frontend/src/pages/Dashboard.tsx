import React, { useState, useEffect } from 'react';
import { Card, Typography, Row, Col, Statistic, Button, Space, Tag, message, Progress } from 'antd';
import {
  DesktopOutlined,
  PlayCircleOutlined,
  StopOutlined,
  ReloadOutlined,
  CloudServerOutlined,
  DashboardOutlined,
  CloudUploadOutlined,
  CloudDownloadOutlined,
} from '@ant-design/icons';

const { Title, Paragraph } = Typography;

interface DashboardData {
  mihomo: {
    running: boolean;
    version: string;
    pid: number;
    uptime: string;
  };
  system: {
    cpu: {
      percent: number;
    };
    memory: {
      used: number;
      total: number;
      percent: number;
    };
    disk: {
      used: number;
      total: number;
      percent: number;
    };
    network: {
      upload: number;
      download: number;
    };
  };
  listeners: {
    total: number;
    enabled: number;
    disabled: number;
  };
}

const API_BASE = 'http://localhost:8080/api/v1';

const formatRate = (bytesPerSec: number): string => {
  if (!bytesPerSec || bytesPerSec === 0) return '0 B/s';
  if (bytesPerSec < 1024) return `${bytesPerSec.toFixed(0)} B/s`;
  if (bytesPerSec < 1024 * 1024) return `${(bytesPerSec / 1024).toFixed(1)} KB/s`;
  return `${(bytesPerSec / (1024 * 1024)).toFixed(1)} MB/s`;
};

const Dashboard: React.FC = () => {
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [errorMsg, setErrorMsg] = useState<string>('');

  const fetchDashboardData = async () => {
    try {
      const res = await fetch(`${API_BASE}/dashboard`);
      if (!res.ok) {
        throw new Error(`HTTP error! status: ${res.status}`);
      }
      const result: DashboardData = await res.json();
      setData(result);
      setErrorMsg('');
    } catch (err: any) {
      setErrorMsg('Backend service unreachable');
      setData(null);
    }
  };

  useEffect(() => {
    fetchDashboardData();
    // Poll every 10 seconds (optimized for lower server load)
    const interval = setInterval(fetchDashboardData, 10000);
    return () => clearInterval(interval);
  }, []);

  const handleAction = async (action: 'start' | 'stop' | 'restart') => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/mihomo/${action}`, {
        method: 'POST',
      });
      const resData = await res.json();
      if (res.ok) {
        message.success(`Mihomo Core ${action}ed successfully!`);
        await fetchDashboardData();
      } else {
        message.error(resData.error || `Failed to ${action} Mihomo Core.`);
      }
    } catch (err) {
      message.error(`Connection failed while trying to ${action} Mihomo.`);
    } finally {
      setLoading(false);
    }
  };

  const isMihomoRunning = data?.mihomo.running || false;

  return (
    <div>
      <Title level={2}>Dashboard</Title>
      <Paragraph>
        Welcome to the 3m-ui management panel dashboard. Monitor VPS resources, network, and control the status of Mihomo Core in real time.
      </Paragraph>

      {/* Row 1: Mihomo Control, CPU Progress, Memory Progress */}
      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        {/* Mihomo Control */}
        <Col xs={24} md={12} lg={8}>
          <Card
            title={
              <Space>
                <DesktopOutlined />
                <span>Mihomo Control</span>
              </Space>
            }
            bordered={false}
            extra={
              errorMsg ? (
                <Tag color="warning">{errorMsg}</Tag>
              ) : isMihomoRunning ? (
                <Tag color="success">Running</Tag>
              ) : (
                <Tag color="error">Stopped</Tag>
              )
            }
            style={{ height: '100%' }}
          >
            <div style={{ marginBottom: 16 }}>
              <div style={{ fontSize: '13px', color: '#8c8c8c', marginBottom: 4 }}>
                PID: <strong style={{ color: '#595959' }}>{isMihomoRunning ? data?.mihomo.pid : 'N/A'}</strong>
              </div>
              <div style={{ fontSize: '13px', color: '#8c8c8c', marginBottom: 4 }}>
                Uptime: <strong style={{ color: '#595959' }}>{isMihomoRunning ? data?.mihomo.uptime : '0s'}</strong>
              </div>
              <div style={{ fontSize: '13px', color: '#8c8c8c' }}>
                Version: <strong style={{ color: '#595959' }}>{data ? data.mihomo.version : 'Unknown'}</strong>
              </div>
            </div>

            <Space size="middle">
              {!isMihomoRunning ? (
                <Button
                  type="primary"
                  icon={<PlayCircleOutlined />}
                  loading={loading}
                  onClick={() => handleAction('start')}
                >
                  Start
                </Button>
              ) : (
                <>
                  <Button
                    type="primary"
                    danger
                    icon={<StopOutlined />}
                    loading={loading}
                    onClick={() => handleAction('stop')}
                  >
                    Stop
                  </Button>
                  <Button
                    icon={<ReloadOutlined />}
                    loading={loading}
                    onClick={() => handleAction('restart')}
                  >
                    Restart
                  </Button>
                </>
              )}
            </Space>
          </Card>
        </Col>

        {/* CPU Progress */}
        <Col xs={12} md={6} lg={8}>
          <Card
            title={
              <Space>
                <DashboardOutlined />
                <span>CPU Allocation</span>
              </Space>
            }
            bordered={false}
            style={{ height: '100%', textAlign: 'center' }}
          >
            <Progress
              type="circle"
              percent={data ? data.system.cpu.percent : 0}
              width={90}
              strokeColor={{
                '0%': '#108ee9',
                '100%': '#87d068',
              }}
            />
            <div style={{ marginTop: 8, fontWeight: 'bold' }}>CPU Usage</div>
          </Card>
        </Col>

        {/* Memory Progress */}
        <Col xs={12} md={6} lg={8}>
          <Card
            title={
              <Space>
                <DashboardOutlined />
                <span>Memory (RAM)</span>
              </Space>
            }
            bordered={false}
            style={{ height: '100%', textAlign: 'center' }}
          >
            <Progress
              type="circle"
              percent={data ? data.system.memory.percent : 0}
              width={90}
              status="normal"
            />
            <div style={{ marginTop: 8, fontWeight: 'bold' }}>Memory Usage</div>
            <div style={{ fontSize: '12px', color: '#8c8c8c' }}>
              {data ? `${data.system.memory.used.toFixed(0)} / ${data.system.memory.total.toFixed(0)} MB` : '0 / 0 MB'}
            </div>
          </Card>
        </Col>
      </Row>

      {/* Row 2: Disk Progress, Network Rates, Listener Stats */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        {/* Disk Progress */}
        <Col xs={24} md={12} lg={8}>
          <Card
            title={
              <Space>
                <DashboardOutlined />
                <span>Disk Storage</span>
              </Space>
            }
            bordered={false}
            style={{ height: '100%', textAlign: 'center' }}
          >
            <Progress
              type="circle"
              percent={data ? data.system.disk.percent : 0}
              width={90}
              status="normal"
            />
            <div style={{ marginTop: 8, fontWeight: 'bold' }}>Disk Storage</div>
            <div style={{ fontSize: '12px', color: '#8c8c8c' }}>
              {data ? `${data.system.disk.used.toFixed(1)} / ${data.system.disk.total.toFixed(1)} GB` : '0 / 0 GB'}
            </div>
          </Card>
        </Col>

        {/* Network Rates */}
        <Col xs={12} md={6} lg={8}>
          <Card
            title={
              <Space>
                <CloudUploadOutlined />
                <span>Real-time Network Rate</span>
              </Space>
            }
            bordered={false}
            style={{ height: '100%' }}
          >
            <Space direction="vertical" size="middle" style={{ width: '100%', marginTop: 8 }}>
              <Statistic
                title="Upload Speed"
                value={data ? formatRate(data.system.network.upload) : '0 B/s'}
                prefix={<CloudUploadOutlined style={{ color: '#1890ff' }} />}
                valueStyle={{ fontSize: '18px', fontWeight: 'bold' }}
              />
              <Statistic
                title="Download Speed"
                value={data ? formatRate(data.system.network.download) : '0 B/s'}
                prefix={<CloudDownloadOutlined style={{ color: '#52c41a' }} />}
                valueStyle={{ fontSize: '18px', fontWeight: 'bold' }}
              />
            </Space>
          </Card>
        </Col>

        {/* Listener Stats */}
        <Col xs={12} md={6} lg={8}>
          <Card
            title={
              <Space>
                <CloudServerOutlined />
                <span>Listener Distribution</span>
              </Space>
            }
            bordered={false}
            style={{ height: '100%' }}
          >
            <Row gutter={16} style={{ marginTop: 8 }}>
              <Col span={24} style={{ marginBottom: 12 }}>
                <Statistic
                  title="Total Inbound Nodes"
                  value={data ? data.listeners.total : 0}
                  valueStyle={{ fontWeight: 'bold', fontSize: '24px' }}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="Active"
                  value={data ? data.listeners.enabled : 0}
                  valueStyle={{ color: '#3f8600', fontSize: '18px' }}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="Disabled"
                  value={data ? data.listeners.disabled : 0}
                  valueStyle={{ color: '#cf1322', fontSize: '18px' }}
                />
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
