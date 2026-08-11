import { useState, useEffect } from 'react';
import { Alert, Button, Card, Descriptions, Space, Spin, Tag, message } from 'antd';
import { apiRequest } from '../api/request';

type Status = { running: boolean; version: string; pid: number; uptime: string };

export default function Core() {
  const [status, setStatus] = useState<Status | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);

  const load = async () => {
    try {
      setStatus(await apiRequest<Status>('/mihomo/status'));
    } catch (e: any) {
      message.error(e.message || '无法获取 Mihomo 状态');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    const id = window.setInterval(load, 5000);
    return () => window.clearInterval(id);
  }, []);

  const action = async (path: string, label: string) => {
    setBusy(true);
    try {
      await apiRequest(`/mihomo/${path}`, { method: 'POST' });
      message.success(label + '成功');
      await load();
    } catch (e: any) {
      message.error(e.message || label + '失败');
    } finally {
      setBusy(false);
    }
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card
        title="Mihomo 核心"
        extra={
          <Space>
            <Button type="primary" disabled={busy || !!status?.running} onClick={() => action('start', '启动')}>
              启动
            </Button>
            <Button danger disabled={busy || !status?.running} onClick={() => action('stop', '停止')}>
              停止
            </Button>
            <Button disabled={busy} onClick={() => action('restart', '重启')}>
              重启
            </Button>
          </Space>
        }
      >
        {loading ? (
          <Spin />
        ) : (
          <Descriptions bordered column={2}>
            <Descriptions.Item label="运行状态">
              {status?.running ? <Tag color="success">运行中</Tag> : <Tag>已停止</Tag>}
            </Descriptions.Item>
            <Descriptions.Item label="版本">{status?.version || '未知'}</Descriptions.Item>
            <Descriptions.Item label="PID">{status?.pid || '-'}</Descriptions.Item>
            <Descriptions.Item label="运行时间">{status?.uptime || '-'}</Descriptions.Item>
          </Descriptions>
        )}
      </Card>
      <Alert
        type="info"
        showIcon
        message="核心管理"
        description="这里管理服务器上的 Mihomo Core。修改配置后，请在配置管理中生成配置，再重载或重启核心。"
      />
    </Space>
  );
}
