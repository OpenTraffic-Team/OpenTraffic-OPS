# RTM 监控运维平台

RTM（Real-Time Monitor）监控运维平台是一个面向边缘计算场景的全栈监控管理系统，由 **后端服务**、**前端管理台** 和 **边缘端 Proxy** 三个独立交付件组成，支持主机管理、健康度采集、阈值告警、远程运维（终端 / 文件 / 进程）、Agent 控制与对话等能力。

> 命名说明：**边缘端 Proxy** 指部署在被监控主机上的采集/控制程序（即 `proxy/` 目录交付件），**系统功能中的 Agent**（Agent 感知 / Agent 控制 / Agent 对话）是平台对接外部 Agent 的业务模块。二者职责不同，下文严格区分使用 **Proxy** 与 **Agent** 两个术语。

## 技术架构

### 后端技术栈（`backend/`，Go module `rtm-server`）

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

### 前端技术栈（`frontend/`，Vue3 SPA）

| 技术           | 版本   | 说明        |
| ------------ | ---- | --------- |
| Vue          | 3.3  | 前端框架      |
| Vite         | 5.x  | 构建工具      |
| Element Plus | 2.8  | UI 组件库    |
| Pinia        | 2.1  | 状态管理      |
| ECharts      | 5.4  | 数据可视化     |
| Axios        | 1.7  | HTTP 客户端  |
| xterm.js     | 5.3  | 浏览器端终端    |

### 边缘端 Proxy 技术栈（`proxy/`，Go module `rtm-proxy`，独立交付件）

| 技术          | 版本     | 说明                 |
| ----------- | ------ | ------------------ |
| Go          | 1.26+  | 编程语言（**仅 Linux**）  |
| gopsutil    | v3     | 主机指标采集             |
| Gorilla WS  | v1.5   | 与平台的 WebSocket 长连接 |
| creack/pty  | v1.1   | 远程终端 PTY 实现        |

> Proxy 与后端不共享代码，通过 HTTP/WS 协议交互；只能运行在 Linux（amd64 / arm64）上，Windows 仅用于交叉编译。

## 项目结构

```
rtm-monitor-platform-go/
├── backend/                        # Go 后端服务（module: rtm-server）
│   ├── cmd/server/main.go          # 主程序入口
│   ├── internal/
│   │   ├── config/                 # 配置加载（Viper + RTM_ 环境变量覆盖）
│   │   ├── constant/               # 状态码、Redis Key 前缀等常量
│   │   ├── dto/                    # 请求/响应 DTO
│   │   ├── handler/                # Gin Handler（含 Agent 代理、聊天会话）
│   │   ├── middleware/             # JWT、XSS、CORS、Recovery、OperLog、Replay、WSAuth
│   │   ├── model/                  # GORM 数据模型（含告警、聊天会话）
│   │   ├── repository/             # 数据访问层
│   │   ├── router/                 # 路由集中注册（router.go）
│   │   ├── service/                # 业务逻辑层 + 内置调度器 + 告警引擎
│   │   ├── ws/                     # WebSocket Hub（前端 ↔ Proxy 桥接）
│   │   └── utils/                  # 工具函数
│   ├── pkg/
│   │   ├── cache/                  # Redis 封装
│   │   ├── captcha/                # 图形/算术验证码
│   │   ├── crypto/                 # RSA 等加密工具
│   │   ├── jwt/                    # JWT 工具
│   │   ├── response/               # 统一响应封装
│   │   └── static/                 # 前端嵌入式静态资源（go:embed all:dist）
│   └── configs/                    # 参考配置模板
│       └── config.yaml
├── frontend/                       # Vue3 + Vite 管理台
│   ├── src/
│   │   ├── api/                    # 按模块分组的 axios 封装
│   │   │   ├── business/           # 主机、健康度、告警
│   │   │   ├── control-agent/      # 控制 Agent 对话
│   │   │   ├── perceive-agent/     # 感知 Agent 对话
│   │   │   ├── remote/             # 终端、文件
│   │   │   ├── system/             # 用户
│   │   │   └── monitor/            # 操作日志、登录日志
│   │   ├── assets/  components/  directive/  layout/
│   │   ├── router/                 # 前端路由（静态业务路由）
│   │   ├── store/                  # Pinia 状态管理
│   │   ├── utils/                  # 工具函数（含 jsencrypt 等）
│   │   └── views/                  # 页面（system / monitor / business）
│   ├── package.json
│   └── vite.config.js
├── proxy/                          # 边缘端 Proxy（module: rtm-proxy，仅 Linux 运行）
│   ├── main.go                     # Proxy 入口（心跳、轮询、WS 客户端）
│   ├── client/                     # HTTP 客户端（注册、心跳、轮询、ACK）
│   ├── collector/                  # 系统/进程指标采集
│   ├── config/                     # Proxy 配置（JSON）
│   ├── executor/                   # 进程启停执行器 + Shell PTY
│   ├── filemanager/                # 远程文件管理（目录遍历防护）
│   ├── wsclient/                   # WebSocket 客户端（自动重连）
│   ├── build-proxy.ps1             # Windows 上交叉编译 → dist/
│   └── README.md
├── sql/                            # PostgreSQL DDL
│   ├── 01_sys_tables.sql           # 系统表（用户、操作日志、登录日志）
│   ├── 03_bu_tables.sql            # 业务表（主机信息、主机健康度）
│   ├── alarm/01_alarm_tables.sql   # 告警通道 / 规则 / 记录 / 通知日志
│   └── chat/01_chat_tables.sql     # Agent 对话会话与消息
├── docs/                           # 中文设计与部署文档
│   ├── 开发环境搭建指南.md
│   ├── 生产环境部署指南.md
│   └── Proxy部署与使用指南.md
├── build-linux.bat                 # Windows 主机交叉编译后端 → Linux 二进制
└── logs/                           # 运行时日志
```

