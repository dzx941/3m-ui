import { useState, useEffect } from 'react';
import { Button, Card, Col, Form, Input, InputNumber, Modal, Row, Select, Space, Table, Tag, DatePicker, message } from 'antd';
import { PlusOutlined, ReloadOutlined, EditOutlined, DeleteOutlined, LinkOutlined, CopyOutlined } from '@ant-design/icons';
import { apiRequest } from '../api/request';
import { ProtocolForm, type ProxyNode } from '../components/protocols';

const protocols = [
  ['shadowsocks', 'Shadowsocks'],
  ['vmess', 'VMess'],
  ['vless', 'VLESS'],
  ['trojan', 'Trojan'],
  ['hysteria2', 'Hysteria2'],
  ['tuic', 'TUIC'],
  ['wireguard', 'WireGuard'],
] as const;

export default function ProxyNodes() {
  const [items, setItems] = useState<ProxyNode[]>([]);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<number | null>(null);
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  // "Generate Client Link" states
  const [linkOpen, setLinkOpen] = useState(false);
  const [nodeIndexForLink, setNodeIndexForLink] = useState<number | null>(null);
  const [generatedLinks, setGeneratedLinks] = useState<{ clash: string; singbox: string; shadowrocket: string } | null>(null);
  const [linkForm] = Form.useForm();

  const load = async () => {
    setLoading(true);
    try {
      setItems(await apiRequest<ProxyNode[]>('/config/proxies'));
    } catch (e: any) {
      message.error(e.message || '加载代理节点失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const submit = async () => {
    const v = await form.validateFields();
    const proxy = {
      name: v.name,
      type: v.type,
      server: v.server,
      port: v.port,
      options: v.protocolConfig || {},
    };
    try {
      if (editing === null) {
        await apiRequest('/config/proxies', { method: 'POST', body: JSON.stringify(proxy) });
      } else {
        await apiRequest(`/config/proxies/${editing}`, { method: 'PUT', body: JSON.stringify(proxy) });
      }
      message.success('代理节点已保存');
      setOpen(false);
      form.resetFields();
      setEditing(null);
      void load();
    } catch (e: any) {
      message.error(e.message || '保存失败');
    }
  };

  const remove = async (i: number) => {
    try {
      await apiRequest(`/config/proxies/${i}`, { method: 'DELETE' });
      message.success('已删除');
      void load();
    } catch (e: any) {
      message.error(e.message || '删除失败');
    }
  };

  const handleOpenLinkModal = (index: number, name: string) => {
    setNodeIndexForLink(index);
    setGeneratedLinks(null);
    linkForm.resetFields();
    linkForm.setFieldsValue({ name: `${name} 接入 Token` });
    setLinkOpen(true);
  };

  const handleGenerateLink = async () => {
    try {
      const values = await linkForm.validateFields();
      const payload = {
        name: values.name,
        type: 'proxy',
        target_id: nodeIndexForLink,
        expire_at: values.expire_at ? values.expire_at.toISOString() : null,
      };

      const res = await apiRequest<{ token: string }>('/access-tokens', {
        method: 'POST',
        body: JSON.stringify(payload),
      });

      // Construct links dynamically
      const base = window.location.origin;
      setGeneratedLinks({
        clash: `${base}/api/v1/client/sub/${res.token}?target=clash`,
        singbox: `${base}/api/v1/client/sub/${res.token}?target=singbox`,
        shadowrocket: `${base}/api/v1/client/sub/${res.token}?target=shadowrocket`,
      });
      message.success('客户端接入 Token 生成成功');
    } catch (e: any) {
      message.error(e.message || '生成链接失败');
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    message.success('链接已复制到剪贴板');
  };

  return (
    <Card
      title="代理节点"
      extra={
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => void load()}>刷新</Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              setEditing(null);
              form.resetFields();
              form.setFieldValue('type', 'vless');
              setOpen(true);
            }}
          >
            添加节点
          </Button>
        </Space>
      }
    >
      <Table<ProxyNode>
        rowKey={(r) => `${r.name}-${r.server}-${r.port}`}
        loading={loading}
        dataSource={items}
        pagination={{ pageSize: 10 }}
        columns={[
          { title: '名称', dataIndex: 'name', render: (v) => <b>{v}</b> },
          { title: '协议', dataIndex: 'type', render: (v) => <Tag>{String(v).toUpperCase()}</Tag> },
          { title: '服务器', dataIndex: 'server' },
          { title: '端口', dataIndex: 'port' },
          {
            title: '操作',
            width: 250,
            render: (_, __, i) => {
              const x = items[i];
              return (
                <Space>
                  <Button
                    size="small"
                    icon={<LinkOutlined />}
                    onClick={() => handleOpenLinkModal(i, x.name)}
                  >
                    生成链接
                  </Button>
                  <Button
                    size="small"
                    icon={<EditOutlined />}
                    onClick={() => {
                      form.setFieldsValue({
                        name: x.name,
                        type: x.type,
                        server: x.server,
                        port: x.port,
                        protocolConfig: x.options || x,
                      });
                      setEditing(i);
                      setOpen(true);
                    }}
                  >
                    编辑
                  </Button>
                  <Button danger size="small" icon={<DeleteOutlined />} onClick={() => void remove(i)}>删除</Button>
                </Space>
              );
            },
          },
        ]}
      />
      <Modal
        title={editing === null ? '添加代理节点' : '编辑代理节点'}
        open={open}
        onCancel={() => setOpen(false)}
        onOk={() => void submit()}
        width={700}
        okText="保存"
        cancelText="取消"
      >
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="name" label="节点名称" rules={[{ required: true, message: '请输入节点名称' }]}>
                <Input placeholder="例如：香港 01" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="type" label="协议" rules={[{ required: true }]}>
                <Select options={protocols.map(([value, label]) => ({ value, label }))} />
              </Form.Item>
            </Col>
            <Col span={16}>
              <Form.Item name="server" label="服务器地址" rules={[{ required: true }]}>
                <Input placeholder="example.com" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="port" label="端口" rules={[{ required: true }]}>
                <InputNumber min={1} max={65535} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <ProtocolFields form={form} />
        </Form>
      </Modal>

      {/* Generate Client Link Modal */}
      <Modal
        title="生成客户端订阅链接"
        open={linkOpen}
        onCancel={() => setLinkOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form form={linkForm} layout="vertical" onFinish={handleGenerateLink} style={{ marginTop: 16 }}>
          <Form.Item name="name" label="链接备注 / 名称" rules={[{ required: true, message: '请输入备注' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="expire_at" label="过期时间 (留空永不过期)">
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Button type="primary" htmlType="submit" block style={{ marginBottom: 16 }}>
            开始生成
          </Button>
        </Form>

        {generatedLinks && (
          <Space direction="vertical" style={{ width: '100%', marginTop: 24, borderTop: '1px solid #eee', paddingTop: 16 }}>
            <div>
              <div style={{ fontWeight: 600, marginBottom: 4 }}>Mihomo / Clash 订阅:</div>
              <Input.Search
                enterButton={<Button icon={<CopyOutlined />}>复制</Button>}
                value={generatedLinks.clash}
                onSearch={() => copyToClipboard(generatedLinks.clash)}
                readOnly
              />
            </div>
            <div style={{ marginTop: 12 }}>
              <div style={{ fontWeight: 600, marginBottom: 4 }}>sing-box 订阅:</div>
              <Input.Search
                enterButton={<Button icon={<CopyOutlined />}>复制</Button>}
                value={generatedLinks.singbox}
                onSearch={() => copyToClipboard(generatedLinks.singbox)}
                readOnly
              />
            </div>
            <div style={{ marginTop: 12 }}>
              <div style={{ fontWeight: 600, marginBottom: 4 }}>Shadowrocket 订阅:</div>
              <Input.Search
                enterButton={<Button icon={<CopyOutlined />}>复制</Button>}
                value={generatedLinks.shadowrocket}
                onSearch={() => copyToClipboard(generatedLinks.shadowrocket)}
                readOnly
              />
            </div>
          </Space>
        )}
      </Modal>
    </Card>
  );
}

function ProtocolFields({ form }: { form: any }) {
  const type = Form.useWatch('type', form) || 'shadowsocks';
  return <ProtocolForm protocol={type} />;
}
