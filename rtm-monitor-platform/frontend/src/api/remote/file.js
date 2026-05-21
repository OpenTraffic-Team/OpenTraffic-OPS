import request from '@/utils/request'

// 获取文件列表
export function listFiles(data, config = {}) {
  return request({
    url: '/api/remote/file/list',
    method: 'post',
    data: data,
    headers: { repeatSubmit: false },
    ...config
  })
}

// 读取文件
export function readFile(data, config = {}) {
  return request({
    url: '/api/remote/file/read',
    method: 'post',
    data: data,
    headers: { repeatSubmit: false },
    ...config
  })
}

// 写入文件
export function writeFile(data, config = {}) {
  return request({
    url: '/api/remote/file/write',
    method: 'post',
    data: data,
    ...config
  })
}

// 删除文件
export function deleteFile(data, config = {}) {
  return request({
    url: '/api/remote/file/delete',
    method: 'post',
    data: data,
    ...config
  })
}

// 上传文件
export function uploadFile(data, config = {}) {
  return request({
    url: '/api/remote/file/upload',
    method: 'post',
    data: data,
    ...config
  })
}

// 下载文件
export function downloadFile(data, config = {}) {
  return request({
    url: '/api/remote/file/download',
    method: 'post',
    data: data,
    headers: { repeatSubmit: false },
    ...config
  })
}

// 创建目录
export function mkdir(data, config = {}) {
  return request({
    url: '/api/remote/file/mkdir',
    method: 'post',
    data: data,
    ...config
  })
}
