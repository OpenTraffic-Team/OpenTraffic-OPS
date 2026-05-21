<template>
  <div class="app-container home">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="panel-group">
      <el-col :xs="12" :sm="12" :md="8" :lg="4" class="card-panel-col">
        <div class="stat-card">
          <div class="stat-glow glow-blue"></div>
          <div class="stat-content">
            <div class="stat-icon icon-bg-blue">
              <svg-icon icon-class="server" class-name="stat-svg" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.hostCount }}</div>
              <div class="stat-label">主机总数</div>
            </div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="12" :md="8" :lg="4" class="card-panel-col">
        <div class="stat-card">
          <div class="stat-glow glow-green"></div>
          <div class="stat-content">
            <div class="stat-icon icon-bg-green">
              <svg-icon icon-class="online" class-name="stat-svg" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ onlineCount }}</div>
              <div class="stat-label">在线主机</div>
            </div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="12" :md="8" :lg="4" class="card-panel-col">
        <div class="stat-card">
          <div class="stat-glow glow-red"></div>
          <div class="stat-content">
            <div class="stat-icon icon-bg-red">
              <svg-icon icon-class="offline" class-name="stat-svg" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.offlineHostCount }}</div>
              <div class="stat-label">离线主机</div>
            </div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="12" :md="8" :lg="4" class="card-panel-col">
        <div class="stat-card">
          <div class="stat-glow glow-purple"></div>
          <div class="stat-content">
            <div class="stat-icon icon-bg-purple">
              <svg-icon icon-class="email" class-name="stat-svg" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.alarmChannelCount }}</div>
              <div class="stat-label">告警通道</div>
            </div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="12" :md="8" :lg="4" class="card-panel-col">
        <div class="stat-card">
          <div class="stat-glow glow-amber"></div>
          <div class="stat-content">
            <div class="stat-icon icon-bg-amber">
              <svg-icon icon-class="slider" class-name="stat-svg" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.alarmRuleCount }}</div>
              <div class="stat-label">告警规则</div>
            </div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="12" :md="8" :lg="4" class="card-panel-col">
        <div class="stat-card">
          <div class="stat-glow glow-red"></div>
          <div class="stat-content">
            <div class="stat-icon icon-bg-red">
              <svg-icon icon-class="bell" class-name="stat-svg" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.unhandledAlarmCount }}</div>
              <div class="stat-label">未处理告警</div>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 快捷入口 -->
    <el-row :gutter="20" class="quick-entry">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <div class="header-title">
                <div class="title-dot dot-blue"></div>
                <span>快捷入口</span>
              </div>
            </div>
          </template>
          <el-row :gutter="20">
            <el-col :xs="12" :sm="8" :md="6" :lg="4">
              <div class="quick-item" @click="goTo('/business/host-info')">
                <div class="quick-icon-wrap icon-bg-blue">
                  <svg-icon icon-class="server" class-name="quick-icon" />
                </div>
                <span>主机信息</span>
              </div>
            </el-col>
            <el-col :xs="12" :sm="8" :md="6" :lg="4">
              <div class="quick-item" @click="goTo('/business/host-health')">
                <div class="quick-icon-wrap icon-bg-cyan">
                  <svg-icon icon-class="chart" class-name="quick-icon" />
                </div>
                <span>健康监控</span>
              </div>
            </el-col>
            <el-col :xs="12" :sm="8" :md="6" :lg="4">
              <div class="quick-item" @click="goTo('/alarm/config')">
                <div class="quick-icon-wrap icon-bg-purple">
                  <svg-icon icon-class="email" class-name="quick-icon" />
                </div>
                <span>告警通道</span>
              </div>
            </el-col>
            <el-col :xs="12" :sm="8" :md="6" :lg="4">
              <div class="quick-item" @click="goTo('/alarm/threshold')">
                <div class="quick-icon-wrap icon-bg-amber">
                  <svg-icon icon-class="slider" class-name="quick-icon" />
                </div>
                <span>告警规则</span>
              </div>
            </el-col>
            <el-col :xs="12" :sm="8" :md="6" :lg="4">
              <div class="quick-item" @click="goTo('/alarm/record')">
                <div class="quick-icon-wrap icon-bg-red">
                  <svg-icon icon-class="bell" class-name="quick-icon" />
                </div>
                <span>告警记录</span>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>
    </el-row>

    <!-- 主机状态列表 -->
    <el-row :gutter="20" class="host-list">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <div class="header-title">
                <div class="title-dot dot-green"></div>
                <span>主机状态概览</span>
              </div>
              <el-button size="small" class="rtm-btn-text" @click="goTo('/business/host-info')">
                查看更多
                <svg-icon icon-class="arrow-right" style="font-size: 12px;" />
              </el-button>
            </div>
          </template>
          <el-table :data="hostList" v-loading="loading" class="dark-table" size="small">
            <el-table-column label="主机IP" prop="ip" align="center" width="160" />
            <el-table-column label="主机名称" prop="name" align="center" width="160" />
            <el-table-column label="操作系统" prop="osType" align="center" width="160">
              <template #default="scope">
                <span v-if="scope.row.osType === 'linux'" class="rtm-badge rtm-badge-info">Linux</span>
                <span v-else-if="scope.row.osType === 'windows'" class="rtm-badge rtm-badge-warning">Windows</span>
                <span v-else-if="scope.row.osType === 'darwin'" class="rtm-badge rtm-badge-default">MacOS</span>
                <span v-else class="rtm-badge rtm-badge-default">{{ scope.row.osType || '-' }}</span>
              </template>
            </el-table-column>
            <el-table-column label="CPU使用率" align="center" min-width="120">
              <template #default="scope">
                <el-progress v-if="scope.row.cpuUsage && scope.row.cpuUsage !== '0'"
                  :percentage="Math.min(parseFloat(scope.row.cpuUsage) || 0, 100)"
                  :color="cpuProgressColor"
                  :stroke-width="8"
                  :show-text="true" />
                <span v-else-if="scope.row.isOnline === true">0%</span>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="内存使用率" align="center" min-width="120">
              <template #default="scope">
                <el-progress v-if="scope.row.memUsage && scope.row.memUsage !== '0'"
                  :percentage="Math.min(parseFloat(scope.row.memUsage) || 0, 100)"
                  :color="memProgressColor"
                  :stroke-width="8"
                  :show-text="true" />
                <span v-else-if="scope.row.isOnline === true">0%</span>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="磁盘使用率" align="center" min-width="120">
              <template #default="scope">
                <el-progress v-if="scope.row.diskUsage && scope.row.diskUsage !== '0'"
                  :percentage="Math.min(parseFloat(scope.row.diskUsage) || 0, 100)"
                  :color="diskProgressColor"
                  :stroke-width="8"
                  :show-text="true" />
                <span v-else-if="scope.row.isOnline === true">0%</span>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="网络入" align="center" width="100">
              <template #default="scope">
                <span v-if="scope.row.netIn && scope.row.netIn !== '0'">{{ parseFloat(scope.row.netIn).toFixed(1) }}</span>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="网络出" align="center" width="100">
              <template #default="scope">
                <span v-if="scope.row.netOut && scope.row.netOut !== '0'">{{ parseFloat(scope.row.netOut).toFixed(1) }}</span>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="状态" align="center" width="120">
              <template #default="scope">
                <span v-if="scope.row.isOnline === true" class="rtm-badge rtm-badge-success">在线</span>
                <span v-else-if="scope.row.isOnline === false" class="rtm-badge rtm-badge-danger">离线</span>
                <span v-else class="rtm-badge rtm-badge-default">未知</span>
              </template>
            </el-table-column>
            <el-table-column label="最后心跳" prop="lastHeartbeat" align="center" width="210" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最新告警记录 -->
    <el-row :gutter="20" class="alarm-list">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <div class="header-title">
                <div class="title-dot dot-red"></div>
                <span>最新告警记录</span>
              </div>
              <el-button size="small" class="rtm-btn-text" @click="goTo('/alarm/record')">
                查看更多
                <svg-icon icon-class="arrow-right" style="font-size: 12px;" />
              </el-button>
            </div>
          </template>
          <el-table :data="recentAlarmList" class="dark-table" size="small">
            <el-table-column label="规则名称" prop="ruleName" align="center" min-width="160" :show-overflow-tooltip="true" />
            <el-table-column label="主机IP" prop="hostIp" align="center" width="160" />
            <el-table-column label="指标/服务" align="center" width="160">
              <template #default="scope">
                {{ metricTypeLabel(scope.row.metricType) }}
              </template>
            </el-table-column>
            <el-table-column label="告警级别" align="center" width="160">
              <template #default="scope">
                <span v-if="scope.row.severity === 'critical'" class="rtm-badge rtm-badge-danger">严重</span>
                <span v-else class="rtm-badge rtm-badge-warning">警告</span>
              </template>
            </el-table-column>
            <el-table-column label="状态" align="center" width="160">
              <template #default="scope">
                <span v-if="scope.row.status === 'triggered'" class="rtm-badge rtm-badge-danger">已触发</span>
                <span v-else-if="scope.row.status === 'resolved'" class="rtm-badge rtm-badge-success">已恢复</span>
                <span v-else class="rtm-badge rtm-badge-default">已确认</span>
              </template>
            </el-table-column>
            <el-table-column label="触发时间" prop="triggerTime" align="center" width="220" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup name="Index">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { getStatisticData } from "@/api/business/home";
