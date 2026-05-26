# OpenTraffic Ops Proxy

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

OpenTraffic Ops — Edge Proxy. **Linux only** (x86_64 / ARM64), deployed on Linux servers to collect system metrics and report to the platform server, with WebSocket remote control support (terminal / file management).

```
┌────────────────────┐     HTTP POST      ┌──────────────────┐
│   OpenTraffic Ops Proxy       │  ──────────────▶  │  OpenTraffic Ops │
│   (Linux Server)              │  ◀──────────────  │  (Server)                 │
└────────────────────┘     Return Metrics   └──────────────────────┘
         │
         │  WebSocket (Long Connection)
         ▼
┌──────────────────────────────┐
│  Remote Terminal / File Mgmt / Shell │
└──────────────────────────────┘
```

The Proxy periodically executes the following tasks:
- **Heartbeat Report** (default 3s): Maintains host online status, while reporting CPU / memory / disk / network / process metrics
- **WebSocket Connection**: Establishes a persistent connection to the platform to receive remote control commands

> **Important**: This Proxy does not support Windows or macOS. Development can be done on Windows, but only for cross-compilation; running and testing must be performed on Linux servers or virtual machines.

### Supported Platforms

| OS | Architecture | Status |
|---------|------|---------|
| Linux   | x86_64 (amd64) | Fully Supported |
| Linux   | ARM64 (aarch64) | Fully Supported |
| Windows | Any | Not Supported |
| macOS   | Any | Not Supported |

---

## Tech Stack

| Technology | Version | Description |
|-----------|---------|-------------|
| Go | 1.26+ | Programming language (**Linux only**) |
| gopsutil | v3 | Host metric collection |
| Gorilla WS | v1.5 | WebSocket long connection to platform |
| creack/pty | v1.1 | Remote terminal PTY implementation |

> Proxy and backend do not share code; they interact via HTTP/WS protocol. Can only run on Linux (amd64 / arm64); Windows is for cross-compilation only.

---

## Features

- **System Info Collection** — Reports OS type/version, CPU arch/cores/model, memory, disk, MAC address on registration
- **System Metric Collection** — Reports CPU / memory / disk / network / load every 3 seconds
- **Process Monitoring** — Collects configured process running status, CPU usage, memory usage
- **WebSocket Long Connection** — Auto-reconnect (exponential backoff), heartbeat keepalive, safe goroutine shutdown
- **Remote Terminal** — PTY-based persistent shell sessions (5-minute timeout auto-close)
- **Remote File Management** — Complete file operations with path security validation

### Collected Metrics

- **CPU**: Overall usage (%)
- **Memory**: Usage (%), Used MB
- **Disk**: Root partition usage (%)
- **Network**: In/Out throughput (KB/s)
- **Load**: 1/5/15 minute average load
- **Processes**: Running status, CPU usage, Memory usage MB

### Platform Interaction APIs

#### HTTP APIs (No Authentication)

| Method | Path | Description |
|------|------|------|
| POST | `/api/v1/proxy/register` | First-time registration, reports hardware info |
| POST | `/api/v1/proxy/heartbeat` | Heartbeat keepalive + monitoring data report |

#### WebSocket API (No Authentication)

| Path | Description |
|------|------|
| `ws://platform/api/v1/proxy/ws?ip=xxx` | Proxy establishes WebSocket long connection |

After the WebSocket connection is established, the platform can send:
- **Terminal Input** (`input`) → Proxy writes to Shell stdin
- **Terminal Resize** (`resize`) → Proxy adjusts terminal size
- **File Operations** (`file_list`/`file_read`/`file_write`/`file_delete`/`file_upload`/`file_download`/`file_mkdir`)

---

## Quick Start

### Prerequisites

- Go 1.22+ (project uses Go 1.26.2)
- Git
- Windows PowerShell (for running the one-click packaging script)

### Verify Cross-Compilation Environment

```powershell
# Check Go version
go version

# Verify cross-compilation to Linux
cd proxy
$env:GOOS = "linux"; $env:GOARCH = "amd64"; $env:CGO_ENABLED = "0"; go build -o opentraffic-ops-proxy .
```

If there is no error output, the cross-compilation environment is working. **Note: this binary cannot run on Windows** and must be uploaded to a Linux server for execution.

### Manual Cross-Compilation (Alternative)

```powershell
cd proxy

# Linux x86_64 (most common servers)
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -ldflags "-s -w" -o opentraffic-ops-proxy-linux-amd64 .

# Linux ARM64 (e.g., Raspberry Pi, ARM cloud servers)
$env:GOOS = "linux"
$env:GOARCH = "arm64"
$env:CGO_ENABLED = "0"
go build -ldflags "-s -w" -o opentraffic-ops-proxy-linux-arm64 .
```

