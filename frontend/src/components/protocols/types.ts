export type ListenerProtocol =
  | 'shadowsocks'
  | 'snell'
  | 'vmess'
  | 'vless'
  | 'trojan'
  | 'hysteria2'
  | 'tuic'
  | 'shadowquic'
  | 'anytls'
  | 'mieru'
  | 'sudoku'
  | 'trusttunnel';

export type Protocol = ListenerProtocol;

export const LISTENER_PROTOCOLS: ListenerProtocol[] = [
  'shadowsocks',
  'snell',
  'vmess',
  'vless',
  'trojan',
  'hysteria2',
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
  type: ListenerProtocol;
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
