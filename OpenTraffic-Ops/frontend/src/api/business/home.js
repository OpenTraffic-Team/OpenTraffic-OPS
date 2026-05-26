import request from '@/utils/request'

// 获取首页统计数据
export function getStatisticData() {
  return request({
    url: '/rtm/home/getStatisticData',
    method: 'get'
  })
}

// 获取异常进程列表
export function getMonProcessAbnormalList() {
  return request({
    url: '/rtm/home/getMonProcessAbnormalList',
    method: 'get'
  })
}
