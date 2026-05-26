<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20">
      <el-col :span="6" :xs="12" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-glow glow-blue"></div>
          <div class="stat-content">
            <div class="stat-icon icon-blue">
              <el-icon :size="24"><Box /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ overview.total_components }}</div>
              <div class="stat-label">总组件数</div>
            </div>
          </div>
        </div>
      </el-col>

      <el-col :span="6" :xs="12" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-glow glow-green"></div>
          <div class="stat-content">
            <div class="stat-icon icon-green">
              <el-icon :size="24"><CircleCheck /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ overview.running_components }}</div>
              <div class="stat-label">运行中</div>
            </div>
          </div>
        </div>
      </el-col>

      <el-col :span="6" :xs="12" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-glow glow-amber"></div>
          <div class="stat-content">
            <div class="stat-icon icon-amber">
              <el-icon :size="24"><VideoPause /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ overview.stopped_components }}</div>
              <div class="stat-label">已停止</div>
            </div>
          </div>
        </div>
      </el-col>

      <el-col :span="6" :xs="12" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-glow glow-red"></div>
          <div class="stat-content">
            <div class="stat-icon icon-red">
              <el-icon :size="24"><CircleClose /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ overview.error_components }}</div>
              <div class="stat-label">错误</div>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 服务器统计卡片 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="6" :xs="12" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-glow glow-cyan"></div>
          <div class="stat-content">
            <div class="stat-icon icon-cyan">
              <el-icon :size="24"><Monitor /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ servers.length }}</div>
              <div class="stat-label">总服务器数</div>
            </div>
          </div>
        </div>
      </el-col>

      <el-col :span="6" :xs="12" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-glow glow-orange"></div>
          <div class="stat-content">
            <div class="stat-icon icon-orange">
              <el-icon :size="24"><Key /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ servers.filter(s => s.auth_type === 'password').length }}</div>
              <div class="stat-label">密码认证</div>
            </div>
          </div>
        </div>
      </el-col>

      <el-col :span="6" :xs="12" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-glow glow-indigo"></div>
          <div class="stat-content">
            <div class="stat-icon icon-indigo">
              <el-icon :size="24"><Lock /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ servers.filter(s => s.auth_type === 'key').length }}</div>
              <div class="stat-label">密钥认证</div>
            </div>
          </div>
        </div>
      </el-col>

      <el-col :span="6" :xs="12" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-glow glow-teal"></div>
          <div class="stat-content">
            <div class="stat-icon icon-teal">
              <el-icon :size="24"><Folder /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ servers.filter(s => s.deploy_path).length }}</div>
              <div class="stat-label">已配置部署</div>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12" :xs="24" :sm="24" :md="12">
        <div class="chart-card">
          <div class="card-header">
            <div class="header-title">
              <div class="title-dot dot-purple"></div>
              <span>组件类型分布</span>
            </div>
          </div>
          <div ref="pieChartRef" class="chart-body"></div>
        </div>
      </el-col>

      <el-col :span="12" :xs="24" :sm="24" :md="12">
        <div class="chart-card">
          <div class="card-header">
            <div class="header-title">
              <div class="title-dot dot-cyan"></div>
              <span>组件状态分布</span>
            </div>
          </div>
          <div ref="barChartRef" class="chart-body"></div>
        </div>
      </el-col>
    </el-row>

    <!-- 实时监控 -->
    <el-row style="margin-top: 20px">
      <el-col :span="24">
        <div class="table-card">
          <div class="card-header">
            <div class="header-title">
              <div class="title-dot dot-green"></div>
              <span>组件实时监控</span>
            </div>
            <el-button
              :type="realtimeEnabled ? 'danger' : 'primary'"
              size="small"
              class="monitor-btn"
              @click="toggleRealtime"
            >
              <el-icon v-if="!realtimeEnabled"><VideoPlay /></el-icon>
              <el-icon v-else><VideoPause /></el-icon>
              {{ realtimeEnabled ? '停止监控' : '开始监控' }}
            </el-button>
          </div>
          <div class="table-body">
            <el-table
              :data="componentDetails"
              style="width: 100%"
              class="dark-table"
            >
              <el-table-column prop="component.name" label="组件名称" width="120" />
              <el-table-column prop="component.type" label="类型" width="160">
                <template #default="{ row }">
                  <span class="type-tag">{{ row.component.type?.toUpperCase() }}</span>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="160">
                <template #default="{ row }">
                  <span class="status-badge" :class="getStatusClass(row.component.status)">
                    {{ row.component.status }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column label="CPU使用率" min-width="180">
                <template #default="{ row }">
                  <div class="metric-cell">
                    <span class="metric-value">{{ row.stats ? formatPercent(row.stats.cpu_usage) : 'N/A' }}</span>
                    <div v-if="row.stats" class="metric-bar">
                      <div class="metric-bar-fill" :style="{ width: formatPercent(row.stats.cpu_usage), background: getBarColor(row.stats.cpu_usage) }"></div>
                    </div>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="内存使用" min-width="180">
                <template #default="{ row }">
                  <span class="metric-value">
                    {{ row.stats ? formatBytes(row.stats.memory_usage) + ' / ' + formatBytes(row.stats.memory_limit) : 'N/A' }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column label="网络IO" min-width="180">
                <template #default="{ row }">
                  <span class="metric-value" v-if="row.stats">
                    <span class="net-down">↓{{ formatBytes(row.stats.network_rx) }}</span>
                    <span class="net-up">↑{{ formatBytes(row.stats.network_tx) }}</span>
                  </span>
                  <span v-else class="metric-value">N/A</span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-col>
    </el-row>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import type { ECharts } from 'echarts'
import { monitorApi } from '@/api/monitor'
import { serverApi } from '@/api/server'
import type { Overview, ComponentDetail, Server } from '@/types'

const overview = ref<Overview>({
  total_components: 0,
  running_components: 0,
  stopped_components: 0,
  error_components: 0,
  components_by_type: {}
})

const componentDetails = ref<ComponentDetail[]>([])
const servers = ref<Server[]>([])
const realtimeEnabled = ref(false)
let refreshTimer: number | null = null
let pieChart: ECharts | null = null
let barChart: ECharts | null = null
const pieChartRef = ref<HTMLElement>()
const barChartRef = ref<HTMLElement>()

const themeColors = ['#6366f1', '#8b5cf6', '#06b6d4', '#10b981', '#f59e0b', '#ef4444']

// 获取总览数据
async function fetchOverview() {
  try {
    overview.value = await monitorApi.getOverview()
    updateCharts()
  } catch (error) {
    console.error('Failed to fetch overview:', error)
  }
}

// 获取组件详情
async function fetchComponentDetails() {
  try {
    componentDetails.value = await monitorApi.getComponentDetails()
  } catch (error) {
    console.error('Failed to fetch component details:', error)
  }
}

// 获取服务器列表
async function fetchServers() {
  try {
    servers.value = await serverApi.list()
  } catch (error) {
    console.error('Failed to fetch servers:', error)
  }
}

// 初始化图表
function initCharts() {
  if (pieChartRef.value) {
    pieChart = echarts.init(pieChartRef.value)
  }
  if (barChartRef.value) {
    barChart = echarts.init(barChartRef.value)
  }
}

// 更新图表
function updateCharts() {
  const commonTextStyle = { color: '#6b7280', fontSize: 12 }
  const commonTooltip = {
    backgroundColor: 'rgba(255, 255, 255, 0.95)',
    borderColor: '#e5e7eb',
    borderWidth: 1,
    textStyle: { color: '#374151' },
    padding: [10, 14]
  }

  if (pieChart) {
    const data = Object.entries(overview.value.components_by_type).map(
      ([name, value]) => ({ name, value })
    )
    pieChart.setOption({
      backgroundColor: 'transparent',
      tooltip: {
        ...commonTooltip,
        trigger: 'item',
        formatter: '{b}: {c} ({d}%)'
      },
      legend: {
        orient: 'vertical',
        left: 'left',
        textStyle: commonTextStyle,
        itemWidth: 12,
        itemHeight: 12,
        itemGap: 16
      },
      series: [
        {
          name: '组件类型',
          type: 'pie',
          radius: ['40%', '65%'],
          center: ['60%', '50%'],
          data,
          color: themeColors,
          emphasis: {
            itemStyle: {
              shadowBlur: 20,
              shadowColor: 'rgba(99, 102, 241, 0.3)'
            }
          },
          label: {
            color: '#6b7280',
            fontSize: 12
          },
          itemStyle: {
            borderRadius: 6,
            borderColor: '#ffffff',
            borderWidth: 2
          }
        }
      ]
    })
  }

  if (barChart) {
    barChart.setOption({
      backgroundColor: 'transparent',
      tooltip: {
        ...commonTooltip,
        trigger: 'axis',
        axisPointer: { type: 'shadow' }
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        top: '10%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        data: ['运行中', '已停止', '错误'],
        axisLine: { lineStyle: { color: '#e5e7eb' } },
        axisLabel: { color: '#6b7280', fontSize: 12 },
        axisTick: { show: false }
      },
      yAxis: {
        type: 'value',
        axisLine: { show: false },
        axisTick: { show: false },
        axisLabel: { color: '#9ca3af', fontSize: 11 },
        splitLine: { lineStyle: { color: '#f3f4f6', type: 'dashed' } }
      },
      series: [
        {
          name: '组件数量',
          type: 'bar',
          data: [
            {
              value: overview.value.running_components,
              itemStyle: { color: new (echarts as any).graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: '#10b981' },
                { offset: 1, color: 'rgba(16, 185, 129, 0.3)' }
              ]), borderRadius: [6, 6, 0, 0] }
            },
            {
              value: overview.value.stopped_components,
              itemStyle: { color: new (echarts as any).graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: '#f59e0b' },
                { offset: 1, color: 'rgba(245, 158, 11, 0.3)' }
              ]), borderRadius: [6, 6, 0, 0] }
            },
            {
              value: overview.value.error_components,
              itemStyle: { color: new (echarts as any).graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: '#ef4444' },
                { offset: 1, color: 'rgba(239, 68, 68, 0.3)' }
              ]), borderRadius: [6, 6, 0, 0] }
            }
          ],
          barWidth: '40%',
          emphasis: {
            itemStyle: {
              shadowBlur: 15,
              shadowColor: 'rgba(0,0,0,0.1)'
            }
          }
        }
      ]
    })
  }
}

