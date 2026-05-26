# OpenTraffic Ops 部署面板

[English](README.md)

## 项目简介

OpenTraffic Ops 部署面板是一个**单包自包含**的综合运维平台，集成了 **Docker 容器组件管理** 与 **SSH 远程服务器部署** 两大核心能力。平台后端由 Go 提供单一 HTTP 服务，前端通过 `go:embed` 嵌入二进制，**无需额外安装 Nginx 或配置反向代理**——只需运行一个二进制文件即可启动完整服务。

### 核心能力

| 能力 | 说明 |
|------|------|
| Docker 组件管理 | 一键安装、启动、停止、卸载常用中间件（PostgreSQL、Redis），支持自定义端口、环境变量、数据卷和启动命令 |
| 实时监控 | 查看组件实时资源占用（CPU / 内存 / 网络 / 磁盘），支持日志实时刷新和自动刷新 |
| SSH 服务器管理 | 统一管理多台远程 Linux 服务器的 SSH 连接配置，支持密码和密钥两种认证方式 |
| 远程二进制部署 | 通过 SSH/SFTP 将 opentraffic-ops-proxy 和 opentraffic-ops 二进制文件一键部署到远程服务器 |
| 远程配置管理 | 在线查看和编辑远程服务器上的软件配置文件（opentraffic-ops-proxy 的 config.json、opentraffic-ops 的 config.yaml） |
| 远程服务管理 | 通过 PID 文件管理远程服务的启动、停止、重启，无需 root 权限 |
| 部署记录追溯 | 完整记录每次远程部署的操作日志、执行结果和历史记录 |

## 技术栈

### 后端
- **语言**: Go 1.21+
- **框架**: Gin
- **数据库**: SQLite（零外部依赖）
- **容器管理**: Docker SDK for Go
- **SSH 客户端**: 基于 `golang.org/x/crypto/ssh`
- **认证**: JWT
- **静态文件托管**: `go:embed` + 自定义 SPA Fallback Handler
- **加密存储**: AES-GCM 加密敏感信息（SSH 密码、私钥）

### 前端
- **框架**: Vue 3 + TypeScript
- **UI 组件**: Element Plus
- **构建工具**: Vite
- **状态管理**: Pinia
- **图表**: ECharts
- **路由模式**: `createWebHistory`（History 模式）

## 功能模块

### 1. 监控大屏
- 组件统计卡片（总数 / 运行中 / 已停止 / 错误）
- 服务器统计卡片（总数 / 密码认证 / 密钥认证 / 已配置部署）
- 组件类型分布饼图
- 组件状态分布柱状图
- 组件实时监控表格（CPU / 内存 / 网络 IO），支持启停实时刷新

### 2. 组件管理
- 组件目录浏览，显示 Docker 连接状态
- 一键安装组件（PostgreSQL、Redis）
- 安装时自定义：组件名称、端口、环境变量、数据卷、启动命令参数
- 启动 / 停止 / 重启 / 卸载已安装组件
- 查看组件详情（资源监控、日志、配置信息）
- 内置离线镜像，无需外网即可部署

#### 支持的组件类型

| 组件 | 类型 | 默认镜像 | 说明 |
|------|------|---------|------|
| PostgreSQL | `postgresql` | `postgres:16-alpine` | 关系型数据库 |
| Redis | `redis` | `redis:7-alpine` | 内存缓存 / 键值数据库 |

#### 组件详情页
- **基本信息**：组件名称、类型、镜像、版本、状态、容器 ID、创建/更新时间
- **配置信息**：JSON 格式的完整配置展示
- **资源监控**：CPU 使用率、内存使用/上限、网络接收/发送、磁盘读取/写入
- **日志查看**：支持最近 100/500/1000 行查看，支持自动刷新

### 3. 服务器管理
- 新增 / 编辑 / 删除远程服务器 SSH 配置
- 支持两种认证方式：
  - **密码认证**：用户名 + 密码
  - **密钥认证**：SSH 私钥（支持带 Passphrase 的私钥）
