import request from './index'
import type { User, LoginRequest, LoginResponse } from '@/types'

export const authApi = {
  // 登录
  login(data: LoginRequest) {
    return request.post<any, LoginResponse>('/auth/login', data)
  },

  // 登出
  logout() {
    return request.post('/auth/logout')
  },

  // 获取用户信息
  getProfile() {
    return request.get<any, User>('/users/profile')
  }
}
