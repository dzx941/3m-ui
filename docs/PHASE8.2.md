# Phase 8.2 Mihomo API Client

Added server-side Mihomo controller API client foundation.

Added:
- backend/internal/mihomo/api/client.go
- backend/internal/mihomo/api/types.go
- backend/internal/mihomo/api/mihomo.go

Purpose:
- keep Mihomo controller access inside backend
- provide typed traffic/connections data
- prepare Phase 8.3 user traffic mapping

Next:
- connect collector scheduler
- map connections to ProxyUser
- extend dashboard APIs
