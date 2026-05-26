import request from '@/utils/request'

// 查询主机监控列表（当前最新数据）
export function listHostMon(query) {
  return request({
    url: '/rtm/hostMon/list',
    method: 'get',
    params: query
  })
}

// 查询主机历史监控数据
export function hostHistory(query) {
  return request({
    url: '/rtm/hostMon/hostHistory',
    method: 'get',
    params: query
  })
}
