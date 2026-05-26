<template>
  <div class="algorithm-tab">
    <el-row :gutter="16">
      <el-col v-for="item in algorithmCards" :key="item.key" :xs="24" :sm="12" :md="8" :lg="6">
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
      <el-form :model="formData" label-width="120px" class="algo-form">
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

<script setup name="AlgorithmTab">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Cpu, DataLine, DocumentCopy, Timer, SetUp,
  DataAnalysis, MapLocation, CircleCheck
} from '@element-plus/icons-vue'
import {
  analyzeParams, v1Decision, configDiff, configFixedTime,
  configV1, modelingParse, modelingRoadnet, modelingValidate
} from '@/api/control-agent/control'

const algorithmCards = [
  {
    key: 'analyze-params',
    title: '参数分析',
    desc: '分析交通信号控制参数配置',
    icon: 'Cpu',
    color: '#409EFF',
    api: analyzeParams,
    fields: [
      { key: 'config', label: '配置数据', type: 'json', placeholder: '请输入JSON格式的配置数据，例如：\n{\n  "cycle_time": 120,\n  "phases": [...]\n}' }
    ]
  },
  {
    key: 'v1-decision',
    title: '决策算法',
    desc: '运行V1智能决策算法',
    icon: 'DataLine',
    color: '#67C23A',
    api: v1Decision,
    fields: [
      { key: 'intersection_id', label: '路口ID', type: 'text', placeholder: '请输入路口ID' },
      { key: 'traffic_data', label: '交通数据', type: 'json', placeholder: '请输入交通流量数据（JSON格式）' },
      { key: 'strategy', label: '策略参数', type: 'json', placeholder: '请输入策略参数（可选，JSON格式）' }
    ]
  },
  {
    key: 'config-diff',
    title: '配置对比',
    desc: '对比两个配置文件的差异',
    icon: 'DocumentCopy',
    color: '#E6A23C',
    api: configDiff,
    fields: [
      { key: 'config_a', label: '配置A', type: 'json', placeholder: '请输入第一个配置（JSON格式）' },
      { key: 'config_b', label: '配置B', type: 'json', placeholder: '请输入第二个配置（JSON格式）' }
    ]
  },
  {
    key: 'fixed-time',
    title: '固定时间配置',
    desc: '生成固定时间信号配置方案',
    icon: 'Timer',
    color: '#F56C6C',
    api: configFixedTime,
    fields: [
      { key: 'cycle_time', label: '周期时长', type: 'number', placeholder: '请输入周期时长（秒）' },
      { key: 'phases', label: '相位配置', type: 'json', placeholder: '请输入相位配置（JSON格式）' }
    ]
  },
  {
    key: 'config-v1',
    title: 'V1配置',
    desc: '生成V1算法配置方案',
    icon: 'SetUp',
    color: '#909399',
    api: configV1,
    fields: [
      { key: 'params', label: '参数配置', type: 'json', placeholder: '请输入V1算法参数配置（JSON格式）' }
    ]
  },
  {
    key: 'modeling-parse',
    title: '建模解析',
    desc: '解析交通网络模型文件',
    icon: 'DataAnalysis',
    color: '#409EFF',
    api: modelingParse,
    fields: [
      { key: 'model_data', label: '模型数据', type: 'textarea', rows: 6, placeholder: '请输入模型数据或上传文件内容' },
      { key: 'format', label: '格式类型', type: 'select', placeholder: '请选择模型格式', options: [{ label: 'SUMO', value: 'sumo' }, { label: 'CityFlow', value: 'cityflow' }, { label: 'JSON', value: 'json' }] }
    ]
  },
  {
    key: 'modeling-roadnet',
    title: '路网建模',
    desc: '构建交通路网模型',
    icon: 'MapLocation',
    color: '#67C23A',
    api: modelingRoadnet,
    fields: [
      { key: 'roadnet_data', label: '路网数据', type: 'json', placeholder: '请输入路网数据（JSON格式）' },
      { key: 'intersections', label: '路口列表', type: 'json', placeholder: '请输入路口列表（JSON数组）' }
    ]
  },
  {
    key: 'modeling-validate',
    title: '模型验证',
    desc: '验证交通模型有效性',
    icon: 'CircleCheck',
    color: '#E6A23C',
    api: modelingValidate,
    fields: [
      { key: 'model', label: '模型数据', type: 'json', placeholder: '请输入需要验证的模型数据（JSON格式）' }
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
  // 重置表单
  Object.keys(formData).forEach(key => delete formData[key])
  card.fields.forEach(field => {
    formData[field.key] = field.type === 'number' ? undefined : ''
  })
  drawerVisible.value = true
}

async function handleSubmit() {
  if (!currentCard.value) return

  // 简单校验
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
    // 尝试自动解析 JSON 字段
    const payload = {}
    for (const field of currentCard.value.fields) {
      let val = formData[field.key]
      if (field.type === 'json') {
        try {
          val = JSON.parse(val)
        } catch {
          // 解析失败保持原样
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
.algorithm-tab {
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

.algo-form {
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
