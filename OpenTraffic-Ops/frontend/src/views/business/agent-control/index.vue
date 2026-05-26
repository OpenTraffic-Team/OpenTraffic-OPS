<template>
  <div class="app-container">
    <!-- 页头：标题 + Agent 健康状态 + 更多功能入口 -->
    <div class="page-header">
      <div class="header-top">
        <div class="header-title-row">
          <div class="title-badge">
            <el-icon :size="20"><SetUp /></el-icon>
          </div>
          <div class="title-text">
            <h3 class="page-title">控制Agent</h3>
            <p class="page-desc">交通信号控制智能体 · 算法决策 · 仿真运行 · 监控运维 · 对话交互</p>
          </div>
        </div>

        <div class="header-actions">
          <el-button
            v-if="activeTab !== 'chat'"
            type="primary"
            plain
            :icon="ChatDotRound"
            class="back-to-chat-btn"
            @click="activeTab = 'chat'"
          >
            返回聊天对话
          </el-button>
          <!-- <el-dropdown trigger="click" placement="bottom-end" @command="handleFeatureSwitch">
            <el-button type="primary" class="more-features-btn">
              <el-icon class="more-features-icon"><Grid /></el-icon>
              <span>更多功能</span>
              <el-icon class="more-features-arrow"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item
                  v-for="tab in featureTabs"
                  :key="tab.name"
                  :command="tab.name"
                  :disabled="activeTab === tab.name"
                >
                  <el-icon><component :is="tab.icon" /></el-icon>
                  <span>{{ tab.label }}</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown> -->
        </div>
      </div>

      <!-- Agent 健康状态面板 -->
      <div class="health-panel">
        <div class="health-card" :class="healthClass" @click="checkHealth">
          <div class="health-dot" />
          <div class="health-info">
            <span class="health-label">服务状态</span>
            <span class="health-value">{{ healthStatus }}</span>
          </div>
          <el-icon v-if="healthLoading" class="health-refresh spinning"><Loading /></el-icon>
          <el-icon v-else class="health-refresh"><Refresh /></el-icon>
        </div>

        <div class="health-card">
          <div class="health-icon" style="background: #EBF5FF; color: #409EFF;">
            <el-icon><Timer /></el-icon>
          </div>
          <div class="health-info">
            <span class="health-label">响应时间</span>
            <span class="health-value" :style="{ color: responseTimeColor }">{{ responseTime }}</span>
          </div>
        </div>

        <div class="health-card">
          <div class="health-icon" style="background: #F0FDF4; color: #22C55E;">
            <el-icon><Clock /></el-icon>
          </div>
          <div class="health-info">
            <span class="health-label">最后检查</span>
            <span class="health-value" style="color: #64748B;">{{ lastCheckTime }}</span>
          </div>
        </div>

        <div class="health-card">
          <div class="health-icon" style="background: #FFFBEB; color: #EAB308;">
            <el-icon><InfoFilled /></el-icon>
          </div>
          <div class="health-info">
            <span class="health-label">API 版本</span>
            <span class="health-value" style="color: #64748B;">{{ apiVersion }}</span>
          </div>
        </div>

        <div class="health-card">
          <div class="health-icon" style="background: #EFF6FF; color: #3B82F6;">
            <el-icon><Tools /></el-icon>
          </div>
          <div class="health-info">
            <span class="health-label">Tools 数量</span>
            <span class="health-value" style="color: #3B82F6;">{{ toolsCount }}</span>
          </div>
        </div>

        <div class="health-card">
          <div class="health-icon" style="background: #F5F3FF; color: #8B5CF6;">
            <el-icon><MagicStick /></el-icon>
          </div>
          <div class="health-info">
            <span class="health-label">Skills 数量</span>
            <span class="health-value" style="color: #8B5CF6;">{{ skillsCount }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 内容区：默认聊天，通过「更多功能」切换到其他三个 Tab，keep-alive 缓存上下文 -->
    <div class="tab-content">
      <keep-alive>
        <ChatTab v-if="activeTab === 'chat'" />
        <AlgorithmTab v-else-if="activeTab === 'algorithm'" />
        <SimulationTab v-else-if="activeTab === 'simulation'" />
        <MonitorTab v-else-if="activeTab === 'monitor'" />
      </keep-alive>
    </div>
  </div>
</template>

<script setup name="AgentControl">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  SetUp, Cpu, VideoPlay, Monitor, ChatDotRound,
  Loading, Refresh, Timer, Clock, InfoFilled, Tools, MagicStick,
  Grid, ArrowDown
} from '@element-plus/icons-vue'
import AlgorithmTab from './components/AlgorithmTab.vue'
import SimulationTab from './components/SimulationTab.vue'
import MonitorTab from './components/MonitorTab.vue'
import ChatTab from './components/ChatTab.vue'
import { healthCheck } from '@/api/control-agent/control'

// 「更多功能」下拉菜单项：智能算法 / 仿真运行 / 监控运维
const featureTabs = [
  { name: 'algorithm', label: '智能算法', icon: 'Cpu' },
  { name: 'simulation', label: '仿真运行', icon: 'VideoPlay' },
  { name: 'monitor', label: '监控运维', icon: 'Monitor' }
]

// 默认进入聊天对话；切换到功能 Tab 后通过「返回聊天对话」按钮回到此态
const activeTab = ref('chat')

function handleFeatureSwitch(command) {
  activeTab.value = command
}

