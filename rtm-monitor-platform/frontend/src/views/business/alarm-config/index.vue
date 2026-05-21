<template>
  <div class="app-container">
    <div class="header-container">
      <el-form :model="queryParams" ref="queryRef" :inline="true" v-show="showSearch" label-width="68px">
        <el-form-item label="通道名称" prop="name">
          <el-input v-model="queryParams.name" placeholder="请输入通道名称" clearable style="width: 240px" @keyup.enter="handleQuery" />
        </el-form-item>
        <el-form-item label="通道类型" prop="channelType">
          <el-select v-model="queryParams.channelType" placeholder="通道类型" clearable style="width: 240px">
            <el-option label="邮件" value="email" />
            <el-option label="钉钉" value="dingtalk" />
            <el-option label="企业微信" value="wechat" />
            <el-option label="平台内部" value="platform" />
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
      <el-table v-loading="loading" :data="channelList" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="60" align="center" />
        <el-table-column label="编号" align="center" prop="id" v-if="columns[0].visible" width="100" />
        <el-table-column label="通道名称" align="center" prop="name" v-if="columns[1].visible" min-width="220" :show-overflow-tooltip="true" />
        <el-table-column label="通道类型" align="center" prop="channelType" v-if="columns[2].visible" width="160">
          <template #default="scope">
            <el-tag :type="channelTypeTag(scope.row.channelType)">{{ channelTypeLabel(scope.row.channelType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" align="center" prop="status" v-if="columns[3].visible" width="160">
          <template #default="scope">
            <el-switch v-model="scope.row.status" active-value="0" inactive-value="1" @change="handleStatusChange(scope.row)" />
          </template>
        </el-table-column>
        <el-table-column label="默认" align="center" prop="isDefault" v-if="columns[4].visible" width="160">
          <template #default="scope">
            <el-tag v-if="scope.row.isDefault === '1'" type="success">是</el-tag>
            <el-tag v-else type="info">否</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="备注" align="center" prop="remark" v-if="columns[5].visible" :show-overflow-tooltip="true" />
        <el-table-column label="创建时间" align="center" prop="createTime" v-if="columns[6].visible" width="220" />
        <el-table-column label="操作" align="center" width="220" class-name="small-padding fixed-width">
          <template #default="scope">
            <el-tooltip content="修改" placement="top">
              <el-button link type="primary" icon="Edit" @click="handleUpdate(scope.row)"></el-button>
            </el-tooltip>
            <el-tooltip content="删除" placement="top">
              <el-button link type="danger" icon="Delete" @click="handleDelete(scope.row)"></el-button>
            </el-tooltip>
            <el-tooltip content="设为默认" placement="top">
              <el-button link type="primary" icon="Star" @click="handleSetDefault(scope.row)"></el-button>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>
      <pagination v-show="total > 0" :total="total" v-model:page="queryParams.pageNum" v-model:limit="queryParams.pageSize" @pagination="getList" />
    </div>

    <!-- 新增/编辑弹窗 -->
    <el-dialog :title="title" v-model="open" width="600px" append-to-body>
      <el-form :model="form" :rules="rules" ref="channelRef" label-width="100px">
        <el-form-item label="通道名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入通道名称" maxlength="100" />
        </el-form-item>
        <el-form-item label="通道类型" prop="channelType">
          <el-select v-model="form.channelType" placeholder="请选择通道类型" style="width: 100%" @change="onChannelTypeChange">
            <el-option label="邮件" value="email" />
            <el-option label="钉钉" value="dingtalk" />
            <el-option label="企业微信" value="wechat" />
            <el-option label="平台内部" value="platform" />
          </el-select>
        </el-form-item>

        <!-- 动态配置表单 -->
        <div v-if="form.channelType === 'email'">
          <el-form-item label="SMTP服务器">
            <el-input v-model="configForm.smtpHost" placeholder="如: smtp.qq.com" />
          </el-form-item>
          <el-form-item label="SMTP端口">
            <el-input v-model="configForm.smtpPort" placeholder="如: 587" />
          </el-form-item>
          <el-form-item label="发件邮箱">
            <el-input v-model="configForm.fromEmail" placeholder="如: alert@example.com" />
          </el-form-item>
          <el-form-item label="授权码/密码">
            <el-input v-model="configForm.password" type="password" placeholder="SMTP授权码" show-password />
          </el-form-item>
          <el-form-item label="收件邮箱">
            <el-input v-model="configForm.toEmails" placeholder="多个邮箱用英文逗号分隔" />
          </el-form-item>
        </div>

        <div v-else-if="form.channelType === 'dingtalk'">
          <el-form-item label="Webhook">
            <el-input v-model="configForm.webhook" placeholder="钉钉机器人Webhook地址" />
          </el-form-item>
          <el-form-item label="密钥">
            <el-input v-model="configForm.secret" type="password" placeholder="钉钉机器人密钥（可选）" show-password />
          </el-form-item>
        </div>

        <div v-else-if="form.channelType === 'wechat'">
          <el-form-item label="Webhook">
            <el-input v-model="configForm.webhook" placeholder="企业微信机器人Webhook地址" />
          </el-form-item>
        </div>

        <div v-else-if="form.channelType === 'platform'">
          <el-alert title="平台内部通知无需额外配置" type="info" :closable="false" />
        </div>

        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio label="0">启用</el-radio>
            <el-radio label="1">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="默认通道">
          <el-radio-group v-model="form.isDefault">
            <el-radio label="1">是</el-radio>
            <el-radio label="0">否</el-radio>
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

<script setup name="AlarmConfig">
import { listAlarmChannel, delAlarmChannel, addAlarmChannel, updateAlarmChannel, updateAlarmChannelStatus, setAlarmChannelDefault } from "@/api/business/alarmChannel";

const { proxy } = getCurrentInstance();

const channelList = ref([]);
const open = ref(false);
const loading = ref(true);
const showSearch = ref(true);
const ids = ref([]);
const single = ref(true);
const multiple = ref(true);
const total = ref(0);
const title = ref("");

const columns = ref([
  { key: 0, label: '编号', visible: true },
  { key: 1, label: '通道名称', visible: true },
  { key: 2, label: '通道类型', visible: true },
  { key: 3, label: '状态', visible: true },
  { key: 4, label: '默认', visible: true },
  { key: 5, label: '备注', visible: false },
  { key: 6, label: '创建时间', visible: true }
]);

const data = reactive({
  form: {},
  queryParams: {
    pageNum: 1,
    pageSize: 10,
    name: undefined,
    channelType: undefined,
    status: undefined
  },
  rules: {
    name: [{ required: true, message: "通道名称不能为空", trigger: "blur" }],
    channelType: [{ required: true, message: "通道类型不能为空", trigger: "change" }]
  }
});

const configForm = reactive({
  smtpHost: '',
  smtpPort: '',
  fromEmail: '',
  password: '',
  toEmails: '',
  webhook: '',
  secret: ''
});

const { queryParams, form, rules } = toRefs(data);

const channelTypeMap = {
  'email': { label: '邮件', tag: '' },
  'dingtalk': { label: '钉钉', tag: 'primary' },
  'wechat': { label: '企业微信', tag: 'success' },
  'platform': { label: '平台内部', tag: 'info' }
};

function channelTypeLabel(type) {
  return channelTypeMap[type]?.label || type;
}

function channelTypeTag(type) {
  return channelTypeMap[type]?.tag || '';
}

function getList() {
  loading.value = true;
  listAlarmChannel(queryParams.value).then(res => {
    loading.value = false;
    channelList.value = res.rows || res.data?.rows || [];
    total.value = res.total || res.data?.total || 0;
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
    channelType: undefined,
    config: '{}',
    status: '0',
    isDefault: '0',
    remark: undefined
  };
  Object.keys(configForm).forEach(key => configForm[key] = '');
  proxy.resetForm("channelRef");
}

function cancel() {
  open.value = false;
  reset();
}

function handleAdd() {
  reset();
  open.value = true;
  title.value = "新增告警通道";
}

function handleUpdate(row) {
  reset();
  const id = row.id || ids.value[0];
  // 从列表中找到对应数据
  const item = channelList.value.find(c => c.id === id);
  if (item) {
    form.value = { ...item };
    // 解析config JSON
    if (item.config) {
      try {
        const cfg = JSON.parse(item.config);
        Object.keys(cfg).forEach(key => {
          if (configForm.hasOwnProperty(key)) {
            configForm[key] = cfg[key];
          }
        });
      } catch (e) {}
    }
    open.value = true;
    title.value = "修改告警通道";
  }
}

function handleDelete(row) {
  const delIds = row.id || ids.value.join(',');
  proxy.$modal.confirm('是否确认删除编号为"' + delIds + '"的数据项？').then(() => {
    return delAlarmChannel(delIds);
  }).then(() => {
    getList();
    proxy.$modal.msgSuccess("删除成功");
  }).catch(() => {});
}

function handleStatusChange(row) {
  const text = row.status === "0" ? "启用" : "禁用";
  proxy.$modal.confirm('确认要"' + text + '""' + row.name + '"吗？').then(() => {
    return updateAlarmChannelStatus({ id: row.id, status: row.status });
  }).then(() => {
    proxy.$modal.msgSuccess(text + "成功");
  }).catch(() => {
    row.status = row.status === "0" ? "1" : "0";
  });
}

function handleSetDefault(row) {
  proxy.$modal.confirm('确认要将"' + row.name + '"设为默认通道吗？').then(() => {
    return setAlarmChannelDefault({ id: row.id, isDefault: '1' });
  }).then(() => {
    getList();
    proxy.$modal.msgSuccess("设置成功");
  }).catch(() => {});
}

function onChannelTypeChange() {
  // 切换通道类型时清空配置
  Object.keys(configForm).forEach(key => configForm[key] = '');
}

function buildConfigJson() {
  const type = form.value.channelType;
  const cfg = {};
  if (type === 'email') {
    cfg.smtpHost = configForm.smtpHost;
    cfg.smtpPort = configForm.smtpPort;
    cfg.fromEmail = configForm.fromEmail;
    cfg.password = configForm.password;
    cfg.toEmails = configForm.toEmails;
  } else if (type === 'dingtalk') {
    cfg.webhook = configForm.webhook;
    cfg.secret = configForm.secret;
  } else if (type === 'wechat') {
    cfg.webhook = configForm.webhook;
  }
  return JSON.stringify(cfg);
}

function validateChannelConfig() {
  const type = form.value.channelType;
  if (type === 'email') {
    if (!configForm.smtpHost || !configForm.smtpPort || !configForm.fromEmail || !configForm.password || !configForm.toEmails) {
      proxy.$modal.msgError("请补全邮件配置");
      return false;
    }
  } else if (type === 'dingtalk') {
    if (!configForm.webhook) {
      proxy.$modal.msgError("请填写钉钉Webhook");
      return false;
    }
  } else if (type === 'wechat') {
    if (!configForm.webhook) {
      proxy.$modal.msgError("请填写企业微信Webhook");
      return false;
    }
  }
  return true;
}

function submitForm() {
  proxy.$refs["channelRef"].validate(valid => {
    if (valid) {
      if (!validateChannelConfig()) {
        return;
      }
      const data = { ...form.value, config: buildConfigJson() };
      if (form.value.id != undefined) {
        updateAlarmChannel(data).then(() => {
          proxy.$modal.msgSuccess("修改成功");
          open.value = false;
          getList();
        });
      } else {
        addAlarmChannel(data).then(() => {
          proxy.$modal.msgSuccess("新增成功");
          open.value = false;
          getList();
        });
      }
    }
  });
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

.table-container {
  background: #FFFFFF;
  padding: 20px 24px;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 1px 2px rgba(0, 0, 0, 0.03);
  border: 1px solid #E2E8F0;
}
</style>
