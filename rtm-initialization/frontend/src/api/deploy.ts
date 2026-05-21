import request from './index'
import type { DeployRequest, DeployRecord } from '@/types'

export const deployApi = {
  // 部署二进制文件
  deploy(data: DeployRequest) {
    return request.post<any, DeployRecord>('/deploy', data)
  },

  // 卸载二进制文件
  undeploy(data: { server_id: string; binary_name: string }) {
    return request.post<any, { message: string }>('/deploy/undeploy', data)
  },

  // 获取部署记录列表
  listRecords(serverId?: string) {
    return request.get<any, DeployRecord[]>('/deploy/records', {
      params: serverId ? { server_id: serverId } : {}
    })
  },

  // 获取部署记录详情
  getRecord(id: number) {
    return request.get<any, DeployRecord>(`/deploy/records/${id}`)
  }
}
