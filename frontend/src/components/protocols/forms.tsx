import { Form, Input, InputNumber, Select, Switch } from 'antd';

export const ShadowsocksForm = () => (
  <>
    <Form.Item name={['protocolConfig', 'cipher']} label="加密算法">
      <Select options={[{ value: 'aes-256-gcm' }, { value: 'chacha20-ietf-poly1305' }, { value: '2022-blake3-aes-256-gcm' }]} />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'password']} label="密码">
      <Input.Password />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'plugin']} label="插件">
      <Input />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'plugin-opts']} label="插件参数">
      <Input.TextArea rows={2} />
    </Form.Item>
  </>
);

export const VmessForm = () => (
  <>
    <Form.Item name={['protocolConfig', 'uuid']} label="UUID">
      <Input />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'alterId']} label="Alter ID">
      <InputNumber min={0} style={{ width: '100%' }} />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'cipher']} label="加密算法">
      <Select options={[{ value: 'auto' }, { value: 'zero' }, { value: 'aes-128-gcm' }, { value: 'chacha20-poly1305' }]} />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'network']} label="传输方式">
      <Select options={[{ value: 'tcp' }, { value: 'ws' }, { value: 'grpc' }, { value: 'h2' }]} />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'tls']} label="TLS" valuePropName="checked">
      <Switch />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'servername']} label="服务器名称">
      <Input />
    </Form.Item>
  </>
);

export const VlessForm = () => (
  <>
    <Form.Item name={['protocolConfig', 'uuid']} label="UUID">
      <Input />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'flow']} label="Flow">
      <Input placeholder="xtls-rprx-vision" />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'encryption']} label="加密">
      <Input placeholder="none" />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'reality']} label="Reality" valuePropName="checked">
      <Switch />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'tls']} label="TLS" valuePropName="checked">
      <Switch />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'network']} label="传输方式">
      <Select options={[{ value: 'tcp' }, { value: 'ws' }, { value: 'grpc' }]} />
    </Form.Item>
  </>
);

export const TrojanForm = () => (
  <>
    <Form.Item name={['protocolConfig', 'password']} label="密码">
      <Input.Password />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'sni']} label="SNI">
      <Input />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'tls']} label="TLS" valuePropName="checked">
      <Switch />
    </Form.Item>
  </>
);

export const Hysteria2Form = () => (
  <>
    <Form.Item name={['protocolConfig', 'password']} label="密码">
      <Input.Password />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'obfs']} label="混淆">
      <Input />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'up']} label="上传速度">
      <Input placeholder="100 Mbps" />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'down']} label="下载速度">
      <Input placeholder="100 Mbps" />
    </Form.Item>
  </>
);

export const WireGuardForm = () => (
  <>
    <Form.Item name={['protocolConfig', 'private-key']} label="私钥">
      <Input.Password />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'public-key']} label="公钥">
      <Input />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'ip']} label="IP 地址">
      <Input />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'mtu']} label="MTU">
      <InputNumber min={576} max={9000} style={{ width: '100%' }} />
    </Form.Item>
  </>
);

export const TuicForm = () => (
  <>
    <Form.Item name={['protocolConfig', 'uuid']} label="UUID">
      <Input />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'password']} label="密码">
      <Input.Password />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'congestion-controller']} label="拥塞控制">
      <Select options={[{ value: 'cubic' }, { value: 'bbr' }, { value: 'new_reno' }]} />
    </Form.Item>
  </>
);
