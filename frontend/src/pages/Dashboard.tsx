import { useState, useEffect } from 'react';
import { Alert, Button, Card, Col, Progress, Row, Space, Statistic, Tag, Typography, message } from 'antd';
import { PlayCircleOutlined, StopOutlined, ReloadOutlined } from '@ant-design/icons';
import { apiRequest } from '../api/request';

const { Title, Paragraph } = Typography;

type Data = {
  mihomo: { running: boolean; version: string; pid: number; uptime: string };
  system: { cpu: { percent: number }; memory: { percent: number; used: number; total: number }; disk: { percent: number } };
  listeners: { total: number; enabled: number };
  traffic: { activeConnections: number; uploadRate: number; downloadRate: number };
};

const fmt = (n: number) => {
  if (!n) return '0 B/s';
  if (n < 1024) return `${Math.round(n)} B/s`;
  if (n < 1048576) return `${(n / 1024).toFixed(1)} KB/s`;
  return `${(n / 1048576).toFixed(1)} MB/s`;
};

export default function Dashboard() {
  const [data, setData] = useState<Data | null>(null);
  const [busy, setBusy] = useState(false);

  const load = async () => {
    try {
      setData(await apiRequest<Data>('/dashboard'));
    } catch (e: any) {
      message.error(e.message || '后端服务不可用');
    }
  };

  useEffect(() => {
    load();
    const id = window.setInterval(load, 10000);
    return () => clearInterval(id);
  }, []);

  const act = async (a: 'start' | 'stop' | 'restart') => {
    setBusy(true);
    try {
      await apiRequest(`/mihomo/${a}`, { method: 'POST' });
      message.success(`${a === 'start' ? '启动' : a === 'stop' ? '停止' : '重启'}成功`);
      load();
    } catch (e: any) {
      message.error(e.message || '操作失败');
    } finally {
      setBusy(false);
    }
  };

  return (
    <>
      <Title level={2}>仪表盘</Title>
      <Paragraph>查看 Mihomo Core、代理节点与服务器运行状态。</Paragraph>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={10}>
          <Card title="Mihomo 核心" extra={data?.mihomo.running ? <Tag color="success">运行中</Tag> : <Tag>已停止</Tag>}>
            <Space direction="vertical" style={{ width: '100%' }}>
              <Statistic title="版本" value={data?.mihomo.version || '未知'} />
              <div>PID：{data?.mihomo.pid || '-'}　运行时间：{data?.mihomo.uptime || '-'}</div>
              <Space>
                <Button type="primary" icon={<PlayCircleOutlined />} disabled={!!data?.mihomo.running} loading={busy} onClick={() => act('start')}>
                  启动
                </Button>
                <Button danger icon={<StopOutlined />} disabled={!data?.mihomo.running} loading={busy} onClick={() => act('stop')}>
                  停止
                </Button>
                <Button icon={<ReloadOutlined />} loading={busy} onClick={() => act('restart')}>
                  重启
                </Button>
              </Space>
            </Space>
          </Card>
        </Col>
        <Col xs={12} lg={5}>
          <Card title="CPU">
            <Progress type="circle" percent={data?.system.cpu.percent || 0} />
          </Card>
        </Col>
        <Col xs={12} lg={5}>
          <Card title="内存">
            <Progress type="circle" percent={data?.system.memory.percent || 0} />
          </Card>
        </Col>
        <Col xs={24} lg={4}>
          <Card title="代理节点">
            <Statistic value={data?.listeners.enabled || 0} suffix="个" />
          </Card>
        </Col>
        <Col xs={24}>
          <Card title="网络">
            <Row gutter={16}>
              <Col span={8}>
                <Statistic title="活动连接" value={data?.traffic.activeConnections || 0} />
              </Col>
              <Col span={8}>
                <Statistic title="上传速率" value={fmt(data?.traffic.uploadRate || 0)} />
              </Col>
              <Col span={8}>
                <Statistic title="下载速率" value={fmt(data?.traffic.downloadRate || 0)} />
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>
      <Alert style={{ marginTop: 16 }} type="info" showIcon message="控制台定位" description="3m-ui 专注 Mihomo Core 管理，不包含机场销售、套餐或终端用户业务。" />
    </>
  );
}
