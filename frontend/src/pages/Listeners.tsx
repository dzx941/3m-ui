import React, { useEffect, useState } from 'react';
import {
  Typography,
  Table,
  Button,
  Space,
  Tag,
  Popconfirm,
  Modal,
  Form,
  Input,
  InputNumber,
  Select,
  Switch,
  message,
  Row,
  Col,
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons';
import { apiRequest } from '../api/request';
import { ProtocolForm } from '../components/protocols';
import { LISTENER_PROTOCOLS, type ListenerProtocol } from '../components/protocols/types';

const { Title, Paragraph } = Typography;

interface NodeRecord {
  ID: number;
  name: string;
  protocol: string;
  port: number;
  bind_address: string;
  tls: boolean;
  udp: boolean;
  enabled: boolean;
  config: string;
  status: string;
}

interface ConnectionView {
  id: string;
  listener_id: number | null;
  listener_name: string;
  proxy_user_id: number | null;
  username: string;
  network: string;
  host: string;
  source_ip: string;
  destination_ip: string;
  destination_port: string;
  upload: number;
  download: number;
}

interface ListenerTrafficStats {
  connections: number;
  upload: number;
  download: number;
}

const protocolLabels: Record<ListenerProtocol, string> = {
  socks: 'SOCKS5',
  http: 'HTTP',
  tproxy: 'TPROXY',
  redir: 'REDIR',
  mixed: 'Mixed',
  tunnel: 'Tunnel',
  tun: 'TUN',
  shadowsocks: 'Shadowsocks',
  snell: 'Snell',
  vmess: 'VMess',
  vless: 'VLESS',
  trojan: 'Trojan',
  hysteria2: 'Hysteria 2',
  'hysteria2-realm': 'Hysteria 2 Realm',
  tuic: 'TUIC',
  shadowquic: 'ShadowQuic',
  anytls: 'AnyTLS',
  mieru: 'Mieru',
  sudoku: 'Sudoku',
  trusttunnel: 'TrustTunnel',
};

const formatBytes = (bytes: number): string => {
  if (!bytes) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = bytes;
  let i = 0;
  while (value >= 1024 && i < units.length - 1) {
    value /= 1024;
    i += 1;
  }
  return `${value.toFixed(value >= 10 || i === 0 ? 0 : 1)} ${units[i]}`;
};

const ListenersPage: React.FC = () => {
  const [data, setData] = useState<NodeRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<NodeRecord | null>(null);
  const [form] = Form.useForm();
  const [trafficByListener, setTrafficByListener] = useState<Record<number, ListenerTrafficStats>>({});
  const selectedProtocol = Form.useWatch('protocol', form) as ListenerProtocol | undefined;

  const fetchNodes = async () => {
    setLoading(true);
    try {
      const list = await apiRequest<NodeRecord[]>('/nodes');
      setData(list || []);
    } catch {
      message.error('Backend service unreachable.');
    } finally {
      setLoading(false);
    }
  };

  const fetchTraffic = async () => {
    try {
      const body = await apiRequest<{ connections: ConnectionView[] }>('/traffic/connections');
      const grouped: Record<number, ListenerTrafficStats> = {};
      for (const conn of body.connections || []) {
        if (conn.listener_id === null) continue;
        const stats = grouped[conn.listener_id] || { connections: 0, upload: 0, download: 0 };
        stats.connections += 1;
        stats.upload += conn.upload;
        stats.download += conn.download;
        grouped[conn.listener_id] = stats;
      }
      setTrafficByListener(grouped);
    } catch {
      // Traffic is non-critical.
    }
  };

  useEffect(() => {
    void fetchNodes();
    void fetchTraffic();
    const interval = setInterval(() => void fetchTraffic(), 10000);
    return () => clearInterval(interval);
  }, []);

  const handleOpenAdd = () => {
    setEditingRecord(null);
    form.resetFields();
    form.setFieldsValue({
      bind_address: '0.0.0.0',
      port: 10086,
      protocol: 'shadowsocks',
      tls: false,
      udp: true,
      enabled: true,
      protocolConfig: {},
    });
    setModalOpen(true);
  };

  const handleOpenEdit = (record: NodeRecord) => {
    setEditingRecord(record);
    form.resetFields();
    let protocolConfig: Record<string, unknown> = {};
    try {
      protocolConfig = JSON.parse(record.config || '{}');
      for (const key of ['reality-config', 'plugin-opts', 'restls-opts', 'jls-opts', 'tls', 'users', 'xhttp-opts']) {
        if (protocolConfig[key] && typeof protocolConfig[key] !== 'string') {
          protocolConfig[key] = JSON.stringify(protocolConfig[key], null, 2);
        }
      }
    } catch {
      // Keep empty config for malformed legacy data.
    }
    form.setFieldsValue({ ...record, protocolConfig });
    setModalOpen(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await apiRequest(`/nodes/${id}`, { method: 'DELETE' });
      message.success('Node deleted successfully.');
      void fetchNodes();
    } catch {
      message.error('Network connection error.');
    }
  };

  const handleToggleEnabled = async (record: NodeRecord, checked: boolean) => {
    try {
      await apiRequest(`/nodes/${record.ID}`, {
        method: 'PUT',
        body: JSON.stringify({ ...record, enabled: checked, status: checked ? 'active' : 'inactive' }),
      });
      message.success(`Node ${checked ? 'enabled' : 'disabled'} successfully.`);
      void fetchNodes();
    } catch {
      message.error('Network connection error.');
    }
  };

  const handleReload = async (id: number) => {
    try {
      await apiRequest(`/nodes/${id}/reload`, { method: 'POST' });
      message.success('Mihomo configuration reloaded.');
    } catch {
      message.error('Network connection error.');
    }
  };

  const handleFormSubmit = async () => {
    try {
      const values = await form.validateFields();
      const configObj: Record<string, unknown> = { ...(values.protocolConfig || {}) };
      for (const key of ['reality-config', 'plugin-opts', 'restls-opts', 'jls-opts', 'tls', 'users', 'xhttp-opts']) {
        const value = configObj[key];
        if (typeof value === 'string' && value.trim() !== '') {
          try {
            configObj[key] = JSON.parse(value);
          } catch {
            message.error(`${key} must contain valid JSON.`);
            return;
          }
        }
      }
      const payload = {
        ...values,
        protocol: values.protocol,
        type: values.protocol,
        listen: values.bind_address,
        config: JSON.stringify(configObj),
        status: values.enabled ? 'active' : 'inactive',
      };
      const method = editingRecord ? 'PUT' : 'POST';
      const url = editingRecord ? `/nodes/${editingRecord.ID}` : '/nodes';
      await apiRequest(url, { method, body: JSON.stringify(payload) });
      message.success(`Node ${editingRecord ? 'updated' : 'created'} successfully.`);
      setModalOpen(false);
      void fetchNodes();
    } catch {
      // Validation failed.
    }
  };

  const columns = [
    { title: 'ID', dataIndex: 'ID', key: 'ID', width: 60 },
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Protocol', dataIndex: 'protocol', key: 'protocol', render: (proto: string) => <Tag color="blue">{protocolLabels[proto as ListenerProtocol] || proto}</Tag> },
    { title: 'Port', dataIndex: 'port', key: 'port' },
    { title: 'TLS', dataIndex: 'tls', key: 'tls', render: (tls: boolean) => (tls ? <Tag color="green">TLS</Tag> : <Tag>Plain</Tag>) },
    {
      title: 'Connections',
      key: 'connections',
      render: (_: unknown, record: NodeRecord) => {
        const stats = trafficByListener[record.ID];
        return <Tag color={stats?.connections ? 'blue' : 'default'}>{stats?.connections || 0}</Tag>;
      },
    },
    {
      title: 'Traffic',
      key: 'traffic',
      render: (_: unknown, record: NodeRecord) => {
        const stats = trafficByListener[record.ID];
        return <span>↑ {formatBytes(stats?.upload || 0)} &nbsp; ↓ {formatBytes(stats?.download || 0)}</span>;
      },
    },
    {
      title: 'Status',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean, record: NodeRecord) => <Switch checked={enabled} onChange={(checked) => void handleToggleEnabled(record, checked)} checkedChildren="On" unCheckedChildren="Off" />,
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, record: NodeRecord) => (
        <Space size="middle">
          <Button type="text" icon={<ReloadOutlined style={{ color: '#52c41a' }} />} onClick={() => void handleReload(record.ID)} title="Hot Reload" />
          <Button type="text" icon={<EditOutlined style={{ color: '#1890ff' }} />} onClick={() => handleOpenEdit(record)} title="Edit" />
          <Popconfirm title="Are you sure you want to delete this node?" onConfirm={() => void handleDelete(record.ID)} okText="Yes" cancelText="No">
            <Button type="text" danger icon={<DeleteOutlined />} title="Delete" />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div>
          <Title level={2} style={{ margin: 0 }}>Mihomo Server Listeners</Title>
          <Paragraph style={{ margin: 0 }}>Configure every listener type supported by the installed Mihomo core, including protocol-specific options.</Paragraph>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleOpenAdd}>Add Listener</Button>
      </div>

      <Table dataSource={data} columns={columns} rowKey="ID" loading={loading} scroll={{ x: 'max-content' }} locale={{ emptyText: 'No listeners found. Create one to get started.' }} />

      <Modal title={editingRecord ? 'Edit Listener' : 'Add Listener'} open={modalOpen} onOk={() => void handleFormSubmit()} onCancel={() => setModalOpen(false)} destroyOnClose width={760}>
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="name" label="Name" rules={[{ required: true, message: 'Please input listener name' }]}>
            <Input placeholder="e.g. vless-reality" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="protocol" label="Listener Protocol" rules={[{ required: true, message: 'Select listener protocol' }]}>
                <Select
                  showSearch
                  optionFilterProp="label"
                  options={LISTENER_PROTOCOLS.map((protocol) => ({ value: protocol, label: protocolLabels[protocol] }))}
                  onChange={() => form.setFieldsValue({ protocolConfig: {} })}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="port" label="Port" rules={[{ required: true, message: 'Please enter port number' }]}>
                <InputNumber min={1} max={65535} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="bind_address" label="Listen Address" rules={[{ required: true, message: 'Please enter listen address' }]}>
                <Input placeholder="0.0.0.0" />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="tls" label="TLS" valuePropName="checked">
                <Switch checkedChildren="Enabled" unCheckedChildren="Disabled" />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="udp" label="UDP" valuePropName="checked">
                <Switch checkedChildren="Enabled" unCheckedChildren="Disabled" />
              </Form.Item>
            </Col>
          </Row>

          <ProtocolForm protocol={selectedProtocol} />

          <Form.Item name="enabled" label="Enabled" valuePropName="checked">
            <Switch checkedChildren="On" unCheckedChildren="Off" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ListenersPage;
