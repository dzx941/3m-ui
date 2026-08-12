import React from 'react';
import {
  AnyTLSForm,
  HttpForm,
  Hysteria2Form,
  Hysteria2RealmForm,
  MieruForm,
  MixedForm,
  RedirForm,
  ShadowsocksForm,
  ShadowQuicForm,
  SnellForm,
  SocksForm,
  SudokuForm,
  TproxyForm,
  TrustTunnelForm,
  TuicForm,
  TunnelForm,
  TunForm,
  VlessForm,
  VmessForm,
  TrojanForm,
} from './forms';
import type { ListenerProtocol } from './types';

const forms: Record<ListenerProtocol, React.ComponentType> = {
  socks: SocksForm,
  http: HttpForm,
  tproxy: TproxyForm,
  redir: RedirForm,
  mixed: MixedForm,
  tunnel: TunnelForm,
  tun: TunForm,
  shadowsocks: ShadowsocksForm,
  snell: SnellForm,
  vmess: VmessForm,
  vless: VlessForm,
  trojan: TrojanForm,
  hysteria2: Hysteria2Form,
  'hysteria2-realm': Hysteria2RealmForm,
  tuic: TuicForm,
  shadowquic: ShadowQuicForm,
  anytls: AnyTLSForm,
  mieru: MieruForm,
  sudoku: SudokuForm,
  trusttunnel: TrustTunnelForm,
};

const ProtocolForm: React.FC<{ protocol?: string }> = ({ protocol }) => {
  const Component = forms[(protocol || 'shadowsocks') as ListenerProtocol] || ShadowsocksForm;
  return <Component />;
};

export default ProtocolForm;