- SSH 连接测试
- 服务器列表展示服务状态（proxy / monitor）
- 展开行查看已部署服务详情
- 支持的操作：启动 / 停止 / 重启 / 配置 / 卸载远程服务

### 4. 远程部署
- 选择目标服务器，部署内置二进制文件：
  - `opentraffic-ops-proxy` — OpenTraffic Ops Proxy 采集代理程序
  - `opentraffic-ops` — OpenTraffic Ops 监控平台服务
- 可选同时部署配置文件
- 支持加载默认配置模板
- 防重复部署检测
- 完整的部署记录和日志追溯

### 5. 配置管理
- 查看所有已安装组件的配置列表
- 在线编辑组件配置（端口、环境变量、数据卷、启动命令）
- 配置保存后需手动重启组件生效

### 6. 使用指南
- 平台简介与特性概览
- 基础环境要求（Docker、浏览器、网络、SSH）
- 组件管理使用说明（支持的组件、配置项、常见操作）
- 服务器管理使用说明（配置项、操作、认证方式）
- 远程部署流程说明
- PostgreSQL / Redis 默认配置与参数说明
- 常见问题 FAQ（手风琴式交互）

## 单包自包含部署（无 Nginx）

本项目的核心设计目标之一是**消除对外部 Web 服务器（如 Nginx）的依赖**。传统 Vue `createWebHistory` 项目通常需要 Nginx 做静态文件服务和路由回退，而本项目通过 Go 原生的 `go:embed` 机制，将前端 `dist` 目录直接嵌入后端二进制中，由 Go 后端统一提供 HTTP 服务。

### 实现原理

#### 1. 前端构建产物嵌入 Go 二进制

在 `backend/pkg/static/static.go` 中，使用 `//go:embed` 指令将 `frontend/dist` 目录编译进二进制：

```go
//go:embed all:../../../frontend/dist
var dist embed.FS
```

这意味着运行 `go build` 后，前端的所有 HTML、JS、CSS、图片等资源都已经包含在单个可执行文件内部，无需在服务器上保留 `frontend/dist` 文件夹。

#### 2. SPA Fallback 逻辑

前端使用 Vue Router 的 `createWebHistory()`（History 模式），当用户直接访问 `/components`、`/configs` 等路由，或刷新页面时，浏览器会请求一个不存在的静态文件路径。如果后端直接返回 404，SPA 将无法正常工作。

因此，我们实现了一个自定义的 `http.Handler`：

```go
func Handler() http.Handler {
    distFS, _ := fs.Sub(dist, "frontend/dist")
    fileServer := http.FileServer(http.FS(distFS))

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. API 和 health 路由不应由静态文件处理器处理
        if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/health" {
            http.NotFound(w, r)
            return
        }

        // 2. 尝试打开请求的文件
        f, err := distFS.Open(strings.TrimPrefix(r.URL.Path, "/"))
        if err != nil {
            // 文件不存在 -> 回退到 index.html
            r.URL.Path = "/"
            fileServer.ServeHTTP(w, r)
            return
        }
        defer f.Close()

        // 3. 如果路径是目录，也回退到 index.html
        if stat, _ := f.Stat(); stat.IsDir() {
            r.URL.Path = "/"
            fileServer.ServeHTTP(w, r)
            return
        }

        // 4. 真实存在的静态文件，直接返回
        fileServer.ServeHTTP(w, r)
    })
}
```

在 `backend/cmd/server/main.go` 中，所有 `/api/*` 和 `/health` 路由注册完成后，通过 `NoRoute` 将此 Handler 设为兜底：

```go
r.NoRoute(gin.WrapH(static.Handler()))
```

Gin 的匹配优先级保证：显式注册的 API 路由优先被命中，未匹配到的路径才进入静态文件处理器。

#### 3. 同域部署，天然无跨域

前端 API 基地址设置为 `/api`（`frontend/src/api/index.ts`）：

```typescript
baseURL: '/api'
```

开发时 Vite 通过 `proxy` 将 `/api` 转发到 `localhost:8080`；生产环境中前后端共享同一个端口和域名，完全不存在跨域问题，也无需 CORS 特殊配置。

