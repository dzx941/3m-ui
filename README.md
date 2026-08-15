# 3M-UI — Mihomo Core Management Console

<p align="center">
  <strong>English</strong> | <a href="#中文">中文</a>
</p>

A lightweight, modern web management console for [Mihomo Core](https://github.com/MetaCubeX/mihomo), designed for VPS proxy node management, client configuration distribution, and core lifecycle control.

---

## 🏗️ Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   React 19  │────▶│  Vite Dev   │────▶│  Ant Design │
│  + Zustand  │     │   Server    │     │      6      │
└─────────────┘     └─────────────┘     └─────────────┘
       │
       ▼ HTTP/REST
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Gin Router │────▶│    JWT      │────▶│   SQLite    │
│  + GORM     │     │   Auth      │     │   (GORM)    │
└─────────────┘     └─────────────┘     └─────────────┘
       │
       ▶ exec.Command
┌─────────────┐
│ Mihomo Core │
└─────────────┘
```

---

## 🚀 Quick Start

### Prerequisites

- Go 1.22+
- Node.js 20+
- Linux VPS (Ubuntu/Debian recommended)

### Backend

```bash
cd backend
go mod tidy
go build -o ../3m-ui ./cmd/server
../3m-ui --config /etc/3m-ui/config.yaml
```

### Frontend

```bash
cd frontend
npm install
npm run dev        # Development
npm run build      # Production build → ../backend/static/
```

### One-line Install (Production)

```bash
curl -fsSL https://raw.githubusercontent.com/kazeyukiro/3m-ui/main/scripts/install.sh | bash
```

---

## ⚙️ Configuration

`/etc/3m-ui/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  public_url: "https://your-domain.com"   # ← Required for X-Forwarded-Host validation

database:
  path: "/etc/3m-ui/3m-ui.db"

jwt:
  secret: "change-me-to-a-32-char-random-string"

mihomo:
  binary_path: "/usr/local/bin/mihomo"
  config_path: "/etc/mihomo/config.yaml"

security:
  cors_origins: []    # ← Empty = deny all cross-origin; add "https://admin.your-domain.com" if needed
```

---

## 🛡️ Security Recommendations

1. **Change default password immediately** after first login (`admin`/`admin`).
2. **Use HTTPS** in production (reverse proxy with Nginx/Caddy).
3. **Set strong JWT secret** (≥ 32 random characters).
4. **Restrict CORS origins** to your admin domain only.
5. **Place mihomo binary** in `/usr/local/bin/` or `/usr/bin/` (whitelist enforced).
6. **Set `public_url`** to prevent Host header injection attacks.

---

## 📡 API Overview

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/auth/login` | No | Login, returns JWT |
| POST | `/api/v1/auth/password` | Yes | Change password |
| GET | `/api/v1/auth/me` | Yes | Current user info |
| GET | `/api/v1/dashboard` | Yes | System & mihomo status |
| GET | `/api/v1/nodes` | Yes | List listeners |
| POST | `/api/v1/nodes` | Yes | Create listener |
| GET | `/api/v1/nodes/:id/uri` | Yes | Export node URI |
| POST | `/api/v1/mihomo/start` | Yes | Start core |
| POST | `/api/v1/mihomo/stop` | Yes | Stop core |

---

## 📄 License

MIT License — see [LICENSE](LICENSE) for details.

---

---

<p id="中文"></p>

<h1 align="center">3M-UI — Mihomo 核心管理控制台</h1>

<p align="center">
  <a href="#3m-ui--mihomo-core-management-console">English</a> | <strong>中文</strong>
</p>

[Mihomo Core](https://github.com/MetaCubeX/mihomo) 的轻量级现代化 Web 管理控制台，专为 VPS 代理节点管理、客户端配置分发和核心生命周期控制而设计。


---

## 🏗️ 架构

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   React 19  │────▶│  Vite 开发  │────▶│  Ant Design │
│  + Zustand  │     │    服务器   │     │      6      │
└─────────────┘     └─────────────┘     └─────────────┘
       │
       ▼ HTTP/REST
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Gin 路由   │────▶│    JWT      │────▶│   SQLite    │
│  + GORM     │     │   认证      │     │   (GORM)    │
└─────────────┘     └─────────────┘     └─────────────┘
       │
       ▶ exec.Command
┌─────────────┐
│ Mihomo 核心 │
└─────────────┘
```

---

## 🚀 快速开始

### 环境要求

- Go 1.22+
- Node.js 20+
- Linux VPS（推荐 Ubuntu/Debian）

### 后端

```bash
cd backend
go mod tidy
go build -o ../3m-ui ./cmd/server
../3m-ui --config /etc/3m-ui/config.yaml
```

### 前端

```bash
cd frontend
npm install
npm run dev        # 开发模式
npm run build      # 生产构建 → ../backend/static/
```

### 一键安装（生产环境）

```bash
curl -fsSL https://raw.githubusercontent.com/kazeyukiro/3m-ui/main/scripts/install.sh | bash
```

---

## ⚙️ 配置说明

`/etc/3m-ui/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  public_url: "https://your-domain.com"   # ← X-Forwarded-Host 校验必需

database:
  path: "/etc/3m-ui/3m-ui.db"

jwt:
  secret: "请更换为32位以上随机字符串"

mihomo:
  binary_path: "/usr/local/bin/mihomo"
  config_path: "/etc/mihomo/config.yaml"

security:
  cors_origins: []    # ← 空列表 = 拒绝所有跨域；如需管理后台可添加 "https://admin.your-domain.com"
```

---

## 🛡️ 安全建议

1. **首次登录后立即修改默认密码**（`admin`/`admin`）。
2. **生产环境使用 HTTPS**（通过 Nginx/Caddy 反向代理）。
3. **设置强 JWT Secret**（≥ 32 位随机字符）。
4. **严格限制 CORS 来源**，仅允许管理域名访问。
5. **将 mihomo 二进制文件**放置在 `/usr/local/bin/` 或 `/usr/bin/`（已启用白名单校验）。
6. **配置 `public_url`**，防止 Host 头注入攻击。

---

## 📡 API 概览

| 方法 | 接口 | 需认证 | 说明 |
|------|------|--------|------|
| POST | `/api/v1/auth/login` | 否 | 登录，返回 JWT |
| POST | `/api/v1/auth/password` | 是 | 修改密码 |
| GET | `/api/v1/auth/me` | 是 | 当前用户信息 |
| GET | `/api/v1/dashboard` | 是 | 系统与核心状态 |
| GET | `/api/v1/nodes` | 是 | 监听器列表 |
| POST | `/api/v1/nodes` | 是 | 创建监听器 |
| GET | `/api/v1/nodes/:id/uri` | 是 | 导出节点 URI |
| POST | `/api/v1/mihomo/start` | 是 | 启动核心 |
| POST | `/api/v1/mihomo/stop` | 是 | 停止核心 |

---

## 📄 许可证

MIT License — 详见 [LICENSE](LICENSE)。