// 健康状态
const healthLoading = ref(false)
const healthStatus = ref('检查中...')
const healthClass = ref('checking')
const responseTime = ref('--')
const responseTimeColor = ref('#64748B')
const lastCheckTime = ref('--')
const apiVersion = ref('--')
const toolsCount = ref('--')
const skillsCount = ref('--')
let healthTimer = null

async function checkHealth() {
  if (healthLoading.value) return
  healthLoading.value = true
  healthStatus.value = '检查中...'
  healthClass.value = 'checking'

  const startTime = performance.now()
  try {
    // healthz 返回裸 JSON：{ status, service, version, tools_count, skills_count }
    const res = await healthCheck()
    const elapsed = Math.round(performance.now() - startTime)

    healthStatus.value = '正常'
    healthClass.value = 'healthy'
    responseTime.value = `${elapsed}ms`
    responseTimeColor.value = elapsed < 300 ? '#22C55E' : elapsed < 1000 ? '#EAB308' : '#EF4444'
    lastCheckTime.value = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })

    // 提取接口返回的字段
    apiVersion.value = res?.version || '--'
    toolsCount.value = res?.tools_count ?? '--'
    skillsCount.value = res?.skills_count ?? '--'
  } catch (error) {
    healthStatus.value = '异常'
    healthClass.value = 'unhealthy'
    responseTime.value = '超时'
    responseTimeColor.value = '#EF4444'
    lastCheckTime.value = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
    apiVersion.value = '--'
    toolsCount.value = '--'
    skillsCount.value = '--'
    ElMessage.warning('Agent 服务连接异常，请检查网络或服务状态')
  } finally {
    healthLoading.value = false
  }
}

onMounted(() => {
  checkHealth()
  // 每 30 秒自动刷新
  healthTimer = setInterval(checkHealth, 30000)
})

onUnmounted(() => {
  if (healthTimer) {
    clearInterval(healthTimer)
  }
})
</script>

<style scoped lang="scss">
.app-container {
  padding: 20px;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  /* 96 = navbar(56) + tags-view(40)，与全局 .app-container 约束一致 */
  height: calc(100vh - 96px);
  min-height: 600px;
  overflow: hidden;
}

// 页头
.page-header {
  flex-shrink: 0;
  background: #FFFFFF;
  border-radius: 16px;
  padding: 24px 28px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 1px 2px rgba(0, 0, 0, 0.03);
  border: 1px solid #E2E8F0;
}

// 页头顶部：标题 + 操作区（更多功能 / 返回聊天）
.header-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 20px;

  .header-title-row {
    display: flex;
    align-items: center;
    gap: 14px;

    .title-badge {
      width: 44px;
      height: 44px;
      border-radius: 12px;
      background: #EBF5FF;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #409EFF;
    }

    .title-text {
      .page-title {
        margin: 0;
        font-size: 20px;
        font-weight: 700;
        color: #1E293B;
        letter-spacing: -0.3px;
      }

      .page-desc {
        margin: 4px 0 0 0;
        font-size: 13px;
        color: #94A3B8;
      }
    }
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-shrink: 0;
  }

  .back-to-chat-btn {
    border-radius: 12px;
    padding: 9px 18px;
    font-weight: 500;
  }

  .more-features-btn {
    border-radius: 12px;
    padding: 9px 18px;
    font-weight: 500;
    display: inline-flex;
    align-items: center;
    gap: 6px;

    .more-features-icon {
      font-size: 16px;
    }

    .more-features-arrow {
      font-size: 12px;
      margin-left: 2px;
    }
  }
}

// 健康状态面板
.health-panel {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;

  .health-card {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 16px;
    border-radius: 12px;
    background: #F8FAFC;
    cursor: pointer;
    transition: all 0.3s;
    border: 1px solid #E2E8F0;

    &:hover {
      background: #F1F5F9;
      border-color: #CBD5E1;
    }

    &.healthy {
      border-color: rgba(34, 197, 94, 0.4);
      background: #F0FDF4;
      .health-dot {
        background: #22C55E;
        box-shadow: 0 0 8px rgba(34, 197, 94, 0.4);
      }
    }

    &.unhealthy {
      border-color: rgba(239, 68, 68, 0.4);
      background: #FEF2F2;
      .health-dot {
        background: #EF4444;
        box-shadow: 0 0 8px rgba(239, 68, 68, 0.4);
      }
    }

    &.checking {
      .health-dot {
        background: #94A3B8;
        animation: pulse 1.5s infinite;
      }
    }

    .health-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      flex-shrink: 0;
    }

    .health-icon {
      width: 32px;
      height: 32px;
      border-radius: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;
    }

    .health-info {
      display: flex;
      flex-direction: column;
      gap: 2px;

      .health-label {
        font-size: 11px;
        color: #94A3B8;
        font-weight: 500;
      }

      .health-value {
        font-size: 13px;
        font-weight: 600;
        color: #1E293B;
      }
    }

    .health-refresh {
      font-size: 13px;
      color: #94A3B8;
      margin-left: 4px;

      &.spinning {
        animation: spin 1s linear infinite;
      }
    }
  }
}

// 内容区
.tab-content {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: #FFFFFF;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 1px 2px rgba(0, 0, 0, 0.03);
  overflow: hidden;

  // 让内部 Tab 组件（keep-alive 的子节点）撑满整个内容区，并允许内部滚动
  > * {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
  }

  // 聊天 Tab 自管理内部滚动（左侧历史列表 + 右侧消息列表），外层禁用滚动
  > .chat-tab {
    overflow: hidden;
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
