# OpenTraffic Ops

A full-stack edge computing operations platform composed of two integrated subsystems: a **deployment panel** for infrastructure provisioning and a **monitoring platform** for edge host management, alerting, and remote operations.

---

## Architecture Overview

```
                    ┌─────────────────────────────────┐
                    │      OpenTraffic Ops            │
                    │  ┌─────────────────────────┐    │
                    │  │  rtm-initialization     │    │
                    │  │  (Deployment Panel)     │    │
                    │  │  - Docker management    │    │
                    │  │  - SSH remote deploy    │────┼──► Deploys rtm-monitor-platform
                    │  │  - Component lifecycle  │    │    and rtm-proxy binaries
                    │  └─────────────────────────┘    │    to remote Linux servers
                    │  ┌─────────────────────────┐    │
                    │  │  rtm-monitor-platform   │    │
                    │  │  (Monitoring & Ops)     │    │
                    │  │  - Host monitoring      │◄───┼──── Receives metrics from
                    │  │  - Alerting engine      │    │    edge proxies
                    │  │  - Remote terminal      │    │
                    │  │  - Agent dialogue       │    │
                    │  └─────────────────────────┘    │
                    └─────────────────────────────────┘
                                         ▲
                                         │ WebSocket / HTTP
                                         │
                    ┌────────────────────┴─────────────┐
                    │      rtm-proxy (Edge Agent)      │
                    │  - System metrics collection     │
                    │  - Process monitoring            │
                    │  - Remote terminal PTY           │
                    │  - Remote file operations        │
                    └──────────────────────────────────┘
                    Deployed on each monitored edge host
```

---

## Subsystems

### 1. rtm-initialization — Deployment Panel

A single-binary, self-contained deployment dashboard that requires no external web server (Nginx) or database (PostgreSQL).

| Capability | Description |
|-----------|-------------|
| Docker Management | One-click install/start/stop/uninstall of middleware (PostgreSQL, Redis) with custom ports, env vars, volumes |
| Real-time Monitoring | Live resource stats (CPU / memory / network / disk) for containers |
| SSH Server Management | Centralized SSH connection configs for multiple remote Linux servers (password or key auth) |
| Remote Binary Deploy | Deploy `rtm-proxy` and `rtm-monitor-platform` binaries to remote servers via SSH/SFTP |
| Remote Config Edit | View and edit remote configuration files (`config.json`, `config.yaml`) online |
| Remote Service Control | Start / stop / restart services on remote hosts via PID files |
| Deployment Audit | Full operation logs, execution results, and deployment history |

**Tech Stack:** Go 1.21+ (Gin, SQLite, Docker SDK, `crypto/ssh`), Vue 3 + TypeScript + Vite, Element Plus

**Key Design:** Frontend is embedded into the Go binary via `go:embed`. The backend serves both API and SPA static files on a single port with custom SPA fallback logic — zero Nginx dependency.

[Details &rarr;](./rtm-initialization/README.md)

---

### 2. rtm-monitor-platform — Monitoring & Operations Platform

A full-stack monitoring and operations platform for edge computing scenarios. Consists of three deliverables: **backend service**, **frontend console**, and **edge proxy**.

| Capability | Description |
|-----------|-------------|
| Host Management | Edge node registration, CRUD, and status display (auto-enrolled on first proxy registration) |
| Health Metrics | Historical host health data with automatic daily rotation (7-day retention) |
| Alerting Engine | Multi-channel notifications (Email, DingTalk, WeCom, In-App), threshold-based rules for CPU / memory / disk / network / load |
| Remote Terminal | Browser-based xterm terminal through WebSocket hub to proxy PTY (colors, resize support) |
| Remote File Ops | Browse, read, edit, upload, download, delete files on proxy hosts (10MB limit, path traversal protection) |
| Process Control | Start / stop / restart processes on edge hosts via platform commands |
| Agent Dialogue | Conversational interaction with control and perception agents for process control and host status queries |

**Tech Stack:**
- Backend: Go 1.25+ (Gin, GORM, PostgreSQL, Redis, JWT v5, Gorilla WebSocket, Zap, Viper)
- Frontend: Vue 3 + Vite, Element Plus, Pinia, ECharts, xterm.js
- Edge Proxy: Go 1.26+ (Linux only, amd64/arm64), gopsutil, Gorilla WebSocket, creack/pty

**Key Design:** The backend serves the SPA via `go:embed` as a single binary. The edge proxy (`proxy/`) is a separate Go module that communicates with the platform via HTTP/WebSocket — deployed independently on each monitored host.

[Details &rarr;](./rtm-monitor-platform/README.md)

---

## Relationship Between Subsystems

