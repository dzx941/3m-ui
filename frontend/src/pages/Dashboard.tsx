import { useState, useEffect } from 'react';
import { Alert, Button, Card, Col, Progress, Row, Space, Statistic, Tag, Typography, message } from 'antd';
import { PlayCircleOutlined, StopOutlined, ReloadOutlined } from '@ant-design/icons';
import { apiRequest } from '../api/request';
import { useI18n } from '../i18n';

const { Title, Paragraph } = Typography;

type Node = {
  id: number;
  name: string;
  protocol: string;
  enabled: boolean;
};

type Data = {
  mihomo: { running: boolean; version: string; pid: number; uptime: string };
  system: { cpu: { percent: number }; memory: { percent: number; used: number; total: number }; disk: { percent: number } };
  listeners: { total: number; enabled: number; disabled?: number };
  traffic: { activeConnections: number; uploadRate: number; downloadRate: number };
};

const fmt = (n: number) => {
  if (!n) return '0 B/s';
  if (n < 1024) return `${Math.round(n)} B/s`;
  if (n < 1048576) return `${(n / 1024).toFixed(1)} KB/s`;
  return `${(n / 1048576).toFixed(1)} MB/s`;
};

export default function Dashboard() {
  const { t } = useI18n();
  const [data, setData] = useState<Data | null>(null);
  const [busy, setBusy] = useState(false);

  const load = async () => {
    try {
      const [dashboard, nodes] = await Promise.all([
        apiRequest<Data>('/dashboard'),
        apiRequest<Node[]>('/nodes'),
      ]);

      // /nodes is the authoritative source for the UI node count.
      const total = Array.isArray(nodes) ? nodes.length : 0;
      const enabled = Array.isArray(nodes) ? nodes.filter((node) => node.enabled).length : 0;

      setData({
        ...dashboard,
        listeners: { total, enabled, disabled: Math.max(0, total - enabled) },
      });
    } catch (e: any) {
      message.error(e.message || t('dashboard.unavailable'));
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
      message.success(t(`dashboard.${a === 'start' ? 'started' : a === 'stop' ? 'stopped' : 'restarted'}`));
      load();
    } catch (e: any) {
      message.error(e.message || t('dashboard.operationFailed'));
    } finally {
      setBusy(false);
    }
  };

  return (
    <>
      <div className="page-heading">
        <Title level={2}>{t('dashboard.title')}</Title>
        <Paragraph>{t('dashboard.subtitle')}</Paragraph>
      </div>

      <Row gutter={[12, 12]}>
        <Col xs={24} lg={10}>
          <Card title={t('core.title')} extra={data?.mihomo.running ? <Tag color="success">{t('dashboard.running')}</Tag> : <Tag>{t('dashboard.stoppedStatus')}</Tag>}>
            <Space direction="vertical" size={14} style={{ width: '100%' }}>
              <Statistic title={t('dashboard.version')} value={data?.mihomo.version || '-'} />
              <div>PID: {data?.mihomo.pid || '-'}　{t('dashboard.uptime')}: {data?.mihomo.uptime || '-'}</div>
              <Space wrap>
                <Button type="primary" icon={<PlayCircleOutlined />} disabled={!!data?.mihomo.running} loading={busy} onClick={() => act('start')}>
                  {t('dashboard.start')}
                </Button>
                <Button danger icon={<StopOutlined />} disabled={!data?.mihomo.running} loading={busy} onClick={() => act('stop')}>
                  {t('dashboard.stop')}
                </Button>
                <Button icon={<ReloadOutlined />} loading={busy} onClick={() => act('restart')}>
                  {t('dashboard.restart')}
                </Button>
              </Space>
            </Space>
          </Card>
        </Col>
        <Col xs={12} sm={6} lg={5}>
          <Card title="CPU">
            <div className="dashboard-progress"><Progress type="circle" percent={data?.system.cpu.percent || 0} /></div>
          </Card>
        </Col>
        <Col xs={12} sm={6} lg={5}>
          <Card title={t('dashboard.memory')}>
            <div className="dashboard-progress"><Progress type="circle" percent={data?.system.memory.percent || 0} /></div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card title={t('dashboard.nodes')}>
            <Statistic value={data?.listeners.total || 0} suffix={t('dashboard.countUnit')} />
            <div className="dashboard-muted">{data?.listeners.enabled || 0} {t('dashboard.running')}</div>
          </Card>
        </Col>
        <Col xs={24}>
          <Card title={t('dashboard.network')}>
            <Row gutter={[12, 12]}>
              <Col xs={24} sm={8}>
                <Statistic title={t('dashboard.connections')} value={data?.traffic.activeConnections || 0} />
              </Col>
              <Col xs={24} sm={8}>
                <Statistic title={t('dashboard.upload')} value={fmt(data?.traffic.uploadRate || 0)} />
              </Col>
              <Col xs={24} sm={8}>
                <Statistic title={t('dashboard.download')} value={fmt(data?.traffic.downloadRate || 0)} />
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>
      <Alert className="dashboard-alert" type="info" showIcon message={t('dashboard.consoleTitle')} description={t('dashboard.consoleDescription')} />
    </>
  );
}
