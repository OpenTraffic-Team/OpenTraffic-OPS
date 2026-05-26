<template>
  <div class="monitor-tab">
    <el-row :gutter="16">
      <el-col v-for="item in monitorCards" :key="item.key" :xs="24" :sm="12" :md="8" :lg="6">
        <el-card shadow="hover" class="func-card" @click="openDrawer(item)">
          <div class="card-icon" :style="{ backgroundColor: item.color + '15', color: item.color }">
            <el-icon :size="28">
              <component :is="item.icon" />
            </el-icon>
          </div>
          <div class="card-title">{{ item.title }}</div>
          <div class="card-desc">{{ item.desc }}</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 表单抽屉 -->
    <el-drawer v-model="drawerVisible" :title="currentCard?.title" direction="rtl" size="80%"
      destroy-on-close class="func-drawer">
      <el-form :model="formData" label-width="120px" class="mon-form">
        <el-form-item v-for="field in currentCard?.fields" :key="field.key" :label="field.label"
          :prop="field.key">
          <el-input v-if="field.type === 'textarea'" v-model="formData[field.key]" type="textarea"
            :rows="field.rows || 4" :placeholder="field.placeholder" />
          <el-input v-else-if="field.type === 'json'" v-model="formData[field.key]" type="textarea"
            :rows="6" :placeholder="field.placeholder" />
          <el-input v-else-if="field.type === 'text'" v-model="formData[field.key]"
            :placeholder="field.placeholder" />
          <el-input-number v-else-if="field.type === 'number'" v-model="formData[field.key]"
            :placeholder="field.placeholder" style="width: 100%;" />
          <el-select v-else-if="field.type === 'select'" v-model="formData[field.key]"
            :placeholder="field.placeholder" style="width: 100%;">
            <el-option v-for="opt in field.options" :key="opt.value" :label="opt.label"
              :value="opt.value" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="drawerVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">执行</el-button>
      </template>
    </el-drawer>

    <!-- 结果抽屉 -->
    <el-drawer v-model="resultDrawerVisible" title="执行结果" direction="rtl" size="80%"
      destroy-on-close class="result-drawer">
      <div v-if="resultLoading" class="result-loading">
        <el-skeleton :rows="6" animated />
      </div>
      <div v-else class="result-content">
        <el-alert v-if="resultError" :title="resultError" type="error" show-icon :closable="false" />
        <div v-else-if="resultData" class="json-result">
          <vue-json-viewer :value="resultData" :expand-depth="3" copyable boxed sort />
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup name="MonitorTab">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Warning, CircleCheck, DataLine, RefreshRight, ArrowDown,
  DocumentChecked, Operation
} from '@element-plus/icons-vue'
import {
  monitoringAnomaly, monitoringEffectiveness, monitoringImpute,
  monitoringRecovery, monitoringDegrade, opsAnalyzeLog, opsSuggestTuning
} from '@/api/control-agent/control'

