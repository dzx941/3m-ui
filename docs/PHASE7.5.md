# Phase 7.5 — Subscription UI & Protocol Conversion

## Goal

Complete the subscription workflow:

ProxyUser → ListenerUser → Listener → client proxy conversion → subscription response.

## Required client proxy fields

- Shadowsocks: type, server, port, cipher, password, udp
- VMess: type, server, port, uuid, alterId, cipher, tls, udp
- VLESS: type, server, port, uuid, tls, flow, udp
- Trojan: type, server, port, password, sni, tls, udp
- Hysteria2: type, server, port, password, sni, tls
- TUIC: type, server, port, uuid, password, sni, tls

The server-side Listener configuration remains the source of truth. Subscription generation must never expose private keys or unrelated server-only configuration.

## Security

Subscription tokens are bearer credentials. Do not log them, return them in normal list responses, or expose server private keys.

Recommended endpoints:

- `GET /api/v1/sub/:token` — YAML
- `GET /api/v1/sub/:token?format=base64` — Base64-encoded YAML
- authenticated admin endpoints for creating, rotating, and revoking subscriptions.

## Validation

Before release, run:

```bash
cd backend
go test ./...
go vet ./...

cd ../frontend
pnpm build
```

Then verify subscription import with a real Mihomo-compatible client.
