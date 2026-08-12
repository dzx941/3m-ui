import { Form, Input, InputNumber, Select, Switch } from 'antd';

const field = (name: string, label: string, placeholder?: string) => (
  <Form.Item name={['protocolConfig', name]} label={label}>
    <Input placeholder={placeholder} />
  </Form.Item>
);

const passwordField = (name: string, label: string, placeholder?: string) => (
  <Form.Item name={['protocolConfig', name]} label={label}>
    <Input.Password placeholder={placeholder} />
  </Form.Item>
);

const numberField = (name: string, label: string, min?: number, max?: number) => (
  <Form.Item name={['protocolConfig', name]} label={label}>
    <InputNumber min={min} max={max} style={{ width: '100%' }} />
  </Form.Item>
);

const boolField = (name: string, label: string) => (
  <Form.Item name={['protocolConfig', name]} label={label} valuePropName="checked">
    <Switch checkedChildren="Enabled" unCheckedChildren="Disabled" />
  </Form.Item>
);

const selectField = (name: string, label: string, options: string[]) => (
  <Form.Item name={['protocolConfig', name]} label={label}>
    <Select allowClear options={options.map((value) => ({ value, label: value }))} />
  </Form.Item>
);

const tagsField = (name: string, label: string, options: string[] = []) => (
  <Form.Item name={['protocolConfig', name]} label={label}>
    <Select mode="tags" options={options.map((value) => ({ value, label: value }))} />
  </Form.Item>
);

const jsonField = (name: string, label: string, placeholder: string) => (
  <Form.Item name={['protocolConfig', name]} label={label}>
    <Input.TextArea rows={6} placeholder={placeholder} />
  </Form.Item>
);

const TLSFields = ({ allowInsecure = false }: { allowInsecure?: boolean }) => (
  <>
    {field('certificate', 'Certificate', '/etc/3m-ui/certs/server.crt')}
    {passwordField('private-key', 'Private Key', '/etc/3m-ui/certs/server.key')}
    {selectField('client-auth-type', 'Client Auth Type', ['', 'request', 'require-any', 'verify-if-given', 'require-and-verify'])}
    {field('client-auth-cert', 'Client Auth Certificate')}
    {field('ech-key', 'ECH Key')}
    {allowInsecure && boolField('allow-insecure', 'Allow Insecure')}
  </>
);

const MuxField = () => jsonField('mux-option', 'MUX Option (JSON)', '{"padding":true,"brutal":{"enabled":true,"up":1000,"down":1000}}');

export const ShadowsocksForm = () => (
  <>
    {selectField('cipher', 'Cipher', ['2022-blake3-aes-128-gcm', '2022-blake3-aes-256-gcm', '2022-blake3-chacha20-poly1305', 'none', 'aes-128-gcm', 'aes-192-gcm', 'aes-256-gcm', 'chacha20-ietf-poly1305', 'xchacha20-ietf-poly1305'])}
    {passwordField('password', 'Password')}
    {boolField('udp', 'UDP')}
    {jsonField('simple-obfs', 'Simple Obfs (JSON)', '{"enable":true,"mode":"http"}')}
    {jsonField('shadow-tls', 'ShadowTLS (JSON)', '{"enable":true,"version":3,"handshake":{"dest":"example.com:443"}}')}
    {jsonField('res-tls', 'ResTLS (JSON)', '{"enable":true,"dest":"example.com:443","password":"..."}')}
    {jsonField('jls-config', 'JLS Config (JSON)', '{"enable":true,"dest":"example.com:443"}')}
    {jsonField('kcp-tun', 'KCP Tunnel (JSON)', '{"enable":false,"key":"...","crypt":"aes","mode":"fast"}')}
    {MuxField()}
  </>
);

export const SnellForm = () => (
  <>
    {passwordField('psk', 'PSK')}
    {numberField('version', 'Version', 1, 5)}
    {boolField('udp', 'UDP')}
    {jsonField('obfs-opts', 'Obfs Options (JSON)', '{"mode":"http","host":"bing.com"}')}
    {jsonField('shadow-tls', 'ShadowTLS (JSON)', '{"enable":true,"version":3,"password":"..."}')}
    {jsonField('res-tls', 'ResTLS (JSON)', '{"enable":true,"dest":"example.com:443","password":"..."}')}
    {jsonField('jls-config', 'JLS Config (JSON)', '{"enable":true,"dest":"example.com:443"}')}
  </>
);

