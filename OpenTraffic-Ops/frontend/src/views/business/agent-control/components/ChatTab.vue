<template>
  <div class="chat-tab">
    <!-- 左侧对话历史 -->
    <div class="chat-sidebar">
      <div class="sidebar-header">
        <el-button type="primary" :icon="Plus" :disabled="sending" @click="createNewChat">新建对话</el-button>
      </div>
      <div class="chat-list">
        <div v-for="(chat, index) in chatHistory" :key="chat.id"
          :class="['chat-item', currentChatId === chat.id ? 'active' : '']"
          @click="switchChat(chat.id)">
          <el-icon class="chat-icon"><ChatLineRound /></el-icon>
          <span class="chat-title">{{ chat.title || `对话 ${index + 1}` }}</span>
          <el-icon class="delete-icon" @click.stop="deleteChat(chat.id)"><Close /></el-icon>
        </div>
        <el-empty v-if="chatHistory.length === 0" description="暂无对话" />
      </div>
    </div>

    <!-- 右侧对话区 -->
    <div class="chat-main">
      <!-- 消息列表 -->
      <div ref="messageContainer" class="message-list">
        <div v-if="currentMessages.length === 0" class="welcome-area">
          <div class="welcome-icon">
            <el-icon :size="48" color="#409EFF"><ChatDotRound /></el-icon>
          </div>
          <h3 class="welcome-title">Agent 智能助手</h3>
          <p class="welcome-desc">我可以帮您进行交通信号控制相关的算法分析、仿真配置和运维咨询</p>
          <div class="quick-prompts">
            <el-tag v-for="prompt in quickPrompts" :key="prompt" class="quick-tag"
              effect="plain" @click="sendQuickPrompt(prompt)">
              {{ prompt }}
            </el-tag>
          </div>
        </div>

        <div v-for="msg in currentMessages" :key="msg.id" :class="['message-item', msg.role]">
          <div class="message-avatar">
            <el-avatar v-if="msg.role === 'user'" :size="36" :icon="UserFilled"
              :style="{ backgroundColor: '#409EFF' }" />
            <el-avatar v-else :size="36" :icon="ChatDotRound"
              :style="{ backgroundColor: '#67C23A' }" />
          </div>
          <div class="message-content">
            <div v-if="!msg.content && msg.streaming" class="thinking-bubble">
              <span class="dot"></span>
              <span class="dot"></span>
              <span class="dot"></span>
              <span class="thinking-text">正在思考中…</span>
            </div>
            <div v-else class="message-bubble" v-html="renderMarkdown(msg.content)"></div>
          </div>
        </div>
      </div>

      <!-- 输入区 -->
      <div class="chat-input-area">
        <div class="input-box">
          <el-input
            v-model="inputMessage"
            type="textarea"
            :autosize="{ minRows: 1, maxRows: 6 }"
            resize="none"
            placeholder="请输入您的问题，Enter 发送 / Shift + Enter 换行"
            @keydown="handleKeydown"
          />
          <div class="input-actions">
            <el-tooltip content="清除当前对话" placement="top">
              <el-button :icon="Delete" circle size="small" class="clear-btn" @click="clearCurrentChat" />
            </el-tooltip>
            <el-button
              type="primary"
              :icon="Promotion"
              :loading="sending"
              :disabled="!inputMessage.trim()"
              class="send-btn"
              @click="sendMessage"
            >
              发送
            </el-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup name="ChatTab">
import { ref, computed, nextTick, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus, ChatLineRound, Close, UserFilled,
  ChatDotRound, Delete, Promotion
} from '@element-plus/icons-vue'
import {
  chatCompletions,
  listChatSessions,
  getChatSessionDetail,
  saveChatTurn,
  deleteChatSessions
} from '@/api/control-agent/control'
import MarkdownIt from 'markdown-it'

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true
})

