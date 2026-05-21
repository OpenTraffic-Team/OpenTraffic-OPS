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
            <el-icon :size="48" color="#8B5CF6"><View /></el-icon>
          </div>
          <h3 class="welcome-title">交通感知智能助手</h3>
          <p class="welcome-desc">我可以帮您分析路口流量、检测拥堵状况、监控感知设备运行状态</p>
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
            <el-avatar v-else :size="36" :icon="View"
              :style="{ backgroundColor: '#8B5CF6' }" />
          </div>
          <div class="message-content">
            <div v-if="!msg.content && msg.streaming" class="thinking-bubble">
              <span class="dot"></span>
              <span class="dot"></span>
              <span class="dot"></span>
              <span class="thinking-text">正在思考中…</span>
            </div>
            <div v-else class="message-bubble" v-html="renderContent(msg)"></div>
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

<script setup name="PerceiveChatTab">
import { ref, computed, nextTick, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus, ChatLineRound, Close, UserFilled,
  View, Delete, Promotion
} from '@element-plus/icons-vue'
import {
  streamChat,
  listChatSessions,
  getChatSessionDetail,
  saveChatTurn,
  deleteChatSessions
} from '@/api/perceive-agent/perceive'
import MarkdownIt from 'markdown-it'

const SESSION_TYPE = 'perceive'

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true
})

// 缓存：仅缓存非流式（已完成）消息的渲染结果，避免历史消息重复 markdown 解析
// 流式消息每次更新 content 都会变，直接走 md.render 不入缓存，防止缓存膨胀
const renderCache = new Map()
function renderContent(msg) {
  const text = msg?.content
  if (!text) return ''
  if (msg.streaming) {
    return md.render(text)
  }
  const cached = renderCache.get(text)
  if (cached !== undefined) return cached
  const html = md.render(text)
  renderCache.set(text, html)
  return html
}

// 快速提示语（感知相关）
const quickPrompts = [
  '帮我看看路口1拥堵吗？',
  '分析一下当前交通流量',
  '检测一下感知设备运行状态',
  '最近 1 小时有哪些异常事件？'
]

// 对话历史：每项 { id, title, agentSessionId, lastMessageAt, isLocal? }
// id 为服务端会话主键（number）或本地占位 'local-<ts>'（isLocal: true）
const chatHistory = ref([])
const currentChatId = ref(null)

// 消息存储 { chatId: [messages] }
const messagesMap = ref({})
const loadingSessions = new Map()

const currentMessages = computed(() => {
  if (currentChatId.value == null) return []
  return messagesMap.value[currentChatId.value] || []
})

const inputMessage = ref('')
const sending = ref(false)
const messageContainer = ref(null)

let messageIdCounter = 0
// 活跃 AbortController，用于组件卸载或切会话时主动关闭 SSE 连接
let activeAbortController = null

function scrollToBottom() {
  nextTick(() => {
    if (messageContainer.value) {
      messageContainer.value.scrollTop = messageContainer.value.scrollHeight
    }
  })
}

// 生成客户端 session_id：优先 crypto.randomUUID，回退到时间戳+随机数
function generateSessionId() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return 'sess-' + Date.now() + '-' + Math.random().toString(36).slice(2, 10)
}

function createNewChat() {
  const id = 'local-' + Date.now()
  chatHistory.value.unshift({
    id,
    title: '新对话',
    // 新会话即刻分配 agentSessionId，使同一会话的多轮对话能在上游保持上下文
    agentSessionId: generateSessionId(),
    lastMessageAt: '',
    isLocal: true
  })
  messagesMap.value[id] = []
  currentChatId.value = id
}