### 关键优势

| 特性 | 传统 Nginx + 后端分离 | 本项目单包自包含 |
|------|---------------------|----------------|
| 部署文件 | 多个文件/目录 + Nginx 配置 | **单个二进制文件** |
| 端口暴露 | 80/443 + 8080 | **仅一个端口** |
| 路由刷新 | 需 Nginx `try_files` 配置 | **后端自动 fallback** |
| 环境依赖 | 需安装 Nginx | **仅需 Docker** |
| 迁移成本 | 需同步前端资源目录 | **复制一个文件即可** |

## 快速开始

### 前置要求

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose（本机运行时使用）
- Git

### 开发模式启动

#### 1. 克隆项目

```bash
git clone <repository-url>
cd OpenTraffic-Ops-Initialization
```

#### 2. 启动后端

```bash
cd backend

# 安装依赖
go mod download

# 运行后端服务
go run cmd/server/main.go
```

后端服务将在 `http://localhost:8080` 启动。

#### 3. 启动前端

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端服务将在 `http://localhost:5173` 启动，开发时通过 Vite Proxy 自动转发 `/api` 到 `http://localhost:8080`。

#### 4. 访问系统

打开浏览器访问 `http://localhost:5173`

默认登录账号：
- 用户名: `admin`
- 密码: `admin123`

### 生产构建（单包自包含）

#### Windows 本地构建

在项目根目录下执行 `build.bat`：

```cmd
build.bat
```

构建完成后，`backend\opentraffic-ops-init.exe` 即为最终产物。

#### Windows 交叉编译 Linux 部署包

执行 `build-opentraffic-ops-initialization.bat` 生成 Linux AMD64 和 ARM64 二进制：

```cmd
build-opentraffic-ops-initialization.bat
```

输出文件为：
- `backend\opentraffic-ops-init-linux-amd64`
- `backend\opentraffic-ops-init-linux-arm64`

上传至 Linux 服务器并运行：

```bash
chmod +x opentraffic-ops-init-linux-amd64
./opentraffic-ops-init-linux-amd64
```

#### Linux / macOS / 手动构建

```bash
# 1. 构建前端
cd frontend
npm install
npm run build

# 2. 将前端产物复制到后端的 embed 目录
cd ..
mkdir -p backend/pkg/static/dist
cp -r frontend/dist/* backend/pkg/static/dist/

# 3. 构建后端单文件（前端 dist 已被嵌入二进制）
cd backend
go build -o opentraffic-ops-init cmd/server/main.go
```

> **注意**：`go:embed` 要求被嵌入的文件必须位于 Go 模块内部，且路径中不能包含 `..`。因此必须先把 `frontend/dist` 复制到 `backend/pkg/static/dist`，再执行 `go build`。

#### Windows 本地开发快速调试（无需每次复制 dist）

开发阶段前端改动频繁，每次都要把 `dist` 复制进后端再编译非常麻烦。我们在 `backend/pkg/static/static.go` 中增加了环境变量开关：

```cmd
# 在 backend 目录下
set RTM_STATIC_DIR=..\frontend\dist
go run cmd\server\main.go
```

设置 `RTM_STATIC_DIR` 后，Go 后端会直接从磁盘加载前端产物，不走 `go:embed`。这样你可以一边改前端、一边刷新浏览器测试，无需重新编译后端。生产构建时**不要**设置该变量，确保前端资源被完整嵌入二进制。

无论通过哪种方式构建，最终产物都是**单个自包含二进制文件**。运行后访问 `http://localhost:8080` 即可看到完整平台，直接刷新如 `http://localhost:8080/components` 也不会 404。

### Docker 部署

#### 使用 Docker Compose（完整环境）

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

#### 单独部署后端二进制

如果你已经将前端嵌入二进制，也可以直接构建并运行后端镜像：

```bash
cd backend

# 构建镜像
docker build -t opentraffic-ops-init .

# 运行容器
docker run -d \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v opentraffic-ops-init-data:/app/data \
  --name opentraffic-ops-init \
  opentraffic-ops-init
```

