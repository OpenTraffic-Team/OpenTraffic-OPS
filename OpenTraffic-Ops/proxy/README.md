# OpenTraffic Ops Proxy

[English](README_EN.md)

OpenTraffic Ops —— 边缘端 Proxy。**仅支持 Linux 操作系统**（x86_64 / ARM64），部署在 Linux 服务器上，负责采集系统指标并上报到平台服务端，同时支持 WebSocket 远程控制（终端/文件管理）。

> ⚠️ **重要说明**：本 Proxy 不支持 Windows 和 macOS。开发环境在 Windows 上，但只能用于交叉编译；运行和测试必须在 Linux 服务器/虚拟机上进行。

---

## 架构说明

```
┌─────────────────────┐     HTTP POST      ┌─────────────────┐
│   OpenTraffic Ops Proxy         │  ───────────────►  │  OpenTraffic Ops Platform   │
│   (Linux 服务器)     │  ◄───────────────  │  (服务端)        │
└─────────────────────┘     返回指令        └─────────────────┘
         │
         │  WebSocket（长连接）
         ▼
┌─────────────────────────────┐
│  远程终端 / 文件管理 / Shell   │
└─────────────────────────────┘
```

Proxy 定时执行以下任务：
- **心跳上报**（默认 3s）：保持主机在线状态，同时上报 CPU/内存/磁盘/网络/进程指标
- **指令轮询**（默认 10s）：拉取平台下发的进程启停指令
- **WebSocket 连接**：建立到平台的持久连接，接收远程控制指令

---

## 支持平台

| 操作系统 | 架构 | 支持状态 |
|---------|------|---------|
| Linux   | x86_64 (amd64) | ✅ 完全支持 |
| Linux   | ARM64 (aarch64) | ✅ 完全支持 |
| Windows | 任意 | ❌ 不支持 |
| macOS   | 任意 | ❌ 不支持 |

---

## 开发环境（Windows 交叉编译）

Proxy 采用 Go 编写，开发环境可以在 Windows 上通过**交叉编译**生成 Linux 二进制文件。

### 前置要求

- Go 1.22+（项目使用 Go 1.26.2）
- Git
- Windows PowerShell（用于执行一键打包脚本）

### 验证交叉编译环境

```powershell
# 检查 Go 版本
go version

# 验证能否交叉编译到 Linux
cd proxy
$env:GOOS = "linux"; $env:GOARCH = "amd64"; $env:CGO_ENABLED = "0"; go build -o opentraffic-ops-proxy .
```

如果输出没有报错，说明交叉编译环境正常。**注意：这个二进制在 Windows 上无法运行**，必须上传到 Linux 服务器执行。

---

## 生产打包（Windows 一键脚本）

在 Windows 开发机上使用提供的 PowerShell 脚本一键打包：

```powershell
cd proxy
.\build-proxy.ps1

# 或指定版本号
.\build-proxy.ps1 -Version "1.1.0"
```

脚本会自动编译以下目标并输出到 `dist/` 目录：

| 输出文件 | 目标平台 |
|---------|---------|
| `opentraffic-ops-proxy-linux-amd64` | Linux x86_64 |
| `opentraffic-ops-proxy-linux-arm64` | Linux ARM64 |

### 手动交叉编译（备用）

如果不用脚本，也可以手动编译：

```powershell
cd proxy

# Linux x86_64（最常见的服务器）
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -ldflags "-s -w" -o opentraffic-ops-proxy-linux-amd64 .

# Linux ARM64（如树莓派、ARM 云服务器）
$env:GOOS = "linux"
$env:GOARCH = "arm64"
$env:CGO_ENABLED = "0"
go build -ldflags "-s -w" -o opentraffic-ops-proxy-linux-arm64 .
```

### 编译参数说明

| 参数 | 说明 |
|------|------|
| `-s` | 去除符号表，减小体积 |
| `-w` | 去除 DWARF 调试信息 |
| `CGO_ENABLED=0` | 禁用 CGO，静态链接，确保跨发行版兼容 |

---

## 部署到 Linux 服务器

### 1. 上传二进制和配置

```bash
# 从 Windows 上传到 Linux 服务器
scp opentraffic-ops-proxy-linux-amd64 root@your-server:/opt/opentraffic-ops-proxy/
scp config.json root@your-server:/opt/opentraffic-ops-proxy/
```

### 2. 配置 systemd 服务（推荐）

在目标 Linux 服务器上执行：

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

# 查看状态
sudo systemctl status opentraffic-ops-proxy

# 查看日志
sudo journalctl -u opentraffic-ops-proxy -f
```

### 3. 直接运行（测试/调试）

```bash
cd /opt/opentraffic-ops-proxy
chmod +x opentraffic-ops-proxy-linux-amd64
./opentraffic-ops-proxy-linux-amd64 -c config.json
```

首次运行会自动在用户目录下创建默认配置文件 `~/.opentraffic-ops-proxy/config.json`。

---

## 配置文件

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

| 配置项 | 类型 | 说明 |
|--------|------|------|
| `platformUrl` | string | 平台服务端地址（HTTP） |
| `ip` | string | 本机 IP（留空则自动检测） |
| `hostName` | string | 主机名（留空则使用系统主机名） |
| `heartbeatInterval` | int | 心跳间隔（秒），默认 3 |
| `pollInterval` | int | 指令轮询间隔（秒），默认 10 |
| `logLevel` | string | 日志级别：debug/info/warn/error |
| `logFile` | string | 日志文件路径（留空则输出到控制台） |
| `enableRemote` | bool | 远程控制开关（终端/文件），默认 `true` |
| `wsEndpoint` | string | WebSocket 端点（留空则自动从 `platformUrl` 推导） |
| `processes` | array | 需要监控的进程列表 |

### 配置项说明

- **`enableRemote`**: 设为 `false` 可禁用远程终端和文件管理功能，Proxy 将拒绝所有远程操作请求
- **`wsEndpoint`**: 当平台 WebSocket 使用独立端口或反向代理时，可手动指定，如 `ws://192.168.1.100:8081/api/v1/proxy/ws`

---

## 与平台交互的接口

### HTTP 接口（无需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/proxy/register` | 首次注册，上报硬件信息 |
| POST | `/api/v1/proxy/heartbeat` | 心跳保活 + 监控数据上报 |
| POST | `/api/v1/proxy/poll` | 轮询待执行指令 |
| POST | `/api/v1/proxy/ack` | 指令执行结果上报 |

### WebSocket 接口（无需认证）

| 路径 | 说明 |
|------|------|
| `ws://platform/api/v1/proxy/ws?ip=xxx` | Proxy 建立 WebSocket 长连接 |

WebSocket 连接建立后，平台可通过该通道下发：
- **终端输入** (`input`) → Proxy 写入 Shell stdin
- **终端 resize** (`resize`) → Proxy 调整终端大小
- **文件操作** (`file_list`/`file_read`/`file_write`/`file_delete`/`file_upload`/`file_download`/`file_mkdir`)

## 支持的指令类型

平台可通过 Redis 指令队列或 WebSocket 向 Proxy 下发以下指令：

| 指令类型 | 说明 |
|----------|------|
| `startProcess` | 启动指定进程 |
| `stopProcess` | 停止指定进程 |
| `restartProcess` | 重启指定进程 |

## 采集指标

- **CPU**：整体使用率（%）
- **内存**：使用率（%）、使用 MB
- **磁盘**：根分区使用率（%）
- **网络**：入/出流量（KB/s）
- **负载**：1/5/15 分钟平均负载
- **进程**：运行状态、CPU 使用率、内存使用 MB
