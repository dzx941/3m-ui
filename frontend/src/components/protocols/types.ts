export type Protocol = 'shadowsocks' | 'vmess' | 'vless' | 'trojan' | 'hysteria2' | 'wireguard' | 'tuic';

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
