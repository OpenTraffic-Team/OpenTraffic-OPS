import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Server, DeployRecord, CreateServerRequest, DeployRequest, ServerServiceStatus } from '@/types'
import { serverApi } from '@/api/server'
import { deployApi } from '@/api/deploy'

export const useServerStore = defineStore('server', () => {
  const servers = ref<Server[]>([])
  const deployRecords = ref<DeployRecord[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchServers() {
    loading.value = true
    error.value = null
    try {
      servers.value = await serverApi.list()
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取服务器列表失败'
    } finally {
      loading.value = false
    }
  }

  async function createServer(data: CreateServerRequest) {
    loading.value = true
    error.value = null
    try {
      const server = await serverApi.create(data)
      servers.value.push(server)
      return server
    } catch (err) {
      error.value = err instanceof Error ? err.message : '创建服务器失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateServer(id: string, data: Partial<CreateServerRequest>) {
    loading.value = true
    error.value = null
    try {
      const server = await serverApi.update(id, data)
      const index = servers.value.findIndex(s => s.id === id)
      if (index !== -1) {
        servers.value[index] = server
      }
      return server
    } catch (err) {
      error.value = err instanceof Error ? err.message : '更新服务器失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteServer(id: string) {
    loading.value = true
    error.value = null
    try {
      await serverApi.delete(id)
      servers.value = servers.value.filter(s => s.id !== id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '删除服务器失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function testConnection(id: string) {
    loading.value = true
    error.value = null
    try {
      const result = await serverApi.testConnection(id)
      return result
    } catch (err) {
      error.value = err instanceof Error ? err.message : '连接测试失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deploy(data: DeployRequest) {
    loading.value = true
    error.value = null
    try {
      const record = await deployApi.deploy(data)
      return record
    } catch (err) {
      error.value = err instanceof Error ? err.message : '部署失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function undeploy(serverId: string, binaryName: string) {
    loading.value = true
    error.value = null
    try {
      const res = await deployApi.undeploy({ server_id: serverId, binary_name: binaryName })
      return res
    } catch (err) {
      error.value = err instanceof Error ? err.message : '卸载失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchDeployRecords(serverId?: string) {
    loading.value = true
    error.value = null
    try {
      deployRecords.value = await deployApi.listRecords(serverId)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取部署记录失败'
    } finally {
      loading.value = false
    }
  }

  async function getProxyConfig(id: string) {
    loading.value = true
    error.value = null
    try {
      const res = await serverApi.getProxyConfig(id)
      return res.content
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取Agent配置失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateProxyConfig(id: string, content: string) {
    loading.value = true
    error.value = null
    try {
      await serverApi.updateProxyConfig(id, content)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '更新Agent配置失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function getSoftwareConfig(id: string, software: string) {
    loading.value = true
    error.value = null
    try {
      const res = await serverApi.getSoftwareConfig(id, software)
      return res.content
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取软件配置失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function getDefaultSoftwareConfig(software: string) {
    loading.value = true
    error.value = null
    try {
      const res = await serverApi.getDefaultSoftwareConfig(software)
      return res.content
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取默认配置失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateSoftwareConfig(id: string, software: string, content: string) {
    loading.value = true
    error.value = null
    try {
      await serverApi.updateSoftwareConfig(id, software, content)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '更新软件配置失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function getServiceStatus(id: string, software: string): Promise<ServerServiceStatus> {
    try {
      const res = await serverApi.getServiceStatus(id, software)
      return {
        software,
        status: res.status as 'running' | 'stopped' | 'unknown',
        label: res.status === 'running' ? '运行中' : res.status === 'stopped' ? '已停止' : '未知'
      }
    } catch (err) {
      return {
        software,
        status: 'unknown',
        label: '未知'
      }
    }
  }

  async function startService(id: string, software: string) {
    loading.value = true
    error.value = null
    try {
      await serverApi.startService(id, software)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '启动服务失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function stopService(id: string, software: string) {
    loading.value = true
    error.value = null
    try {
      await serverApi.stopService(id, software)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '停止服务失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function restartService(id: string, software: string) {
    loading.value = true
    error.value = null
    try {
      await serverApi.restartService(id, software)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '重启服务失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    servers,
    deployRecords,
    loading,
    error,
    fetchServers,
    createServer,
    updateServer,
    deleteServer,
    testConnection,
    deploy,
    undeploy,
    fetchDeployRecords,
    getProxyConfig,
    updateProxyConfig,
    getSoftwareConfig,
    getDefaultSoftwareConfig,
    updateSoftwareConfig,
    getServiceStatus,
    startService,
    stopService,
    restartService
  }
})
