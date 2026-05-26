import { createWebHistory, createRouter } from 'vue-router'
/* Layout */
import Layout from '@/layout'

/**
 * Note: 路由配置项
 *
 * hidden: true                     // 当设置 true 的时候该路由不会再侧边栏出现
 * alwaysShow: true                 // 不管子路由个数，始终显示根路由
 * redirect: noRedirect             // 面包屑中不可点击
 * name:'router-name'               // 设定路由的名字，用于 keep-alive
 * meta : {
    noCache: true                   // 不会被 <keep-alive> 缓存
    title: 'title'                  // 侧边栏和面包屑中展示的名字
    icon: 'svg-name'                // 菜单图标
    activeMenu: '/system/user'      // 高亮对应侧边栏
  }
 */

// 公共路由（无需登录即可访问的基础页面）
export const constantRoutes = [
  {
    path: '/redirect',
    component: Layout,
    hidden: true,
    children: [
      {
        path: '/redirect/:path(.*)',
        component: () => import('@/views/redirect/index.vue')
      }
    ]
  },
  {
    path: '/login',
    component: () => import('@/views/login'),
    hidden: true
  },
  {
    path: '/:pathMatch(.*)*',
    component: () => import('@/views/error/404'),
    hidden: true
  },
  {
    path: '/401',
    component: () => import('@/views/error/401'),
    hidden: true
  },
  {
    path: '',
    component: Layout,
    redirect: '/index',
    children: [
      {
        path: '/index',
        component: () => import('@/views/index.vue'),
        name: 'Index',
        meta: { title: '首页', icon: 'dashboard', affix: true }
      }
    ]
  },
  {
    path: '/user',
    component: Layout,
    hidden: true,
    redirect: 'noredirect',
    children: [
      {
        path: 'profile',
        component: () => import('@/views/system/user/profile/index'),
        name: 'Profile',
        meta: { title: '个人中心', icon: 'user' }
      }
    ]
  }
]

// 静态业务路由（登录后动态添加，不再依赖后端菜单接口）
export const asyncRoutes = [
  {
    path: '/business',
    component: Layout,
    redirect: 'noRedirect',
    alwaysShow: true,
    name: 'HostManage',
    meta: { title: '主机管理', icon: 'server' },
    children: [
      {
        path: 'host-info',
        component: () => import('@/views/business/host-info/index'),
        name: 'HostInfoList',
        meta: { title: '主机信息', icon: 'server' }
      },
      {
        path: 'remote-ops',
        component: () => import('@/views/business/remote-ops/index'),
        name: 'RemoteOps',
        meta: { title: '主机运维', icon: 'tool' }
      }
    ]
  },
  {
    path: '/business',
    component: Layout,
    hidden: true,
    children: [
      {
        path: 'host-health',
        component: () => import('@/views/business/host-health/index'),
        name: 'HostHealth',
        meta: { title: '主机健康度', activeMenu: '/business/host-info' }
      },
      {
        path: 'remote-terminal',
        component: () => import('@/views/business/remote-terminal/index'),
        name: 'RemoteTerminal',
        meta: { title: '远程终端', activeMenu: '/business/remote-ops' }
      },
      {
        path: 'remote-file',
        component: () => import('@/views/business/remote-file/index'),
        name: 'RemoteFile',
        meta: { title: '文件管理', activeMenu: '/business/remote-ops' }
      }
    ]
  },
  {
    path: '/agent',
    component: Layout,
    redirect: 'noRedirect',
    alwaysShow: true,
    name: 'AgentManage',
    meta: { title: 'Agent管理', icon: 'monitor' },
    children: [
      {
        path: 'control',
        component: () => import('@/views/business/agent-control/index'),
        name: 'AgentControl',
        meta: { title: '控制Agent', icon: 'setting' }
      },
      {
        path: 'perceive',
        component: () => import('@/views/business/agent-perceive/index'),
        name: 'AgentPerceive',
        meta: { title: '感知Agent', icon: 'eye-open' }
      }
    ]
  },
  {
    path: '/alarm',
    component: Layout,
    redirect: 'noRedirect',
    alwaysShow: true,
    name: 'AlarmManage',
    meta: { title: '告警管理', icon: 'email' },
    children: [
      {
        path: 'config',
        component: () => import('@/views/business/alarm-config/index'),
        name: 'AlarmConfig',
        meta: { title: '告警通道', icon: 'email' }
      },
      {
        path: 'threshold',
        component: () => import('@/views/business/threshold-config/index'),
        name: 'ThresholdConfig',
        meta: { title: '告警规则', icon: 'slider' }
      },
      {
        path: 'record',
        component: () => import('@/views/business/alarm-record/index'),
        name: 'AlarmRecord',
        meta: { title: '告警记录', icon: 'bell' }
      }
    ]
  },
  {
    path: '/system',
    component: Layout,
    redirect: 'noRedirect',
    alwaysShow: true,
    name: 'System',
    meta: { title: '系统管理', icon: 'system' },
    children: [
      {
        path: 'user',
        component: () => import('@/views/system/user/index'),
        name: 'User',
        meta: { title: '用户管理', icon: 'user' }
      },
      {
        path: 'login-log',
        component: () => import('@/views/monitor/loginlog/index'),
        name: 'LoginLog',
        meta: { title: '登录日志', icon: 'log' }
      },
      {
        path: 'oper-log',
        component: () => import('@/views/monitor/operlog/index'),
        name: 'OperLog',
        meta: { title: '操作日志', icon: 'log' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes: constantRoutes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0 }
    }
  }
})

export default router
