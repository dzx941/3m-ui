# 3m-ui

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
   go run cmd/server/main.go
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
