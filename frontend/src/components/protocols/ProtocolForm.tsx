import React from 'react';
import type { 协议 } from './types';
import { Hysteria2Form, ShadowsocksForm, TrojanForm, TuicForm, VlessForm, VmessForm, WireGuardForm } from './forms';

const forms: Record<协议, React.ComponentType> = {
  shadowsocks: ShadowsocksForm,
  vmess: VmessForm,
  vless: VlessForm,
  trojan: TrojanForm,
  hysteria2: Hysteria2Form,
  wireguard: WireGuardForm,
  tuic: TuicForm,
};

const 协议Form: React.FC<{ protocol?: string }> = ({ protocol }) => {
  const Component = forms[(protocol || 'shadowsocks') as 协议] || ShadowsocksForm;
  return <Component />;
};
export default 协议Form;
