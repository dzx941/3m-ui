import React, { useEffect, useState } from 'react';
import {
  Typography, Table, Button, Space, Tag, Modal, Form, Input, InputNumber,
  Switch, DatePicker, Select, message, Popconfirm, Progress,
} from 'antd';
import { UserAddOutlined, EditOutlined, DeleteOutlined, LinkOutlined } from '@ant-design/icons';
import { apiRequest } from '../api/request';
import dayjs, { Dayjs } from 'dayjs';

const { Title, Paragraph } = Typography;

interface ProxyUser {
  id: number;
  username: string;
  uuid_masked: string;
  traffic_limit: number;
  traffic_used: number;
  upload_bytes: number;
  download_bytes: number;
  last_seen: string | null;
  online: boolean;
  expire_time: string;
  enabled: boolean;
  blocked: boolean;
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
  if (!bytes) return 'Unlimited';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = bytes;
  let i = 0;
  while (value >= 1024 && i < units.length - 1) { value /= 1024; i++; }
  return `${value.toFixed(value >= 10 || i === 0 ? 0 : 1)} ${units[i]}`;
};

// Same formatting as formatBytes but for absolute traffic amounts (upload/
// download counters), where 0 means "none transferred yet" rather than
// "unlimited".
const formatDataAmount = (bytes: number) => {
  if (!bytes) return '0 B';
  return formatBytes(bytes);
};

