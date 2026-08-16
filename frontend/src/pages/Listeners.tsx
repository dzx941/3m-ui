import React, { useEffect, useState } from 'react';
import { Table, Button, Space, Tag, Modal, Form, Input, Select, Switch, message, Popconfirm, Tooltip, Card } from 'antd';
import { PlusOutlined, ReloadOutlined, QrcodeOutlined, DeleteOutlined, EditOutlined, CopyOutlined } from '@ant-design/icons';
import { fetchListeners, createListener, updateListener, deleteListener, reloadListener, exportNodeURI, Listener } from '../api/nodes';
import { useI18n } from '../i18n';

const PROTOCOLS = ['shadowsocks', 'snell', 'vmess', 'vless', 'trojan', 'hysteria2', 'tuic', 'shadowquic', 'anytls', 'mieru', 'sudoku', 'trusttunnel'];

const Listeners: React.FC = () => {
  const { t } = useI18n();
  const [data, setData] = useState<Listener[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Listener | null>(null);
  const [form] = Form.useForm();
  const [uris, setUris] = useState<string[]>([]);
  const [uriModal, setUriModal] = useState(false);

  const load = async () => {
    setLoading(true);
    try { const res = await fetchListeners(); setData(res); }
    catch (e: any) { message.error(e.message); }
    finally { setLoading(false); }
  };

  useEffect(() => { load(); }, []);

  const onSubmit = async (values: any) => {
    try {
      const payload = { ...values, config: values.config ? JSON.stringify(JSON.parse(values.config)) : '{}' };
      if (editing) { await updateListener(editing.id, payload); message.success(t('listeners.updated')); }
      else { await createListener(payload); message.success(t('listeners.created')); }
      setModalOpen(false); setEditing(null); form.resetFields(); load();
    } catch (e: any) { message.error(e.message); }
  };

  const onDelete = async (id: number) => {
    try { await deleteListener(id); message.success(t('listeners.deleted')); load(); }
    catch (e: any) { message.error(e.message); }
  };

  const onReload = async (id: number) => {
    try { await reloadListener(id); message.success(t('listeners.reloaded')); }
    catch (e: any) { message.error(e.message); }
  };

  const showURIs = async (id: number) => {
    try { const res = await exportNodeURI(id); setUris(res.uris); setUriModal(true); }
    catch (e: any) { message.error(e.message); }
  };

  const columns = [
    { title: t('listeners.name'), dataIndex: 'name', key: 'name' },
    { title: t('listeners.protocol'), dataIndex: 'protocol', key: 'protocol', render: (p: string) => <Tag>{p}</Tag> },
    { title: t('listeners.port'), dataIndex: 'port', key: 'port' },
    { title: t('listeners.status'), dataIndex: 'enabled', key: 'enabled', render: (v: boolean) => <Tag color={v ? 'success' : 'default'}>{v ? t('common.enabled') : t('common.disabled')}</Tag> },
    { title: t('common.actions'), key: 'actions', render: (_: any, record: Listener) => (
      <Space>
        <Tooltip title={t('listeners.copyURI')}><Button icon={<QrcodeOutlined />} onClick={() => showURIs(record.id)} /></Tooltip>
        <Tooltip title={t('common.refresh')}><Button icon={<ReloadOutlined />} onClick={() => onReload(record.id)} /></Tooltip>
        <Button icon={<EditOutlined />} onClick={() => { setEditing(record); form.setFieldsValue({ ...record, config: record.config ? JSON.stringify(JSON.parse(record.config), null, 2) : '{}' }); setModalOpen(true); }} />
        <Popconfirm title={t('listeners.deleteConfirm')} onConfirm={() => onDelete(record.id)}><Button icon={<DeleteOutlined />} danger /></Popconfirm>
      </Space>
    )},
  ];

  return (
    <div>
      <h2>{t('listeners.title')}</h2>
      <Card extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setModalOpen(true); }}>{t('listeners.create')}</Button>}>
        <Table dataSource={data} columns={columns} rowKey="id" loading={loading} />
      </Card>
      <Modal open={modalOpen} title={editing ? t('listeners.edit') : t('listeners.create')} onCancel={() => { setModalOpen(false); setEditing(null); form.resetFields(); }} onOk={() => form.submit()} width={640}>
        <Form form={form} layout="vertical" onFinish={onSubmit}>
          <Form.Item name="name" label={t('listeners.name')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="protocol" label={t('listeners.protocol')} rules={[{ required: true }]}>
            <Select>{PROTOCOLS.map((p) => <Select.Option key={p} value={p}>{p}</Select.Option>)}</Select>
          </Form.Item>
          <Form.Item name="port" label={t('listeners.port')} rules={[{ required: true }]}><Input type="number" /></Form.Item>
          <Form.Item name="bind_address" label={t('listeners.bindAddress')} initialValue="0.0.0.0"><Input /></Form.Item>
          <Form.Item name="enabled" label={t('listeners.status')} valuePropName="checked" initialValue={true}><Switch /></Form.Item>
          <Form.Item name="udp" label={t('listeners.udp')} valuePropName="checked" initialValue={false}><Switch /></Form.Item>
          <Form.Item name="tls" label={t('listeners.tls')} valuePropName="checked" initialValue={false}><Switch /></Form.Item>
          <Form.Item name="config" label={t('listeners.config')} rules={[{ validator: (_: any, v: string) => { if (!v) return Promise.resolve(); try { JSON.parse(v); return Promise.resolve(); } catch { return Promise.reject(new Error(t('listeners.invalidJSON'))); } } }]}>
            <Input.TextArea rows={4} />
          </Form.Item>
        </Form>
      </Modal>
      <Modal open={uriModal} title={t('listeners.urisTitle')} onCancel={() => setUriModal(false)} footer={null}>
        <Space direction="vertical" style={{ width: '100%' }}>
          {uris.map((uri, i) => (
            <Card key={i} size="small">
              <Space><Input value={uri} readOnly style={{ width: 400 }} /><Button icon={<CopyOutlined />} onClick={() => { navigator.clipboard.writeText(uri); message.success(t('common.copy')); }} /></Space>
            </Card>
          ))}
        </Space>
      </Modal>
    </div>
  );
};

export default Listeners;