## 功能模块

### 系统管理

- **用户管理** —— 用户增删改查、密码策略、登录失败锁定
- **个人中心** —— 用户信息维护、密码修改、头像上传

### 主机管理

- **主机信息** —— 边缘节点主机的注册、CRUD 与状态展示（Proxy 首次注册自动入库）
- **主机健康度** —— 主机历史健康度数据采集与查询（自动按日轮转、保留 7 天）
- **主机运维** —— 远程运维操作汇总入口（终端、文件、进程控制）

### 监控与告警

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

### Agent 管理（业务侧 Agent）

- **控制 Agent** —— 通过对话方式与控制 Agent 交互，执行进程启停、参数下发等控制操作
- **感知 Agent** —— 通过对话方式与感知 Agent 交互，获取主机在线状态与基础信息
- **Agent 对话会话** —— 会话创建、列表分页、消息历史、重命名、删除

### 远程运维

- **远程终端** —— 浏览器内 xterm 终端，经平台 WebSocket Hub 直达 Proxy PTY（支持颜色、resize）
- **远程文件** —— Proxy 所在主机文件浏览、读取、编辑、上传、下载、删除、创建目录（单文件 10MB 限制，目录遍历防护）
- **进程控制** —— 通过平台向 Proxy 下发进程启动 / 停止 / 重启指令

### 系统日志

- **操作日志** —— 通过 `OperLog` 中间件自动记录受保护接口的操作行为
- **登录日志** —— 登录成功 / 失败记录

### 边缘端 Proxy 功能

- **系统信息采集** —— 注册时上报 OS 类型/版本、CPU 架构/核数/型号、内存、磁盘、MAC 地址
- **系统指标采集** —— 3 秒周期上报 CPU / 内存 / 磁盘 / 网络 / 负载
- **进程监控** —— 采集配置进程的运行状态、CPU 使用率、内存使用
- **指令执行** —— 接收平台下发的 `startProcess` / `stopProcess` / `restartProcess` 指令
- **WebSocket 长连接** —— 自动重连（指数退避）、心跳保活、读写 goroutine 安全退出
- **远程终端** —— 基于 PTY 的持久 Shell 会话（5 分钟超时自动关闭）
- **远程文件管理** —— 完整的文件操作能力，支持路径安全校验

## 快速开始

### 环境要求

- Go 1.25+ （Proxy 构建额外需要 Go 1.26+）
- Node.js 18+
- PostgreSQL 15+
- Redis 7+（推荐准备 **两个实例 / 两个 db**：平台与边缘分离）

### 1. 克隆项目