export const VmessForm = () => (
  <>
    {jsonField('users', 'Users (JSON)', '[{"username":"user","uuid":"...","alterId":0}]')}
    {field('ws-path', 'WebSocket Path', '/')}
    {field('grpc-service-name', 'gRPC Service Name', 'GunService')}
    {jsonField('mekya-config', 'Mekya Config (JSON)', '{"enable":true}')}
    {jsonField('mkcp-config', 'mKCP Config (JSON)', '{"enable":true,"mtu":1350}')}
    {jsonField('jls-config', 'JLS Config (JSON)', '{"enable":true,"dest":"example.com:443"}')}
    {jsonField('shadow-tls', 'ShadowTLS (JSON)', '{"enable":true,"version":3}')}
    {jsonField('res-tls', 'ResTLS (JSON)', '{"enable":true,"dest":"example.com:443","password":"..."}')}
    {jsonField('reality-config', 'Reality Config (JSON)', '{"dest":"example.com:443","private-key":"...","short-id":["0123456789abcdef"],"server-names":["example.com"]}')}
    {jsonField('tlsmirror-config', 'TLS Mirror Config (JSON)', '{"dest":"example.com:443","primary-key":"..."}')}
    <TLSFields />
    {MuxField()}
  </>
);

export const VlessForm = () => (
  <>
    {jsonField('users', 'Users (JSON)', '[{"username":"user","uuid":"...","flow":"xtls-rprx-vision"}]')}
    {field('ws-path', 'WebSocket Path', '/')}
    {field('grpc-service-name', 'gRPC Service Name', 'GunService')}
    {jsonField('xhttp-config', 'XHTTP Config (JSON)', '{"path":"/","host":"","mode":"auto"}')}
    {field('decryption', 'Decryption')}
    {jsonField('reality-config', 'Reality Config (JSON)', '{"dest":"example.com:443","private-key":"...","short-id":["0123456789abcdef"],"server-names":["example.com"]}')}
    {jsonField('shadow-tls', 'ShadowTLS (JSON)', '{"enable":true,"version":3}')}
    {jsonField('res-tls', 'ResTLS (JSON)', '{"enable":true,"dest":"example.com:443","password":"..."}')}
    {jsonField('jls-config', 'JLS Config (JSON)', '{"enable":true,"dest":"example.com:443"}')}
    <TLSFields allowInsecure />
    {MuxField()}
  </>
);

export const TrojanForm = () => (
  <>
    {jsonField('users', 'Users (JSON)', '[{"username":"user","password":"password"}]')}
    {field('ws-path', 'WebSocket Path', '/')}
    {field('grpc-service-name', 'gRPC Service Name', 'GunService')}
    {jsonField('reality-config', 'Reality Config (JSON)', '{"dest":"example.com:443","private-key":"...","short-id":["0123456789abcdef"],"server-names":["example.com"]}')}
    {jsonField('shadow-tls', 'ShadowTLS (JSON)', '{"enable":true,"version":3}')}
    {jsonField('res-tls', 'ResTLS (JSON)', '{"enable":true,"dest":"example.com:443","password":"..."}')}
    {jsonField('jls-config', 'JLS Config (JSON)', '{"enable":true,"dest":"example.com:443"}')}
    {jsonField('ss-option', 'Shadowsocks Option (JSON)', '{"enabled":true,"method":"aes-128-gcm","password":"..."}')}
    <TLSFields allowInsecure />
    {MuxField()}
  </>
);

export const Hysteria2Form = () => (
  <>
    {jsonField('users', 'Users (JSON)', '{"user":"password"}')}
    {field('up', 'Upload Speed (Mbps)', '1000')}
    {field('down', 'Download Speed (Mbps)', '1000')}
    {boolField('ignore-client-bandwidth', 'Ignore Client Bandwidth')}
    {selectField('obfs', 'Obfuscation', ['salamander'])}
    {passwordField('obfs-password', 'Obfuscation Password')}
    {field('masquerade', 'Masquerade', 'https://example.com')}
    {field('bbr-profile', 'BBR Profile')}
    {jsonField('realm-opts', 'Realm Options (JSON)', '{"enable":true,"server-url":"https://example.com","token":"public","realm-id":"my-realm"}')}
    {tagsField('alpn', 'ALPN', ['h3'])}
    <TLSFields />
    {MuxField()}
  </>
);

export const Hysteria2RealmForm = () => (
  <>
    {passwordField('token', 'Bearer Token', 'public')}
    {numberField('max-realms', 'Max Realms', 0)}
    {numberField('max-realms-per-ip', 'Max Realms Per IP', 0)}
    {field('trusted-proxy-header', 'Trusted Proxy Header', 'X-Forwarded-For')}
    {field('realm-name-pattern', 'Realm Name Pattern', '^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$')}
    <TLSFields />
    {tagsField('alpn', 'ALPN', ['h2', 'http/1.1'])}
  </>
);

