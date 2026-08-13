import { useState, useEffect } from 'react';
import { Alert, Button, Card, Descriptions, Space, Spin, Tag, message } from 'antd';
import { apiRequest } from '../api/request';
import { useI18n } from '../i18n';

type Status = { running: boolean; version: string; pid: number; uptime: string };

export default function Core() {
  const { t } = useI18n();
  const [status, setStatus] = useState<Status | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);

  const load = async () => {
    try {
      setStatus(await apiRequest<Status>('/mihomo/status'));
    } catch (e: any) {
      message.error(e.message || t('core.unavailable'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    const id = window.setInterval(load, 5000);
    return () => window.clearInterval(id);
  }, []);

  const action = async (path: string, successKey: 'started' | 'stopped' | 'restarted') => {
    setBusy(true);
    try {
      await apiRequest(`/mihomo/${path}`, { method: 'POST' });
      message.success(t(`core.${successKey}`));
      await load();
    } catch (e: any) {
      message.error(e.message || t('core.operationFailed'));
    } finally {
      setBusy(false);
    }
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card
        className="core-card"
        title={t('core.title')}
        extra={
          <Space className="core-actions" wrap size={[8, 8]}>
            <Button type="primary" disabled={busy || !!status?.running} onClick={() => action('start', 'started')}>
              {t('core.start')}
            </Button>
            <Button danger disabled={busy || !status?.running} onClick={() => action('stop', 'stopped')}>
              {t('core.stop')}
            </Button>
            <Button disabled={busy} onClick={() => action('restart', 'restarted')}>
              {t('core.restart')}
            </Button>
          </Space>
        }
      >
        {loading ? (
          <div className="core-loading"><Spin /></div>
        ) : (
          <Descriptions bordered column={{ xs: 1, sm: 2 }}>
            <Descriptions.Item label={t('core.status')}>
              {status?.running ? <Tag color="success">{t('core.running')}</Tag> : <Tag>{t('core.stoppedStatus')}</Tag>}
            </Descriptions.Item>
            <Descriptions.Item label={t('core.version')}>{status?.version || '-'}</Descriptions.Item>
            <Descriptions.Item label="PID">{status?.pid || '-'}</Descriptions.Item>
            <Descriptions.Item label={t('core.uptime')}>{status?.uptime || '-'}</Descriptions.Item>
          </Descriptions>
        )}
      </Card>
      <Alert
        type="info"
        showIcon
        message={t('core.management')}
        description={t('core.description')}
      />
    </Space>
  );
}