const 用户管理: React.FC = () => {
  const [users, set用户管理] = useState<ProxyUser[]>([]);
  const [nodes, set节点管理] = useState<Node[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [bindOpen, setBindOpen] = useState(false);
  const [editing, setEditing] = useState<ProxyUser | null>(null);
  const [bindingUser, setBindingUser] = useState<ProxyUser | null>(null);
  const [boundNodeIds, setBoundNodeIds] = useState<number[]>([]);
  const [form] = Form.useForm();
  const [bindForm] = Form.useForm();

  const fetch用户管理 = async (silent = false) => {
    if (!silent) setLoading(true);
    try {
      set用户管理(await apiRequest<ProxyUser[]>('/users'));
    } catch {
      if (!silent) message.error('Failed to load proxy users.');
    } finally { if (!silent) setLoading(false); }
  };

  const fetch节点管理 = async () => {
    try {
      set节点管理(await apiRequest<Node[]>('/nodes'));
    } catch { /* handled when binding */ }
  };

  useEffect(() => {
    fetch用户管理();
    fetch节点管理();
    // Traffic/online state changes as the backend's traffic scheduler ticks
    // (10s), so refresh in the background to keep those columns current.
    const interval = setInterval(() => fetch用户管理(true), 10000);
    return () => clearInterval(interval);
  }, []);

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ enabled: true, traffic_limit: 0 });
    setModalOpen(true);
  };

  const openEdit = (u: ProxyUser) => {
    setEditing(u);
    form.setFieldsValue({
      username: u.username,
      traffic_limit: u.traffic_limit,
      enabled: u.enabled,
      expire_time: u.expire_time ? dayjs(u.expire_time) : undefined,
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

    const url = editing ? `/users/${editing.id}` : '/users';
    try {
      await apiRequest(url, { method: editing ? 'PUT' : 'POST', body: JSON.stringify(payload) });
    } catch (err) {
      message.error((err as { message?: string }).message || 'Failed to save user.');
      return;
    }
    message.success(editing ? 'User updated.' : 'User created.');
    setModalOpen(false);
    fetch用户管理();
  };

  const removeUser = async (id: number) => {
    try { await apiRequest(`/users/${id}`, { method: 'DELETE' }); message.success('User deleted.'); fetch用户管理(); }
    catch (err) { message.error((err as { message?: string }).message || 'Failed to delete user.'); }
  };

  const openBind = async (u: ProxyUser) => {
    setBindingUser(u);
    try {
      const list = await apiRequest<Node[]>(`/users/${u.id}/listeners`);
      setBoundNodeIds(list.map(n => n.id));
    } catch { setBoundNodeIds([]); }
    bindForm.resetFields();
    setBindOpen(true);
  };

  const saveBindings = async () => {
    const values = await bindForm.validateFields();
    try {
      await apiRequest(`/users/${bindingUser!.id}/listeners`, { method: 'POST', body: JSON.stringify({ listener_ids: values.listener_ids || [] }) });
    } catch (err) {
      message.error((err as { message?: string }).message || 'Failed to update node bindings.');
      return;
    }
    message.success('Node bindings updated.');
    setBindOpen(false);
  };

  const columns = [
    { title: '用户名', dataIndex: 'username', key: 'username' },
    { title: 'UUID', dataIndex: 'uuid_masked', key: 'uuid_masked' },
    {
      title: 'Traffic',
      key: 'traffic',
      render: (_: unknown, u: ProxyUser) => {
        const percent = u.traffic_limit ? Math.min(100, (u.traffic_used / u.traffic_limit) * 100) : 0;
        return (
          <div style={{ minWidth: 150 }}>
            <div>{formatBytes(u.traffic_used)} / {formatBytes(u.traffic_limit)}</div>
            {u.traffic_limit > 0 && <Progress percent={Number(percent.toFixed(1))} size="small" showInfo={false} />}
          </div>
        );
      },
    },
    {
      title: 'Upload',
      dataIndex: 'upload_bytes',
      key: 'upload_bytes',
      render: (v: number) => formatDataAmount(v),
    },
    {
      title: 'Download',
      dataIndex: 'download_bytes',
      key: 'download_bytes',
      render: (v: number) => formatDataAmount(v),
    },
    {
      title: 'Remaining Quota',
      key: 'remaining_quota',
      render: (_: unknown, u: ProxyUser) => (
        u.traffic_limit > 0 ? formatDataAmount(Math.max(0, u.traffic_limit - u.traffic_used)) : 'Unlimited'
      ),
    },
    {
      title: 'Expire',
      dataIndex: 'expire_time',
      key: 'expire_time',
      render: (v: string) => v ? dayjs(v).format('YYYY-MM-DD HH:mm') : 'Never',
    },
    {
      title: 'Online',
      key: 'online',
      render: (_: unknown, u: ProxyUser) => (
        <Tag color={u.online ? 'success' : 'default'}>{u.online ? 'Online' : 'Offline'}</Tag>
      ),
    },
    {
      title: 'Last Seen',
      dataIndex: 'last_seen',
      key: 'last_seen',
      render: (v: string | null) => v ? dayjs(v).format('YYYY-MM-DD HH:mm') : 'Never',
    },
    {
      title: 'Status',
      key: 'status',
      render: (_: unknown, u: ProxyUser) => {
        const expired = u.expire_time && dayjs(u.expire_time).isBefore(dayjs());
        const overQuota = u.traffic_limit > 0 && u.traffic_used >= u.traffic_limit;
        const label = !u.enabled ? 'Disabled' : expired ? 'Expired' : overQuota ? 'Over Quota' : 'Active';
        return <Tag color={label === 'Active' ? 'green' : label === 'Disabled' ? 'default' : 'red'}>{label}</Tag>;
      },
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, u: ProxyUser) => (
        <Space>
          <Button icon={<LinkOutlined />} onClick={() => openBind(u)}>节点管理</Button>
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
          <Title level={2} style={{ margin: 0 }}>Proxy 用户管理</Title>
          <Paragraph style={{ margin: 0 }}>Manage credentials and node access for Mihomo users.</Paragraph>
        </div>
        <Button type="primary" icon={<UserAddOutlined />} onClick={openCreate}>Create User</Button>
      </div>

      <Table rowKey="id" dataSource={users} columns={columns} loading={loading} scroll={{ x: 'max-content' }} />

      <Modal
        title={editing ? 'Edit Proxy User' : 'Create Proxy User'}
        open={modalOpen}
        onOk={saveUser}
        onCancel={() => setModalOpen(false)}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Form.Item name="username" label="用户名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item
            name="password"
            label={editing ? '密码 (leave empty to keep current)' : '密码'}
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
        title={`Bind 节点管理${bindingUser ? ` · ${bindingUser.username}` : ''}`}
        open={bindOpen}
        onOk={saveBindings}
        onCancel={() => setBindOpen(false)}
        destroyOnClose
      >
        <Form form={bindForm} initialValues={{ listener_ids: boundNodeIds }}>
          <Form.Item name="listener_ids" label="Allowed 节点管理">
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

export default 用户管理;
