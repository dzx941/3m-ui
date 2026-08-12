export type ListenerProtocol =
  | 'socks'
  | 'http'
  | 'tproxy'
  | 'redir'
  | 'mixed'
  | 'tunnel'
  | 'tun'
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

export const LISTENER_PROTOCOLS: ListenerProtocol[] = [
  'socks',
  'http',
  'tproxy',
  'redir',
  'mixed',
  'tunnel',
  'tun',
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
