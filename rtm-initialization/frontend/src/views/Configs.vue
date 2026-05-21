<template>
  <div class="configs-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <div class="header-title-section">
          <h2>配置管理</h2>
          <p class="subtitle">查看和编辑已安装组件的运行配置</p>
        </div>
      </div>
    </div>

    <!-- 表格 -->
    <div class="table-card">
      <div class="card-header">
        <div class="header-title">
          <div class="title-dot dot-purple"></div>
          <span>组件配置列表</span>
        </div>
      </div>
      <div class="table-body">
        <el-table
          :data="componentStore.components"
          v-loading="componentStore.loading"
          style="width: 100%"
          class="dark-table"
          element-loading-background="rgba(245, 247, 250, 0.8)"
          element-loading-text="加载中..."
        >
        <el-table-column prop="name" label="组件名称" width="120">
          <template #default="{ row }">
            <span class="cell-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="160">
          <template #default="{ row }">
            <span class="type-tag">{{ row.type }}</span>
          </template>
        </el-table-column>
        <el-table-column label="端口" width="120">
          <template #default="{ row }">
            <span class="cell-mono">{{ row.config?.port || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="环境变量摘要" min-width="220">
          <template #default="{ row }">
            <span v-if="!row.config?.env || !Object.keys(row.config.env).length" class="cell-empty">-</span>
            <span v-else class="env-summary">
              {{ formatEnvSummary(row.config.env) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="数据卷摘要" min-width="220">
          <template #default="{ row }">
            <span v-if="!row.config?.volumes?.length" class="cell-empty">-</span>
            <span v-else class="volume-summary">
              {{ formatVolumeSummary(row.config.volumes) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160">
          <template #default="{ row }">
            <button class="action-btn btn-edit" @click="openEditDialog(row)">
              <el-icon><Edit /></el-icon>编辑
            </button>
          </template>
        </el-table-column>
      </el-table>
      </div>
    </div>

    <!-- 编辑配置对话框 -->
    <el-dialog
      v-model="showEditDialog"
      title="编辑组件配置"
      width="900px"
      class="dark-dialog"
      destroy-on-close
    >
      <el-alert
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 20px"
        class="dark-alert"
      >
        配置已保存，重启容器后生效
      </el-alert>

      <el-row :gutter="24">
        <!-- 左侧表单 -->
        <el-col :span="15">
          <el-form :model="editForm" label-width="90px" class="dark-form">
            <el-form-item label="端口">
              <el-input v-model="editForm.port" placeholder="如5432" />
            </el-form-item>

            <el-form-item label="环境变量">
              <div
                v-for="(item, index) in editForm.envList"
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
                v-for="(_, index) in editForm.volumesList"
                :key="index"
                class="dynamic-row"
              >
                <el-input
                  v-model="editForm.volumesList[index]"
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
                v-for="(_, index) in editForm.commandList"
                :key="index"
                class="dynamic-row"
              >
                <el-input
                  v-model="editForm.commandList[index]"
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
          <div class="info-panel" v-if="currentComponent">
            <div class="info-panel-header">
              <el-icon><InfoFilled /></el-icon>
              <span>当前配置信息</span>
            </div>
            <div class="info-panel-body">
              <div class="info-row">
                <span class="info-label">组件名称</span>
                <span class="info-value">{{ currentComponent.name }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">组件类型</span>
                <span class="info-value type-tag-small">{{ currentComponent.type }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">当前镜像</span>
                <span class="info-value">{{ currentComponent.image }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">当前端口</span>
                <span class="info-value">{{ currentComponent.config?.port || '-' }}</span>
              </div>

              <div class="info-divider"></div>

              <div class="info-section-title">当前环境变量</div>
              <div v-if="!currentEnvList.length" class="info-empty">无</div>
              <div v-for="(e, i) in currentEnvList" :key="i" class="info-kv">
                <span class="info-k">{{ e.key }}</span>
                <span class="info-sep">=</span>
                <span class="info-v">{{ e.value }}</span>
              </div>

              <div class="info-section-title">当前数据卷</div>
              <div v-if="!currentVolumesList.length" class="info-empty">无</div>
              <div v-for="(v, i) in currentVolumesList" :key="i" class="info-line">
                {{ v }}
              </div>

              <div class="info-section-title">当前启动命令参数</div>
              <div v-if="!currentCommandList.length" class="info-empty">无</div>
              <div v-for="(c, i) in currentCommandList" :key="i" class="info-line command">
                {{ c }}
              </div>
            </div>
          </div>
        </el-col>
      </el-row>

      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveConfig">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import { useComponentStore } from '@/stores/component'
import { componentApi } from '@/api/component'
import type { Component, ComponentConfig } from '@/types'

const componentStore = useComponentStore()

const showEditDialog = ref(false)
const currentComponent = ref<Component | null>(null)

interface EnvItem {
  key: string
  value: string
}

const editForm = reactive({
  port: '',
  envList: [] as EnvItem[],
  volumesList: [] as string[],
  commandList: [] as string[]
})

const currentEnvList = computed(() => {
  if (!currentComponent.value?.config?.env) return []
  return Object.entries(currentComponent.value.config.env).map(([key, value]) => ({
    key,
    value: String(value)
  }))
})

const currentVolumesList = computed(() => {
  return currentComponent.value?.config?.volumes || []
})

const currentCommandList = computed(() => {
  return currentComponent.value?.config?.command || []
})

onMounted(() => {
  componentStore.fetchComponents()
})

function formatEnvSummary(env: Record<string, string>) {
  const entries = Object.entries(env).slice(0, 2)
  const parts = entries.map(([k, v]) => `${k}=${v}`)
  const suffix = Object.keys(env).length > 2 ? ` 等${Object.keys(env).length}项` : ''
  return parts.join(', ') + suffix
}

function formatVolumeSummary(volumes: string[]) {
  if (!volumes.length) return '-'
  const first = volumes[0]
  if (volumes.length === 1) return first
  return `${first} 等${volumes.length}项`
}

function openEditDialog(component: Component) {
  currentComponent.value = component
  editForm.port = component.config?.port || ''

  const env = component.config?.env || {}
  editForm.envList = Object.entries(env).map(([key, value]) => ({
    key,
    value: String(value)
  }))

  editForm.volumesList = component.config?.volumes
    ? [...component.config.volumes]
    : []

  editForm.commandList = component.config?.command
    ? [...component.config.command]
    : []

  showEditDialog.value = true
}

function addEnv() {
  editForm.envList.push({ key: '', value: '' })
}

function removeEnv(index: number) {
  editForm.envList.splice(index, 1)
}

function addVolume() {
  editForm.volumesList.push('')
}

function removeVolume(index: number) {
  editForm.volumesList.splice(index, 1)
}

function addCommandArg() {
  editForm.commandList.push('')
}

function removeCommandArg(index: number) {
  editForm.commandList.splice(index, 1)
}

async function handleSaveConfig() {
  if (!currentComponent.value) return

  const env: Record<string, string> = {}
  editForm.envList.forEach(({ key, value }) => {
    if (key.trim()) {
      env[key.trim()] = value
    }
  })

  const config: ComponentConfig = {
    port: editForm.port,
    env,
    volumes: editForm.volumesList.filter(v => v.trim()),
    command: editForm.commandList.filter(v => v.trim())
  }

  try {
    await componentApi.updateConfig(currentComponent.value.id, config)
    ElMessage.success('配置更新成功')
    showEditDialog.value = false
    await componentStore.fetchComponents()
  } catch (error) {
    ElMessage.error('配置更新失败')
  }
}
</script>

<style scoped>
.configs-page {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-shrink: 0;
  padding: 24px 24px 0;
}

.header-left {
  display: flex;
  align-items: center;
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

/* ========== 卡片式表格 ========== */
.table-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  flex: 1;
  margin: 0 24px 24px;
  display: flex;
  flex-direction: column;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #f3f4f6;
  flex-shrink: 0;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 15px;
  font-weight: 600;
  color: #374151;
}

.title-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot-purple { background: #8b5cf6; box-shadow: 0 0 8px rgba(139, 92, 246, 0.4); }

.table-body {
  padding: 0 4px 4px;
  flex: 1;
  overflow: auto;
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

.cell-name {
  font-weight: 500;
  color: #374151;
  font-size: 14px;
}

.cell-mono {
  font-family: 'SF Mono', 'Consolas', monospace;
  color: #6b7280;
  font-size: 13px;
}

.cell-empty {
  color: #d1d5db;
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

.env-summary,
.volume-summary {
  font-size: 13px;
  color: #6b7280;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 3px;
  padding: 4px 8px;
  border-radius: 6px;
  border: none;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
  white-space: nowrap;
  line-height: 1;
}

.btn-edit {
  background: rgba(99, 102, 241, 0.1);
  color: #6366f1;
}

.btn-edit:hover {
  background: rgba(99, 102, 241, 0.2);
  color: #818cf8;
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
