import React, { useState, useEffect } from 'react';
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
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, CopyOutlined } from '@ant-design/icons';
import { apiRequest } from '../api/request';
import { ProtocolForm } from '../components/protocols';

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

interface ClientAccess {
  id: number;
  name: string;
  type: 'listener';
  listener_id: number;
  mihomo_link: string;
  clash_link: string;
  singbox_link: string;
  shadowrocket_link: string;
}

const formatBytes = (bytes: number): string => {
  if (!bytes) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = bytes;
  let i = 0;
  while (value >= 1024 && i < units.length - 1) {
    value /= 1024;
    i++;
  }
  return `${value.toFixed(value >= 10 || i === 0 ? 0 : 1)} ${units[i]}`;
};

const ListenersPage: React.FC = () => {
  const [data, setData] = useState<NodeRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<NodeRecord | null>(null);
  const [clientAccess, setClientAccess] = useState<ClientAccess | null>(null);
  const [clientModalOpen, setClientModalOpen] = useState(false);
  const [form] = Form.useForm();
  const [trafficByListener, setTrafficByListener] = useState<Record<number, ListenerTrafficStats>>({});
  const selectedProtocol = Form.useWatch('protocol', form);

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
      // Traffic is non-critical UI data.
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
    let flow = '';
    try {
      const parsed = JSON.parse(record.config || '{}');
      if (parsed.flow) flow = parsed.flow;
      protocolConfig = { ...parsed };
    } catch {
      // Keep the form usable for legacy malformed records.
    }
    form.setFieldsValue({ ...record, flow, protocolConfig });
    setModalOpen(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await apiRequest(`/nodes/${id}`, { method: 'DELETE' });
      message.success('Listener deleted successfully.');
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
      message.success(`Listener ${checked ? 'enabled' : 'disabled'} successfully.`);
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

  const generateClientAccess = async (id: number) => {
    try {
      const access = await apiRequest<ClientAccess>(`/nodes/${id}/client-access`, { method: 'POST' });
      setClientAccess(access);
      setClientModalOpen(true);
    } catch (error: unknown) {
      message.error(error instanceof Error ? error.message : 'Failed to generate client configuration.');
    }
  };

  const handleFormSubmit = async () => {
    try {
      const values = await form.validateFields();
      const configObj: Record<string, unknown> = values.protocolConfig || {};
      if (values.flow) configObj.flow = values.flow;
      const payload = {
        ...values,
        config: JSON.stringify(configObj),
        status: values.enabled ? 'active' : 'inactive',
      };
      const method = editingRecord ? 'PUT' : 'POST';
      const url = editingRecord ? `/nodes/${editingRecord.ID}` : '/nodes';
      const created = await apiRequest<NodeRecord>(url, { method, body: JSON.stringify(payload) });

      message.success(`Listener ${editingRecord ? 'updated' : 'created'} successfully.`);
      setModalOpen(false);
      void fetchNodes();

      // A new listener immediately receives a listener-bound client access token.
      if (!editingRecord && created?.ID) {
        await generateClientAccess(created.ID);
      }
    } catch {
      // Validation or API failure is already surfaced by the form/request layer.
    }
  };

  const copyLink = async (link: string) => {
    try {
      await navigator.clipboard.writeText(link);
      message.success('Link copied.');
    } catch {
      message.error('Failed to copy link.');
    }
  };

  const columns = [
    { title: 'ID', dataIndex: 'ID', key: 'ID', width: 60 },
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Listener Type', dataIndex: 'protocol', key: 'protocol', render: (proto: string) => <Tag color="blue">{proto}</Tag> },
    { title: 'Port', dataIndex: 'port', key: 'port' },
    { title: 'TLS', dataIndex: 'tls', key: 'tls', render: (tls: boolean) => (tls ? <Tag color="green">TLS</Tag> : <Tag>Plain</Tag>) },
    {
      title: 'Connections', key: 'connections', render: (_: unknown, record: NodeRecord) => {
        const stats = trafficByListener[record.ID];
        return <Tag color={stats?.connections ? 'blue' : 'default'}>{stats?.connections || 0}</Tag>;
      },
    },
    {
      title: 'Traffic', key: 'traffic', render: (_: unknown, record: NodeRecord) => {
        const stats = trafficByListener[record.ID];
        return <span>↑ {formatBytes(stats?.upload || 0)} &nbsp; ↓ {formatBytes(stats?.download || 0)}</span>;
      },
    },
    {
      title: 'Status', dataIndex: 'enabled', key: 'enabled', render: (enabled: boolean, record: NodeRecord) => (
        <Switch checked={enabled} onChange={(checked) => void handleToggleEnabled(record, checked)} checkedChildren="On" unCheckedChildren="Off" />
      ),
    },
    {
      title: 'Actions', key: 'actions', render: (_: unknown, record: NodeRecord) => (
        <Space size="middle">
          <Button type="text" icon={<CopyOutlined style={{ color: '#722ed1' }} />} onClick={() => void generateClientAccess(record.ID)} title="Generate Client Link" />
          <Button type="text" icon={<ReloadOutlined style={{ color: '#52c41a' }} />} onClick={() => void handleReload(record.ID)} title="Hot Reload" />
          <Button type="text" icon={<EditOutlined style={{ color: '#1890ff' }} />} onClick={() => handleOpenEdit(record)} title="Edit" />
          <Popconfirm title="Are you sure you want to delete this listener?" onConfirm={() => void handleDelete(record.ID)} okText="Yes" cancelText="No">
            <Button type="text" danger icon={<DeleteOutlined />} title="Delete" />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const clientLinks = clientAccess
    ? [
        ['Mihomo / Clash', clientAccess.clash_link],
        ['sing-box', clientAccess.singbox_link],
        ['Shadowrocket', clientAccess.shadowrocket_link],
      ]
    : [];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div>
          <Title level={2} style={{ margin: 0 }}>Mihomo Listeners</Title>
          <Paragraph style={{ margin: 0 }}>Create Mihomo listeners and immediately generate client configuration links.</Paragraph>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleOpenAdd}>Add Listener</Button>
      </div>

      <Table dataSource={data} columns={columns} rowKey="ID" loading={loading} scroll={{ x: 'max-content' }} locale={{ emptyText: 'No listeners found. Create one to get started.' }} />

      <Modal title={editingRecord ? 'Edit Listener' : 'Add Listener'} open={modalOpen} onOk={() => void handleFormSubmit()} onCancel={() => setModalOpen(false)} destroyOnClose width={600}>
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="name" label="Name" rules={[{ required: true, message: 'Please input listener name' }]}>
            <Input placeholder="e.g. hk-shadowsocks" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="protocol" label="Listener Type" rules={[{ required: true, message: 'Select listener type' }]}>
                <Select placeholder="Choose listener type" onChange={() => form.setFieldsValue({ protocolConfig: {} })}>
                  <Select.Option value="shadowsocks">Shadowsocks</Select.Option>
                  <Select.Option value="vmess">VMess</Select.Option>
                  <Select.Option value="vless">VLESS</Select.Option>
                  <Select.Option value="trojan">Trojan</Select.Option>
                  <Select.Option value="hysteria2">Hysteria 2</Select.Option>
                  <Select.Option value="tuic">TUIC</Select.Option>
                  <Select.Option value="wireguard">WireGuard</Select.Option>
                </Select>
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
              <Form.Item name="bind_address" label="Bind Address" rules={[{ required: true, message: 'Please enter bind IP' }]}>
                <Input placeholder="e.g. 0.0.0.0" />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="tls" label="TLS" valuePropName="checked"><Switch checkedChildren="Enabled" unCheckedChildren="Disabled" /></Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="udp" label="UDP" valuePropName="checked"><Switch checkedChildren="Enabled" unCheckedChildren="Disabled" /></Form.Item>
            </Col>
          </Row>

          <Typography.Paragraph type="secondary">
            SOCKS, HTTP, TProxy, Redir, Mixed, Tunnel and TUN are not listener choices here. Only listener protocols that can be exported as client proxy configurations are exposed.
          </Typography.Paragraph>

          <ProtocolForm protocol={selectedProtocol} />

          <Form.Item name="enabled" label="Status Enabled" valuePropName="checked">
            <Switch checkedChildren="On" unCheckedChildren="Off" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title={`Client configuration — ${clientAccess?.name || ''}`} open={clientModalOpen} onCancel={() => setClientModalOpen(false)} footer={null} width={700} destroyOnClose>
        <Typography.Paragraph type="secondary">
          The listener has been created and its client distribution links are ready. Use these URLs directly in the corresponding client.
        </Typography.Paragraph>
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          {clientLinks.map(([label, link]) => (
            <div key={label} style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
              <Typography.Text strong style={{ width: 130 }}>{label}</Typography.Text>
              <Input value={link} readOnly />
              <Button icon={<CopyOutlined />} onClick={() => void copyLink(link)}>Copy</Button>
            </div>
          ))}
          {clientAccess?.mihomo_link && (
            <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
              <Typography.Text strong style={{ width: 130 }}>Raw Mihomo</Typography.Text>
              <Input value={clientAccess.mihomo_link} readOnly />
              <Button icon={<CopyOutlined />} onClick={() => void copyLink(clientAccess.mihomo_link)}>Copy</Button>
            </div>
          )}
        </Space>
      </Modal>
    </div>
  );
};

export default ListenersPage;
