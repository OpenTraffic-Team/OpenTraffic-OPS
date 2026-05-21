# RTM Monitoring & Operations Platform

RTM (Real-Time Monitor) is a full-stack monitoring and operations management platform for edge computing scenarios. It consists of two independent deliverables: **monitoring platform service** (backend with embedded frontend via `go:embed`, deployed as a single binary) and **edge proxy**, supporting host management, health metric collection, threshold alerting, remote operations (terminal / file / process), and Agent dialogue (control agent / perception agent).

> Naming clarification: **Edge Proxy** refers to the data collection/control program deployed on monitored hosts (deliverable in `proxy/`). **System Agent dialogue** (control agent / perception agent) is the business module for interfacing with external Agents. The two have different responsibilities — **Proxy** and **Agent** are used strictly below.

## Tech Stack

### Backend (`backend/`, Go module `rtm-server`)

| Technology | Version | Description |
|-----------|---------|-------------|
| Go | 1.25+ | Programming language |
| Gin | v1.10 | Web framework |
| GORM | v1.25 | ORM framework |
| PostgreSQL | 15+ | Primary database |
| Redis | 7+ | Platform cache / edge messaging (dual instances) |
| JWT v5 | v5.3 | Authentication |
| Gorilla WS | v1.5 | WebSocket (terminal, file) |
| Zap | v1.27 | Logging framework |
| Viper | v1.19 | Configuration management |

### Frontend (`frontend/`, Vue3 SPA)

| Technology | Version | Description |
|-----------|---------|-------------|
| Vue | 3.3 | Frontend framework |
| Vite | 5.x | Build tool |
| Element Plus | 2.8 | UI component library |
| Pinia | 2.1 | State management |
| ECharts | 5.4 | Data visualization |
| Axios | 1.7 | HTTP client |
| xterm.js | 5.3 | Browser terminal |

### Edge Proxy (`proxy/`, Go module `rtm-proxy`, independent deliverable)

| Technology | Version | Description |
|-----------|---------|-------------|
| Go | 1.26+ | Programming language (**Linux only**) |
| gopsutil | v3 | Host metric collection |
| Gorilla WS | v1.5 | WebSocket long connection to platform |
| creack/pty | v1.1 | Remote terminal PTY implementation |

> Proxy and backend do not share code; they interact via HTTP/WS protocol. Can only run on Linux (amd64 / arm64); Windows is for cross-compilation only.

## Project Structure

