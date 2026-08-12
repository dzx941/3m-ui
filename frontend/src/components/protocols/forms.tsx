import { Form, Input, InputNumber, Select, Switch } from 'antd';

const field = (name: string, label: string, placeholder?: string) => (
  <Form.Item name={['protocolConfig', name]} label={label}>
    <Input placeholder={placeholder} />
  </Form.Item>
);

const numberField = (name: string, label: string, min = 0, max?: number) => (
  <Form.Item name={['protocolConfig', name]} label={label}>
    <InputNumber min={min} max={max} style={{ width: '100%' }} />
  </Form.Item>
);

const boolField = (name: string, label: string) => (
  <Form.Item name={['protocolConfig', name]} label={label} valuePropName="checked">
    <Switch checkedChildren="Enabled" unCheckedChildren="Disabled" />
  </Form.Item>
);

const tagsField = (name: string, label: string, options: string[] = []) => (
  <Form.Item name={['protocolConfig', name]} label={label}>
    <Select mode="tags" options={options.map((value) => ({ value }))} />
  </Form.Item>
);

const JSONField = (name: string, label: string, placeholder: string) => (
  <Form.Item name={['protocolConfig', name]} label={label}>
    <Input.TextArea rows={5} placeholder={placeholder} />
  </Form.Item>
);

const ServerTLSFields = () => (
  <>
    {field('certificate', 'Certificate', '/etc/3m-ui/certs/server.crt')}
    <Form.Item name={['protocolConfig', 'private-key']} label="Private Key">
      <Input.Password placeholder="/etc/3m-ui/certs/server.key" />
    </Form.Item>
    {tagsField('alpn', 'ALPN', ['h2', 'http/1.1', 'h3'])}
    {boolField('allow-insecure', 'Allow Insecure')}
  </>
);

export const SocksForm = () => (
  <>
    {boolField('udp', 'UDP Support')}
    {JSONField('users', 'Users (JSON)', '[{"username":"user","password":"password"}]')}
  </>
);

export const HttpForm = () => (
  <>
    {JSONField('users', 'Users (JSON)', '[{"username":"user","password":"password"}]')}
    {boolField('tls', 'TLS')}
    <ServerTLSFields />
  </>
);

export const TproxyForm = () => (
  <>
    {boolField('udp', 'UDP Support')}
    {numberField('routing-mark', 'Routing Mark')}
  </>
);

export const RedirForm = () => <>{numberField('routing-mark', 'Routing Mark')}</>;

export const MixedForm = () => (
  <>
    {boolField('udp', 'UDP Support')}
    {JSONField('users', 'Users (JSON)', '[{"username":"user","password":"password"}]')}
  </>
);

export const TunnelForm = () => (
  <>
    {field('target', 'Target', '127.0.0.1:8080')}
    {field('network', 'Network', 'tcp')}
    {field('headers', 'Headers (JSON)', '{"Host":"example.com"}')}
  </>
);

export const TunForm = () => (
  <>
    <Form.Item name={['protocolConfig', 'stack']} label="Stack">
      <Select options={['system', 'gvisor', 'mixed'].map((value) => ({ value }))} />
    </Form.Item>
    {tagsField('dns-hijack', 'DNS Hijack', ['any:53', 'tcp://any:53', 'udp://any:53'])}
    {boolField('auto-route', 'Auto Route')}
    {boolField('auto-detect-interface', 'Auto Detect Interface')}
    {boolField('strict-route', 'Strict Route')}
    {boolField('endpoint-independent-nat', 'Endpoint Independent NAT')}
    {numberField('mtu', 'MTU', 576, 9000)}
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
    {field('password', 'Password')}
    {field('plugin', 'Plugin')}
    {JSONField('plugin-opts', 'Plugin Options (JSON)', '{"mode":"tls"}')}
    {field('obfs', 'Simple Obfs')}
    {JSONField('restls-opts', 'RESTLS Options (JSON)', '{"host":"example.com"}')}
    {JSONField('jls-opts', 'JLS Options (JSON)', '{"host":"example.com"}')}
  </>
);

