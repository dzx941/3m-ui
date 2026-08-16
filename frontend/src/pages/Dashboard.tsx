import React, { useEffect, useState } from 'react';
import { Row, Col, Card, Statistic, Button, Space, Tag, message } from 'antd';
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  ReloadOutlined,
  NodeIndexOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import { fetchDashboard, startMihomo, stopMihomo, restartMihomo } from '../api/system';
import { fetchListeners } from '../api/nodes';

const Dashboard: React.FC = () => {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [busy, setBusy] = useState(false);

  const load = async () => {
    setLoading(true);
    try {
      const [dash, nodes] = await Promise.all([fetchDashboard(), fetchListeners()]);
      const total = nodes.length;
      const enabled = nodes.filter((n: any) => n.enabled).length;
      setData({ ...dash, listeners: { total, enabled, disabled: total - enabled } });
    } catch (e: any) {
      message.error(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    const id = setInterval(load, 10000);
    return () => clearInterval(id);
  }, []);

  const act = async (action: 'start' | 'stop' | 'restart') => {
    setBusy(true);
    try {
      if (action === 'start') await startMihomo();
      else if (action === 'stop') await stopMihomo();
      else await restartMihomo();
      message.success(`${action === 'start' ? '启动' : action === 'stop' ? '停止' : '重启'}成功`);
      load();
    } catch (e: any) {
      message.error(e.message);
    } finally {
      setBusy(false);
    }
  };

  const mihomo = data?.mihomo;
  const running = mihomo?.running;

  return (
    <div>
      <h2>仪表盘</h2>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title="核心状态"
              value={running ? '运行中' : '已停止'}
              valueStyle={{ color: running ? '#52c41a' : '#ff4d4f' }}
            />
            <div style={{ marginTop: 8 }}>
              <Tag>{mihomo?.version || '-'}</Tag>
              <Tag icon={<ClockCircleOutlined />}>{mihomo?.uptime || '-'}</Tag>
            </div>
            <Space style={{ marginTop: 16 }}>
              {!running && (
                <Button icon={<PlayCircleOutlined />} onClick={() => act('start')} loading={busy}>
                  启动
                </Button>
              )}
              {running && (
                <Button icon={<PauseCircleOutlined />} danger onClick={() => act('stop')} loading={busy}>
                  停止
                </Button>
              )}
              <Button icon={<ReloadOutlined />} onClick={() => act('restart')} loading={busy}>
                重启
              </Button>
            </Space>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title="监听器"
              value={data?.listeners?.enabled || 0}
              suffix={`/ ${data?.listeners?.total || 0}`}
              prefix={<NodeIndexOutlined />}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic title="CPU" value={data?.cpu?.percent || 0} suffix="%" precision={1} />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic title="内存" value={data?.memory?.percent || 0} suffix="%" precision={1} />
            <div style={{ fontSize: 12, color: '#888' }}>
              {((data?.memory?.used || 0) / 1024).toFixed(1)} GB / {((data?.memory?.total || 0) / 1024).toFixed(1)} GB
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
