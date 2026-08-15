import React, { useEffect, useState } from 'react';
import {
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  InputNumber,
  Select,
  Switch,
  message,
  Popconfirm,
  Tooltip,
  Card,
} from 'antd';
import {
  PlusOutlined,
  ReloadOutlined,
  QrcodeOutlined,
  DeleteOutlined,
  EditOutlined,
  CopyOutlined,
} from '@ant-design/icons';
import {
  fetchListeners,
  createListener,
  updateListener,
  deleteListener,
  reloadListener,
  exportNodeURI,
  Listener,
} from '../api/nodes';

const PROTOCOLS = [
  'shadowsocks', 'snell', 'vmess', 'vless', 'trojan',
  'hysteria2', 'tuic', 'shadowquic', 'anytls', 'mieru',
  'sudoku', 'trusttunnel',
];

const Listeners: React.FC = () => {
  const [data, setData] = useState<Listener[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Listener | null>(null);
  const [form] = Form.useForm();
  const [uris, setUris] = useState<string[]>([]);
  const [uriModal, setUriModal] = useState(false);

  const load = async () => {
    setLoading(true);
    try {
      const res = await fetchListeners();
      setData(res);
    } catch (e: any) {
      message.error(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const onSubmit = async (values: any) => {
    try {
      const payload = {
        ...values,
        config: values.config ? JSON.stringify(JSON.parse(values.config)) : '{}',
      };
      if (editing) {
        await updateListener(editing.id, payload);
        message.success('Updated');
      } else {
        await createListener(payload);
        message.success('Created');
      }
      setModalOpen(false);
      setEditing(null);
      form.resetFields();
      load();
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const onDelete = async (id: number) => {
    try {
      await deleteListener(id);
      message.success('Deleted');
      load();
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const onReload = async (id: number) => {
    try {
      await reloadListener(id);
      message.success('Reloaded');
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const showURIs = async (id: number) => {
    try {
      const res = await exportNodeURI(id);
      setUris(res.uris);
      setUriModal(true);
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const columns = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Protocol', dataIndex: 'protocol', key: 'protocol', render: (p: string) => <Tag>{p}</Tag> },
    { title: 'Port', dataIndex: 'port', key: 'port' },
    {
      title: 'Status',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (v: boolean) => (
        <Tag color={v ? 'green' : 'red'}>{v ? 'Enabled' : 'Disabled'}</Tag>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: Listener) => (
        <Space>
          <Tooltip title="Export URI">
            <Button icon={<QrcodeOutlined />} size="small" onClick={() => showURIs(record.id)} />
          </Tooltip>
          <Tooltip title="Reload">
            <Button icon={<ReloadOutlined />} size="small" onClick={() => onReload(record.id)} />
          </Tooltip>
          <Tooltip title="Edit">
            <Button
              icon={<EditOutlined />}
              size="small"
              onClick={() => {
                setEditing(record);
                form.setFieldsValue({
                  ...record,
                  config: record.config ? JSON.stringify(JSON.parse(record.config), null, 2) : '{}',
                });
                setModalOpen(true);
              }}
            />
          </Tooltip>
          <Popconfirm title="Delete?" onConfirm={() => onDelete(record.id)}>
            <Button icon={<DeleteOutlined />} danger size="small" />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <h2 style={{ margin: 0 }}>Listeners</h2>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => {
            setEditing(null);
            form.resetFields();
            setModalOpen(true);
          }}
        >
          Add Listener
        </Button>
      </div>

      <Card>
        <Table rowKey="id" columns={columns} dataSource={data} loading={loading} />
      </Card>

      <Modal
        open={modalOpen}
        title={editing ? 'Edit Listener' : 'New Listener'}
        onCancel={() => {
          setModalOpen(false);
          setEditing(null);
          form.resetFields();
        }}
        onOk={() => form.submit()}
        width={640}
      >
        <Form form={form} layout="vertical" onFinish={onSubmit}>
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="protocol" label="Protocol" rules={[{ required: true }]} initialValue="shadowsocks">
            <Select options={PROTOCOLS.map((p) => ({ value: p, label: p }))} />
          </Form.Item>
          <Form.Item name="port" label="Port" rules={[{ required: true, type: 'number', min: 1, max: 65535 }]}>
            <InputNumber style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="bind_address" label="Bind Address" initialValue="0.0.0.0">
            <Input />
          </Form.Item>
          <Form.Item name="enabled" label="Enabled" valuePropName="checked" initialValue={true}>
            <Switch />
          </Form.Item>
          <Form.Item name="udp" label="UDP" valuePropName="checked" initialValue={false}>
            <Switch />
          </Form.Item>
          <Form.Item name="tls" label="TLS" valuePropName="checked" initialValue={false}>
            <Switch />
          </Form.Item>
          <Form.Item
            name="config"
            label="Config (JSON)"
            rules={[
              {
                validator: (_, v) => {
                  if (!v) return Promise.resolve();
                  try {
                    JSON.parse(v);
                    return Promise.resolve();
                  } catch {
                    return Promise.reject(new Error('Invalid JSON'));
                  }
                },
              },
            ]}
          >
            <Input.TextArea rows={6} placeholder='{"password":"..."}' />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        open={uriModal}
        title="Node URIs"
        onCancel={() => setUriModal(false)}
        footer={null}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          {uris.map((uri, i) => (
            <Input
              key={i}
              value={uri}
              readOnly
              addonAfter={
                <Button
                  icon={<CopyOutlined />}
                  type="text"
                  onClick={() => {
                    navigator.clipboard.writeText(uri);
                    message.success('Copied');
                  }}
                />
              }
            />
          ))}
        </Space>
      </Modal>
    </div>
  );
};

export default Listeners;
