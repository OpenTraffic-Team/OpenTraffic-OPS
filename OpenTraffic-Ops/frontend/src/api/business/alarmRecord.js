import request from '@/utils/request'

// 查询告警记录列表
export function listAlarmRecord(query) {
  return request({
    url: '/rtm/alarmRecord/list',
    method: 'get',
    params: query
  })
}

// 查询告警记录详细
export function getAlarmRecord(id) {
  return request({
    url: '/rtm/alarmRecord/' + id,
    method: 'get'
  })
}

// 确认告警
export function ackAlarmRecord(id) {
  return request({
    url: '/rtm/alarmRecord/ack/' + id,
    method: 'put'
  })
}

// 批量确认告警
export function batchAckAlarmRecord(data) {
  return request({
    url: '/rtm/alarmRecord/batchAck',
    method: 'put',
    data: data
  })
}

// 查询未读告警数量
export function getUnreadCount() {
  return request({
    url: '/rtm/alarmRecord/unreadCount',
    method: 'get'
  })
}
