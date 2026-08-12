import React, { useEffect, useState } from 'react';
import {
  Typography,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  DatePicker,
  Select,
  Switch,
  message,
  Popconfirm,
} from 'antd';
import {
  PlusOutlined,
  CopyOutlined,
  QrcodeOutlined,
  DeleteOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { apiRequest } from '../api/request';
import { useI18n } from '../i18n';

const { Title, Paragraph } = Typography;

interface Listener {
  id: number;
  ID?: number;
  name: string;
  protocol: string;
  port: number;
  enabled?: boolean;
}

interface AccessToken {
  ID: number;
  name: string;
  enabled: boolean;
  expire_at: string | null;
  listener_id: number;
  listener_name?: string;
  listener_protocol?: string;
  mihomo_link: string;
  clash_link: string;
  singbox_link: string;
  shadowrocket_link: string;
}

const ClientAccessPage: React.FC = () => {
  const { t } = useI18n();
  const [tokens, setTokens] = useState<AccessToken[]>([]);
  const [listeners, setListeners] = useState<Listener[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [qrOpen, setQrOpen] = useState(false);
  const [qrUrl, setQrUrl] = useState('');
  const [form] = Form.useForm();

  const fetchTokens = async () => {
    setLoading(true);
    try {
      const data = await apiRequest<AccessToken[]>('/access-tokens');
      setTokens(data || []);
    } catch (error: unknown) {
      message.error(error instanceof Error ? error.message : '加载客户端接入失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchListeners = async () => {
    try {
      const data = await apiRequest<Listener[]>('/listeners');
      setListeners(data || []);
    } catch (error) {
      console.error('[ClientAccess] failed to load listeners', error);
    }
  };

  useEffect(() => {
    void fetchTokens();
    void fetchListeners();
  }, []);

  const handleCreateToken = async () => {
    try {
      const values = await form.validateFields();
      await apiRequest('/access-tokens', {
        method: 'POST',
        body: JSON.stringify({
          name: values.name,
          listener_id: Number(values.listener_id),
          expire_at: values.expire_at ? values.expire_at.toISOString() : null,
        }),
      });
      message.success(t('clientAccess.tokenCreated') || '客户端接入 Token 创建成功');
      setModalOpen(false);
      form.resetFields();
      await fetchTokens();
    } catch (error: unknown) {
      message.error(error instanceof Error ? error.message : '创建 Token 失败');
    }
  };

  const handleToggleEnabled = async (id: number, enabled: boolean) => {
    try {
      await apiRequest(`/access-tokens/${id}`, {
        method: 'PUT',
        body: JSON.stringify({ enabled }),
      });
      message.success(enabled ? 'Token 已启用' : 'Token 已禁用');
      await fetchTokens();
    } catch (error: unknown) {
      message.error(error instanceof Error ? error.message : '操作失败');
    }
  };

  const handleDeleteToken = async (id: number) => {
    try {
      await apiRequest(`/access-tokens/${id}`, { method: 'DELETE' });
      message.success('Token 已删除');
      await fetchTokens();
    } catch (error: unknown) {
      message.error(error instanceof Error ? error.message : '删除失败');
    }
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      message.success('链接已复制到剪贴板');
    } catch {
      message.error('复制失败，请手动复制链接');
    }
  };

  const showQrCode = (url: string) => {
    if (!url) {
      message.warning('客户端链接不可用');
      return;
    }
    setQrUrl(
      `https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=${encodeURIComponent(url)}`,
    );
    setQrOpen(true);
  };

  const getListenerName = (record: AccessToken) => {
    if (record.listener_name) return record.listener_name;
    const listener = listeners.find(
      (item) => (item.id ?? item.ID) === record.listener_id,
    );
    return listener?.name || `Listener #${record.listener_id}`;
  };

  const columns = [
    {
      title: 'Token 名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Listener',
      key: 'listener',
      render: (_: unknown, record: AccessToken) => (
        <Space direction="vertical" size={0}>
          <span>{getListenerName(record)}</span>
          <span style={{ color: '#888', fontSize: 12 }}>
            {record.listener_protocol || 'unknown'}
          </span>
        </Space>
      ),
    },
    {
      title: '过期时间',
      dataIndex: 'expire_at',
      key: 'expire_at',
      render: (value: string | null) =>
        value ? dayjs(value).format('YYYY-MM-DD HH:mm') : '永不过期',
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean, record: AccessToken) => (
        <Switch
          checked={enabled}
          onChange={(checked) => void handleToggleEnabled(record.ID, checked)}
        />
      ),
    },
    {
      title: '客户端配置',
      key: 'links',
      render: (_: unknown, record: AccessToken) => {
        if (!record.enabled) {
          return <Tag>{t('clientAccess.disabled') || '已禁用'}</Tag>;
        }
        const links = [
          ['Mihomo / Clash', record.clash_link],
          ['sing-box', record.singbox_link],
          ['Shadowrocket', record.shadowrocket_link],
        ].filter(([, url]) => Boolean(url));

        return (
          <Space direction="vertical" size="small" style={{ width: '100%' }}>
            {links.map(([label, url]) => (
              <div key={label}>
                <span style={{ fontWeight: 600, marginRight: 8 }}>{label}:</span>
                <Space>
                  <Button
                    size="small"
                    icon={<CopyOutlined />}
                    onClick={() => void copyToClipboard(url)}
                  >
                    复制
                  </Button>
                  <Button
                    size="small"
                    icon={<QrcodeOutlined />}
                    onClick={() => showQrCode(url)}
                  >
                    二维码
                  </Button>
                </Space>
              </div>
            ))}
          </Space>
        );
      },
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: unknown, record: AccessToken) => (
        <Popconfirm
          title="确定删除此 Token 吗？"
          onConfirm={() => void handleDeleteToken(record.ID)}
        >
          <Button danger type="text" icon={<DeleteOutlined />}>
            删除
          </Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 24,
        }}
      >
        <div>
          <Title level={2} style={{ margin: 0 }}>
            {t('clientAccess.title')}
          </Title>
          <Paragraph style={{ margin: 0 }}>
            {t('clientAccess.subtitle')}
          </Paragraph>
        </div>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => {
            form.resetFields();
            setModalOpen(true);
          }}
        >
          {t('clientAccess.createToken') || '生成客户端 Token'}
        </Button>
      </div>

      <Table dataSource={tokens} columns={columns} rowKey="ID" loading={loading} />

      <Modal
        title={t('clientAccess.createToken') || '生成客户端 Token'}
        open={modalOpen}
        onOk={() => void handleCreateToken()}
        onCancel={() => setModalOpen(false)}
        destroyOnClose
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="name"
            label="Token 名称 / 备注"
            rules={[{ required: true, message: '请输入名称' }]}
          >
            <Input placeholder="例如：我的手机" />
          </Form.Item>

          <Form.Item
            name="listener_id"
            label="Listener"
            rules={[{ required: true, message: '请选择 Listener' }]}
          >
            <Select placeholder="选择要导出的 Listener">
              {listeners.map((listener) => {
                const id = listener.id ?? listener.ID;
                return (
                  <Select.Option key={id} value={id}>
                    {listener.name} — {listener.protocol}:{listener.port}
                  </Select.Option>
                );
              })}
            </Select>
          </Form.Item>

          <Form.Item name="expire_at" label="过期时间（留空永不过期）">
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="扫描二维码订阅"
        open={qrOpen}
        onCancel={() => setQrOpen(false)}
        footer={null}
        destroyOnClose
        width={300}
      >
        <div style={{ textAlign: 'center', padding: 24 }}>
          {qrUrl && (
            <img
              src={qrUrl}
              alt="QR Code"
              style={{ maxWidth: '100%', height: 'auto', display: 'block', margin: '0 auto' }}
            />
          )}
        </div>
      </Modal>
    </div>
  );
};

export default ClientAccessPage;