```
rtm-monitor-platform/
├── backend/                        # Go backend service (module: rtm-server)
│   ├── cmd/server/main.go          # Main entry point
│   ├── internal/
│   │   ├── config/                 # Config loading (Viper + RTM_ env var overrides)
│   │   ├── constant/               # Status codes, Redis key prefixes, etc.
│   │   ├── dto/                    # Request/response DTOs
│   │   ├── handler/                # Gin handlers (including Agent proxy, chat sessions)
│   │   ├── middleware/             # JWT, XSS, CORS, Recovery, OperLog, Replay, WSAuth
│   │   ├── model/                  # GORM data models (alarms, chat sessions)
│   │   ├── repository/             # Data access layer
│   │   ├── router/                 # Centralized route registration (router.go)
│   │   ├── service/                # Business logic + built-in scheduler + alarm engine
│   │   ├── ws/                     # WebSocket Hub (frontend ↔ Proxy bridge)
│   │   └── utils/                  # Utility functions
│   ├── pkg/
│   │   ├── cache/                  # Redis wrapper
│   │   ├── captcha/                # Image/arithmetic captcha
│   │   ├── crypto/                 # RSA and other encryption tools
│   │   ├── jwt/                    # JWT utilities
│   │   ├── response/               # Unified response wrapper
│   │   └── static/                 # Frontend embedded static resources (go:embed all:dist)
│   └── configs/                    # Reference config templates
│       └── config.yaml
├── frontend/                       # Vue3 + Vite management console
│   ├── src/
│   │   ├── api/                    # Axios wrappers grouped by module
│   │   │   ├── business/           # Host, health, alarms
│   │   │   ├── control-agent/      # Control Agent dialogue
│   │   │   ├── perceive-agent/     # Perception Agent dialogue
│   │   │   ├── remote/             # Terminal, file
│   │   │   ├── system/             # Users
│   │   │   └── monitor/            # Operation logs, login logs
│   │   ├── assets/  components/  directive/  layout/
│   │   ├── router/                 # Frontend routes (static business routes)
│   │   ├── store/                  # Pinia state management
│   │   ├── utils/                  # Utilities (including jsencrypt)
│   │   └── views/                  # Pages (system / monitor / business)
│   ├── package.json
│   └── vite.config.js
├── proxy/                          # Edge Proxy (module: rtm-proxy, Linux only)
│   ├── main.go                     # Proxy entry (heartbeat, polling, WS client)
│   ├── client/                     # HTTP client (register, heartbeat, poll, ACK)
│   ├── collector/                  # System/process metric collection
│   ├── config/                     # Proxy config (JSON)
│   ├── executor/                   # Process start/stop executor + Shell PTY
│   ├── filemanager/                # Remote file management (path traversal protection)
│   ├── wsclient/                   # WebSocket client (auto-reconnect)
│   ├── build-proxy.ps1             # Windows cross-compile → dist/
│   └── README.md
├── sql/                            # PostgreSQL DDL
│   ├── 01_sys_tables.sql           # System tables (users, operation logs, login logs)
│   ├── 03_bu_tables.sql            # Business tables (host info, host health)
│   ├── alarm/01_alarm_tables.sql   # Alarm channels / rules / records / notification logs
│   └── chat/01_chat_tables.sql     # Agent dialogue sessions and messages
├── docs/                           # Chinese design and deployment documents
│   ├── Development Environment Setup Guide.md
│   ├── Production Deployment Guide.md
│   └── Proxy Deployment and Usage Guide.md
├── build-linux.bat                 # Windows host cross-compiles backend → Linux binary
└── logs/                           # Runtime logs
```

## Feature Modules

### System Management

- **User Management** — User CRUD, password policy, login failure lockout
- **Personal Center** — User info maintenance, password change, avatar upload

### Host Management

- **Host Information** — Edge node host registration, CRUD, and status display (auto-enrolled on first proxy registration)
- **Host Health** — Historical host health data collection and query (auto daily rotation, 7-day retention)
- **Host Operations** — Remote operations entry point (terminal, file, process control)

### Monitoring & Alerting

- **Alarm Channels** — Supports email, DingTalk, WeCom, and in-app notification channels, with multiple channels configurable
- **Alarm Rules** — Multi-dimensional rule orchestration:
  - Metric-based: CPU / memory / disk / network / load
  - Service-based: host offline, control Agent offline
- **Alarm Records** — Historical alarm query, confirmation, and recovery tracking
- **Alarm Notification Logs** — Detailed records of send status per channel
- **Built-in Scheduler** (no external cron dependency):
  - `dealOffline` (60s) — Host offline detection
  - `alarmCheck` (30s) — Alarm detection
  - `cleanHostHealth` (daily at 03:30) — Clean health data older than 7 days

### Agent Management (Business-side Agents)

- **Control Agent** — Interact with the control Agent through dialogue, executing process start/stop, parameter distribution, and other control operations
- **Perception Agent** — Interact with the perception Agent through dialogue, obtaining host online status and basic information
- **Agent Dialogue Sessions** — Session creation, paginated list, message history, rename, delete

### Remote Operations

- **Remote Terminal** — Browser-based xterm terminal, routed through platform WebSocket Hub directly to Proxy PTY (color and resize support)
- **Remote File** — File browse, read, edit, upload, download, delete, and directory creation on proxy hosts (10MB single file limit, path traversal protection)
- **Process Control** — Start / stop / restart process commands sent from platform to Proxy

### System Logs

- **Operation Logs** — Automatically records protected interface operations via `OperLog` middleware
- **Login Logs** — Login success / failure records

### Edge Proxy Features

- **System Info Collection** — Reports OS type/version, CPU arch/cores/model, memory, disk, MAC address on registration
- **System Metric Collection** — Reports CPU / memory / disk / network / load every 3 seconds
- **Process Monitoring** — Collects configured process running status, CPU usage, memory usage
- **Command Execution** — Receives platform-issued `startProcess` / `stopProcess` / `restartProcess` commands
- **WebSocket Long Connection** — Auto-reconnect (exponential backoff), heartbeat keepalive, safe goroutine shutdown
- **Remote Terminal** — PTY-based persistent shell sessions (5-minute timeout auto-close)
- **Remote File Management** — Complete file operations with path security validation

