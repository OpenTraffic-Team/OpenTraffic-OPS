<template>
  <div class="app-container">
    <div class="header-container">
      <el-form :model="queryParams" ref="queryRef" :inline="true" v-show="showSearch" label-width="68px">
        <el-form-item label="主机IP" prop="ip">
          <el-input v-model="queryParams.ip" placeholder="请输入主机IP" clearable style="width: 240px"
            @keyup.enter="handleQuery" />
        </el-form-item>
        <el-form-item label="主机名称" prop="name">
          <el-input v-model="queryParams.name" placeholder="请输入主机名称" clearable style="width: 240px"
            @keyup.enter="handleQuery" />
        </el-form-item>
        <el-form-item label="在线状态" prop="isOnline">
          <el-select v-model="queryParams.isOnline" placeholder="在线状态" clearable style="width: 240px">
            <el-option label="在线" :value="true" />
            <el-option label="离线" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item label="操作系统" prop="osType">
          <el-select v-model="queryParams.osType" placeholder="操作系统" clearable style="width: 240px">
            <el-option label="Linux" value="linux" />
            <el-option label="Windows" value="windows" />
            <el-option label="MacOS" value="darwin" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="handleQuery">搜索</el-button>
          <el-button icon="Refresh" @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>

      <el-row :gutter="10" class="mb8" justify="end">
        <right-toolbar v-model:showSearch="showSearch" @queryTable="getList"></right-toolbar>
      </el-row>
    </div>

    <!-- 主机卡片网格 -->
    <div v-loading="loading" class="cards-container">
      <el-empty v-if="filteredHostList.length === 0" description="暂无主机数据" />
      <el-row v-else :gutter="15">
        <el-col v-for="host in filteredHostList" :key="host.id" :xs="24" :sm="12" :lg="8" :xl="6"
          class="card-col">
          <el-card shadow="hover" :class="['host-card', !isHostOnline(host) ? 'offline' : '']">
            <!-- 卡片头部 -->
            <div class="card-header">
              <div class="header-left">
                <div class="host-ip">{{ host.ip }}</div>
                <div class="host-name">{{ host.name || '未命名主机' }}</div>
              </div>
              <div class="header-right">
                <el-tag v-if="isHostOnline(host)" type="success" size="small">在线</el-tag>
                <el-tag v-else-if="isHostOffline(host)" type="danger" size="small">离线</el-tag>
                <el-tag v-else type="info" size="small">未知</el-tag>
              </div>
            </div>

            <!-- 硬件信息 -->
            <div class="hardware-section">
              <div class="hw-row">
                <span class="hw-label">OS</span>
                <span class="hw-value">{{ osLabel(host.osType) }} {{ host.osVersion || '' }}</span>
              </div>
              <div class="hw-row">
                <span class="hw-label">CPU</span>
                <span class="hw-value">{{ host.cpuCores ? host.cpuCores + '核 ' : '' }}{{ host.cpuModel || '-' }}</span>
              </div>
              <div class="hw-row">
                <span class="hw-label">内存</span>
                <span class="hw-value">{{ formatMem(host.memTotalMb) }}</span>
              </div>
              <div class="hw-row">
                <span class="hw-label">磁盘</span>
                <span class="hw-value">{{ host.diskTotalGb ? host.diskTotalGb + ' GB' : '-' }}</span>
              </div>
              <div class="hw-row">
                <span class="hw-label">MAC</span>
                <span class="hw-value">{{ host.macAddress || '-' }}</span>
              </div>
              <div class="hw-row">
                <span class="hw-label">Agent</span>
                <span class="hw-value">{{ host.agentVersion || '-' }}</span>
              </div>
              <div class="hw-row">
                <span class="hw-label">心跳</span>
                <span class="hw-value">{{ host._health?.lastHeartbeat || host.lastHeartbeat || '-' }}</span>
              </div>
            </div>

            <!-- 分隔线 -->
            <el-divider style="margin: 12px 0;" />

            <!-- 实时状态 -->
            <div class="realtime-section">
              <div class="rt-row">
                <span class="rt-label">CPU</span>
                <el-progress class="rt-progress"
                  :percentage="Math.min(parseFloat(host._health?.cpuUsage) || 0, 100)"
                  :color="progressColors" :stroke-width="10" :show-text="false" />
                <span class="rt-value">{{ host._health?.cpuUsage || '0' }}%</span>
              </div>
              <div class="rt-row">
                <span class="rt-label">内存</span>
                <el-progress class="rt-progress"
                  :percentage="Math.min(parseFloat(host._health?.memUsage) || 0, 100)"
                  :color="progressColors" :stroke-width="10" :show-text="false" />
                <span class="rt-value">{{ host._health?.memUsage || '0' }}%</span>
              </div>
              <div class="rt-row">
                <span class="rt-label">磁盘</span>
                <el-progress class="rt-progress"
                  :percentage="Math.min(parseFloat(host._health?.diskUsage) || 0, 100)"
                  :color="progressColors" :stroke-width="10" :show-text="false" />
                <span class="rt-value">{{ host._health?.diskUsage || '0' }}%</span>
              </div>
              <div class="rt-row-inline">
                <span class="rt-inline-item">
                  <span class="rt-inline-label">入流量</span>
                  <span class="rt-inline-value">{{ host._health?.netIn || '0' }} KB/s</span>
                </span>
                <span class="rt-inline-item">
                  <span class="rt-inline-label">出流量</span>
                  <span class="rt-inline-value">{{ host._health?.netOut || '0' }} KB/s</span>
                </span>
              </div>
              <div class="rt-row-inline">
                <span class="rt-inline-item">
                  <span class="rt-inline-label">负载</span>
                  <span class="rt-inline-value">{{ host._health?.loadAvg || '-' }}</span>
                </span>
                <span class="rt-inline-item">
                  <span class="rt-inline-label">已用内存</span>
                  <span class="rt-inline-value">{{ host._health?.memUsageMB || '0' }} MB</span>
                </span>
              </div>
            </div>
            <el-divider style="margin: 12px 0;" />
            <div class="card-footer">
              <el-button class="history-btn" type="primary" size="small" icon="TrendCharts" @click="goToHistory(host.ip)"
                >历史数据</el-button
              >
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <pagination v-show="total > 0" :total="total" v-model:page="queryParams.pageNum"
      v-model:limit="queryParams.pageSize" @pagination="getList" />
  </div>
