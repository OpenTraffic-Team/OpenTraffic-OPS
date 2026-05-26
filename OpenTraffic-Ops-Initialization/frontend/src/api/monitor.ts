import request from './index'
import type { Overview, ComponentDetail } from '@/types'

export const monitorApi = {
  // 获取总览信息
  getOverview() {
    return request.get<any, Overview>('/monitor/overview')
  },

  // 获取所有组件详情
  getComponentDetails() {
    return request.get<any, ComponentDetail[]>('/monitor/components')
  }
}