### Build Parameter Reference

| Parameter | Description |
|------|------|
| `-s` | Strip symbol table to reduce size |
| `-w` | Strip DWARF debug info |
| `CGO_ENABLED=0` | Disable CGO, static linking, ensuring cross-distro compatibility |

---

## Server Deployment

### Production Packaging (Windows One-Click Script)

On the Windows development machine, use the provided PowerShell script for one-click packaging:

```batch
cd proxy
build-opentraffic-ops-proxy.bat
```

The script will automatically compile the following targets and output to the `dist/` directory:

| Output File | Target Platform |
|---------|---------|
| `opentraffic-ops-proxy-linux-amd64` | Linux x86_64 |
| `opentraffic-ops-proxy-linux-arm64` | Linux ARM64 |

### Deploy to Linux Server

#### 1. Upload Binary and Config

```bash
# Upload from Windows to Linux server
scp opentraffic-ops-proxy-linux-amd64 root@your-server:/opt/opentraffic-ops-proxy/
scp config.json root@your-server:/opt/opentraffic-ops-proxy/
```

#### 2. Run Directly (Testing / Debugging)

```bash
cd /opt/opentraffic-ops-proxy
chmod +x opentraffic-ops-proxy-linux-amd64
./opentraffic-ops-proxy-linux-amd64 -c config.json
```

On first run, a default config file will be automatically created at `~/.opentraffic-ops-proxy/config.json`.

### Configuration

```json
{
  "platformUrl": "http://192.168.1.100:8080",
  "ip": "",
  "hostName": "",
  "version": "1.0.0",
  "heartbeatInterval": 3,
  "logLevel": "info",
  "logFile": "",
  "enableRemote": true,
  "wsEndpoint": "",
  "processes": [
    {
      "name": "my-app",
      "pattern": "my-app",
      "execCmd": "/opt/my-app/bin/start.sh"
    }
  ]
}
```

| Config Item | Type | Description |
|--------|------|------|
| `platformUrl` | string | Platform server address (HTTP) |
| `ip` | string | Local IP (auto-detected if empty) |
| `hostName` | string | Host name (uses system hostname if empty) |
| `heartbeatInterval` | int | Heartbeat interval (seconds), default 3 |
| `logLevel` | string | Log level: debug/info/warn/error |
| `logFile` | string | Log file path (outputs to console if empty) |
| `enableRemote` | bool | Remote control switch (terminal/file), default `true` |
| `wsEndpoint` | string | WebSocket endpoint (auto-derived from `platformUrl` if empty) |
| `processes` | array | List of processes to monitor |

**Configuration Notes**:
- **`enableRemote`**: Set to `false` to disable remote terminal and file management; the Proxy will reject all remote operation requests
- **`wsEndpoint`**: When the platform WebSocket uses an independent port or reverse proxy, you can manually specify it, e.g., `ws://192.168.1.100:8081/api/v1/proxy/ws`

---

## FAQ

### Why can't I run the compiled binary on Windows?

Proxy uses Linux-specific system calls (`creack/pty`, `/proc` filesystem access) that are not available on Windows. The binary **must** run on Linux (amd64 or arm64). Windows is only a build host for cross-compilation.

### Connection to platform refused / timeout

- Verify the `platformUrl` in `config.json` points to the correct OpenTraffic Ops server address and port
- Check firewall rules between the Proxy host and the platform server
- Ensure the platform server is running and accessible

### Config file not found on first run

If you start Proxy without specifying `-c config.json`, it will attempt to create a default config at `~/.opentraffic-ops-proxy/config.json`. Ensure the user running Proxy has write permissions to their home directory.

### WebSocket keeps disconnecting

- Check network stability between Proxy and platform
- Verify the platform's WebSocket endpoint (`wsEndpoint`) is correctly configured
- Check `journalctl -u opentraffic-ops-proxy -f` for error logs

### Process monitoring shows no data

- Ensure the `processes` array in `config.json` is properly configured with correct `pattern` values
- The `pattern` is used to match process names via `ps` / `pgrep`

---

## Acknowledgments

OpenTraffic Ops Proxy is built with the following open-source projects:

- [Go](https://golang.org/) — Programming language
- [gopsutil](https://github.com/shirou/gopsutil) — System metrics collection
- [Gorilla WebSocket](https://github.com/gorilla/websocket) — WebSocket client implementation
- [creack/pty](https://github.com/creack/pty) — PTY for remote terminal

[MIT License](../LICENSE)