async function loadSessionMessages(id) {
  const existing = messagesMap.value[id]
  if (Array.isArray(existing) && existing.length > 0) return
  if (loadingSessions.has(id)) return loadingSessions.get(id)

  const task = (async () => {
    try {
      const res = await getChatSessionDetail(id)
      const detail = res?.data ?? res
      const current = messagesMap.value[id]
      if (Array.isArray(current) && current.length > 0) return
      const msgs = (detail?.messages || []).map(m => ({
        id: ++messageIdCounter,
        role: m.role,
        content: m.content
      }))
      messagesMap.value[id] = msgs
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
  if (e.key === 'Enter' && !e.shiftKey && !e.isComposing && e.keyCode !== 229) {
    e.preventDefault()
    sendMessage()
  }
}

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

  const currentChat = chatHistory.value.find(c => c.id === chatIdAtSend)
  // 保险：若历史会话从后端加载时没有 agentSessionId，临时生成一个用于本次及后续
  if (currentChat && !currentChat.agentSessionId) {
    currentChat.agentSessionId = generateSessionId()
  }
  const sessionIdForUpstream = currentChat?.agentSessionId || generateSessionId()

  let receivedAnyChunk = false
  // 查找消息辅助：会话被删除后 messagesMap[chatIdAtSend] 会变 undefined
  const findAssistant = () => {
    const arr = messagesMap.value[chatIdAtSend]
    if (!Array.isArray(arr)) return null
    return arr.find(m => m.id === assistantMsgId) || null
  }

  // 创建 AbortController 支持取消
  const controller = new AbortController()
  activeAbortController = controller

  let streamError = null
  try {
    await streamChat(
      { session_id: sessionIdForUpstream, query: text },
      {
        onChunk: (chunk) => {
          const assistantMsg = findAssistant()
          if (!assistantMsg) return
          assistantMsg.content += chunk
          receivedAnyChunk = true
          // 频繁的 chunk 不必每次都滚到底，但用户体验上滚动更自然
          scrollToBottom()
        },
        onDone: () => {
          const assistantMsg = findAssistant()
          if (assistantMsg) {
            // 上游可能返回空内容，给个兜底提示
            if (!receivedAnyChunk && !assistantMsg.content) {
              assistantMsg.content = '（无回复内容）'
            }
            assistantMsg.streaming = false
          }
        },
        onError: (err) => {
          streamError = err
          const assistantMsg = findAssistant()
          if (assistantMsg) {
            assistantMsg.content = '抱歉，请求出现了错误：' + (err?.message || '未知错误')
            assistantMsg.streaming = false
          }
        },
        signal: controller.signal
      }
    )
  } catch (err) {
    // 错误已通过 onError 处理，这里只防止 unhandled rejection
    if (!streamError) {
      streamError = err
      const assistantMsg = findAssistant()
      if (assistantMsg) {
        assistantMsg.content = '抱歉，请求出现了错误：' + (err?.message || '未知错误')
        assistantMsg.streaming = false
      }
    }
  } finally {
    if (activeAbortController === controller) {
      activeAbortController = null
    }
    sending.value = false
    scrollToBottom()
  }

  if (streamError) {
    ElMessage.error('聊天请求失败')
    return
  }

  // 持久化本轮对话
  const assistantFinal = findAssistant()
  const replyText = assistantFinal?.content || ''
  if (!replyText) return

  try {
    const chat = chatHistory.value.find(c => c.id === chatIdAtSend)
    if (!chat) return // 发送过程中可能被删除
    const isLocal = !!chat.isLocal
    const turnRes = await saveChatTurn({
      sessionId: isLocal ? 0 : Number(chatIdAtSend),
      sessionType: SESSION_TYPE,
      agentSessionId: chat.agentSessionId || '',
      userMessage: text,
      assistantMessage: replyText
    })
    const data = turnRes?.data ?? turnRes

    const chatAfter = chatHistory.value.find(c => c.id === chatIdAtSend)
    if (!chatAfter || !data) return

    if (isLocal) {
      const newId = data.sessionId
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
    moveChatToTop(chatAfter.id)
  } catch (e) {
    ElMessage.warning('对话保存失败，刷新后可能无法恢复本轮')
  }
}

watch(() => currentMessages.value.length, scrollToBottom)

onMounted(async () => {
  try {
    const res = await listChatSessions({
      pageNum: 1,
      pageSize: 50,
      sessionType: SESSION_TYPE
    })
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

onUnmounted(() => {
  // 组件卸载时中断正在进行的 SSE 连接，避免后台继续写入孤儿对象
  if (activeAbortController) {
    try { activeAbortController.abort() } catch (_) { /* ignore */ }
    activeAbortController = null
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
        background: #F5F3FF;
        color: #8B5CF6;
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
            background: #F5F3FF;
            border-color: #8B5CF6;
            color: #8B5CF6;
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
          background: #F5F3FF;
          color: #1E293B;
          border-radius: 12px 12px 12px 2px;

          :deep(a) {
            color: #8B5CF6;
          }

          :deep(code) {
            background: #E9D5FF;
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
          background: #F5F3FF;
          color: #64748B;
          border-radius: 12px 12px 12px 2px;
          font-size: 13px;

          .dot {
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: #8B5CF6;
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
