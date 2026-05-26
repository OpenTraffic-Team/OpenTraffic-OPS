<template>
  <div class="help-page">
    <div class="page-header">
      <div class="header-title-section">
        <h2>使用指南</h2>
        <p class="subtitle">了解平台的使用方法和最佳实践</p>
      </div>
    </div>

    <div class="help-body">
      <el-tabs class="dark-tabs">
        <!-- 平台简介 -->
        <el-tab-pane label="平台简介">
          <div class="help-section">
            <p class="lead">
              RTM部署面板 是一个集 Docker 容器组件部署与 SSH 远程服务器管理于一体的综合运维平台。
              它既能帮助你在本地环境快速安装、启动、监控常用中间件，也支持通过 SSH 将 opentraffic-ops-proxy 等二进制文件
              一键部署到远程 Linux 服务器，实现分布式节点的统一管理。
            </p>
            <div class="feature-grid">
              <div v-for="(feat, i) in features" :key="i" class="feature-card">
                <div class="feature-icon" :class="feat.iconClass">
                  <el-icon :size="22"><component :is="feat.icon" /></el-icon>
                </div>
                <div class="feature-title">{{ feat.title }}</div>
                <div class="feature-desc">{{ feat.desc }}</div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- 基础环境 -->
        <el-tab-pane label="基础环境">
          <div class="help-section">
            <div class="env-list">
              <div class="env-item">
                <div class="env-icon icon-docker">
                  <el-icon><Box /></el-icon>
                </div>
                <div class="env-content">
                  <div class="env-title">Docker</div>
                  <div class="env-desc">必须在本机安装并启动 Docker Desktop 或 Docker Engine，平台通过 Docker API 与守护进程通信来管理容器。</div>
                </div>
              </div>
              <div class="env-item">
                <div class="env-icon icon-browser">
                  <el-icon><Monitor /></el-icon>
                </div>
                <div class="env-content">
                  <div class="env-title">浏览器</div>
                  <div class="env-desc">推荐使用 Chrome / Edge / Firefox 最新版本访问前端页面。</div>
                </div>
              </div>
              <div class="env-item">
                <div class="env-icon icon-network">
                  <el-icon><Connection /></el-icon>
                </div>
                <div class="env-content">
                  <div class="env-title">网络</div>
                  <div class="env-desc">组件镜像部署无需外网，镜像包已内嵌在平台中。SSH 远程部署需要本机与目标服务器之间的网络可达。</div>
                </div>
              </div>
              <div class="env-item">
                <div class="env-icon icon-ssh">
                  <el-icon><Key /></el-icon>
                </div>
                <div class="env-content">
                  <div class="env-title">SSH 远程部署</div>
                  <div class="env-desc">如需使用远程服务器部署功能，需要目标 Linux 服务器开启 SSH 服务（默认端口 22），并准备好密码或私钥认证凭据。</div>
                </div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- 组件管理 -->
        <el-tab-pane label="组件管理">
          <div class="help-section">
            <h3 class="section-title"><span class="title-num">01</span>支持的组件类型</h3>
            <div class="component-types">
              <div v-for="(comp, i) in componentTypes" :key="i" class="comp-type-card">
                <div class="comp-type-name">{{ comp.name }}</div>
                <div class="comp-type-desc">{{ comp.desc }}</div>
              </div>
            </div>

            <h3 class="section-title"><span class="title-num">02</span>安装时的配置项</h3>
            <div class="config-list">
              <div v-for="(cfg, i) in installConfigs" :key="i" class="config-item">
                <div class="config-name">{{ cfg.name }}</div>
                <div class="config-desc" v-html="cfg.desc"></div>
              </div>
            </div>

            <div class="tip-box">
              <el-icon><WarningFilled /></el-icon>
              <div>
                <strong>数据卷权限提示</strong>：如果你使用命名卷（如
                <code>postgres-data:/var/lib/postgresql/data</code>），Docker 会自动处理权限，无需额外配置。
                如果你改用宿主机路径绑定挂载，请确保宿主机目录的属主与容器镜像的默认用户 UID 一致，否则容器启动时会报
                <code>Permission denied</code>。
              </div>
            </div>

            <h3 class="section-title"><span class="title-num">03</span>常见操作</h3>
            <div class="operation-list">
              <div v-for="(op, i) in operations" :key="i" class="op-item">
                <div class="op-badge">{{ op.name }}</div>
                <div class="op-desc">{{ op.desc }}</div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- 服务器管理 -->
        <el-tab-pane label="服务器管理">
          <div class="help-section">
            <p class="lead">
              服务器管理模块用于维护远程 Linux 服务器的 SSH 连接配置。配置完成后，即可通过平台将 opentraffic-ops-proxy
              等二进制文件一键部署到目标服务器，并支持远程配置编辑和部署历史追溯。
            </p>

            <h3 class="section-title"><span class="title-num">01</span>服务器配置项</h3>
            <div class="config-list">
              <div v-for="(cfg, i) in serverConfigs" :key="i" class="config-item">
                <div class="config-name">{{ cfg.name }}</div>
                <div class="config-desc" v-html="cfg.desc"></div>
              </div>
            </div>

            <div class="tip-box">
              <el-icon><WarningFilled /></el-icon>
              <div>
                <strong>安全提示</strong>：服务器的密码和私钥内容在存储时会被加密处理。
                建议使用<strong>密钥认证</strong>方式，安全性高于密码认证。私钥密码（Passphrase）为可选项。
              </div>
            </div>

            <h3 class="section-title"><span class="title-num">02</span>服务器操作</h3>
            <div class="operation-list">
              <div v-for="(op, i) in serverOperations" :key="i" class="op-item">
                <div class="op-badge">{{ op.name }}</div>
                <div class="op-desc">{{ op.desc }}</div>
              </div>
            </div>

            <h3 class="section-title"><span class="title-num">03</span>认证方式说明</h3>
            <div class="config-table">
              <div v-for="(cfg, i) in authTypes" :key="i" class="config-table-row">
                <div class="config-table-key">{{ cfg.key }}</div>
                <div class="config-table-val">{{ cfg.value }}</div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- 远程部署 -->
        <el-tab-pane label="远程部署">
          <div class="help-section">
            <p class="lead">
              远程部署功能允许你通过 SSH 将内置的二进制文件上传到目标 Linux 服务器，上传后可通过平台直接管理服务的启停。
              目前支持部署 opentraffic-ops-proxy 和 opentraffic-ops 两个组件。
            </p>

            <h3 class="section-title"><span class="title-num">01</span>可部署的二进制文件</h3>
            <div class="config-table">
              <div v-for="(cfg, i) in deployBinaries" :key="i" class="config-table-row">
                <div class="config-table-key">{{ cfg.key }}</div>
                <div class="config-table-val">{{ cfg.value }}</div>
              </div>
            </div>

            <h3 class="section-title"><span class="title-num">02</span>部署流程</h3>
            <div class="operation-list">
              <div v-for="(op, i) in deploySteps" :key="i" class="op-item">
                <div class="op-badge">{{ op.name }}</div>
                <div class="op-desc">{{ op.desc }}</div>
              </div>
            </div>

            <h3 class="section-title"><span class="title-num">03</span>服务进程管理</h3>
            <p class="section-desc">部署完成后，平台使用进程文件（PID 文件）来管理服务生命周期，无需 root 权限：</p>
            <pre class="code-block"># 上传二进制文件到部署路径
