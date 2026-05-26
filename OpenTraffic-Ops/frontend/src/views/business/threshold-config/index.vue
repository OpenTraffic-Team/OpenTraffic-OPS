<template>
  <div class="app-container">
    <div class="header-container">
      <el-form :model="queryParams" ref="queryRef" :inline="true" v-show="showSearch" label-width="68px">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="queryParams.name" placeholder="请输入规则名称" clearable style="width: 240px" @keyup.enter="handleQuery" />
        </el-form-item>
        <el-form-item label="规则类型" prop="ruleType">
          <el-select v-model="queryParams.ruleType" placeholder="规则类型" clearable style="width: 240px">
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
        <el-form-item label="状态" prop="status">
          <el-select v-model="queryParams.status" placeholder="状态" clearable style="width: 240px">
            <el-option label="启用" value="0" />
            <el-option label="禁用" value="1" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="handleQuery">搜索</el-button>
          <el-button icon="Refresh" @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>

      <el-row :gutter="10" class="mb8">
        <el-col :span="1.5">
          <el-button type="primary" plain icon="Plus" @click="handleAdd">新增</el-button>
        </el-col>
        <el-col :span="1.5">
          <el-button type="success" plain icon="Edit" :disabled="single" @click="handleUpdate">修改</el-button>
        </el-col>
        <el-col :span="1.5">
          <el-button type="danger" plain icon="Delete" :disabled="multiple" @click="handleDelete">删除</el-button>
        </el-col>
        <right-toolbar v-model:showSearch="showSearch" @queryTable="getList" :columns="columns"></right-toolbar>
      </el-row>
    </div>

    <div class="table-container">
      <el-table v-loading="loading" :data="ruleList" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="60" align="center" />
        <el-table-column label="编号" align="center" prop="id" v-if="columns[0].visible" width="100" />
        <el-table-column label="规则名称" align="center" prop="name" v-if="columns[1].visible" min-width="220" :show-overflow-tooltip="true" />
        <el-table-column label="规则类型" align="center" prop="ruleType" v-if="columns[2].visible" width="160">
          <template #default="scope">
            <el-tag v-if="scope.row.ruleType === 'metric'" type="primary">指标</el-tag>
            <el-tag v-else type="warning">服务</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="指标/服务" align="center" prop="metricType" v-if="columns[3].visible" width="160">
          <template #default="scope">
            {{ metricTypeLabel(scope.row.metricType) }}
          </template>
        </el-table-column>
        <el-table-column label="阈值条件" align="center" v-if="columns[4].visible" width="160">
          <template #default="scope">
            <span v-if="scope.row.ruleType === 'metric'">
              {{ compareOpLabel(scope.row.compareOp) }} {{ scope.row.threshold }}
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="持续时间" align="center" prop="duration" v-if="columns[5].visible" width="160">
          <template #default="scope">
            {{ scope.row.duration }}秒
          </template>
        </el-table-column>
        <el-table-column label="告警级别" align="center" prop="severity" v-if="columns[6].visible" width="160">
          <template #default="scope">
            <el-tag v-if="scope.row.severity === 'critical'" type="danger">严重</el-tag>
            <el-tag v-else type="warning">警告</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" align="center" prop="status" v-if="columns[7].visible" width="160">
          <template #default="scope">
            <el-switch v-model="scope.row.status" active-value="0" inactive-value="1" @change="handleStatusChange(scope.row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" align="center" width="220" class-name="small-padding fixed-width">
          <template #default="scope">
            <el-tooltip content="修改" placement="top">
              <el-button link type="primary" icon="Edit" @click="handleUpdate(scope.row)"></el-button>
            </el-tooltip>
            <el-tooltip content="删除" placement="top">
              <el-button link type="danger" icon="Delete" @click="handleDelete(scope.row)"></el-button>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>
      <pagination v-show="total > 0" :total="total" v-model:page="queryParams.pageNum" v-model:limit="queryParams.pageSize" @pagination="getList" />
    </div>

    <!-- 新增/编辑弹窗 -->
    <el-dialog :title="title" v-model="open" width="650px" append-to-body>
      <el-form :model="form" :rules="rules" ref="ruleRef" label-width="100px">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入规则名称" maxlength="100" />
        </el-form-item>
        <el-form-item label="规则类型" prop="ruleType">
          <el-radio-group v-model="form.ruleType" @change="onRuleTypeChange">
            <el-radio label="metric">指标告警</el-radio>
            <el-radio label="service">服务告警</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="form.ruleType === 'metric' ? '指标类型' : '服务类型'" prop="metricType">
          <el-select v-model="form.metricType" placeholder="请选择类型" style="width: 100%">
            <template v-if="form.ruleType === 'metric'">
              <el-option label="CPU使用率" value="cpu" />
              <el-option label="内存使用率" value="mem" />
              <el-option label="磁盘使用率" value="disk" />
              <el-option label="网络流量" value="network" />
              <el-option label="系统负载" value="load" />
            </template>
            <template v-else>
              <el-option label="主机离线" value="host_offline" />
              <el-option label="Agent离线" value="agent_offline" />
            </template>
          </el-select>
        </el-form-item>
        <el-form-item label="关联主机" prop="hostId">
          <el-select-v2 v-model="form.hostId" :options="hostOptions" placeholder="请选择主机（0表示全部主机）" style="width: 100%" clearable filterable />
        </el-form-item>
        <template v-if="form.ruleType === 'metric'">
          <el-row>
            <el-col :span="12">
              <el-form-item label="比较运算符" prop="compareOp">
                <el-select v-model="form.compareOp" placeholder="运算符" style="width: 100%">
                  <el-option label=">" value="gt" />
                  <el-option label="<" value="lt" />
                  <el-option label=">=" value="ge" />
                  <el-option label="<=" value="le" />
                  <el-option label="=" value="eq" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="阈值" prop="threshold">
                <el-input-number v-model="form.threshold" :precision="2" :step="1" :min="0" style="width: 100%" />
              </el-form-item>
            </el-col>
          </el-row>
        </template>
        <el-form-item label="持续时间" prop="duration">
          <el-input-number v-model="form.duration" :min="0" :step="30" style="width: 100%">
            <template #suffix>秒</template>
          </el-input-number>
          <span class="form-tip">持续超过该时间才触发告警</span>
        </el-form-item>
        <el-form-item label="告警级别" prop="severity">
          <el-radio-group v-model="form.severity">
            <el-radio label="warning">警告</el-radio>
            <el-radio label="critical">严重</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="告警通道">
          <el-select v-model="selectedChannels" multiple placeholder="请选择告警通道" style="width: 100%">
            <el-option v-for="ch in channelOptions" :key="ch.id" :label="ch.name" :value="ch.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio label="0">启用</el-radio>
            <el-radio label="1">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" placeholder="请输入备注"></el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button type="primary" @click="submitForm">确 定</el-button>
          <el-button @click="cancel">取 消</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup name="ThresholdConfig">
