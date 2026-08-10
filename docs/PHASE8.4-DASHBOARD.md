# Phase 8.4 — Traffic Dashboard

## Goal

Expose traffic information through the existing unified dashboard without
making the browser call Mihomo's controller directly.

Dashboard data flow:

3m-ui `/api/v1/dashboard`
→ Mihomo status
→ system metrics
→ listener statistics
→ traffic status

## UI

Add/retain these cards:

- Mihomo service status
- CPU / Memory / Disk
- Upload / Download rate
- Total upload / download
- Active connections
- Listener statistics

Use the existing dashboard polling interval (10 seconds) unless the backend
adds a websocket/SSE stream later.

## Users and Nodes

The Users page should show:
- online/offline
- traffic used
- remaining quota
- last seen

The Nodes page should show:
- active connections
- traffic counters/rates where available

## Important

Do not fabricate per-user traffic from global `/traffic` counters. Per-user
accounting must be based on connection metadata or an explicit authenticated
Mihomo accounting mechanism.
