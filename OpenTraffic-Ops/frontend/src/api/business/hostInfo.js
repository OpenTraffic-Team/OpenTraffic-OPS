import request from '@/utils/request'

// 查询主机信息列表
export function listHostInfo(query) {
  return request({
    url: '/rtm/hostInfo/list',
    method: 'get',
    params: query
  })
}

// 查询主机信息详细
export function getHostInfo(id) {
  return request({
    url: '/rtm/hostInfo/' + id,
    method: 'get'
  })
}

// 修改主机信息（仅支持修改名称）
export function updateHostInfo(data) {
  return request({
    url: '/rtm/hostInfo',
    method: 'put',
    data: data
  })
}

// 删除主机信息
export function delHostInfo(id) {
  return request({
    url: '/rtm/hostInfo/' + id,
    method: 'delete'
  })
}

// 查询主机信息树列表
export function selectHostInfoTreeNode() {
  return request({
    url: '/rtm/hostInfo/selectHostInfoTreeNode',
    method: 'get',
  })
}