```bash
git clone <repository-url>
cd rtm-monitor-platform-go
```

### 2. 初始化数据库

创建 PostgreSQL 数据库（默认库名 `rtm`）：

```sql
CREATE DATABASE rtm WITH ENCODING = 'UTF8';
```

按顺序导入 `sql/` 目录下的 DDL：

```bash
psql -d rtm -f sql/01_sys_tables.sql
psql -d rtm -f sql/03_bu_tables.sql
psql -d rtm -f sql/alarm/01_alarm_tables.sql
psql -d rtm -f sql/chat/01_chat_tables.sql
```

在 `~/.rtm-monitor-platform/` 目录下创建 `config.yaml`（可参考 `backend/configs/config.yaml`），修改数据库连接配置：

```yaml
datasource:
  host: 127.0.0.1
  port: 5432
  database: rtm
  username: postgres
  password: your_password
```

### 3. 启动后端（开发模式）

```bash
cd backend
go mod download
go run cmd/server/main.go
```

后端服务默认运行在 `http://localhost:18084`。

### 4. 启动前端（开发模式）

```bash
cd frontend
npm install
npm run dev
```

前端开发服务器默认运行在 `http://localhost:80`，并将 `/dev-api`、`/dev-ws-api` 代理到 `127.0.0.1:18084`。

### 5. 访问系统

打开浏览器访问 <http://localhost> ，默认账号密码：

- 用户名：`admin`
- 密码：`admin123`

### 6.（可选）部署 Proxy 到 Linux 主机

参见 [`proxy/README.md`](proxy/README.md) 与 [`docs/Proxy部署与使用指南.md`](docs/Proxy部署与使用指南.md)。

## 构建部署

### Linux 交叉编译（后端 + 内嵌前端）

在 Windows 开发机上一键交叉编译后端，并把前端打包嵌入到二进制中：

```bash
build-linux.bat
```

脚本执行流程：

1. 清理 `backend/pkg/static/dist/` 与历史构建产物；
2. 在 `frontend/` 中执行 `npm install && npm run build:prod`；
3. 将 `frontend/dist/*` 复制到 `backend/pkg/static/dist/`；
4. 以 `GOOS=linux CGO_ENABLED=0` 同时构建 amd64 与 arm64 二进制。

构建产物直接输出到 `backend/` 目录：

```
backend/
├── rtm-monitor-platform-linux-amd64   # AMD64 二进制（前端已嵌入）
├── rtm-monitor-platform-linux-arm64   # ARM64 二进制（前端已嵌入）
└── configs/
    └── config.yaml                    # 参考配置模板
```

二进制自带前端静态资源（`go:embed`），无需额外部署 Nginx。Linux 服务器上需先将配置文件放到固定路径，然后直接启动：

```bash
mkdir -p ~/.rtm-monitor-platform
cp backend/configs/config.yaml ~/.rtm-monitor-platform/config.yaml
# 编辑 ~/.rtm-monitor-platform/config.yaml 修改生产环境配置

chmod +x rtm-monitor-platform-linux-amd64
./rtm-monitor-platform-linux-amd64
```

### Proxy 交叉编译

```powershell
cd proxy
.\build-proxy.ps1                  # 默认输出到 proxy/dist/
.\build-proxy.ps1 -Version "1.1.0"
```

产物：

- `proxy/dist/rtm-proxy-linux-amd64`
- `proxy/dist/rtm-proxy-linux-arm64`

> Proxy 仅支持 Linux 运行；Windows / macOS 仅用作构建主机。

### 前端单独构建

```bash
cd frontend
npm run build:prod    # 生产环境
npm run build:stage   # 测试环境
```

## 配置文件说明

后端使用单一配置文件 `config.yaml`，固定从 `~/.rtm-monitor-platform/config.yaml` 加载，开发和生产环境共用。

首次运行前，在对应用户目录下创建配置文件（可参考 `backend/configs/config.yaml`）：

```bash
# Linux / macOS
mkdir -p ~/.rtm-monitor-platform
cp backend/configs/config.yaml ~/.rtm-monitor-platform/config.yaml

# Windows
mkdir %USERPROFILE%\.rtm-monitor-platform
copy backend\configs\config.yaml %USERPROFILE%\.rtm-monitor-platform\config.yaml
```

