import React from 'react';
import { Button, Form, Input, InputNumber, Select, Space, Switch } from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';

const item = (name: string, label: string, placeholder?: string, required = false) => (
  <Form.Item name={['protocolConfig', name]} label={label} rules={required ? [{ required: true, message: `${label} is required` }] : undefined}>
    <Input placeholder={placeholder} />
  </Form.Item>
);

const password = (name: string, label: string, placeholder?: string, required = false) => (
  <Form.Item name={['protocolConfig', name]} label={label} rules={required ? [{ required: true, message: `${label} is required` }] : undefined}>
    <Input.Password placeholder={placeholder} />
  </Form.Item>
);

const number = (name: string, label: string, min?: number, max?: number) => (
  <Form.Item name={['protocolConfig', name]} label={label}>
    <InputNumber min={min} max={max} style={{ width: '100%' }} />
  </Form.Item>
);

const boolean = (name: string, label: string) => (
  <Form.Item name={['protocolConfig', name]} label={label} valuePropName="checked">
    <Switch checkedChildren="Enabled" unCheckedChildren="Disabled" />
  </Form.Item>
);

const select = (name: string, label: string, options: string[]) => (
  <Form.Item name={['protocolConfig', name]} label={label}>
    <Select allowClear options={options.map((value) => ({ value, label: value }))} />
  </Form.Item>
);

const tags = (name: string, label: string, options: string[] = []) => (
  <Form.Item name={['protocolConfig', name]} label={label}>
    <Select mode="tags" options={options.map((value) => ({ value, label: value }))} />
  </Form.Item>
);

const nested = (path: string[], name: string, label: string, placeholder?: string) => (
  <Form.Item name={['protocolConfig', ...path, name]} label={label}>
    <Input placeholder={placeholder} />
  </Form.Item>
);

const nestedPassword = (path: string[], name: string, label: string, placeholder?: string) => (
  <Form.Item name={['protocolConfig', ...path, name]} label={label}>
    <Input.Password placeholder={placeholder} />
  </Form.Item>
);

const nestedNumber = (path: string[], name: string, label: string, min?: number, max?: number) => (
  <Form.Item name={['protocolConfig', ...path, name]} label={label}>
    <InputNumber min={min} max={max} style={{ width: '100%' }} />
  </Form.Item>
);

const nestedBoolean = (path: string[], name: string, label: string) => (
  <Form.Item name={['protocolConfig', ...path, name]} label={label} valuePropName="checked">
    <Switch checkedChildren="Enabled" unCheckedChildren="Disabled" />
  </Form.Item>
);

const nestedSelect = (path: string[], name: string, label: string, options: string[]) => (
  <Form.Item name={['protocolConfig', ...path, name]} label={label}>
    <Select allowClear options={options.map((value) => ({ value, label: value }))} />
  </Form.Item>
);

const UserList = ({ kind }: { kind: 'password' | 'uuid' | 'tuic' }) => (
  <Form.List name={['protocolConfig', 'users']}>
    {(fields, { add, remove }) => (
      <>
        <Form.Item label="Users">
          {fields.map((field) => (
            <Space key={field.key} align="baseline" style={{ display: 'flex', marginBottom: 8 }}>
              <Form.Item {...field} name={[field.name, 'username']} noStyle>
                <Input placeholder="Username / UUID" />
              </Form.Item>
              {kind === 'uuid' ? (
                <Form.Item {...field} name={[field.name, 'uuid']} noStyle>
                  <Input placeholder="UUID" />
                </Form.Item>
              ) : (
                <Form.Item {...field} name={[field.name, 'password']} noStyle>
                  <Input.Password placeholder="Password" />
                </Form.Item>
              )}
              {kind === 'uuid' && (
                <Form.Item {...field} name={[field.name, 'flow']} noStyle>
                  <Input placeholder="Flow" />
                </Form.Item>
              )}
              {kind === 'uuid' && (
                <Form.Item {...field} name={[field.name, 'alterId']} noStyle>
                  <InputNumber min={0} placeholder="alterId" />
                </Form.Item>
              )}
              <MinusCircleOutlined onClick={() => remove(field.name)} />
            </Space>
          ))}
          <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>Add user</Button>
        </Form.Item>
      </>
    )}
  </Form.List>
);