# 设置可执行权限
chmod +x /opt/rtm/opentraffic-ops-proxy-linux-amd64

# 启动服务（通过平台按钮操作）
cd /opt/rtm && setsid ./opentraffic-ops-proxy-linux-amd64 > /dev/null 2>&1 &lt;/dev/null & echo $! > opentraffic-ops-proxy.pid

# 停止服务（通过平台按钮操作）
kill $(cat /opt/rtm/opentraffic-ops-proxy.pid)
rm -f /opt/rtm/opentraffic-ops-proxy.pid</pre>

            <div class="tip-box">
              <el-icon><WarningFilled /></el-icon>
              <div>
                <strong>进程管理说明</strong>：平台通过 PID 文件记录服务进程号，使用 <code>kill -0</code> 检测进程是否存活。
                服务以普通用户身份运行，无需 <code>sudo</code> 或 root 权限。请勿手动删除 PID 文件，否则平台将无法正确判断服务状态。
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- PostgreSQL -->
        <el-tab-pane label="PostgreSQL">
          <div class="help-section">
            <h3 class="section-title"><span class="title-num">01</span>默认内置配置</h3>
            <p class="section-desc">使用内置镜像安装 PostgreSQL 时，系统会自动预填充以下默认值。核心参数均通过启动命令参数直接传入：</p>
            <div class="config-table">
              <div v-for="(cfg, i) in pgDefaults" :key="i" class="config-table-row">
                <div class="config-table-key">{{ cfg.key }}</div>
                <div class="config-table-val">{{ cfg.value }}</div>
              </div>
            </div>

            <h3 class="section-title"><span class="title-num">02</span>默认启动命令参数</h3>
            <pre class="code-block">postgres
