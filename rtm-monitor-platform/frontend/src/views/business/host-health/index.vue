<template>
  <div class="app-container">
    <el-row :gutter="15">
      <!-- 左侧主机列表 -->
      <el-col :span="4">
        <div class="host-list-panel">
          <div class="panel-title">主机列表</div>
          <el-input v-model="filterText" placeholder="搜索主机IP/名称" clearable size="small" style="margin-bottom: 10px;" />
          <el-scrollbar :height="listHeight">
            <div v-for="host in filteredHosts" :key="host.id" class="host-item"
              :class="{ active: selectedIp === host.ip }" @click="selectHost(host)">
              <div class="host-item-name">{{ host.name || '未命名' }}</div>
              <div class="host-item-ip">{{ host.ip }}</div>
              <el-tag v-if="host.isOnline === true" type="success" size="small">在线</el-tag>
              <el-tag v-else-if="host.isOnline === false" type="danger" size="small">离线</el-tag>
              <el-tag v-else type="info" size="small">未知</el-tag>
            </div>
          </el-scrollbar>
        </div>
      </el-col>

      <!-- 右侧图表区域 -->
      <el-col :span="20">
        <div class="chart-panel">
          <!-- 查询条件 -->
          <div class="query-bar">
            <el-radio-group v-model="queryLevel" size="small" @change="onLevelChange">
              <el-radio-button label="1">日期级别（近15天）</el-radio-button>
              <el-radio-button label="2">小时级别</el-radio-button>
              <el-radio-button label="3">分钟级别</el-radio-button>
            </el-radio-group>
            <el-date-picker v-if="queryLevel !== '1'" v-model="queryDate" type="date" placeholder="选择日期"
              value-format="YYYY-MM-DD" size="small" style="margin-left: 10px;" />
            <el-select v-if="queryLevel === '3'" v-model="queryHour" placeholder="选择小时" size="small"
              style="width: 120px; margin-left: 10px;">
              <el-option v-for="h in 24" :key="h - 1" :label="String(h - 1).padStart(2, '0') + ' 时'"
                :value="String(h - 1).padStart(2, '0')" />
            </el-select>
            <el-button type="primary" size="small" icon="Search" style="margin-left: 10px;" @click="queryData"
              :disabled="!selectedIp">查询</el-button>
          </div>

          <!-- 当前选中主机信息 -->
          <div v-if="selectedHost" class="host-info-bar">
            <el-descriptions :column="6" size="small" border>
              <el-descriptions-item label="IP">{{ selectedHost.ip }}</el-descriptions-item>
              <el-descriptions-item label="名称">{{ selectedHost.name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="操作系统">{{ selectedHost.osType || '-' }}</el-descriptions-item>
              <el-descriptions-item label="CPU">{{ selectedHost.cpuCores ? selectedHost.cpuCores + ' 核' : '-' }}</el-descriptions-item>
              <el-descriptions-item label="内存">{{ formatMem(selectedHost.memTotalMb) }}</el-descriptions-item>
              <el-descriptions-item label="磁盘">{{ selectedHost.diskTotalGb ? selectedHost.diskTotalGb + ' GB' : '-' }}</el-descriptions-item>
            </el-descriptions>
          </div>

          <!-- 图表区域 -->
          <div v-loading="loading" class="charts-wrapper">
            <!-- 空数据提示 -->
            <div v-if="noData" class="empty-data-tip">
              <el-empty description="暂无监控数据">
                <template #description>
                  <div style="color: #909399; font-size: 14px;">
                    <p>暂无监控数据</p>
                    <p style="font-size: 12px; margin-top: 8px;">
                      该主机在选定时间段内没有上报健康度数据
                    </p>
                  </div>
                </template>
              </el-empty>
            </div>
            <template v-else>
              <el-row :gutter="15">
                <el-col :span="12">
                  <div ref="chartCpu" class="chart-box" />
                </el-col>
                <el-col :span="12">
                  <div ref="chartMem" class="chart-box" />
                </el-col>
              </el-row>
              <el-row :gutter="15" style="margin-top: 15px;">
                <el-col :span="12">
                  <div ref="chartDisk" class="chart-box" />
                </el-col>
                <el-col :span="12">
                  <div ref="chartNetIn" class="chart-box" />
                </el-col>
              </el-row>
              <el-row :gutter="15" style="margin-top: 15px;">
                <el-col :span="12">
                  <div ref="chartNetOut" class="chart-box" />
                </el-col>
                <el-col :span="12">
                  <div ref="chartMemMb" class="chart-box" />
                </el-col>
              </el-row>
            </template>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup name="HostHealth">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import * as echarts from 'echarts'
import { listHostInfo } from "@/api/business/hostInfo"
import { hostHistory } from "@/api/business/hostHealth"

const route = useRoute()
const { proxy } = getCurrentInstance()

// 主机列表
const hostList = ref([])
const filterText = ref('')
const selectedIp = ref('')
const selectedHost = ref(null)
const loading = ref(false)
const noData = ref(false)

// 查询条件
const queryLevel = ref('1')
const queryDate = ref(null)
const queryHour = ref(null)

// 图表引用
const chartCpu = ref(null)
const chartMem = ref(null)
const chartDisk = ref(null)
const chartNetIn = ref(null)
const chartNetOut = ref(null)
const chartMemMb = ref(null)

// 图表实例
let charts = {}
const listHeight = ref('calc(100vh - 236px)')

// 过滤主机列表
const filteredHosts = computed(() => {
  if (!filterText.value) return hostList.value
  const text = filterText.value.toLowerCase()
  return hostList.value.filter(h =>
    (h.ip && h.ip.toLowerCase().includes(text)) ||
    (h.name && h.name.toLowerCase().includes(text))
  )
})

/** 格式化内存 */
function formatMem(mb) {
  if (!mb || mb <= 0) return '-'
  if (mb >= 1024) return (mb / 1024).toFixed(2) + ' GB'
  return mb + ' MB'
}

/** 检查数据是否全部为空/零 */
function isAllEmpty(data) {
  const arrays = [
    data.cpuUsages, data.memUsages, data.diskUsages,
    data.netIns, data.netOuts, data.memUsageMBs
  ]
  for (const arr of arrays) {
    if (Array.isArray(arr) && arr.length > 0) {
      for (const v of arr) {
        if (v !== 0 && v !== 0.0 && v !== '0') {
          return false
        }
      }
    }
  }
  return true
}

/** 加载主机列表 */
function loadHosts() {
  listHostInfo({ pageNum: 1, pageSize: 9999 }).then(response => {
    const pageData = response.data || response
    hostList.value = pageData.rows || []
    if (hostList.value.length === 0) return
    // 如果 URL 带了 ip 参数，自动选中该主机
    const queryIp = route.query.ip
    if (queryIp) {
      const found = hostList.value.find(h => h.ip === queryIp)
      if (found) {
        selectHost(found)
        return
      }
    }
    // 否则默认选中第一个
    if (!selectedIp.value) {
      selectHost(hostList.value[0])
    }
  }).catch(err => {
    proxy.$modal.msgError('加载主机列表失败：' + (err.message || '未知错误'))
  })
}

/** 选择主机 */
function selectHost(host) {
  selectedIp.value = host.ip
  selectedHost.value = host
  nextTick(() => {
    initCharts()
    queryData()
  })
}

/** 级别切换 */
function onLevelChange() {
  queryDate.value = null
  queryHour.value = null
}

/** 初始化图表 */
function initCharts() {
  disposeCharts()
  charts.cpu = echarts.init(chartCpu.value)
  charts.mem = echarts.init(chartMem.value)
  charts.disk = echarts.init(chartDisk.value)
  charts.netIn = echarts.init(chartNetIn.value)
  charts.netOut = echarts.init(chartNetOut.value)
  charts.memMb = echarts.init(chartMemMb.value)
}

/** 销毁图表 */
function disposeCharts() {
  Object.values(charts).forEach(c => c?.dispose())
  charts = {}
}

/** 查询数据 */
function queryData() {
  if (!selectedIp.value) {
    proxy.$modal.msgWarning('请先选择主机')
    return
  }
  loading.value = true
  noData.value = false
  hostHistory({
    ip: selectedIp.value,
    queryLevel: queryLevel.value,
    queryDate: queryDate.value,
    queryHour: queryHour.value
  }).then(response => {
    const data = response.data || response
    if (!data.times || data.times.length === 0 || isAllEmpty(data)) {
      noData.value = true
      // 清空图表
      Object.values(charts).forEach(c => c?.clear())
    } else {
      noData.value = false
      updateCharts(data)
    }
    loading.value = false
  }).catch(err => {
    loading.value = false
    noData.value = true
    proxy.$modal.msgError('查询监控数据失败：' + (err.message || '未知错误'))
  })
}

/** 更新图表 */
function updateCharts(data) {
  const times = data.times || []
  const commonOption = {
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', top: '15%', containLabel: true },
    toolbox: { feature: { saveAsImage: {} } },
    xAxis: { type: 'category', boundaryGap: false, data: times },
    yAxis: { type: 'value', min: 0 }
  }

  charts.cpu?.setOption({
    ...commonOption,
    title: { text: 'CPU使用率（%）', left: 'center', textStyle: { fontSize: 14 } },
    series: [{ name: 'CPU使用率', type: 'line', smooth: true, data: data.cpuUsages || [], areaStyle: { opacity: 0.2 } }]
  }, true)

  charts.mem?.setOption({
    ...commonOption,
    title: { text: '内存使用率（%）', left: 'center', textStyle: { fontSize: 14 } },
    series: [{ name: '内存使用率', type: 'line', smooth: true, data: data.memUsages || [], areaStyle: { opacity: 0.2 } }]
  }, true)

  charts.disk?.setOption({
    ...commonOption,
    title: { text: '磁盘使用率（%）', left: 'center', textStyle: { fontSize: 14 } },
    series: [{ name: '磁盘使用率', type: 'line', smooth: true, data: data.diskUsages || [], areaStyle: { opacity: 0.2 } }]
  }, true)

  charts.netIn?.setOption({
    ...commonOption,
    title: { text: '网络入流量（KB/s）', left: 'center', textStyle: { fontSize: 14 } },
    series: [{ name: '网络入流量', type: 'line', smooth: true, data: data.netIns || [], areaStyle: { opacity: 0.2 } }]
  }, true)

  charts.netOut?.setOption({
    ...commonOption,
    title: { text: '网络出流量（KB/s）', left: 'center', textStyle: { fontSize: 14 } },
    series: [{ name: '网络出流量', type: 'line', smooth: true, data: data.netOuts || [], areaStyle: { opacity: 0.2 } }]
  }, true)

  charts.memMb?.setOption({
    ...commonOption,
    title: { text: '内存已使用（MB）', left: 'center', textStyle: { fontSize: 14 } },
    series: [{ name: '内存已使用', type: 'line', smooth: true, data: data.memUsageMBs || [], areaStyle: { opacity: 0.2 } }]
  }, true)
}

// 窗口大小变化
let resizeTimer = null
function handleResize() {
  if (resizeTimer) clearTimeout(resizeTimer)
  resizeTimer = setTimeout(() => {
    Object.values(charts).forEach(c => c?.resize())
  }, 100)
}

onMounted(() => {
  loadHosts()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  disposeCharts()
})
</script>

<style scoped lang="scss">
.app-container {
  min-height: calc(100vh - 84px);
  padding: 24px;
}

.host-list-panel {
  background: #FFFFFF;
  padding: 20px;
  border-radius: 12px;
  height: calc(100vh - 144px);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 1px 2px rgba(0, 0, 0, 0.03);
  border: 1px solid #E2E8F0;

  .panel-title {
    font-weight: 700;
    font-size: 16px;
    margin-bottom: 14px;
    color: #1E293B;
    letter-spacing: -0.3px;
  }
}

.host-item {
  padding: 12px 14px;
  margin-bottom: 10px;
  border-radius: 10px;
  cursor: pointer;
  border: 1px solid #E2E8F0;
  transition: all 0.2s ease;

  &:hover {
    background-color: #F8FAFC;
    border-color: #CBD5E1;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.04);
  }

  &.active {
    background-color: #EFF6FF;
    border-color: #2563EB;
    box-shadow: 0 2px 8px rgba(37, 99, 235, 0.1);

    .host-item-name {
      color: #2563EB;
    }
  }

  .host-item-name {
    font-weight: 600;
    font-size: 14px;
    color: #334155;
    margin-bottom: 5px;
  }

  .host-item-ip {
    font-size: 12px;
    color: #94A3B8;
    margin-bottom: 5px;
    font-weight: 500;
  }
}

.chart-panel {
  background: #FFFFFF;
  padding: 20px;
  border-radius: 12px;
  height: calc(100vh - 144px);
  overflow-y: auto;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 1px 2px rgba(0, 0, 0, 0.03);
  border: 1px solid #E2E8F0;

  .query-bar {
    margin-bottom: 16px;
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
  }

  .host-info-bar {
    margin-bottom: 16px;
  }
}

.charts-wrapper {
  .chart-box {
    width: 100%;
    height: 260px;
    border: 1px solid #E2E8F0;
    border-radius: 10px;
    padding: 8px;
    background: #FAFBFC;
  }

  .empty-data-tip {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100%;
    min-height: 400px;
  }
}
</style>
