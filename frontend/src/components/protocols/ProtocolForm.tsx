import React from 'react';
import type { ListenerProtocol } from './types';
import {
  AnyTLSForm,
  Hysteria2Form,
  Hysteria2RealmForm,
  MieruForm,
  ShadowsocksForm,
  ShadowQuicForm,
  SnellForm,
  SudokuForm,
  TrojanForm,
  TrustTunnelForm,
  TuicForm,
  VlessForm,
  VmessForm,
} from './forms';

const forms: Record<ListenerProtocol, React.ComponentType> = {
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