-c max_connections=200
-c shared_buffers=512MB
-c effective_cache_size=1536MB
-c work_mem=16MB
-c maintenance_work_mem=128MB
-c log_statement=all</pre>

            <h3 class="section-title"><span class="title-num">03</span>常用启动参数说明</h3>
            <div class="param-table">
              <div class="param-table-header">
                <div>参数</div>
                <div>说明</div>
                <div>示例值</div>
              </div>
              <div v-for="(p, i) in pgCommandArgs" :key="i" class="param-table-row">
                <div class="param-arg">{{ p.arg }}</div>
                <div>{{ p.desc }}</div>
                <div class="param-example">{{ p.example }}</div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- Redis -->
        <el-tab-pane label="Redis">
          <div class="help-section">
            <h3 class="section-title"><span class="title-num">01</span>默认内置配置</h3>
            <p class="section-desc">使用内置镜像安装 Redis 时，系统会自动预填充以下默认值。所有配置均通过启动命令参数直接传入：</p>
            <div class="config-table">
              <div v-for="(cfg, i) in redisDefaults" :key="i" class="config-table-row">
                <div class="config-table-key">{{ cfg.key }}</div>
                <div class="config-table-val">{{ cfg.value }}</div>
              </div>
            </div>

            <h3 class="section-title"><span class="title-num">02</span>默认启动命令参数</h3>
            <pre class="code-block">redis-server
--appendonly yes
--requirepass admin123
--maxmemory 1536mb
--maxmemory-policy allkeys-lru
--protected-mode yes
--loglevel notice
--save "900 1"
--save "300 10"
--save "60 10000"</pre>

            <h3 class="section-title"><span class="title-num">03</span>常用启动参数说明</h3>
            <div class="param-table">
              <div class="param-table-header">
                <div>参数</div>
                <div>说明</div>
                <div>示例值</div>
              </div>
              <div v-for="(p, i) in redisCommandArgs" :key="i" class="param-table-row">
                <div class="param-arg">{{ p.arg }}</div>
                <div>{{ p.desc }}</div>
                <div class="param-example">{{ p.example }}</div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- 常见问题 -->
        <el-tab-pane label="常见问题">
          <div class="help-section">
            <div class="faq-list">
              <div
                v-for="(faq, i) in faqs"
                :key="i"
                class="faq-item"
                :class="{ 'faq-open': faqOpen === i }"
                @click="faqOpen = faqOpen === i ? null : i"
              >
                <div class="faq-question">
                  <div class="faq-q-icon">Q</div>
                  <span>{{ faq.q }}</span>
                  <el-icon class="faq-arrow"><ArrowRight /></el-icon>
                </div>
                <div class="faq-answer" v-show="faqOpen === i">
                  <div v-html="faq.a"></div>
                </div>
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import {
  Monitor, Connection, Box, WarningFilled, ArrowRight, Key
} from '@element-plus/icons-vue'

const faqOpen = ref<number | null>(null)

