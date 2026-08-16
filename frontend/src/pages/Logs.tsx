import React, { useEffect, useState } from 'react';
import { Card, List, Tag } from 'antd';
import { fetchLogs } from '../api/system';

const Logs: React.FC = () => {
  const [logs, setLogs] = useState<Awaited<ReturnType<typeof fetchLogs>>>([]);
  useEffect(() => {
    fetchLogs().then(setLogs).catch(() => setLogs([]));
  }, []);
  return (
    <div>
      <h2>Logs</h2>
      <Card>
        <List
          size="small"
          dataSource={logs}
          renderItem={(item) => (
            <List.Item>
              <Tag color="blue">{item.level?.toUpperCase() || 'INFO'}</Tag>
              <code style={{ fontSize: 12 }}>{item.payload}</code>
            </List.Item>
          )}
        />
      </Card>
    </div>
  );
};

export default Logs;
