<template>
  <div class="app-container">
    <div class="terminal-header">
      <el-page-header @back="goBack" :content="`WebSSH - ${hostIp}`" />
      <el-tag :type="wsStatus === 'connected' ? 'success' : wsStatus === 'connecting' ? 'warning' : 'danger'"
        size="small">
        {{ wsStatusText }}
      </el-tag>
    </div>
    <div class="terminal-body">
      <div class="terminal-wrapper" v-loading="loading">
        <div ref="terminalRef" class="terminal-container" @click="handleContainerClick"></div>
      </div>
    </div>
  </div>
</template>

<script>
/**
 * RemoteTerminal 全局共享状态
 * 用于防止多个组件实例同时创建 WebSocket 连接时弹出多次提示
 */
const TerminalGlobal = {
  connectGen: 0,      // 全局连接世代，所有实例共享
  debounceTimer: null // 全局防抖定时器，所有实例共享
}
</script>

<script setup name="RemoteTerminal">
import { ref, onMounted, onBeforeUnmount, onActivated, onDeactivated, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'
import { buildTerminalWSUrl } from '@/api/remote/terminal'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const terminalRef = ref(null)
const loading = ref(true)

const hostIp = ref(route.query.ip || '')
const wsStatus = ref('connecting')
const wsStatusText = ref('连接中...')

let term = null
let fitAddon = null
let ws = null
let reconnectTimer = null
let pingTimer = null
let connectTimeout = null
let isActive = true  // 组件是否处于活跃状态

/** 返回 */
function goBack() {
  router.push('/business/remote-ops')
}

/** 初始化终端 */
function initTerminal() {
  try {
    if (!terminalRef.value) {
      console.error('[Terminal] terminalRef is null, DOM not ready')
      return false
    }

    term = new Terminal({
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      cursorBlink: true,
      cursorStyle: 'block',
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
        cursor: '#d4d4d4',
        selectionBackground: '#264f78',
        black: '#000000',
        red: '#cd3131',
        green: '#0dbc79',
        yellow: '#e5e510',
        blue: '#2472c8',
        magenta: '#bc3fbc',
        cyan: '#11a8cd',
        white: '#e5e5e5',
      },
      rows: 30,
      cols: 120,
      scrollback: 10000,
    })

    fitAddon = new FitAddon()
    term.loadAddon(fitAddon)
    term.open(terminalRef.value)

    // 确保容器有尺寸后再 fit
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        try {
          if (fitAddon && terminalRef.value) {
            fitAddon.fit()
            console.log('[Terminal] terminal size:', term.cols, 'x', term.rows)
          }
        } catch (e) {
          console.warn('[Terminal] fit error:', e)
        }
      })
    })

    // 终端输入事件
    term.onData((data) => {
      if (term) term.scrollToBottom()
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'input',
          data: data,
          sessionId: hostIp.value
        }))
      }
    })

    // 终端大小调整
    term.onResize(({ cols, rows }) => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'resize',
          cols: cols,
          rows: rows,
          sessionId: hostIp.value
        }))
      }
    })

    loading.value = false
    console.log('[Terminal] initTerminal success')
    if (term) {
      term.write(`\x1B[32mWebSSH Terminal - ${hostIp.value}\x1B[0m\r\n`)
      term.write('\x1B[33m正在连接...\x1B[0m\r\n\r\n')
      term.focus()
      console.log('[Terminal] term focused')
    }
    return true
  } catch (e) {
    console.error('[Terminal] initTerminal failed:', e)
    loading.value = false
    ElMessage.error('终端初始化失败: ' + e.message)
    return false
  }
}

/** 连接WebSocket（全局防抖，所有实例共享） */
function connectWS() {
  console.log('[Terminal] connectWS called, hostIp=', hostIp.value, 'ws readyState=', ws ? ws.readyState : 'null')

  if (!hostIp.value) {
    wsStatus.value = 'error'
    wsStatusText.value = '未指定IP'
    ElMessage.error('未指定主机IP')
    return
  }

  // 全局防抖：清除所有实例的延迟连接请求，300ms内只执行最后一次
  if (TerminalGlobal.debounceTimer) {
    clearTimeout(TerminalGlobal.debounceTimer)
    TerminalGlobal.debounceTimer = null
  }

  TerminalGlobal.debounceTimer = setTimeout(() => {
    TerminalGlobal.debounceTimer = null
    doConnectWS()
  }, 300)
}

