# OpenTraffic Ops Init Deployment Panel

[中文](README_CN.md)

## Table of Contents

- [Project Introduction](#project-introduction)
- [Tech Stack](#tech-stack)
- [Features](#features)
- [Quick Start](#quick-start)
- [Server Deployment](#server-deployment)
- [FAQ](#faq)
- [Acknowledgments](#acknowledgments)

---

## Project Introduction

OpenTraffic Ops Init Deployment Panel is a **single-binary, self-contained** comprehensive operations platform that integrates two core capabilities: **Docker container management** and **SSH remote server deployment**. The backend is powered by a single Go HTTP service, and the frontend is embedded into the binary via `go:embed` — **no additional Nginx or reverse proxy configuration required**. Simply run one binary to launch the complete service.

### Core Capabilities

| Capability | Description |
|-----------|-------------|
| Docker Component Management | One-click install, start, stop, and uninstall common middleware (PostgreSQL, Redis) with custom ports, environment variables, volumes, and startup commands |
| Real-time Monitoring | View real-time resource usage (CPU / memory / network / disk) of components, with live log refresh and auto-refresh support |
| SSH Server Management | Centrally manage SSH connection configurations for multiple remote Linux servers, supporting both password and key authentication |
| Remote Binary Deployment | Deploy `opentraffic-ops-proxy` and `opentraffic-ops` binaries to remote servers via SSH/SFTP with one click |
| Remote Configuration Management | View and edit software configuration files on remote servers online (`config.json` for proxy, `config.yaml` for platform) |
| Remote Service Management | Start, stop, and restart services on remote servers via PID files, without requiring root privileges |
| Deployment Audit Trail | Complete logging of every remote deployment operation, execution results, and historical records |

### Single-Binary Self-Contained Deployment

A core design goal is **eliminating dependency on external web servers** (like Nginx). The frontend `dist` directory is embedded directly into the Go binary via `go:embed`, with a custom SPA fallback handler so routes like `/components` work on direct access or refresh. The final artifact is always a **single self-contained binary file** — just copy and run.

---

## Tech Stack

### Backend

| Technology | Description |
|-----------|-------------|
| Go 1.21+ | Programming language |
| Gin | Web framework |
| SQLite | Zero-dependency embedded database |
| Docker SDK for Go | Container lifecycle management |
| `golang.org/x/crypto/ssh` | SSH client implementation |
| JWT | Authentication |
| `go:embed` + custom SPA fallback | Static file hosting without Nginx |
| AES-GCM | Encryption for sensitive data (SSH passwords, private keys) |

### Frontend

| Technology | Description |
|-----------|-------------|
| Vue 3 + TypeScript | Frontend framework |
| Element Plus | UI component library |
| Vite | Build tool |
| Pinia | State management |
| ECharts | Data visualization |
| `createWebHistory` | History-mode routing |

---

## Features

### Dashboard
- Component stat cards (total / running / stopped / error)
- Server stat cards (total / password auth / key auth / configured for deployment)
- Component type distribution pie chart
- Component status distribution bar chart
- Real-time component monitoring table (CPU / memory / network IO) with start/stop live refresh

> 🖼️ **Screenshot placeholder**: Add a screenshot of the Dashboard page here.

### Component Management
- Browse component catalog with Docker connection status
- One-click install components (PostgreSQL, Redis)
- Customize during installation: component name, port, environment variables, volumes, startup command arguments
- Start / stop / restart / uninstall installed components
- View component details (resource monitoring, logs, configuration)
- Built-in offline images — no internet required for deployment

> 🖼️ **Screenshot placeholder**: Add a screenshot of the Component Management page here.

#### Supported Component Types

| Component | Type | Default Image | Description |
|-----------|------|--------------|-------------|
| PostgreSQL | `postgresql` | `postgres:16-alpine` | Relational database |
| Redis | `redis` | `redis:7-alpine` | In-memory cache / key-value store |

### Server Management
- Add / edit / delete remote server SSH configurations
- Two authentication methods supported: password auth and key auth (supports Passphrase)
- SSH connection test
- Server list displays service status (proxy / monitor)
- Expandable rows to view deployed service details
- Supported operations: start / stop / restart / configure / uninstall remote services

> 🖼️ **Screenshot placeholder**: Add a screenshot of the Server Management page here.

### Remote Deployment
- Select target servers to deploy built-in binaries (`opentraffic-ops-proxy`, `opentraffic-ops`)
- Optionally deploy configuration files simultaneously
- Support loading default configuration templates
- Duplicate deployment detection
- Complete deployment records and log traceability

> 🖼️ **Screenshot placeholder**: Add a screenshot of the Remote Deployment page here.

### Configuration Management
- View configuration list of all installed components
- Edit component configurations online (ports, environment variables, volumes, startup commands)
- Manual restart required after configuration changes take effect

> 🖼️ **Screenshot placeholder**: Add a screenshot of the Configuration Management page here.

### User Guide
- Platform introduction and feature overview
- Basic environment requirements (Docker, browser, network, SSH)
- Component and server management usage instructions
- Remote deployment process instructions
- PostgreSQL / Redis default configuration and parameter descriptions
- Common FAQ (accordion-style interaction)

> 🖼️ **Screenshot placeholder**: Add a screenshot of the User Guide page here.

---

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose (for local runtime)
- Git

### Development Mode

#### 1. Clone Project

```bash
git clone <repository-url>
cd OpenTraffic-Ops-Initialization
```

#### 2. Start Backend

```bash
cd backend
go mod download
go run cmd/server/main.go
```

Backend service starts at `http://localhost:8080`.

#### 3. Start Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend dev server starts at `http://localhost:5173`, with Vite Proxy automatically forwarding `/api` to `http://localhost:8080`.

#### 4. Access System

Open browser and visit `http://localhost:5173`

Default login credentials:
- Username: `admin`
- Password: `admin123`

#### Windows Local Development Quick Debug (No dist copy needed)

During development, frontend changes are frequent. An environment variable switch in `backend/pkg/static/static.go` allows loading frontend assets directly from disk:

```cmd
# In backend directory
set RTM_STATIC_DIR=..\frontend\dist
go run cmd\server\main.go
```

**Do not** set this variable for production builds, to ensure frontend resources are fully embedded in the binary.

---

## Server Deployment

### Production Build (Single Binary)

#### Windows Cross-Compile for Linux

Run `build-opentraffic-ops-initialization.bat` to generate Linux AMD64 and ARM64 binaries:

```cmd
build-opentraffic-ops-initialization.bat
```

Output files:
- `backend\opentraffic-ops-init-linux-amd64`
- `backend\opentraffic-ops-init-linux-arm64`

Upload to Linux server and run:

```bash
chmod +x opentraffic-ops-init-linux-amd64
./opentraffic-ops-init-linux-amd64
```

### Configuration

Create `backend/.env` file:

```env
# Server config
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database config
DATA_DIR=./data

# JWT config
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRE_HOURS=24

# Encryption key (32 bytes, for encrypting SSH passwords and private keys)
ENCRYPTION_KEY=your-encryption-key-32-bytes-long
```

### Security Recommendations

1. Change the default admin password
2. Replace JWT Secret and encryption key
3. Use HTTPS (can terminate TLS via external reverse proxy or load balancer)
4. Configure firewall rules
5. Regularly back up the SQLite database

### Backup Strategy

```bash
# Backup SQLite database
cp backend/data/opentraffic-ops-init.db backup/opentraffic-ops-init_$(date +%Y%m%d).db

# Backup Docker volumes
docker run --rm \
  -v opentraffic-ops-init-data:/data \
  -v $(pwd)/backup:/backup \
  alpine tar czf /backup/opentraffic-ops-init-data_$(date +%Y%m%d).tar.gz /data
```

---

## FAQ

### Docker connection failed
- Ensure Docker service is running
- Check `/var/run/docker.sock` permissions

### Port conflict
- Modify port configuration in `.env` file
- Ensure port is not occupied

### Frontend refresh returns 404
- Confirm `frontend/dist` exists during `go build`
- Check embed path in `backend/pkg/static/static.go`

### CORS issues
- Production environment is same-origin, should not have CORS
- Development environment: ensure Vite Proxy config is correct and backend is running

### SSH connection test failed
- Check if target server SSH service is running
- Confirm host address, port, username, password/private key are correct
- Check if firewall allows SSH port

### Remote deployment failed (insufficient permissions)
- Check if SSH user has read/write permissions for deployment path
- Confirm deployment path disk has sufficient space
- Check target server's SELinux or AppArmor restrictions

### Component container start failed (Permission denied)
- When using bind mounts, ensure host directory owner matches container default user UID
- PostgreSQL UID is 70, Redis UID is 999
- Recommend using named volumes (e.g., `postgres-data:/var/lib/postgresql/data`), Docker handles permissions automatically

---

## Acknowledgments

OpenTraffic Ops Init is built with the following open-source projects:

- [Go](https://golang.org/) / [Gin](https://github.com/gin-gonic/gin) — Backend framework
- [Vue.js](https://vuejs.org/) / [Vite](https://vitejs.dev/) — Frontend framework and build tool
- [Element Plus](https://element-plus.org/) — UI component library
- [Docker SDK for Go](https://github.com/docker/docker-ce) — Container management
- [ECharts](https://echarts.apache.org/) — Data visualization

[MIT License](../LICENSE)
