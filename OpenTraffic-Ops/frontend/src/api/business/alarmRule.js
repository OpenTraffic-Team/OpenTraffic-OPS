import request from '@/utils/request'

// 查询告警规则列表
export function listAlarmRule(query) {
  return request({
    url: '/rtm/alarmRule/list',
    method: 'get',
    params: query
  })
}

// 查询告警规则详细
export function getAlarmRule(id) {
  return request({
    url: '/rtm/alarmRule/' + id,
    method: 'get'
  })
}

// 新增告警规则
export function addAlarmRule(data) {
  return request({
    url: '/rtm/alarmRule',
    method: 'post',
    data: data
  })
}

// 修改告警规则
export function updateAlarmRule(data) {
  return request({
    url: '/rtm/alarmRule',
    method: 'put',
    data: data
  })
}

// 删除告警规则
export function delAlarmRule(id) {
  return request({
    url: '/rtm/alarmRule/' + id,
    method: 'delete'
  })
}

// 更新告警规则状态
export function updateAlarmRuleStatus(data) {
  return request({
    url: '/rtm/alarmRule/status',
    method: 'put',
    data: data
  })
}