// 切换实时监控
function toggleRealtime() {
  realtimeEnabled.value = !realtimeEnabled.value
  if (realtimeEnabled.value) {
    refreshData()
    refreshTimer = window.setInterval(refreshData, 2000)
  } else {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
  }
}

// 刷新数据
function refreshData() {
  fetchOverview()
  fetchComponentDetails()
  fetchServers()
}

// 状态样式
function getStatusClass(status: string) {
  const classes: Record<string, string> = {
    running: 'status-running',
    stopped: 'status-stopped',
    error: 'status-error',
    installing: 'status-installing'
  }
  return classes[status] || 'status-unknown'
}

// 进度条颜色
function getBarColor(value: number) {
  if (value < 50) return '#10b981'
  if (value < 80) return '#f59e0b'
  return '#ef4444'
}

// 格式化百分比
function formatPercent(value: number) {
  return value.toFixed(2) + '%'
}

// 格式化字节
function formatBytes(bytes: number) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

onMounted(() => {
  fetchOverview()
  fetchComponentDetails()
  fetchServers()
  initCharts()

  window.addEventListener('resize', () => {
    pieChart?.resize()
    barChart?.resize()
  })
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  pieChart?.dispose()
  barChart?.dispose()
})
</script>

<style scoped>
.dashboard {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

/* ========== 统计卡片 ========== */
.stat-card {
  position: relative;
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  padding: 24px;
  overflow: hidden;
  transition: all 0.3s ease;
  cursor: pointer;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.stat-card:hover {
  transform: translateY(-3px);
  border-color: #d1d5db;
  box-shadow: 0 10px 40px -10px rgba(0, 0, 0, 0.1);
}

.stat-glow {
  position: absolute;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  filter: blur(40px);
  opacity: 0.15;
  top: -20px;
  right: -20px;
  transition: opacity 0.3s ease;
}

.stat-card:hover .stat-glow {
  opacity: 0.3;
}

.glow-blue { background: #6366f1; }
.glow-green { background: #10b981; }
.glow-amber { background: #f59e0b; }
.glow-red { background: #ef4444; }
.glow-cyan { background: #06b6d4; }
.glow-orange { background: #f97316; }
.glow-indigo { background: #4f46e5; }
.glow-teal { background: #14b8a6; }

.stat-content {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.icon-blue {
  background: linear-gradient(135deg, #6366f1 0%, #4f46e5 100%);
  box-shadow: 0 4px 15px -4px rgba(99, 102, 241, 0.4);
}

.icon-green {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  box-shadow: 0 4px 15px -4px rgba(16, 185, 129, 0.4);
}

.icon-amber {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  box-shadow: 0 4px 15px -4px rgba(245, 158, 11, 0.4);
}

.icon-cyan {
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
  box-shadow: 0 4px 15px -4px rgba(6, 182, 212, 0.4);
}

.icon-orange {
  background: linear-gradient(135deg, #f97316 0%, #ea580c 100%);
  box-shadow: 0 4px 15px -4px rgba(249, 115, 22, 0.4);
}

.icon-indigo {
  background: linear-gradient(135deg, #4f46e5 0%, #4338ca 100%);
  box-shadow: 0 4px 15px -4px rgba(79, 70, 229, 0.4);
}

.icon-teal {
  background: linear-gradient(135deg, #14b8a6 0%, #0d9488 100%);
  box-shadow: 0 4px 15px -4px rgba(20, 184, 166, 0.4);
}

.icon-red {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  box-shadow: 0 4px 15px -4px rgba(239, 68, 68, 0.4);
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #1f2937;
  line-height: 1.2;
  letter-spacing: -0.5px;
}

.stat-label {
  font-size: 13px;
  color: #9ca3af;
  margin-top: 4px;
}

/* ========== 图表卡片 ========== */
.chart-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #f3f4f6;
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
.dot-cyan { background: #06b6d4; box-shadow: 0 0 8px rgba(6, 182, 212, 0.4); }
.dot-green { background: #10b981; box-shadow: 0 0 8px rgba(16, 185, 129, 0.4); }

.chart-body {
  height: 300px;
  padding: 10px;
}

/* ========== 表格卡片 ========== */
.table-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.table-body {
  padding: 0 4px 4px;
}

.monitor-btn {
  border-radius: 8px;
  font-size: 13px;
}

/* ========== 浅色表格覆盖 ========== */
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
  color: #6b7280 !important;
  font-weight: 500;
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

/* 类型标签 */
.type-tag {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 6px;
  background: rgba(99, 102, 241, 0.1);
  color: #6366f1;
  font-size: 12px;
  font-weight: 500;
}

/* 状态徽章 */
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

/* 指标单元格 */
.metric-cell {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.metric-value {
  font-size: 13px;
  color: #6b7280;
  font-family: 'SF Mono', 'Consolas', monospace;
}

.metric-bar {
  width: 60px;
  height: 4px;
  background: #f3f4f6;
  border-radius: 2px;
  overflow: hidden;
}

.metric-bar-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.5s ease;
}

.net-down {
  color: #059669;
  margin-right: 10px;
}

.net-up {
  color: #dc2626;
}

/* 滚动条 */
.dashboard::-webkit-scrollbar {
  width: 6px;
}

.dashboard::-webkit-scrollbar-track {
  background: transparent;
}

.dashboard::-webkit-scrollbar-thumb {
  background: #e5e7eb;
  border-radius: 3px;
}

.dashboard::-webkit-scrollbar-thumb:hover {
  background: #d1d5db;
}

@media (max-width: 768px) {
  .dashboard {
    padding: 16px;
  }
}
</style>
