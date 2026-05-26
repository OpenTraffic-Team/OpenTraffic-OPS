# OpenTraffic Ops Init Deployment Panel

## Overview

RTM Deployment Panel is a **single-binary, self-contained** comprehensive operations platform that integrates two core capabilities: **Docker container management** and **SSH remote server deployment**. The backend is powered by a single Go HTTP service, and the frontend is embedded into the binary via `go:embed` — **no additional Nginx or reverse proxy configuration required**. Simply run one binary to launch the complete service.

### Core Capabilities

| Capability | Description |
|-----------|-------------|
| Docker Component Management | One-click install, start, stop, and uninstall common middleware (PostgreSQL, Redis) with custom ports, environment variables, volumes, and startup commands |
| Real-time Monitoring | View real-time resource usage (CPU / memory / network / disk) of components, with live log refresh and auto-refresh support |
| SSH Server Management | Centrally manage SSH connection configurations for multiple remote Linux servers, supporting both password and key authentication |
| Remote Binary Deployment | Deploy `opentraffic-ops-proxy` and `opentraffic-ops` binaries to remote servers via SSH/SFTP with one click |
| Remote Configuration Management | View and edit software configuration files on remote servers online (`config.json` for opentraffic-ops-proxy, `config.yaml` for opentraffic-ops) |
| Remote Service Management | Start, stop, and restart services on remote servers via PID files, without requiring root privileges |
| Deployment Audit Trail | Complete logging of every remote deployment operation, execution results, and historical records |

## Tech Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: SQLite (zero external dependencies)
- **Container Management**: Docker SDK for Go
- **SSH Client**: Based on `golang.org/x/crypto/ssh`
- **Authentication**: JWT
- **Static File Hosting**: `go:embed` + custom SPA fallback handler
- **Encryption**: AES-GCM for sensitive data (SSH passwords, private keys)

### Frontend
- **Framework**: Vue 3 + TypeScript
- **UI Library**: Element Plus
- **Build Tool**: Vite
- **State Management**: Pinia
- **Charts**: ECharts
- **Routing**: `createWebHistory` (History mode)

## Feature Modules

### 1. Dashboard
- Component stat cards (total / running / stopped / error)
- Server stat cards (total / password auth / key auth / configured for deployment)
- Component type distribution pie chart
- Component status distribution bar chart
- Real-time component monitoring table (CPU / memory / network IO) with start/stop live refresh

### 2. Component Management
- Browse component catalog with Docker connection status
- One-click install components (PostgreSQL, Redis)
- Customize during installation: component name, port, environment variables, volumes, startup command arguments
- Start / stop / restart / uninstall installed components
- View component details (resource monitoring, logs, configuration)
- Built-in offline images — no internet required for deployment

#### Supported Component Types

| Component | Type | Default Image | Description |
|-----------|------|--------------|-------------|
| PostgreSQL | `postgresql` | `postgres:16-alpine` | Relational database |
| Redis | `redis` | `redis:7-alpine` | In-memory cache / key-value store |

#### Component Detail Page
- **Basic Info**: component name, type, image, version, status, container ID, create/update timestamps
- **Configuration**: full configuration displayed in JSON format
- **Resource Monitoring**: CPU usage, memory usage/limit, network rx/tx, disk read/write
- **Log Viewer**: support for recent 100/500/1000 lines, with auto-refresh

### 3. Server Management
- Add / edit / delete remote server SSH configurations
- Two authentication methods supported:
  - **Password Auth**: username + password
  - **Key Auth**: SSH private key (supports keys with Passphrase)
- SSH connection test
- Server list displays service status (proxy / monitor)
- Expandable rows to view deployed service details
- Supported operations: start / stop / restart / configure / uninstall remote services

### 4. Remote Deployment
- Select target servers to deploy built-in binaries:
  - `opentraffic-ops-proxy` — OpenTraffic Ops Proxy data collection agent
  - `opentraffic-ops` — OpenTraffic Ops monitoring platform service
- Optionally deploy configuration files simultaneously
- Support loading default configuration templates
- Duplicate deployment detection
- Complete deployment records and log traceability

### 5. Configuration Management
- View configuration list of all installed components
- Edit component configurations online (ports, environment variables, volumes, startup commands)
- Manual restart required after configuration changes take effect

### 6. User Guide
- Platform introduction and feature overview
- Basic environment requirements (Docker, browser, network, SSH)
- Component management usage instructions (supported components, configuration items, common operations)
- Server management usage instructions (configuration items, operations, auth methods)
- Remote deployment process instructions
- PostgreSQL / Redis default configuration and parameter descriptions
- Common FAQ (accordion-style interaction)