## Quick Start

### Requirements

- Go 1.25+ (Proxy build additionally requires Go 1.26+)
- Node.js 18+
- PostgreSQL 15+
- Redis 7+ (recommend preparing **two instances / two dbs**: platform and edge separated)

### 1. Clone Project

```bash
git clone <repository-url>
cd rtm-monitor-platform
```

### 2. Initialize Database

Create PostgreSQL database (default name `rtm`):

```sql
CREATE DATABASE rtm WITH ENCODING = 'UTF8';
```

Import DDL from `sql/` directory in order:

```bash
psql -d rtm -f sql/01_sys_tables.sql
psql -d rtm -f sql/03_bu_tables.sql
psql -d rtm -f sql/alarm/01_alarm_tables.sql
psql -d rtm -f sql/chat/01_chat_tables.sql
```

Create `config.yaml` under `~/.rtm-monitor-platform/` (reference `backend/configs/config.yaml`), and modify database connection config:

```yaml
datasource:
  host: 127.0.0.1
  port: 5432
  database: rtm
  username: postgres
  password: your_password
```

### 3. Start Backend (Development Mode)

```bash
cd backend
go mod download
go run cmd/server/main.go
```

Backend service runs at `http://localhost:18084` by default.

### 4. Start Frontend (Development Mode)

```bash
cd frontend
npm install
npm run dev
```

Frontend dev server runs at `http://localhost:80`, proxying `/dev-api` and `/dev-ws-api` to `127.0.0.1:18084`.

### 5. Access System

Open browser and visit `http://localhost`. Default credentials:

- Username: `admin`
- Password: `admin123`

### 6. (Optional) Deploy Proxy to Linux Host

See [`proxy/README.md`](proxy/README.md) and [`docs/Proxy Deployment and Usage Guide.md`](docs/Proxy Deployment and Usage Guide.md).

## Build & Deploy

### Linux Cross-Compilation (Backend + Embedded Frontend)

On Windows development machine, one-click cross-compile backend with frontend embedded:

```bash
build-linux.bat
```

Script execution flow:

1. Clean `backend/pkg/static/dist/` and previous build artifacts
2. Run `npm install && npm run build:prod` in `frontend/`
3. Copy `frontend/dist/*` to `backend/pkg/static/dist/`
4. Build amd64 and arm64 binaries with `GOOS=linux CGO_ENABLED=0`

Build artifacts output to `backend/` directory:

```
backend/
├── rtm-monitor-platform-linux-amd64   # AMD64 binary (frontend embedded)
├── rtm-monitor-platform-linux-arm64   # ARM64 binary (frontend embedded)
└── configs/
    └── config.yaml                    # Reference config template
```

The binary includes frontend static resources (`go:embed`), no additional Nginx deployment needed. On Linux server, first place config file at the fixed path, then start directly:

```bash
mkdir -p ~/.rtm-monitor-platform
cp backend/configs/config.yaml ~/.rtm-monitor-platform/config.yaml
# Edit ~/.rtm-monitor-platform/config.yaml for production settings

chmod +x rtm-monitor-platform-linux-amd64
./rtm-monitor-platform-linux-amd64
```

### Proxy Cross-Compilation

```powershell
cd proxy
.\build-proxy.ps1                  # Default output to proxy/dist/
.\build-proxy.ps1 -Version "1.1.0"
```

Artifacts:

- `proxy/dist/rtm-proxy-linux-amd64`
- `proxy/dist/rtm-proxy-linux-arm64`

> Proxy only supports Linux runtime; Windows / macOS are build hosts only.

### Frontend Standalone Build

```bash
cd frontend
npm run build:prod    # Production
npm run build:stage   # Staging
```

## Configuration

The backend uses a single `config.yaml` file, always loaded from `~/.rtm-monitor-platform/config.yaml`, shared between development and production.

Before first run, create the config file in the corresponding user directory (reference `backend/configs/config.yaml`):

