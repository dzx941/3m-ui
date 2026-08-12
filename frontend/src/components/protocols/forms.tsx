import React from 'react';
import { Button, Form, Input, InputNumber, Select, Space, Switch } from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';

const path = (name: string) => ['protocolConfig', ...name.split('.')];

const text = (name: string, label: string, placeholder?: string, required = false) => (
  <Form.Item name={path(name)} label={label} rules={required ? [{ required: true, message: `${label} is required` }] : undefined}>
    <Input placeholder={placeholder} />
  </Form.Item>
);

const secret = (name: string, label: string, placeholder?: string, required = false) => (
  <Form.Item name={path(name)} label={label} rules={required ? [{ required: true, message: `${label} is required` }] : undefined}>
    <Input.Password placeholder={placeholder} />
  </Form.Item>
);

const number = (name: string, label: string, min?: number, max?: number) => (
  <Form.Item name={path(name)} label={label}>
    <InputNumber min={min} max={max} style={{ width: '100%' }} />
  </Form.Item>
);

const toggle = (name: string, label: string) => (
  <Form.Item name={path(name)} label={label} valuePropName="checked">
    <Switch checkedChildren="On" unCheckedChildren="Off" />
  </Form.Item>
);

const select = (name: string, label: string, options: string[]) => (
  <Form.Item name={path(name)} label={label}>
    <Select allowClear options={options.map((value) => ({ value, label: value }))} />
  </Form.Item>
);

const multi = (name: string, label: string, options: string[] = []) => (
  <Form.Item name={path(name)} label={label}>
    <Select mode="tags" options={options.map((value) => ({ value, label: value }))} />
  </Form.Item>
);

const UsersArray = ({ mode }: { mode: 'password' | 'uuid' }) => (
  <Form.List name={['protocolConfig', 'users']}>
    {(fields, { add, remove }) => (
      <Form.Item label="Users">
        {fields.map((field) => (
          <Space key={field.key} align="baseline" style={{ display: 'flex', marginBottom: 8 }}>
            <Form.Item {...field} name={[field.name, 'username']} noStyle><Input placeholder="Username" /></Form.Item>
            {mode === 'uuid' ? (
              <>
                <Form.Item {...field} name={[field.name, 'uuid']} noStyle><Input placeholder="UUID" /></Form.Item>
                <Form.Item {...field} name={[field.name, 'flow']} noStyle><Input placeholder="Flow" /></Form.Item>
                <Form.Item {...field} name={[field.name, 'alterId']} noStyle><InputNumber min={0} placeholder="alterId" /></Form.Item>
              </>
            ) : (
              <Form.Item {...field} name={[field.name, 'password']} noStyle><Input.Password placeholder="Password" /></Form.Item>
            )}
            <MinusCircleOutlined onClick={() => remove(field.name)} />
          </Space>
        ))}
        <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>Add user</Button>
      </Form.Item>
    )}
  </Form.List>
);

const UsersMap = () => (
  <Form.List name={['protocolConfig', 'users']}>
    {(fields, { add, remove }) => (
      <Form.Item label="Users">
        {fields.map((field) => (
          <Space key={field.key} align="baseline" style={{ display: 'flex', marginBottom: 8 }}>
            <Form.Item {...field} name={[field.name, 'username']} noStyle><Input placeholder="Username / UUID" /></Form.Item>
            <Form.Item {...field} name={[field.name, 'password']} noStyle><Input.Password placeholder="Password" /></Form.Item>
            <MinusCircleOutlined onClick={() => remove(field.name)} />
          </Space>
        ))}
        <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>Add user</Button>
      </Form.Item>
    )}
  </Form.List>
);

const TLS = ({ allowInsecure = false }: { allowInsecure?: boolean }) => (
  <>
    {text('certificate', 'Certificate', '/etc/3m-ui/certs/server.crt')}
    {text('private-key', 'Private Key', '/etc/3m-ui/certs/server.key')}
    {select('client-auth-type', 'Client Auth Type', ['request', 'require-any', 'verify-if-given', 'require-and-verify'])}
    {text('client-auth-cert', 'Client Auth Certificate')}
    {text('ech-key', 'ECH Key')}
    {allowInsecure && toggle('allow-insecure', 'Allow Insecure')}
  </>
);