import { listHostMon } from "@/api/business/hostHealth";
import { metricTypeLabel } from "@/utils/alarm";

const router = useRouter();
const loading = ref(false);

const statistics = ref({
  hostCount: 0,
  offlineHostCount: 0,
  alarmChannelCount: 0,
  alarmRuleCount: 0,
  unhandledAlarmCount: 0
});

const hostList = ref([]);
const recentAlarmList = ref([]);

// 在线主机数 = 总数 - 离线数
const onlineCount = computed(() => {
  const total = statistics.value.hostCount || 0;
  const offline = statistics.value.offlineHostCount || 0;
  return Math.max(0, total - offline);
});

// CPU进度条颜色
const cpuProgressColor = [
  { color: '#10B981', percentage: 60 },
  { color: '#F59E0B', percentage: 80 },
  { color: '#EF4444', percentage: 100 }
];

// 内存进度条颜色
const memProgressColor = [
  { color: '#10B981', percentage: 60 },
  { color: '#F59E0B', percentage: 80 },
  { color: '#EF4444', percentage: 100 }
];

// 磁盘进度条颜色
const diskProgressColor = [
  { color: '#10B981', percentage: 70 },
  { color: '#F59E0B', percentage: 85 },
  { color: '#EF4444', percentage: 100 }
];