任意 Key 都可以通过 `RTM_` 前缀的环境变量覆盖（`.` → `_`）：

```bash
export RTM_DATASOURCE_HOST=192.168.1.100
export RTM_DATASOURCE_PASSWORD=secret
export RTM_REDIS_PLATFORM_PASSWORD=***
export RTM_REDIS_EDGE_HOST=192.168.1.101
```

### 关键配置项

```yaml
server:
  port: 18084          # HTTP / WebSocket 端口（前端 vite 代理固定到该端口）
  mode: release        # 运行模式: debug/test/release

datasource:
  driver: postgres
  host: 127.0.0.1
  port: 5432
  database: rtm
  username: postgres
  password: ***

redis:
  platform:            # 平台 Redis：会话、验证码、登录锁、在线用户
    host: 127.0.0.1
    port: 6379
    db: 3
  edge:                # 边缘 Redis：监控数据 / Proxy 指令队列
    host: 127.0.0.1
    port: 6379
    db: 1

jwt:
  header: Authorization
  secret: ***
  expireTime: 480      # Token 过期时间（分钟）

agent:
  control: ""          # 控制 Agent 外部 API 地址
  perceive: ""         # 感知 Agent 外部 API 地址
```

> 平台与边缘两个 Redis 角色必须分开配置（可以是同一物理实例的不同 db，也可以是两套实例）。
> Agent 配置用于对接外部 Agent 服务，为空时对应功能不可用。

### 开发热重载（前后端分离）

通过设置环境变量让后端从磁盘读取前端文件，跳过 `go:embed`，避免每次改前端就重新构建：

```bash
# Windows
cd backend
set RTM_STATIC_DIR=..\frontend\dist
go run cmd\server\main.go
```

详见 `backend/pkg/static/static.go` 的开发 / 生产切换逻辑。

## 后端架构

标准分层结构。`cmd/server/main.go` 完成依赖装配，`internal/router/router.go` 是**唯一**的路由注册中心，新增 handler 必须在这里挂载（无自动发现）。

```
handler   →  service  →  repository  →  model (GORM)
   ↑           ↑
  dto      (业务逻辑，可调用多个 repo)
```

- 公开路由组 `public`：登录、获取公钥、Proxy 上报（`/api/v1/proxy/*`）等。
- 鉴权路由组 `auth`：经过 `middleware.JWTAuth()` 校验后挂载所有业务接口。
- WebSocket：前端终端 `/ws/terminal`（`WSAuth` 校验查询参数中的 token）；Proxy 长连 `/api/v1/proxy/ws`（按网络可达性放行，无 JWT）。
- WebSocket Hub（`internal/ws/hub.go`）作为前端会话与 Proxy 连接之间的桥接，承担远程终端透传与远程文件操作。
- 调度器（`internal/service/scheduler.go`）由 `main.go` 启动，承载离线检测、告警检测、健康度清理三类内置任务。
- 告警引擎（`internal/service/alarm_engine.go`）每 30 秒检查一次告警规则，支持阈值突破持续时间判断与自动恢复。

### 标准响应

所有 HTTP 响应都通过 `pkg/response` 输出，HTTP 状态固定为 200，业务真实状态在 `code` 字段中：

```go
response.Success(c, data)                 // 200 / "操作成功"
response.SuccessWithMsg(c, msg, data)
response.SuccessPage(c, total, rows)      // {code, msg, data: {total, rows}}
response.Error(c, msg)                    // 500
response.Unauthorized(c, msg)             // 401（JWT 中间件使用）
response.Forbidden(c, msg)                // 403
```

前端拦截器以 `code`（200 / 401 / 403 / 500 / 601）判断业务状态。

### 认证流程

1. `GET /getPublicKey` 获取 RSA 公钥；
2. 前端用公钥加密密码后 `POST /login`；
3. 后端签发 JWT，前端在 `Authorization: Bearer <token>` 中携带；
4. `JWTAuth` 中间件解析 token，从平台 Redis 的 `login_tokens:<uuid>` 读取 `LoginUser`，并把 `userId` / `username` / `uuid` / `claims` 注入 Gin Context（通过 `middleware.GetUserID(c)` 等访问）；
5. `GetInfo` 在 `loginUser.NeedRefresh()` 为真时自动续签。