## Single-Binary Self-Contained Deployment (No Nginx)

A core design goal of this project is to **eliminate dependency on external web servers** (like Nginx). Traditional Vue `createWebHistory` projects typically require Nginx for static file serving and route fallback, while this project uses Go's native `go:embed` mechanism to embed the frontend `dist` directory directly into the backend binary, with the Go backend providing unified HTTP service.

### Implementation

#### 1. Frontend Build Artifacts Embedded in Go Binary

In `backend/pkg/static/static.go`, the `//go:embed` directive embeds the `frontend/dist` directory at compile time:

```go
//go:embed all:../../../frontend/dist
var dist embed.FS
```

This means after running `go build`, all frontend HTML, JS, CSS, and image resources are contained within a single executable file — no need to keep a `frontend/dist` folder on the server.

#### 2. SPA Fallback Logic

The frontend uses Vue Router's `createWebHistory()` (History mode). When users directly access routes like `/components`, `/configs`, or refresh the page, the browser requests a static file path that doesn't exist on disk. If the backend returns 404 directly, the SPA will not work.

Therefore, a custom `http.Handler` is implemented:

```go
func Handler() http.Handler {
    distFS, _ := fs.Sub(dist, "frontend/dist")
    fileServer := http.FileServer(http.FS(distFS))

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. API and health routes should not be handled by static file handler
        if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/health" {
            http.NotFound(w, r)
            return
        }

        // 2. Try to open requested file
        f, err := distFS.Open(strings.TrimPrefix(r.URL.Path, "/"))
        if err != nil {
            // File not found -> fallback to index.html
            r.URL.Path = "/"
            fileServer.ServeHTTP(w, r)
            return
        }
        defer f.Close()

        // 3. If path is a directory, also fallback to index.html
        if stat, _ := f.Stat(); stat.IsDir() {
            r.URL.Path = "/"
            fileServer.ServeHTTP(w, r)
            return
        }

        // 4. Real static file exists, serve directly
        fileServer.ServeHTTP(w, r)
    })
}
```

In `backend/cmd/server/main.go`, after all `/api/*` and `/health` routes are registered, this Handler is set as the catch-all via `NoRoute`:

```go
r.NoRoute(gin.WrapH(static.Handler()))
```

Gin's matching priority ensures: explicitly registered API routes are matched first, and only unmatched paths enter the static file handler.

#### 3. Same-Origin Deployment, No CORS Issues

The frontend API base URL is set to `/api` (`frontend/src/api/index.ts`):

```typescript
baseURL: '/api'
```

During development, Vite proxies `/api` to `localhost:8080`; in production, frontend and backend share the same port and domain, so there are no cross-origin issues and no special CORS configuration is needed.

### Key Advantages

| Feature | Traditional Nginx + Backend Separation | This Project (Single Binary) |
|--------|--------------------------------------|------------------------------|
| Deployment Files | Multiple files/directories + Nginx config | **Single binary file** |
| Port Exposure | 80/443 + 8080 | **Single port only** |
| Route Refresh | Requires Nginx `try_files` config | **Backend auto fallback** |
| Environment Dependencies | Requires Nginx installation | **Docker only** |
| Migration Cost | Must sync frontend resource directory | **Copy one file** |

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

# Install dependencies
go mod download

# Run backend service
go run cmd/server/main.go
```

Backend service starts at `http://localhost:8080`.

#### 3. Start Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start dev server
npm run dev
```

Frontend dev server starts at `http://localhost:5173`, with Vite Proxy automatically forwarding `/api` to `http://localhost:8080`.

#### 4. Access System

Open browser and visit `http://localhost:5173`

Default login credentials:
- Username: `admin`
- Password: `admin123`

### Production Build (Single Binary)

#### Windows Local Build

Run `build.bat` in the project root:

```cmd
build.bat
```

After build completes, `backend\opentraffic-ops-init.exe` is the final artifact.

#### Windows Cross-Compile for Linux

Run `build-linux.bat` to generate Linux AMD64 and ARM64 binaries:

```cmd
build-linux.bat
```

Output files:
- `backend\opentraffic-ops-init-linux-amd64`
- `backend\opentraffic-ops-init-linux-arm64`

Upload to Linux server and run:

```bash
chmod +x opentraffic-ops-init-linux-amd64
./opentraffic-ops-init-linux-amd64
```

