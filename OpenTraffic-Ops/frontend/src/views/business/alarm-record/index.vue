<template>
  <div class="app-container">
    <div class="header-container">
      <el-form :model="queryParams" ref="queryRef" :inline="true" v-show="showSearch" label-width="68px">
        <el-form-item label="主机IP" prop="hostIp">
          <el-input v-model="queryParams.hostIp" placeholder="请输入主机IP" clearable style="width: 240px" @keyup.enter="handleQuery" />
        </el-form-item>
        <el-form-item label="告警类型" prop="alarmType">
          <el-select v-model="queryParams.alarmType" placeholder="告警类型" clearable style="width: 240px">
            <el-option label="指标告警" value="metric" />
            <el-option label="服务告警" value="service" />
          </el-select>
        </el-form-item>
        <el-form-item label="指标类型" prop="metricType">
          <el-select v-model="queryParams.metricType" placeholder="指标类型" clearable style="width: 240px">
            <el-option label="CPU" value="cpu" />
            <el-option label="内存" value="mem" />
            <el-option label="磁盘" value="disk" />
            <el-option label="网络" value="network" />
            <el-option label="负载" value="load" />
            <el-option label="主机离线" value="host_offline" />
            <el-option label="Agent离线" value="agent_offline" />
          </el-select>
        </el-form-item>
        <el-form-item label="告警状态" prop="status">
          <el-select v-model="queryParams.status" placeholder="告警状态" clearable style="width: 240px">
            <el-option label="已触发" value="triggered" />
            <el-option label="已恢复" value="resolved" />
            <el-option label="已确认" value="acknowledged" />
          </el-select>
        </el-form-item>
        <el-form-item label="告警级别" prop="severity">
          <el-select v-model="queryParams.severity" placeholder="告警级别" clearable style="width: 240px">
            <el-option label="警告" value="warning" />
            <el-option label="严重" value="critical" />
          </el-select>
        </el-form-item>
        <el-form-item label="触发时间" style="width: 308px;">
          <el-date-picker v-model="dateRange" value-format="YYYY-MM-DD" type="daterange" range-separator="-" start-placeholder="开始日期" end-placeholder="结束日期"></el-date-picker>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="handleQuery">搜索</el-button>
          <el-button icon="Refresh" @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>

      <el-row :gutter="10" class="mb8">
        <el-col :span="1.5">
          <el-button type="primary" plain icon="Check" @click="handleBatchAck" :disabled="multiple">批量确认</el-button>
        </el-col>
        <el-col :span="1.5">
          <el-badge :value="unreadCount" :hidden="unreadCount === 0" class="item">
            <el-button type="warning" plain icon="Bell" @click="loadUnread">刷新未读</el-button>
          </el-badge>
        </el-col>
        <right-toolbar v-model:showSearch="showSearch" @queryTable="getList" :columns="columns"></right-toolbar>
      </el-row>
    </div>

    <div class="table-container">
      <el-table v-loading="loading" :data="recordList" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="60" align="center" />
        <el-table-column label="编号" align="center" prop="id" v-if="columns[0].visible" width="100" />
        <el-table-column label="规则名称" align="center" prop="ruleName" v-if="columns[1].visible" min-width="220" :show-overflow-tooltip="true" />
        <el-table-column label="主机IP" align="center" prop="hostIp" v-if="columns[2].visible" width="160" />
        <el-table-column label="主机名称" align="center" prop="hostName" v-if="columns[3].visible" width="160" />
        <el-table-column label="告警类型" align="center" prop="alarmType" v-if="columns[4].visible" width="160">
          <template #default="scope">
            <el-tag v-if="scope.row.alarmType === 'metric'" type="primary" size="small">指标</el-tag>
            <el-tag v-else type="warning" size="small">服务</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="指标/服务" align="center" prop="metricType" v-if="columns[5].visible" width="160">
          <template #default="scope">
            {{ metricTypeLabel(scope.row.metricType) }}
          </template>
        </el-table-column>
        <el-table-column label="当前值/阈值" align="center" v-if="columns[6].visible" width="160">
          <template #default="scope">
            <span v-if="scope.row.alarmType === 'metric'">
              <span :class="valueClass(scope.row)">{{ formatValue(scope.row.currentValue) }}</span>
              <span class="text-muted"> / {{ formatValue(scope.row.threshold) }}</span>
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="告警级别" align="center" prop="severity" v-if="columns[7].visible" width="160">
          <template #default="scope">
            <el-tag v-if="scope.row.severity === 'critical'" type="danger" size="small">严重</el-tag>
            <el-tag v-else type="warning" size="small">警告</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" align="center" prop="status" v-if="columns[8].visible" width="160">
          <template #default="scope">
            <el-tag v-if="scope.row.status === 'triggered'" type="danger" size="small" effect="dark">已触发</el-tag>
            <el-tag v-else-if="scope.row.status === 'resolved'" type="success" size="small">已恢复</el-tag>
            <el-tag v-else type="info" size="small">已确认</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="触发时间" align="center" prop="triggerTime" v-if="columns[9].visible" width="220" />
        <!-- <el-table-column label="恢复时间" align="center" prop="resolveTime" v-if="columns[10].visible" width="220" /> -->
        <el-table-column label="操作" align="center" width="220" class-name="small-padding fixed-width">
          <template #default="scope">
            <el-tooltip content="查看详情" placement="top">
              <el-button link type="primary" icon="View" @click="handleView(scope.row)"></el-button>
            </el-tooltip>
            <el-tooltip v-if="scope.row.status === 'triggered'" content="确认告警" placement="top">
              <el-button link type="success" icon="Check" @click="handleAck(scope.row)"></el-button>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>
      <pagination v-show="total > 0" :total="total" v-model:page="queryParams.pageNum" v-model:limit="queryParams.pageSize" @pagination="getList" />
    </div>

    <!-- 详情弹窗 -->
    <el-dialog title="告警详情" v-model="detailOpen" width="750px" append-to-body>
      <el-descriptions :column="2" border v-if="detailRow">
        <el-descriptions-item label="规则名称" :span="2">{{ detailRow.ruleName }}</el-descriptions-item>
        <el-descriptions-item label="主机IP">{{ detailRow.hostIp }}</el-descriptions-item>
        <el-descriptions-item label="主机名称">{{ detailRow.hostName || '-' }}</el-descriptions-item>
        <el-descriptions-item label="告警类型">
          <el-tag v-if="detailRow.alarmType === 'metric'" type="primary">指标</el-tag>
          <el-tag v-else type="warning">服务</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="指标/服务">{{ metricTypeLabel(detailRow.metricType) }}</el-descriptions-item>
        <el-descriptions-item label="当前值">{{ formatValue(detailRow.currentValue) }}</el-descriptions-item>
        <el-descriptions-item label="阈值">{{ formatValue(detailRow.threshold) }}</el-descriptions-item>
        <el-descriptions-item label="告警级别">
          <el-tag v-if="detailRow.severity === 'critical'" type="danger">严重</el-tag>
          <el-tag v-else type="warning">警告</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag v-if="detailRow.status === 'triggered'" type="danger" effect="dark">已触发</el-tag>
          <el-tag v-else-if="detailRow.status === 'resolved'" type="success">已恢复</el-tag>
          <el-tag v-else type="info">已确认</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="触发时间">{{ detailRow.triggerTime }}</el-descriptions-item>
        <el-descriptions-item label="恢复时间">{{ detailRow.resolveTime || '-' }}</el-descriptions-item>
        <el-descriptions-item label="通知状态">
          <el-tag v-if="detailRow.notifyStatus === 'success'" type="success">成功</el-tag>
          <el-tag v-else-if="detailRow.notifyStatus === 'failed'" type="danger">失败</el-tag>
          <el-tag v-else type="info">待发送</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="告警内容" :span="2">
          <pre style="white-space: pre-wrap; margin: 0; font-family: inherit;">{{ detailRow.content }}</pre>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="detailOpen = false">关 闭</el-button>
          <el-button v-if="detailRow && detailRow.status === 'triggered'" type="success" @click="handleAck(detailRow); detailOpen = false">确认告警</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup name="AlarmRecord">