```bash
# Linux / macOS
mkdir -p ~/.rtm-monitor-platform
cp backend/configs/config.yaml ~/.rtm-monitor-platform/config.yaml

# Windows
mkdir %USERPROFILE%\.rtm-monitor-platform
copy backend\configs\config.yaml %USERPROFILE%\.rtm-monitor-platform\config.yaml
```

Any key can be overridden via `RTM_` prefixed environment variables (`.` → `_`):

```bash
export RTM_DATASOURCE_HOST=192.168.1.100
export RTM_DATASOURCE_PASSWORD=secret
export RTM_REDIS_PLATFORM_PASSWORD=***
export RTM_REDIS_EDGE_HOST=192.168.1.101
```

### Key Configuration Items

```yaml
server:
  port: 18084          # HTTP / WebSocket port (frontend vite proxy fixed to this port)
  mode: release        # Run mode: debug/test/release

datasource:
  driver: postgres
  host: 127.0.0.1
  port: 5432
  database: rtm
  username: postgres
  password: ***

redis:
  platform:            # Platform Redis: sessions, captcha, login locks, online users
    host: 127.0.0.1
    port: 6379
    db: 3
  edge:                # Edge Redis: monitoring data / Proxy command queue
    host: 127.0.0.1
    port: 6379
    db: 1

jwt:
  header: Authorization
  secret: ***
  expireTime: 480      # Token expiration time (minutes)

agent:
  control: ""          # Control Agent external API address
  perceive: ""         # Perception Agent external API address
```

> Platform and edge Redis roles must be configured separately (can be different dbs on the same physical instance, or two separate instances).
> Agent configs are for interfacing with external Agent services; corresponding features are unavailable when empty.

### Development Hot Reload (Frontend-Backend Separation)

Set environment variable to let backend read frontend files from disk, skipping `go:embed`, avoiding recompiling backend for every frontend change:

```bash
# Windows
cd backend
set RTM_STATIC_DIR=..\frontend\dist
go run cmd\server\main.go
```

See `backend/pkg/static/static.go` for development / production switching logic.

## Backend Architecture

Standard layered structure. `cmd/server/main.go` handles dependency injection, `internal/router/router.go` is the **sole** route registration center — new handlers must be mounted there (no auto-discovery).

```
handler   →  service  →  repository  →  model (GORM)
   ↑           ↑
  dto      (business logic, may call multiple repos)
```

- Public route group `public`: login, get public key, proxy reporting (`/api/v1/proxy/*`), etc.
- Auth route group `auth`: all business APIs mounted after `middleware.JWTAuth()` validation.
- WebSocket: frontend terminal `/ws/terminal` (`WSAuth` validates token in query params); Proxy long connection `/api/v1/proxy/ws` (no JWT, security by network reachability).
- WebSocket Hub (`internal/ws/hub.go`) bridges frontend sessions and Proxy connections, handling remote terminal passthrough and remote file operations.
- Scheduler (`internal/service/scheduler.go`) started from `main.go`, carries three built-in jobs: offline detection, alarm detection, health data cleanup.
- Alarm Engine (`internal/service/alarm_engine.go`) checks alarm rules every 30 seconds, supporting threshold breach duration judgment and auto-recovery.

### Standard Response

All HTTP responses go through `pkg/response`, HTTP status fixed at 200, real business status in `code` field:

```go
response.Success(c, data)                 // 200 / "Operation successful"
response.SuccessWithMsg(c, msg, data)
response.SuccessPage(c, total, rows)      // {code, msg, data: {total, rows}}
response.Error(c, msg)                    // 500
response.Unauthorized(c, msg)             // 401 (used by JWT middleware)
response.Forbidden(c, msg)                // 403
```

Frontend interceptors judge business state by `code` (200 / 401 / 403 / 500 / 601).

### Auth Flow

1. `GET /getPublicKey` retrieves RSA public key
2. Frontend encrypts password with public key, then `POST /login`
3. Backend issues JWT, frontend carries it in `Authorization: Bearer <token>`
4. `JWTAuth` middleware parses token, reads `LoginUser` from platform Redis `login_tokens:<uuid>`, and injects `userId` / `username` / `uuid` / `claims` into Gin Context (accessed via `middleware.GetUserID(c)` etc.)
5. `GetInfo` auto-refreshes token when `loginUser.NeedRefresh()` is true

