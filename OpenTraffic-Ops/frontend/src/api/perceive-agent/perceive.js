import request from '@/utils/request'
import { getToken } from '@/utils/auth'

const AGENT_BASE = '/api/agent-perceive'

// ==================== 健康检查 ====================

/**
 * 健康检查
 * 返回 { status, service, api_version, metrics: { tools_count, macro_tasks_count, skills_count }, ... }
 */
export function healthCheck() {
  return request({
    url: `${AGENT_BASE}/api/health`,
    method: 'get'
  })
}

// ==================== SSE 流式对话 ====================

/**
 * 流式对话（OpenAI SSE 格式）
 *
 * 请求体：{ session_id?: string, query: string }
 *
 * 上游返回 text/event-stream，每行形如：
 *   data: {"choices":[{"delta":{"content":"片段"}}]}
 *   data: [DONE]
 *
 * @param {Object} data { session_id?, query }
 * @param {Object} handlers
 *   - onChunk(text): 每收到一个内容片段触发
 *   - onDone(): 流正常结束（收到 [DONE] 或连接被服务端关闭）
 *   - onError(err): 网络/解析错误
 *   - signal?: AbortSignal，外部取消用
 * @returns {Promise<void>} 解析完成后 resolve；失败时 reject（已经触发 onError）
 */
export function streamChat(data, { onChunk, onDone, onError, signal } = {}) {
  const token = getToken()
  const baseApi = import.meta.env.VITE_APP_BASE_API || ''
  const url = `${baseApi}${AGENT_BASE}/api/stream_chat`

  console.log('[streamChat] 请求URL:', url)
  console.log('[streamChat] 请求体:', data)

  return fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'text/event-stream',
      ...(token ? { 'Authorization': 'Bearer ' + token } : {})
    },
    body: JSON.stringify(data),
    signal
  }).then(async (response) => {
    console.log('[streamChat] 响应状态:', response.status, response.statusText)
    console.log('[streamChat] Content-Type:', response.headers.get('Content-Type'))

    if (!response.ok) {
      let errMsg = `请求失败 (${response.status})`
      try {
        const body = await response.json()
        console.log('[streamChat] 错误响应体:', body)
        if (body?.msg) errMsg = body.msg
      } catch (_) {
        // 忽略 JSON 解析失败
      }
      throw new Error(errMsg)
    }

    if (!response.body || !response.body.getReader) {
      throw new Error('当前环境不支持流式读取')
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder('utf-8')
    let buffer = ''
    let streamDone = false
    let totalContent = ''

    try {
      while (!streamDone) {
        const { value, done } = await reader.read()
        if (done) {
          streamDone = true
          break
        }
        const chunk = decoder.decode(value, { stream: true })
        buffer += chunk
        console.log('[streamChat] 收到原始数据块:', JSON.stringify(chunk))

        // SSE 事件以 \n\n 分隔；先按 \n 拆行，最后一段不完整的留在 buffer
        let newlineIdx
        while ((newlineIdx = buffer.indexOf('\n')) !== -1) {
          const line = buffer.slice(0, newlineIdx).replace(/\r$/, '')
          buffer = buffer.slice(newlineIdx + 1)
          if (!line) {
            // 空行表示一个 SSE 事件结束，此时若 buffer 里还有 data: 开头的残留需要处理
            continue
          }

          console.log('[streamChat] 解析行:', line)

          if (line.startsWith('data:')) {
            const dataStr = line.slice(5).trimStart()
            if (dataStr === '[DONE]') {
              console.log('[streamChat] 收到 [DONE]')
              streamDone = true
              break
            }
            try {
              const parsed = JSON.parse(dataStr)
              console.log('[streamChat] 解析JSON:', parsed)
              // 同时支持 OpenAI 标准格式和感知Agent自定义格式
              const content = parsed?.choices?.[0]?.delta?.content
                ?? parsed?.data?.text_answer
                ?? parsed?.text_answer
                ?? null
              if (content != null && content !== '') {
                totalContent += content
                console.log('[streamChat] 提取内容片段:', JSON.stringify(content), '累计:', totalContent.length)
                onChunk && onChunk(content)
              }
            } catch (e) {
              console.log('[streamChat] JSON解析失败:', dataStr, e?.message)
            }
          }
          // 其他事件（event:、id:、retry: 等）暂不处理
        }
      }

      // 处理 buffer 中最后可能残留的一行（没有换行符结尾）
      const remaining = buffer.trim()
      if (remaining && remaining.startsWith('data:')) {
        const dataStr = remaining.slice(5).trimStart()
        if (dataStr !== '[DONE]') {
          try {
            const parsed = JSON.parse(dataStr)
            const content = parsed?.choices?.[0]?.delta?.content
              ?? parsed?.data?.text_answer
              ?? parsed?.text_answer
              ?? null
            if (content != null && content !== '') {
              console.log('[streamChat] 提取最后内容片段:', JSON.stringify(content))
              onChunk && onChunk(content)
            }
          } catch (e) {
            console.log('[streamChat] 最后残留JSON解析失败:', dataStr)
          }
        }
      }

      console.log('[streamChat] 流结束，调用 onDone。累计字符数:', totalContent.length)
      onDone && onDone()
    } catch (err) {
      if (err?.name === 'AbortError') {
        console.log('[streamChat] 请求被取消')
        onDone && onDone()
        return
      }
      throw err
    }
  }).catch((err) => {
    if (err?.name === 'AbortError') return
    console.error('[streamChat] 错误:', err)
    onError && onError(err)
  })
}

// ==================== 聊天会话持久化（复用后端 /rtm/chatSession） ====================

const CHAT_BASE = '/rtm/chatSession'

/**
 * 当前用户的感知Agent聊天会话列表（分页）
 * @param {Object} params { pageNum, pageSize, sessionType }
 */
export function listChatSessions(params) {
  return request({
    url: `${CHAT_BASE}/list`,
    method: 'get',
    params
  })
}

/**
 * 获取会话详情（含全部消息）
 * @param {number} id
 */
export function getChatSessionDetail(id) {
  return request({
    url: `${CHAT_BASE}/${id}`,
    method: 'get'
  })
}

/**
 * 保存一轮对话（user + assistant），Agent 调用成功后再调用
 * @param {Object} data { sessionId?, sessionType, agentSessionId?, userMessage, assistantMessage }
 *   sessionId 为 0 或缺省时表示新建会话
 */
export function saveChatTurn(data) {
  return request({
    url: `${CHAT_BASE}/turn`,
    method: 'post',
    data,
    headers: { repeatSubmit: false }
  })
}

/**
 * 删除会话（支持批量，逗号分隔）
 * @param {number|string} ids
 */
export function deleteChatSessions(ids) {
  return request({
    url: `${CHAT_BASE}/${ids}`,
    method: 'delete'
  })
}