const features = reactive([
  { icon: 'Download', iconClass: 'icon-purple', title: '一键安装', desc: '常用中间件（PostgreSQL、Redis）一键部署' },
  { icon: 'Box', iconClass: 'icon-cyan', title: '离线镜像', desc: '内置离线镜像包，无需外网即可部署' },
  { icon: 'Setting', iconClass: 'icon-green', title: '灵活配置', desc: '安装阶段即可自定义端口、环境变量、数据卷和启动命令' },
  { icon: 'Monitor', iconClass: 'icon-amber', title: '实时监控', desc: '实时查看容器状态、日志和资源占用' },
  { icon: 'DataAnalysis', iconClass: 'icon-red', title: '配置管理', desc: '配置支持在线修改，重启后生效' },
  { icon: 'Upload', iconClass: 'icon-teal', title: '远程部署', desc: '通过 SSH 将二进制文件部署到远程 Linux 服务器' },
  { icon: 'Cpu', iconClass: 'icon-indigo', title: '服务器管理', desc: '统一管理多台远程服务器的 SSH 连接配置' },
  { icon: 'Document', iconClass: 'icon-orange', title: '部署记录', desc: '记录每次远程部署的日志和结果，支持历史追溯' },
])

const componentTypes = reactive([
  { name: 'PostgreSQL', desc: '关系型数据库' },
  { name: 'Redis', desc: '内存缓存 / 键值数据库' },
])

const installConfigs = reactive([
  { name: '组件名称', desc: '自定义该组件在本平台的显示名称，默认与组件类型相同（如 <code>postgresql</code>）。' },
  { name: '端口', desc: '将容器内部端口映射到宿主机同一端口，例如填写 <code>5432</code>，即可通过 <code>localhost:5432</code> 访问服务。' },
  { name: '环境变量', desc: '以键值对形式配置容器启动参数，例如数据库密码、时区、日志级别等。支持动态增删行。' },
  { name: '数据卷', desc: '以 <code>主机路径:容器路径</code> 形式挂载宿主机目录到容器内，用于持久化存储数据或挂载配置文件，避免容器删除后数据丢失。支持动态增删行。' },
  { name: '启动命令', desc: '自定义容器启动时的命令参数。平台已为 PostgreSQL 和 Redis 预填充了生产级默认参数，可直接使用或按需修改。' },
])

const operations = reactive([
  { name: '安装', desc: '创建组件记录并自动创建、启动 Docker 容器。安装弹窗中的默认值来自各组件类型的内置默认配置。' },
  { name: '启动 / 停止 / 重启', desc: '对已有容器进行生命周期控制，状态会实时同步到列表。' },
  { name: '卸载', desc: '停止并删除容器，同时清理本地数据库中的组件记录，请谨慎操作。' },
  { name: '详情', desc: '查看组件的实时资源占用、日志输出及容器信息。' },
])

const serverConfigs = reactive([
  { name: '名称', desc: '服务器的显示名称，用于在平台中快速识别，如 <code>生产服务器-01</code>。' },
  { name: '主机地址', desc: '目标服务器的 IP 地址或域名，如 <code>192.168.1.100</code>。' },
  { name: 'SSH 端口', desc: 'SSH 服务端口，默认为 <code>22</code>。' },
  { name: '用户名', desc: '用于 SSH 登录的用户名，如 <code>root</code>。' },
  { name: '认证方式', desc: '支持<strong>密码认证</strong>和<strong>密钥认证</strong>两种方式。密码认证需填写 SSH 密码；密钥认证需粘贴私钥内容，可选项填写私钥密码（Passphrase）。' },
  { name: '部署路径', desc: '远程服务器上的部署目录，二进制文件将上传到此路径。默认为 <code>/opt/rtm</code>。' },
  { name: '描述', desc: '服务器的补充说明信息，便于团队成员理解用途。' },
])

const serverOperations = reactive([
  { name: '新增', desc: '填写服务器 SSH 连接信息，保存到平台数据库中。敏感信息（密码、私钥）会被加密存储。' },
  { name: '编辑', desc: '修改服务器的连接配置。如果密码或私钥留空，则保留原有值不做修改。' },
  { name: '测试', desc: '使用保存的凭据尝试建立 SSH 连接，验证配置是否正确。' },
  { name: '部署', desc: '选择二进制文件（opentraffic-ops-proxy 或 opentraffic-ops），通过 SFTP 上传到远程服务器的部署路径。同一服务器上同一服务只能部署一次。' },
  { name: '配置', desc: '查看和编辑远程服务器上 opentraffic-ops-proxy 的 <code>config.json</code> 配置文件。' },
  { name: '删除', desc: '从平台中移除服务器配置记录。此操作不会删除远程服务器上的任何文件。' },
])

