export type ListenerFieldKind = 'text' | 'password' | 'number' | 'boolean' | 'text-list' | 'select';
export type ListenerField = {
  path: string;
  label: string;
  kind: ListenerFieldKind;
  required?: boolean;
  options?: string[];
  protocols?: string[];
  visibleWhen?: (values: Record<string, any>) => boolean;
  placeholder?: string;
};

const TLS = ['vmess', 'vless', 'trojan', 'hysteria2', 'tuic', 'anytls', 'trusttunnel'];
const REALITY = ['vmess', 'vless', 'trojan'];
const TRANSPORT = ['vmess', 'vless', 'trojan'];

/**
 * Framework-neutral listener capability manifest.
 * Every UI edition renders this same schema; only the widgets differ.
 */
export const LISTENER_FIELDS: ListenerField[] = [
  { path: 'users', label: 'Users', kind: 'text-list', protocols: ['vless', 'vmess', 'trojan', 'hysteria2', 'tuic', 'anytls'] },
  { path: 'flow', label: 'Flow', kind: 'select', options: ['', 'xtls-rprx-vision'], protocols: ['vless'] },
  { path: 'decryption', label: 'Decryption', kind: 'text', protocols: ['vless'] },
  { path: 'certificate', label: 'Certificate', kind: 'text', protocols: TLS },
  { path: 'private-key', label: 'Private Key', kind: 'password', protocols: TLS },
  { path: 'server-name', label: 'Server Name', kind: 'text', protocols: TLS },
  { path: 'alpn', label: 'ALPN', kind: 'text-list', protocols: TLS },
  { path: 'fingerprint', label: 'Fingerprint', kind: 'text', protocols: TLS },
  { path: 'allow-insecure', label: 'Allow Insecure', kind: 'boolean', protocols: TLS },
  { path: 'ws-path', label: 'WebSocket Path', kind: 'text', protocols: TRANSPORT },
  { path: 'ws-headers.Host', label: 'WebSocket Host', kind: 'text', protocols: TRANSPORT },
  { path: 'grpc-service-name', label: 'gRPC Service Name', kind: 'text', protocols: TRANSPORT },
  { path: 'xhttp-config.path', label: 'XHTTP Path', kind: 'text', protocols: ['vless'] },
  { path: 'xhttp-config.host', label: 'XHTTP Host', kind: 'text', protocols: ['vless'] },
  { path: 'xhttp-config.mode', label: 'XHTTP Mode', kind: 'select', options: ['auto', 'packet-up', 'stream-up', 'stream-one'], protocols: ['vless'] },
  { path: 'reality-config.dest', label: 'Reality Dest', kind: 'text', required: true, protocols: REALITY },
  { path: 'reality-config.private-key', label: 'Reality Private Key', kind: 'password', required: true, protocols: REALITY },
  { path: 'reality-config.short-id', label: 'Reality Short IDs', kind: 'text-list', protocols: REALITY },
  { path: 'reality-config.server-names', label: 'Reality Server Names', kind: 'text-list', protocols: REALITY },
  { path: 'mux.enabled', label: 'MUX Enabled', kind: 'boolean', protocols: ['shadowsocks', 'vmess', 'vless', 'trojan'] },
  { path: 'mux.concurrency', label: 'MUX Concurrency', kind: 'number', protocols: ['shadowsocks', 'vmess', 'vless', 'trojan'] },
  { path: 'shadow-tls.enabled', label: 'Shadow TLS Enabled', kind: 'boolean' },
  { path: 'shadow-tls.version', label: 'Shadow TLS Version', kind: 'number' },
  { path: 'shadow-tls.password', label: 'Shadow TLS Password', kind: 'password' },
  { path: 'shadow-tls.host', label: 'Shadow TLS Host', kind: 'text' },
  { path: 'shadow-tls.port', label: 'Shadow TLS Port', kind: 'number' },
  { path: 'res-tls.enabled', label: 'RES TLS Enabled', kind: 'boolean' },
  { path: 'res-tls.server', label: 'RES TLS Server', kind: 'text' },
  { path: 'res-tls.server-port', label: 'RES TLS Port', kind: 'number' },
  { path: 'res-tls.private-key', label: 'RES TLS Private Key', kind: 'password' },
  { path: 'jls-config.enabled', label: 'JLS Enabled', kind: 'boolean' },
  { path: 'jls-config.interval', label: 'JLS Interval', kind: 'number' },
  { path: 'jls-config.noise', label: 'JLS Noise', kind: 'number' },
  { path: 'obfs.type', label: 'Obfuscation Type', kind: 'text', protocols: ['shadowsocks'] },
  { path: 'obfs.password', label: 'Obfuscation Password', kind: 'password', protocols: ['shadowsocks'] },
  { path: 'kcp-tun.enabled', label: 'KCP Tun Enabled', kind: 'boolean', protocols: ['shadowsocks'] },
  { path: 'kcp-tun.mtu', label: 'KCP MTU', kind: 'number', protocols: ['shadowsocks'] },
  { path: 'kcp-tun.sndwnd', label: 'KCP Send Window', kind: 'number', protocols: ['shadowsocks'] },
  { path: 'kcp-tun.rcvwnd', label: 'KCP Receive Window', kind: 'number', protocols: ['shadowsocks'] },
  { path: 'password', label: 'Password', kind: 'password', protocols: ['shadowsocks', 'hysteria2', 'tuic', 'snell', 'anytls'] },
  { path: 'method', label: 'Method', kind: 'text', protocols: ['shadowsocks'] },
  { path: 'up', label: 'Upload Bandwidth', kind: 'number', protocols: ['hysteria2'] },
  { path: 'down', label: 'Download Bandwidth', kind: 'number', protocols: ['hysteria2'] },
  { path: 'hop-interval', label: 'Hop Interval', kind: 'text', protocols: ['hysteria2'] },
  { path: 'masquerade', label: 'Masquerade', kind: 'text', protocols: ['hysteria2'] },
  { path: 'recv-window-conn', label: 'Receive Window Conn', kind: 'number', protocols: ['hysteria2'] },
  { path: 'recv-window', label: 'Receive Window', kind: 'number', protocols: ['hysteria2'] },
  { path: 'congestion-controller', label: 'Congestion Controller', kind: 'text', protocols: ['hysteria2'] },
  { path: 'ignore-client-bandwidth', label: 'Ignore Client Bandwidth', kind: 'boolean', protocols: ['hysteria2'] },
];

export function fieldsForProtocol(protocol: string): ListenerField[] {
  return LISTENER_FIELDS.filter((field) => !field.protocols || field.protocols.includes(protocol));
}

export function getPath(obj: Record<string, any>, path: string): any {
  return path.split('.').reduce((v, key) => v == null ? undefined : v[key], obj);
}

export function setPath(obj: Record<string, any>, path: string, value: any): void {
  const parts = path.split('.');
  let cursor = obj;
  parts.slice(0, -1).forEach((part) => {
    if (!cursor[part] || typeof cursor[part] !== 'object') cursor[part] = {};
    cursor = cursor[part];
  });
  cursor[parts[parts.length - 1]] = value;
}