/** 真正执行WebSocket连接 */
function doConnectWS() {
  console.log('[Terminal] doConnectWS called, hostIp=', hostIp.value, 'ws readyState=', ws ? ws.readyState : 'null')

  // 如果已经连接中或已连接，不再重复创建
  if (ws && (ws.readyState === WebSocket.CONNECTING || ws.readyState === WebSocket.OPEN)) {
    console.log('[Terminal] ws already connecting/open, skip')
    return
  }

  // 清理旧连接，递增全局世代使旧连接回调失效
  cleanup()

  // 分配新的全局世代号
  TerminalGlobal.connectGen++
  const myGen = TerminalGlobal.connectGen
  console.log('[Terminal] new global generation:', myGen)

  wsStatus.value = 'connecting'
  wsStatusText.value = '连接中...'

  let wsUrl
  try {
    wsUrl = buildTerminalWSUrl(hostIp.value)
    console.log('[Terminal] WebSocket URL:', wsUrl)
  } catch (e) {
    console.error('[Terminal] buildTerminalWSUrl failed:', e)
    wsStatus.value = 'error'
    wsStatusText.value = 'URL错误'
    return
  }

  try {
    ws = new WebSocket(wsUrl)
    console.log('[Terminal] new WebSocket created, readyState=', ws.readyState)
  } catch (e) {
    console.error('[Terminal] new WebSocket failed:', e)
    wsStatus.value = 'error'
    wsStatusText.value = '创建失败'
    return
  }

  // 连接超时检测（5秒）
  connectTimeout = setTimeout(() => {
    if (ws && ws.readyState === WebSocket.CONNECTING) {
      console.error('[Terminal] connection timeout')
      wsStatus.value = 'error'
      wsStatusText.value = '连接超时'
      if (term) term.write('\r\n[连接超时，请检查后端服务是否运行]\r\n')
      ws.close()
    }
  }, 5000)

  ws.onopen = () => {
    console.log('[Terminal] ws.onopen, gen=', myGen)
    // 全局世代检查：只有最新世代的连接才能执行关键操作
    if (myGen !== TerminalGlobal.connectGen) {
      console.log('[Terminal] ws.onopen ignored, gen mismatch, myGen=', myGen, 'current=', TerminalGlobal.connectGen)
      return
    }
    if (connectTimeout) {
      clearTimeout(connectTimeout)
      connectTimeout = null
    }
    wsStatus.value = 'connected'
    wsStatusText.value = '已连接'
    ElMessage.success('终端连接成功')

    if (term) {
      term.write('\x1B[2K\r')
      term.write('\x1B[32m连接成功\x1B[0m\r\n\r\n')
      term.focus()
    }

    try {
      const dims = fitAddon ? fitAddon.proposeDimensions() : null
      if (dims) {
        ws.send(JSON.stringify({
          type: 'resize',
          cols: dims.cols,
          rows: dims.rows,
          sessionId: hostIp.value
        }))
      }
    } catch (e) {
      console.warn('[Terminal] send initial resize error:', e)
    }

    startPing()
  }

  ws.onmessage = (event) => {
    // 全局世代检查：只处理最新世代的消息
    if (myGen !== TerminalGlobal.connectGen) {
      return
    }
    try {
      const msg = JSON.parse(event.data)
      if (msg.type === 'output' && msg.data) {
        if (term) {
          term.write(msg.data)
          term.scrollToBottom()
        }
      } else if (msg.type === 'error') {
        if (term) {
          term.write(`\r\n[ERROR] ${msg.error || '未知错误'}\r\n`)
          term.scrollToBottom()
        }
      }
    } catch (e) {
      if (term) {
        term.write(event.data)
        term.scrollToBottom()
      }
    }
  }

  ws.onclose = (event) => {
    console.log('[Terminal] ws.onclose, gen=', myGen, 'code=', event.code, 'reason=', event.reason)
    if (connectTimeout) {
      clearTimeout(connectTimeout)
      connectTimeout = null
    }
    // 全局世代检查：只有最新世代的断开才更新状态并触发重连
    if (myGen !== TerminalGlobal.connectGen) {
      console.log('[Terminal] ws.onclose ignored, gen mismatch, myGen=', myGen, 'current=', TerminalGlobal.connectGen)
      return
    }
    wsStatus.value = 'disconnected'
    wsStatusText.value = '已断开'
    stopPing()
    if (term) {
      term.write('\r\n\r\n[连接已关闭]\r\n')
    }
    // 自动重连（先清理旧的定时器，防止重复）
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    reconnectTimer = setTimeout(() => {
      console.log('[Terminal] reconnecting...')
      connectWS()
    }, 5000)
  }

  ws.onerror = (err) => {
    console.error('[Terminal] ws.onerror, gen=', myGen, ':', err)
    if (connectTimeout) {
      clearTimeout(connectTimeout)
      connectTimeout = null
    }
    // 全局世代检查：只有最新世代的错误才更新状态
    if (myGen !== TerminalGlobal.connectGen) {
      console.log('[Terminal] ws.onerror ignored, gen mismatch, myGen=', myGen, 'current=', TerminalGlobal.connectGen)
      return
    }
    wsStatus.value = 'error'
    wsStatusText.value = '连接错误'
    if (term) term.write('\r\n[WebSocket 连接错误]\r\n')
  }
}

