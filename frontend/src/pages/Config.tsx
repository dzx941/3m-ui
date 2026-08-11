import React, { useEffect, useState } from 'react';
import {
  Typography, Row, Col, Card, Button, Space, Form, Input, Select, Switch, Tabs,
  InputNumber, Divider, message, Alert,
} from 'antd';
import { SaveOutlined, DownloadOutlined, ReloadOutlined, SettingOutlined, PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { apiRequest, downloadFile } from '../api/request';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;

interface VisualDNS {
  enable: boolean;
  enhancedMode: string;
  listen: string;
  nameserver: string[];
  fallback?: string[];
}

interface VisualConfig {
  mode: string;
  logLevel: string;
  allowLan: boolean;
  ipv6: boolean;
  dns: VisualDNS;
  proxies: Array<{ name: string; type: string; server: string; port: number; options?: Record<string, unknown> }>;
  proxyGroups: Array<{ name: string; type: string; proxies: string[]; url?: string; interval?: number }>;
  rules: string | string[];
}

const defaults: VisualConfig = {
  mode: 'rule',
  logLevel: 'info',
  allowLan: true,
  ipv6: false,
  dns: {
    enable: true,
    enhancedMode: 'fake-ip',
    listen: '0.0.0.0:1053',
    nameserver: ['119.29.29.29', '223.5.5.5'],
    fallback: [],
  },
  proxies: [],
  proxyGroups: [],
  rules: 'GEOIP,CN,DIRECT\nMATCH,DIRECT',
};

const Config: React.FC = () => {
  const [form] = Form.useForm<VisualConfig>();
  const [loading, setLoading] = useState(false);
  const [initializing, setInitializing] = useState(true);
  const [preview, setPreview] = useState('');

  const loadConfig = async () => {
    setInitializing(true);
    try {
      const visual = await apiRequest<VisualConfig>('/config/visual');
      form.setFieldsValue({
        ...defaults,
        ...visual,
        dns: { ...defaults.dns, ...(visual.dns || {}) },
        rules: Array.isArray(visual.rules) ? visual.rules.join('\n') : visual.rules || '',
      });
      const generated = await apiRequest<{ config: string }>('/config');
      setPreview(generated.config);
    } catch (err) {
      message.error((err as { message?: string }).message || '加载配置失败');
    } finally {
      setInitializing(false);
    }
  };

  useEffect(() => {
    void loadConfig();
  }, []);

  const handleSave = async () => {
    const values = await form.validateFields();
    setLoading(true);
    try {
      const payload: VisualConfig = {
        ...defaults,
        ...values,
        dns: { ...defaults.dns, ...(values.dns || {}) },
        rules: typeof values.rules === 'string' ? values.rules.split(/\r?\n/).map(v => v.trim()).filter(Boolean) : values.rules || [],
        proxies: values.proxies || [],
        proxyGroups: values.proxyGroups || [],
      };
      await apiRequest('/config/visual', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
      await apiRequest('/config/generate', { method: 'POST' });
      const generated = await apiRequest<{ config: string }>('/config');
      setPreview(generated.config);
      message.success('配置已保存并生成');
    } catch (err) {
      message.error((err as { message?: string }).message || '配置保存失败');
    } finally {
      setLoading(false);
    }
  };

  const handleHotReload = async () => {
    setLoading(true);
    try {
      await apiRequest('/mihomo/restart', { method: 'POST' });
      message.success('Mihomo 已重新加载配置');
    } catch (err) {
      message.error((err as { message?: string }).message || 'Mihomo 重载失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDownload = async () => {
    try {
      await downloadFile('/config/download', 'config.yaml');
    } catch (err) {
      message.error((err as { message?: string }).message || '配置下载失败');
    }
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div>
          <Title level={2} style={{ margin: 0 }}>Mihomo Config Engine</Title>
          <Paragraph style={{ margin: 0 }}>
            可视化管理 Mihomo 配置。节点协议与入站认证在 Nodes 页面独立管理。
          </Paragraph>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => void loadConfig()} loading={initializing}>刷新</Button>
        </Space>
      </div>

      <Alert
        type="info"
        showIcon
        message="配置生成逻辑"
        description="页面保存的是可视化配置项，后端负责生成最终 config.yaml；Nodes 页面管理 listeners，Proxy Users 管理认证凭据。"
        style={{ marginBottom: 24 }}
      />

      <Row gutter={[24, 24]}>
        <Col xs={24} lg={16}>
          <Card title={<Space><SettingOutlined /><span>Visual Configuration Modules</span></Space>} bordered={false} loading={initializing}>
            <Form form={form} layout="vertical" initialValues={defaults}>
              <Tabs items={[
                {
                  key: 'general',
                  label: '基础',
                  children: (
                    <Row gutter={16}>
                      <Col xs={24} md={8}><Form.Item name="mode" label="运行模式"><Select options={['rule', 'global', 'direct', 'script'].map(value => ({ value, label: value }))} /></Form.Item></Col>
                      <Col xs={24} md={8}><Form.Item name="logLevel" label="日志级别"><Select options={['silent', 'error', 'warning', 'info', 'debug'].map(value => ({ value, label: value }))} /></Form.Item></Col>
                      <Col xs={12} md={4}><Form.Item name="allowLan" label="允许 LAN" valuePropName="checked"><Switch /></Form.Item></Col>
                      <Col xs={12} md={4}><Form.Item name="ipv6" label="IPv6" valuePropName="checked"><Switch /></Form.Item></Col>
                    </Row>
                  ),
                },
                {
                  key: 'dns',
                  label: 'DNS',
                  children: (
                    <>
                      <Row gutter={16}>
                        <Col xs={12} md={6}><Form.Item name={['dns', 'enable']} label="启用 DNS" valuePropName="checked"><Switch /></Form.Item></Col>
                        <Col xs={12} md={9}><Form.Item name={['dns', 'enhancedMode']} label="增强模式"><Select options={[{ value: 'fake-ip' }, { value: 'redir-host' }]} /></Form.Item></Col>
                        <Col xs={24} md={9}><Form.Item name={['dns', 'listen']} label="DNS 监听"><Input /></Form.Item></Col>
                      </Row>
                      <Form.Item name={['dns', 'nameserver']} label="Nameserver"><Select mode="tags" tokenSeparators={[',']} /></Form.Item>
                      <Form.Item name={['dns', 'fallback']} label="Fallback"><Select mode="tags" tokenSeparators={[',']} /></Form.Item>
                    </>
                  ),
                },
                {
                  key: 'proxy',
                  label: '代理节点',
                  children: (
                    <Form.List name="proxies">
                      {(fields, { add, remove }) => (
                        <Space direction="vertical" style={{ width: '100%' }}>
                          {fields.map(field => (
                            <Card key={field.key} size="small" title={`代理 ${field.name + 1}`} extra={<Button danger type="text" icon={<DeleteOutlined />} onClick={() => remove(field.name)} />}>
                              <Row gutter={12}>
                                <Col xs={24} md={6}><Form.Item {...field} name={[field.name, 'name']} label="名称" rules={[{ required: true }]}><Input /></Form.Item></Col>
                                <Col xs={24} md={6}><Form.Item {...field} name={[field.name, 'type']} label="协议" rules={[{ required: true }]}><Select options={['ss','vmess','vless','trojan','hysteria2','tuic','wireguard'].map(value => ({ value, label: value }))} /></Form.Item></Col>
                                <Col xs={24} md={6}><Form.Item {...field} name={[field.name, 'server']} label="服务器"><Input /></Form.Item></Col>
                                <Col xs={24} md={6}><Form.Item {...field} name={[field.name, 'port']} label="端口"><InputNumber min={1} max={65535} style={{ width: '100%' }} /></Form.Item></Col>
                              </Row>
                            </Card>
                          ))}
                          <Button icon={<PlusOutlined />} onClick={() => add({ type: 'vless', port: 443 })}>添加代理节点</Button>
                        </Space>
                      )}
                    </Form.List>
                  ),
                },
                {
                  key: 'group',
                  label: '代理组',
                  children: (
                    <Form.List name="proxyGroups">
                      {(fields, { add, remove }) => (
                        <Space direction="vertical" style={{ width: '100%' }}>
                          {fields.map(field => (
                            <Card key={field.key} size="small" title={`代理组 ${field.name + 1}`} extra={<Button danger type="text" icon={<DeleteOutlined />} onClick={() => remove(field.name)} />}>
                              <Row gutter={12}>
                                <Col xs={24} md={8}><Form.Item {...field} name={[field.name, 'name']} label="名称"><Input /></Form.Item></Col>
                                <Col xs={24} md={8}><Form.Item {...field} name={[field.name, 'type']} label="类型"><Select options={['select','url-test','fallback','load-balance','relay'].map(value => ({ value, label: value }))} /></Form.Item></Col>
                                <Col xs={24} md={8}><Form.Item {...field} name={[field.name, 'proxies']} label="成员"><Select mode="tags" tokenSeparators={[',']} /></Form.Item></Col>
                              </Row>
                            </Card>
                          ))}
                          <Button icon={<PlusOutlined />} onClick={() => add({ type: 'select', proxies: [] })}>添加代理组</Button>
                        </Space>
                      )}
                    </Form.List>
                  ),
                },
                {
                  key: 'rule',
                  label: '规则',
                  children: <Form.Item name="rules" label="每行一条规则"><TextArea rows={12} placeholder="DOMAIN-SUFFIX,example.com,DIRECT" /></Form.Item>,
                },
              ]} />

              <Divider />
              <Button type="primary" icon={<SaveOutlined />} loading={loading} onClick={() => void handleSave()}>
                保存并生成 config.yaml
              </Button>
            </Form>
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title="Config Control Center" bordered={false}>
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <Button icon={<DownloadOutlined />} block onClick={() => void handleDownload()}>下载 config.yaml</Button>
              <Button icon={<ReloadOutlined />} danger block loading={loading} onClick={() => void handleHotReload()}>重新加载 Mihomo</Button>
              <Text type="secondary">生成后的配置预览</Text>
              <TextArea value={preview} readOnly rows={24} style={{ fontFamily: 'monospace' }} />
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Config;