const authTypes = reactive([
  { key: '密码认证', value: '使用用户名 + 密码进行 SSH 登录。适合快速测试场景，但安全性相对较低。' },
  { key: '密钥认证', value: '使用 SSH 私钥进行免密登录。安全性更高，推荐在生产环境中使用。支持带密码的私钥（Passphrase）。' },
])

const deployBinaries = reactive([
  { key: 'opentraffic-ops-proxy', value: 'OpenTraffic Ops Proxy 采集代理程序，部署到远程服务器后负责采集系统指标并上报。' },
  { key: 'opentraffic-ops', value: 'OpenTraffic Ops 监控平台服务，提供监控数据的聚合、存储和展示能力。' },
])

const deploySteps = reactive([
  { name: '选择服务器', desc: '在服务器管理列表中点击「部署」按钮，进入部署对话框。' },
  { name: '选择二进制', desc: '选择要部署的文件：opentraffic-ops-proxy 或 opentraffic-ops。' },
  { name: '加载配置（可选）', desc: '可加载默认配置或自定义配置文件内容，部署时会自动写入到远程服务器。' },
  { name: '执行部署', desc: '平台通过 SSH 连接到目标服务器，使用 SFTP 上传二进制文件，设置可执行权限，并创建配置文件。' },
  { name: '查看记录', desc: '部署完成后可在「部署记录」中查看详细日志，包括每一步的执行结果。' },
])

const pgDefaults = reactive([
  { key: '端口', value: '5432' },
  { key: '默认镜像', value: 'postgres:16-alpine' },
  { key: 'POSTGRES_USER', value: 'admin（超级管理员用户名）' },
  { key: 'POSTGRES_PASSWORD', value: 'admin123（超级管理员密码）' },
  { key: 'POSTGRES_DB', value: 'myappdb（初始化时自动创建的数据库名）' },
  { key: 'TZ', value: 'Asia/Shanghai（容器时区）' },
  { key: '默认数据卷', value: 'postgres-data:/var/lib/postgresql/data（命名卷持久化）' },
  { key: '容器默认用户', value: 'postgres（UID 70）' },
])

const redisDefaults = reactive([
  { key: '端口', value: '6379' },
  { key: '默认镜像', value: 'redis:7-alpine' },
  { key: '默认数据卷', value: 'redis-data:/data（命名卷持久化）' },
  { key: '容器默认用户', value: 'redis（UID 999）' },
])

const pgCommandArgs = reactive([
  { arg: '-c max_connections', desc: '设置数据库最大并发连接数。', example: '200' },
  { arg: '-c shared_buffers', desc: '设置 PostgreSQL 用于缓存数据块的共享内存大小。', example: '512MB' },
  { arg: '-c effective_cache_size', desc: '告诉查询优化器操作系统和 PostgreSQL 缓存的总大小，影响执行计划选择。', example: '1536MB' },
  { arg: '-c work_mem', desc: '每个查询操作（排序、哈希表等）可使用的内存上限。', example: '16MB' },
  { arg: '-c maintenance_work_mem', desc: '维护操作（如 VACUUM、CREATE INDEX）可使用的内存上限。', example: '128MB' },
  { arg: '-c log_statement', desc: '控制记录哪些 SQL 语句到日志。all 表示记录所有语句。', example: 'all' }
])

const redisCommandArgs = reactive([
  { arg: '--appendonly', desc: '开启 AOF 持久化，每次写操作都记录到日志，确保数据安全。', example: 'yes' },
  { arg: '--requirepass', desc: '设置 Redis 访问密码，必须设置以防止未授权访问。', example: 'admin123' },
  { arg: '--maxmemory', desc: '设置 Redis 最大可用内存，超出后触发淘汰策略。', example: '1536mb' },
  { arg: '--maxmemory-policy', desc: '内存达到上限时的键淘汰策略。', example: 'allkeys-lru' },
  { arg: '--protected-mode', desc: '保护模式，开启后只允许本地回环或已认证连接访问。', example: 'yes' },
  { arg: '--loglevel', desc: '日志级别，控制日志输出的详细程度。', example: 'notice' },
  { arg: '--save', desc: 'RDB 快照保存策略，格式为"秒数 变更次数"。可配置多条。', example: '"900 1"' }
])

