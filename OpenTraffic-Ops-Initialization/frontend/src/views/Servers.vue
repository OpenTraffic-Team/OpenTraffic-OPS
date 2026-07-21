<template>
  <div class="servers-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <div class="header-title-section">
          <h2>服务器管理</h2>
          <p class="subtitle">管理和部署到远程Linux服务器</p>
        </div>
      </div>
      <button class="action-btn btn-primary" @click="openAddDialog">
        <el-icon><Plus /></el-icon>
        新增服务器
      </button>
    </div>

    <!-- 服务器卡片网格 -->
    <div class="server-grid-wrap" v-loading="serverStore.loading" element-loading-background="rgba(245, 247, 250, 0.8)" element-loading-text="加载中...">
      <el-empty v-if="!serverStore.loading && serverStore.servers.length === 0" description="暂无服务器，点击右上角新增" />
      <el-row :gutter="20">
        <el-col
          v-for="server in serverStore.servers"
          :key="server.id"
          :xs="24"
          :sm="12"
          :md="8"
          :lg="8"
          :xl="6"
          class="card-col"
        >
          <div class="server-card">
          <div class="server-card-head">
            <div class="server-title">
              <span class="server-name">{{ server.name }}</span>
              <el-tag size="small" :type="server.auth_type === 'password' ? 'warning' : 'success'" effect="dark">
                {{ server.auth_type === 'password' ? '密码' : '密钥' }}
              </el-tag>
            </div>
            <span class="host-text">{{ server.host }}:{{ server.port }}</span>
          </div>
          <div class="server-meta">
            <span class="meta-item">用户 {{ server.username }}</span>
            <span class="meta-item meta-path" :title="server.deploy_path">{{ server.deploy_path }}</span>
          </div>
          <div class="service-status-row">
            <div
              class="status-item"
              title="opentraffic-ops-proxy"
              @click="openServiceConfig(server, 'opentraffic-ops-proxy')"
            >
              <span class="status-dot" :class="getServiceStatusClass(server.id, 'opentraffic-ops-proxy')"></span>
              <span class="status-name">proxy</span>
            </div>
            <div
              class="status-item"
              title="opentraffic-ops"
              @click="openServiceConfig(server, 'opentraffic-ops')"
            >
              <span class="status-dot" :class="getServiceStatusClass(server.id, 'opentraffic-ops')"></span>
              <span class="status-name">monitor</span>
            </div>
            <div
              class="status-item"
              title="opentraffic-control"
            >
              <span class="status-dot" :class="getServiceStatusClass(server.id, 'opentraffic-control')"></span>
              <span class="status-name">control</span>
            </div>
          </div>
          <div class="server-card-foot">
            <div class="row-actions">
              <button class="action-btn btn-edit" @click="openEditDialog(server)">
                <el-icon><Edit /></el-icon>编辑
              </button>
              <button class="action-btn btn-test" @click="handleTest(server)">
                <el-icon><Connection /></el-icon>测试
              </button>
              <button class="action-btn btn-deploy" @click="openDeployDialog(server)">
                <el-icon><Upload /></el-icon>部署
              </button>
              <button class="action-btn btn-delete" @click="handleDelete(server)">
                <el-icon><Delete /></el-icon>删除
              </button>
            </div>
            <button class="expand-toggle" @click="toggleServerExpand(server)">
              服务
              <el-icon>
                <ArrowUp v-if="expandedServers.has(server.id)" />
                <ArrowDown v-else />
              </el-icon>
            </button>
          </div>
          <div v-if="expandedServers.has(server.id)" class="server-services">
            <div class="service-list">
              <div v-for="sw in softwareList" :key="sw" class="service-card">
                <div class="service-card-header">
                  <span class="service-card-name">{{ sw }}</span>
                  <el-tag
                    size="small"
                    :type="getServiceStatusType(server.id, sw)"
                    effect="dark"
                  >
                    {{ getServiceStatusLabel(server.id, sw) }}
                  </el-tag>
                </div>
                <div class="service-card-actions">
                  <template v-if="isDeployed(server.id, sw)">
                    <button
                      class="action-btn btn-start"
                      :disabled="getServiceStatus(server.id, sw) === 'running'"
                      @click="handleStartService(server.id, sw)"
                    >
                      <el-icon><VideoPlay /></el-icon>启动
                    </button>
                    <button
                      class="action-btn btn-stop"
                      :disabled="getServiceStatus(server.id, sw) !== 'running'"
                      @click="handleStopService(server.id, sw)"
                    >
                      <el-icon><VideoPause /></el-icon>停止
                    </button>
                    <button
                      class="action-btn btn-restart"
                      @click="handleRestartService(server.id, sw)"
                    >
                      <el-icon><RefreshRight /></el-icon>重启
                    </button>
                    <button
                      class="action-btn btn-config"
                      @click="openServiceConfig(server, sw)"
                    >
                      <el-icon><Setting /></el-icon>配置
                    </button>
                    <button
                      class="action-btn btn-undeploy"
                      @click="handleUndeployService(server.id, sw)"
                    >
                      <el-icon><Delete /></el-icon>卸载
                    </button>
                  </template>
                  <template v-else>
                    <span class="not-deployed-tag">未部署</span>
                  </template>
                </div>
              </div>
            </div>
          </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <!-- 新增/编辑服务器对话框 -->
    <el-dialog
      v-model="showServerDialog"
      :title="isEdit ? '编辑服务器' : '新增服务器'"
      width="700px"
      class="dark-dialog"
      destroy-on-close
    >
      <el-form :model="serverForm" label-width="100px" class="dark-form">
        <el-form-item label="名称" required>
          <el-input v-model="serverForm.name" placeholder="如：生产服务器01" />
        </el-form-item>
        <el-form-item label="主机地址" required>
          <el-input v-model="serverForm.host" placeholder="如：192.168.1.100" />
        </el-form-item>
        <el-form-item label="SSH端口" required>
          <el-input-number v-model="serverForm.port" :min="1" :max="65535" :value-on-clear="22" />
        </el-form-item>
        <el-form-item label="用户名" required>
          <el-input v-model="serverForm.username" placeholder="如：root" />
        </el-form-item>
        <el-form-item label="认证方式" required>
          <el-radio-group v-model="serverForm.auth_type">
            <el-radio-button label="password">密码</el-radio-button>
            <el-radio-button label="key">密钥</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="serverForm.auth_type === 'password'" label="密码" required>
          <el-input v-model="serverForm.password" type="password" show-password placeholder="SSH密码" />
        </el-form-item>
        <template v-if="serverForm.auth_type === 'key'">
          <el-form-item label="私钥" required>
            <el-input v-model="serverForm.private_key" type="textarea" :rows="4" placeholder="粘贴SSH私钥内容" />
          </el-form-item>
          <el-form-item label="密钥密码">
            <el-input v-model="serverForm.passphrase" type="password" show-password placeholder="私钥密码（如有）" />
          </el-form-item>
        </template>
        <el-form-item label="部署路径" required>
          <el-input v-model="serverForm.deploy_path" placeholder="如：/opt/rtm" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="serverForm.description" type="textarea" :rows="2" placeholder="服务器描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showServerDialog = false" :disabled="serverStore.loading">取消</el-button>
        <el-button type="primary" @click="handleSaveServer" :loading="serverStore.loading">确定</el-button>
      </template>
    </el-dialog>

    <!-- 部署对话框 -->
    <el-dialog
      v-model="showDeployDialog"
      title="部署"
      width="800px"
      class="dark-dialog"
      destroy-on-close
    >
      <el-form :model="deployForm" label-width="120px" class="dark-form">
        <el-form-item label="目标服务器">
          <el-input :model-value="currentServer?.name" disabled />
        </el-form-item>
        <el-form-item label="选择部署包" required>
          <el-select v-model="deployForm.binary_name" placeholder="选择要部署的资源" style="width: 100%" popper-class="light-select-dropdown">
            <el-option
              :label="deployedSoftwares.has('opentraffic-ops-proxy') ? 'opentraffic-ops-proxy (Linux AMD64) — 已部署' : 'opentraffic-ops-proxy (Linux AMD64)'"
              value="opentraffic-ops-proxy"
              :disabled="deployedSoftwares.has('opentraffic-ops-proxy')"
            />
            <el-option
              :label="deployedSoftwares.has('opentraffic-ops') ? 'opentraffic-ops (Linux AMD64) — 已部署' : 'opentraffic-ops (Linux AMD64)'"
              value="opentraffic-ops"
              :disabled="deployedSoftwares.has('opentraffic-ops')"
            />
            <el-option label="opentraffic-control（自动识别架构）" value="opentraffic-control" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="deployForm.binary_name && deployForm.binary_name !== 'opentraffic-control' && deployedSoftwares.has(deployForm.binary_name)">
          <el-alert type="warning" :closable="false" show-icon>
            <template #title>
              该服务已部署，请勿重复部署
            </template>
          </el-alert>
        </el-form-item>
        <el-form-item v-if="deployForm.binary_name === 'opentraffic-control'" label="版本号">
          <el-input v-model="deployForm.version" placeholder="如：v1.0.0（留空则自动生成时间戳版本）" />
        </el-form-item>
        <el-form-item v-if="deployForm.binary_name === 'opentraffic-control'">
          <el-alert type="info" :closable="false" show-icon>
            <template #title>
              龙芯 LoongArch64 与 ARM aarch64 首次部署会自动上传 Python 环境（trafficlight_env）到部署目录，后续重复部署只更新算法包
            </template>
          </el-alert>
        </el-form-item>
        <template>
          <el-form-item label="同时配置">
            <el-switch v-model="deployWithConfig" active-text="是" inactive-text="否" />
          </el-form-item>
          <template v-if="deployWithConfig">
            <el-form-item>
              <div class="config-actions-row">
                <el-button type="primary" text size="small" style="color: #ffffff" @click="loadDefaultDeployConfig">
                  <el-icon><Refresh /></el-icon> 加载默认配置
                </el-button>
                <el-button type="info" text size="small" @click="deployConfigContent = ''">
                  <el-icon><Delete /></el-icon> 清空
                </el-button>
              </div>
            </el-form-item>
            <el-form-item>
              <el-input
                v-model="deployConfigContent"
                type="textarea"
                :rows="12"
                placeholder="请输入JSON配置内容，或点击「加载默认配置」"
                class="config-textarea"
              />
            </el-form-item>
          </template>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="showDeployDialog = false" :disabled="serverStore.loading">取消</el-button>
        <el-button type="primary" @click="handleDeploy" :loading="serverStore.loading">开始部署</el-button>
      </template>
    </el-dialog>

    <!-- 部署记录对话框 -->
    <el-dialog
      v-model="showRecordsDialog"
      title="部署记录"
      width="1000px"
      class="dark-dialog"
    >
      <el-table :data="serverStore.deployRecords" class="dark-table" max-height="400">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="server_name" label="服务器" width="140" />
        <el-table-column prop="binary_name" label="资源" width="160" />
        <el-table-column prop="version" label="版本" width="140" show-overflow-tooltip />
        <el-table-column prop="remote_path" label="远程路径" min-width="200" show-overflow-tooltip />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="getStatusType(row.status)" effect="dark">
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="170" />
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button type="primary" text size="small" @click="showLogDialog(row)">查看日志</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 日志详情对话框 -->
    <el-dialog
      v-model="showLogDialogVisible"
      :title="`部署日志 - ${currentRecord?.binary_name}`"
      width="800px"
      class="dark-dialog"
    >
      <div class="log-content">
        <pre v-if="currentRecord?.log">{{ currentRecord.log }}</pre>
        <div v-else class="log-empty">暂无日志</div>
      </div>
    </el-dialog>

    <!-- 软件配置编辑对话框 -->
    <el-dialog
      v-model="showConfigDialog"
      :title="`${configSoftwareName} 配置 - ${currentServer?.name}`"
      width="850px"
      class="dark-dialog"
      destroy-on-close
    >
      <el-form class="dark-form">
        <el-form-item label="软件类型">
          <el-select v-model="configSoftwareName" placeholder="选择软件" style="width: 100%" @change="onConfigSoftwareChange" popper-class="light-select-dropdown">
            <el-option label="opentraffic-ops-proxy" value="opentraffic-ops-proxy" />
            <el-option label="opentraffic-ops" value="opentraffic-ops" />
            <el-option label="opentraffic-control" value="opentraffic-control" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <div class="config-path-hint">{{ getConfigPathHint() }}</div>
        </el-form-item>
        <el-form-item>
          <el-input
            v-model="configContent"
            type="textarea"
            :rows="18"
            placeholder="配置文件内容"
            class="config-textarea"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showConfigDialog = false" :disabled="serverStore.loading">取消</el-button>
        <el-button type="primary" @click="handleSaveConfig" :loading="serverStore.loading">保存</el-button>
      </template>
    </el-dialog>

    <!-- 部署记录按钮 -->
    <div class="records-floating">
      <el-button type="info" text @click="openRecordsDialog">
        <el-icon><Document /></el-icon>
        部署记录
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useServerStore } from '@/stores/server'
import type { Server, CreateServerRequest, DeployRequest, ServerServiceStatus } from '@/types'

