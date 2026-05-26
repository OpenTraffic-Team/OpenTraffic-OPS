// 告警相关字典与工具函数

// 指标/服务类型映射
export const metricTypeMap = {
  'cpu': 'CPU使用率',
  'mem': '内存使用率',
  'disk': '磁盘使用率',
  'network': '网络流量',
  'load': '系统负载',
  'host_offline': '主机离线',
  'agent_offline': 'Agent离线'
}

// 获取指标/服务类型显示名
export function metricTypeLabel(type) {
  return metricTypeMap[type] || type || '-'
}
