import React, { useState } from 'react';
import { Card, Form, Input, Select, Button, message, Space, Alert } from 'antd';
import { ThunderboltOutlined } from '@ant-design/icons';
import { createListener } from '../api/nodes';
import { useI18n } from '../i18n';

const PROTOCOLS = ['shadowsocks', 'vmess', 'vless', 'trojan', 'hysteria2', 'tuic'];

const InboundTemplates: React.FC = () => {
  const { t } = useI18n();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: any) => {
    setLoading(true);
    try { await createListener({ ...values, enabled: true, config: '{}' }); message.success(t('listeners.created')); form.resetFields(); }
    catch (e: any) { message.error(e.message || t('common.error')); }
    finally { setLoading(false); }
  };

  return (
    <div>
      <h2>{t('templates.title')}</h2>
      <Alert message={t('config.subtitle')} type="info" showIcon style={{ marginBottom: 16 }} />
      <Card>
        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Form.Item name="name" label={t('listeners.name')} rules={[{ required: true }]}><Input placeholder="e.g. hk-ss-01" /></Form.Item>
          <Form.Item name="protocol" label={t('listeners.protocol')} rules={[{ required: true }]} initialValue="shadowsocks">
            <Select>{PROTOCOLS.map((p) => <Select.Option key={p} value={p}>{p}</Select.Option>)}</Select>
          </Form.Item>
          <Form.Item name="port" label={t('listeners.port')} rules={[{ required: true }]}><Input type="number" placeholder="e.g. 8388" /></Form.Item>
          <Form.Item name="bind_address" label={t('listeners.bindAddress')} initialValue="0.0.0.0"><Input /></Form.Item>
          <Button type="primary" htmlType="submit" icon={<ThunderboltOutlined />} loading={loading}>{t('common.create')}</Button>
        </Form>
      </Card>
    </div>
  );
};

export default InboundTemplates;