/** 获取统计数据 */
function getStatistic() {
  getStatisticData().then(response => {
    const data = response.data || {};
    statistics.value = {
      hostCount: data.hostCount || 0,
      offlineHostCount: data.offlineHostCount || 0,
      alarmChannelCount: data.alarmChannelCount || 0,
      alarmRuleCount: data.alarmRuleCount || 0,
      unhandledAlarmCount: data.unhandledAlarmCount || 0
    };
    recentAlarmList.value = data.recentAlarms || [];
  });
}

/** 获取主机监控列表 */
function getHostList() {
  loading.value = true;
  listHostMon().then(response => {
    const data = response.data || response;
    hostList.value = (data.rows || data || []).slice(0, 10);
    loading.value = false;
  }).catch(() => {
    loading.value = false;
  });
}

/** 跳转页面 */
function goTo(path) {
  router.push(path);
}

// 轮询定时器
let pollTimer = null;

/** 启动轮询 */
function startPolling() {
  stopPolling();
  pollTimer = setInterval(() => {
    getStatistic();
    getHostList();
  }, 5000);
}

/** 停止轮询 */
function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

onMounted(() => {
  getStatistic();
  getHostList();
  startPolling();
});

onBeforeUnmount(() => {
  stopPolling();
});
</script>

<style scoped lang="scss">
.home {
  min-height: calc(100vh - 84px);
  padding: 20px;
}

// ========== 统计卡片 (rtm-initialization style) ==========
.panel-group {
  margin-bottom: 20px;

  .card-panel-col {
    margin-bottom: 20px;
  }
}