import { listAlarmRecord, ackAlarmRecord, batchAckAlarmRecord, getUnreadCount } from "@/api/business/alarmRecord";
import { metricTypeLabel } from "@/utils/alarm";

const { proxy } = getCurrentInstance();

const recordList = ref([]);
const loading = ref(true);
const showSearch = ref(true);
const ids = ref([]);
const single = ref(true);
const multiple = ref(true);
const total = ref(0);
const unreadCount = ref(0);
const detailOpen = ref(false);
const detailRow = ref(null);
const dateRange = ref([]);

const columns = ref([
  { key: 0, label: '编号', visible: true },
  { key: 1, label: '规则名称', visible: true },
  { key: 2, label: '主机IP', visible: true },
  { key: 3, label: '主机名称', visible: true },
  { key: 4, label: '告警类型', visible: true },
  { key: 5, label: '指标/服务', visible: true },
  { key: 6, label: '当前值/阈值', visible: true },
  { key: 7, label: '告警级别', visible: true },
  { key: 8, label: '状态', visible: true },
  { key: 9, label: '触发时间', visible: true },
  { key: 10, label: '恢复时间', visible: false }
]);

const data = reactive({
  queryParams: {
    pageNum: 1,
    pageSize: 10,
    hostIp: undefined,
    alarmType: undefined,
    metricType: undefined,
    status: undefined,
    severity: undefined
  }
});

