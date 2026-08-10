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

interface ListenerRecord {
  ID: number;
  name: string;
  type: string;
  listen: string;
  port: number;
  udp: boolean;
  enabled: boolean;
  proxy: string;
  rule: string;
  config: string;
  CreatedAt?: string;
}

const API_BASE = 'http://localhost:8080/api/v1';

const Listeners: React.FC = () => {
  const [data, setData] = useState<ListenerRecord[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [modalOpen, setModalOpen] = useState<boolean>(false);
  const [editingRecord, setEditingRecord] = useState<ListenerRecord | null>(null);
  const [form] = Form.useForm();

  const fetchListeners = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/listeners`);
      if (res.ok) {
        const list = await res.json();
        setData(list || []);
      } else {
        message.error('Failed to fetch listeners from backend.');
      }
    } catch (err) {
      message.error('Backend service unreachable.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchListeners();
  }, []);

  const handleOpenAdd = () => {
    setEditingRecord(null);
    form.resetFields();
    form.setFieldsValue({
      listen: '0.0.0.0',
      port: 7890,
      type: 'mixed',
      udp: false,
      enabled: true,
    });
    setModalOpen(true);
  };

  const handleOpenEdit = (record: ListenerRecord) => {
    setEditingRecord(record);
    form.resetFields();
    form.setFieldsValue(record);
    setModalOpen(true);
  };

  const handleDelete = async (id: number) => {
    try {
      const res = await fetch(`${API_BASE}/listeners/${id}`, {
        method: 'DELETE',
      });
      if (res.ok) {
        message.success('Listener deleted successfully.');
        fetchListeners();
      } else {
        message.error('Failed to delete listener.');
      }
    } catch (err) {
      message.error('Network connection error.');
    }
  };

  const handleToggleEnabled = async (record: ListenerRecord, checked: boolean) => {
    const updated = { ...record, enabled: checked };
    try {
      const res = await fetch(`${API_BASE}/listeners/${record.ID}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updated),
      });
      if (res.ok) {
        message.success(`Listener ${checked ? 'enabled' : 'disabled'} successfully.`);
        fetchListeners();
      } else {
        message.error('Failed to update status.');
      }
    } catch (err) {
      message.error('Network connection error.');
    }
  };

  const handleReload = async (id: number) => {
    try {
      const res = await fetch(`${API_BASE}/listeners/${id}/reload`, {
        method: 'POST',
      });
      if (res.ok) {
        message.success('Configuration hot-reloaded into Mihomo Core!');
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
      const method = editingRecord ? 'PUT' : 'POST';
      const url = editingRecord ? `${API_BASE}/listeners/${editingRecord.ID}` : `${API_BASE}/listeners`;

      const res = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(values),
      });

      if (res.ok) {
        message.success(`Listener ${editingRecord ? 'updated' : 'created'} successfully.`);
        setModalOpen(false);
        fetchListeners();
      } else {
        const errorData = await res.json();
        message.error(errorData.error || 'Failed to save listener.');
      }
    } catch (err) {
      // Validation failed or network issue
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
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => <Tag color="blue">{type}</Tag>,
    },
    {
      title: 'Listen Address',
      dataIndex: 'listen',
      key: 'listen',
    },
    {
      title: 'Port',
      dataIndex: 'port',
      key: 'port',
    },
    {
      title: 'UDP',
      dataIndex: 'udp',
      key: 'udp',
      render: (udp: boolean) => (udp ? <Tag color="purple">UDP</Tag> : <Tag color="default">TCP</Tag>),
    },
    {
      title: 'Status',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean, record: ListenerRecord) => (
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
      render: (_: any, record: ListenerRecord) => (
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
            title="Are you sure you want to delete this listener?"
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
          <Title level={2} style={{ margin: 0 }}>Listeners</Title>
          <Paragraph style={{ margin: 0 }}>
            Manage and generate configurations for Mihomo Core inbound listener ports.
          </Paragraph>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleOpenAdd}>
          Add Listener
        </Button>
      </div>

      <Table
        dataSource={data}
        columns={columns}
        rowKey="ID"
        loading={loading}
        locale={{ emptyText: 'No Inbound Listeners Found. Create one to get started!' }}
      />

      {/* Add / Edit Modal */}
      <Modal
        title={editingRecord ? 'Edit Inbound Listener' : 'Add Inbound Listener'}
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
            rules={[{ required: true, message: 'Please input listener name' }]}
          >
            <Input placeholder="e.g. hk-mixed" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="type"
                label="Type"
                rules={[{ required: true, message: 'Select listener type' }]}
              >
                <Select placeholder="Choose protocol type">
                  <Option value="mixed">mixed</Option>
                  <Option value="socks">socks</Option>
                  <Option value="http">http</Option>
                  <Option value="redir">redir</Option>
                  <Option value="tproxy">tproxy</Option>
                  <Option value="tunnel">tunnel</Option>
                  <Option value="shadowsocks">shadowsocks</Option>
                  <Option value="vmess">vmess</Option>
                  <Option value="tuic">tuic</Option>
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
                name="listen"
                label="Listen IP"
                rules={[{ required: true, message: 'Please enter binding IP' }]}
              >
                <Input placeholder="e.g. 0.0.0.0" />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="udp" label="UDP" valuePropName="checked">
                <Switch checkedChildren="Enabled" unCheckedChildren="Disabled" />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="enabled" label="Enabled" valuePropName="checked">
                <Switch checkedChildren="On" unCheckedChildren="Off" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="proxy" label="Proxy Node Selector">
            <Input placeholder="e.g. DIRECT, Proxy, etc." />
          </Form.Item>

          <Form.Item name="rule" label="Rule Set">
            <Input placeholder="e.g. GEOIP,CN,DIRECT" />
          </Form.Item>

          <Form.Item
            name="config"
            label="Extra Parameters (JSON / YAML format)"
            extra="Extra settings merged dynamically to this listener block."
          >
            <TextArea rows={4} placeholder={`e.g. \n{\n  "sniffing": true\n}`} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Listeners;
