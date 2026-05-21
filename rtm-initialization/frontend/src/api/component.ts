import request from './index'
import type { Component, ComponentConfig, ComponentStats, ComponentCatalogItem } from '@/types'

export interface InstallComponentRequest {
  name: string
  type: string
  image?: string
  version?: string
  config: ComponentConfig
  image_source?: 'pull' | 'upload' | 'embedded'
  image_file?: File
}

export const componentApi = {
  // 获取组件列表
  list() {
    return request.get<any, Component[]>('/components')
  },

  // 获取组件目录
  getCatalog() {
    return request.get<any, ComponentCatalogItem[]>('/components/catalog')
  },

  // 获取组件详情
  get(id: string) {
    return request.get<any, Component>(`/components/${id}`)
  },

  // 安装组件
  install(data: InstallComponentRequest) {
    const formData = new FormData()
    formData.append('name', data.name)
    formData.append('type', data.type)
    if (data.image) {
      formData.append('image', data.image)
    }
    if (data.version) {
      formData.append('version', data.version)
    }
    formData.append('config', JSON.stringify(data.config))
    if (data.image_source) {
      formData.append('image_source', data.image_source)
    }
    if (data.image_file) {
      formData.append('image_file', data.image_file)
    }
    return request.post<any, Component>('/components', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },

  // 卸载组件
  uninstall(id: string) {
    return request.delete(`/components/${id}`)
  },

  // 启动组件
  start(id: string) {
    return request.post(`/components/${id}/start`)
  },

  // 停止组件
  stop(id: string) {
    return request.post(`/components/${id}/stop`)
  },

  // 重启组件
  restart(id: string) {
    return request.post(`/components/${id}/restart`)
  },

  // 获取日志
  getLogs(id: string, tail: string = '100') {
    return request.get<any, { logs: string }>(`/components/${id}/logs`, { params: { tail } })
  },

  // 获取统计信息
  getStats(id: string) {
    return request.get<any, ComponentStats>(`/components/${id}/stats`)
  },

  // 更新配置
  updateConfig(id: string, config: ComponentConfig) {
    return request.put(`/components/${id}/config`, config)
  },

  // 控制组件（通用方法）
  control(id: string, action: 'start' | 'stop' | 'restart') {
    return request.post(`/components/${id}/${action}`)
  }
}