#### Linux / macOS / Manual Build

```bash
# 1. Build frontend
cd frontend
npm install
npm run build

# 2. Copy frontend artifacts to backend embed directory
cd ..
mkdir -p backend/pkg/static/dist
cp -r frontend/dist/* backend/pkg/static/dist/

# 3. Build backend single binary (frontend dist embedded)
cd backend
go build -o opentraffic-ops-init cmd/server/main.go
```

> **Note**: `go:embed` requires embedded files to be within the Go module, and paths cannot contain `..`. Therefore, `frontend/dist` must first be copied to `backend/pkg/static/dist` before running `go build`.

#### Windows Local Development Quick Debug (No dist copy needed)

During development, frontend changes are frequent, and copying `dist` into the backend for every recompile is tedious. An environment variable switch is added in `backend/pkg/static/static.go`:

```cmd
# In backend directory
set RTM_STATIC_DIR=..\frontend\dist
go run cmd\server\main.go
```

After setting `RTM_STATIC_DIR`, the Go backend loads frontend assets directly from disk instead of `go:embed`. This allows you to modify frontend code and refresh the browser without recompiling the backend. **Do not** set this variable for production builds, to ensure frontend resources are fully embedded in the binary.

Regardless of the build method, the final artifact is always a **single self-contained binary file**. After running, visit `http://localhost:8080` to see the complete platform. Direct refresh of `http://localhost:8080/components` will not return 404.

### Docker Deployment

#### Using Docker Compose (Full Environment)

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

#### Deploy Backend Binary Alone

If you have already embedded the frontend into the binary, you can build and run the backend image directly:

```bash
cd backend

# Build image
docker build -t opentraffic-ops-init .

# Run container
docker run -d \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v rtm-data:/app/data \
  --name opentraffic-ops-init \
  opentraffic-ops-init
```

## Project Structure

```
opentraffic-ops-init/
├── backend/                      # Go backend
│   ├── cmd/server/              # Entry files
│   ├── internal/
│   │   ├── controller/          # API controllers
│   │   │   ├── auth_controller.go
│   │   │   ├── component_controller.go
│   │   │   ├── deploy_controller.go
│   │   │   ├── monitor_controller.go
│   │   │   └── server_controller.go
│   │   ├── service/             # Business logic
│   │   ├── repository/          # Data access
│   │   ├── model/               # Data models
│   │   └── middleware/          # Middleware (JWT, CORS, Recovery, ErrorHandler)
│   ├── pkg/
│   │   ├── docker/              # Docker client wrapper
│   │   ├── ssh/                 # SSH client wrapper
│   │   ├── config/              # Configuration management
│   │   ├── crypto/              # Encryption utilities (AES-GCM)
│   │   ├── static/              # Static file hosting (go:embed)
│   │   └── assets/              # Embedded resources (default configs, binaries)
│   └── configs/                 # Configuration files
│
├── frontend/                     # Vue frontend
│   ├── src/
│   │   ├── views/               # Page components
│   │   │   ├── Dashboard.vue    # Monitoring dashboard
│   │   │   ├── Components.vue   # Component management
│   │   │   ├── ComponentDetail.vue  # Component details
│   │   │   ├── Servers.vue      # Server management
│   │   │   ├── Configs.vue      # Configuration management
│   │   │   ├── Help.vue         # User guide
│   │   │   ├── Login.vue        # Login page
│   │   │   └── Layout.vue       # Layout page
│   │   ├── components/          # Common components
│   │   ├── api/                 # API call wrappers
│   │   ├── stores/              # Pinia state management
│   │   └── router/              # Router config (createWebHistory)
│   └── package.json
│
├── components/                   # Component templates (Docker Compose, config templates)
│   ├── postgresql/
│   └── redis/
│
├── docker-compose.yaml           # Docker orchestration
└── README.md                     # This document
```

## API Documentation

### Auth Endpoints

#### Login
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

#### Logout
```http
POST /api/auth/logout
Authorization: Bearer <token>
```

#### Get User Profile
```http
GET /api/users/profile
Authorization: Bearer <token>
```

### Component Management Endpoints

```http
# Get component catalog (with install status)
GET /api/components/catalog

# Get component list
GET /api/components

# Get component details
GET /api/components/:id

# Install component
POST /api/components

# Uninstall component
DELETE /api/components/:id

# Start/stop/restart component
POST /api/components/:id/start
POST /api/components/:id/stop
POST /api/components/:id/restart

# Get component logs
GET /api/components/:id/logs

# Get component resource stats
GET /api/components/:id/stats

# Update component config
PUT /api/components/:id/config
```