const { queryParams } = toRefs(data);

function formatValue(val) {
  if (val === undefined || val === null) return '-';
  return Number(val).toFixed(2);
}

function valueClass(row) {
  if (row.currentValue >= row.threshold) return 'text-danger';
  return '';
}

function getList() {
  loading.value = true;
  const params = proxy.addDateRange(queryParams.value, dateRange.value);
  listAlarmRecord(params).then(res => {
    loading.value = false;
    recordList.value = res.rows || res.data?.rows || [];
    total.value = res.total || res.data?.total || 0;
  });
}

function loadUnread() {
  getUnreadCount().then(res => {
    unreadCount.value = res.data || 0;
  });
}

function handleQuery() {
  queryParams.value.pageNum = 1;
  getList();
}

function resetQuery() {
  dateRange.value = [];
  proxy.resetForm("queryRef");
  handleQuery();
}

function handleSelectionChange(selection) {
  ids.value = selection.map(item => item.id);
  single.value = selection.length != 1;
  multiple.value = !selection.length;
}

function handleView(row) {
  detailRow.value = row;
  detailOpen.value = true;
}

function handleAck(row) {
  proxy.$modal.confirm('确认已处理该告警？').then(() => {
    return ackAlarmRecord(row.id);
  }).then(() => {
    proxy.$modal.msgSuccess("确认成功");
    getList();
    loadUnread();
  }).catch(() => {});
}

function handleBatchAck() {
  if (ids.value.length === 0) {
    proxy.$modal.msgWarning("请选择要确认的告警");
    return;
  }
  proxy.$modal.confirm('确认批量处理选中的 ' + ids.value.length + ' 条告警？').then(() => {
    return batchAckAlarmRecord({ ids: ids.value });
  }).then(() => {
    proxy.$modal.msgSuccess("批量确认成功");
    getList();
    loadUnread();
  }).catch(() => {});
}

getList();
loadUnread();
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

.table-container {
  background: #FFFFFF;
  padding: 20px 24px;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 1px 2px rgba(0, 0, 0, 0.03);
  border: 1px solid #E2E8F0;
}

.text-danger {
  color: #F56C6C;
  font-weight: bold;
}

.text-muted {
  color: #909399;
}
</style>
