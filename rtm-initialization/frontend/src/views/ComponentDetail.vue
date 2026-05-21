<template>
  <div class="component-detail">
    <!-- 页面头部 -->
    <div class="detail-header">
      <div class="back-btn" @click="$router.back()">
        <el-icon><ArrowLeft /></el-icon>
        <span>返回</span>
      </div>
      <div class="header-center">
        <h2>{{ component?.name || '组件详情' }}</h2>
        <span v-if="component" class="detail-status" :class="getStatusClass(component.status)">
          {{ component.status }}
        </span>
      </div>
      <div class="header-spacer"></div>
    </div>

    <div class="detail-body">
      <el-tabs v-model="activeTab" class="dark-tabs">
        <!-- 基本信息 -->
        <el-tab-pane label="基本信息" name="info">
          <div class="info-card">
            <div class="info-grid">
              <div class="info-cell">
                <div class="info-cell-label">组件名称</div>
                <div class="info-cell-value">{{ component?.name || 'N/A' }}</div>
              </div>
              <div class="info-cell">
                <div class="info-cell-label">组件类型</div>
                <div class="info-cell-value">
                  <span class="type-tag">{{ component?.type || 'N/A' }}</span>
                </div>
              </div>
              <div class="info-cell">
                <div class="info-cell-label">镜像</div>
                <div class="info-cell-value">{{ component?.image || 'N/A' }}</div>
              </div>
              <div class="info-cell">
                <div class="info-cell-label">版本</div>
                <div class="info-cell-value">{{ component?.version || 'N/A' }}</div>
              </div>
              <div class="info-cell">
                <div class="info-cell-label">状态</div>
                <div class="info-cell-value">
                  <span class="status-badge" :class="getStatusClass(component?.status)">
                    {{ component?.status || 'N/A' }}
                  </span>
                </div>
              </div>
              <div class="info-cell">
                <div class="info-cell-label">容器ID</div>
                <div class="info-cell-value mono">{{ component?.container_id || 'N/A' }}</div>
              </div>
              <div class="info-cell">
                <div class="info-cell-label">创建时间</div>
                <div class="info-cell-value">{{ formatDate(component?.created_at) }}</div>
              </div>
              <div class="info-cell">
                <div class="info-cell-label">更新时间</div>
                <div class="info-cell-value">{{ formatDate(component?.updated_at) }}</div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- 配置信息 -->
        <el-tab-pane label="配置信息" name="config">
          <div class="config-card">
            <pre class="config-display">{{ JSON.stringify(component?.config, null, 2) }}</pre>
          </div>
        </el-tab-pane>

        <!-- 资源监控 -->
        <el-tab-pane label="资源监控" name="stats">
          <div v-if="stats" class="stats-grid">
            <!-- CPU -->
            <div class="stats-card">
              <div class="stats-card-header">
                <div class="stats-icon icon-cpu">
                  <el-icon><Cpu /></el-icon>
                </div>
                <div class="stats-title">CPU 使用率</div>
              </div>
              <div class="stats-value">{{ formatPercent(stats.cpu_usage) }}</div>
              <div class="stats-progress">
                <div class="progress-track">
                  <div
                    class="progress-fill"
                    :class="getProgressClass(safePercent(stats.cpu_usage, 100))"
                    :style="{ width: safePercent(stats.cpu_usage, 100) + '%' }"
                  ></div>
                </div>
              </div>
            </div>

            <!-- 内存 -->
            <div class="stats-card">
              <div class="stats-card-header">
                <div class="stats-icon icon-mem">
                  <el-icon><Memo /></el-icon>
                </div>
                <div class="stats-title">内存使用</div>
              </div>
              <div class="stats-value">
                {{ formatBytes(stats.memory_usage) }}
                <span class="stats-total">/ {{ formatBytes(stats.memory_limit) }}</span>
              </div>
              <div class="stats-progress">
                <div class="progress-track">
                  <div
                    class="progress-fill"
                    :class="getProgressClass(safePercent(stats.memory_usage, stats.memory_limit))"
                    :style="{ width: safePercent(stats.memory_usage, stats.memory_limit) + '%' }"
                  ></div>
                </div>
              </div>
            </div>

            <!-- 网络接收 -->
            <div class="stats-card">
              <div class="stats-card-header">
                <div class="stats-icon icon-net-down">
                  <el-icon><Download /></el-icon>
                </div>
                <div class="stats-title">网络接收</div>
              </div>
              <div class="stats-value net-down-text">↓ {{ formatBytes(stats.network_rx) }}</div>
            </div>

            <!-- 网络发送 -->
            <div class="stats-card">
              <div class="stats-card-header">
                <div class="stats-icon icon-net-up">
                  <el-icon><Upload /></el-icon>
                </div>
                <div class="stats-title">网络发送</div>
              </div>
              <div class="stats-value net-up-text">↑ {{ formatBytes(stats.network_tx) }}</div>
            </div>

            <!-- 磁盘读取 -->
            <div class="stats-card">
              <div class="stats-card-header">
                <div class="stats-icon icon-disk">
                  <el-icon><Reading /></el-icon>
                </div>
                <div class="stats-title">磁盘读取</div>
              </div>
              <div class="stats-value">{{ formatBytes(stats.block_read) }}</div>
            </div>

            <!-- 磁盘写入 -->
            <div class="stats-card">
              <div class="stats-card-header">
                <div class="stats-icon icon-disk">
                  <el-icon><Edit /></el-icon>
                </div>
                <div class="stats-title">磁盘写入</div>
              </div>
              <div class="stats-value">{{ formatBytes(stats.block_write) }}</div>
            </div>
          </div>
          <div v-else class="empty-state">
            <el-icon :size="48"><DataLine /></el-icon>
            <p>暂无监控数据</p>
          </div>
        </el-tab-pane>

        <!-- 日志查看 -->
        <el-tab-pane label="日志查看" name="logs">
          <div class="log-card">
            <div class="log-toolbar">
              <div class="log-actions">
                <button class="toolbar-btn" @click="fetchLogs">
                  <el-icon><Refresh /></el-icon>
                  刷新日志
                </button>
                <button
                  class="toolbar-btn"
                  :class="{ 'btn-active': autoRefresh }"
                  @click="toggleAutoRefresh"
                >
                  <el-icon><Timer /></el-icon>
                  {{ autoRefresh ? '停止刷新' : '自动刷新' }}
                </button>
              </div>
              <el-select v-model="logTail" size="small" class="dark-select" style="width: 130px">
                <el-option label="最近100行" value="100" />
                <el-option label="最近500行" value="500" />
                <el-option label="最近1000行" value="1000" />
              </el-select>
            </div>
            <div class="log-container">
              <pre class="log-content">{{ logs || '暂无日志' }}</pre>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { componentApi } from '@/api/component'
