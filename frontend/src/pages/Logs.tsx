import React, { useState, useEffect } from 'react';
import { Typography, Card, Space, Button, message } from 'antd';
import { apiRequest } from '../api/request';
import dayjs from 'dayjs';
import { useI18n } from '../i18n';

const { Title, Paragraph } = Typography;

interface LogLine {
  timestamp: string;
  level: string;
  payload: string;
}

const Logs: React.FC = () => {
  const { t } = useI18n();
  const [logs, setLogs] = useState<LogLine[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchLogs = async () => {
    try {
      const data = await apiRequest<LogLine[]>('/mihomo/logs');
      setLogs(data || []);
    } catch (e: any) {
      message.error(e.message || t('logs.loadFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void fetchLogs();
    const id = setInterval(() => void fetchLogs(), 5000);
    return () => clearInterval(id);
  }, []);

  return (
    <div>
      <Title level={2}>{t('logs.title')}</Title>
      <Paragraph>{t('logs.subtitle')}</Paragraph>
      <Space style={{ marginBottom: 16 }}>
        <Button onClick={() => setLogs([])}>{t('logs.clear')}</Button>
        <Button type="primary" onClick={() => void fetchLogs()} loading={loading}>{t('logs.refresh')}</Button>
      </Space>
      <Card
        style={{
          backgroundColor: '#001529',
          color: '#ffffff',
          fontFamily: 'monospace',
          minHeight: '300px',
          borderRadius: '4px',
        }}
        styles={{
          body: {
            backgroundColor: '#001529',
            color: '#ffffff',
            padding: '16px',
            overflowY: 'auto',
            maxHeight: '500px',
          }
        }}
      >
        {logs.length === 0 ? (
          <div style={{ color: '#888' }}>{t('logs.empty')}</div>
        ) : (
          logs.map((log, i) => (
            <div key={i} style={{ marginBottom: 4 }}>
              <span style={{ color: '#00ff00', marginRight: 8 }}>
                [{dayjs(log.timestamp).format('YYYY-MM-DD HH:mm:ss')}]
              </span>
              <span style={{ color: log.level === 'error' ? '#ff4d4f' : '#1890ff', marginRight: 8, textTransform: 'uppercase' }}>
                [{log.level}]
              </span>
              <span>{log.payload}</span>
            </div>
          ))
        )}
      </Card>
    </div>
  );
};

export default Logs;