export const SnellForm = () => (
  <>
    {numberField('version', 'Version', 1, 5)}
    {field('psk', 'PSK')}
    {field('obfs-opts.host', 'Obfuscation Host')}
    {field('obfs-opts.mode', 'Obfuscation Mode')}
    {JSONField('shadow-tls-opts', 'ShadowTLS Options (JSON)', '{"version":3,"password":"..."}')}
    {JSONField('restls-opts', 'RESTLS Options (JSON)', '{"host":"example.com"}')}
    {JSONField('jls-opts', 'JLS Options (JSON)', '{"host":"example.com"}')}
  </>
);

const WebTransportFields = () => (
  <>
    {field('ws-path', 'WebSocket Path', '/')}
    {field('ws-headers', 'WebSocket Headers (JSON)', '{"Host":"example.com"}')}
    {field('grpc-service-name', 'gRPC Service Name', 'GunService')}
    {numberField('grpc-max-concurrent-streams', 'gRPC Max Concurrent Streams')}
    {field('xhttp-path', 'XHTTP Path', '/')}
    {field('xhttp-mode', 'XHTTP Mode', 'auto')}
    {JSONField('xhttp-opts', 'XHTTP Options (JSON)', '{"mode":"auto"}')}
  </>
);

export const VmessForm = () => (
  <>
    {JSONField('users', 'Users (JSON)', '[{"username":"user","uuid":"...","alterId":0}]')}
    {field('cipher', 'Cipher', 'auto')}
    {field('packet-encoding', 'Packet Encoding', 'xudp')}
    {WebTransportFields()}
    {JSONField('reality-config', 'Reality Server Config (JSON)', '{"dest":"example.com:443","private-key":"...","short-id":["0123456789abcdef"],"server-names":["example.com"]}')}
    {JSONField('restls-opts', 'RESTLS Options (JSON)', '{"host":"example.com"}')}
    {JSONField('jls-opts', 'JLS Options (JSON)', '{"host":"example.com"}')}
    {JSONField('tlsmirror-config', 'TLS Mirror Config (JSON)', '{"host":"example.com"}')}
    {JSONField('mekya-config', 'MEKYA Config (JSON)', '{}')}
    <ServerTLSFields />
  </>
);

export const VlessForm = () => (
  <>
    {JSONField('users', 'Users (JSON)', '[{"username":"user","uuid":"...","flow":"xtls-rprx-vision"}]')}
    {field('decryption', 'Decryption', 'none')}
    {field('flow', 'Flow', 'xtls-rprx-vision')}
    {WebTransportFields()}
    {JSONField('reality-config', 'Reality Server Config (JSON)', '{"dest":"example.com:443","private-key":"...","short-id":["0123456789abcdef"],"server-names":["example.com"]}')}
    {JSONField('restls-opts', 'RESTLS Options (JSON)', '{"host":"example.com"}')}
    {JSONField('jls-opts', 'JLS Options (JSON)', '{"host":"example.com"}')}
    <ServerTLSFields />
  </>
);

export const TrojanForm = () => (
  <>
    {JSONField('users', 'Users (JSON)', '[{"username":"user","password":"password"}]')}
    {WebTransportFields()}
    {JSONField('reality-config', 'Reality Server Config (JSON)', '{"dest":"example.com:443","private-key":"...","short-id":["0123456789abcdef"],"server-names":["example.com"]}')}
    {JSONField('shadow-tls-opts', 'ShadowTLS Options (JSON)', '{"version":3,"password":"..."}')}
    {JSONField('restls-opts', 'RESTLS Options (JSON)', '{"host":"example.com"}')}
    {JSONField('jls-opts', 'JLS Options (JSON)', '{"host":"example.com"}')}
    {boolField('allow-insecure', 'Allow Insecure')}
    <ServerTLSFields />
  </>
);

export const Hysteria2Form = () => (
  <>
    {JSONField('users', 'Users (JSON)', '{"user":"password"}')}
    {field('up', 'Upload Speed', '1000 mbps')}
    {field('down', 'Download Speed', '1000 mbps')}
    <Form.Item name={['protocolConfig', 'obfs']} label="Obfuscation">
      <Select allowClear options={['salamander', 'gecko'].map((value) => ({ value }))} />
    </Form.Item>
    {field('obfs-password', 'Obfuscation Password')}
    {field('masquerade', 'Masquerade', 'https://example.com')}
    {field('ignore-client-bandwidth', 'Ignore Client Bandwidth')}
    {field('hop-interval', 'Hop Interval', '30s')}
    {JSONField('realm-opts', 'Realm Options (JSON)', '{"proxy":"127.0.0.1:8080"}')}
    {JSONField('tls', 'TLS Options (JSON)', '{"certificate":"/path/cert.pem","private-key":"/path/key.pem"}')}
    <ServerTLSFields />
  </>
);