const Mux = () => (
  <>
    {toggle('mux-option.padding', 'MUX Padding')}
    {toggle('mux-option.brutal.enabled', 'MUX Brutal')}
    {number('mux-option.brutal.up', 'MUX Brutal Up (Mbps)', 0)}
    {number('mux-option.brutal.down', 'MUX Brutal Down (Mbps)', 0)}
  </>
);

const ShadowTLS = () => (
  <>
    {toggle('shadow-tls.enable', 'ShadowTLS Enabled')}
    {number('shadow-tls.version', 'ShadowTLS Version', 1, 3)}
    {secret('shadow-tls.password', 'ShadowTLS Password')}
    {text('shadow-tls.handshake.dest', 'ShadowTLS Handshake Destination', 'example.com:443')}
    {text('shadow-tls.handshake.proxy', 'ShadowTLS Handshake Proxy')}
  </>
);

const ResTLS = () => (
  <>
    {toggle('res-tls.enable', 'ResTLS Enabled')}
    {text('res-tls.dest', 'ResTLS Destination', 'example.com:443')}
    {secret('res-tls.password', 'ResTLS Password')}
    {text('res-tls.restls-script', 'ResTLS Script')}
    {number('res-tls.min-record-len', 'ResTLS Min Record Length', 0)}
    {text('res-tls.proxy', 'ResTLS Proxy')}
  </>
);

const JLS = () => (
  <>
    {toggle('jls-config.enable', 'JLS Enabled')}
    {text('jls-config.dest', 'JLS Destination', 'example.com:443')}
    {text('jls-config.sni', 'JLS SNI')}
    {multi('jls-config.alpn', 'JLS ALPN', ['h2', 'http/1.1'])}
    {text('jls-config.proxy', 'JLS Proxy')}
    {number('jls-config.rate-limit', 'JLS Rate Limit (bit/s)', 0)}
  </>
);

const Reality = () => (
  <>
    {text('reality-config.dest', 'Reality Destination', 'example.com:443')}
    {secret('reality-config.private-key', 'Reality Private Key')}
    {multi('reality-config.short-id', 'Reality Short IDs')}
    {multi('reality-config.server-names', 'Reality Server Names')}
    {number('reality-config.limit-fallback-upload.after-bytes', 'Reality Upload After Bytes', 0)}
    {number('reality-config.limit-fallback-upload.bytes-per-sec', 'Reality Upload Bytes/s', 0)}
    {number('reality-config.limit-fallback-upload.burst-bytes-per-sec', 'Reality Upload Burst Bytes/s', 0)}
    {number('reality-config.limit-fallback-download.after-bytes', 'Reality Download After Bytes', 0)}
    {number('reality-config.limit-fallback-download.bytes-per-sec', 'Reality Download Bytes/s', 0)}
    {number('reality-config.limit-fallback-download.burst-bytes-per-sec', 'Reality Download Burst Bytes/s', 0)}
  </>
);

const Transport = () => (
  <>
    {text('ws-path', 'WebSocket Path', '/')}
    {text('grpc-service-name', 'gRPC Service Name', 'GunService')}
  </>
);