.stat-card {
  position: relative;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  padding: 24px;
  overflow: hidden;
  transition: all 0.3s ease;
  cursor: pointer;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);

  &:hover {
    transform: translateY(-3px);
    border-color: #d1d5db;
    box-shadow: 0 10px 40px -10px rgba(0, 0, 0, 0.1);
  }
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

  .stat-card:hover & {
    opacity: 0.3;
  }
}

.glow-blue { background: #3b82f6; }
.glow-green { background: #10b981; }
.glow-red { background: #ef4444; }
.glow-purple { background: #8b5cf6; }
.glow-amber { background: #f59e0b; }

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

  .stat-svg {
    font-size: 24px;
  }
}

.icon-bg-blue {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  box-shadow: 0 4px 15px -4px rgba(59, 130, 246, 0.4);
}
.icon-bg-green {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  box-shadow: 0 4px 15px -4px rgba(16, 185, 129, 0.4);
}
.icon-bg-red {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  box-shadow: 0 4px 15px -4px rgba(239, 68, 68, 0.4);
}
.icon-bg-purple {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
  box-shadow: 0 4px 15px -4px rgba(139, 92, 246, 0.4);
}
.icon-bg-amber {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  box-shadow: 0 4px 15px -4px rgba(245, 158, 11, 0.4);
}
.icon-bg-cyan {
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
  box-shadow: 0 4px 15px -4px rgba(6, 182, 212, 0.4);
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

// ========== 快捷入口 ==========
.quick-entry {
  margin-bottom: 20px;

  :deep(.el-card) {
    border-radius: 16px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
    border: 1px solid #E2E8F0;
    transition: box-shadow 0.3s ease;

    &:hover {
      box-shadow: 0 10px 40px -10px rgba(0, 0, 0, 0.08);
    }
  }

  :deep(.el-card__header) {
    padding: 16px 24px !important;
    border-bottom: 1px solid #F1F5F9;
  }

  .quick-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 24px 0;
    cursor: pointer;
    border-radius: 12px;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    gap: 10px;

    &:hover {
      background-color: #F8FAFC;
      transform: translateY(-3px);
    }

    .quick-icon-wrap {
      width: 48px;
      height: 48px;
      border-radius: 14px;
      display: flex;
      align-items: center;
      justify-content: center;

      .quick-icon {
        font-size: 22px;
        color: #fff;
      }
    }

    span {
      font-size: 14px;
      color: #475569;
      font-weight: 500;
    }
  }
}

// ========== 卡片通用头部 ==========
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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

.dot-blue {
  background: #3b82f6;
  box-shadow: 0 0 8px rgba(59, 130, 246, 0.4);
}
.dot-green {
  background: #10b981;
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.4);
}
.dot-red {
  background: #ef4444;
  box-shadow: 0 0 8px rgba(239, 68, 68, 0.4);
}

// ========== 列表区域 ==========
.host-list,
.alarm-list {
  margin-bottom: 20px;

  :deep(.el-card) {
    border-radius: 16px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
    border: 1px solid #E2E8F0;
    transition: box-shadow 0.3s ease;

    &:hover {
      box-shadow: 0 10px 40px -10px rgba(0, 0, 0, 0.08);
    }
  }

  :deep(.el-card__header) {
    padding: 16px 24px !important;
    border-bottom: 1px solid #F1F5F9;
  }
}

// ========== 表格覆盖 ==========
:deep(.dark-table) {
  background: transparent;
}

:deep(.dark-table .el-table__inner-wrapper::before) {
  display: none;
}

:deep(.dark-table .el-table__header th) {
  background: #f9fafb !important;
  color: #6b7280 !important;
  font-weight: 500 !important;
  font-size: 13px !important;
  border-bottom: 1px solid #f3f4f6 !important;
  padding: 12px 16px !important;
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
  font-size: 13px !important;
  border-bottom: 1px solid #f3f4f6 !important;
  padding: 14px 16px !important;
}

:deep(.dark-table .el-table__empty-text) {
  color: #9ca3af;
}
</style>
