import React, { useEffect, useState } from 'react';
import { Card, Table, Button, Space, Modal, Form, Input, Switch, message, Popconfirm } from 'antd';
import { PlusOutlined, DeleteOutlined, EditOutlined } from '@ant-design/icons';
import { fetchUsers, createUser, updateUser, deleteUser, ProxyUser } from '../api/users';
import { useI18n } from '../i18n';

const Users: React.FC = () => {
  const { t } = useI18n();
  const [data, setData] = useState<ProxyUser[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<ProxyUser | null>(null);
  const [form] = Form.useForm();

  const load = async () => {
    setLoading(true);
    try { const res = await fetchUsers(); setData(res); }
    catch (e: any) { message.error(e.message || t('common.error')); }
    finally { setLoading(false); }
  };

  useEffect(() => { load(); }, []);

  const onSubmit = async (values: any) => {
    try {
      if (editing) { await updateUser(editing.id, values); message.success(t('users.updated')); }
      else { await createUser(values); message.success(t('users.created')); }
      setModalOpen(false); setEditing(null); form.resetFields(); load();
    } catch (e: any) { message.error(e.message || t('common.error')); }
  };

  const onDelete = async (id: number) => {
    try { await deleteUser(id); message.success(t('users.deleted')); load(); }
    catch (e: any) { message.error(e.message || t('common.error')); }
  };

  const columns = [
    { title: t('users.username'), dataIndex: 'username', key: 'username' },
    { title: t('common.status'), dataIndex: 'enabled', key: 'enabled', render: (v: boolean) => (v ? t('common.enabled') : t('common.disabled')) },
    { title: t('common.actions'), key: 'actions', render: (_: any, record: ProxyUser) => (
      <Space>
        <Button icon={<EditOutlined />} onClick={() => { setEditing(record); form.setFieldsValue(record); setModalOpen(true); }} />
        <Popconfirm title={t('users.deleteConfirm')} onConfirm={() => onDelete(record.id)}><Button icon={<DeleteOutlined />} danger /></Popconfirm>
      </Space>
    )},
  ];

  return (
    <div>
      <h2>{t('users.title')}</h2>
      <Card extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setModalOpen(true); }}>{t('users.create')}</Button>}>
        <Table scroll={{ x: 480 }} size="middle" dataSource={data} columns={columns} rowKey="id" loading={loading} />
      </Card>
      <Modal open={modalOpen} title={editing ? t('users.edit') : t('users.create')} onCancel={() => { setModalOpen(false); setEditing(null); form.resetFields(); }} onOk={() => form.submit()} destroyOnClose>
        <Form form={form} layout="vertical" onFinish={onSubmit}>
          <Form.Item name="username" label={t('users.username')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="password" label={t('users.password')} rules={[{ required: !editing }]}><Input.Password placeholder={editing ? t('common.empty') : ''} /></Form.Item>
          <Form.Item name="enabled" label={t('users.enabled')} valuePropName="checked" initialValue={true}><Switch /></Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Users;