export const Hysteria2RealmForm = () => (
  <>
    {field('listen', 'Realm Listen', '0.0.0.0:443')}
    {field('server', 'Realm Server')}
    {field('up', 'Upload Speed', '1000 mbps')}
    {field('down', 'Download Speed', '1000 mbps')}
    {field('hop-interval', 'Hop Interval', '30s')}
    {JSONField('realm-opts', 'Realm Options (JSON)', '{"proxy":"127.0.0.1:8080"}')}
    <ServerTLSFields />
  </>
);

export const TuicForm = () => (
  <>
    {JSONField('users', 'Users (JSON)', '{"uuid":"password"}')}
    <Form.Item name={['protocolConfig', 'congestion-controller']} label="Congestion Controller">
      <Select options={[{ value: 'cubic' }, { value: 'bbr' }, { value: 'new_reno' }]} />
    </Form.Item>
    {numberField('max-idle-time', 'Max Idle Time (ms)')}
    {numberField('authentication-timeout', 'Authentication Timeout (ms)')}
    {numberField('max-udp-relay-packet-size', 'Max UDP Relay Packet Size', 1, 65535)}
    {tagsField('alpn', 'ALPN', ['h3'])}
    {boolField('disable-sni', 'Disable SNI')}
    {boolField('reduce-rtt', 'Reduce RTT')}
    {field('heartbeat-interval', 'Heartbeat Interval')}
    <ServerTLSFields />
  </>
);

export const ShadowQuicForm = () => (
  <>
    {field('password', 'Password')}
    {field('transport', 'Transport')}
    {numberField('max-idle-time', 'Max Idle Time (ms)')}
    {numberField('max-datagram-frame-size', 'Max Datagram Frame Size', 1, 65535)}
    <Form.Item name={['protocolConfig', 'congestion-controller']} label="Congestion Controller">
      <Select options={['cubic', 'bbr', 'new_reno'].map((value) => ({ value }))} />
    </Form.Item>
    {boolField('zero-rtt', 'Zero RTT')}
    {tagsField('alpn', 'ALPN', ['h3'])}
  </>
);

export const AnyTLSForm = () => (
  <>
    {JSONField('users', 'Users (JSON)', '[{"name":"user","password":"password"}]')}
    {numberField('padding-scheme', 'Padding Scheme')}
    {numberField('idle-session-check-interval', 'Idle Session Check Interval (ms)')}
    {numberField('idle-session-timeout', 'Idle Session Timeout (ms)')}
    {numberField('min-idle-session', 'Minimum Idle Session')}
    {JSONField('tls', 'TLS Options (JSON)', '{"certificate":"/path/cert.pem","private-key":"/path/key.pem"}')}
    {JSONField('restls-opts', 'RESTLS Options (JSON)', '{"host":"example.com"}')}
    {JSONField('jls-opts', 'JLS Options (JSON)', '{"host":"example.com"}')}
  </>
);

export const MieruForm = () => (
  <>
    {JSONField('users', 'Users (JSON)', '[{"name":"user","password":"password"}]')}
    {field('protocol-version', 'Protocol Version')}
    {field('mtu', 'MTU')}
    {boolField('client-reconnect', 'Client Reconnect')}
    {field('logging-level', 'Logging Level')}
  </>
);

export const SudokuForm = () => (
  <>
    {field('password', 'Password')}
    {numberField('idle-timeout', 'Idle Timeout (ms)')}
    {numberField('max-connections', 'Max Connections')}
    {JSONField('tls', 'TLS Options (JSON)', '{"certificate":"/path/cert.pem","private-key":"/path/key.pem"}')}
  </>
);

export const TrustTunnelForm = () => (
  <>
    {field('username', 'Username')}
    {field('password', 'Password')}
    {field('target', 'Target', '127.0.0.1:8080')}
    {field('server-name', 'Server Name')}
    {JSONField('tls', 'TLS Options (JSON)', '{"certificate":"/path/cert.pem","private-key":"/path/key.pem"}')}
    {JSONField('padding', 'Padding Options (JSON)', '{}')}
  </>
);
