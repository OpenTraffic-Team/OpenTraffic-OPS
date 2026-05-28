# OpenTraffic Ops

<p align="center">
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-Apache_2.0-blue.svg" alt="License: Apache 2.0"></a>
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white" alt="Go 1.25+">
  <img src="https://img.shields.io/badge/Vue-3.3-4FC08D?logo=vue.js&logoColor=white" alt="Vue 3">
  <img src="https://img.shields.io/badge/Postgres-15+-4169E1?logo=postgresql&logoColor=white" alt="PostgreSQL 15+">
  <img src="https://img.shields.io/badge/Redis-7+-DC382D?logo=redis&logoColor=white" alt="Redis 7+">
</p>

<p align="center">
  <img src="https://img.shields.io/badge/HuggingFace-%F0%9F%A4%97-FFD21E" alt="HuggingFace">
  <img src="https://img.shields.io/badge/X-000000?logo=x&logoColor=white" alt="X">
  <img src="https://img.shields.io/badge/%E5%B0%8F%E7%BA%A2%E4%B9%A6-FF2442" alt="小红书">
</p>

<p align="center">
  <a href="README.md">English</a>
</p>

一套全栈边缘计算运维平台，由两个紧密集成的子系统组成：用于基础设施供给的**部署面板**，以及用于边缘主机管理、告警和远程运维的**监控平台**。

---

## 📑 目录

