import router, { constantRoutes, asyncRoutes } from '@/router'

const usePermissionStore = defineStore(
  'permission',
  {
    state: () => ({
      routes: [],
      addRoutes: [],
      defaultRoutes: [],
      topbarRouters: [],
      sidebarRouters: []
    }),
    actions: {
      setRoutes(routes) {
        this.addRoutes = routes
        this.routes = constantRoutes.concat(routes)
      },
      setDefaultRoutes(routes) {
        this.defaultRoutes = constantRoutes.concat(routes)
      },
      setTopbarRoutes(routes) {
        this.topbarRouters = routes
      },
      setSidebarRouters(routes) {
        this.sidebarRouters = routes
      },
      generateRoutes(roles) {
        return new Promise((resolve) => {
          // 使用静态路由，不再请求后端菜单接口
          const accessedRoutes = asyncRoutes

          // 将业务路由注册到 router
          accessedRoutes.forEach(route => {
            router.addRoute(route)
          })

          // 设置侧边栏路由：constantRoutes（首页+个人中心） + 业务路由
          this.setRoutes(accessedRoutes)
          this.setSidebarRouters(constantRoutes.concat(accessedRoutes))
          this.setDefaultRoutes(accessedRoutes)
          this.setTopbarRoutes(accessedRoutes)

          resolve(accessedRoutes)
        })
      }
    }
  })

export default usePermissionStore