export const ShadowsocksForm = () => (
  <>
    {select('cipher', 'Cipher', ['2022-blake3-aes-128-gcm', '2022-blake3-aes-256-gcm', '2022-blake3-chacha20-poly1305', 'none', 'aes-128-gcm', 'aes-192-gcm', 'aes-256-gcm', 'chacha20-ietf-poly1305', 'xchacha20-ietf-poly1305'])}
    {secret('password', 'Password', undefined, true)}
    {toggle('udp', 'UDP')}
    {toggle('simple-obfs.enable', 'Simple Obfs Enabled')}
    {select('simple-obfs.mode', 'Simple Obfs Mode', ['http', 'tls'])}
    <ShadowTLS />
    <ResTLS />
    <JLS />
    {toggle('kcp-tun.enable', 'KCP Tunnel Enabled')}
    {secret('kcp-tun.key', 'KCP Key')}
    {select('kcp-tun.crypt', 'KCP Crypt', ['aes', 'aes-128', 'aes-128-gcm', 'aes-192', 'salsa20', 'blowfish', 'twofish', 'cast5', '3des', 'tea', 'xtea', 'xor', 'none', 'null'])}
    {select('kcp-tun.mode', 'KCP Mode', ['fast3', 'fast2', 'fast', 'normal', 'manual'])}
    {number('kcp-tun.conn', 'KCP Connections', 1)}
    {number('kcp-tun.autoexpire', 'KCP Auto Expire (s)', 0)}
    {number('kcp-tun.scavengettl', 'KCP Scavenge TTL (s)', 0)}
    {number('kcp-tun.ratelimit', 'KCP Rate Limit (bytes/s)', 0)}
    {number('kcp-tun.mtu', 'KCP MTU', 1)}
    {number('kcp-tun.sndwnd', 'KCP Send Window', 1)}
    {number('kcp-tun.rcvwnd', 'KCP Receive Window', 1)}
    {number('kcp-tun.datashard', 'KCP Data Shard', 0)}
    {number('kcp-tun.parityshard', 'KCP Parity Shard', 0)}
    {number('kcp-tun.dscp', 'KCP DSCP', 0, 63)}
    {toggle('kcp-tun.nocomp', 'KCP No Compression')}
    {toggle('kcp-tun.acknodelay', 'KCP Ack No Delay')}
    {number('kcp-tun.nodelay', 'KCP No Delay', 0)}
    {number('kcp-tun.interval', 'KCP Interval (ms)', 0)}
    {number('kcp-tun.resend', 'KCP Resend', 0)}
    {number('kcp-tun.sockbuf', 'KCP Socket Buffer', 0)}
    {number('kcp-tun.smuxver', 'KCP SMUX Version', 1, 2)}
    {number('kcp-tun.smuxbuf', 'KCP SMUX Buffer', 0)}
    {number('kcp-tun.framesize', 'KCP Frame Size', 0)}
    {number('kcp-tun.streambuf', 'KCP Stream Buffer', 0)}
    {number('kcp-tun.keepalive', 'KCP Keepalive (s)', 0)}
    <Mux />
  </>
);

export const SnellForm = () => (
  <>
    {secret('psk', 'PSK', undefined, true)}
    {number('version', 'Version', 1, 5)}
    {toggle('udp', 'UDP')}
    {select('obfs-opts.mode', 'Obfs Mode', ['http', 'tls'])}
    {text('obfs-opts.host', 'Obfs Host')}
    <ShadowTLS />
    <ResTLS />
    <JLS />
  </>
);

export const VmessForm = () => (
  <>
    <UsersArray mode="uuid" />
    <Transport />
    {toggle('mekya-config.enable', 'Mekya Enabled')}
    {number('mekya-config.max-write-size', 'Mekya Max Write Size', 0)}
    {number('mekya-config.max-write-duration-ms', 'Mekya Max Write Duration (ms)', 0)}
    {number('mekya-config.max-simultaneous-write-connection', 'Mekya Max Simultaneous Writes', 0)}
    {number('mekya-config.packet-writing-buffer', 'Mekya Packet Buffer', 0)}
    {number('mkcp-config.mtu', 'mKCP MTU', 1)}
    {number('mkcp-config.tti', 'mKCP TTI', 0)}
    {number('mkcp-config.uplink-capacity', 'mKCP Uplink', 0)}
    {number('mkcp-config.downlink-capacity', 'mKCP Downlink', 0)}
    {toggle('mkcp-config.congestion', 'mKCP Congestion')}
    {number('mkcp-config.write-buffer', 'mKCP Write Buffer', 0)}
    {number('mkcp-config.read-buffer', 'mKCP Read Buffer', 0)}
    {text('mkcp-config.seed', 'mKCP Seed')}
    {select('mkcp-config.header', 'mKCP Header', ['none', 'srtp', 'utp', 'wechat-video', 'dtls', 'wireguard'])}
    <TLS />
    <ShadowTLS />
    <ResTLS />
    <JLS />
    <Reality />
    {text('tlsmirror-config.dest', 'TLSMirror Destination')}
    {secret('tlsmirror-config.primary-key', 'TLSMirror Primary Key')}
    <Mux />
  </>
);