## Frontend Architecture

- `src/api/` grouped by domain: `system/`, `monitor/`, `business/`, `remote/`, `control-agent/`, `perceive-agent/`, plus `login.js`, `menu.js`.
- `src/views/` aligns with API grouping: `business/host-info/`, `business/alarm-config/`, `business/remote-terminal/`, `business/agent-control/`, `business/agent-perceive/`, etc.
- Login flow (`src/store/modules/user.js`): fetch public key → `utils/jsencrypt.js` encrypt password → `login()`, uniformly unpacks `{code, msg, data}` responses.
- Path aliases: `@` → `src/`, `~` → project root (see `vite.config.js`).
- Dev API base `/dev-api`, WebSocket base `/dev-ws-api`, both proxied to `127.0.0.1:18084`.

## Development Guide

### Backend Development Conventions

- **Handler** — Gin request entry, responsible for parameter validation and response wrapping; each handler type must provide `RegisterRoutes(*gin.RouterGroup)`, explicitly called from `router.go`.
- **Service** — Business logic layer, constructor receives `*gorm.DB` and (when necessary) other services.
- **Repository** — Data access layer, wraps GORM queries; not directly exposed externally.
- **DTO / Model** — `internal/dto/*` for external interaction, `internal/model/*` are GORM models. **Do not return models directly to frontend**.
- **Constants Reuse** — Status codes, Redis key prefixes, `del_flag`, captcha types, etc. centralized in `internal/constant/constant.go`. Hardcoded literals are prohibited.
- **Audit Logging** — CRUD handlers requiring operation logs should receive `operLogService` in constructor, recorded uniformly by `OperLog` middleware.

### Frontend Development Conventions

- API interfaces go in `src/api/`, grouped by module.
- Page components go in `src/views/` under corresponding modules.
- Common components go in `src/components/`.
- State management uses Pinia, modules defined in `src/store/modules/`.

## API Documentation

The project follows RESTful design style with unified response format:

```json
{
  "code": 200,
  "msg": "Operation successful",
  "data": {}
}
```

Authentication: request header `Authorization: Bearer <token>`; WebSocket passes token via query parameter `?token=<...>`.

### Proxy Protocol Interfaces (Public, No Authentication Required)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/proxy/register` | Proxy first-time registration, reports hardware info |
| POST | `/api/v1/proxy/heartbeat` | Heartbeat keepalive + monitoring data report (3s cycle) |
| POST | `/api/v1/proxy/poll` | Poll pending commands (process start/stop) |
| POST | `/api/v1/proxy/ack` | Report command execution result |
| GET | `/api/v1/proxy/ws?ip=<host>` | WebSocket long connection (terminal/file) |

## Security Features

- JWT Token authentication + Token auto-refresh
- RSA password encryption in transit
- XSS filtering middleware (enabled by `xss.urlPatterns` / `xss.excludes`)
- Replay attack protection (`Replay` middleware)
- Login failure lockout (`user.password.maxRetryCount` / `lockTime`)
- SQL injection protection (GORM parameterized queries)
- CORS cross-origin control
- Remote file path security validation (directory traversal prevention)

## Logging

Logs are output via Zap, default writing to `logs/` directory, with size / day-based rotation (implemented by lumberjack):

```
logs/
├── rtm-server.log          # Current log
└── rtm-server-*.log        # Historical rotated logs
```

Log level, filename, single file size, retention count, retention days, and compression are all configurable in `config.yaml` `log` block:

```yaml
log:
  level: info           # debug / info / warn / error
  filename: logs/rtm-server.log
  maxSize: 100          # Single file MB
  maxBackups: 30        # Max retention copies
  maxAge: 30            # Max retention days
  compress: true
```

## Documentation

More design and deployment details in `docs/`:

- Development Environment Setup Guide
- Production Deployment Guide
- Edge Proxy Deployment and Usage Guide
- Remote Host Management Feature Design
- Control Agent Documentation (System-side Agent Control Features)

## License

[MIT License](../LICENSE)
