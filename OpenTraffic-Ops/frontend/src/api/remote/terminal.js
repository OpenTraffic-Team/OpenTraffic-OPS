import { getToken } from '@/utils/auth'

/**
 * 构建终端WebSocket URL
 * @param {string} hostIp 目标主机IP
 * @returns {string} WebSocket URL
 */
export function buildTerminalWSUrl(hostIp) {
  const token = getToken()
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'

  // 开发环境：VITE_APP_WS_BASE_API 是路径前缀（如 /dev-ws-api），走 Vite 代理
  const wsBaseApi = import.meta.env.VITE_APP_WS_BASE_API || ''
  if (wsBaseApi && wsBaseApi.startsWith('/')) {
    const baseUrl = `${protocol}//${window.location.host}${wsBaseApi}`
    return `${baseUrl}/ws/terminal?hostIp=${encodeURIComponent(hostIp)}&token=${encodeURIComponent(token || '')}`
  }

  // 生产环境：VITE_APP_WS_BASE_API 未设置或为空，直接使用当前域名
  const baseUrl = `${protocol}//${window.location.host}`
  return `${baseUrl}/ws/terminal?hostIp=${encodeURIComponent(hostIp)}&token=${encodeURIComponent(token || '')}`
}