const serverStore = useServerStore()

const showServerDialog = ref(false)
const showDeployDialog = ref(false)
const showRecordsDialog = ref(false)
const showLogDialogVisible = ref(false)
const showConfigDialog = ref(false)
const isEdit = ref(false)
const currentServer = ref<Server | null>(null)
const currentRecord = ref<any>(null)
const configContent = ref('')

const serverForm = reactive<CreateServerRequest>({
  name: '',
  host: '',
  port: 22,
  username: '',
  auth_type: 'password',
  password: '',
  private_key: '',
  passphrase: '',
  deploy_path: '/opt/rtm',
  description: ''
})

const deployForm = reactive<DeployRequest>({
  server_id: '',
  binary_name: 'opentraffic-ops-proxy',
  version: ''
})

const deployWithConfig = ref(false)
const deployConfigContent = ref('')
const configSoftwareName = ref('opentraffic-ops-proxy')
const deployedSoftwares = ref<Set<string>>(new Set())

const softwareList = ['opentraffic-ops-proxy', 'opentraffic-ops', 'opentraffic-control']

const serviceStatuses = ref<Record<string, Record<string, ServerServiceStatus>>>({})
const serverDeployedMap = ref<Record<string, Set<string>>>({})
const expandedServers = ref<Set<string>>(new Set())