import type { Component, ComponentStats, ComponentStatus } from '@/types'

const route = useRoute()
const componentId = route.params.id as string

const activeTab = ref('info')
const component = ref<Component | null>(null)
const stats = ref<ComponentStats | null>(null)
const logs = ref('')
const logTail = ref('100')
const autoRefresh = ref(false)
let refreshTimer: number | null = null

async function fetchComponent() {
  try {
    component.value = await componentApi.get(componentId)
  } catch (error) {
    ElMessage.error('获取组件详情失败')
  }
}

async function fetchStats() {
  try {
    stats.value = await componentApi.getStats(componentId)
  } catch (error) {
    console.error('Failed to fetch stats:', error)
  }
}

async function fetchLogs() {
  try {
    const result = await componentApi.getLogs(componentId, logTail.value)
    logs.value = result.logs
  } catch (error) {
    ElMessage.error('获取日志失败')
  }
}

function toggleAutoRefresh() {
  autoRefresh.value = !autoRefresh.value
  if (autoRefresh.value) {
    refreshData()
    refreshTimer = window.setInterval(refreshData, 5000)
  } else {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
  }
}

function refreshData() {
  fetchStats()
  if (activeTab.value === 'logs') {
    fetchLogs()
  }
}

function getStatusClass(status: ComponentStatus | undefined) {
  if (!status) return 'status-unknown'
  const classes: Record<ComponentStatus, string> = {
    running: 'status-running',
    stopped: 'status-stopped',
    error: 'status-error',
    installing: 'status-installing'
  }
  return classes[status] || 'status-unknown'
}

function safeRatio(numerator: number | undefined | null, denominator: number | undefined | null) {
  if (numerator == null || denominator == null || Number.isNaN(numerator) || Number.isNaN(denominator)) return 0
  if (denominator <= 0) return 0
  return Math.min(numerator / denominator, 1)
}

function safePercent(numerator: number | undefined | null, denominator: number | undefined | null) {
  return safeRatio(numerator, denominator) * 100
}

function formatPercent(value: number | undefined | null) {
  if (value == null || Number.isNaN(value)) return '-'
  return value.toFixed(2) + '%'
}

function formatBytes(bytes: number | undefined | null) {
  if (bytes == null || Number.isNaN(bytes) || bytes < 0) return '-'
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(k)), sizes.length - 1)
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

function getProgressClass(percentage: number) {
  if (percentage < 50) return 'fill-green'
  if (percentage < 80) return 'fill-amber'
  return 'fill-red'
}