// 已渲染 markdown 的缓存：消息内容串 → HTML。避免每次 Vue re-render
// 都把所有历史消息重新跑一遍 markdown-it。
const renderCache = new Map()
function renderMarkdown(text) {
  if (!text) return ''
  const cached = renderCache.get(text)
  if (cached !== undefined) return cached
  const html = md.render(text)
  renderCache.set(text, html)
  return html
}

// 快速提示
const quickPrompts = [
  '如何配置固定时间信号方案？',
  '分析一下这个路口的流量数据',
  '帮我生成一个仿真配置',
  '信号控制异常如何排查？'
]

// 对话历史：每项 { id, title, agentSessionId, lastMessageAt, isLocal? }
// id 为服务端会话主键（number）或本地占位 'local-<ts>'（isLocal: true）
const chatHistory = ref([])
const currentChatId = ref(null)

// 消息存储 { chatId: [messages] }
const messagesMap = ref({})
// 正在进行的会话消息拉取请求：id → Promise，避免并发期间重复 GET，
// 也用于 sendMessage 时判断会话是否处于加载中。
const loadingSessions = new Map()

const currentMessages = computed(() => {
  if (currentChatId.value == null) return []
  return messagesMap.value[currentChatId.value] || []
})

const inputMessage = ref('')
const sending = ref(false)
const messageContainer = ref(null)

let messageIdCounter = 0

function scrollToBottom() {
  nextTick(() => {
    if (messageContainer.value) {
      messageContainer.value.scrollTop = messageContainer.value.scrollHeight
    }
  })
}

function createNewChat() {
  const id = 'local-' + Date.now()
  chatHistory.value.unshift({
    id,
    title: '新对话',
    agentSessionId: '',
    lastMessageAt: '',
    isLocal: true
  })
  messagesMap.value[id] = []
  currentChatId.value = id
}

async function loadSessionMessages(id) {
  // 已有消息（用户已发送或之前已加载）→ 不再拉取
  const existing = messagesMap.value[id]
  if (Array.isArray(existing) && existing.length > 0) return
  // 同一会话拉取中 → 复用 in-flight Promise
  if (loadingSessions.has(id)) return loadingSessions.get(id)

  const task = (async () => {
    try {
      const res = await getChatSessionDetail(id)
      const detail = res?.data ?? res
      // await 期间用户可能已经发了消息，避免覆盖
      const current = messagesMap.value[id]
      if (Array.isArray(current) && current.length > 0) return
      const msgs = (detail?.messages || []).map(m => ({
        id: ++messageIdCounter,
        role: m.role,
        content: m.content
      }))
      messagesMap.value[id] = msgs
      // 同步会话头部信息（标题等服务端可能与本地不一致）
      const chat = chatHistory.value.find(c => c.id === id)
      if (chat && detail) {
        chat.title = detail.title || chat.title
        chat.agentSessionId = detail.agentSessionId || chat.agentSessionId
        chat.lastMessageAt = detail.lastMessageAt || chat.lastMessageAt
      }
    } catch (e) {
      ElMessage.error('加载会话消息失败')
    } finally {
      loadingSessions.delete(id)
    }
  })()
  loadingSessions.set(id, task)
  return task
}

async function switchChat(id) {
  if (currentChatId.value === id) return
  currentChatId.value = id
  const chat = chatHistory.value.find(c => c.id === id)
  if (chat && !chat.isLocal) {
    await loadSessionMessages(id)
  }
  scrollToBottom()
}

async function deleteChat(id) {
  const chat = chatHistory.value.find(c => c.id === id)
  if (!chat) return

  if (!chat.isLocal) {
    try {
      await deleteChatSessions(id)
    } catch (e) {
      ElMessage.error('删除会话失败')
      return
    }
  }

  const idx = chatHistory.value.findIndex(c => c.id === id)
  if (idx !== -1) chatHistory.value.splice(idx, 1)
  delete messagesMap.value[id]

  if (currentChatId.value === id) {
    if (chatHistory.value.length > 0) {
      const next = chatHistory.value[0]
      currentChatId.value = next.id
      if (!next.isLocal) await loadSessionMessages(next.id)
    } else {
      createNewChat()
    }
  }
}