## 项目结构

```
opentraffic-ops-init/
├── backend/                      # Go 后端
│   ├── cmd/server/              # 入口文件
│   ├── internal/
│   │   ├── controller/          # API 控制器
│   │   │   ├── auth_controller.go
│   │   │   ├── component_controller.go
│   │   │   ├── deploy_controller.go
│   │   │   ├── monitor_controller.go
│   │   │   └── server_controller.go
│   │   ├── service/             # 业务逻辑
│   │   ├── repository/          # 数据访问
│   │   ├── model/               # 数据模型
│   │   └── middleware/          # 中间件（JWT、CORS、Recovery、ErrorHandler）
│   ├── pkg/
│   │   ├── docker/              # Docker 客户端封装
│   │   ├── ssh/                 # SSH 客户端封装
│   │   ├── config/              # 配置管理
│   │   ├── crypto/              # 加密工具（AES-GCM）
│   │   ├── static/              # 静态文件托管（go:embed）
│   │   └── assets/              # 嵌入资源（默认配置文件、二进制文件）
│   └── configs/                 # 配置文件
│
├── frontend/                     # Vue 前端
│   ├── src/
│   │   ├── views/               # 页面组件
│   │   │   ├── Dashboard.vue    # 监控大屏
│   │   │   ├── Components.vue   # 组件管理
│   │   │   ├── ComponentDetail.vue  # 组件详情
│   │   │   ├── Servers.vue      # 服务器管理
│   │   │   ├── Configs.vue      # 配置管理
│   │   │   ├── Help.vue         # 使用指南
│   │   │   ├── Login.vue        # 登录页
│   │   │   └── Layout.vue       # 布局页
│   │   ├── components/          # 通用组件
│   │   ├── api/                 # API 调用封装
│   │   ├── stores/              # Pinia 状态管理
│   │   └── router/              # 路由配置（createWebHistory）
│   └── package.json
│
├── components/                   # 组件模板（Docker Compose、配置模板）
│   ├── postgresql/
│   └── redis/
│
├── docker-compose.yaml           # Docker 编排
└── README.md                     # 本文档
```

## API 文档

### 认证接口

#### 登录
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

#### 登出
```http
POST /api/auth/logout
Authorization: Bearer <token>
```

#### 获取用户信息
```http
GET /api/users/profile
Authorization: Bearer <token>
```

### 组件管理接口

```http
# 获取组件目录（含安装状态）
GET /api/components/catalog

# 获取组件列表
GET /api/components

# 获取组件详情
GET /api/components/:id

# 安装组件
POST /api/components

# 卸载组件
DELETE /api/components/:id

# 启动/停止/重启组件
POST /api/components/:id/start
POST /api/components/:id/stop
POST /api/components/:id/restart

# 获取组件日志
GET /api/components/:id/logs

# 获取组件资源统计
GET /api/components/:id/stats

# 更新组件配置
PUT /api/components/:id/config
```

### 监控接口

```http
# 获取系统总览
GET /api/monitor/overview

# 获取组件详情（含统计信息）
GET /api/monitor/components

# WebSocket 实时监控
GET /api/monitor/realtime
```

### 服务器管理接口

```http
# 获取服务器列表
GET /api/servers

# 创建服务器
POST /api/servers

# 获取服务器详情
GET /api/servers/:id

# 更新服务器
PUT /api/servers/:id

# 删除服务器
DELETE /api/servers/:id

# 测试 SSH 连接
POST /api/servers/:id/test

# 获取 opentraffic-ops-proxy 配置
GET /api/servers/:id/proxy-config
PUT /api/servers/:id/proxy-config

# 获取/更新指定软件配置
GET /api/servers/:id/configs/:software
PUT /api/servers/:id/configs/:software

# 获取默认软件配置（嵌入资源）
GET /api/servers/configs/:software/default

# 获取服务运行状态
GET /api/servers/:id/services/:software/status

# 启动/停止/重启远程服务
POST /api/servers/:id/services/:software/start
POST /api/servers/:id/services/:software/stop
POST /api/servers/:id/services/:software/restart
```

