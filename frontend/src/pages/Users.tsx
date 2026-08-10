import React, { useEffect, useState } from 'react';
import {
  Typography, Table, Button, Space, Tag, Modal, Form, Input, InputNumber,
  Switch, DatePicker, Select, message, Popconfirm, Progress, Badge
} from 'antd';
import { UserAddOutlined, EditOutlined, DeleteOutlined, LinkOutlined } from '@ant-design/icons';
import dayjs, { Dayjs } from 'dayjs';

const { Title, Paragraph } = Typography;
const API_BASE = 'http://localhost:8080/api/v1';

interface ProxyUser {
  id: number;
  username: string;
  uuid_masked: string;
  traffic_limit: number;
  traffic_used: number;
  expire_time: string;
  enabled: boolean;
  upload_bytes: number;
  download_bytes: number;
  last_seen: string;
  online: boolean;
}

interface Node {
  id: number;
  name: string;
  protocol: string;
  port: number;
  bind_address: string;
  enabled: boolean;
  tls: boolean;
  udp: boolean;
  status: string;
}

const formatBytes = (bytes: number) => {
  if (!bytes && bytes !== 0) return 'Unlimited';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = bytes;
  let i = 0;
  while (value >= 1024 && i < units.length - 1) { value /= 1024; i++; }
  return `${value.toFixed(value >= 10 || i === 0 ? 0 : 1)} ${units[i]}`;
};