const faqs = reactive([
  {
    q: '容器一直处于 installing 状态？',
    a: '加载内置镜像并创建容器通常只需几秒到几分钟。若长时间无变化，请检查 Docker 守护进程是否正常运行，或查看组件详情页的日志输出排查错误。'
  },
  {
    q: '端口冲突导致容器启动失败？',
    a: '请确保填写的宿主机端口未被其他程序占用，或尝试更换为其他可用端口。'
  },
  {
    q: '修改了配置但容器行为没变化？',
    a: '配置管理页面保存的只是数据库记录，已运行的容器不会自动感知配置变更。请前往组件管理页面手动<strong>重启</strong>对应组件。'
  },
  {
    q: '使用宿主机目录做数据卷时报 Permission denied？',
    a: `当使用绑定挂载（宿主机具体路径映射到容器）时，容器内的默认用户（PostgreSQL UID 70，Redis UID 999）需要对宿主机目录拥有读写权限。
        请在服务器执行如下命令修正属主：<br><br>
        <pre class="code-block inline"># PostgreSQL\nsudo chown -R 70:70 /app/postgres-data\n\n# Redis\nsudo chown -R 999:999 /app/redis-data</pre><br>
        如果不想手动处理权限，推荐直接使用平台默认的<strong>命名卷</strong>（如 <code>postgres-data:/var/lib/postgresql/data</code>），Docker 会自动管理权限。`
  },
  {
    q: 'PostgreSQL / Redis 密码忘记了怎么办？',
    a: '前往<strong>配置管理</strong>页面找到对应组件，点击<strong>编辑配置</strong>即可查看当前保存的环境变量密码。修改后记得重启容器。'
  },
  {
    q: 'SSH 连接测试失败？',
    a: '请检查以下几点：<br>1. 目标服务器的 SSH 服务是否正常运行；<br>2. 主机地址和端口是否正确；<br>3. 用户名是否存在且具有 SSH 登录权限；<br>4. 密码是否正确，或私钥是否与服务器上的公钥匹配；<br>5. 防火墙是否放行了 SSH 端口。'
  },
  {
    q: '远程部署失败，提示权限不足？',
    a: '请检查以下几点：<br>1. SSH 用户是否对部署路径（默认 <code>/opt/rtm</code>）有读写权限；<br>2. 部署路径所在磁盘是否有足够空间；<br>3. 目标服务器的 SELinux 或 AppArmor 是否限制了文件执行权限。'
  },
  {
    q: '部署记录中的「进行中」状态是什么意思？',
    a: '「进行中」表示部署任务已提交但尚未完成。如果长时间处于此状态，可能是网络中断或 SSH 连接异常导致的。可以尝试重新部署或查看具体日志排查原因。'
  },
  {
    q: '如何更新已部署的 opentraffic-ops-proxy 配置？',
    a: '在服务器管理页面找到目标服务器，点击「配置」按钮，即可查看和编辑远程服务器上的 <code>~/.opentraffic-ops-proxy/config.json</code> 文件。保存后需要手动重启 opentraffic-ops-proxy 服务使配置生效。'
  },
])
</script>

