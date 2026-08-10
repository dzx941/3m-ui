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

const { Title, Paragraph } = Typography;
const { TextArea } = Input;
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
  connections?: number;
  upload_bytes?: number;
  download_bytes?: number;
}

const API_BASE = 'http://localhost:8080/api/v1';

const formatBytes = (bytes?: number) => {
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

  // Track selected protocol to dynamically show/hide inputs
  const selectedProtocol = Form.useWatch('protocol', form);

  const fetchNodes = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/nodes`);
      if (res.ok) {
        const list = await res.json();
        setData(list || []);
      } else {
        message.error('Failed to fetch nodes from backend.');
      }
    } catch (err) {
      message.error('Backend service unreachable.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchNodes();
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
      rawConfig: '{}',
    });
    setModalOpen(true);
  };

  const handleOpenEdit = (record: NodeRecord) => {
    setEditingRecord(record);
    form.resetFields();

    // Node-level config contains protocol options only. Authentication
    // credentials are managed by Proxy Users and are intentionally never
    // displayed or accepted on the node form.
    let rawConfig = record.config || '{}';
    let flow = '';

    try {
      const parsed = JSON.parse(record.config || '{}');
      if (parsed.flow) flow = parsed.flow;
      const cleaned = { ...parsed };
      delete cleaned.password;
      delete cleaned.uuid;
      rawConfig = JSON.stringify(cleaned, null, 2);
    } catch (e) {
      // Ignore malformed legacy config here; backend validation will report it.
    }

    form.setFieldsValue({ ...record, flow, rawConfig });
    setModalOpen(true);
  };

  const handleDelete = async (id: number) => {
    try {
      const res = await fetch(`${API_BASE}/nodes/${id}`, {
        method: 'DELETE',
      });
      if (res.ok) {
        message.success('Node deleted successfully.');
        fetchNodes();
      } else {
        message.error('Failed to delete node.');
      }
    } catch (err) {
      message.error('Network connection error.');
    }
  };

  const handleToggleEnabled = async (record: NodeRecord, checked: boolean) => {
    const updated = { ...record, enabled: checked, status: checked ? 'active' : 'inactive' };
    try {
      const res = await fetch(`${API_BASE}/nodes/${record.ID}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updated),
      });
      if (res.ok) {
        message.success(`Node ${checked ? 'enabled' : 'disabled'} successfully.`);
        fetchNodes();
      } else {
        message.error('Failed to update status.');
      }
    } catch (err) {
      message.error('Network connection error.');
    }
  };

  const handleReload = async (id: number) => {
    try {
      const res = await fetch(`${API_BASE}/nodes/${id}/reload`, {
        method: 'POST',
      });
      if (res.ok) {
        message.success('Mihomo configuration reloaded!');
      } else {
        message.error('Failed to hot-reload configuration.');
      }
    } catch (err) {
      message.error('Network connection error.');
    }
  };

  const handleFormSubmit = async () => {
    try {
      const values = await form.validateFields();

      // Combine password/uuid/flow inputs with rawConfig to create the unified config field
      let configObj: Record<string, any> = {};
      try {
        configObj = JSON.parse(values.rawConfig || '{}');
      } catch (e) {
        message.error('Extra Parameters must be valid JSON syntax.');
        return;
      }

      // Only node-level options belong here. ProxyUser credentials are injected
      // by the backend Config Engine based on ListenerUser bindings.
      if (values.flow) configObj.flow = values.flow;

      const payload = {
        ...values,
        config: JSON.stringify(configObj),
        status: values.enabled ? 'active' : 'inactive',
      };

      const method = editingRecord ? 'PUT' : 'POST';
      const url = editingRecord ? `${API_BASE}/nodes/${editingRecord.ID}` : `${API_BASE}/nodes`;

      const res = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        message.success(`Node ${editingRecord ? 'updated' : 'created'} successfully.`);
        setModalOpen(false);
        fetchNodes();
      } else {
        const errorData = await res.json();
        message.error(errorData.error || 'Failed to save node.');
      }
    } catch (err) {
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
      dataIndex: 'connections',
      key: 'connections',
      render: (conns?: number) => <Tag color={conns ? 'green' : 'default'}>{conns || 0} active</Tag>,
    },
    {
      title: 'Traffic Stats',
      key: 'traffic',
      render: (_: any, record: NodeRecord) => (
        <div style={{ fontSize: '12px' }}>
          <span style={{ color: '#1890ff' }}>↑ {formatBytes(record.upload_bytes)}</span> |{' '}
          <span style={{ color: '#52c41a' }}>↓ {formatBytes(record.download_bytes)}</span>
        </div>
      ),
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
            Manage server nodes, protocols, inbounds, security configurations, credentials, connections, and traffic statistics.
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
                  form.setFieldsValue({ rawConfig: '{}' });
                }}>
                  <Option value="shadowsocks">Shadowsocks</Option>
                  <Option value="vmess">VMess</Option>
                  <Option value="vless">VLESS</Option>
                  <Option value="trojan">Trojan</Option>
                  <Option value="hysteria2">Hysteria 2</Option>
                  <Option value="tuic">TUIC</Option>
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

          {selectedProtocol === 'vless' && (
            <Form.Item name="flow" label="Flow Parameter">
              <Input placeholder="e.g. xtls-rprx-vision" />
            </Form.Item>
          )}

          <Form.Item name="enabled" label="Status Enabled" valuePropName="checked">
            <Switch checkedChildren="On" unCheckedChildren="Off" />
          </Form.Item>

          <Form.Item
            name="rawConfig"
            label="Extra JSON Settings"
            extra="Additional settings merged dynamically to this server node."
          >
            <TextArea rows={4} placeholder={`e.g. \n{\n  "obfs": "http"\n}`} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Listeners;