### 部署接口

```http
# 部署二进制文件到远程服务器
POST /api/deploy

# 卸载远程服务
POST /api/deploy/undeploy

# 获取部署记录列表
GET /api/deploy/records?server_id=

# 获取部署记录详情
GET /api/deploy/records/:id
```

## 配置说明

### 后端配置

创建 `backend/.env` 文件：

```env
# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# 数据库配置
DATA_DIR=./data

# JWT 配置
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRE_HOURS=24

# 加密密钥（32 字节，用于加密 SSH 密码和私钥）
ENCRYPTION_KEY=your-encryption-key-32-bytes-long
```

### 前端配置

前端配置在 `frontend/vite.config.ts` 中：

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

- `base` 默认为 `/`，生产环境下以根路径部署，无需修改。
- `proxy` 仅在 `npm run dev` 开发模式下生效，生产构建后由 Go 后端直接托管。

## 生产部署

### 安全建议

1. 修改默认管理员密码
2. 更换 JWT Secret 和加密密钥
3. 使用 HTTPS（可通过外部反向代理或负载均衡器终止 TLS）
4. 配置防火墙规则
5. 定期备份 SQLite 数据库

### 性能优化

1. 使用生产环境构建（`npm run build`）
2. 后端已默认启用 Gzip（Gin 框架支持）
3. 静态资源由 `go:embed` 提供，内存访问速度极快
4. 如需更高性能，可在外部增加 CDN 或缓存层

### 备份策略

```bash
# 备份 SQLite 数据库
cp backend/data/opentraffic-ops-init.db backup/opentraffic-ops-init_$(date +%Y%m%d).db

# 备份 Docker volumes
docker run --rm \
  -v opentraffic-ops-init-data:/data \
  -v $(pwd)/backup:/backup \
  alpine tar czf /backup/opentraffic-ops-init-data_$(date +%Y%m%d).tar.gz /data
```

## 故障排查

### 常见问题

1. **Docker 连接失败**
   - 确保 Docker 服务正在运行
   - 检查 `/var/run/docker.sock` 权限

2. **端口冲突**
   - 修改 `.env` 文件中的端口配置
   - 确保端口未被占用

3. **前端刷新 404**
   - 确认 `go build` 时 `frontend/dist` 已存在
   - 检查 `backend/pkg/static/static.go` 中的 embed 路径是否正确

4. **跨域问题**
   - 生产环境前后端同域，不应出现跨域
   - 开发环境确保 Vite Proxy 配置正确且后端已启动

5. **SSH 连接测试失败**
   - 检查目标服务器的 SSH 服务是否正常运行
   - 确认主机地址、端口、用户名、密码/私钥是否正确
   - 检查防火墙是否放行了 SSH 端口

6. **远程部署失败（权限不足）**
   - 检查 SSH 用户对部署路径是否有读写权限
   - 确认部署路径所在磁盘有足够空间
   - 检查目标服务器的 SELinux 或 AppArmor 限制

7. **组件容器启动失败（Permission denied）**
   - 使用绑定挂载时，确保宿主机目录的属主与容器默认用户 UID 一致
   - PostgreSQL UID 为 70，Redis UID 为 999
   - 推荐使用命名卷（如 `postgres-data:/var/lib/postgresql/data`），Docker 会自动处理权限

## 开发指南

### 添加新的组件类型

1. 在 `components/` 下创建新的组件目录
2. 添加 `config.yaml.template` 和 `docker-compose.yaml.template`
3. 在后端 `internal/model/component.go` 中添加新的组件类型
4. 在前端 `src/types/index.ts` 中更新 `ComponentType` 类型

### 添加新的可部署二进制文件

1. 将二进制文件放入 `backend/pkg/assets/` 目录
2. 在 `backend/pkg/assets/assets.go` 中添加 `//go:embed` 指令
3. 更新前端部署对话框中的选项列表
4. 在 `backend/pkg/assets/images/` 下添加对应的默认配置文件

## 贡献指南

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License