### Monitoring Endpoints

```http
# Get system overview
GET /api/monitor/overview

# Get component details (with stats)
GET /api/monitor/components

# WebSocket real-time monitoring
GET /api/monitor/realtime
```

### Server Management Endpoints

```http
# Get server list
GET /api/servers

# Create server
POST /api/servers

# Get server details
GET /api/servers/:id

# Update server
PUT /api/servers/:id

# Delete server
DELETE /api/servers/:id

# Test SSH connection
POST /api/servers/:id/test

# Get opentraffic-ops-proxy config
GET /api/servers/:id/proxy-config
PUT /api/servers/:id/proxy-config

# Get/update software config
GET /api/servers/:id/configs/:software
PUT /api/servers/:id/configs/:software

# Get default software config (embedded resource)
GET /api/servers/configs/:software/default

# Get service running status
GET /api/servers/:id/services/:software/status

# Start/stop/restart remote service
POST /api/servers/:id/services/:software/start
POST /api/servers/:id/services/:software/stop
POST /api/servers/:id/services/:software/restart
```

### Deployment Endpoints

```http
# Deploy binary to remote server
POST /api/deploy

# Undeploy remote service
POST /api/deploy/undeploy

# Get deployment records list
GET /api/deploy/records?server_id=

# Get deployment record details
GET /api/deploy/records/:id
```

## Configuration

### Backend Configuration

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

### Frontend Configuration

Frontend configuration is in `frontend/vite.config.ts`:

```typescript
export default defineConfig({
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
```

- `base` defaults to `/`, deployed at root path in production — no modification needed.
- `proxy` only takes effect in `npm run dev` development mode; after production build, Go backend serves directly.

## Production Deployment

### Security Recommendations

1. Change the default admin password
2. Replace JWT Secret and encryption key
3. Use HTTPS (can terminate TLS via external reverse proxy or load balancer)
4. Configure firewall rules
5. Regularly back up the SQLite database

### Performance Optimization

1. Use production build (`npm run build`)
2. Backend has Gzip enabled by default (Gin framework)
3. Static resources served by `go:embed` with extremely fast memory access
4. For higher performance, external CDN or cache layer can be added

### Backup Strategy

```bash
# Backup SQLite database
cp backend/data/rtm_init.db backup/rtm_init_$(date +%Y%m%d).db

# Backup Docker volumes
docker run --rm \
  -v rtm-data:/data \
  -v $(pwd)/backup:/backup \
  alpine tar czf /backup/rtm-data_$(date +%Y%m%d).tar.gz /data
```

## Troubleshooting

### Common Issues

1. **Docker connection failed**
   - Ensure Docker service is running
   - Check `/var/run/docker.sock` permissions

2. **Port conflict**
   - Modify port configuration in `.env` file
   - Ensure port is not occupied

3. **Frontend refresh returns 404**
   - Confirm `frontend/dist` exists during `go build`
   - Check embed path in `backend/pkg/static/static.go`

4. **CORS issues**
   - Production environment is same-origin, should not have CORS
   - Development environment: ensure Vite Proxy config is correct and backend is running

5. **SSH connection test failed**
   - Check if target server SSH service is running
   - Confirm host address, port, username, password/private key are correct
   - Check if firewall allows SSH port

6. **Remote deployment failed (insufficient permissions)**
   - Check if SSH user has read/write permissions for deployment path
   - Confirm deployment path disk has sufficient space
   - Check target server's SELinux or AppArmor restrictions

7. **Component container start failed (Permission denied)**
   - When using bind mounts, ensure host directory owner matches container default user UID
   - PostgreSQL UID is 70, Redis UID is 999
   - Recommend using named volumes (e.g., `postgres-data:/var/lib/postgresql/data`), Docker handles permissions automatically

## Development Guide

### Adding New Component Types

1. Create new component directory under `components/`
2. Add `config.yaml.template` and `docker-compose.yaml.template`
3. Add new component type in backend `internal/model/component.go`
4. Update `ComponentType` type in frontend `src/types/index.ts`

### Adding New Deployable Binaries

1. Place binary file in `backend/pkg/assets/` directory
2. Add `//go:embed` directive in `backend/pkg/assets/assets.go`
3. Update option list in frontend deployment dialog
4. Add corresponding default config file under `backend/pkg/assets/images/`

## Contributing

Issues and Pull Requests are welcome!

## License

[MIT License](../LICENSE)