function formatDate(dateString: string | undefined) {
  if (!dateString) return 'N/A'
  return new Date(dateString).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchComponent()
  fetchStats()
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.component-detail {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* ========== 头部 ========== */
.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  border-bottom: 1px solid #f3f4f6;
  flex-shrink: 0;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #9ca3af;
  font-size: 14px;
  cursor: pointer;
  transition: color 0.2s ease;
  width: 80px;
}

.back-btn:hover {
  color: #6b7280;
}

.header-center {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  justify-content: center;
}

.header-center h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.header-spacer {
  width: 80px;
}

.detail-status {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

/* ========== 标签页 ========== */
.detail-body {
  flex: 1;
  overflow: hidden;
  padding: 0 24px 24px;
}

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

:deep(.dark-tabs .el-tab-pane) {
  height: calc(100% - 64px);
  overflow-y: auto;
}

:deep(.dark-tabs .el-tab-pane::-webkit-scrollbar) {
  width: 6px;
}
:deep(.dark-tabs .el-tab-pane::-webkit-scrollbar-thumb) {
  background: #e5e7eb;
  border-radius: 3px;
}

/* ========== 信息卡片 ========== */
.info-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

@media (max-width: 768px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
}

.info-cell {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 14px 16px;
  background: #f9fafb;
  border-radius: 12px;
  border: 1px solid #f3f4f6;
}

.info-cell-label {
  font-size: 12px;
  color: #9ca3af;
  font-weight: 500;
}

.info-cell-value {
  font-size: 14px;
  color: #374151;
  word-break: break-all;
}

.info-cell-value.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 13px;
}

/* ========== 配置卡片 ========== */
.config-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.config-display {
  margin: 0;
  padding: 20px 24px;
  background: #f9fafb;
  font-size: 13px;
  overflow-x: auto;
  color: #374151;
  font-family: 'SF Mono', 'Consolas', monospace;
  line-height: 1.7;
}

/* ========== 资源监控网格 ========== */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}

.stats-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  padding: 20px;
  transition: all 0.3s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.stats-card:hover {
  border-color: #d1d5db;
}

.stats-card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}

.stats-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 18px;
}

.icon-cpu {
  background: linear-gradient(135deg, #6366f1, #4f46e5);
  box-shadow: 0 4px 15px -4px rgba(99, 102, 241, 0.4);
}

.icon-mem {
  background: linear-gradient(135deg, #8b5cf6, #7c3aed);
  box-shadow: 0 4px 15px -4px rgba(139, 92, 246, 0.4);
}

.icon-net-down {
  background: linear-gradient(135deg, #10b981, #059669);
  box-shadow: 0 4px 15px -4px rgba(16, 185, 129, 0.4);
}

.icon-net-up {
  background: linear-gradient(135deg, #f59e0b, #d97706);
  box-shadow: 0 4px 15px -4px rgba(245, 158, 11, 0.4);
}

.icon-disk {
  background: linear-gradient(135deg, #06b6d4, #0891b2);
  box-shadow: 0 4px 15px -4px rgba(6, 182, 212, 0.4);
}

.stats-title {
  font-size: 13px;
  color: #9ca3af;
  font-weight: 500;
}

.stats-value {
  font-size: 24px;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 16px;
  letter-spacing: -0.5px;
}

.stats-total {
  font-size: 14px;
  font-weight: 400;
  color: #d1d5db;
}

.net-down-text {
  color: #059669;
}

.net-up-text {
  color: #d97706;
}

.stats-progress {
  margin-top: 8px;
}

.progress-track {
  width: 100%;
  height: 6px;
  background: #f3f4f6;
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.5s ease;
}

.fill-green { background: linear-gradient(90deg, #10b981, #34d399); }
.fill-amber { background: linear-gradient(90deg, #f59e0b, #fbbf24); }
.fill-red { background: linear-gradient(90deg, #ef4444, #f87171); }

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px;
  color: #d1d5db;
  gap: 12px;
}

.empty-state p {
  margin: 0;
  font-size: 14px;
  color: #9ca3af;
}

/* ========== 日志卡片 ========== */
.log-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  height: calc(100vh - 220px);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.log-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  border-bottom: 1px solid #f3f4f6;
  flex-shrink: 0;
}

.log-actions {
  display: flex;
  gap: 8px;
}

.toolbar-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
  background: #f9fafb;
  color: #6b7280;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
}

.toolbar-btn:hover {
  background: #f3f4f6;
  color: #374151;
}

.toolbar-btn.btn-active {
  background: rgba(99, 102, 241, 0.1);
  border-color: rgba(99, 102, 241, 0.3);
  color: #6366f1;
}

.log-container {
  flex: 1;
  overflow: auto;
  background: #f9fafb;
}

.log-container::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.log-container::-webkit-scrollbar-track {
  background: transparent;
}

.log-container::-webkit-scrollbar-thumb {
  background: #e5e7eb;
  border-radius: 4px;
}

.log-content {
  margin: 0;
  padding: 20px 24px;
  color: #374151;
  font-family: 'SF Mono', 'Consolas', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.7;
  min-height: 100%;
  white-space: pre-wrap;
  word-break: break-all;
}

/* ========== 通用状态徽章 ========== */
.status-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
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

.type-tag {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 6px;
  background: rgba(99, 102, 241, 0.1);
  color: #6366f1;
  font-size: 12px;
  font-weight: 500;
}

/* 深色选择器 */
:deep(.dark-select .el-input__wrapper) {
  background: #f9fafb !important;
  border: 1px solid #e5e7eb !important;
  box-shadow: none !important;
}

:deep(.dark-select .el-input__inner) {
  color: #374151 !important;
}
</style>