const monitorCards = [
  {
    key: 'anomaly',
    title: '异常检测',
    desc: '检测交通数据中的异常模式',
    icon: 'Warning',
    color: '#F56C6C',
    api: monitoringAnomaly,
    fields: [
      { key: 'data', label: '监测数据', type: 'json', placeholder: '请输入监测数据（JSON格式）\n例如：\n{\n  "metrics": [...],\n  "threshold": 0.95\n}' }
    ]
  },
  {
    key: 'effectiveness',
    title: '有效性评估',
    desc: '评估信号控制策略效果',
    icon: 'CircleCheck',
    color: '#67C23A',
    api: monitoringEffectiveness,
    fields: [
      { key: 'strategy_id', label: '策略ID', type: 'text', placeholder: '请输入策略ID' },
      { key: 'metrics', label: '评估指标', type: 'json', placeholder: '请输入评估指标数据（JSON格式）' }
    ]
  },
  {
    key: 'impute',
    title: '缺失值插补',
    desc: '对缺失交通数据进行插补修复',
    icon: 'DataLine',
    color: '#409EFF',
    api: monitoringImpute,
    fields: [
      { key: 'dataset', label: '数据集', type: 'json', placeholder: '请输入包含缺失值的数据集（JSON格式）' },
      { key: 'method', label: '插补方法', type: 'select', placeholder: '请选择插补方法', options: [{ label: '线性插值', value: 'linear' }, { label: '均值填充', value: 'mean' }, { label: 'KNN', value: 'knn' }] }
    ]
  },
  {
    key: 'recovery',
    title: '恢复监控',
    desc: '监控信号控制系统恢复状态',
    icon: 'RefreshRight',
    color: '#67C23A',
    api: monitoringRecovery,
    fields: [
      { key: 'system_id', label: '系统ID', type: 'text', placeholder: '请输入系统ID' }
    ]
  },
  {
    key: 'degrade',
    title: '降级监控',
    desc: '监控系统降级运行状态',
    icon: 'ArrowDown',
    color: '#E6A23C',
    api: monitoringDegrade,
    fields: [
      { key: 'system_id', label: '系统ID', type: 'text', placeholder: '请输入系统ID' },
      { key: 'mode', label: '降级模式', type: 'select', placeholder: '请选择降级模式', options: [{ label: '固定周期', value: 'fixed_cycle' }, { label: '黄闪', value: 'flash' }, { label: '全红', value: 'all_red' }] }
    ]
  },
  {
    key: 'analyze-log',
    title: '日志分析',
    desc: '分析系统运行日志',
    icon: 'DocumentChecked',
    color: '#909399',
    api: opsAnalyzeLog,
    fields: [
      { key: 'logs', label: '日志内容', type: 'textarea', rows: 6, placeholder: '请输入或粘贴日志内容' },
      { key: 'analysis_type', label: '分析类型', type: 'select', placeholder: '请选择分析类型', options: [{ label: '错误分析', value: 'error' }, { label: '性能分析', value: 'performance' }, { label: '异常检测', value: 'anomaly' }] }
    ]
  },
  {
    key: 'suggest-tuning',
    title: '调优建议',
    desc: '获取信号参数调优建议',
    icon: 'Operation',
    color: '#409EFF',
    api: opsSuggestTuning,
    fields: [
      { key: 'current_config', label: '当前配置', type: 'json', placeholder: '请输入当前信号配置（JSON格式）' },
      { key: 'performance_data', label: '性能数据', type: 'json', placeholder: '请输入性能监测数据（JSON格式）' }
    ]
  }
]

const drawerVisible = ref(false)
const resultDrawerVisible = ref(false)
const submitting = ref(false)
const resultLoading = ref(false)
const resultData = ref(null)
const resultError = ref('')
const currentCard = ref(null)
const formData = reactive({})

function openDrawer(card) {
  currentCard.value = card
  Object.keys(formData).forEach(key => delete formData[key])
  card.fields.forEach(field => {
    formData[field.key] = field.type === 'number' ? undefined : ''
  })
  drawerVisible.value = true
}

async function handleSubmit() {
  if (!currentCard.value) return

  for (const field of currentCard.value.fields) {
    const val = formData[field.key]
    if (val === '' || val === undefined || val === null) {
      ElMessage.warning(`请填写 ${field.label}`)
      return
    }
  }

  submitting.value = true
  drawerVisible.value = false
  resultDrawerVisible.value = true
  resultLoading.value = true
  resultError.value = ''
  resultData.value = null

  try {
    const payload = {}
    for (const field of currentCard.value.fields) {
      let val = formData[field.key]
      if (field.type === 'json') {
        try {
          val = JSON.parse(val)
        } catch {
          // keep as is
        }
      }
      payload[field.key] = val
    }

    const res = await currentCard.value.api(payload)
    resultData.value = res.data || res
  } catch (error) {
    resultError.value = error?.message || '请求失败，请检查网络连接或Agent服务状态'
  } finally {
    resultLoading.value = false
    submitting.value = false
  }
}
</script>

<style scoped lang="scss">
.monitor-tab {
  .func-card {
    cursor: pointer;
    border-radius: 12px;
    margin-bottom: 16px;
    transition: all 0.3s ease;
    border: 1px solid #E2E8F0;

    &:hover {
      transform: translateY(-4px);
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
      border-color: #409EFF;
    }

    :deep(.el-card__body) {
      padding: 24px;
      text-align: center;
    }

    .card-icon {
      width: 56px;
      height: 56px;
      border-radius: 14px;
      display: flex;
      align-items: center;
      justify-content: center;
      margin: 0 auto 16px;
    }

    .card-title {
      font-size: 15px;
      font-weight: 600;
      color: #1E293B;
      margin-bottom: 8px;
    }

    .card-desc {
      font-size: 12px;
      color: #94A3B8;
      line-height: 1.5;
    }
  }
}

.mon-form {
  :deep(.el-form-item__label) {
    font-weight: 500;
    color: #475569;
  }
}

.result-content {
  .json-result {
    max-height: 500px;
    overflow: auto;

    :deep(.jv-container) {
      background: #F8FAFC;
      border-radius: 8px;
    }
  }
}

.result-loading {
  padding: 20px;
}
</style>