## 前端架构

- `src/api/` 按域分组：`system/`、`monitor/`、`business/`、`remote/`、`control-agent/`、`perceive-agent/`，以及 `login.js`、`menu.js`。
- `src/views/` 与 API 分组对齐：`business/host-info/`、`business/alarm-config/`、`business/remote-terminal/`、`business/agent-control/`、`business/agent-perceive/` 等。
- 登录流程（`src/store/modules/user.js`）：拉取公钥 → `utils/jsencrypt.js` 加密密码 → `login()`，统一按 `{code, msg, data}` 解包响应。
- 路径别名：`@` → `src/`，`~` → 项目根（见 `vite.config.js`）。
- 开发期 API 基地址 `/dev-api`，WebSocket 基地址 `/dev-ws-api`，均代理到 `127.0.0.1:18084`。

## 开发指南

### 后端开发规范

- **Handler** —— Gin 请求入口，负责参数校验与响应封装；每个 handler 类型都需提供 `RegisterRoutes(*gin.RouterGroup)`，由 `router.go` 显式调用。
- **Service** —— 业务逻辑层，构造函数接收 `*gorm.DB` 以及（必要时）其它 service。
- **Repository** —— 数据访问层，封装 GORM 查询；不直接对外暴露。
- **DTO / Model** —— `internal/dto/*` 用于对外交互，`internal/model/*` 是 GORM 模型，**不要把 model 直接返回给前端**。
- **常量复用** —— 状态码、Redis Key 前缀、`del_flag`、验证码类型等常量集中在 `internal/constant/constant.go`，禁止硬编码字面量。
- **审计日志** —— 需要写操作日志的 CRUD handler，构造函数应接收 `operLogService`，由 `OperLog` 中间件统一记录。

### 前端开发规范

- API 接口放在 `src/api/`，按模块分组。
- 页面组件放在 `src/views/` 对应模块下。
- 公共组件放在 `src/components/`。
- 状态管理使用 Pinia，模块定义在 `src/store/modules/`。

## API 文档

项目采用 RESTful 设计风格，统一响应格式：

```json
{
  "code": 200,
  "msg": "操作成功",
  "data": {}
}
```

认证方式：请求头 `Authorization: Bearer <token>`；WebSocket 通过查询参数 `?token=<...>` 传递。

### Proxy 协议接口（公开，无需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/proxy/register` | Proxy 首次注册，上报硬件信息 |
| POST | `/api/v1/proxy/heartbeat` | 心跳保活 + 监控数据上报（3s 周期） |
| POST | `/api/v1/proxy/poll` | 轮询待执行指令（进程启停） |
| POST | `/api/v1/proxy/ack` | 指令执行结果上报 |
| GET | `/api/v1/proxy/ws?ip=<host>` | WebSocket 长连接（终端/文件） |

## 安全特性

- JWT Token 认证 + Token 自动续签
- 密码 RSA 加密传输
- XSS 过滤中间件（按 `xss.urlPatterns` / `xss.excludes` 生效）
- 操作防重放（`Replay` 中间件）
- 登录失败锁定（`user.password.maxRetryCount` / `lockTime`）
- SQL 注入防护（GORM 参数化查询）
- CORS 跨域控制
- 远程文件路径安全校验（禁止目录遍历）

## 日志说明

日志通过 Zap 输出，默认写入 `logs/` 目录，按大小 / 天数轮转（由 lumberjack 实现）：

```
logs/
├── rtm-server.log          # 当前日志
└── rtm-server-*.log        # 历史轮转日志
```

日志级别、文件名、单文件大小、保留份数、保留天数、是否压缩等均可在 `config.yaml` 的 `log` 块中配置：

```yaml
log:
  level: info           # debug / info / warn / error
  filename: logs/rtm-server.log
  maxSize: 100          # 单文件 MB
  maxBackups: 30        # 最多保留份数
  maxAge: 30            # 最长保留天数
  compress: true
```

## 文档

更多设计与部署细节请见 `docs/`：

- 开发环境搭建指南
- 生产环境部署指南
- 边缘端 Proxy 部署与使用指南
- 远程主机管理功能设计方案
- 控制 Agent 说明（系统侧 Agent 控制功能）

## 开源协议

[MIT License](LICENSE)
