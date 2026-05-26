import request from '@/utils/request'

// 查询告警通道列表
export function listAlarmChannel(query) {
  return request({
    url: '/rtm/alarmChannel/list',
    method: 'get',
    params: query
  })
}

// 查询所有告警通道（下拉选择）
export function listAllAlarmChannel() {
  return request({
    url: '/rtm/alarmChannel/all',
    method: 'get'
  })
}

// 查询告警通道详细
export function getAlarmChannel(id) {
  return request({
    url: '/rtm/alarmChannel/' + id,
    method: 'get'
  })
}

// 新增告警通道
export function addAlarmChannel(data) {
  return request({
    url: '/rtm/alarmChannel',
    method: 'post',
    data: data
  })
}

// 修改告警通道
export function updateAlarmChannel(data) {
  return request({
    url: '/rtm/alarmChannel',
    method: 'put',
    data: data
  })
}

// 删除告警通道
export function delAlarmChannel(id) {
  return request({
    url: '/rtm/alarmChannel/' + id,
    method: 'delete'
  })
}

// 更新告警通道状态
export function updateAlarmChannelStatus(data) {
  return request({
    url: '/rtm/alarmChannel/status',
    method: 'put',
    data: data
  })
}

// 设置默认通道
export function setAlarmChannelDefault(data) {
  return request({
    url: '/rtm/alarmChannel/setDefault',
    method: 'put',
    data: data
  })
}
