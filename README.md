# 3m-ui

**Current version:** `v0.1.0-rc.1`

3m-ui is an easy-to-use VPS Web management panel built on top of the Mihomo Core, leveraging its inbound and listener capabilities. It is designed to act similarly to 3x-ui, but with a robust Mihomo backend core.

## Tech Stack

### Backend
- **Language**: Go 1.25+
- **Framework**: Gin Gonic
- **ORM**: GORM
- **Database**: SQLite
- **Auth**: Bcrypt & standard JWT placeholder

### Frontend
- **Framework**: React 19
- **Build Tool**: Vite
- **Language**: TypeScript
- **UI Components**: Ant Design (v6)
- **Routing**: React Router (BrowserRouter)

---

## Directory Structure

```text
3m-ui/
├── backend/                  # Go Backend Code
│   ├── cmd/
│   │   └── server/           # Main entry point (main.go)
│   ├── config/               # Configuration YAML files (config.yaml)
│   ├── data/                 # SQLite databases (.db files)
│   └── internal/             # Private application and business logic
│       ├── auth/             # Cryptography/Auth helpers (bcrypt, jwt)
│       ├── config/           # Config loading package
│       ├── database/         # Database GORM models and initialization
│       │   └── models/       # Model definitions (User, Listener, etc.)
│       ├── listener/         # Listener management logic
│       ├── mihomo/           # Mihomo Core interaction logic
│       ├── router/           # Gin Web routing setup
│       ├── subscription/     # Subscription logic
│       └── system/           # Linux system operating tasks
├── frontend/                 # React Frontend Code
│   ├── src/
│   │   ├── api/              # API Client interfaces
│   │   ├── components/       # Shared UI components
│   │   ├── layouts/          # Page layouts (Sidebar, header, etc.)
│   │   ├── pages/            # View pages (Dashboard, Listeners, Subscriptions, etc.)
│   │   ├── stores/           # Zustand / Redux state stores
│   │   ├── App.tsx           # Route registrations
│   │   └── main.tsx          # Application bootstrap
├── deploy/                   # Deployment scripts (.gitkeep)
└── docs/                     # Architectural documents (.gitkeep)
```

---

## Getting Started

### Prerequisites
- [Go 1.25+](https://go.dev/doc/install)
- [NodeJS v22+](https://nodejs.org/en)
- [pnpm](https://pnpm.io/)

### Running Backend (Development)

1. Navigate to the backend folder:
   ```bash
   cd backend
   ```
2. Initialize dependencies & run test checks:
   ```bash
   go vet ./...
   go test ./...
   ```
3. Run the Go server:
   ```bash
   go run ./cmd/server
   ```
   The backend API will be accessible at `http://localhost:8080/api/v1`.

### Running Frontend (Development)

1. Navigate to the frontend folder:
   ```bash
   cd frontend
   ```
2. Install dependencies:
   ```bash
   pnpm install
   ```
3. Run the Vite development server:
   ```bash
   pnpm dev
   ```
   The frontend will be accessible at `http://localhost:5173`.

### Production Build

#### Backend Compile
```bash
cd backend
go build -o server cmd/server/main.go
```

#### Frontend Compile
```bash
cd frontend
pnpm build
```
This generates the static bundle in `frontend/dist` which can be served by any static asset host or backend handler.

---

## Production Installation

3m-ui is distributed as a single Linux server binary with the built frontend embedded into the Go executable. Official release assets include Linux binaries for `amd64`, `arm64`, and `armv7`, plus installer, updater, and uninstaller scripts. The first release-candidate line is `v0.1.0-rc.1`.

### Supported Platforms

- Debian and Ubuntu with systemd
- Alpine Linux with OpenRC
- CPU architectures: `amd64`, `arm64`, `armv7`

### Install

Run the installer as root on the target server:

```bash
curl -fsSL https://github.com/dzx941/3m-ui/releases/latest/download/install.sh | sh
```

The installer detects the OS, CPU architecture, and init system, downloads the matching release binary from `https://github.com/dzx941/3m-ui/releases`, writes the service definition, enables the service, and starts 3m-ui automatically.

After installation, open:

```text
http://SERVER_IP:8080/
```

### Version

After installation, verify the installed binary and build metadata:

```bash
3m-ui --version
```

### Upgrade

Upgrade keeps the database, user data, Mihomo configuration files, and existing application configuration. It backs up `/etc/3m-ui` before replacing the binary.

```bash
curl -fsSL https://github.com/dzx941/3m-ui/releases/latest/download/update.sh | sh
```

To upgrade to a specific release candidate manually, download the matching binary from that GitHub Release, stop the service, replace `/usr/local/bin/3m-ui`, and start the service again. Keep `/etc/3m-ui` and `/var/lib/3m-ui` in place to preserve configuration and the SQLite database.

### Uninstall

Remove the binary, service, configuration, and logs while keeping persistent data in `/var/lib/3m-ui`:

```bash
curl -fsSL https://github.com/dzx941/3m-ui/releases/latest/download/uninstall.sh | sh
```

Remove all data as well:

```bash
curl -fsSL https://github.com/dzx941/3m-ui/releases/latest/download/uninstall.sh -o uninstall.sh
sh uninstall.sh --purge
```

### Production Directory Layout

```text
/usr/local/bin/3m-ui       # Installed server binary
/etc/3m-ui/config.yaml     # Application configuration
/var/lib/3m-ui/            # Persistent application data
/var/lib/3m-ui/3m-ui.db    # SQLite database
/var/lib/3m-ui/mihomo/     # Generated Mihomo configuration files
/var/log/3m-ui/            # Log directory reserved for deployments
```

### Service Management

systemd:

```bash
systemctl status 3m-ui
systemctl restart 3m-ui
```

OpenRC:

```bash
rc-service 3m-ui status
rc-service 3m-ui restart
```

## Release and Packaging

The release pipeline runs from `.github/workflows/release.yml` on `v*` tags and manual dispatch. It builds the frontend with pnpm, copies `frontend/dist` into the backend embed directory, cross-compiles Linux binaries into `dist/`, copies deployment scripts, and publishes all artifacts to a GitHub Release.

Local packaging commands:

```bash
make build        # Build embedded frontend and host binary into dist/3m-ui
make build-linux  # Build Linux release binaries
make release      # Clean, build, cross-compile, and copy scripts
make clean        # Remove dist/
```

## Release Candidate Checklist

See [`RELEASE_CHECKLIST.md`](RELEASE_CHECKLIST.md) before tagging or publishing a release candidate.
