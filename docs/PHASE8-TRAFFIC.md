# Phase 8 — Traffic & Connections

## Goal

Add real-time Mihomo traffic and connection monitoring without coupling the
dashboard directly to the Mihomo HTTP API.

Architecture:

Mihomo API
  ↓
backend/internal/mihomo
  ↓
backend/internal/traffic
  ↓
Dashboard / Users / Nodes

## Required metrics

### Global
- upload bytes/sec
- download bytes/sec
- total upload bytes
- total download bytes
- active connection count

### Per user
- traffic used
- last seen
- online state

### Per node
- active connections
- upload/download rate where available

## Safety

- Never expose Mihomo's controller secret to the browser.
- Keep controller access server-side.
- Use short-lived in-memory polling/cache for expensive endpoints.
- Treat Mihomo API fields as optional because different core versions may expose
  different connection metadata.
- Do not count failed/closed connections as active users.

## API proposal

GET /api/v1/traffic/status
GET /api/v1/traffic/users
GET /api/v1/traffic/nodes
GET /api/v1/connections

The dashboard should continue using its unified endpoint and should not make
direct requests to Mihomo's controller.

## Verification

Run:

```bash
cd backend
go test ./...
go vet ./...

cd ../frontend
pnpm build
```

Then verify against a real Mihomo instance with active traffic.