import { listAlarmRule, delAlarmRule, addAlarmRule, updateAlarmRule, updateAlarmRuleStatus } from "@/api/business/alarmRule";
import { listAllAlarmChannel } from "@/api/business/alarmChannel";
import { listHostInfo } from "@/api/business/hostInfo";

const { proxy } = getCurrentInstance();

const ruleList = ref([]);
const open = ref(false);
const loading = ref(true);
const showSearch = ref(true);
const ids = ref([]);
const single = ref(true);
const multiple = ref(true);
const total = ref(0);
const title = ref("");
const channelOptions = ref([]);
const hostOptions = ref([{ value: 0, label: '全部主机' }]);
const selectedChannels = ref([]);

const columns = ref([
  { key: 0, label: '编号', visible: true },
  { key: 1, label: '规则名称', visible: true },
  { key: 2, label: '规则类型', visible: true },
  { key: 3, label: '指标/服务', visible: true },
  { key: 4, label: '阈值条件', visible: true },
  { key: 5, label: '持续时间', visible: true },
  { key: 6, label: '告警级别', visible: true },
  { key: 7, label: '状态', visible: true }
]);

const data = reactive({
  form: {
    ruleType: 'metric',
    compareOp: 'gt',
    severity: 'warning',
    status: '0',
    duration: 60,
    hostId: 0
  },
  queryParams: {
    pageNum: 1,
    pageSize: 10,
    name: undefined,
    ruleType: undefined,
    metricType: undefined,
    status: undefined
  },
  rules: {
    name: [{ required: true, message: "规则名称不能为空", trigger: "blur" }],
    ruleType: [{ required: true, message: "规则类型不能为空", trigger: "change" }],
    metricType: [{ required: true, message: "指标/服务类型不能为空", trigger: "change" }]
  }
});

