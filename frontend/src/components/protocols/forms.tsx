import { Form, Input, InputNumber, Select } from 'antd';

const ServerTLSFields = () => (
  <>
    <Form.Item name={['protocolConfig', 'certificate']} label="Certificate">
      <Input placeholder="/etc/3m-ui/certs/server.crt" />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'private-key']} label="Private Key">
      <Input.Password placeholder="/etc/3m-ui/certs/server.key" />
    </Form.Item>
  </>
);

export const ShadowsocksForm = () => (
  <>
    <Form.Item name={['protocolConfig', 'cipher']} label="Cipher">
      <Select options={[
        { value: '2022-blake3-aes-128-gcm' },
        { value: '2022-blake3-aes-256-gcm' },
        { value: '2022-blake3-chacha20-poly1305' },
        { value: 'aes-128-gcm' },
        { value: 'aes-192-gcm' },
        { value: 'aes-256-gcm' },
        { value: 'chacha20-ietf-poly1305' },
        { value: 'none' },
      ]} />
    </Form.Item>
  </>
);

export const VmessForm = () => (
  <>
    <Form.Item name={['protocolConfig', 'ws-path']} label="WebSocket Path">
      <Input placeholder="/" />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'grpc-service-name']} label="gRPC Service Name">
      <Input placeholder="GunService" />
    </Form.Item>
    <ServerTLSFields />
  </>
);

export const VlessForm = () => (
  <>
    <Form.Item name={['protocolConfig', 'ws-path']} label="WebSocket Path">
      <Input placeholder="/" />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'grpc-service-name']} label="gRPC Service Name">
      <Input placeholder="GunService" />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'decryption']} label="Decryption">
      <Input placeholder="Optional VLESS encryption server configuration" />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'allow-insecure']} label="Allow Insecure">
      <Select options={[{ value: true, label: 'Enabled' }, { value: false, label: 'Disabled' }]} />
    </Form.Item>
    <ServerTLSFields />
    <Form.Item name={['protocolConfig', 'reality-config']} label="Reality Config (JSON)">
      <Input.TextArea rows={5} placeholder='{"dest":"example.com:443","private-key":"...","short-id":["0123456789abcdef"],"server-names":["example.com"]}' />
    </Form.Item>
  </>
);

export const TrojanForm = () => (
  <>
    <Form.Item name={['protocolConfig', 'ws-path']} label="WebSocket Path">
      <Input placeholder="/" />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'grpc-service-name']} label="gRPC Service Name">
      <Input placeholder="GunService" />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'allow-insecure']} label="Allow Insecure">
      <Select options={[{ value: true, label: 'Enabled' }, { value: false, label: 'Disabled' }]} />
    </Form.Item>
    <ServerTLSFields />
    <Form.Item name={['protocolConfig', 'reality-config']} label="Reality Config (JSON)">
      <Input.TextArea rows={5} placeholder='{"dest":"example.com:443","private-key":"...","short-id":["0123456789abcdef"],"server-names":["example.com"]}' />
    </Form.Item>
  </>
);

export const Hysteria2Form = () => (
  <>
    <Form.Item name={['protocolConfig', 'up']} label="Upload Speed">
      <Input placeholder="1000" />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'down']} label="Download Speed">
      <Input placeholder="1000" />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'obfs']} label="Obfuscation">
      <Select options={[{ value: 'salamander' }]} allowClear />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'obfs-password']} label="Obfuscation Password">
      <Input.Password />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'masquerade']} label="Masquerade">
      <Input placeholder="https://example.com" />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'alpn']} label="ALPN">
      <Select mode="tags" options={[{ value: 'h3' }]} />
    </Form.Item>
    <ServerTLSFields />
  </>
);

export const TuicForm = () => (
  <>
    <Form.Item name={['protocolConfig', 'congestion-controller']} label="Congestion Controller">
      <Select options={[{ value: 'cubic' }, { value: 'bbr' }, { value: 'new_reno' }]} />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'max-idle-time']} label="Max Idle Time (ms)">
      <InputNumber min={0} style={{ width: '100%' }} />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'authentication-timeout']} label="Authentication Timeout (ms)">
      <InputNumber min={0} style={{ width: '100%' }} />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'max-udp-relay-packet-size']} label="Max UDP Relay Packet Size">
      <InputNumber min={1} max={65535} style={{ width: '100%' }} />
    </Form.Item>
    <Form.Item name={['protocolConfig', 'alpn']} label="ALPN">
      <Select mode="tags" options={[{ value: 'h3' }]} />
    </Form.Item>
    <ServerTLSFields />
  </>
);