</template>

<script setup name="HostInfo">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { listHostInfo } from "@/api/business/hostInfo";
import { listHostMon } from "@/api/business/hostHealth";

const router = useRouter();
const { proxy } = getCurrentInstance();

const hostInfoList = ref([]);
const loading = ref(true);
const showSearch = ref(true);
const total = ref(0);

const queryParams = ref({
  pageNum: 1,
  pageSize: 12,
  ip: undefined,
  name: undefined,
  isOnline: undefined,
  osType: undefined
});

const progressColors = [
  { color: '#10B981', percentage: 60 },
  { color: '#F59E0B', percentage: 80 },
  { color: '#EF4444', percentage: 100 }
];

/** 过滤后的列表 */
const filteredHostList = computed(() => {
  let list = hostInfoList.value;
  if (queryParams.value.ip) {
    list = list.filter(h => h.ip && h.ip.includes(queryParams.value.ip));
  }
  if (queryParams.value.name) {
    list = list.filter(h => h.name && h.name.includes(queryParams.value.name));
  }
  if (queryParams.value.isOnline !== undefined && queryParams.value.isOnline !== null && queryParams.value.isOnline !== '') {
    list = list.filter(h => h.isOnline === queryParams.value.isOnline);
  }
  if (queryParams.value.osType) {
    list = list.filter(h => h.osType === queryParams.value.osType);
  }
  return list;
});

/** 判断主机是否在线（优先使用实时健康度数据） */
function isHostOnline(host) {
  if (host._health?.isOnline !== undefined) {
    return host._health.isOnline === true;
  }
  return host.isOnline === true;
}

/** 判断主机是否离线（优先使用实时健康度数据） */
function isHostOffline(host) {
  if (host._health?.isOnline !== undefined) {
    return host._health.isOnline === false;
  }
  return host.isOnline === false;
}

/** OS 标签 */
function osLabel(osType) {
  if (osType === 'linux') return 'Linux';
  if (osType === 'windows') return 'Windows';
  if (osType === 'darwin') return 'MacOS';
  return osType || '-';
}

/** 格式化内存显示 */
function formatMem(mb) {
  if (!mb || mb <= 0) return '-';
  if (mb >= 1024 * 1024) return (mb / 1024 / 1024).toFixed(2) + ' TB';
  if (mb >= 1024) return (mb / 1024).toFixed(2) + ' GB';
  return mb + ' MB';
}

/** 查询主机信息列表 */
function getList() {
  loading.value = true;
  listHostInfo(queryParams.value).then(response => {
    const pageData = response.data || response;
    hostInfoList.value = pageData.rows || [];
    total.value = pageData.total || 0;
    loading.value = false;
    // 加载完主机列表后，获取实时健康度数据
    loadHealthData();
  }).catch(() => {
    loading.value = false;
  });
}

