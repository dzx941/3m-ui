import { useState, useEffect } from 'react';
import {
  Typography,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  QRCode,
  Input,
  DatePicker,
  Select,
  Switch,
  message,
  Popconfirm,
} from 'antd';
import { PlusOutlined, CopyOutlined, QrcodeOutlined, DeleteOutlined } from '@ant-design/icons';
import { apiRequest } from '../api/request';
import { useI18n } from '../i18n';
import dayjs from 'dayjs';

const { Title, Paragraph } = Typography;

interface AccessToken {
  ID: number;
  name: string;
  token: string;
  enabled: boolean;
  expire_at: string | null;
  type: 'user' | 'proxy';
  target_id: number;
  mihomo_link: string;
  clash_link: string;
  singbox_link: string;
  shadowrocket_link: string;
}

interface ProxyUser {
  id: number;
  username: string;
}

interface ProxyNode {
  name: string;
}

const ClientAccessPage: React.FC = () => {
  const { t } = useI18n();
  const [tokens, setTokens] = useState<AccessToken[]>([]);
  const [users, setUsers] = useState<ProxyUser[]>([]);
  const [proxyNodes, setProxyNodes] = useState<ProxyNode[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [qrOpen, setQrOpen] = useState(false);
  const [qrUrl, setQrUrl] = useState('');
  const [form] = Form.useForm();

  const selectedType = Form.useWatch('type', form);

  const fetchTokens = async () => {
    setLoading(true);
    try {
      const data = await apiRequest<AccessToken[]>('/access-tokens');
      setTokens(data || []);
    } catch (e: any) {
      message.error(e.message || '加载 Token 失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchUsersAndNodes = async () => {
    try {
      const userData = await apiRequest<ProxyUser[]>('/users');
      setUsers(userData || []);
      const nodeData = await apiRequest<ProxyNode[]>('/config/proxies');
      setProxyNodes(nodeData || []);
    } catch {
      // Ignored for fallback
    }
  };

  useEffect(() => {
    void fetchTokens();
    void fetchUsersAndNodes();
  }, []);

  const handleCreateToken = async () => {
    try {
      const values = await form.validateFields();
      const payload = {
        name: values.name,
        type: values.type,
        target_id: Number(values.target_id),
        expire_at: values.expire_at ? values.expire_at.toISOString() : null,
      };

      await apiRequest('/access-tokens', {
        method: 'POST',
        body: JSON.stringify(payload),
      });

      message.success('Token 创建成功');
      setModalOpen(false);
      form.resetFields();
      void fetchTokens();
    } catch (e: any) {
      message.error(e.message || '创建 Token 失败');
    }
  };

  const handleToggleEnabled = async (id: number, checked: boolean) => {
    try {
      await apiRequest(`/access-tokens/${id}`, {
        method: 'PUT',
        body: JSON.stringify({ enabled: checked }),
      });
      message.success(checked ? 'Token 已启用' : 'Token 已禁用');
      void fetchTokens();
    } catch (e: any) {
      message.error(e.message || '操作失败');
    }
  };

  const handleDeleteToken = async (id: number) => {
    try {
      await apiRequest(`/access-tokens/${id}`, {
        method: 'DELETE',
      });
      message.success('Token 已删除');
      void fetchTokens();
    } catch (e: any) {
      message.error(e.message || '删除失败');
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    message.success('链接已复制到剪贴板');
  };

  const showQrCode = (url: string) => {
    setQrUrl(url);
    setQrOpen(true);
  };

  const getTargetName = (type: string, id: number) => {
    if (type === 'user') {
      const found = users.find((u) => u.id === id);
      return found ? `用户: ${found.username}` : `用户 ID: ${id}`;
    } else {
      const found = proxyNodes[id];
      return found ? `节点: ${found.name}` : `节点索引: ${id}`;
    }
  };

  const columns = [
    {
      title: 'Token 名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: 'user' | 'proxy') => (
        <Tag color={type === 'user' ? 'blue' : 'green'}>
          {type === 'user' ? '入站用户' : '代理节点'}
        </Tag>
      ),
    },
    {
      title: '绑定目标',
      key: 'target',
      render: (_: any, r: AccessToken) => getTargetName(r.type, r.target_id),
    },
    {
      title: '过期时间',
      dataIndex: 'expire_at',
      key: 'expire_at',
      render: (v: string | null) => (v ? dayjs(v).format('YYYY-MM-DD HH:mm') : '永不过期'),
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean, r: AccessToken) => (
        <Switch checked={enabled} onChange={(checked) => void handleToggleEnabled(r.ID, checked)} />
      ),
    },
    {
      title: '客户端订阅配置',
      key: 'links',
      render: (_: any, r: AccessToken) => {
        if (!r.enabled) return <span style={{ color: '#ccc' }}>Token 已禁用</span>;
        return (
          <Space direction="vertical" size="small" style={{ width: '100%' }}>
            <div>
              <span style={{ fontWeight: 600, marginRight: 8 }}>Mihomo / Clash:</span>
              <Space>
                <Button size="small" icon={<CopyOutlined />} onClick={() => copyToClipboard(r.clash_link)}>复制</Button>
                <Button size="small" icon={<QrcodeOutlined />} onClick={() => showQrCode(r.clash_link)}>二维码</Button>
              </Space>
            </div>
            <div>
              <span style={{ fontWeight: 600, marginRight: 8 }}>sing-box:</span>
              <Space>
                <Button size="small" icon={<CopyOutlined />} onClick={() => copyToClipboard(r.singbox_link)}>复制</Button>
                <Button size="small" icon={<QrcodeOutlined />} onClick={() => showQrCode(r.singbox_link)}>二维码</Button>
              </Space>
            </div>
            <div>
              <span style={{ fontWeight: 600, marginRight: 8 }}>Shadowrocket:</span>
              <Space>
                <Button size="small" icon={<CopyOutlined />} onClick={() => copyToClipboard(r.shadowrocket_link)}>复制</Button>
                <Button size="small" icon={<QrcodeOutlined />} onClick={() => showQrCode(r.shadowrocket_link)}>二维码</Button>
              </Space>
            </div>
          </Space>
        );
      },
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, r: AccessToken) => (
        <Popconfirm title="确定删除此 Token 吗?" onConfirm={() => void handleDeleteToken(r.ID)}>
          <Button danger type="text" icon={<DeleteOutlined />}>删除</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div>
          <Title level={2} style={{ margin: 0 }}>{t('clientAccess.title')}</Title>
          <Paragraph style={{ margin: 0 }}>{t('clientAccess.subtitle')}</Paragraph>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => {
          form.resetFields();
          form.setFieldsValue({ type: 'user' });
          setModalOpen(true);
        }}>
          生成客户端 Token
        </Button>
      </div>

      <Table dataSource={tokens} columns={columns} rowKey="ID" loading={loading} />

      <Modal
        title="生成客户端 Token"
        open={modalOpen}
        onOk={() => void handleCreateToken()}
        onCancel={() => setModalOpen(false)}
        destroyOnClose
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="name" label="Token 名称 / 备注" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="e.g. 我的手机订阅" />
          </Form.Item>

          <Form.Item name="type" label="Token 类型" rules={[{ required: true }]}>
            <Select options={[
              { value: 'user', label: '绑定入站用户 (管理 Listener 接入)' },
              { value: 'proxy', label: '绑定特定代理节点 (转发外部 Proxy)' },
            ]} />
          </Form.Item>

          {selectedType === 'user' ? (
            <Form.Item name="target_id" label="绑定 Proxy 用户" rules={[{ required: true, message: '请选择用户' }]}>
              <Select placeholder="选择代理用户">
                {users.map((u) => (
                  <Select.Option key={u.id} value={u.id}>{u.username}</Select.Option>
                ))}
              </Select>
            </Form.Item>
          ) : (
            <Form.Item name="target_id" label="绑定外部代理节点" rules={[{ required: true, message: '请选择节点' }]}>
              <Select placeholder="选择代理节点">
                {proxyNodes.map((p, idx) => (
                  <Select.Option key={idx} value={idx}>{p.name}</Select.Option>
                ))}
              </Select>
            </Form.Item>
          )}

          <Form.Item name="expire_at" label="过期时间 (留空永不过期)">
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
        bodyStyle={{ textAlign: 'center', padding: '24px' }}
      >
        <QRCode value={qrUrl} size={250} bordered />
      </Modal>
    </div>
  );
};

export default ClientAccessPage;
