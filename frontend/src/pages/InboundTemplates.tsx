import React, { useEffect, useState } from 'react';
import { Alert, Button, Card, Form, Input, InputNumber, Select, Space, message } from 'antd';
import client from '../api/client';

type Field = { name: string; label: string; required: boolean; secret: boolean; default?: string };
type Template = { id: string; name: string; description: string; protocol: string; fields: Field[] };

const InboundTemplates: React.FC = () => {
  const [templates, setTemplates] = useState<Template[]>([]);
  const [selected, setSelected] = useState<string>();
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  useEffect(() => {
    client.get('/inbound-templates').then(({ data }) => setTemplates(data.templates || [])).catch((err) => message.error(err.message));
  }, []);

  const template = templates.find((item) => item.id === selected);

  const submit = async (values: Record<string, string>) => {
    if (!template) return;
    setLoading(true);
    try {
      const { data } = await client.post(`/inbound-templates/${template.id}/create`, {
        template_id: template.id,
        name: values.name,
        port: String(values.port || '443'),
        bind: values.bind || '0.0.0.0',
        username: values.username || 'user1',
        password: values.password,
        uuid: values.uuid,
        values: Object.fromEntries(template.fields.map((field) => [field.name, values[field.name] || ''])),
      });
      message.success(`节点「${data.listener.name}」创建成功`);
      form.resetFields();
    } catch (err: any) {
      message.error(err.message || '创建失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <div>
        <h2 style={{ marginBottom: 4 }}>快速创建节点</h2>
        <div style={{ color: '#666' }}>基于 Mihomo Inbound 模板一键创建 Listener、凭据并应用配置。</div>
      </div>
      <Card>
        <Form layout="vertical" form={form} onFinish={submit} initialValues={{ port: 443, bind: '0.0.0.0', username: 'user1' }}>
          <Form.Item label="节点模板" required>
            <Select placeholder="选择模板" value={selected} onChange={(value) => { setSelected(value); form.resetFields(); }} options={templates.map((item) => ({ value: item.id, label: `${item.name} — ${item.description}` }))} />
          </Form.Item>
          {template && <Alert style={{ marginBottom: 16 }} type="info" message={`协议：${template.protocol}`} description="模板参数会写入 Listener 配置；VLESS 的 UUID/Reality 密钥可以自动生成。证书、Decryption、ECH 等外部材料仍需提前准备。" />}
          <Space wrap style={{ width: '100%' }}>
            <Form.Item name="name" label="节点名称" rules={[{ required: true, message: '请输入节点名称' }]}><Input style={{ width: 220 }} /></Form.Item>
            <Form.Item name="port" label="端口" rules={[{ required: true }]}><InputNumber min={1} max={65535} style={{ width: 140 }} /></Form.Item>
            <Form.Item name="bind" label="监听地址"><Input style={{ width: 180 }} /></Form.Item>
            <Form.Item name="username" label="用户名"><Input style={{ width: 180 }} /></Form.Item>
            {template?.protocol !== 'vless' && template && <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}><Input.Password style={{ width: 220 }} /></Form.Item>}
          </Space>
          {template?.protocol === 'vless' && <Form.Item name="uuid" label="UUID（留空自动生成）"><Input style={{ maxWidth: 420 }} /></Form.Item>}
          {template?.fields.map((field) => (
            <Form.Item key={field.name} name={field.name} label={field.label} initialValue={field.default === 'auto' ? undefined : field.default} rules={[{ required: field.required, message: `请输入${field.label}` }]}>
              {field.secret ? <Input.Password placeholder={field.default === 'auto' ? '留空自动生成' : undefined} /> : <Input />}
            </Form.Item>
          ))}
          <Button type="primary" htmlType="submit" loading={loading} disabled={!template}>一键创建并应用</Button>
        </Form>
      </Card>
    </Space>
  );
};

export default InboundTemplates;
