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
        <el-form-item>
          <el-button type="primary" icon="Search" @click="handleQuery">搜索</el-button>
          <el-button icon="Refresh" @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>

      <el-row :gutter="10" class="mb8" justify="end">
        <right-toolbar v-model:showSearch="showSearch" @queryTable="getList"></right-toolbar>
      </el-row>
    </div>

    <!-- 主机运维卡片网格 -->
    <div v-loading="loading" class="cards-container">
      <el-empty v-if="filteredHostList.length === 0" description="暂无在线主机" />
      <el-row v-else :gutter="15">
        <el-col v-for="host in filteredHostList" :key="host.id" :xs="24" :sm="12" :lg="8" :xl="6"
          class="card-col">
          <el-card shadow="hover" class="ops-card">
            <!-- 卡片头部 -->
            <div class="card-header">
              <div class="header-left">
                <div class="host-ip">{{ host.ip }}</div>
                <div class="host-name">{{ host.name || '未命名主机' }}</div>
              </div>
              <div class="header-right">
                <el-tag v-if="isHostOnline(host)" type="success" size="small">在线</el-tag>
                <el-tag v-else type="danger" size="small">离线</el-tag>
              </div>
            </div>

            <!-- 简要信息 -->
            <div class="info-section">
              <div class="info-row">
                <span class="info-label">OS</span>
                <span class="info-value">{{ osLabel(host.osType) }} {{ host.osVersion || '' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">CPU</span>
                <span class="info-value">{{ host.cpuCores ? host.cpuCores + '核 ' : '' }}{{ host.cpuModel || '-' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">内存</span>
                <span class="info-value">{{ formatMem(host.memTotalMb) }}</span>
              </div>
            </div>

            <el-divider style="margin: 12px 0;" />

            <!-- 操作按钮 -->
            <div class="card-actions">
              <el-button
                v-if="isHostOnline(host)"
                type="success"
                size="small"
                icon="Terminal"
                @click="goToTerminal(host.ip)"
              >WebSSH 终端</el-button>
              <el-button
                v-if="isHostOnline(host)"
                type="warning"
                size="small"
                icon="Folder"
                @click="goToFile(host.ip)"
              >文件管理</el-button>
              <el-button
                v-if="!isHostOnline(host)"
                disabled
                size="small"
                icon="Warning"
              >主机离线</el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <pagination v-show="total > 0" :total="total" v-model:page="queryParams.pageNum"
      v-model:limit="queryParams.pageSize" @pagination="getList" />
  </div>
</template>

<script setup name="RemoteOps">
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
  name: undefined
});

/** 过滤后的列表（仅显示在线主机） */
const filteredHostList = computed(() => {
  let list = hostInfoList.value.filter(h => isHostOnline(h));
  if (queryParams.value.ip) {
    list = list.filter(h => h.ip && h.ip.includes(queryParams.value.ip));
  }
  if (queryParams.value.name) {
    list = list.filter(h => h.name && h.name.includes(queryParams.value.name));
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

.ops-card {
  border-radius: 12px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid #E2E8F0;

  &:hover {
    box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.08), 0 8px 10px -6px rgba(0, 0, 0, 0.03);
    transform: translateY(-2px);
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

.info-section {
  .info-row {
    display: flex;
    align-items: baseline;
    margin-bottom: 6px;
    font-size: 13px;
    line-height: 1.5;

    .info-label {
      color: #94A3B8;
      width: 44px;
      flex-shrink: 0;
      font-weight: 500;
    }

    .info-value {
      color: #475569;
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      font-weight: 500;
    }
  }
}

.card-actions {
  display: flex;
  justify-content: center;
  gap: 10px;

  .el-button {
    flex: 1;
    border-radius: 8px;
    font-weight: 600;
  }
}
</style>