const UserMap = ({ label = 'Users' }: { label?: string }) => (
  <Form.List name={['protocolConfig', 'users']}>
    {(fields, { add, remove }) => (
      <Form.Item label={label}>
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

const TLSFields = ({ allowInsecure = false }: { allowInsecure?: boolean }) => (
  <>
    {item('certificate', 'Certificate', '/etc/3m-ui/certs/server.crt')}
    {item('private-key', 'Private Key', '/etc/3m-ui/certs/server.key')}
    {select('client-auth-type', 'Client Auth Type', ['request', 'require-any', 'verify-if-given', 'require-and-verify'])}
    {item('client-auth-cert', 'Client Auth Certificate')}
    {item('ech-key', 'ECH Key')}
    {allowInsecure && boolean('allow-insecure', 'Allow Insecure')}
  </>
);

const MuxFields = () => (
  <>
    {boolean('mux-option.enabled', 'MUX Enabled')}
    {boolean('mux-option.padding', 'MUX Padding')}
    {boolean('mux-option.brutal.enabled', 'MUX Brutal')}
    {number('mux-option.brutal.up', 'MUX Brutal Up (Mbps)', 0)}
    {number('mux-option.brutal.down', 'MUX Brutal Down (Mbps)', 0)}
  </>
);

const JlsFields = () => (
  <>
    {boolean('jls-config.enable', 'JLS Enabled')}
    {item('jls-config.dest', 'JLS Destination', 'example.com:443')}
    {item('jls-config.sni', 'JLS SNI')}
    {tags('jls-config.alpn', 'JLS ALPN', ['h2', 'http/1.1'])}
    {item('jls-config.proxy', 'JLS Proxy')}
    {number('jls-config.rate-limit', 'JLS Rate Limit (bit/s)', 0)}
    <Form.List name={['protocolConfig', 'jls-config', 'users']}>
      {(fields, { add, remove }) => (
        <Form.Item label="JLS Users">
          {fields.map((field) => (
            <Space key={field.key} style={{ display: 'flex', marginBottom: 8 }}>
              <Form.Item {...field} name={[field.name, 'username']} noStyle><Input placeholder="Username" /></Form.Item>
              <Form.Item {...field} name={[field.name, 'password']} noStyle><Input.Password placeholder="Password" /></Form.Item>
              <MinusCircleOutlined onClick={() => remove(field.name)} />
            </Space>
          ))}
          <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>Add JLS user</Button>
        </Form.Item>
      )}
    </Form.List>
  </>
);

const ShadowTlsFields = () => (
  <>
    {boolean('shadow-tls.enable', 'ShadowTLS Enabled')}
    {number('shadow-tls.version', 'ShadowTLS Version', 1, 3)}
    {password('shadow-tls.password', 'ShadowTLS Password')}
    {item('shadow-tls.handshake.dest', 'ShadowTLS Handshake Destination', 'example.com:443')}
    {item('shadow-tls.handshake.proxy', 'ShadowTLS Handshake Proxy')}
    <Form.List name={['protocolConfig', 'shadow-tls', 'users']}>
      {(fields, { add, remove }) => (
        <Form.Item label="ShadowTLS v3 Users">
          {fields.map((field) => (
            <Space key={field.key} style={{ display: 'flex', marginBottom: 8 }}>
              <Form.Item {...field} name={[field.name, 'name']} noStyle><Input placeholder="Name" /></Form.Item>
              <Form.Item {...field} name={[field.name, 'password']} noStyle><Input.Password placeholder="Password" /></Form.Item>
              <MinusCircleOutlined onClick={() => remove(field.name)} />
            </Space>
          ))}
          <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>Add ShadowTLS user</Button>
        </Form.Item>
      )}
    </Form.List>
  </>
);

const ResTlsFields = () => (
  <>
    {boolean('res-tls.enable', 'ResTLS Enabled')}
    {item('res-tls.dest', 'ResTLS Destination', 'example.com:443')}
    {password('res-tls.password', 'ResTLS Password')}
    {item('res-tls.restls-script', 'ResTLS Script')}
    {number('res-tls.min-record-len', 'ResTLS Min Record Length', 0)}
    {item('res-tls.proxy', 'ResTLS Proxy')}
  </>
);

const RealityFields = ({ withLimits = true }: { withLimits?: boolean }) => (
  <>
    {item('reality-config.dest', 'Reality Destination', 'example.com:443')}
    {password('reality-config.private-key', 'Reality Private Key')}
    {tags('reality-config.short-id', 'Reality Short IDs')}
    {tags('reality-config.server-names', 'Reality Server Names')}
    {withLimits && <>
      {number('reality-config.limit-fallback-upload.after-bytes', 'Reality Upload Limit After Bytes', 0)}
      {number('reality-config.limit-fallback-upload.bytes-per-sec', 'Reality Upload Bytes/s', 0)}
      {number('reality-config.limit-fallback-upload.burst-bytes-per-sec', 'Reality Upload Burst Bytes/s', 0)}
      {number('reality-config.limit-fallback-download.after-bytes', 'Reality Download Limit After Bytes', 0)}
      {number('reality-config.limit-fallback-download.bytes-per-sec', 'Reality Download Bytes/s', 0)}
      {number('reality-config.limit-fallback-download.burst-bytes-per-sec', 'Reality Download Burst Bytes/s', 0)}
    </>}
  </>
);

export const ShadowsocksForm = () => (
  <>
    {select('cipher', 'Cipher', ['2022-blake3-aes-128-gcm', '2022-blake3-aes-256-gcm', '2022-blake3-chacha20-poly1305', 'none', 'aes-128-gcm', 'aes-192-gcm', 'aes-256-gcm', 'chacha20-ietf-poly1305', 'xchacha20-ietf-poly1305'])}
    {password('password', 'Password', undefined, true)}
    {boolean('udp', 'UDP')}
    {boolean('simple-obfs.enable', 'Simple Obfs Enabled')}
    {select('simple-obfs.mode', 'Simple Obfs Mode', ['http', 'tls'])}
    <ShadowTlsFields />
    <ResTlsFields />
    <JlsFields />
    {boolean('kcp-tun.enable', 'KCP Tunnel Enabled')}
    {password('kcp-tun.key', 'KCP Key')}
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
    {boolean('kcp-tun.nocomp', 'KCP No Compression')}
    {boolean('kcp-tun.acknodelay', 'KCP Ack No Delay')}
    {number('kcp-tun.nodelay', 'KCP No Delay', 0)}
    {number('kcp-tun.interval', 'KCP Interval (ms)', 0)}
    {number('kcp-tun.resend', 'KCP Resend', 0)}
    {number('kcp-tun.sockbuf', 'KCP Socket Buffer', 0)}
    {number('kcp-tun.smuxver', 'KCP SMUX Version', 1, 2)}
    {number('kcp-tun.smuxbuf', 'KCP SMUX Buffer', 0)}
    {number('kcp-tun.framesize', 'KCP Frame Size', 0)}
    {number('kcp-tun.streambuf', 'KCP Stream Buffer', 0)}
    {number('kcp-tun.keepalive', 'KCP Keepalive (s)', 0)}
    <MuxFields />
  </>
);

export const SnellForm = () => (
  <>
    {password('psk', 'PSK', undefined, true)}
    {number('version', 'Version', 1, 5)}
    {boolean('udp', 'UDP')}
    {select('obfs-opts.mode', 'Obfs Mode', ['http', 'tls'])}
    {item('obfs-opts.host', 'Obfs Host')}
    <ShadowTlsFields />
    <ResTlsFields />
    <JlsFields />
  </>
);

const TransportFields = () => (
  <>
    {item('ws-path', 'WebSocket Path', '/')}
    {item('grpc-service-name', 'gRPC Service Name', 'GunService')}
  </>
);

export const VmessForm = () => (
  <>
    <UserList kind="uuid" />
    {TransportFields()}
    {boolean('mekya-config.enable', 'Mekya Enabled')}
    {number('mekya-config.max-write-size', 'Mekya Max Write Size', 0)}
    {number('mekya-config.max-write-duration-ms', 'Mekya Max Write Duration (ms)', 0)}
    {number('mekya-config.max-simultaneous-write-connection', 'Mekya Max Simultaneous Writes', 0)}
    {number('mekya-config.packet-writing-buffer', 'Mekya Packet Buffer', 0)}
    {number('mekya-config.kcp.mtu', 'Mekya KCP MTU', 1)}
    {number('mekya-config.kcp.tti', 'Mekya KCP TTI', 0)}
    {number('mekya-config.kcp.uplink-capacity', 'Mekya KCP Uplink', 0)}
    {number('mekya-config.kcp.downlink-capacity', 'Mekya KCP Downlink', 0)}
    {boolean('mekya-config.kcp.congestion', 'Mekya KCP Congestion')}
    {number('mekya-config.kcp.write-buffer', 'Mekya KCP Write Buffer', 0)}
    {number('mekya-config.kcp.read-buffer', 'Mekya KCP Read Buffer', 0)}
    {item('mekya-config.kcp.seed', 'Mekya KCP Seed')}
    {select('mekya-config.kcp.header', 'Mekya KCP Header', ['none', 'srtp', 'utp', 'wechat-video', 'dtls', 'wireguard'])}
    {boolean('mkcp-config.enable', 'mKCP Enabled')}
    {number('mkcp-config.mtu', 'mKCP MTU', 1)}
    {number('mkcp-config.tti', 'mKCP TTI', 0)}
    {number('mkcp-config.uplink-capacity', 'mKCP Uplink', 0)}
    {number('mkcp-config.downlink-capacity', 'mKCP Downlink', 0)}
    {boolean('mkcp-config.congestion', 'mKCP Congestion')}
    {number('mkcp-config.write-buffer', 'mKCP Write Buffer', 0)}
    {number('mkcp-config.read-buffer', 'mKCP Read Buffer', 0)}
    {item('mkcp-config.seed', 'mKCP Seed')}
    {select('mkcp-config.header', 'mKCP Header', ['none', 'srtp', 'utp', 'wechat-video', 'dtls', 'wireguard'])}
    <TLSFields />
    <JlsFields />
    <ShadowTlsFields />
    <ResTlsFields />
    <RealityFields />
    {item('tlsmirror-config.dest', 'TLSMirror Destination')}
    {password('tlsmirror-config.primary-key', 'TLSMirror Primary Key')}
    {MuxFields />}
  </>
);

export const VlessForm = () => (
  <>
    <UserList kind="uuid" />
    {TransportFields()}
    {item('xhttp-config.path', 'XHTTP Path', '/')}
    {item('xhttp-config.host', 'XHTTP Host')}
    {select('xhttp-config.mode', 'XHTTP Mode', ['auto', 'stream-one', 'stream-up', 'packet-up'])}
    {boolean('xhttp-config.no-sse-header', 'XHTTP No SSE Header')}
    {item('xhttp-config.x-padding-bytes', 'XHTTP Padding Bytes')}
    {boolean('xhttp-config.x-padding-obfs-mode', 'XHTTP Padding Obfs Mode')}
    {item('xhttp-config.x-padding-key', 'XHTTP Padding Key')}
    {item('xhttp-config.x-padding-header', 'XHTTP Padding Header')}
    {select('xhttp-config.x-padding-placement', 'XHTTP Padding Placement', ['queryInHeader', 'cookie', 'header', 'query'])}
    {select('xhttp-config.x-padding-method', 'XHTTP Padding Method', ['repeat-x', 'tokenish'])}
    {select('xhttp-config.uplink-http-method', 'XHTTP Uplink HTTP Method', ['POST', 'PUT', 'PATCH', 'DELETE'])}
    {select('xhttp-config.session-placement', 'XHTTP Session Placement', ['path', 'query', 'cookie', 'header'])}
    {item('xhttp-config.session-key', 'XHTTP Session Key')}
    {select('xhttp-config.session-table', 'XHTTP Session Table', ['', 'uuid', 'ALPHABET', 'Alphabet', 'BASE36', 'Base62', 'HEX', 'alphabet', 'base36', 'hex', 'number'])}
    {item('xhttp-config.session-length', 'XHTTP Session Length', '16-32')}
    {select('xhttp-config.seq-placement', 'XHTTP Sequence Placement', ['path', 'query', 'cookie', 'header'])}
    {item('xhttp-config.seq-key', 'XHTTP Sequence Key')}
    {select('xhttp-config.uplink-data-placement', 'XHTTP Uplink Data Placement', ['body', 'cookie', 'header'])}
    {item('xhttp-config.uplink-data-key', 'XHTTP Uplink Data Key')}
    {number('xhttp-config.uplink-chunk-size', 'XHTTP Uplink Chunk Size', 0)}
    {number('xhttp-config.sc-max-buffered-posts', 'XHTTP Max Buffered Posts', 0)}
    {item('xhttp-config.sc-stream-up-server-secs', 'XHTTP Stream Up Server Seconds')}
    {number('xhttp-config.sc-max-each-post-bytes', 'XHTTP Max Each Post Bytes', 0)}
    {item('decryption', 'VLESS Decryption')}
    <TLSFields allowInsecure />
    <RealityFields />
    <ShadowTlsFields />
    <ResTlsFields />
    <JlsFields />
    <MuxFields />
  </>
);

export const TrojanForm = () => (
  <>
    <UserList kind="password" />
    {TransportFields()}
    <TLSFields allowInsecure />
    <RealityFields />
    <ShadowTlsFields />
    <ResTlsFields />
    <JlsFields />
    {boolean('ss-option.enabled', 'Trojan SS Enabled')}
    {select('ss-option.method', 'Trojan SS Method', ['aes-128-gcm', 'aes-256-gcm', 'chacha20-ietf-poly1305'])}
    {password('ss-option.password', 'Trojan SS Password')}
    <MuxFields />
  </>
);

export const Hysteria2Form = () => (
  <>
    <UserMap />
    {item('up', 'Upload Speed', '1000')}
    {item('down', 'Download Speed', '1000')}
    {boolean('ignore-client-bandwidth', 'Ignore Client Bandwidth')}
    {select('obfs', 'Obfuscation', ['salamander'])}
    {password('obfs-password', 'Obfuscation Password')}
    {item('masquerade', 'Masquerade', 'https://example.com')}
    {item('bbr-profile', 'BBR Profile')}
    {tags('alpn', 'ALPN', ['h3'])}
    <TLSFields />
    {boolean('realm-opts.enable', 'Realm Options Enabled')}
    {item('realm-opts.server-url', 'Realm Server URL')}
    {password('realm-opts.token', 'Realm Token')}
    {item('realm-opts.realm-id', 'Realm ID')}
    {tags('realm-opts.stun-servers', 'Realm STUN Servers')}
    <MuxFields />
  </>
);

export const Hysteria2RealmForm = () => null;

export const TuicForm = () => (
  <>
    <UserMap label="TUIC V5 Users" />
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
    <TLSFields />
    {select('congestion-controller', 'Congestion Controller', ['cubic', 'new_reno', 'bbr'])}
    {select('bbr-profile', 'BBR Profile', ['standard', 'conservative', 'aggressive'])}
    {number('max-idle-time', 'Max Idle Time (ms)', 0)}
    {number('authentication-timeout', 'Authentication Timeout (ms)', 0)}
    {tags('alpn', 'ALPN', ['h3'])}
    {number('max-udp-relay-packet-size', 'Max UDP Relay Packet Size', 1, 65535)}
    <MuxFields />
  </>
);

export const ShadowQuicForm = () => (
  <>
    <UserList kind="password" />
    {item('jls-upstream.addr', 'JLS Upstream Address', 'www.example.com:443')}
    {item('jls-upstream.sni', 'JLS Upstream SNI')}
    {item('jls-upstream.proxy', 'JLS Upstream Proxy')}
    {number('jls-upstream.rate-limit', 'JLS Upstream Rate Limit (bit/s)', 0)}
    {tags('alpn', 'ALPN', ['h3'])}
    {tags('quic-versions', 'QUIC Versions', ['v1', 'v2'])}
    {boolean('zero-rtt', 'Zero RTT')}
    {select('congestion-controller', 'Congestion Controller', ['cubic', 'new_reno', 'bbr'])}
    {item('up', 'Upload Speed')}
    {item('down', 'Download Speed')}
    {boolean('ignore-client-bandwidth', 'Ignore Client Bandwidth')}
    {number('cwnd', 'CWND', 1)}
    {select('bbr-profile', 'BBR Profile', ['standard', 'conservative', 'aggressive'])}
    {number('max-idle-time', 'Max Idle Time (ms)', 0)}
    {number('max-datagram-frame-size', 'Max Datagram Frame Size', 1, 65535)}
    {number('recv-window-conn', 'Receive Window Conn', 0)}
    {number('recv-window', 'Receive Window', 0)}
    {boolean('disable-mtu-discovery', 'Disable MTU Discovery')}
  </>
);

export const AnyTLSForm = () => (
  <>
    <UserMap />
    <TLSFields allowInsecure />
    <ShadowTlsFields />
    <ResTlsFields />
    <JlsFields />
    {item('padding-scheme', 'Padding Scheme')}
  </>
);

export const MieruForm = () => (
  <>
    {select('transport', 'Transport', ['TCP', 'UDP'])}
    <UserMap />
    {item('traffic-pattern', 'Traffic Pattern (Base64)')}
    {boolean('user-hint-is-mandatory', 'User Hint Is Mandatory')}
  </>
);

export const SudokuForm = () => (
  <>
    {item('key', 'Server Key', undefined, true)}
    {select('aead-method', 'AEAD Method', ['chacha20-poly1305', 'aes-128-gcm', 'none'])}
    {number('padding-min', 'Padding Min', 0, 100)}
    {number('padding-max', 'Padding Max', 0, 100)}
    {select('table-type', 'Table Type', ['prefer_ascii', 'prefer_entropy', 'up_ascii_down_entropy', 'up_entropy_down_ascii'])}
    {item('custom-table', 'Custom Table')}
    {tags('custom-tables', 'Custom Tables')}
    {number('handshake-timeout', 'Handshake Timeout (s)', 0)}
    {boolean('enable-pure-downlink', 'Pure Downlink')}
    {boolean('httpmask.disable', 'HTTPMask Disabled')}
    {select('httpmask.mode', 'HTTPMask Mode', ['legacy', 'stream', 'poll', 'auto', 'ws'])}
    {item('httpmask.path-root', 'HTTPMask Path Root')}
    {boolean('disable-http-mask', 'Disable HTTP Mask (compatibility)')}
    {select('http-mask-mode', 'HTTP Mask Mode (compatibility)', ['legacy', 'stream', 'poll', 'auto', 'ws'])}
    {item('path-root', 'Path Root (compatibility)')}
    {item('fallback', 'Fallback Address')}
    <MuxFields />
  </>
);

export const TrustTunnelForm = () => (
  <>
    <UserList kind="password" />
    <TLSFields />
    {tags('network', 'Network', ['tcp', 'udp'])}
    {select('congestion-controller', 'Congestion Controller', ['cubic', 'new_reno', 'bbr'])}
    {select('bbr-profile', 'BBR Profile', ['standard', 'conservative', 'aggressive'])}
  </>
);
