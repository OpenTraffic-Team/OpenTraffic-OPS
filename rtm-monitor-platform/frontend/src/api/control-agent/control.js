import request from '@/utils/request'

const AGENT_BASE = '/api/agent-control'

// ==================== 算法 ====================

/**
 * 参数分析
 * @param {Object} data
 */
export function analyzeParams(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/algorithm/analyze-params`,
    method: 'post',
    data
  })
}

/**
 * 决策算法
 * @param {Object} data
 */
export function v1Decision(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/algorithm/v1-decision`,
    method: 'post',
    data
  })
}

/**
 * 配置对比
 * @param {Object} data
 */
export function configDiff(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/config/diff`,
    method: 'post',
    data
  })
}

/**
 * 固定时间配置
 * @param {Object} data
 */
export function configFixedTime(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/config/fixed-time`,
    method: 'post',
    data
  })
}

/**
 * V1配置
 * @param {Object} data
 */
export function configV1(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/config/v1`,
    method: 'post',
    data
  })
}

/**
 * 建模解析
 * @param {Object} data
 */
export function modelingParse(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/modeling/parse`,
    method: 'post',
    data
  })
}

/**
 * 路网建模
 * @param {Object} data
 */
export function modelingRoadnet(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/modeling/roadnet`,
    method: 'post',
    data
  })
}

/**
 * 模型验证
 * @param {Object} data
 */
export function modelingValidate(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/modeling/validate`,
    method: 'post',
    data
  })
}

// ==================== 仿真 ====================

/**
 * 运行仿真
 * @param {Object} data
 */
export function runSimulation(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/simulation/run`,
    method: 'post',
    data
  })
}

/**
 * 流量仿真
 * @param {Object} data
 */
export function flowSimulation(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/simulation/flow`,
    method: 'post',
    data
  })
}

/**
 * 仿真报告
 * @param {Object} data
 */
export function reportSimulation(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/simulation/report`,
    method: 'post',
    data
  })
}

// ==================== 监控运维 ====================

/**
 * 异常检测
 * @param {Object} data
 */
export function monitoringAnomaly(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/monitoring/anomaly`,
    method: 'post',
    data
  })
}

/**
 * 有效性评估
 * @param {Object} data
 */
export function monitoringEffectiveness(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/monitoring/effectiveness`,
    method: 'post',
    data
  })
}

/**
 * 缺失值插补
 * @param {Object} data
 */
export function monitoringImpute(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/monitoring/impute`,
    method: 'post',
    data
  })
}

/**
 * 恢复监控
 * @param {Object} data
 */
export function monitoringRecovery(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/monitoring/recovery`,
    method: 'post',
    data
  })
}

/**
 * 降级监控
 * @param {Object} data
 */
export function monitoringDegrade(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/monitoring/degrade`,
    method: 'post',
    data
  })
}

/**
 * 日志分析
 * @param {Object} data
 */
export function opsAnalyzeLog(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/ops/analyze-log`,
    method: 'post',
    data
  })
}

/**
 * 调优建议
 * @param {Object} data
 */
export function opsSuggestTuning(data) {
  return request({
    url: `${AGENT_BASE}/api/agent/ops/suggest-tuning`,
    method: 'post',
    data
  })
}

// ==================== 聊天 ====================

/**
 * 获取聊天配置
 */
export function getChatConfig() {
  return request({
    url: `${AGENT_BASE}/api/chat/config`,
    method: 'get'
  })
}

/**
 * 简单聊天（非流式）
 * @param {Object} data
 */
export function chatSimple(data) {
  return request({
    url: `${AGENT_BASE}/api/chat/simple`,
    method: 'post',
    data
  })
}

/**
 * 聊天对话 completions
 * 后端 ChatRequest 结构：{ message, module?, context?, history?, session_id? }
 * 后端 ChatResponse 结构：{ reply, tool_calls_made, tool_call_details, module_used, success, session_id, latency_ms }
 * 接口为同步 JSON 响应（非 SSE 流式），LLM 工具调用耗时较长，单独放宽超时到 120s
 * @param {Object} data { message: string, history?: Array<{role,content}>, module?, context?, session_id? }
 */
export function chatCompletions(data) {
  return request({
    url: `${AGENT_BASE}/api/chat/completions`,
    method: 'post',
    data,
    timeout: 120000,
    headers: { repeatSubmit: false }
  })
}

// ==================== 健康检查 ====================

/**
 * 健康检查
 */
export function healthCheck() {
  return request({
    url: `${AGENT_BASE}/healthz`,
    method: 'get'
  })
}

// ==================== 聊天会话持久化 ====================

const CHAT_BASE = '/rtm/chatSession'

/**
 * 当前用户的聊天会话列表（分页）
 * @param {Object} params { pageNum, pageSize }
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
 * @param {Object} data { sessionId?, agentSessionId?, userMessage, assistantMessage }
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
 * 重命名会话
 * @param {Object} data { id, title }
 */
export function renameChatSession(data) {
  return request({
    url: CHAT_BASE,
    method: 'put',
    data
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