const configFileNameMap: Record<string, string> = {
  'opentraffic-ops-proxy': 'config.json',
  'opentraffic-ops': 'config.yaml'
}

function getConfigPathHint(): string {
  if (configSoftwareName.value === 'opentraffic-control') {
    return `远程路径：${currentServer.value?.deploy_path || '/opt/rtm'}/opentraffic-control/config/mq_config.json`
  }
  return `远程路径：~/.${configSoftwareName.value}/${configFileNameMap[configSoftwareName.value] || 'config.json'}`
}

onMounted(() => {
  serverStore.fetchServers().then(() => {
    fetchAllServiceStatuses()
    for (const server of serverStore.servers) {
      loadServerDeployedSoftwares(server.id)
    }
  })
})

function resetServerForm() {
  serverForm.name = ''
  serverForm.host = ''
  serverForm.port = 22
  serverForm.username = ''
  serverForm.auth_type = 'password'
  serverForm.password = ''
  serverForm.private_key = ''
  serverForm.passphrase = ''
  serverForm.deploy_path = '/opt/rtm'
  serverForm.description = ''
}

function openAddDialog() {
  isEdit.value = false
  resetServerForm()
  showServerDialog.value = true
}

function openEditDialog(server: Server) {
  isEdit.value = true
  currentServer.value = server
  serverForm.name = server.name
  serverForm.host = server.host
  serverForm.port = server.port
  serverForm.username = server.username
  serverForm.auth_type = server.auth_type
  serverForm.password = ''
  serverForm.private_key = ''
  serverForm.passphrase = ''
  serverForm.deploy_path = server.deploy_path
  serverForm.description = server.description
  showServerDialog.value = true
}