export const TuicForm = () => (
  <>
    {jsonField('users', 'Users (JSON, TUIC V5)', '{"00000000-0000-0000-0000-000000000001":"password"}')}
    {jsonField('token', 'Token (JSON, TUIC V4)', '["TOKEN"]')}
    <TLSFields />
    {selectField('congestion-controller', 'Congestion Controller', ['cubic', 'new_reno', 'bbr'])}
    {field('bbr-profile', 'BBR Profile')}
    {numberField('max-idle-time', 'Max Idle Time (ms)', 0)}
    {numberField('authentication-timeout', 'Authentication Timeout (ms)', 0)}
    {tagsField('alpn', 'ALPN', ['h3'])}
    {numberField('max-udp-relay-packet-size', 'Max UDP Relay Packet Size', 1, 65535)}
    {MuxField()}
  </>
);

export const ShadowQuicForm = () => (
  <>
    {jsonField('users', 'Users (JSON)', '[{"username":"user","password":"password"}]')}
    {jsonField('jls-upstream', 'JLS Upstream (JSON)', '{"addr":"www.example.com:443","sni":"example.com","rate-limit":0}')}
    {tagsField('alpn', 'ALPN', ['h3'])}
    {tagsField('quic-versions', 'QUIC Versions', ['v1', 'v2'])}
    {boolField('zero-rtt', 'Zero RTT')}
    {selectField('congestion-controller', 'Congestion Controller', ['cubic', 'new_reno', 'bbr'])}
    {field('up', 'Upload Speed')}
    {field('down', 'Download Speed')}
    {boolField('ignore-client-bandwidth', 'Ignore Client Bandwidth')}
    {numberField('cwnd', 'CWND', 1)}
    {field('bbr-profile', 'BBR Profile')}
    {numberField('max-idle-time', 'Max Idle Time (ms)', 0)}
    {numberField('max-datagram-frame-size', 'Max Datagram Frame Size', 1, 65535)}
    {numberField('recv-window-conn', 'Receive Window Conn', 0)}
    {numberField('recv-window', 'Receive Window', 0)}
    {boolField('disable-mtu-discovery', 'Disable MTU Discovery')}
  </>
);

export const AnyTLSForm = () => (
  <>
    {jsonField('users', 'Users (JSON)', '{"username":"password"}')}
    <TLSFields allowInsecure />
    {jsonField('shadow-tls', 'ShadowTLS (JSON)', '{"enable":true,"version":3,"handshake":{"dest":"example.com:443"}}')}
    {jsonField('res-tls', 'ResTLS (JSON)', '{"enable":true,"dest":"example.com:443","password":"..."}')}
    {jsonField('jls-config', 'JLS Config (JSON)', '{"enable":true,"dest":"example.com:443"}')}
    {field('padding-scheme', 'Padding Scheme')}
  </>
);

export const MieruForm = () => (
  <>
    {selectField('transport', 'Transport', ['TCP', 'UDP'])}
    {jsonField('users', 'Users (JSON)', '{"username":"password"}')}
    {field('traffic-pattern', 'Traffic Pattern (Base64)')}
    {boolField('user-hint-is-mandatory', 'User Hint Is Mandatory')}
  </>
);

export const SudokuForm = () => (
  <>
    {field('key', 'Server Key')}
    {selectField('aead-method', 'AEAD Method', ['chacha20-poly1305', 'aes-128-gcm', 'none'])}
    {numberField('padding-min', 'Padding Min', 0, 100)}
    {numberField('padding-max', 'Padding Max', 0, 100)}
    {selectField('table-type', 'Table Type', ['prefer_ascii', 'prefer_entropy', 'up_ascii_down_entropy', 'up_entropy_down_ascii'])}
    {field('custom-table', 'Custom Table')}
    {tagsField('custom-tables', 'Custom Tables')}
    {numberField('handshake-timeout', 'Handshake Timeout', 0)}
    {boolField('enable-pure-downlink', 'Pure Downlink')}
    {jsonField('httpmask', 'HTTPMask (JSON)', '{"disable":false,"mode":"legacy","path-root":""}')}
    {boolField('disable-http-mask', 'Disable HTTP Mask')}
    {selectField('http-mask-mode', 'HTTP Mask Mode', ['legacy', 'stream', 'poll', 'auto', 'ws'])}
    {field('path-root', 'HTTP Mask Path Root')}
    {field('fallback', 'Fallback Address')}
    {MuxField()}
  </>
);

export const TrustTunnelForm = () => (
  <>
    {jsonField('users', 'Users (JSON)', '[{"username":"user","password":"password"}]')}
    <TLSFields />
    {tagsField('network', 'Network', ['tcp', 'udp'])}
    {selectField('congestion-controller', 'Congestion Controller', ['cubic', 'new_reno', 'bbr'])}
    {field('bbr-profile', 'BBR Profile')}
  </>
);