- [🏗️ 架构概览](#架构概览)
- [🔧 子系统](#子系统)
  - [1. OpenTraffic-Ops-Initialization — 部署面板](#1-opentraffic-ops-initialization--部署面板)
  - [2. OpenTraffic-Ops — 监控与运维平台](#2-opentraffic-ops--监控与运维平台)
  - [3. proxy — 边缘代理](#3-proxy--边缘代理)
- [🔗 子系统之间的关系](#子系统之间的关系)
- [🚀 快速开始](#快速开始)
  - [📋 前置要求](#前置要求)
  - [🖥️ 启动部署面板](#启动部署面板)
  - [📊 启动监控平台](#启动监控平台)
  - [📦 构建生产二进制文件](#构建生产二进制文件windows-主机交叉编译到-linux)
- [📁 项目结构](#项目结构)
- [📚 文档](#文档)
- [🔒 安全特性](#安全特性)
- [📄 许可证](#许可证)

---

## 架构概览

```
                    +-------------------------------+
                    |     OpenTraffic Ops           |
                    |  +-------------------------+  |
                    |  | OpenTraffic-Ops-Init    |  |
                    |  | (部署面板)               |  |
                    |  | - Docker 管理           |--|---> 部署 OpenTraffic-Ops
                    |  | - SSH 远程部署          |  |    和 Proxy 二进制文件
                    |  | - 组件生命周期管理       |  |    到远程 Linux 服务器
                    |  +-------------------------+  |
                    |  +-------------------------+  |
                    |  | OpenTraffic-Ops         |  |
                    |  | (监控与运维平台)         |  |
                    |  | - 主机监控              |<--|---- 接收来自边缘 Proxy
                    |  | - 告警引擎              |  |    的指标数据
                    |  | - 远程终端              |  |
                    |  | - Agent 对话            |  |
                    |  +-------------------------+  |
                    +-------------------------------+
                                         ^
                                         | WebSocket / HTTP
                                         |
                    +--------------------+------------------+
                    |     proxy (边缘代理)                   |
                    |  - 系统指标采集                        |
                    |  - 进程监控                            |
                    |  - 远程终端 PTY                        |
                    |  - 远程文件操作                        |
                    +--------------------------------------+
                    部署在每个被监控的边缘主机上
```

---

## 子系统

### 1. OpenTraffic-Ops-Initialization — 部署面板

一个单二进制、自包含的部署仪表盘，无需外部 Web 服务器（Nginx）或数据库（PostgreSQL）。

| 能力 | 说明 |
|------|------|
| Docker 管理 | 一键安装/启动/停止/卸载中间件（PostgreSQL、Redis），支持自定义端口、环境变量、数据卷 |
| 实时监控 | 容器实时资源统计（CPU / 内存 / 网络 / 磁盘） |
| SSH 服务器管理 | 集中管理多台远程 Linux 服务器的 SSH 连接配置（密码或密钥认证） |
| 远程二进制部署 | 通过 SSH/SFTP 将 `proxy` 和 `OpenTraffic-Ops` 二进制文件部署到远程服务器 |
| 远程配置编辑 | 在线查看和编辑远程配置文件（`config.json`、`config.yaml`） |
| 远程服务控制 | 通过 PID 文件在远程主机上启动 / 停止 / 重启服务 |
| 部署审计 | 完整的操作日志、执行结果和部署历史 |

**关键特性：**
- **Docker 组件管理**：一键安装/启动/停止/卸载 PostgreSQL、Redis，支持自定义端口、环境变量和数据卷
- **实时监控**：组件实时资源统计（CPU / 内存 / 网络 / 磁盘），日志自动刷新
- **SSH 服务器管理**：集中管理多台远程 Linux 服务器的 SSH 配置，支持密码和密钥两种认证方式
- **远程二进制部署**：一键通过 SSH/SFTP 将 `opentraffic-ops` 和 `opentraffic-ops-proxy` 二进制文件部署到远程 Linux 服务器，支持重复部署检测
- **远程配置管理**：在线查看和编辑远程服务器配置文件
- **远程服务控制**：通过 PID 文件在远程主机上启动/停止/重启服务
- **部署审计追溯**：每次部署的完整操作日志和执行结果

**技术栈：** Go 1.21+ (Gin, SQLite, Docker SDK, `crypto/ssh`), Vue 3 + TypeScript + Vite, Element Plus

**关键设计：** 前端通过 `go:embed` 嵌入 Go 二进制中。后端在同一端口上同时提供 API 和 SPA 静态文件服务，并带有自定义 SPA 回退逻辑 —— 零 Nginx 依赖。

[详情 &rarr;](./OpenTraffic-Ops-Initialization/README_CN.md)

---

### 2. OpenTraffic-Ops — 监控与运维平台

面向边缘计算场景的全栈监控运维平台。包含两个交付件：**监控平台服务**（后端通过 `go:embed` 内嵌前端，单二进制部署）和**边缘代理（Proxy）**。

| 能力 | 说明 |
|------|------|
| 主机管理 | 边缘节点注册、增删改查和状态展示（Proxy 首次注册时自动入库） |
| 健康指标 | 主机历史健康数据，自动按日轮转（保留 7 天） |
| 告警引擎 | 多渠道通知（邮件、钉钉、企业微信、站内信），基于阈值的 CPU / 内存 / 磁盘 / 网络 / 负载 规则 |
| 远程终端 | 浏览器内 xterm 终端，通过 WebSocket Hub 连接到 Proxy PTY（支持颜色、resize） |
| 远程文件操作 | 在 Proxy 主机上浏览、读取、编辑、上传、下载、删除文件（单文件 10MB 限制，目录遍历防护） |
| 进程控制 | 通过平台命令在边缘主机上启动 / 停止 / 重启进程 |
| Agent 对话 | 与控制 Agent 和感知 Agent 进行对话式交互，用于运维协助和主机状态查询 |

**关键特性：**
- **系统管理**：用户管理、个人中心/资料管理
- **主机管理**：边缘节点增删改查，7 天健康历史数据自动按日清理，运维操作入口
- **监控与告警**：多渠道告警通知（邮件、钉钉、企业微信、站内信），基于阈值的 CPU / 内存 / 磁盘 / 网络 / 负载 / 主机离线 / Agent 离线 规则，告警记录，通知日志
- **内置调度器**：`dealOffline`（60s）、`alarmCheck`（30s）、`cleanHostHealth`（每日 03:30）
- **Agent 对话**：控制 Agent 和感知 Agent 对话，会话管理
- **远程运维**：浏览器内 xterm 终端（WebSocket Hub + PTY），远程文件操作（10MB 限制，目录遍历防护），进程控制（启动/停止/重启）
- **系统日志**：操作日志、登录日志

**技术栈：**
- 后端：Go 1.25+ (Gin, GORM, PostgreSQL, Redis, JWT v5, Gorilla WebSocket, Zap, Viper)
- 前端：Vue 3 + Vite, Element Plus, Pinia, ECharts, xterm.js
- 边缘代理：Go 1.26+（仅 Linux，amd64/arm64），gopsutil, Gorilla WebSocket, creack/pty

**关键设计：** 后端通过 `go:embed` 提供 SPA 服务，为单个二进制文件。边缘代理（`proxy/`）是一个独立的 Go 模块，通过 HTTP/WebSocket 与平台通信 —— 独立部署在每个被监控主机上。

[详情 &rarr;](./OpenTraffic-Ops/README_CN.md)

---

### 3. proxy — 边缘代理

部署在每个被监控的边缘主机上。负责系统指标采集和向平台服务器上报，支持 WebSocket 远程控制（终端 / 文件管理）。

**关键特性：**
- 系统信息采集（操作系统、CPU、内存、磁盘、MAC 地址）
- 3 秒周期指标上报（CPU / 内存 / 磁盘 / 网络 / 负载）
- 进程监控（运行状态、CPU%、内存使用量）
- 指令执行（startProcess / stopProcess / restartProcess）
- WebSocket 长连接（自动重连、指数退避、心跳保活）
- 远程终端（持久 PTY Shell，5 分钟超时）
- 远程文件管理（路径安全校验）

**平台支持：** 仅 Linux x86_64 (amd64) 和 Linux ARM64 (aarch64)。Windows 和 macOS 仅用于交叉编译。

[详情 &rarr;](./OpenTraffic-Ops/proxy/README_CN.md)

---

## 子系统之间的关系

```
+-----------------------------+      部署      +-------------------------+
| OpenTraffic-Ops-Init        | ------------> | OpenTraffic-Ops         |
| (本机)                       |  SSH/SFTP      | (远程 Linux 服务器)      |
|                             |                |                         |
| - Docker 管理               |      部署      | - 主机监控              |
| - SSH 配置                  | ------------> | - 告警                  |
| - 二进制部署                |                | - 远程运维              |
+-----------------------------+                +-------------+-----------+
                                                             |
                                                             | HTTP / WebSocket
                                                             |
                                                  +----------v----------+
                                                  | proxy               |
                                                  | (每个边缘主机)       |
                                                  | - 指标采集          |
                                                  | - 远程终端          |
                                                  | - 文件操作          |
                                                  +---------------------+
```

1. **`OpenTraffic-Ops-Initialization`** 是你的控制平面 —— 在本地机器或堡垒主机上运行。它管理 Docker 容器（PostgreSQL、Redis）并将监控栈部署到远程服务器。

2. **`OpenTraffic-Ops`** 作为服务器运行在中心节点或边缘节点上。它收集指标、触发告警，并为运维人员提供 Web UI。

3. **`proxy`** 运行在你想要监控的每台主机上。它每 3 秒上报一次指标，并接受来自平台的远程命令（终端、文件、进程）。

---

## 快速开始

### 前置要求

- Go 1.25+（Proxy 构建需要 Go 1.26+）
- Node.js 18+
- Docker & Docker Compose（用于 `OpenTraffic-Ops-Initialization` 容器管理）
- PostgreSQL 15+（用于 `OpenTraffic-Ops`）
- Redis 7+（推荐两个实例：平台和边缘分离）

### 启动部署面板

```bash
cd OpenTraffic-Ops-Initialization/backend
go mod download
go run cmd/server/main.go
# 服务运行在 http://localhost:8080
```

### 启动监控平台

```bash
# 1. 创建 PostgreSQL 数据库
psql -c "CREATE DATABASE rtm WITH ENCODING = 'UTF8';"

# 2. 导入 DDL
cd OpenTraffic-Ops
psql -d rtm -f sql/01_sys_tables.sql
psql -d rtm -f sql/03_bu_tables.sql
psql -d rtm -f sql/alarm/01_alarm_tables.sql
psql -d rtm -f sql/chat/01_chat_tables.sql

# 3. 启动后端
cd backend
go mod download
go run cmd/server/main.go
# 服务运行在 http://localhost:18084

# 4. 启动前端（开发模式）
cd ../frontend
npm install
npm run dev
# 开发服务器运行在 http://localhost:80
```

两个系统的默认凭据：`admin` / `admin123`

### 构建生产二进制文件（Windows 主机交叉编译到 Linux）

```bash
# 监控平台（后端 + 内嵌前端）
cd OpenTraffic-Ops
build-opentraffic-ops.bat
# 输出：backend/opentraffic-ops-linux-amd64, backend/opentraffic-ops-linux-arm64

# 边缘代理
cd proxy
build-opentraffic-ops-proxy.bat
# 输出：proxy/dist/opentraffic-ops-proxy-linux-amd64, proxy/dist/opentraffic-ops-proxy-linux-arm64

# 部署面板
cd ../../OpenTraffic-Ops-Initialization
build-opentraffic-ops-initialization.bat
# 输出：backend/opentraffic-ops-init-linux-amd64, backend/opentraffic-ops-init-linux-arm64
```

---

## 项目结构

```
opentraffic-ops/
├── OpenTraffic-Ops-Initialization/  # 部署面板
│   ├── backend/                     # Go 后端（Gin, SQLite, Docker SDK）
│   ├── frontend/                    # Vue 3 + TypeScript SPA
│   ├── components/                  # Docker Compose 模板
│   ├── docker-compose.yaml
│   └── README.md                    # 详情文档
│
├── OpenTraffic-Ops/                 # 监控与运维平台
│   ├── backend/                     # Go 后端（Gin, GORM, PostgreSQL, Redis）
│   ├── frontend/                    # Vue 3 SPA
│   ├── proxy/                       # 边缘代理（仅 Linux，独立 Go 模块）
│   ├── sql/                         # PostgreSQL DDL
│   ├── docs/                        # 设计与部署指南
│   └── README.md                    # 详情文档
│
├── README.md                        # 本文件
├── .gitignore                       # 根目录合并的忽略规则
└── LICENSE                          # Apache 许可证
```

---

## 文档

- [OpenTraffic-Ops-Initialization README](./OpenTraffic-Ops-Initialization/README_CN.md) — 部署面板详情
- [OpenTraffic-Ops README](./OpenTraffic-Ops/README_CN.md) — 监控平台详情
- [Proxy README](./OpenTraffic-Ops/proxy/README_CN.md) — 边缘代理部署指南

---

## 安全特性

- JWT Token 认证与自动续签
- RSA 密码加密传输
- XSS 过滤中间件
- 重放攻击防护
- 登录失败锁定
- 参数化 SQL 查询（GORM）
- CORS 控制
- 远程文件路径遍历防护
- SSH 凭据 AES-GCM 加密（在部署面板中）

---

## 许可证

[MIT License](./LICENSE)