async function handleSaveServer() {
  if (!serverForm.name || !serverForm.host || !serverForm.username) {
    ElMessage.warning('请填写必填项')
    return
  }
  if (serverForm.auth_type === 'password' && !serverForm.password && !isEdit.value) {
    ElMessage.warning('请填写密码')
    return
  }
  if (serverForm.auth_type === 'key' && !serverForm.private_key && !isEdit.value) {
    ElMessage.warning('请填写私钥')
    return
  }

  try {
    if (isEdit.value && currentServer.value) {
      const updateData: Partial<CreateServerRequest> = {
        name: serverForm.name,
        host: serverForm.host,
        port: serverForm.port,
        username: serverForm.username,
        auth_type: serverForm.auth_type,
        deploy_path: serverForm.deploy_path,
        description: serverForm.description
      }
      if (serverForm.password) updateData.password = serverForm.password
      if (serverForm.private_key) updateData.private_key = serverForm.private_key
      if (serverForm.passphrase) updateData.passphrase = serverForm.passphrase

      await serverStore.updateServer(currentServer.value.id, updateData)
      ElMessage.success('服务器更新成功')
    } else {
      await serverStore.createServer({ ...serverForm })
      ElMessage.success('服务器创建成功')
    }
    showServerDialog.value = false
    resetServerForm()
  } catch (error: any) {
    ElMessage.error(error?.message || '操作失败')
  }
}