/** 启动ping */
function startPing() {
  stopPing()
  pingTimer = setInterval(() => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'ping' }))
    }
  }, 30000)
}

/** 停止ping */
function stopPing() {
  if (pingTimer) {
    clearInterval(pingTimer)
    pingTimer = null
  }
}

/** 窗口大小调整 */
function handleResize() {
  if (fitAddon) {
    try {
      fitAddon.fit()
    } catch (e) {
      console.warn('[Terminal] resize fit error:', e)
    }
  }
}

/** 点击终端容器获取焦点 */
function handleContainerClick() {
  if (term) {
    term.focus()
  }
}

/** 清理资源 */
function cleanup() {
  // 递增全局世代，使所有旧实例的旧连接回调失效
  TerminalGlobal.connectGen++
  console.log('[Terminal] cleanup, new global generation:', TerminalGlobal.connectGen)

  // 清除全局防抖定时器（阻止任何实例的延迟连接）
  if (TerminalGlobal.debounceTimer) {
    clearTimeout(TerminalGlobal.debounceTimer)
    TerminalGlobal.debounceTimer = null
  }
  stopPing()
  if (connectTimeout) {
    clearTimeout(connectTimeout)
    connectTimeout = null
  }
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  if (ws) {
    ws.onopen = null
    ws.onmessage = null
    ws.onclose = null
    ws.onerror = null
    ws.close()
    ws = null
  }
}

/** 设置并连接 */
function setup() {
  console.log('[Terminal] setup, hostIp=', hostIp.value, 'query=', route.query)
  if (!term) {
    initTerminal()
  }
  connectWS()
}

onMounted(() => {
  isActive = true
  setup()
  window.addEventListener('resize', handleResize)
})

// keep-alive 缓存时清理资源，防止后台累积重连定时器
onDeactivated(() => {
  console.log('[Terminal] onDeactivated, cleaning up')
  isActive = false
  cleanup()
})

// keep-alive 缓存后重新激活
onActivated(() => {
  console.log('[Terminal] onActivated, hostIp=', hostIp.value, 'query=', route.query)
  isActive = true
  // 更新IP（如果从不同主机跳转过来）
  const newIp = route.query.ip || ''
  if (newIp !== hostIp.value) {
    hostIp.value = newIp
    cleanup()
    // 终端已存在，直接连接
    connectWS()
  } else {
    // 同一个主机，且已经连接中/已连接，不做任何操作
    if (!ws || ws.readyState === WebSocket.CLOSED || ws.readyState === WebSocket.CLOSING) {
      connectWS()
    }
  }
})

// 监听路由 query 参数变化（同一组件实例内导航）
watch(() => route.query.ip, (newIp) => {
  console.log('[Terminal] watch route.query.ip changed:', newIp)
  // 组件不活跃时不响应（防止 keep-alive 缓存期间触发）
  if (!isActive) {
    console.log('[Terminal] watch ignored, component not active')
    return
  }
  if (newIp && newIp !== hostIp.value) {
    hostIp.value = newIp
    cleanup()
    if (term) {
      term.clear()
      term.reset()
    }
    connectWS()
  }
})

onBeforeUnmount(() => {
  isActive = false
  window.removeEventListener('resize', handleResize)
  cleanup()
  if (term) {
    term.dispose()
    term = null
  }
})
</script>

<style scoped lang="scss">
.app-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 84px);
  padding: 24px;
}

.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  padding: 16px 24px;
  border-radius: 12px;
  margin-bottom: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 1px 2px rgba(0, 0, 0, 0.03);
  border: 1px solid #E2E8F0;
}

.terminal-body {
  flex: 1;
  overflow: hidden;
}

.terminal-wrapper {
  width: 100%;
  height: 100%;
  background: #1e1e1e;
  border-radius: 12px;
  padding: 25px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1);
}

.terminal-container {
  width: 100%;
  height: 100%;
}

:deep(.xterm) {
  height: 100%;
}

:deep(.xterm-viewport) {
  overflow-y: auto !important;
}
</style>