export const VlessForm = () => (
  <>
    <UsersArray mode="uuid" />
    <Transport />
    {text('xhttp-config.path', 'XHTTP Path', '/')}
    {text('xhttp-config.host', 'XHTTP Host')}
    {select('xhttp-config.mode', 'XHTTP Mode', ['auto', 'stream-one', 'stream-up', 'packet-up'])}
    {toggle('xhttp-config.no-sse-header', 'XHTTP No SSE Header')}
    {text('xhttp-config.x-padding-bytes', 'XHTTP Padding Bytes')}
    {toggle('xhttp-config.x-padding-obfs-mode', 'XHTTP Padding Obfs Mode')}
    {text('xhttp-config.x-padding-key', 'XHTTP Padding Key')}
    {text('xhttp-config.x-padding-header', 'XHTTP Padding Header')}
    {select('xhttp-config.x-padding-placement', 'XHTTP Padding Placement', ['queryInHeader', 'cookie', 'header', 'query'])}
    {select('xhttp-config.x-padding-method', 'XHTTP Padding Method', ['repeat-x', 'tokenish'])}
    {select('xhttp-config.uplink-http-method', 'XHTTP Uplink HTTP Method', ['POST', 'PUT', 'PATCH', 'DELETE'])}
    {select('xhttp-config.session-placement', 'XHTTP Session Placement', ['path', 'query', 'cookie', 'header'])}
    {text('xhttp-config.session-key', 'XHTTP Session Key')}
    {select('xhttp-config.session-table', 'XHTTP Session Table', ['', 'uuid', 'ALPHABET', 'Alphabet', 'BASE36', 'Base62', 'HEX', 'alphabet', 'base36', 'hex', 'number'])}
    {text('xhttp-config.session-length', 'XHTTP Session Length', '16-32')}
    {select('xhttp-config.seq-placement', 'XHTTP Sequence Placement', ['path', 'query', 'cookie', 'header'])}
    {text('xhttp-config.seq-key', 'XHTTP Sequence Key')}
    {select('xhttp-config.uplink-data-placement', 'XHTTP Uplink Data Placement', ['body', 'cookie', 'header'])}
    {text('xhttp-config.uplink-data-key', 'XHTTP Uplink Data Key')}
    {number('xhttp-config.uplink-chunk-size', 'XHTTP Uplink Chunk Size', 0)}
    {number('xhttp-config.sc-max-buffered-posts', 'XHTTP Max Buffered Posts', 0)}
    {text('xhttp-config.sc-stream-up-server-secs', 'XHTTP Stream Up Server Seconds')}
    {number('xhttp-config.sc-max-each-post-bytes', 'XHTTP Max Each Post Bytes', 0)}
    {text('decryption', 'VLESS Decryption')}
    <TLS allowInsecure />
    <Reality />
    <ShadowTLS />
    <ResTLS />
    <JLS />
    <Mux />
  </>
);

export const TrojanForm = () => (
  <>
    <UsersArray mode="password" />
    <Transport />
    <TLS allowInsecure />
    <Reality />
    <ShadowTLS />
    <ResTLS />
    <JLS />
    {toggle('ss-option.enabled', 'Trojan SS Enabled')}
    {select('ss-option.method', 'Trojan SS Method', ['aes-128-gcm', 'aes-256-gcm', 'chacha20-ietf-poly1305'])}
    {secret('ss-option.password', 'Trojan SS Password')}
    <Mux />
  </>
);

export const Hysteria2Form = () => (
  <>
    <UsersMap />
    {text('up', 'Upload Speed', '1000')}
    {text('down', 'Download Speed', '1000')}
    {toggle('ignore-client-bandwidth', 'Ignore Client Bandwidth')}
    {select('obfs', 'Obfuscation', ['salamander'])}
    {secret('obfs-password', 'Obfuscation Password')}
    {text('masquerade', 'Masquerade', 'https://example.com')}
    {select('bbr-profile', 'BBR Profile', ['standard', 'conservative', 'aggressive'])}
    {multi('alpn', 'ALPN', ['h3'])}
    <TLS />
    <Mux />
  </>
);

