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
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons';
import { apiRequest } from '../api/request';
import { ProtocolForm } from '../components/protocols';

const { Title, Paragraph } = Typography;
const { Option } = Select;

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

const formatBytes = (bytes: number): string => {
  if (!bytes) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = bytes;
  let i = 0;
  while (value >= 1024 && i < units.length - 1) { value /= 1024; i++; }
  return `${value.toFixed(value >= 10 || i === 0 ? 0 : 1)} ${units[i]}`;
};


const Listeners: React.FC = () => {
  const [data, setData] = useState<NodeRecord[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [modalOpen, setModalOpen] = useState<boolean>(false);
  const [editingRecord, setEditingRecord] = useState<NodeRecord | null>(null);
  const [form] = Form.useForm();
  const [trafficByListener, setTrafficByListener] = useState<Record<number, ListenerTrafficStats>>({});

  // Track selected protocol to dynamically show/hide inputs
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
      // Traffic stats are a non-critical overlay; ignore transient failures.
    }
  };

  useEffect(() => {
    fetchNodes();
    fetchTraffic();
    const interval = setInterval(fetchTraffic, 10000);
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

    // Node-level config contains protocol options only. Authentication
    // credentials are managed by Proxy Users and are intentionally never
    // displayed or accepted on the node form.
    let protocolConfig: Record<string, unknown> = {};
    let flow = '';

    try {
      const parsed = JSON.parse(record.config || '{}');
      if (parsed.flow) flow = parsed.flow;
      const cleaned = { ...parsed };
      protocolConfig = cleaned;
    } catch {
      // Ignore malformed legacy config here; backend validation will report it.
    }

    form.setFieldsValue({ ...record, flow, protocolConfig });
    setModalOpen(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await apiRequest(`/nodes/${id}`, { method: 'DELETE' });
      message.success('Node deleted successfully.');
      fetchNodes();
    } catch {
      message.error('Network connection error.');
    }
  };

  const handleToggleEnabled = async (record: NodeRecord, checked: boolean) => {
    const updated = { ...record, enabled: checked, status: checked ? 'active' : 'inactive' };
    try {
      await apiRequest(`/nodes/${record.ID}`, { method: 'PUT', body: JSON.stringify(updated) });
      message.success(`Node ${checked ? 'enabled' : 'disabled'} successfully.`);
      fetchNodes();
    } catch {
      message.error('Network connection error.');
    }
  };

  const handleReload = async (id: number) => {
    try {
      await apiRequest(`/nodes/${id}/reload`, { method: 'POST' });
      message.success('Mihomo configuration reloaded!');
    } catch {
      message.error('Network connection error.');
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

      await apiRequest(url, { method, body: JSON.stringify(payload) });

      message.success(`Node ${editingRecord ? 'updated' : 'created'} successfully.`);
      setModalOpen(false);
      fetchNodes();
    } catch {
      // Validation failed
    }
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'ID',
      key: 'ID',
      width: 60,
    },
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Protocol',
      dataIndex: 'protocol',
      key: 'protocol',
      render: (proto: string) => <Tag color="blue">{proto}</Tag>,
    },
    {
      title: 'Port',
      dataIndex: 'port',
      key: 'port',
    },
    {
      title: 'TLS',
      dataIndex: 'tls',
      key: 'tls',
      render: (tls: boolean) => (tls ? <Tag color="green">TLS</Tag> : <Tag color="default">Plain</Tag>),
    },
    {
      title: 'Connections',
      key: 'connections',
      render: (_: any, record: NodeRecord) => {
        const stats = trafficByListener[record.ID];
        return <Tag color={stats?.connections ? 'blue' : 'default'}>{stats?.connections || 0}</Tag>;
      },
    },
    {
      title: 'Traffic',
      key: 'traffic',
      render: (_: any, record: NodeRecord) => {
        const stats = trafficByListener[record.ID];
        return (
          <span>
            ↑ {formatBytes(stats?.upload || 0)} &nbsp; ↓ {formatBytes(stats?.download || 0)}
          </span>
        );
      },
    },
    {
      title: 'Status',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean, record: NodeRecord) => (
        <Switch
          checked={enabled}
          onChange={(checked) => handleToggleEnabled(record, checked)}
          checkedChildren="On"
          unCheckedChildren="Off"
        />
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: NodeRecord) => (
        <Space size="middle">
          <Button
            type="text"
            icon={<ReloadOutlined style={{ color: '#52c41a' }} />}
            onClick={() => handleReload(record.ID)}
            title="Hot Reload"
          />
          <Button
            type="text"
            icon={<EditOutlined style={{ color: '#1890ff' }} />}
            onClick={() => handleOpenEdit(record)}
            title="Edit"
          />
          <Popconfirm
            title="Are you sure you want to delete this node?"
            onConfirm={() => handleDelete(record.ID)}
            okText="Yes"
            cancelText="No"
          >
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              title="Delete"
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div>
          <Title level={2} style={{ margin: 0 }}>Mihomo Server Nodes</Title>
          <Paragraph style={{ margin: 0 }}>
            Manage server nodes, protocols, inbounds, security configurations, and credentials.
          </Paragraph>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleOpenAdd}>
          Add Server Node
        </Button>
      </div>

      <Table
        dataSource={data}
        columns={columns}
        rowKey="ID"
        loading={loading}
        scroll={{ x: 'max-content' }}
        locale={{ emptyText: 'No Server Nodes Found. Create one to get started!' }}
      />

      {/* Add / Edit Node Modal */}
      <Modal
        title={editingRecord ? 'Edit Server Node' : 'Add Server Node'}
        open={modalOpen}
        onOk={handleFormSubmit}
        onCancel={() => setModalOpen(false)}
        destroyOnClose
        width={600}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: 'Please input node name' }]}
          >
            <Input placeholder="e.g. hk-shadowsocks" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="protocol"
                label="Protocol"
                rules={[{ required: true, message: 'Select node protocol' }]}
              >
                <Select placeholder="Choose protocol" onChange={() => {
                  form.setFieldsValue({ protocolConfig: {} });
                }}>
                  <Option value="shadowsocks">Shadowsocks</Option>
                  <Option value="vmess">VMess</Option>
                  <Option value="vless">VLESS</Option>
                  <Option value="trojan">Trojan</Option>
                  <Option value="hysteria2">Hysteria 2</Option>
                  <Option value="tuic">TUIC</Option>
                  <Option value="wireguard">WireGuard</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="port"
                label="Port"
                rules={[{ required: true, message: 'Please enter port number' }]}
              >
                <InputNumber min={1} max={65535} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="bind_address"
                label="Bind Address"
                rules={[{ required: true, message: 'Please enter bind IP' }]}
              >
                <Input placeholder="e.g. 0.0.0.0" />
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

          <Typography.Paragraph type="secondary" style={{ marginBottom: 16 }}>
            Authentication credentials are managed under Proxy Users and bound to nodes separately.
            Do not place passwords or UUIDs in the node configuration.
          </Typography.Paragraph>

          <ProtocolForm protocol={selectedProtocol} />

          <Form.Item name="enabled" label="Status Enabled" valuePropName="checked">
            <Switch checkedChildren="On" unCheckedChildren="Off" />
          </Form.Item>

        </Form>
      </Modal>
    </div>
  );
};

export default Listeners;
