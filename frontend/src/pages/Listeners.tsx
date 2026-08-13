import React, { useEffect, useState } from 'react';
import { Button, Form, Input, InputNumber, Modal, Popconfirm, Select, Space, Switch, Table, Tag, Typography, message } from 'antd';
import { CopyOutlined, DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { apiRequest } from '../api/request';
import { ProtocolForm } from '../components/protocols';
import { LISTENER_PROTOCOLS, type ListenerProtocol } from '../components/protocols/types';
import { useI18n } from '../i18n';

const { Title, Paragraph, Text } = Typography;

type NodeRecord = { ID: number; name: string; protocol: ListenerProtocol; port: number; bind_address: string; listen?: string; enabled: boolean; config: string; status: string };
type ClientAccess = { name: string; mihomo_link: string; clash_link: string; singbox_link: string; shadowrocket_link: string };
const protocolLabels: Record<ListenerProtocol, string> = { shadowsocks: 'Shadowsocks', snell: 'Snell', vmess: 'VMess', vless: 'VLESS', trojan: 'Trojan', hysteria2: 'Hysteria 2', tuic: 'TUIC V4/V5', shadowquic: 'ShadowQUIC', anytls: 'AnyTLS', mieru: 'Mieru', sudoku: 'Sudoku', trusttunnel: 'TrustTunnel' };
const mapUserProtocols = new Set<ListenerProtocol>(['anytls', 'hysteria2', 'mieru', 'tuic']);

const prune = (value: unknown): unknown => {
  if (Array.isArray(value)) return value.map(prune).filter((item) => item !== undefined);
  if (value && typeof value === 'object') {
    const out: Record<string, unknown> = {};
    for (const [key, raw] of Object.entries(value)) {
      if (raw === undefined || raw === null || raw === '') continue;
      const cleaned = prune(raw);
      if (cleaned === undefined) continue;
      if (typeof cleaned === 'object' && cleaned !== null && !Array.isArray(cleaned) && Object.keys(cleaned).length === 0) continue;
      if (Array.isArray(cleaned) && cleaned.length === 0) continue;
      out[key] = cleaned;
    }
    return out;
  }
  return value;
};

const serializeProtocolConfig = (protocol: ListenerProtocol, raw: Record<string, unknown>) => {
  const cfg = JSON.parse(JSON.stringify(prune(raw || {}))) as Record<string, unknown>;
  if (mapUserProtocols.has(protocol) && Array.isArray(cfg.users)) {
    const users: Record<string, string> = {};
    for (const row of cfg.users as Array<Record<string, unknown>>) {
      const username = String(row.username ?? row.uuid ?? '').trim();
      const password = String(row.password ?? '').trim();
      if (username && password) users[username] = password;
    }
    if (Object.keys(users).length) cfg.users = users; else delete cfg.users;
  }
  return cfg;
};
const hydrateProtocolConfig = (protocol: ListenerProtocol, raw: Record<string, unknown>) => {
  const cfg = JSON.parse(JSON.stringify(raw || {})) as Record<string, unknown>;
  if (mapUserProtocols.has(protocol) && cfg.users && typeof cfg.users === 'object' && !Array.isArray(cfg.users)) cfg.users = Object.entries(cfg.users as Record<string, unknown>).map(([username, password]) => ({ username, password }));
  return cfg;
};

const NodesPage: React.FC = () => {
  const { t, locale } = useI18n();
  const [rows, setRows] = useState<NodeRecord[]>([]); const [loading, setLoading] = useState(false); const [open, setOpen] = useState(false); const [editing, setEditing] = useState<NodeRecord | null>(null); const [access, setAccess] = useState<ClientAccess | null>(null); const [accessOpen, setAccessOpen] = useState(false); const [saving, setSaving] = useState(false); const [form] = Form.useForm();
  const protocol = Form.useWatch('protocol', form) as ListenerProtocol | undefined;

  const load = async () => { setLoading(true); try { setRows((await apiRequest<NodeRecord[]>('/nodes')) || []); } catch (error: unknown) { message.error(error instanceof Error ? error.message : t('nodes.loadFailed')); } finally { setLoading(false); } };
  useEffect(() => { void load(); }, []);
  const openCreate = () => { setEditing(null); form.resetFields(); form.setFieldsValue({ bind_address: '0.0.0.0', port: 443, protocol: 'vless', enabled: true, protocolConfig: {} }); setOpen(true); };
  const openEdit = (row: NodeRecord) => { let config: Record<string, unknown> = {}; try { config = hydrateProtocolConfig(row.protocol, JSON.parse(row.config || '{}')); } catch { message.error(t('nodes.invalidConfig')); return; } setEditing(row); form.resetFields(); form.setFieldsValue({ ...row, bind_address: row.bind_address || row.listen || '0.0.0.0', protocolConfig: config }); setOpen(true); };
  const generateClientAccess = async (id: number) => { try { const data = await apiRequest<ClientAccess>(`/nodes/${id}/client-access`, { method: 'POST' }); setAccess(data); setAccessOpen(true); } catch (error: unknown) { message.error(error instanceof Error ? error.message : t('nodes.clientUnavailable')); } };

  const save = async () => {
    if (saving) return; setSaving(true);
    try {
      const values = await form.validateFields(); const selected = values.protocol as ListenerProtocol;
      const payload = { name: String(values.name || '').trim(), protocol: selected, type: selected, port: Number(values.port), bind_address: String(values.bind_address || '0.0.0.0').trim(), listen: String(values.bind_address || '0.0.0.0').trim(), enabled: Boolean(values.enabled), status: values.enabled ? 'active' : 'inactive', config: JSON.stringify(serializeProtocolConfig(selected, values.protocolConfig || {})) };
      const id = editing?.ID; const result = await apiRequest<NodeRecord>(id ? `/nodes/${id}` : '/nodes', { method: id ? 'PUT' : 'POST', body: JSON.stringify(payload) });
      setOpen(false); await load(); message.success(id ? t('nodes.updated') : t('nodes.created')); if (!id && result?.ID) await generateClientAccess(result.ID);
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'errorFields' in error) { const first = (error as { errorFields?: Array<{ name?: Array<string | number>; errors?: string[] }> }).errorFields?.[0]; const field = first?.name?.join('.') || ''; const detail = first?.errors?.[0] || t('nodes.invalidConfig'); message.error(field ? `${field}: ${detail}` : detail); if (first?.name) form.scrollToField(first.name); }
      else if (error instanceof Error) message.error(error.message); else if (error && typeof error === 'object' && 'message' in error) message.error(String((error as { message: unknown }).message)); else message.error(t('nodes.saveFailed'));
    } finally { setSaving(false); }
  };
  const remove = async (id: number) => { try { await apiRequest(`/nodes/${id}`, { method: 'DELETE' }); await load(); message.success(t('nodes.deleted')); } catch (error: unknown) { message.error(error instanceof Error ? error.message : t('nodes.deleteFailed')); } };
  const toggle = async (row: NodeRecord, enabled: boolean) => { try { await apiRequest(`/nodes/${row.ID}`, { method: 'PUT', body: JSON.stringify({ ...row, enabled, status: enabled ? 'active' : 'inactive' }) }); await load(); } catch (error: unknown) { message.error(error instanceof Error ? error.message : t('nodes.updatedFailed')); } };
  const reload = async (id: number) => { try { await apiRequest(`/nodes/${id}/reload`, { method: 'POST' }); message.success(t('nodes.reloaded')); } catch (error: unknown) { message.error(error instanceof Error ? error.message : t('nodes.reloadFailed')); } };
  const copy = async (value: string) => { try { await navigator.clipboard.writeText(value); message.success(t('nodes.linkCopied')); } catch { message.error(t('nodes.copyFailed')); } };

  const title = locale === 'zh-CN' ? '节点' : 'Nodes';
  const subtitle = locale === 'zh-CN' ? '一个节点就是一个 Mihomo 入站配置，并自动生成对应的客户端代理配置。' : 'Each node is a Mihomo inbound configuration that automatically generates the corresponding client proxy configuration.';
  const listenerLabel = locale === 'zh-CN' ? '监听协议' : 'Listener Protocol';
  const clientHint = locale === 'zh-CN' ? '以下链接由当前节点自动生成，可直接提供给对应客户端。' : 'These links are generated from the current node and can be provided directly to the corresponding client.';

  const columns = [
    { title: t('nodes.name'), dataIndex: 'name', key: 'name' },
    { title: t('nodes.protocol'), dataIndex: 'protocol', key: 'protocol', render: (value: ListenerProtocol) => <Tag color="blue">{protocolLabels[value]}</Tag> },
    { title: t('nodes.listen'), key: 'listen', render: (_: unknown, row: NodeRecord) => `${row.bind_address || row.listen || '0.0.0.0'}:${row.port}` },
    { title: t('nodes.status'), dataIndex: 'enabled', key: 'enabled', render: (enabled: boolean, row: NodeRecord) => <Switch checked={enabled} onChange={(checked) => void toggle(row, checked)} /> },
    { title: t('nodes.actions'), key: 'actions', render: (_: unknown, row: NodeRecord) => <Space><Button type="text" icon={<CopyOutlined />} title={t('nodes.clientConfig')} onClick={() => void generateClientAccess(row.ID)} /><Button type="text" icon={<ReloadOutlined />} title={t('nodes.reload')} onClick={() => void reload(row.ID)} /><Button type="text" icon={<EditOutlined />} title={t('nodes.edit')} onClick={() => openEdit(row)} /><Popconfirm title={t('nodes.deleteConfirm')} onConfirm={() => void remove(row.ID)} okText={t('nodes.yes')} cancelText={t('nodes.no')}><Button type="text" danger icon={<DeleteOutlined />} /></Popconfirm></Space> },
  ];

  return <div>
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}><div><Title level={2} style={{ margin: 0 }}>{title}</Title><Paragraph style={{ margin: 0 }}>{subtitle}</Paragraph></div><Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>{t('nodes.create')}</Button></div>
    <Table rowKey="ID" loading={loading} dataSource={rows} columns={columns} scroll={{ x: 'max-content' }} />
    <Modal title={editing ? t('nodes.editTitle') : t('nodes.createTitle')} open={open} onOk={() => void save()} onCancel={() => setOpen(false)} confirmLoading={saving} destroyOnClose width={820}>
      <Form form={form} layout="vertical"><Form.Item name="name" label={t('nodes.name')} rules={[{ required: true, message: t('nodes.nameRequired') }]}><Input /></Form.Item>
        <Space style={{ display: 'flex' }} align="start"><Form.Item name="protocol" label={listenerLabel} rules={[{ required: true }]} style={{ width: 280 }}><Select options={LISTENER_PROTOCOLS.map((value) => ({ value, label: protocolLabels[value] }))} onChange={() => form.setFieldValue('protocolConfig', {})} /></Form.Item><Form.Item name="port" label={t('nodes.port')} rules={[{ required: true }]} style={{ width: 180 }}><InputNumber min={1} max={65535} style={{ width: '100%' }} /></Form.Item><Form.Item name="bind_address" label={t('nodes.listenAddress')} rules={[{ required: true }]} style={{ width: 220 }}><Input placeholder="0.0.0.0" /></Form.Item></Space>
        <ProtocolForm protocol={protocol} /><Form.Item name="enabled" label={t('nodes.enabled')} valuePropName="checked"><Switch /></Form.Item><Text type="secondary">{t('nodes.protocolHint')}</Text>
      </Form>
    </Modal>
    <Modal title={`${t('nodes.clientConfig')} — ${access?.name || ''}`} open={accessOpen} onCancel={() => setAccessOpen(false)} footer={null} width={760} destroyOnClose><Paragraph type="secondary">{clientHint}</Paragraph><Space direction="vertical" style={{ width: '100%' }}>{access && ([[t('nodes.mihomoClash'), access.clash_link], [t('nodes.singbox'), access.singbox_link], [t('nodes.shadowrocket'), access.shadowrocket_link], [t('nodes.rawMihomo'), access.mihomo_link]] as Array<[string, string]>).map(([label, link]) => <Space key={label} style={{ display: 'flex' }}><Text strong style={{ width: 130 }}>{label}</Text><Input value={link} readOnly style={{ minWidth: 430 }} /><Button icon={<CopyOutlined />} onClick={() => void copy(link)}>{t('nodes.copy')}</Button></Space>)}</Space></Modal>
  </div>;
};

export default NodesPage;
