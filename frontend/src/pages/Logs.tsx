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
    <div className="logs-page">
      <div className="page-heading logs-heading">
        <Title level={2}>{t('logs.title')}</Title>
        <Paragraph className="logs-subtitle">{t('logs.subtitle')}</Paragraph>
      </div>

      <div className="logs-toolbar">
        <Space wrap size={8}>
          <Button onClick={() => setLogs([])}>{t('logs.clear')}</Button>
          <Button type="primary" onClick={() => void fetchLogs()} loading={loading}>{t('logs.refresh')}</Button>
        </Space>
      </div>

      <Card className="logs-card">
        <div className="logs-viewer" role="log" aria-live="polite" aria-busy={loading}>
          {logs.length === 0 ? (
            <div className="logs-empty">{t('logs.empty')}</div>
          ) : (
            logs.map((log, i) => (
              <div className="logs-line" key={`${log.timestamp}-${i}`}>
                <span className="logs-time">
                  [{dayjs(log.timestamp).format('YYYY-MM-DD HH:mm:ss')}]
                </span>
                <span className={`logs-level logs-level-${log.level.toLowerCase()}`}>
                  [{log.level}]
                </span>
                <span className="logs-payload">{log.payload}</span>
              </div>
            ))
          )}
        </div>
      </Card>
    </div>
  );
};

export default Logs;
