export type ListenerProtocol =
  | 'socks'
  | 'http'
  | 'tproxy'
  | 'redir'
  | 'mixed'
  | 'tunnel'
  | 'shadowsocks'
  | 'snell'
  | 'vmess'
  | 'vless'
  | 'trojan'
  | 'hysteria2'
  | 'hysteria2-realm'
  | 'tuic'
  | 'shadowquic'
  | 'anytls'
  | 'mieru'
  | 'sudoku'
  | 'trusttunnel';

export type Protocol = ListenerProtocol;

/**
 * Mihomo listener types supported by the listener configuration UI.
 *
 * TUN is intentionally excluded: it is a top-level transparent-proxy/TUN
 * feature, not a listener protocol, and therefore must not be serialized as
 * an entry in the `listeners` array.
 */
export const LISTENER_PROTOCOLS: ListenerProtocol[] = [
  'socks',
  'http',
  'tproxy',
  'redir',
  'mixed',
  'tunnel',
  'shadowsocks',
  'snell',
  'vmess',
  'vless',
  'trojan',
  'hysteria2',
  'hysteria2-realm',
  'tuic',
  'shadowquic',
  'anytls',
  'mieru',
  'sudoku',
  'trusttunnel',
];

export interface ProxyNode {
  id?: string;
  name: string;
  type: Protocol;
  server: string;
  port: number;
  uuid?: string;
  password?: string;
  cipher?: string;
  tls?: boolean;
  sni?: string;
  servername?: string;
  flow?: string;
  path?: string;
  host?: string;
  [key: string]: unknown;
}
