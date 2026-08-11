import React from 'react';
import type { Protocol } from './types';
import { Hysteria2Form, ShadowsocksForm, TrojanForm, TuicForm, VlessForm, VmessForm, WireGuardForm } from './forms';

const forms: Record<Protocol, React.ComponentType> = {
  shadowsocks: ShadowsocksForm,
  vmess: VmessForm,
  vless: VlessForm,
  trojan: TrojanForm,
  hysteria2: Hysteria2Form,
  wireguard: WireGuardForm,
  tuic: TuicForm,
};

const ProtocolForm: React.FC<{ protocol?: string }> = ({ protocol }) => {
  const Component = forms[(protocol || 'shadowsocks') as Protocol] || ShadowsocksForm;
  return <Component />;
};
export default ProtocolForm;
