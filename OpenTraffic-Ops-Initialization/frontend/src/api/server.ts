import request from './index'
import type { Server, CreateServerRequest } from '@/types'

export const serverApi = {
  // 获取服务器列表
  list() {
    return request.get<any, Server[]>('/servers')
  },

  // 获取服务器详情
  get(id: string) {
    return request.get<any, Server>(`/servers/${id}`)
  },

  // 创建服务器
  create(data: CreateServerRequest) {
    return request.post<any, Server>('/servers', data)
  },

  // 更新服务器
  update(id: string, data: Partial<CreateServerRequest>) {
    return request.put<any, Server>(`/servers/${id}`, data)
  },

  // 删除服务器
  delete(id: string) {
    return request.delete(`/servers/${id}`)
  },

  // 测试SSH连接
  testConnection(id: string) {
    return request.post<any, { message: string }>(`/servers/${id}/test`)
  },

  // 获取opentraffic-ops-proxy配置（兼容旧接口）
  getProxyConfig(id: string) {
    return request.get<any, { content: string }>(`/servers/${id}/proxy-config`)
  },

  // 更新opentraffic-ops-proxy配置（兼容旧接口）
  updateProxyConfig(id: string, content: string) {
    return request.put(`/servers/${id}/proxy-config`, { content })
  },

  // 获取指定软件的配置
  getSoftwareConfig(id: string, software: string) {
    return request.get<any, { content: string }>(`/servers/${id}/configs/${software}`)
  },

  // 获取指定软件的默认配置（嵌入资源）
  getDefaultSoftwareConfig(software: string) {
    return request.get<any, { content: string }>(`/servers/configs/${software}/default`)
  },

  // 更新指定软件的配置
  updateSoftwareConfig(id: string, software: string, content: string) {
    return request.put(`/servers/${id}/configs/${software}`, { content })
  },

  // 获取指定服务的运行状态
  getServiceStatus(id: string, software: string) {
    return request.get<any, { status: string }>(`/servers/${id}/services/${software}/status`)
  },

  // 启动指定服务
  startService(id: string, software: string) {
    return request.post<any, { message: string }>(`/servers/${id}/services/${software}/start`)
  },

  // 停止指定服务
  stopService(id: string, software: string) {
    return request.post<any, { message: string }>(`/servers/${id}/services/${software}/stop`)
  },

  // 重启指定服务
  restartService(id: string, software: string) {
    return request.post<any, { message: string }>(`/servers/${id}/services/${software}/restart`)
  }
}
