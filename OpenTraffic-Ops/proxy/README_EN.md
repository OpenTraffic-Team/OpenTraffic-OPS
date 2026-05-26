# OpenTraffic Ops Proxy

[中文](README.md)

OpenTraffic Ops — Edge Proxy. **Linux only** (x86_64 / ARM64), deployed on Linux servers to collect system metrics and report to the platform server, with WebSocket remote control support (terminal / file management).

> ⚠️ **Important**: This Proxy does not support Windows or macOS. Development can be done on Windows, but only for cross-compilation; running and testing must be performed on Linux servers or virtual machines.

---

## Architecture

```
┌─────────────────────┐     HTTP POST      ┌─────────────────┐
│   OpenTraffic Ops Proxy         │  ───────────────►  │  OpenTraffic Ops Platform   │
│   (Linux Server)     │  ◄───────────────  │  (Server)        │
└─────────────────────┘     Return Commands  └─────────────────┘
         │
         │  WebSocket (Long Connection)
         ▼
┌─────────────────────────────┐
│  Remote Terminal / File Mgmt / Shell  │
└─────────────────────────────┘
```

The Proxy periodically executes the following tasks:
- **Heartbeat Report** (default 3s): Maintains host online status, while reporting CPU / memory / disk / network / process metrics
- **Command Polling** (default 10s): Pulls process start/stop commands issued by the platform
- **WebSocket Connection**: Establishes a persistent connection to the platform to receive remote control commands

---

## Supported Platforms

| OS | Architecture | Status |
|---------|------|---------|
| Linux   | x86_64 (amd64) | ✅ Fully Supported |
| Linux   | ARM64 (aarch64) | ✅ Fully Supported |
| Windows | Any | ❌ Not Supported |
| macOS   | Any | ❌ Not Supported |

---

## Development Environment (Windows Cross-Compilation)

The Proxy is written in Go. Development can be done on Windows via **cross-compilation** to produce Linux binaries.

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

---

## Production Packaging (Windows One-Click Script)

On the Windows development machine, use the provided PowerShell script for one-click packaging:

```powershell
cd proxy
.\build-proxy.ps1

# Or specify a version number
.\build-proxy.ps1 -Version "1.1.0"
```

The script will automatically compile the following targets and output to the `dist/` directory:

| Output File | Target Platform |
|---------|---------|
| `opentraffic-ops-proxy-linux-amd64` | Linux x86_64 |
| `opentraffic-ops-proxy-linux-arm64` | Linux ARM64 |

### Manual Cross-Compilation (Alternative)

If you prefer not to use the script, you can compile manually:

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

## Deploy to Linux Server

### 1. Upload Binary and Config

```bash
# Upload from Windows to Linux server
scp opentraffic-ops-proxy-linux-amd64 root@your-server:/opt/opentraffic-ops-proxy/
scp config.json root@your-server:/opt/opentraffic-ops-proxy/
```

### 2. Configure systemd Service (Recommended)

On the target Linux server:

```bash
sudo tee /etc/systemd/system/opentraffic-ops-proxy.service > /dev/null << 'EOF'
[Unit]
Description=OpenTraffic Ops Proxy
After=network.target

[Service]
Type=simple
ExecStart=/opt/opentraffic-ops-proxy/opentraffic-ops-proxy-linux-amd64 -c /opt/opentraffic-ops-proxy/config.json
Restart=always
RestartSec=10
User=root
WorkingDirectory=/opt/opentraffic-ops-proxy

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable opentraffic-ops-proxy
sudo systemctl start opentraffic-ops-proxy

# Check status
sudo systemctl status opentraffic-ops-proxy

# View logs
sudo journalctl -u opentraffic-ops-proxy -f
```

### 3. Run Directly (Testing / Debugging)

```bash
cd /opt/opentraffic-ops-proxy
chmod +x opentraffic-ops-proxy-linux-amd64
./opentraffic-ops-proxy-linux-amd64 -c config.json
```

On first run, a default config file will be automatically created at `~/.opentraffic-ops-proxy/config.json`.

---

## Configuration

```json
{
  "platformUrl": "http://192.168.1.100:8080",
  "ip": "",
  "hostName": "",
  "version": "1.0.0",
  "heartbeatInterval": 3,
  "pollInterval": 10,
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
| `pollInterval` | int | Command polling interval (seconds), default 10 |
| `logLevel` | string | Log level: debug/info/warn/error |
| `logFile` | string | Log file path (outputs to console if empty) |
| `enableRemote` | bool | Remote control switch (terminal/file), default `true` |
| `wsEndpoint` | string | WebSocket endpoint (auto-derived from `platformUrl` if empty) |
| `processes` | array | List of processes to monitor |

### Configuration Notes

- **`enableRemote`**: Set to `false` to disable remote terminal and file management; the Proxy will reject all remote operation requests
- **`wsEndpoint`**: When the platform WebSocket uses an independent port or reverse proxy, you can manually specify it, e.g., `ws://192.168.1.100:8081/api/v1/proxy/ws`

---

## Platform Interaction APIs

### HTTP APIs (No Authentication)

| Method | Path | Description |
|------|------|------|
| POST | `/api/v1/proxy/register` | First-time registration, reports hardware info |
| POST | `/api/v1/proxy/heartbeat` | Heartbeat keepalive + monitoring data report |
| POST | `/api/v1/proxy/poll` | Poll pending commands |
| POST | `/api/v1/proxy/ack` | Report command execution result |

### WebSocket API (No Authentication)

| Path | Description |
|------|------|
| `ws://platform/api/v1/proxy/ws?ip=xxx` | Proxy establishes WebSocket long connection |

After the WebSocket connection is established, the platform can send:
- **Terminal Input** (`input`) → Proxy writes to Shell stdin
- **Terminal Resize** (`resize`) → Proxy adjusts terminal size
- **File Operations** (`file_list`/`file_read`/`file_write`/`file_delete`/`file_upload`/`file_download`/`file_mkdir`)

## Supported Command Types

The platform can send the following commands to the Proxy via the Redis command queue or WebSocket:

| Command Type | Description |
|----------|------|
| `startProcess` | Start the specified process |
| `stopProcess` | Stop the specified process |
| `restartProcess` | Restart the specified process |

## Collected Metrics

- **CPU**: Overall usage (%)
- **Memory**: Usage (%), Used MB
- **Disk**: Root partition usage (%)
- **Network**: In/Out throughput (KB/s)
- **Load**: 1/5/15 minute average load
- **Processes**: Running status, CPU usage, Memory usage MB