<style scoped>
.help-page {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.page-header {
  flex-shrink: 0;
  padding: 24px 24px 0;
  margin-bottom: 16px;
}

.header-title-section h2 {
  margin: 0 0 4px 0;
  font-size: 22px;
  font-weight: 700;
  color: #1f2937;
  letter-spacing: -0.3px;
}

.subtitle {
  margin: 0;
  font-size: 13px;
  color: #9ca3af;
}

.help-body {
  flex: 1;
  overflow: hidden;
  padding: 0 24px 24px;
}

/* ========== 浅色标签页 ========== */
:deep(.dark-tabs .el-tabs__header) {
  margin-bottom: 20px;
  border-bottom: 1px solid #f3f4f6;
}

:deep(.dark-tabs .el-tabs__nav-wrap::after) {
  background: transparent;
}

:deep(.dark-tabs .el-tabs__item) {
  color: #9ca3af;
  font-size: 14px;
  padding: 0 20px;
  height: 44px;
  line-height: 44px;
}

:deep(.dark-tabs .el-tabs__item:hover) {
  color: #6b7280;
}

:deep(.dark-tabs .el-tabs__item.is-active) {
  color: #6366f1;
  font-weight: 500;
}

:deep(.dark-tabs .el-tabs__active-bar) {
  background: linear-gradient(90deg, #6366f1, #8b5cf6);
  height: 2px;
}

:deep(.dark-tabs) {
  height: 100%;
  display: flex;
  flex-direction: column;
}

:deep(.dark-tabs .el-tabs__content) {
  flex: 1;
  overflow: hidden;
}

:deep(.dark-tabs .el-tab-pane) {
  height: 100%;
  overflow-y: auto;
  padding-right: 4px;
}

:deep(.dark-tabs .el-tab-pane::-webkit-scrollbar) {
  width: 6px;
}
:deep(.dark-tabs .el-tab-pane::-webkit-scrollbar-thumb) {
  background: #e5e7eb;
  border-radius: 3px;
}

/* ========== 帮助内容区 ========== */
.help-section {
  padding-bottom: 24px;
}

.lead {
  font-size: 15px;
  line-height: 1.8;
  color: #6b7280;
  margin: 0 0 24px 0;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
  margin: 28px 0 16px;
}

.section-title:first-child {
  margin-top: 0;
}

.title-num {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
}

.section-desc {
  font-size: 14px;
  color: #9ca3af;
  line-height: 1.7;
  margin: 0 0 16px;
}

/* ========== 特性网格 ========== */
.feature-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

@media (max-width: 1200px) {
  .feature-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .feature-grid {
    grid-template-columns: 1fr;
  }
}

.feature-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 14px;
  padding: 20px;
  transition: all 0.3s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.feature-card:hover {
  border-color: #d1d5db;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.06);
}

.feature-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  margin-bottom: 14px;
  font-size: 20px;
}

.icon-purple { background: linear-gradient(135deg, #6366f1, #4f46e5); box-shadow: 0 4px 15px -4px rgba(99, 102, 241, 0.4); }
.icon-cyan { background: linear-gradient(135deg, #06b6d4, #0891b2); box-shadow: 0 4px 15px -4px rgba(6, 182, 212, 0.4); }
.icon-green { background: linear-gradient(135deg, #10b981, #059669); box-shadow: 0 4px 15px -4px rgba(16, 185, 129, 0.4); }
.icon-amber { background: linear-gradient(135deg, #f59e0b, #d97706); box-shadow: 0 4px 15px -4px rgba(245, 158, 11, 0.4); }
.icon-red { background: linear-gradient(135deg, #ef4444, #dc2626); box-shadow: 0 4px 15px -4px rgba(239, 68, 68, 0.4); }
.icon-teal { background: linear-gradient(135deg, #14b8a6, #0d9488); box-shadow: 0 4px 15px -4px rgba(20, 184, 166, 0.4); }
.icon-indigo { background: linear-gradient(135deg, #4f46e5, #4338ca); box-shadow: 0 4px 15px -4px rgba(79, 70, 229, 0.4); }
.icon-orange { background: linear-gradient(135deg, #f97316, #ea580c); box-shadow: 0 4px 15px -4px rgba(249, 115, 22, 0.4); }

.feature-title {
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 6px;
}

.feature-desc {
  font-size: 13px;
  color: #9ca3af;
  line-height: 1.5;
}

/* ========== 环境列表 ========== */
.env-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.env-item {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 14px;
  padding: 20px;
  transition: all 0.3s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.env-item:hover {
  border-color: #d1d5db;
}

.env-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 20px;
  flex-shrink: 0;
}

.icon-docker { background: linear-gradient(135deg, #2496ed, #1a7bc8); }
.icon-browser { background: linear-gradient(135deg, #6366f1, #4f46e5); }
.icon-network { background: linear-gradient(135deg, #10b981, #059669); }
.icon-ssh { background: linear-gradient(135deg, #f59e0b, #d97706); }

.env-title {
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 6px;
}

.env-desc {
  font-size: 13px;
  color: #9ca3af;
  line-height: 1.6;
}

/* ========== 组件类型卡片 ========== */
.component-types {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 24px;
}

@media (max-width: 768px) {
  .component-types {
    grid-template-columns: repeat(2, 1fr);
  }
}

.comp-type-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 16px;
  text-align: center;
  transition: all 0.3s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.comp-type-card:hover {
  border-color: rgba(99, 102, 241, 0.3);
  background: rgba(99, 102, 241, 0.03);
}

.comp-type-name {
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 4px;
}

.comp-type-desc {
  font-size: 12px;
  color: #9ca3af;
}

/* ========== 配置列表 ========== */
.config-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 24px;
}

.config-item {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 16px 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.config-name {
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 6px;
}

.config-desc {
  font-size: 13px;
  color: #9ca3af;
  line-height: 1.6;
}

/* ========== 提示框 ========== */
.tip-box {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  background: rgba(245, 158, 11, 0.06);
  border: 1px solid rgba(245, 158, 11, 0.12);
  border-radius: 12px;
  padding: 16px 20px;
  margin-bottom: 24px;
  font-size: 13px;
  line-height: 1.7;
  color: #6b7280;
}

.tip-box .el-icon {
  color: #f59e0b;
  font-size: 18px;
  flex-shrink: 0;
  margin-top: 2px;
}

/* ========== 操作列表 ========== */
.operation-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.op-item {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 16px 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.op-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 8px;
  background: rgba(99, 102, 241, 0.1);
  color: #6366f1;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
  margin-top: 2px;
}

.op-desc {
  font-size: 13px;
  color: #6b7280;
  line-height: 1.6;
}

/* ========== 配置表格 ========== */
.config-table {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  overflow: hidden;
  margin-bottom: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.config-table-row {
  display: flex;
  padding: 12px 20px;
  border-bottom: 1px solid #f3f4f6;
}

.config-table-row:last-child {
  border-bottom: none;
}

.config-table-key {
  width: 160px;
  flex-shrink: 0;
  font-size: 13px;
  font-weight: 500;
  color: #6b7280;
}

.config-table-val {
  flex: 1;
  font-size: 13px;
  color: #9ca3af;
}

/* ========== 代码块 ========== */
.code-block {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 16px 20px;
  font-family: 'SF Mono', 'Consolas', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.7;
  color: #374151;
  overflow-x: auto;
  white-space: pre;
  margin: 0 0 24px;
}

.code-block.inline {
  margin: 8px 0 0;
  padding: 12px 16px;
  font-size: 12px;
  border-radius: 8px;
}

/* ========== 参数表格 ========== */
.param-table {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  overflow: hidden;
  margin-bottom: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.param-table-header,
.param-table-row {
  display: grid;
  grid-template-columns: 200px 1fr 160px;
  gap: 16px;
  padding: 12px 20px;
}

.param-table-header {
  background: #f9fafb;
  border-bottom: 1px solid #f3f4f6;
  font-size: 12px;
  font-weight: 600;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.param-table-row {
  border-bottom: 1px solid #f3f4f6;
  font-size: 13px;
  color: #6b7280;
  align-items: center;
}

.param-table-row:last-child {
  border-bottom: none;
}

.param-arg {
  color: #6366f1;
  font-weight: 500;
  font-family: 'SF Mono', 'Consolas', monospace;
}

.param-example {
  color: #059669;
  font-family: 'SF Mono', 'Consolas', monospace;
}

/* ========== FAQ ========== */
.faq-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.faq-item {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.faq-item:hover {
  border-color: #d1d5db;
}

.faq-item.faq-open {
  border-color: rgba(99, 102, 241, 0.2);
}

.faq-question {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  font-size: 14px;
  font-weight: 500;
  color: #374151;
}

.faq-q-icon {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.faq-arrow {
  margin-left: auto;
  color: #d1d5db;
  font-size: 14px;
  transition: transform 0.3s ease;
}

.faq-open .faq-arrow {
  transform: rotate(90deg);
  color: #6366f1;
}

.faq-answer {
  padding: 0 20px 16px 56px;
  font-size: 13px;
  color: #9ca3af;
  line-height: 1.7;
}

/* ========== 通用样式 ========== */
code {
  background: #f3f4f6;
  padding: 2px 7px;
  border-radius: 5px;
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
  color: #d97706;
}

strong {
  color: #374151;
  font-weight: 600;
}
</style>
