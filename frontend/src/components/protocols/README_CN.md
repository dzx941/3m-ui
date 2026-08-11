# 节点可视化配置

当前节点编辑器通过 ProtocolForm 根据协议动态渲染：

- Shadowsocks
- VMess
- VLESS
- Trojan
- Hysteria2
- TUIC
- WireGuard

新增协议时只需扩展 forms.tsx 与 types.ts。

配置页面通过表单生成 Mihomo YAML。