const Users: React.FC = () => {
  const [users, setUsers] = useState<ProxyUser[]>([]);
  const [nodes, setNodes] = useState<Node[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [bindOpen, setBindOpen] = useState(false);
  const [editing, setEditing] = useState<ProxyUser | null>(null);
  const [bindingUser, setBindingUser] = useState<ProxyUser | null>(null);
  const [boundNodeIds, setBoundNodeIds] = useState<number[]>([]);
  const [form] = Form.useForm();
  const [bindForm] = Form.useForm();

  const fetchUsers = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/users`);
      if (!res.ok) throw new Error();
      setUsers(await res.json());
    } catch {
      message.error('Failed to load proxy users.');
    } finally { setLoading(false); }
  };

  const fetchNodes = async () => {
    try {
      const res = await fetch(`${API_BASE}/nodes`);
      if (res.ok) setNodes(await res.json());
    } catch { /* handled when binding */ }
  };

  useEffect(() => { fetchUsers(); fetchNodes(); }, []);

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ enabled: true, traffic_limit: 0 });
    setModalOpen(true);
  };

  const openEdit = (u: ProxyUser) => {
    setEditing(u);
    form.resetFields();
    form.setFieldsValue({
      username: u.username,
      traffic_limit: u.traffic_limit,
      enabled: u.enabled,
      expire_time: u.expire_time && !dayjs(u.expire_time).isBefore(dayjs('1971-01-01')) ? dayjs(u.expire_time) : undefined,
    });
    setModalOpen(true);
  };

  const saveUser = async () => {
    const values = await form.validateFields();
    const payload = {
      ...values,
      expire_time: values.expire_time ? (values.expire_time as Dayjs).toISOString() : undefined,
    };
    if (!payload.password) delete payload.password;

    const url = editing ? `${API_BASE}/users/${editing.id}` : `${API_BASE}/users`;
    const res = await fetch(url, {
      method: editing ? 'PUT' : 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      message.error(data.error || 'Failed to save user.');
      return;
    }
    message.success(editing ? 'User updated.' : 'User created.');
    setModalOpen(false);
    fetchUsers();
  };

  const removeUser = async (id: number) => {
    const res = await fetch(`${API_BASE}/users/${id}`, { method: 'DELETE' });
    if (res.ok) { message.success('User deleted.'); fetchUsers(); }
    else message.error('Failed to delete user.');
  };

  const openBind = async (u: ProxyUser) => {
    setBindingUser(u);
    const res = await fetch(`${API_BASE}/users/${u.id}/listeners`);
    if (res.ok) {
      const list: Node[] = await res.json();
      setBoundNodeIds(list.map(n => n.id));
    } else setBoundNodeIds([]);
    bindForm.resetFields();
    setBindOpen(true);
  };

  const saveBindings = async () => {
    const values = await bindForm.validateFields();
    const res = await fetch(`${API_BASE}/users/${bindingUser!.id}/listeners`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ listener_ids: values.listener_ids || [] }),
    });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      message.error(data.error || 'Failed to update node bindings.');
      return;
    }
    message.success('Node bindings updated.');
    setBindOpen(false);
  };

  const columns = [
    {
      title: 'Online',
      key: 'online',
      width: 90,
      render: (_: unknown, u: ProxyUser) => {
        return u.online ? (
          <Badge status="processing" text="Online" style={{ color: '#52c41a', fontWeight: 'bold' }} />
        ) : (
          <Badge status="default" text="Offline" style={{ color: '#8c8c8c' }} />
        );
      }
    },
    { title: 'Username', dataIndex: 'username', key: 'username' },
    { title: 'UUID', dataIndex: 'uuid_masked', key: 'uuid_masked' },
    {
      title: 'Traffic Stats',
      key: 'traffic',
      render: (_: unknown, u: ProxyUser) => {
        const percent = u.traffic_limit ? Math.min(100, (u.traffic_used / u.traffic_limit) * 100) : 0;
        const remaining = u.traffic_limit > 0 ? Math.max(0, u.traffic_limit - u.traffic_used) : null;
        return (
          <div style={{ minWidth: 200, fontSize: '12px' }}>
            <div>
              <strong>Used:</strong> {formatBytes(u.traffic_used)} / {formatBytes(u.traffic_limit)}
            </div>
            <div>
              <span style={{ color: '#1890ff' }}>↑ {formatBytes(u.upload_bytes || 0)}</span> |{' '}
              <span style={{ color: '#52c41a' }}>↓ {formatBytes(u.download_bytes || 0)}</span>
            </div>
            {remaining !== null && (
              <div style={{ color: remaining < 1024 * 1024 * 1024 ? 'red' : 'inherit' }}>
                <strong>Remaining:</strong> {formatBytes(remaining)}
              </div>
            )}
            {u.traffic_limit > 0 && <Progress percent={Number(percent.toFixed(1))} size="small" showInfo={false} />}
          </div>
        );
      },
    },
    {
      title: 'Last Seen',
      key: 'last_seen',
      render: (_: unknown, u: ProxyUser) => {
        if (!u.last_seen || dayjs(u.last_seen).year() <= 1970) {
          return 'Never';
        }
        return dayjs(u.last_seen).format('YYYY-MM-DD HH:mm:ss');
      }
    },
    {
      title: 'Expire',
      dataIndex: 'expire_time',
      key: 'expire_time',
      render: (v: string) => {
        if (!v || dayjs(v).year() <= 1970) {
          return 'Never';
        }
        return dayjs(v).format('YYYY-MM-DD HH:mm');
      },
    },
    {
      title: 'Status',
      key: 'status',
      render: (_: unknown, u: ProxyUser) => {
        const expired = u.expire_time && dayjs(u.expire_time).year() > 1970 && dayjs(u.expire_time).isBefore(dayjs());
        const overlimit = u.traffic_limit > 0 && u.traffic_used >= u.traffic_limit;
        return (
          <Space direction="vertical" size={4}>
            <Tag color={!u.enabled ? 'default' : expired ? 'red' : overlimit ? 'warning' : 'green'}>
              {!u.enabled ? 'Disabled' : expired ? 'Expired' : overlimit ? 'Exceeded' : 'Active'}
            </Tag>
          </Space>
        );
      },
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, u: ProxyUser) => (
        <Space>
          <Button icon={<LinkOutlined />} onClick={() => openBind(u)}>Nodes</Button>
          <Button icon={<EditOutlined />} onClick={() => openEdit(u)}>Edit</Button>
          <Popconfirm title="Delete this proxy user?" onConfirm={() => removeUser(u.id)}>
            <Button danger icon={<DeleteOutlined />}>Delete</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div>
          <Title level={2} style={{ margin: 0 }}>Proxy Users</Title>
          <Paragraph style={{ margin: 0 }}>Manage credentials, detailed traffic statistics, online status, and node access for Mihomo users.</Paragraph>
        </div>
        <Button type="primary" icon={<UserAddOutlined />} onClick={openCreate}>Create User</Button>
      </div>

      <Table rowKey="id" dataSource={users} columns={columns} loading={loading} />

      <Modal
        title={editing ? 'Edit Proxy User' : 'Create Proxy User'}
        open={modalOpen}
        onOk={saveUser}
        onCancel={() => setModalOpen(false)}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Form.Item name="username" label="Username" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item
            name="password"
            label={editing ? 'Password (leave empty to keep current)' : 'Password'}
            rules={editing ? [] : [{ required: true }]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>
          <Form.Item name="uuid" label="UUID (leave empty to auto-generate)">
            <Input />
          </Form.Item>
          <Form.Item name="traffic_limit" label="Traffic Limit (Bytes, 0 = unlimited)">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="expire_time" label="Expire Time">
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="enabled" label="Enabled" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={`Bind Nodes${bindingUser ? ` · ${bindingUser.username}` : ''}`}
        open={bindOpen}
        onOk={saveBindings}
        onCancel={() => setBindOpen(false)}
        destroyOnClose
      >
        <Form form={bindForm} initialValues={{ listener_ids: boundNodeIds }}>
          <Form.Item name="listener_ids" label="Allowed Nodes">
            <Select
              mode="multiple"
              placeholder="Select nodes"
              options={nodes.map(n => ({
                value: n.id,
                label: `${n.name} · ${n.protocol} · :${n.port}`,
              }))}
            />
          </Form.Item>
          <Typography.Text type="secondary">
            Saving replaces the existing bindings. Selecting no nodes revokes access to all nodes.
          </Typography.Text>
        </Form>
      </Modal>
    </div>
  );
};

export default Users;