```
┌─────────────────────┐      deploys      ┌──────────────────────────┐
│ rtm-initialization  │ ─────────────────►│ rtm-monitor-platform     │
│ (this machine)      │  SSH/SFTP         │ (remote Linux server)    │
│                     │                   │                          │
│ - Docker mgmt       │      deploys      │ - Host monitoring        │
│ - SSH configs       │ ─────────────────►│ - Alerting               │
│ - Binary deploy     │                   │ - Remote ops             │
└─────────────────────┘                   └────────────┬─────────────┘
                                                       │
                                                       │ HTTP / WebSocket
                                                       │
                                              ┌────────▼──────────────┐
                                              │ rtm-proxy             │
                                              │ (on each edge host)   │
                                              │ - Metrics collection  │
                                              │ - Remote terminal     │
                                              │ - File operations     │
                                              └───────────────────────┘
```

1. **`rtm-initialization`** is your control plane — run it on your local machine or a bastion host. It manages Docker containers (PostgreSQL, Redis) and deploys the monitoring stack to remote servers.

2. **`rtm-monitor-platform`** runs as a server on a central or edge node. It collects metrics, triggers alerts, and provides the Web UI for operators.

3. **`rtm-proxy`** runs on each host you want to monitor. It reports metrics every 3 seconds and accepts remote commands (terminal, file, process) from the platform.

---

## Quick Start

### Prerequisites

- Go 1.25+ (proxy build requires Go 1.26+)
- Node.js 18+
- Docker & Docker Compose (for `rtm-initialization` container management)
- PostgreSQL 15+ (for `rtm-monitor-platform`)
- Redis 7+ (two instances recommended: platform + edge)

### Start the Deployment Panel

```bash
cd rtm-initialization/backend
go mod download
go run cmd/server/main.go
# Service runs on http://localhost:8080
```

### Start the Monitoring Platform

```bash
# 1. Create PostgreSQL database
psql -c "CREATE DATABASE rtm WITH ENCODING = 'UTF8';"

# 2. Import DDL
cd rtm-monitor-platform
psql -d rtm -f sql/01_sys_tables.sql
psql -d rtm -f sql/03_bu_tables.sql
psql -d rtm -f sql/alarm/01_alarm_tables.sql
psql -d rtm -f sql/chat/01_chat_tables.sql

# 3. Start backend
cd backend
go mod download
go run cmd/server/main.go
# Service runs on http://localhost:18084

# 4. Start frontend (dev mode)
cd ../frontend
npm install
npm run dev
# Dev server on http://localhost:80
```

Default credentials for both systems: `admin` / `admin123`

### Build Production Binaries (Windows host cross-compiling to Linux)

```bash
# Monitoring platform (backend + embedded frontend)
cd rtm-monitor-platform
build-linux.bat
# Outputs: backend/rtm-monitor-platform-linux-amd64, backend/rtm-monitor-platform-linux-arm64

# Edge proxy
cd proxy
.\build-proxy.ps1
# Outputs: proxy/dist/rtm-proxy-linux-amd64, proxy/dist/rtm-proxy-linux-arm64

# Deployment panel
cd ../../rtm-initialization
build-linux.bat
# Outputs: backend/rtm-initialization-linux-amd64, backend/rtm-initialization-linux-arm64
```

---

## Project Structure

```
opentraffic-ops/
├── rtm-initialization/          # Deployment Panel
│   ├── backend/                 # Go backend (Gin, SQLite, Docker SDK)
│   ├── frontend/                # Vue 3 + TypeScript SPA
│   ├── components/              # Docker Compose templates
│   ├── docker-compose.yaml
│   └── README.md                # (Chinese, detailed)
│
├── rtm-monitor-platform/        # Monitoring & Operations Platform
│   ├── backend/                 # Go backend (Gin, GORM, PostgreSQL, Redis)
│   ├── frontend/                # Vue 3 SPA
│   ├── proxy/                   # Edge proxy (Linux only, separate Go module)
│   ├── sql/                     # PostgreSQL DDL
│   ├── docs/                    # Design & deployment guides (Chinese)
│   └── README.md                # (Chinese, detailed)
│
├── README.md                    # This file
├── .gitignore                   # Root-level combined ignore rules
└── LICENSE                      # MIT License
```

---

## Documentation

- [rtm-initialization README](./rtm-initialization/README.md) — Deployment panel details (Chinese)
- [rtm-monitor-platform README](./rtm-monitor-platform/README.md) — Monitoring platform details (Chinese)
- [rtm-monitor-platform CLAUDE.md](./rtm-monitor-platform/CLAUDE.md) — Developer reference for Claude Code
- [Proxy README](./rtm-monitor-platform/proxy/README.md) — Edge proxy deployment guide

---

## Security Features

- JWT token authentication with auto-refresh
- RSA password encryption in transit
- XSS filtering middleware
- Replay attack protection
- Login failure lockout
- Parameterized SQL queries (GORM)
- CORS control
- Remote file path traversal protection
- AES-GCM encryption for SSH credentials (in deployment panel)

---

## License

[MIT License](./LICENSE)