async function handleDelete(server: Server) {
  try {
    await ElMessageBox.confirm(
      `确定要删除服务器 "${server.name}" 吗？`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    await serverStore.deleteServer(server.id)
    ElMessage.success('服务器删除成功')
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

async function handleTest(server: Server) {
  try {
    await serverStore.testConnection(server.id)
    ElMessage.success(`服务器 "${server.name}" 连接成功`)
  } catch (error: any) {
    ElMessage.error(`连接失败: ${error?.message || '未知错误'}`)
  }
}

async function openDeployDialog(server: Server) {
  currentServer.value = server
  deployForm.server_id = server.id
  deployForm.binary_name = 'opentraffic-ops-proxy'
  deployForm.version = ''
  deployWithConfig.value = false
  deployConfigContent.value = ''
  await loadDeployedSoftwares(server.id)
  showDeployDialog.value = true
}

async function loadDeployedSoftwares(serverId: string) {
  try {
    await serverStore.fetchDeployRecords(serverId)
    deployedSoftwares.value = new Set(
      serverStore.deployRecords
        .filter(r => r.status === 'success')
        .map(r => normalizeBinaryName(r.binary_name))
    )
  } catch {
    deployedSoftwares.value = new Set()
  }
}

async function loadDefaultDeployConfig() {
  if (!deployForm.binary_name) return
  try {
    const content = await serverStore.getDefaultSoftwareConfig(deployForm.binary_name)
    deployConfigContent.value = content
    ElMessage.success('默认配置已加载')
  } catch (error: any) {
    ElMessage.error(`加载默认配置失败: ${error?.message || '未知错误'}`)
  }
}

async function handleDeploy() {
  if (!deployForm.binary_name) {
    ElMessage.warning('请选择要部署的资源')
    return
  }
  if (deployForm.binary_name !== 'opentraffic-control' && deployedSoftwares.value.has(deployForm.binary_name)) {
    ElMessage.warning(`服务 ${deployForm.binary_name} 已部署，请勿重复部署`)
    return
  }
  try {
    const payload: DeployRequest = { ...deployForm }
    if (payload.binary_name === 'opentraffic-control' && !payload.version?.trim()) {
      payload.version = `v${new Date().toISOString().replace(/[-:T.Z]/g, '').slice(0, 14)}`
    }
    if (deployWithConfig.value && deployConfigContent.value.trim()) {
      payload.config_content = deployConfigContent.value.trim()
    }
    const record = await serverStore.deploy(payload)
    if (record.status === 'success') {
      ElMessage.success('部署成功')
    } else {
      ElMessage.warning('部署失败，请查看日志')
    }
    showDeployDialog.value = false
    // 部署成功后刷新部署记录和服务状态
    if (currentServer.value) {
      await loadServerDeployedSoftwares(currentServer.value.id)
      await refreshServiceStatus(currentServer.value.id, payload.binary_name)
    }
  } catch (error: any) {
    ElMessage.error(`部署失败: ${error?.message || '未知错误'}`)
  }
}

async function openRecordsDialog() {
  await serverStore.fetchDeployRecords()
  showRecordsDialog.value = true
}

function showLogDialog(record: any) {
  currentRecord.value = record
  showLogDialogVisible.value = true
}

async function openServiceConfig(server: Server, software: string) {
  currentServer.value = server
  configSoftwareName.value = software
  configContent.value = ''
  showConfigDialog.value = true
  try {
    const content = await serverStore.getSoftwareConfig(server.id, software)
    configContent.value = content
  } catch (error: any) {
    ElMessage.error(`获取配置失败: ${error?.message || '未知错误'}`)
  }
}

async function onConfigSoftwareChange() {
  if (!currentServer.value) return
  configContent.value = ''
  try {
    const content = await serverStore.getSoftwareConfig(currentServer.value.id, configSoftwareName.value)
    configContent.value = content
  } catch (error: any) {
    ElMessage.error(`获取配置失败: ${error?.message || '未知错误'}`)
  }
}

async function handleSaveConfig() {
  if (!currentServer.value) return
  try {
    await serverStore.updateSoftwareConfig(currentServer.value.id, configSoftwareName.value, configContent.value)
    ElMessage.success('配置保存成功')
    showConfigDialog.value = false
  } catch (error: any) {
    ElMessage.error(`保存失败: ${error?.message || '未知错误'}`)
  }
}

// 部署相关
function isDeployed(serverId: string, software: string): boolean {
  return serverDeployedMap.value[serverId]?.has(software) || false
}

function normalizeBinaryName(binaryName: string): string {
  if (binaryName === 'opentraffic-control-linux-amd64') {
    return 'opentraffic-control'
  }
  return binaryName
}

async function loadServerDeployedSoftwares(serverId: string) {
  try {
    await serverStore.fetchDeployRecords(serverId)
    const set = new Set(
      serverStore.deployRecords
        .filter(r => r.status === 'success')
        .map(r => normalizeBinaryName(r.binary_name))
    )
    serverDeployedMap.value[serverId] = set
  } catch {
    serverDeployedMap.value[serverId] = new Set()
  }
}

async function handleUndeployService(serverId: string, software: string) {
  try {
    const isControl = software === 'opentraffic-control'
    await ElMessageBox.confirm(
      isControl
        ? `确定要卸载算法包 "${software}" 吗？这将删除远程目录并清除部署记录。`
        : `确定要卸载服务 "${software}" 吗？这将停止服务、删除远程文件并清除部署记录。`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    await serverStore.undeploy(serverId, software)
    ElMessage.success(`${software} 卸载成功`)
    // 刷新部署状态并清空本地服务状态，使界面立即显示"未部署"
    await loadServerDeployedSoftwares(serverId)
    if (serviceStatuses.value[serverId]) {
      delete serviceStatuses.value[serverId][software]
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(`卸载失败: ${error?.message || '未知错误'}`)
    }
  }
}

// 服务状态相关
function getServiceStatus(serverId: string, software: string): string {
  return serviceStatuses.value[serverId]?.[software]?.status || 'unknown'
}

function getServiceStatusType(serverId: string, software: string): string {
  if (!isDeployed(serverId, software)) {
    return 'info'
  }
  const status = getServiceStatus(serverId, software)
  const map: Record<string, string> = {
    running: 'success',
    stopped: 'info',
    unknown: 'danger'
  }
  return map[status] || 'info'
}

function getServiceStatusLabel(serverId: string, software: string): string {
  if (!isDeployed(serverId, software)) {
    return '未部署'
  }
  const status = getServiceStatus(serverId, software)
  const map: Record<string, string> = {
    running: '运行中',
    stopped: '已停止',
    unknown: '未知'
  }
  return map[status] || '未知'
}

function getServiceStatusClass(serverId: string, software: string): string {
  if (!isDeployed(serverId, software)) {
    return 'status-stopped'
  }
  const status = getServiceStatus(serverId, software)
  return `status-${status}`
}

async function fetchAllServiceStatuses() {
  const promises: Promise<void>[] = []
  for (const server of serverStore.servers) {
    for (const sw of softwareList) {
      promises.push(
        serverStore.getServiceStatus(server.id, sw).then(status => {
          if (!serviceStatuses.value[server.id]) {
            serviceStatuses.value[server.id] = {}
          }
          serviceStatuses.value[server.id][sw] = status
        }).catch(() => {
          if (!serviceStatuses.value[server.id]) {
            serviceStatuses.value[server.id] = {}
          }
          serviceStatuses.value[server.id][sw] = {
            software: sw,
            status: 'unknown',
            label: '未知'
          }
        })
      )
    }
  }
  await Promise.all(promises)
}

async function refreshServiceStatus(serverId: string, software: string) {
  try {
    const status = await serverStore.getServiceStatus(serverId, software)
    if (!serviceStatuses.value[serverId]) {
      serviceStatuses.value[serverId] = {}
    }
    serviceStatuses.value[serverId][software] = status
  } catch (e) {
    // ignore
  }
}

async function handleStartService(serverId: string, software: string) {
  try {
    await serverStore.startService(serverId, software)
    ElMessage.success(`${software} 启动成功`)
    await new Promise(r => setTimeout(r, 1500))
    await refreshServiceStatus(serverId, software)
  } catch (error: any) {
    ElMessage.error(`启动失败: ${error?.message || '未知错误'}`)
    setTimeout(() => refreshServiceStatus(serverId, software), 2000)
  }
}

async function handleStopService(serverId: string, software: string) {
  try {
    await serverStore.stopService(serverId, software)
    ElMessage.success(`${software} 停止成功`)
    await new Promise(r => setTimeout(r, 800))
    await refreshServiceStatus(serverId, software)
  } catch (error: any) {
    ElMessage.error(`停止失败: ${error?.message || '未知错误'}`)
  }
}

async function handleRestartService(serverId: string, software: string) {
  try {
    await serverStore.restartService(serverId, software)
    ElMessage.success(`${software} 重启成功`)
    await new Promise(r => setTimeout(r, 2000))
    await refreshServiceStatus(serverId, software)
  } catch (error: any) {
    ElMessage.error(`重启失败: ${error?.message || '未知错误'}`)
    setTimeout(() => refreshServiceStatus(serverId, software), 2500)
  }
}

async function toggleServerExpand(server: Server) {
  if (expandedServers.value.has(server.id)) {
    expandedServers.value.delete(server.id)
    return
  }
  expandedServers.value.add(server.id)
  await loadServerDeployedSoftwares(server.id)
  await fetchAllServiceStatusesForServer(server.id)
}

async function fetchAllServiceStatusesForServer(serverId: string) {
  for (const sw of softwareList) {
    try {
      const status = await serverStore.getServiceStatus(serverId, sw)
      if (!serviceStatuses.value[serverId]) {
        serviceStatuses.value[serverId] = {}
      }
      serviceStatuses.value[serverId][sw] = status
    } catch {
      if (!serviceStatuses.value[serverId]) {
        serviceStatuses.value[serverId] = {}
      }
      serviceStatuses.value[serverId][sw] = {
        software: sw,
        status: 'unknown',
        label: '未知'
      }
    }
  }
}

function getStatusType(status: string) {
  const map: Record<string, string> = {
    pending: 'warning',
    success: 'success',
    failed: 'danger'
  }
  return map[status] || 'info'
}

function getStatusLabel(status: string) {
  const map: Record<string, string> = {
    pending: '进行中',
    success: '成功',
    failed: '失败'
  }
  return map[status] || status
}
</script>

<style scoped>
.servers-page {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  flex-shrink: 0;
  padding: 24px 24px 0;
}

.header-left {
  display: flex;
  align-items: flex-start;
  gap: 20px;
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

/* 服务器卡片网格 */
.server-grid-wrap {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 0 24px 24px;
}

.card-col {
  margin-bottom: 20px;
}

.server-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  overflow: hidden;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
  box-sizing: border-box;
  transition: all 0.3s ease;
}

.server-card:hover {
  border-color: #d1d5db;
  box-shadow: 0 10px 40px -10px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.server-card-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.server-title {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.server-name {
  font-size: 17px;
  font-weight: 600;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.server-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #9ca3af;
  min-width: 0;
}

.meta-item {
  white-space: nowrap;
}

.meta-path {
  font-family: 'SF Mono', 'Consolas', monospace;
  color: #6b7280;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

.server-card-foot {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  border-top: 1px solid #f3f4f6;
  padding-top: 10px;
}

.expand-toggle {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: none;
  background: rgba(99, 102, 241, 0.1);
  color: #6366f1;
  border-radius: 6px;
  padding: 4px 10px;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
  white-space: nowrap;
}

.expand-toggle:hover {
  background: rgba(99, 102, 241, 0.2);
}

.server-services {
  border-top: 1px solid #f3f4f6;
  padding-top: 12px;
}

/* 操作按钮 */
.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  outline: none;
  font-family: inherit;
}

.btn-primary {
  height: 36px;
  padding: 0 16px;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  color: #fff;
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 15px -4px rgba(99, 102, 241, 0.5);
}

.row-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  white-space: nowrap;
}

.row-actions .action-btn,
.service-card-actions .action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 3px;
  padding: 4px 8px;
  border-radius: 6px;
  border: none;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
  white-space: nowrap;
  line-height: 1;
}

.row-actions .action-btn:disabled,
.service-card-actions .action-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.row-actions .btn-edit,
.service-card-actions .btn-config {
  background: rgba(99, 102, 241, 0.1);
  color: #6366f1;
}
.row-actions .btn-edit:hover:not(:disabled),
.service-card-actions .btn-config:hover:not(:disabled) {
  background: rgba(99, 102, 241, 0.2);
}

.row-actions .btn-test,
.service-card-actions .btn-start {
  background: rgba(16, 185, 129, 0.1);
  color: #059669;
}
.row-actions .btn-test:hover:not(:disabled),
.service-card-actions .btn-start:hover:not(:disabled) {
  background: rgba(16, 185, 129, 0.2);
}

.row-actions .btn-deploy,
.service-card-actions .btn-restart {
  background: rgba(245, 158, 11, 0.1);
  color: #d97706;
}
.row-actions .btn-deploy:hover:not(:disabled),
.service-card-actions .btn-restart:hover:not(:disabled) {
  background: rgba(245, 158, 11, 0.2);
}

.row-actions .btn-delete,
.service-card-actions .btn-stop {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
}
.row-actions .btn-delete:hover:not(:disabled),
.service-card-actions .btn-stop:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.2);
}

.service-card-actions .btn-undeploy {
  background: rgba(107, 114, 128, 0.1);
  color: #6b7280;
}
.service-card-actions .btn-undeploy:hover:not(:disabled) {
  background: rgba(107, 114, 128, 0.2);
}

/* 服务状态行 */
.service-status-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 5px;
  cursor: pointer;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}

.status-running { background: #22c55e; box-shadow: 0 0 6px rgba(34, 197, 94, 0.4); }
.status-stopped { background: #9ca3af; }
.status-unknown { background: #ef4444; box-shadow: 0 0 6px rgba(239, 68, 68, 0.4); }

.status-name {
  font-size: 12px;
  color: #6b7280;
  font-family: 'SF Mono', 'Consolas', monospace;
}

.service-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.service-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 12px 16px;
}

.service-card-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.service-card-name {
  font-size: 13px;
  font-weight: 600;
  color: #374151;
  font-family: 'SF Mono', 'Consolas', monospace;
}

.service-card-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.service-card-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.not-deployed-tag {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 6px;
  background: #f3f4f6;
  color: #9ca3af;
  font-size: 12px;
  font-weight: 500;
}

/* 配置对话框 */
.config-path-hint {
  font-size: 12px;
  color: #9ca3af;
  padding: 6px 10px;
  background: #f9fafb;
  border-radius: 6px;
  font-family: 'SF Mono', 'Consolas', monospace;
}

.config-actions-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

:deep(.config-textarea .el-textarea__inner) {
  background: #f9fafb;
  border-color: #e5e7eb;
  color: #059669;
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.6;
}

.host-text {
  font-family: 'SF Mono', 'Consolas', monospace;
  color: #6b7280;
}

/* 浮动按钮 */
.records-floating {
  position: absolute;
  bottom: 24px;
  right: 24px;
}

/* 日志内容 */
.log-content {
  background: #f9fafb;
  border-radius: 8px;
  padding: 16px;
  max-height: 400px;
  overflow: auto;
}

.log-content pre {
  margin: 0;
  color: #374151;
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}

.log-empty {
  color: #9ca3af;
  text-align: center;
  padding: 20px;
}

/* 浅色表格 */
:deep(.dark-table) {
  background: transparent;
}

:deep(.dark-table .el-table__inner-wrapper::before) {
  display: none;
}

:deep(.dark-table .el-table__header-wrapper) {
  background: transparent;
}

:deep(.dark-table .el-table__header th) {
  background: #f9fafb !important;
  color: #1f2937 !important;
  font-weight: 800;
  font-size: 13px;
  border-bottom: 1px solid #f3f4f6 !important;
  padding: 12px 16px;
}

:deep(.dark-table .el-table__row) {
  background: transparent !important;
}

:deep(.dark-table .el-table__row:hover td) {
  background: #f9fafb !important;
}

:deep(.dark-table .el-table__row td) {
  background: transparent !important;
  color: #374151 !important;
  font-size: 13px;
  border-bottom: 1px solid #f3f4f6 !important;
  padding: 14px 16px;
}

:deep(.dark-table .el-table__empty-text) {
  color: #9ca3af;
}

/* 展开行样式 */
:deep(.dark-table .el-table__expand-icon) {
  color: #6b7280;
}

:deep(.dark-table .el-table__expanded-cell) {
  padding: 0 16px;
  background: transparent !important;
}

/* 浅色对话框 */
:deep(.dark-dialog .el-dialog) {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
}

:deep(.dark-dialog .el-dialog__header) {
  border-bottom: 1px solid #f3f4f6;
  padding: 20px 24px;
  margin-right: 0;
}

:deep(.dark-dialog .el-dialog__title) {
  color: #1f2937;
  font-size: 16px;
  font-weight: 600;
}

:deep(.dark-dialog .el-dialog__body) {
  padding: 24px;
}

:deep(.dark-dialog .el-dialog__footer) {
  border-top: 1px solid #f3f4f6;
  padding: 16px 24px;
}

/* 浅色表单 */
:deep(.dark-form .el-form-item__label) {
  color: #6b7280;
}

:deep(.dark-form .el-input__wrapper) {
  background: #f9fafb;
  box-shadow: 0 0 0 1px #e5e7eb inset;
}

:deep(.dark-form .el-input__inner) {
  color: #374151;
}

:deep(.dark-form .el-textarea__inner) {
  background: #f9fafb;
  border-color: #e5e7eb;
  color: #374151;
}

:deep(.dark-form .el-input-number__decrease,
      .dark-form .el-input-number__increase) {
  background: #f3f4f6;
  border-color: #e5e7eb;
  color: #6b7280;
}

:deep(.dark-form .el-radio-button__inner) {
  background: #f9fafb;
  border-color: #e5e7eb;
  color: #6b7280;
}

:deep(.dark-form .el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  border-color: transparent;
  color: #fff;
}

/* el-select 浅色 */
:deep(.dark-form .el-select .el-input__wrapper) {
  background: #f9fafb;
  box-shadow: 0 0 0 1px #e5e7eb inset;
}

:deep(.dark-form .el-select .el-input__inner) {
  color: #374151;
}

:deep(.dark-form .el-select .el-input__inner::placeholder) {
  color: #9ca3af;
}

:deep(.dark-form .el-select .el-select__icon) {
  color: #9ca3af;
}

:deep(.dark-form .el-switch__label) {
  color: #9ca3af;
}

:deep(.dark-form .el-switch__label.is-active) {
  color: #374151;
}

:deep(.dark-form .el-alert) {
  background: rgba(99, 102, 241, 0.08);
  border: 1px solid rgba(99, 102, 241, 0.15);
}

:deep(.dark-form .el-alert__title) {
  color: #6366f1;
  font-size: 12px;
}
</style>

<style>
/* el-select 下拉框浅色（全局，因为 popper 挂载在 body 上） */
.light-select-dropdown.el-popper {
  background: #ffffff !important;
  border: 1px solid #e5e7eb !important;
  border-radius: 8px !important;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1) !important;
}

.light-select-dropdown .el-select-dropdown {
  background: transparent !important;
}

.light-select-dropdown .el-scrollbar {
  background: transparent !important;
}

.light-select-dropdown .el-scrollbar__wrap {
  background: transparent !important;
}

.light-select-dropdown .el-select-dropdown__list {
  padding: 4px;
  background: transparent !important;
}

.light-select-dropdown .el-select-dropdown__item {
  color: #374151 !important;
  border-radius: 6px;
  font-size: 13px;
  background: transparent !important;
}

.light-select-dropdown .el-select-dropdown__item:hover,
.light-select-dropdown .el-select-dropdown__item.hover {
  background: #f3f4f6 !important;
  color: #1f2937 !important;
}

.light-select-dropdown .el-select-dropdown__item.selected {
  background: rgba(99, 102, 241, 0.1) !important;
  color: #6366f1 !important;
  font-weight: 500;
}

.light-select-dropdown .el-popper__arrow::before {
  background: #ffffff !important;
  border-color: #e5e7eb !important;
}
</style>