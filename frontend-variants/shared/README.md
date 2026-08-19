# Shared UI Contract

All alternative editions consume the same backend API and the same listener configuration normalization contract. UI libraries are intentionally isolated from business logic.

## Parity contract

Every edition must implement the same routes and actions:

- Login / first-password-change / change-password
- Dashboard
- Mihomo core control and logs
- Listeners: CRUD, enable/disable, reload, URI, advanced protocol/transport/security forms
- Users: CRUD, traffic reset, limits, expiration, blocking and subscription data
- Traffic
- Routing
- Cluster: CRUD and health checks
- Config: view, validate and generate
- Settings

Listener form data MUST round-trip through the canonical JSON representation. In particular Reality values are represented as `reality-config.dest`, `reality-config.private-key`, `reality-config.short-id`, and `reality-config.server-names`; UI aliases are never sent to the backend.
