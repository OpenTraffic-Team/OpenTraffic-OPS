<template>
  <div class="components-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <div class="header-title-section">
          <h2>组件管理</h2>
          <p class="subtitle">管理和部署基础设施组件</p>
        </div>
        <div class="header-tags">
          <span
            v-if="dockerStatusKnown"
            class="status-pill"
            :class="dockerAvailable ? 'pill-success' : 'pill-danger'"
          >
            <span class="pill-dot"></span>
            Docker {{ dockerAvailable ? '已连接' : '不可用' }}
          </span>
          <el-tooltip
            v-if="!dockerAvailable && dockerError"
            effect="dark"
            :content="dockerError"
            placement="bottom"
          >
            <el-icon class="error-hint"><Warning /></el-icon>
          </el-tooltip>
        </div>
      </div>
    </div>

    <!-- 组件卡片网格 -->
    <div v-loading="componentStore.loading" class="card-grid" element-loading-background="rgba(245, 247, 250, 0.8)" element-loading-text="加载中...">
      <el-row :gutter="20">
        <el-col
          v-for="item in componentStore.catalog"
          :key="item.type"
          :xs="24"
          :sm="12"
          :md="8"
          :lg="8"
          :xl="6"
          class="card-col"
        >
          <div class="component-card">
            <div class="card-top">
              <div class="card-header">
                <h3 class="component-name">{{ item.name }}</h3>
                <span class="type-tag">{{ item.type }}</span>
              </div>

              <p class="component-desc">{{ item.description }}</p>

              <div class="component-meta">
                <div class="meta-item">
                  <el-icon><Box /></el-icon>
                  <span class="meta-value" :title="item.default_image">{{ item.default_image }}</span>
                </div>
                <div class="meta-item">
                  <el-icon><Connection /></el-icon>
                  <span class="meta-value">{{ item.default_port }}</span>
                </div>
              </div>
            </div>

            <div class="card-bottom">
              <div class="component-status">
                <span class="status-badge" :class="getCatalogStatusClass(item)">
                  {{ getCatalogStatusLabel(item) }}
                </span>
              </div>

              <div class="card-actions">
                <template v-if="!item.installed">
                  <button
                    class="action-btn btn-primary"
                    :disabled="!item.docker_available"
                    @click="openInstallConfigDialog(item)"
                  >
                    <el-icon><Download /></el-icon>
                    安装
                  </button>
                </template>
                <template v-else>
                  <div class="action-group">
                    <button
                      class="action-btn btn-small"
                      :class="{ 'btn-active': item.status === 'running' }"
                      :disabled="item.status === 'running'"
                      @click="handleControl(item.component_id!, 'start')"
                    >
                      <el-icon><VideoPlay /></el-icon>
                    </button>
                    <button
                      class="action-btn btn-small"
                      :class="{ 'btn-active': item.status === 'stopped' }"
                      :disabled="item.status === 'stopped'"
                      @click="handleControl(item.component_id!, 'stop')"
                    >
                      <el-icon><VideoPause /></el-icon>
                    </button>
                    <button
                      class="action-btn btn-small"
                      @click="handleControl(item.component_id!, 'restart')"
                    >
                      <el-icon><RefreshRight /></el-icon>
                    </button>
                    <button
                      class="action-btn btn-small btn-danger"
                      @click="handleUninstall(item)"
                    >
                      <el-icon><Delete /></el-icon>
                    </button>
                  </div>
                  <button
                    class="action-btn btn-ghost"
                    @click="$router.push(`/components/${item.component_id}`)"
                  >
                    详情
                    <el-icon><ArrowRight /></el-icon>
                  </button>
                </template>
              </div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <!-- 安装配置对话框 -->
    <el-dialog
      v-model="showInstallConfigDialog"
      title="安装组件"
      width="900px"
      class="dark-dialog"
      destroy-on-close
    >
      <div v-loading="componentStore.loading" element-loading-background="rgba(245, 247, 250, 0.8)">
        <el-row :gutter="24">
          <!-- 左侧表单 -->
          <el-col :span="15">
            <el-form :model="installConfigForm" label-width="90px" class="dark-form">
              <el-form-item label="组件名称">
                <el-input v-model="installConfigForm.name" placeholder="请输入组件名称" />
              </el-form-item>

              <el-form-item label="端口">
                <el-input v-model="installConfigForm.port" placeholder="如5432" />
              </el-form-item>

              <el-form-item label="环境变量">
                <div
                  v-for="(item, index) in installConfigForm.envList"
                  :key="index"
                  class="dynamic-row"
                >
                  <el-input v-model="item.key" placeholder="KEY" style="width: 150px" />
                  <el-input v-model="item.value" placeholder="VALUE" style="width: 150px; margin-left: 8px" />
                  <el-button
                    type="danger"
                    :icon="Delete"
                    circle
                    size="small"
                    style="margin-left: 8px"
                    @click="removeEnv(index)"
                  />
                </div>
                <el-button type="primary" text class="add-btn" @click="addEnv">
                  <el-icon><Plus /></el-icon> 添加环境变量
                </el-button>
              </el-form-item>

              <el-form-item label="数据卷">
                <div
                  v-for="(_, index) in installConfigForm.volumesList"
                  :key="index"
                  class="dynamic-row"
                >
                  <el-input
                    v-model="installConfigForm.volumesList[index]"
                    placeholder="主机路径:容器路径"
                    style="width: 308px"
                  />
                  <el-button
                    type="danger"
                    :icon="Delete"
                    circle
                    size="small"
                    style="margin-left: 8px"
                    @click="removeVolume(index)"
                  />
                </div>
                <el-button type="primary" text class="add-btn" @click="addVolume">
                  <el-icon><Plus /></el-icon> 添加数据卷
                </el-button>
              </el-form-item>

              <el-form-item label="启动命令">
                <div
                  v-for="(_, index) in installConfigForm.commandList"
                  :key="index"
                  class="dynamic-row"
                >
                  <el-input
                    v-model="installConfigForm.commandList[index]"
                    placeholder="参数"
                    style="width: 308px"
                  />
                  <el-button
                    type="danger"
                    :icon="Delete"
                    circle
                    size="small"
                    style="margin-left: 8px"
                    @click="removeCommandArg(index)"
                  />
                </div>
                <el-button type="primary" text class="add-btn" @click="addCommandArg">
                  <el-icon><Plus /></el-icon> 添加命令参数
                </el-button>
              </el-form-item>
            </el-form>
          </el-col>

          <!-- 右侧信息面板 -->
          <el-col :span="9">
            <div class="info-panel" v-if="currentCatalogItem">
              <div class="info-panel-header">
                <el-icon><InfoFilled /></el-icon>
                <span>组件基本信息</span>
              </div>
              <div class="info-panel-body">
                <div class="info-row">
                  <span class="info-label">组件名称</span>
                  <span class="info-value">{{ currentCatalogItem.name }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">组件类型</span>
                  <span class="info-value type-tag-small">{{ currentCatalogItem.type }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">默认镜像</span>
                  <span class="info-value">{{ currentCatalogItem.default_image }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">默认端口</span>
                  <span class="info-value">{{ currentCatalogItem.default_port }}</span>
                </div>

                <div class="info-divider"></div>

                <div class="info-section-title">默认环境变量</div>
                <div v-if="!defaultEnvList.length" class="info-empty">无</div>
                <div v-for="(e, i) in defaultEnvList" :key="i" class="info-kv">
                  <span class="info-k">{{ e.key }}</span>
                  <span class="info-sep">=</span>
                  <span class="info-v">{{ e.value }}</span>
                </div>

                <div class="info-section-title">建议数据卷</div>
                <div v-if="!defaultVolumesList.length" class="info-empty">无</div>
                <div v-for="(v, i) in defaultVolumesList" :key="i" class="info-line">
                  {{ v }}
                </div>

                <div class="info-section-title">默认启动命令参数</div>
                <div v-if="!defaultCommandList.length" class="info-empty">无</div>
                <div v-for="(c, i) in defaultCommandList" :key="i" class="info-line command">
                  {{ c }}
                </div>
              </div>
            </div>
          </el-col>
        </el-row>
      </div>

      <template #footer>
        <el-button @click="showInstallConfigDialog = false" :disabled="componentStore.loading">取消</el-button>
        <el-button type="primary" @click="handleInstallConfigConfirm" :loading="componentStore.loading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import { useComponentStore } from '@/stores/component'
import type { ComponentCatalogItem, ComponentConfig } from '@/types'

const componentStore = useComponentStore()

const showInstallConfigDialog = ref(false)
const currentCatalogItem = ref<ComponentCatalogItem | null>(null)

interface EnvItem {
  key: string
  value: string
}

const installConfigForm = reactive({
  name: '',
  port: '',
  envList: [] as EnvItem[],
  volumesList: [] as string[],
  commandList: [] as string[]
})

const defaultEnvList = computed(() => {
  if (!currentCatalogItem.value?.default_config?.env) return []
  return Object.entries(currentCatalogItem.value.default_config.env).map(([key, value]) => ({
    key,
    value: String(value)
  }))
})

const defaultVolumesList = computed(() => {
  return currentCatalogItem.value?.default_config?.volumes || []
})

const defaultCommandList = computed(() => {
  return currentCatalogItem.value?.default_config?.command || []
})

const dockerStatusKnown = computed(() => componentStore.catalog.length > 0)
const dockerAvailable = computed(() =>
  componentStore.catalog.some(item => item.docker_available)
)
const dockerError = computed(() => {
  const item = componentStore.catalog.find(i => i.docker_error)
  return item?.docker_error || ''
})

onMounted(() => {
  componentStore.fetchCatalog()
})

function openInstallConfigDialog(item: ComponentCatalogItem) {
  currentCatalogItem.value = item
  installConfigForm.name = item.type
  installConfigForm.port = item.default_port

  const env = item.default_config?.env || {}
  installConfigForm.envList = Object.entries(env).map(([key, value]) => ({
    key,
    value: String(value)
  }))

  installConfigForm.volumesList = item.default_config?.volumes
    ? [...item.default_config.volumes]
    : []

  installConfigForm.commandList = item.default_config?.command
    ? [...item.default_config.command]
    : []

  showInstallConfigDialog.value = true
}

function addEnv() {
  installConfigForm.envList.push({ key: '', value: '' })
}

function removeEnv(index: number) {
  installConfigForm.envList.splice(index, 1)
}

function addVolume() {
  installConfigForm.volumesList.push('')
}

function removeVolume(index: number) {
  installConfigForm.volumesList.splice(index, 1)
}

function addCommandArg() {
  installConfigForm.commandList.push('')
}

function removeCommandArg(index: number) {
  installConfigForm.commandList.splice(index, 1)
}

async function handleInstallConfigConfirm() {
  if (!currentCatalogItem.value) return

  const item = currentCatalogItem.value
  const env: Record<string, string> = {}
  installConfigForm.envList.forEach(({ key, value }) => {
    if (key.trim()) {
      env[key.trim()] = value
    }
  })

  const config: ComponentConfig = {
    port: installConfigForm.port,
    env,
    volumes: installConfigForm.volumesList.filter(v => v.trim()),
    command: installConfigForm.commandList.filter(v => v.trim())
  }

  try {
    await componentStore.installComponent({
      name: installConfigForm.name,
      type: item.type,
      image_source: 'embedded',
      version: item.default_version,
      config
    })
    ElMessage.success('组件安装成功')
    await componentStore.fetchCatalog()
    showInstallConfigDialog.value = false
    resetInstallConfigForm()
  } catch (error: any) {
    ElMessage.error('组件安装失败')
  }
}

function resetInstallConfigForm() {
  installConfigForm.name = ''
  installConfigForm.port = ''
  installConfigForm.envList = []
  installConfigForm.volumesList = []
  installConfigForm.commandList = []
  currentCatalogItem.value = null
}

async function handleControl(id: string, action: 'start' | 'stop' | 'restart') {
  try {
    await componentStore.controlComponent(id, action)
    ElMessage.success(`${action === 'start' ? '启动' : action === 'stop' ? '停止' : '重启'}成功`)
    await componentStore.fetchCatalog()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

async function handleUninstall(item: ComponentCatalogItem) {
  if (!item.component_id) return
  try {
    await ElMessageBox.confirm(
      `确定要卸载组件 "${item.name}" 吗？此操作不可恢复。`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await componentStore.uninstallComponent(item.component_id)
    ElMessage.success('组件卸载成功')
    await componentStore.fetchCatalog()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('组件卸载失败')
    }
  }
}

function getCatalogStatusClass(item: ComponentCatalogItem) {
  if (!item.installed) return 'status-not-installed'
  const classes: Record<string, string> = {
    running: 'status-running',
    stopped: 'status-stopped',
    error: 'status-error',
    installing: 'status-installing',
    unknown: 'status-unknown'
  }
  return classes[item.status || 'unknown'] || 'status-unknown'
}

function getCatalogStatusLabel(item: ComponentCatalogItem) {
  if (!item.installed) return '未安装'
  const labels: Record<string, string> = {
    running: '运行中',
    stopped: '已停止',
    error: '错误',
    installing: '安装中',
    unknown: '未知'
  }
  return labels[item.status || 'unknown'] || item.status || '未知'
}
</script>

<style scoped>
.components-page {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
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

.header-tags {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 4px;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
}

.pill-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.pill-success {
  background: rgba(16, 185, 129, 0.1);
  color: #059669;
}
.pill-success .pill-dot {
  background: #10b981;
  box-shadow: 0 0 6px rgba(16, 185, 129, 0.4);
}

.pill-danger {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
}
.pill-danger .pill-dot {
  background: #ef4444;
  box-shadow: 0 0 6px rgba(239, 68, 68, 0.4);
}

.error-hint {
  color: #ef4444;
  font-size: 16px;
  cursor: pointer;
}

.card-grid {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 0 24px 24px;
}

.card-grid::-webkit-scrollbar {
  width: 6px;
}
.card-grid::-webkit-scrollbar-track {
  background: transparent;
}
.card-grid::-webkit-scrollbar-thumb {
  background: #e5e7eb;
  border-radius: 3px;
}
.card-grid::-webkit-scrollbar-thumb:hover {
  background: #d1d5db;
}

.card-col {
  margin-bottom: 20px;
}

/* ========== 组件卡片 ========== */
.component-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  overflow: hidden;
  transition: all 0.3s ease;
  height: 100%;
  display: flex;
  flex-direction: column;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.component-card:hover {
  border-color: #d1d5db;
  box-shadow: 0 10px 40px -10px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.card-top {
  flex: 1;
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.component-name {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  color: #1f2937;
}

.type-tag {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 6px;
  background: rgba(99, 102, 241, 0.1);
  color: #6366f1;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.5px;
  text-transform: uppercase;
}

.type-tag-small {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  background: rgba(99, 102, 241, 0.1);
  color: #6366f1;
}

.component-desc {
  margin: 0 0 16px 0;
  font-size: 13px;
  color: #9ca3af;
  line-height: 1.6;
  min-height: 42px;
}

.component-meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.meta-item .el-icon {
  color: #d1d5db;
  font-size: 14px;
}

.meta-value {
  color: #6b7280;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-bottom {
  padding: 14px 20px;
  border-top: 1px solid #f3f4f6;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

/* 状态徽章 */
.status-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

.status-not-installed {
  background: #f3f4f6;
  color: #9ca3af;
}

.status-running {
  background: rgba(16, 185, 129, 0.1);
  color: #059669;
}

.status-stopped {
  background: rgba(245, 158, 11, 0.1);
  color: #d97706;
}

.status-error {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
}

.status-installing {
  background: rgba(99, 102, 241, 0.1);
  color: #6366f1;
}

.status-unknown {
  background: #f3f4f6;
  color: #9ca3af;
}

/* 操作按钮 */
.card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-group {
  display: flex;
  gap: 4px;
}

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
  height: 32px;
  padding: 0 14px;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  color: #fff;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 15px -4px rgba(99, 102, 241, 0.5);
}

.btn-primary:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-small {
  width: 30px;
  height: 30px;
  background: #f3f4f6;
  color: #9ca3af;
  padding: 0;
}

.btn-small:hover:not(:disabled) {
  background: #e5e7eb;
  color: #6b7280;
}

.btn-small:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.btn-active {
  background: rgba(16, 185, 129, 0.15) !important;
  color: #059669 !important;
}

.btn-danger:hover {
  background: rgba(239, 68, 68, 0.15) !important;
  color: #dc2626 !important;
}

.btn-ghost {
  height: 30px;
  padding: 0 12px;
  background: transparent;
  color: #9ca3af;
  border: 1px solid #e5e7eb;
}

.btn-ghost:hover {
  border-color: #d1d5db;
  color: #6b7280;
}

/* ========== 动态行 ========== */
.dynamic-row {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.add-btn {
  color: #ffffff !important;
  font-size: 13px;
  background-color: #6366f1 !important;
  border: 1px solid #6366f1 !important;
  padding: 6px 14px !important;
  border-radius: 6px !important;
  height: auto !important;
}

/* ========== 信息面板 ========== */
.info-panel {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  overflow: hidden;
}

.info-panel-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 16px;
  background: #f3f4f6;
  border-bottom: 1px solid #e5e7eb;
  font-size: 14px;
  font-weight: 600;
  color: #374151;
}

.info-panel-header .el-icon {
  color: #6366f1;
  font-size: 16px;
}

.info-panel-body {
  padding: 14px 16px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #f3f4f6;
}

.info-label {
  font-size: 12px;
  color: #9ca3af;
}

.info-value {
  font-size: 12px;
  color: #374151;
}

.info-divider {
  height: 1px;
  background: #e5e7eb;
  margin: 14px 0;
}

.info-section-title {
  font-size: 12px;
  font-weight: 600;
  color: #6b7280;
  margin: 12px 0 8px;
}

.info-empty {
  font-size: 12px;
  color: #9ca3af;
  padding: 4px 0;
}

.info-kv {
  font-size: 12px;
  padding: 4px 0;
}

.info-k {
  color: #6366f1;
  font-weight: 500;
}

.info-sep {
  margin: 0 4px;
  color: #d1d5db;
}

.info-v {
  color: #059669;
}

.info-line {
  font-size: 12px;
  color: #6b7280;
  padding: 3px 0;
  word-break: break-all;
}

.info-line.command {
  color: #d97706;
  font-family: 'SF Mono', 'Consolas', monospace;
}
</style>
