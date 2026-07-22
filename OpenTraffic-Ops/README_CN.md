# OpenTraffic Ops 监控运维平台

<p align="center">
  <a href="../LICENSE"><img src="https://img.shields.io/badge/License-Apache_2.0-blue.svg" alt="License: Apache 2.0"></a>
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

## 📑 目录

- [📖 项目简介](#项目简介)
- [🔧 技术栈](#技术栈)
  - [后端](#后端backend-go-module-opentraffic-ops-backend)
  - [前端](#前端frontend-vue3-spa)
  - [边缘端 Proxy](#边缘端-proxyproxy-go-module-opentraffic-ops-proxy独立交付件)
- [✨ 功能描述](#功能描述)
  - [👤 系统管理](#系统管理)
  - [🖥️ 主机管理](#主机管理)
  - [🔔 监控与告警](#监控与告警)
  - [🤖 Agent 对话](#agent-对话)
  - [🔧 远程运维](#远程运维)
  - [📝 系统日志](#系统日志)
  - [📡 边缘端 Proxy 功能](#边缘端-proxy-功能)
  - [🔌 Proxy 协议接口](#proxy-协议接口公开无需认证)
- [🚀 快速开始](#快速开始)
  - [📋 前置要求](#前置要求)
  - [💻 开发模式启动](#开发模式启动)
- [🖥️ 服务器部署](#服务器部署)
  - [📦 生产构建](#生产构建单包自包含)
  - [⚙️ 配置说明](#配置说明)
  - [📝 日志说明](#日志说明)
- [❓ 常见问题](#常见问题)
- [🙏 致谢](#致谢)

---

## 📖 项目简介

OpenTraffic Ops 监控运维平台是一个面向边缘计算场景的全栈监控管理系统，由**监控平台服务**（后端内嵌前端，通过 `go:embed` 单二进制部署）和**边缘端 Proxy** 两个独立交付件组成，支持主机管理、健康度采集、阈值告警、远程运维（终端 / 文件 / 进程）、Agent 对话（控制 Agent / 感知 Agent）等能力。

> **命名说明**：**边缘端 Proxy** 指部署在被监控主机上的采集/控制程序（即 `proxy/` 目录交付件）。**系统功能中的 Agent 对话**分为控制 Agent 和感知 Agent 两种类型，是平台对接外部 Agent 的业务模块。二者职责不同，下文严格区分使用 **Proxy** 与 **Agent** 两个术语。

---

## 🔧 技术栈

### 后端（`backend/`，Go module `opentraffic-ops-backend`）

| 技术         | 版本     | 说明                |
| ---------- | ------ | ----------------- |
| Go         | 1.25+  | 编程语言              |
| Gin        | v1.10  | Web 框架            |
| GORM       | v1.25  | ORM 框架            |
| PostgreSQL | 15+    | 主数据库              |
| Redis      | 7+     | 平台缓存 / 边缘消息（双实例） |
| JWT v5     | v5.3   | 认证授权              |
| Gorilla WS | v1.5   | WebSocket（终端、文件）  |
| Zap        | v1.27  | 日志框架              |
| Viper      | v1.19  | 配置管理              |

### 前端（`frontend/`，Vue3 SPA）

| 技术           | 版本   | 说明        |
| ------------ | ---- | --------- |
| Vue          | 3.3  | 前端框架      |
| Vite         | 5.x  | 构建工具      |
| Element Plus | 2.8  | UI 组件库    |
| Pinia        | 2.1  | 状态管理      |
| ECharts      | 5.4  | 数据可视化     |
| Axios        | 1.7  | HTTP 客户端  |
| xterm.js     | 5.3  | 浏览器端终端    |

### 边缘端 Proxy（`proxy/`，Go module `opentraffic-ops-proxy`，独立交付件）

| 技术          | 版本     | 说明                 |
| ----------- | ------ | ------------------ |
| Go          | 1.26+  | 编程语言（**仅 Linux**）  |
| gopsutil    | v3     | 主机指标采集             |
| Gorilla WS  | v1.5   | 与平台的 WebSocket 长连接 |
| creack/pty  | v1.1   | 远程终端 PTY 实现        |

> Proxy 与后端不共享代码，通过 HTTP/WS 协议交互；只能运行在 Linux（amd64 / arm64）上，Windows 仅用于交叉编译。

---

## ✨ 功能描述

### 👤 系统管理
- **用户管理** —— 用户增删改查、密码策略、登录失败锁定
- **个人中心** —— 用户信息维护、密码修改、头像上传

![用户管理](images/image-1.png)

### 🖥️ 主机管理
- **主机信息** —— 边缘节点主机的注册、CRUD 与状态展示（Proxy 首次注册自动入库）
- **主机健康度** —— 主机历史健康度数据采集与查询（自动按日轮转、保留 7 天）
- **主机运维** —— 远程运维操作汇总入口（终端、文件、进程控制）

![主机管理](images/image.png)

### 🔔 监控与告警
- **告警通道** —— 支持邮件、钉钉、企业微信、平台内部四类通知渠道，可配置多个通道
- **告警规则** —— 多维度规则编排：
  - 指标类：CPU / 内存 / 磁盘 / 网络 / 负载
  - 服务类：主机离线、控制 Agent 离线
- **告警记录** —— 历史告警查询、确认与恢复追踪
- **告警通知日志** —— 各通道发送状态的详细记录
- **内置调度器**（无外部 cron 依赖）：
  - `dealOffline`（60s）—— 主机离线检测
  - `alarmCheck`（30s）—— 告警检测
  - `cleanHostHealth`（每日 03:30）—— 清理 7 天前的健康度数据

![告警规则](images/image-2.png)
![告警通道](images/image-3.png)
![告警记录](images/image-4.png)

### 🤖 Agent 对话
- **控制 Agent 对话** —— 通过对话方式与控制 Agent 交互，执行进程启停、参数下发等操作
- **感知 Agent 对话** —— 通过对话方式与感知 Agent 交互，获取主机在线状态与基础信息
- **会话管理** —— 会话创建、列表分页、消息历史、重命名、删除

![Agent 对话](images/image-5.png)
![Agent 会话](images/image-6.png)

### 🔧 远程运维
- **远程终端** —— 浏览器内 xterm 终端，经平台 WebSocket Hub 直达 Proxy PTY（支持颜色、resize）
- **远程文件** —— Proxy 所在主机文件浏览、读取、编辑、上传、下载、删除、创建目录（单文件 10MB 限制，目录遍历防护）
- **进程控制** —— 通过平台向 Proxy 下发进程启动 / 停止 / 重启指令

![远程终端](images/image-8.png)
![远程文件](images/image-9.png)
![进程控制](images/image-10.png)

### 📝 系统日志
- **操作日志** —— 通过 `OperLog` 中间件自动记录受保护接口的操作行为
- **登录日志** —— 登录成功 / 失败记录

![系统日志](images/image-11.png)

### 📡 边缘端 Proxy 功能
- **系统信息采集** —— 注册时上报 OS 类型/版本、CPU 架构/核数/型号、内存、磁盘、MAC 地址
- **系统指标采集** —— 3 秒周期上报 CPU / 内存 / 磁盘 / 网络 / 负载
- **进程监控** —— 采集配置进程的运行状态、CPU 使用率、内存使用
- **指令执行** —— 接收平台下发的 `startProcess` / `stopProcess` / `restartProcess` 指令
- **WebSocket 长连接** —— 自动重连（指数退避）、心跳保活、读写 goroutine 安全退出
- **远程终端** —— 基于 PTY 的持久 Shell 会话（5 分钟超时自动关闭）
- **远程文件管理** —— 完整的文件操作能力，支持路径安全校验

### 🔌 Proxy 协议接口（公开，无需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/proxy/register` | Proxy 首次注册，上报硬件信息 |
| POST | `/api/v1/proxy/heartbeat` | 心跳保活 + 监控数据上报（3s 周期） |
| POST | `/api/v1/proxy/poll` | 轮询待执行指令（进程启停） |
| POST | `/api/v1/proxy/ack` | 指令执行结果上报 |
| GET | `/api/v1/proxy/ws?ip=<host>` | WebSocket 长连接（终端/文件） |

---

## 🚀 快速开始

### 📋 前置要求

- Go 1.25+ （Proxy 构建额外需要 Go 1.26+）
- Node.js 18+
- PostgreSQL 15+
- Redis 7+（推荐准备 **两个实例 / 两个 db**：平台与边缘分离）

### 💻 开发模式启动

#### 1. 克隆项目

```bash
git clone <repository-url>
cd OpenTraffic-Ops
```

#### 2. 初始化数据库

```sql
CREATE DATABASE rtm WITH ENCODING = 'UTF8';
```

```bash
psql -d rtm -f sql/01_sys_tables.sql
psql -d rtm -f sql/02_bu_tables.sql
psql -d rtm -f sql/03_chat_tables.sql
psql -d rtm -f sql/04_alarm_tables.sql
```

创建 `~/.opentraffic-ops/opentraffic-ops-config.yaml`（参考 `backend/configs/opentraffic-ops-config.yaml`），修改数据库连接：

```yaml
datasource:
  host: 127.0.0.1
  port: 5432
  database: rtm
  username: postgres
  password: your_password
```

#### 3. 启动后端

```bash
cd backend
go mod download
go run cmd/server/main.go
```

后端服务默认运行在 `http://localhost:18081`。

#### 4. 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端开发服务器默认运行在 `http://localhost:80`，并将 `/dev-api`、`/dev-ws-api` 代理到 `127.0.0.1:18081`。

#### 5. 访问系统

打开浏览器访问 `http://localhost`，默认账号密码：
- 用户名：`admin`
- 密码：`admin123`

Proxy 部署到 Linux 主机，参见 [`proxy/README_CN.md`](proxy/README_CN.md)。

#### Windows 本地开发快速调试（无需每次复制 dist）

开发阶段前端改动频繁，设置环境变量让后端直接从磁盘加载前端资源：

```cmd
# 在 backend 目录下
set RTM_STATIC_DIR=..\frontend\dist
go run cmd\server\main.go
```

生产构建时**不要**设置该变量，确保前端资源被完整嵌入二进制。

---

## 🖥️ 服务器部署

### 📦 生产构建（单包自包含）

#### Windows 交叉编译 Linux 部署包

执行 `build-opentraffic-ops.bat` 生成内嵌前端的 Linux AMD64、ARM64 和 Loong64 二进制：

```cmd
build-opentraffic-ops.bat
```

输出文件：
- `backend\opentraffic-ops-linux-amd64`
- `backend\opentraffic-ops-linux-arm64`
- `backend\opentraffic-ops-linux-loong64`

上传至 Linux 服务器并运行：

```bash
mkdir -p ~/.opentraffic-ops
cp backend/configs/opentraffic-ops-config.yaml ~/.opentraffic-ops/opentraffic-ops-config.yaml
# 编辑 ~/.opentraffic-ops/opentraffic-ops-config.yaml 修改生产环境配置

chmod +x opentraffic-ops-linux-amd64
./opentraffic-ops-linux-amd64
```

#### Proxy 交叉编译

```batch
cd proxy
build-opentraffic-ops-proxy.bat
```

输出文件：
- `proxy/dist/opentraffic-ops-proxy-linux-amd64`
- `proxy/dist/opentraffic-ops-proxy-linux-arm64`
- `proxy/dist/opentraffic-ops-proxy-linux-loong64`

> Proxy 仅支持 Linux 运行；Windows / macOS 仅用作构建主机。

### ⚙️ 配置说明

后端使用单一配置文件 `opentraffic-ops-config.yaml`，固定从 `~/.opentraffic-ops/opentraffic-ops-config.yaml` 加载，开发和生产环境共用。

首次运行前，创建配置文件（参考 `backend/configs/opentraffic-ops-config.yaml`）：

```bash
# Linux / macOS
mkdir -p ~/.opentraffic-ops
cp backend/configs/opentraffic-ops-config.yaml ~/.opentraffic-ops/opentraffic-ops-config.yaml

# Windows
mkdir %USERPROFILE%\.opentraffic-ops
copy backend\configs\opentraffic-ops-config.yaml %USERPROFILE%\.opentraffic-ops\opentraffic-ops-config.yaml
```

任意 Key 都可以通过 `RTM_` 前缀的环境变量覆盖（`.` → `_`）：

```bash
export RTM_DATASOURCE_HOST=192.168.1.100
export RTM_DATASOURCE_PASSWORD=secret
```

#### 关键配置项

```yaml
server:
  port: 18081
  mode: release        # debug / test / release

datasource:
  driver: postgres
  host: 127.0.0.1
  port: 5432
  database: rtm
  username: postgres
  password: ***

redis:
  platform:            # 会话、验证码、登录锁、在线用户
    host: 127.0.0.1
    port: 6379
    db: 3
  edge:                # 监控数据 / Proxy 消息队列
    host: 127.0.0.1
    port: 6379
    db: 1

jwt:
  header: Authorization
  secret: ***
  expireTime: 480      # 分钟

agent:
  control: ""          # 控制 Agent 外部 API 地址
  perceive: ""         # 感知 Agent 外部 API 地址
```

> 平台与边缘两个 Redis 角色必须分开配置（可以是同一物理实例的不同 db，也可以是两套实例）。
> Agent 配置用于对接外部 Agent 服务，为空时对应功能不可用。

### 📝 日志说明

日志通过 Zap 输出，默认写入 `logs/` 目录，按大小 / 天数轮转：

```
logs/
├── opentraffic-ops-backend.log
└── opentraffic-ops-backend-*.log
```

`opentraffic-ops-config.yaml` 的 `log` 块中可配置日志级别、文件名、单文件大小、保留份数、保留天数、是否压缩：

```yaml
log:
  level: info
  filename: logs/opentraffic-ops-backend.log
  maxSize: 100          # MB
  maxBackups: 30
  maxAge: 30            # 天
  compress: true
```

---

## ❓ 常见问题

### 数据库连接失败 / 迁移错误
- 确认 PostgreSQL 正在运行且可访问
- 检查 `~/.opentraffic-ops/opentraffic-ops-config.yaml` 中的 `datasource` 配置
- 确保按正确顺序执行了 `sql/` 目录下的 DDL 脚本

### Redis 连接失败
- 确认 Redis 实例正在运行
- 检查 `redis.platform` 和 `redis.edge` 配置
- 平台与边缘 Redis 可以是同一实例的不同 db

### 前端提示无法连接后端
- 确认后端在 18081 端口正常运行
- 检查 `frontend/vite.config.js` 中的 Vite 代理配置
- 生产环境确保 `opentraffic-ops-config.yaml` 中的 `server.port` 与预期端口一致

### WebSocket 终端无法连接
- 检查目标 Proxy 是否在线（平台中的主机状态）
- 确认 token 通过查询参数正确传递
- 检查 WebSocket 端口的防火墙规则

### 告警通知未发送
- 检查告警通道配置（邮件服务器、钉钉 webhook 等）
- 确认告警规则阈值设置正确且已启用
- 查看通知日志获取发送失败原因

### Proxy 未向平台注册
- 检查 Proxy `opentraffic-ops-proxy-config.json` 中的 `platformUrl` 是否指向正确的平台地址
- 确认 Proxy 主机与平台之间的网络连通性
- 确保平台的 `/api/v1/proxy/register` 接口可达

---

## 🙏 致谢

OpenTraffic Ops 基于以下开源项目构建：

- [Go](https://golang.org/) / [Gin](https://github.com/gin-gonic/gin) / [GORM](https://gorm.io/) —— 后端框架与 ORM
- [Vue.js](https://vuejs.org/) / [Vite](https://vitejs.dev/) —— 前端框架与构建工具
- [Element Plus](https://element-plus.org/) —— UI 组件库
- [PostgreSQL](https://www.postgresql.org/) —— 主数据库
- [Redis](https://redis.io/) —— 缓存与消息
- [Gorilla WebSocket](https://github.com/gorilla/websocket) —— WebSocket 实现
- [Zap](https://github.com/uber-go/zap) —— 日志框架
- [xterm.js](https://xtermjs.org/) —— 浏览器端终端

[Apache License 2.0](../LICENSE)
