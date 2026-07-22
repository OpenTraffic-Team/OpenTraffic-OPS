# OpenTraffic Ops Init Deployment Panel

<p align="center">
  <a href="../LICENSE"><img src="https://img.shields.io/badge/License-Apache_2.0-blue.svg" alt="License: Apache 2.0"></a>
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white" alt="Go 1.21+">
  <img src="https://img.shields.io/badge/Vue-3.3-4FC08D?logo=vue.js&logoColor=white" alt="Vue 3">
  <img src="https://img.shields.io/badge/Docker-Supported-2496ED?logo=docker&logoColor=white" alt="Docker">
</p>

<p align="center">
  <img src="https://img.shields.io/badge/HuggingFace-%F0%9F%A4%97-FFD21E" alt="HuggingFace">
  <img src="https://img.shields.io/badge/X-000000?logo=x&logoColor=white" alt="X">
  <img src="https://img.shields.io/badge/%E5%B0%8F%E7%BA%A2%E4%B9%A6-FF2442" alt="小红书">
</p>

<p align="center">
  <a href="README_CN.md">中文</a>
</p>

## 📑 Table of Contents

- [📖 Project Introduction](#project-introduction)
  - [Core Capabilities](#core-capabilities)
  - [Single-Binary Self-Contained Deployment](#single-binary-self-contained-deployment)
- [🔧 Tech Stack](#tech-stack)
  - [Backend](#backend)
  - [Frontend](#frontend)
- [✨ Features](#features)
  - [📊 Dashboard](#dashboard)
  - [🧩 Component Management](#component-management)
  - [🖧 Server Management](#server-management)
  - [📦 Remote Deployment](#remote-deployment)
  - [⚙️ Configuration Management](#configuration-management)
  - [📖 User Guide](#user-guide)
- [🚀 Quick Start](#quick-start)
  - [📋 Prerequisites](#prerequisites)
  - [💻 Development Mode](#development-mode)
- [🖥️ Server Deployment](#server-deployment)
  - [📦 Production Build](#production-build-single-binary)
  - [⚙️ Configuration](#configuration)
  - [🔒 Security Recommendations](#security-recommendations)
  - [💾 Backup Strategy](#backup-strategy)
- [❓ FAQ](#faq)
- [🙏 Acknowledgments](#acknowledgments)

---

## 📖 Project Introduction

OpenTraffic Ops Init Deployment Panel is a **single-binary, self-contained** comprehensive operations platform that integrates two core capabilities: **Docker container management** and **SSH remote server deployment**. The backend is powered by a single Go HTTP service, and the frontend is embedded into the binary via `go:embed` — **no additional Nginx or reverse proxy configuration required**. Simply run one binary to launch the complete service.

### 🎯 Core Capabilities

| Capability | Description |
|-----------|-------------|
| Docker Component Management | One-click install, start, stop, and uninstall common middleware (PostgreSQL, Redis) with custom ports, environment variables, volumes, and startup commands |
| Real-time Monitoring | View real-time resource usage (CPU / memory / network / disk) of components, with live log refresh and auto-refresh support |
| SSH Server Management | Centrally manage SSH connection configurations for multiple remote Linux servers, supporting both password and key authentication |
| Remote Binary Deployment | Deploy `opentraffic-ops-proxy` and `opentraffic-ops` binaries to remote servers via SSH/SFTP with one click |
| Remote Configuration Management | View and edit software configuration files on remote servers online (`opentraffic-ops-proxy-config.json` for proxy, `opentraffic-ops-config.yaml` for platform) |
| Remote Service Management | Start, stop, and restart services on remote servers via PID files, without requiring root privileges |
| Deployment Audit Trail | Complete logging of every remote deployment operation, execution results, and historical records |

### 📦 Single-Binary Self-Contained Deployment

A core design goal is **eliminating dependency on external web servers** (like Nginx). The frontend `dist` directory is embedded directly into the Go binary via `go:embed`, with a custom SPA fallback handler so routes like `/components` work on direct access or refresh. The final artifact is always a **single self-contained binary file** — just copy and run.

---

## 🔧 Tech Stack

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

## ✨ Features

### 📊 Dashboard
- Component stat cards (total / running / stopped / error)
- Server stat cards (total / password auth / key auth / configured for deployment)
- Component type distribution pie chart
- Component status distribution bar chart
- Real-time component monitoring table (CPU / memory / network IO) with start/stop live refresh

![Dashboard](images/image.png)

### 🧩 Component Management
- Browse component catalog with Docker connection status
- One-click install components (PostgreSQL, Redis)
- Customize during installation: component name, port, environment variables, volumes, startup command arguments
- Start / stop / restart / uninstall installed components
- View component details (resource monitoring, logs, configuration)
- Built-in offline images — no internet required for deployment

![Component Management](images/image-2.png)

#### Supported Component Types

| Component | Type | Default Image | Description |
|-----------|------|--------------|-------------|
| PostgreSQL | `postgresql` | `postgres:16-alpine` | Relational database |
| Redis | `redis` | `redis:7-alpine` | In-memory cache / key-value store |

### 🖧 Server Management
- Add / edit / delete remote server SSH configurations
- Two authentication methods supported: password auth and key auth (supports Passphrase)
- SSH connection test
- Server list displays service status (proxy / monitor / control / perception)
- Expandable rows to view deployed service details
- Supported operations: start / stop / restart / configure / uninstall remote services; control and perception packages support start / stop / restart / configure / uninstall, with config paths `{deploy_path}/opentraffic-control/config/mq_config.json` and `{deploy_path}/opentraffic-perception/opentraffic-perception-linux-{amd64,arm64}/drivers/config.json` respectively

![Server Management](images/image-1.png)

### 📦 Remote Deployment
- Select target servers to deploy built-in binaries (`opentraffic-ops-proxy`, `opentraffic-ops`)
- Deploy the `opentraffic-control` algorithm package (tar archive) to remote servers with architecture auto-detection (amd64 / arm64 / loong64) and version tracking
- **LoongArch64 (龙芯)**: uses a two-package model — Python environment (`trafficlight-loong64.tar.gz`, extracted as `trafficlight_env/` with all dependencies bundled) is deployed to `{deploy_path}/opentraffic-control/trafficlight_env` on first deploy, while the algorithm package (`opentraffic-control-linux-loong64.tar`, pre-compiled .so) is updated incrementally; runs directly after extraction, no on-board compilation required
- **ARM aarch64**: uses a two-package model — Python environment (`trafficlight-arm64.tar.gz`, extracted as `trafficlight_env/` with all dependencies bundled) is deployed to `{deploy_path}/opentraffic-control/trafficlight_env` on first deploy, while the algorithm package (`opentraffic-control-linux-arm64.tar`) is updated incrementally; runs directly after extraction, no conda / pip / build tools required
- **x86/amd64**: uses a two-package model — Python environment (`trafficlight-amd64.tar.gz`, extracted as `trafficlight_env/` with all dependencies bundled) is deployed to `{deploy_path}/opentraffic-control/trafficlight_env` on first deploy, while the algorithm package (`opentraffic-control-linux-amd64.tar`) is updated incrementally; runs directly after extraction, no conda / pip / build tools required
- **x86/amd64 and ARM aarch64 perception**: deploy the `opentraffic-perception-linux-{amd64,arm64}.tar` runtime package to remote servers with architecture auto-detection. On first deploy it automatically runs `deploy/install.sh` to create `.venv`, runs `deploy/configure.sh` to generate a default `drivers/config.json`, and writes the user-supplied configuration. Subsequent redeployments update only the algorithm package. The ARM64 build uses RKNN inference
- Optionally deploy configuration files simultaneously for binaries
- Support loading default configuration templates
- Duplicate deployment detection for binaries; algorithm packages allow repeated deployments with version history
- Complete deployment records and log traceability

![Remote Deployment](images/image-3.png)

### ⚙️ Configuration Management
- View configuration list of all installed components
- Edit component configurations online (ports, environment variables, volumes, startup commands)
- Manual restart required after configuration changes take effect

![Configuration Management](images/image-4.png)

### 📖 User Guide
- Platform introduction and feature overview
- Basic environment requirements (Docker, browser, network, SSH)
- Component and server management usage instructions
- Remote deployment process instructions
- PostgreSQL / Redis default configuration and parameter descriptions
- Common FAQ (accordion-style interaction)

![User Guide](images/image-5.png)

---

## 🚀 Quick Start

### 📋 Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose (for local runtime)
- Git

### 💻 Development Mode

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

Backend service starts at `http://localhost:18080`.

#### 3. Start Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend dev server starts at `http://localhost:5173`, with Vite Proxy automatically forwarding `/api` to `http://localhost:18080`.

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

## 🖥️ Server Deployment

### 📦 Production Build (Single Binary)

#### Windows Cross-Compile for Linux

Run `build-opentraffic-ops-initialization.bat` to generate Linux AMD64, ARM64 and Loong64 binaries:

```cmd
build-opentraffic-ops-initialization.bat
```

Output files:
- `backend\opentraffic-ops-init-linux-amd64`
- `backend\opentraffic-ops-init-linux-arm64`
- `backend\opentraffic-ops-init-linux-loong64`

Upload to Linux server and run:

```bash
chmod +x opentraffic-ops-init-linux-amd64
./opentraffic-ops-init-linux-amd64
```

### ⚙️ Configuration

Create `backend/.env` file:

```env
# Server config
SERVER_HOST=0.0.0.0
SERVER_PORT=18080

# Database config
DATA_DIR=./data

# JWT config
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRE_HOURS=24

# Encryption key (32 bytes, for encrypting SSH passwords and private keys)
ENCRYPTION_KEY=your-encryption-key-32-bytes-long
```

### 🔒 Security Recommendations

1. Change the default admin password
2. Replace JWT Secret and encryption key
3. Use HTTPS (can terminate TLS via external reverse proxy or load balancer)
4. Configure firewall rules
5. Regularly back up the SQLite database

### 💾 Backup Strategy

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

## ❓ FAQ

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

### LoongArch64 control service fails to start
- Confirm `{deploy_path}/opentraffic-control/trafficlight_env/bin/python3` exists after first deployment
- Run `file trafficlight_env/bin/python3` and confirm it shows a LoongArch ELF binary; a mismatch means the wrong env package was used
- Verify Redis address, port and password in `config/mq_config.json`
- Check `{deploy_path}/opentraffic-control/run.log` for detailed errors

### ARM aarch64 control service fails to start
- Confirm `{deploy_path}/opentraffic-control/trafficlight_env/bin/python3` exists after first deployment
- Run `file trafficlight_env/bin/python3` and confirm it shows `ELF 64-bit LSB ... ARM aarch64`; a mismatch means the wrong env package was used
- Verify Redis address, port and password in `config/mq_config.json`
- Check `{deploy_path}/opentraffic-control/run.log` for detailed errors

### x86/amd64 control service fails to start
- Confirm `{deploy_path}/opentraffic-control/trafficlight_env/bin/python3` exists after first deployment
- Run `file trafficlight_env/bin/python3` and confirm it shows `ELF 64-bit LSB ... x86-64`; a mismatch means the wrong env package was used
- Verify Redis address, port and password in `config/mq_config.json`
- Check `{deploy_path}/opentraffic-control/run.log` for detailed errors

### x86/amd64 perception service fails to start
- Confirm `{deploy_path}/opentraffic-perception/opentraffic-perception-linux-amd64/.venv/bin/python3` exists after first deployment
- Confirm the remote server has `python3.13` installed before deploying
- Verify `video_path`, `radar_reference_path`, `output_path` and Redis settings in `drivers/config.json`
- Check `{deploy_path}/opentraffic-perception/opentraffic-perception-linux-amd64/run.log` for detailed errors

### ARM aarch64 perception service fails to start
- Confirm `{deploy_path}/opentraffic-perception/opentraffic-perception-linux-arm64/.venv/bin/python3` exists after first deployment
- Confirm the remote server has `python3.13` installed before deploying
- Verify `video_path`, `radar_reference_path`, `output_path` and Redis settings in `drivers/config.json`
- Check `{deploy_path}/opentraffic-perception/opentraffic-perception-linux-arm64/run.log` for detailed errors

### Component container start failed (Permission denied)
- When using bind mounts, ensure host directory owner matches container default user UID
- PostgreSQL UID is 70, Redis UID is 999
- Recommend using named volumes (e.g., `postgres-data:/var/lib/postgresql/data`), Docker handles permissions automatically

---

## 🙏 Acknowledgments

OpenTraffic Ops Init is built with the following open-source projects:

- [Go](https://golang.org/) / [Gin](https://github.com/gin-gonic/gin) — Backend framework
- [Vue.js](https://vuejs.org/) / [Vite](https://vitejs.dev/) — Frontend framework and build tool
- [Element Plus](https://element-plus.org/) — UI component library
- [Docker SDK for Go](https://github.com/docker/docker-ce) — Container management
- [ECharts](https://echarts.apache.org/) — Data visualization

[Apache License 2.0](../LICENSE)