/** 加载实时健康度数据：直接合并到 hostInfoList 中 */
function loadHealthData() {
  listHostMon().then(response => {
    const data = response.data || response;
    const list = data.rows || data || [];
    const map = {};
    list.forEach(item => {
      if (item.ip) {
        map[item.ip] = item;
      }
    });
    // 直接给每个 host 对象附加 _health 属性，触发响应式更新
    hostInfoList.value = hostInfoList.value.map(host => {
      const health = map[host.ip];
      if (health) {
        return { ...host, _health: health };
      }
      return host;
    });
  });
}

// 轮询定时器
let pollTimer = null;

/** 启动轮询 */
function startPolling() {
  stopPolling();
  pollTimer = setInterval(() => {
    loadHealthData();
  }, 3000);
}

/** 停止轮询 */
function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

onMounted(() => {
  startPolling();
});

onBeforeUnmount(() => {
  stopPolling();
});

/** 跳转历史数据页面 */
function goToHistory(ip) {
  router.push({
    path: '/business/host-health',
    query: { ip: ip }
  });
}

/** 跳转终端页面 */
function goToTerminal(ip) {
  router.push({
    path: '/business/remote-terminal',
    query: { ip: ip }
  });
}

/** 跳转文件管理页面 */
function goToFile(ip) {
  router.push({
    path: '/business/remote-file',
    query: { ip: ip }
  });
}

/** 搜索按钮操作 */
function handleQuery() {
  queryParams.value.pageNum = 1;
  getList();
}

/** 重置按钮操作 */
function resetQuery() {
  proxy.resetForm("queryRef");
  queryParams.value.isOnline = undefined;
  queryParams.value.osType = undefined;
  handleQuery();
}

getList();
</script>

<style scoped lang="scss">
.app-container {
  padding: 24px;
}

.header-container {
  background: #FFFFFF;
  padding: 20px 24px;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 1px 2px rgba(0, 0, 0, 0.03);
  margin-bottom: 20px;
}

.cards-container {
  margin-top: 0;
}

.card-col {
  margin-bottom: 20px;
}

.host-card {
  border-radius: 12px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid #E2E8F0;

  &:hover {
    box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.08), 0 8px 10px -6px rgba(0, 0, 0, 0.03);
    transform: translateY(-2px);
  }

  &.offline {
    opacity: 0.65;
    border-color: #E2E8F0;
    box-shadow: none;

    &:hover {
      transform: none;
      box-shadow: none;
    }

    .card-header {
      .host-ip {
        color: #94A3B8;
      }
    }
  }

  :deep(.el-card__body) {
    padding: 20px;
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;

  .header-left {
    flex: 1;
    min-width: 0;

    .host-ip {
      font-size: 17px;
      font-weight: 700;
      color: #1E293B;
      line-height: 1.4;
      letter-spacing: -0.3px;
    }

    .host-name {
      font-size: 13px;
      color: #64748B;
      margin-top: 6px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      font-weight: 500;
    }
  }

  .header-right {
    margin-left: 12px;
    flex-shrink: 0;
  }
}

.hardware-section {
  .hw-row {
    display: flex;
    align-items: baseline;
    margin-bottom: 6px;
    font-size: 13px;
    line-height: 1.5;

    .hw-label {
      color: #94A3B8;
      width: 44px;
      flex-shrink: 0;
      font-weight: 500;
    }

    .hw-value {
      color: #475569;
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      font-weight: 500;
    }
  }
}

.realtime-section {
  .rt-row {
    display: flex;
    align-items: center;
    margin-bottom: 10px;

    .rt-label {
      width: 36px;
      font-size: 12px;
      color: #64748B;
      flex-shrink: 0;
      font-weight: 600;
    }

    .rt-progress {
      flex: 1;
      margin: 0 10px;
    }

    .rt-value {
      width: 44px;
      font-size: 12px;
      color: #334155;
      text-align: right;
      font-weight: 600;
      flex-shrink: 0;
    }
  }

  .rt-row-inline {
    display: flex;
    justify-content: space-between;
    margin-bottom: 6px;

    .rt-inline-item {
      font-size: 12px;

      .rt-inline-label {
        color: #94A3B8;
        margin-right: 6px;
        font-weight: 500;
      }

      .rt-inline-value {
        color: #334155;
        font-weight: 600;
      }
    }
  }
}

.card-footer {
  text-align: center;

  .history-btn {
    width: 100%;
    border-radius: 8px;
    font-weight: 600;
  }
}
</style>