const { queryParams, form, rules } = toRefs(data);

const metricTypeMap = {
  'cpu': 'CPU使用率',
  'mem': '内存使用率',
  'disk': '磁盘使用率',
  'network': '网络流量',
  'load': '系统负载',
  'host_offline': '主机离线',
  'agent_offline': 'Agent离线'
};

const compareOpMap = {
  'gt': '>',
  'lt': '<',
  'ge': '>=',
  'le': '<=',
  'eq': '='
};

function metricTypeLabel(type) {
  return metricTypeMap[type] || type;
}

function compareOpLabel(op) {
  return compareOpMap[op] || op;
}

function getList() {
  loading.value = true;
  listAlarmRule(queryParams.value).then(res => {
    loading.value = false;
    ruleList.value = res.rows || res.data?.rows || [];
    total.value = res.total || res.data?.total || 0;
  });
}

function loadChannels() {
  listAllAlarmChannel().then(res => {
    const data = res.data || res;
    channelOptions.value = Array.isArray(data) ? data : (data.rows || []);
  });
}

function loadHosts() {
  listHostInfo({ pageNum: 1, pageSize: 999 }).then(res => {
    const rows = res.rows || res.data?.rows || [];
    const options = [{ value: 0, label: '全部主机' }];
    rows.forEach(h => {
      options.push({ value: h.id, label: (h.name || '未命名') + ' [' + h.ip + ']' });
    });
    hostOptions.value = options;
  });
}

function handleQuery() {
  queryParams.value.pageNum = 1;
  getList();
}

function resetQuery() {
  proxy.resetForm("queryRef");
  handleQuery();
}

function handleSelectionChange(selection) {
  ids.value = selection.map(item => item.id);
  single.value = selection.length != 1;
  multiple.value = !selection.length;
}

function reset() {
  form.value = {
    id: undefined,
    name: undefined,
    ruleType: 'metric',
    metricType: undefined,
    hostId: 0,
    threshold: 80,
    compareOp: 'gt',
    duration: 60,
    severity: 'warning',
    channelIds: '[]',
    status: '0',
    remark: undefined
  };
  selectedChannels.value = [];
  proxy.resetForm("ruleRef");
}

function cancel() {
  open.value = false;
  reset();
}

function handleAdd() {
  reset();
  open.value = true;
  title.value = "新增告警规则";
}

function handleUpdate(row) {
  reset();
  const id = row.id || ids.value[0];
  const item = ruleList.value.find(r => r.id === id);
  if (item) {
    form.value = { ...item };
    // 解析channelIds
    if (item.channelIds) {
      try {
        selectedChannels.value = JSON.parse(item.channelIds);
      } catch (e) {
        selectedChannels.value = [];
      }
    }
    open.value = true;
    title.value = "修改告警规则";
  }
}

function handleDelete(row) {
  const delIds = row.id || ids.value.join(',');
  proxy.$modal.confirm('是否确认删除编号为"' + delIds + '"的数据项？').then(() => {
    return delAlarmRule(delIds);
  }).then(() => {
    getList();
    proxy.$modal.msgSuccess("删除成功");
  }).catch(() => {});
}

function handleStatusChange(row) {
  const text = row.status === "0" ? "启用" : "禁用";
  proxy.$modal.confirm('确认要"' + text + '""' + row.name + '"吗？').then(() => {
    return updateAlarmRuleStatus({ id: row.id, status: row.status });
  }).then(() => {
    proxy.$modal.msgSuccess(text + "成功");
  }).catch(() => {
    row.status = row.status === "0" ? "1" : "0";
  });
}

function onRuleTypeChange() {
  form.value.metricType = undefined;
  form.value.threshold = form.value.ruleType === 'metric' ? 80 : 0;
}

function submitForm() {
  proxy.$refs["ruleRef"].validate(valid => {
    if (valid) {
      const data = { ...form.value, channelIds: JSON.stringify(selectedChannels.value) };
      if (form.value.id != undefined) {
        updateAlarmRule(data).then(() => {
          proxy.$modal.msgSuccess("修改成功");
          open.value = false;
          getList();
        });
      } else {
        addAlarmRule(data).then(() => {
          proxy.$modal.msgSuccess("新增成功");
          open.value = false;
          getList();
        });
      }
    }
  });
}

getList();
loadChannels();
loadHosts();
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

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-left: 8px;
}
</style>