async function clearCurrentChat() {
  const id = currentChatId.value
  if (id == null) return
  const chat = chatHistory.value.find(c => c.id === id)
  if (!chat) return

  try {
    await ElMessageBox.confirm('确定要清除当前对话吗？此操作不可恢复。', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }

  if (chat.isLocal) {
    messagesMap.value[id] = []
    return
  }

  try {
    await deleteChatSessions(id)
  } catch (e) {
    ElMessage.error('删除会话失败')
    return
  }

  const idx = chatHistory.value.findIndex(c => c.id === id)
  if (idx !== -1) chatHistory.value.splice(idx, 1)
  delete messagesMap.value[id]
  createNewChat()
}

function sendQuickPrompt(prompt) {
  inputMessage.value = prompt
  sendMessage()
}

function handleKeydown(e) {
  // Enter 发送，Shift + Enter 换行；中文等输入法组词期间不触发发送
  if (e.key === 'Enter' && !e.shiftKey && !e.isComposing && e.keyCode !== 229) {
    e.preventDefault()
    sendMessage()
  }
}

// 把指定会话移到 chatHistory 顶部（保持对象引用）
function moveChatToTop(id) {
  const idx = chatHistory.value.findIndex(c => c.id === id)
  if (idx > 0) {
    const [chat] = chatHistory.value.splice(idx, 1)
    chatHistory.value.unshift(chat)
  }
}

async function sendMessage() {
  const text = inputMessage.value.trim()
  if (!text || sending.value) return

  // 确保有当前会话（理论上 onMounted 已保证，但兜底）
  if (currentChatId.value == null) {
    createNewChat()
  }

  const chatIdAtSend = currentChatId.value
  const messages = messagesMap.value[chatIdAtSend] || (messagesMap.value[chatIdAtSend] = [])
  const userMsgId = ++messageIdCounter
  const assistantMsgId = ++messageIdCounter

  // 添加用户消息
  messages.push({
    id: userMsgId,
    role: 'user',
    content: text
  })

  // 添加助手消息（流式占位）
  messages.push({
    id: assistantMsgId,
    role: 'assistant',
    content: '',
    streaming: true
  })

  inputMessage.value = ''
  sending.value = true
  scrollToBottom()

  let agentSucceeded = false
  let replyText = ''

  try {
    // 构建历史消息（不含当前用户消息和助手占位消息）
    const history = messages
      .filter(m => !m.streaming && m.content)
      .slice(0, -1)
      .map(m => ({ role: m.role, content: m.content }))

    // 当前聊天对应的 Agent session_id（用于上游上下文 / 工具状态保持）
    const currentChat = chatHistory.value.find(c => c.id === chatIdAtSend)

    const result = await chatCompletions({
      message: text,
      history,
      ...(currentChat?.agentSessionId ? { session_id: currentChat.agentSessionId } : {})
    })

    // 保存外部 Agent 返回的 session_id
    if (currentChat && result?.session_id) {
      currentChat.agentSessionId = result.session_id
    }

    replyText = result?.reply || '（无回复内容）'
    const assistantMsg = messages.find(m => m.id === assistantMsgId)
    if (assistantMsg) {
      assistantMsg.content = replyText
      assistantMsg.streaming = false
    }
    agentSucceeded = true
    scrollToBottom()
  } catch (error) {
    const assistantMsg = messages.find(m => m.id === assistantMsgId)
    if (assistantMsg) {
      assistantMsg.content = '抱歉，请求出现了错误：' + (error?.message || '未知错误')
      assistantMsg.streaming = false
    }
    ElMessage.error('聊天请求失败')
  } finally {
    sending.value = false
  }

  // Agent 调用失败 → 不持久化，DB 保持干净
  if (!agentSucceeded) return

  // 持久化本轮对话
  try {
    const chat = chatHistory.value.find(c => c.id === chatIdAtSend)
    // 发送过程中用户可能已删除/清除该会话 → 跳过持久化，避免造成幽灵记录
    if (!chat) return
    const isLocal = !!chat.isLocal
    const turnRes = await saveChatTurn({
      sessionId: isLocal ? 0 : Number(chatIdAtSend),
      sessionType: 'control',
      agentSessionId: chat.agentSessionId || '',
      userMessage: text,
      assistantMessage: replyText
    })
    const data = turnRes?.data ?? turnRes

    // 二次确认：await 期间可能被删除
    const chatAfter = chatHistory.value.find(c => c.id === chatIdAtSend)
    if (!chatAfter || !data) return

    if (isLocal) {
      const newId = data.sessionId
      // 迁移 messagesMap key：local-xxx → 服务端 id
      messagesMap.value[newId] = messagesMap.value[chatIdAtSend]
      delete messagesMap.value[chatIdAtSend]
      if (currentChatId.value === chatIdAtSend) {
        currentChatId.value = newId
      }
      chatAfter.id = newId
      chatAfter.title = data.title || chatAfter.title
      chatAfter.lastMessageAt = data.lastMessageAt
      chatAfter.isLocal = false
    } else {
      chatAfter.title = data.title || chatAfter.title
      chatAfter.lastMessageAt = data.lastMessageAt
    }
    // 标题/时间更新后，将其移到列表顶部
    moveChatToTop(chatAfter.id)
  } catch (e) {
    // 不影响用户已经看到的回复，仅提示一次
    ElMessage.warning('对话保存失败，刷新后可能无法恢复本轮')
  }
}

// 仅在消息条数变化时滚到底；内容更新已在 sendMessage 内显式 scrollToBottom
watch(() => currentMessages.value.length, scrollToBottom)

onMounted(async () => {
  try {
    const res = await listChatSessions({ pageNum: 1, pageSize: 50, sessionType: 'control' })
    const payload = res?.data ?? res
    const rows = payload?.rows || []
    chatHistory.value = rows.map(s => ({
      id: s.id,
      title: s.title || '新对话',
      agentSessionId: s.agentSessionId || '',
      lastMessageAt: s.lastMessageAt || '',
      isLocal: false
    }))
    if (chatHistory.value.length === 0) {
      createNewChat()
    } else {
      const first = chatHistory.value[0]
      currentChatId.value = first.id
      await loadSessionMessages(first.id)
    }
  } catch (e) {
    ElMessage.error('加载会话列表失败')
    createNewChat()
  }
})
</script>

<style scoped lang="scss">
.chat-tab {
  display: flex;
  height: 100%;
  overflow: hidden;
}

.chat-sidebar {
  width: 240px;
  background: #F8FAFC;
  border-right: 1px solid #E2E8F0;
  display: flex;
  flex-direction: column;

  .sidebar-header {
    padding: 16px;
    border-bottom: 1px solid #E2E8F0;

    .el-button {
      width: 100%;
    }
  }

  .chat-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px;

    .chat-item {
      display: flex;
      align-items: center;
      padding: 10px 12px;
      border-radius: 8px;
      cursor: pointer;
      transition: all 0.2s;
      gap: 8px;
      margin-bottom: 4px;

      &:hover {
        background: #E2E8F0;

        .delete-icon {
          opacity: 1;
        }
      }

      &.active {
        background: #EBF5FF;
        color: #409EFF;
      }

      .chat-icon {
        font-size: 16px;
        color: #94A3B8;
      }

      .chat-title {
        flex: 1;
        font-size: 13px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .delete-icon {
        font-size: 14px;
        color: #94A3B8;
        opacity: 0;
        transition: opacity 0.2s;

        &:hover {
          color: #F56C6C;
        }
      }
    }
  }
}

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #FFFFFF;

  .message-list {
    flex: 1;
    overflow-y: auto;
    padding: 16px 20px;

    .welcome-area {
      text-align: center;
      padding: 60px 20px;

      .welcome-icon {
        margin-bottom: 16px;
      }

      .welcome-title {
        font-size: 20px;
        font-weight: 600;
        color: #1E293B;
        margin: 0 0 8px;
      }

      .welcome-desc {
        font-size: 14px;
        color: #94A3B8;
        margin: 0 0 24px;
      }

      .quick-prompts {
        display: flex;
        flex-wrap: wrap;
        justify-content: center;
        gap: 10px;

        .quick-tag {
          cursor: pointer;
          font-size: 13px;
          padding: 8px 14px;
          border-radius: 20px;
          transition: all 0.2s;

          &:hover {
            background: #EBF5FF;
            border-color: #409EFF;
            color: #409EFF;
          }
        }
      }
    }

    .message-item {
      display: flex;
      gap: 10px;
      margin-bottom: 14px;

      &.user {
        flex-direction: row-reverse;

        .message-content {
          align-items: flex-end;

          .message-bubble {
            padding: 6px 12px;
            background: #409EFF;
            color: #FFFFFF;
            border-radius: 12px 12px 2px 12px;
            line-height: 1.5;

            :deep(p) {
              margin: 0;
            }

            :deep(a) {
              color: #FFFFFF;
              text-decoration: underline;
            }

            :deep(code) {
              background: rgba(255, 255, 255, 0.2);
              color: #FFFFFF;
            }

            :deep(pre) {
              background: rgba(0, 0, 0, 0.2);
            }
          }
        }
      }

      &.assistant {
        .message-bubble {
          background: #F1F5F9;
          color: #1E293B;
          border-radius: 12px 12px 12px 2px;

          :deep(a) {
            color: #409EFF;
          }

          :deep(code) {
            background: #E2E8F0;
            padding: 2px 6px;
            border-radius: 4px;
            font-family: monospace;
            font-size: 13px;
          }

          :deep(pre) {
            background: #0F172A;
            color: #E2E8F0;
            padding: 12px;
            border-radius: 8px;
            overflow-x: auto;

            code {
              background: transparent;
              color: inherit;
              padding: 0;
            }
          }

          :deep(p) {
            margin: 0 0 8px;

            &:last-child {
              margin-bottom: 0;
            }
          }

          :deep(ul, ol) {
            margin: 0 0 8px;
            padding-left: 20px;
          }
        }
      }

      .message-content {
        display: flex;
        flex-direction: column;
        max-width: 78%;
        gap: 4px;

        .message-bubble {
          padding: 8px 12px;
          font-size: 14px;
          line-height: 1.6;
          word-break: break-word;
        }

        .thinking-bubble {
          display: inline-flex;
          align-items: center;
          gap: 6px;
          padding: 8px 14px;
          background: #F1F5F9;
          color: #64748B;
          border-radius: 12px 12px 12px 2px;
          font-size: 13px;

          .dot {
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: #94A3B8;
            animation: bounce 1.4s infinite ease-in-out;

            &:nth-child(2) {
              animation-delay: 0.2s;
            }

            &:nth-child(3) {
              animation-delay: 0.4s;
            }
          }

          .thinking-text {
            margin-left: 4px;
          }
        }
      }
    }
  }

  .chat-input-area {
    border-top: 1px solid #E2E8F0;
    padding: 10px 16px 12px;
    background: #FFFFFF;

    .input-box {
      display: flex;
      gap: 10px;
      align-items: flex-end;

      .el-textarea {
        flex: 1;

        :deep(.el-textarea__inner) {
          border-radius: 10px;
          padding: 8px 12px;
          font-size: 14px;
          line-height: 1.5;
          resize: none;
          min-height: 36px;
        }
      }

      .input-actions {
        display: flex;
        align-items: center;
        gap: 6px;
        flex-shrink: 0;
      }

      .clear-btn {
        height: 36px;
        width: 36px;
      }

      .send-btn {
        height: 36px;
        padding: 0 16px;
        border-radius: 10px;
      }
    }
  }
}

@keyframes bounce {

  0%,
  80%,
  100% {
    transform: translateY(0);
    opacity: 0.5;
  }

  40% {
    transform: translateY(-4px);
    opacity: 1;
  }
}
</style>
