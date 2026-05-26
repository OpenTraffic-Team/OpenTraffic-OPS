<template>
  <div class="simulation-tab">
    <el-row :gutter="16">
      <el-col v-for="item in simulationCards" :key="item.key" :xs="24" :sm="12" :md="8">
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
      <el-form :model="formData" label-width="120px" class="sim-form">
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

<script setup name="SimulationTab">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { VideoPlay, TrendCharts, Document } from '@element-plus/icons-vue'
import { runSimulation, flowSimulation, reportSimulation } from '@/api/control-agent/control'

const simulationCards = [
  {
    key: 'simulation-run',
    title: '运行仿真',
    desc: '配置并启动交通信号仿真运行',
    icon: 'VideoPlay',
    color: '#409EFF',
    api: runSimulation,
    fields: [
      { key: 'config', label: '仿真配置', type: 'json', placeholder: '请输入仿真配置（JSON格式）\n例如：\n{\n  "duration": 3600,\n  "scenario": "cityflow",\n  "roadnet": "..."\n}' },
      { key: 'algorithm', label: '算法类型', type: 'select', placeholder: '请选择算法类型', options: [{ label: '固定时间', value: 'fixed_time' }, { label: 'V1智能', value: 'v1' }, { label: '自适应', value: 'adaptive' }] }
    ]
  },
  {
    key: 'simulation-flow',
    title: '流量仿真',
    desc: '基于流量数据运行仿真分析',
    icon: 'TrendCharts',
    color: '#67C23A',
    api: flowSimulation,
    fields: [
      { key: 'flow_data', label: '流量数据', type: 'json', placeholder: '请输入流量数据（JSON格式）\n{\n  "intersections": [...],\n  "flows": [...]\n}' },
      { key: 'time_range', label: '时间范围', type: 'text', placeholder: '例如: 08:00-09:00' }
    ]
  },
  {
    key: 'simulation-report',
    title: '仿真报告',
    desc: '生成仿真运行结果报告',
    icon: 'Document',
    color: '#E6A23C',
    api: reportSimulation,
    fields: [
      { key: 'run_id', label: '运行ID', type: 'text', placeholder: '请输入仿真运行ID' },
      { key: 'metrics', label: '评估指标', type: 'json', placeholder: '请输入需要评估的指标（JSON数组，可选）' }
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
.simulation-tab {
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

.sim-form {
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