export const TuicForm = () => (
  <>
    <UsersMap />
    <Form.List name={['protocolConfig', 'token']}>
      {(fields, { add, remove }) => (
        <Form.Item label="TUIC V4 Tokens">
          {fields.map((field) => (
            <Space key={field.key} style={{ display: 'flex', marginBottom: 8 }}>
              <Form.Item {...field} noStyle><Input placeholder="TOKEN" /></Form.Item>
              <MinusCircleOutlined onClick={() => remove(field.name)} />
            </Space>
          ))}
          <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>Add token</Button>
        </Form.Item>
      )}
    </Form.List>
    <TLS />
    {select('congestion-controller', 'Congestion Controller', ['cubic', 'new_reno', 'bbr'])}
    {select('bbr-profile', 'BBR Profile', ['standard', 'conservative', 'aggressive'])}
    {number('max-idle-time', 'Max Idle Time (ms)', 0)}
    {number('authentication-timeout', 'Authentication Timeout (ms)', 0)}
    {multi('alpn', 'ALPN', ['h3'])}
    {number('max-udp-relay-packet-size', 'Max UDP Relay Packet Size', 1, 65535)}
    <Mux />
  </>
);

export const ShadowQuicForm = () => (
  <>
    <UsersArray mode="password" />
    {text('jls-upstream.addr', 'JLS Upstream Address', 'www.example.com:443')}
    {text('jls-upstream.sni', 'JLS Upstream SNI')}
    {text('jls-upstream.proxy', 'JLS Upstream Proxy')}
    {number('jls-upstream.rate-limit', 'JLS Upstream Rate Limit (bit/s)', 0)}
    {multi('alpn', 'ALPN', ['h3'])}
    {multi('quic-versions', 'QUIC Versions', ['v1', 'v2'])}
    {toggle('zero-rtt', 'Zero RTT')}
    {select('congestion-controller', 'Congestion Controller', ['cubic', 'new_reno', 'bbr'])}
    {text('up', 'Upload Speed')}
    {text('down', 'Download Speed')}
    {toggle('ignore-client-bandwidth', 'Ignore Client Bandwidth')}
    {number('cwnd', 'CWND', 1)}
    {select('bbr-profile', 'BBR Profile', ['standard', 'conservative', 'aggressive'])}
    {number('max-idle-time', 'Max Idle Time (ms)', 0)}
    {number('max-datagram-frame-size', 'Max Datagram Frame Size', 1, 65535)}
    {number('recv-window-conn', 'Receive Window Conn', 0)}
    {number('recv-window', 'Receive Window', 0)}
    {toggle('disable-mtu-discovery', 'Disable MTU Discovery')}
  </>
);

export const AnyTLSForm = () => (
  <>
    <UsersMap />
    <TLS allowInsecure />
    <ShadowTLS />
    <ResTLS />
    <JLS />
    {text('padding-scheme', 'Padding Scheme')}
  </>
);

export const MieruForm = () => (
  <>
    {select('transport', 'Transport', ['TCP', 'UDP'])}
    <UsersMap />
    {text('traffic-pattern', 'Traffic Pattern (Base64)')}
    {toggle('user-hint-is-mandatory', 'User Hint Is Mandatory')}
  </>
);

export const SudokuForm = () => (
  <>
    {text('key', 'Server Key', undefined, true)}
    {select('aead-method', 'AEAD Method', ['chacha20-poly1305', 'aes-128-gcm', 'none'])}
    {number('padding-min', 'Padding Min', 0, 100)}
    {number('padding-max', 'Padding Max', 0, 100)}
    {select('table-type', 'Table Type', ['prefer_ascii', 'prefer_entropy', 'up_ascii_down_entropy', 'up_entropy_down_ascii'])}
    {text('custom-table', 'Custom Table')}
    {multi('custom-tables', 'Custom Tables')}
    {number('handshake-timeout', 'Handshake Timeout (s)', 0)}
    {toggle('enable-pure-downlink', 'Pure Downlink')}
    {toggle('httpmask.disable', 'HTTPMask Disabled')}
    {select('httpmask.mode', 'HTTPMask Mode', ['legacy', 'stream', 'poll', 'auto', 'ws'])}
    {text('httpmask.path-root', 'HTTPMask Path Root')}
    {text('fallback', 'Fallback Address')}
    <Mux />
  </>
);

export const TrustTunnelForm = () => (
  <>
    <UsersArray mode="password" />
    <TLS />
    {multi('network', 'Network', ['tcp', 'udp'])}
    {select('congestion-controller', 'Congestion Controller', ['cubic', 'new_reno', 'bbr'])}
    {select('bbr-profile', 'BBR Profile', ['standard', 'conservative', 'aggressive'])}
  </>
);
