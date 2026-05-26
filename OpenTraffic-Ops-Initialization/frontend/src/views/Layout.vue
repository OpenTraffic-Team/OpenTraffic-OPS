<template>
  <div class="layout-container">
    <!-- 动态背景 -->
    <div class="layout-bg">
      <div class="bg-orb orb-1"></div>
      <div class="bg-orb orb-2"></div>
    </div>

    <el-container class="outer-container">
      <!-- 侧边栏 -->
      <el-aside width="220px">
        <div class="logo">
          <div class="logo-icon">
            <svg viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg">
              <rect x="8" y="8" width="20" height="20" rx="4" fill="#6366f1"/>
              <rect x="36" y="8" width="20" height="20" rx="4" fill="#8b5cf6"/>
              <rect x="8" y="36" width="20" height="20" rx="4" fill="#8b5cf6"/>
              <rect x="36" y="36" width="20" height="20" rx="4" fill="#a78bfa"/>
            </svg>
          </div>
          <h2>OpenTraffic Ops 部署面板</h2>
        </div>
        <el-menu
          class="side-menu"
          :default-active="activeMenu"
          router
        >
          <el-menu-item index="/dashboard">
            <el-icon><Monitor /></el-icon>
            <span>首页</span>
          </el-menu-item>
          <el-menu-item index="/servers">
            <el-icon><Link /></el-icon>
            <span>服务器管理</span>
          </el-menu-item>
          <el-menu-item index="/components">
            <el-icon><Box /></el-icon>
            <span>组件管理</span>
          </el-menu-item>
          <el-menu-item index="/configs">
            <el-icon><Setting /></el-icon>
            <span>配置管理</span>
          </el-menu-item>
          <el-menu-item index="/help">
            <el-icon><Document /></el-icon>
            <span>使用指南</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主内容区 -->
      <el-container class="inner-container">
        <!-- 顶部栏 -->
        <el-header>
          <div class="header-content">
            <div class="breadcrumb">
              <el-breadcrumb separator="/">
                <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
                <el-breadcrumb-item v-if="currentRoute.meta.title">
                  {{ currentRoute.meta.title }}
                </el-breadcrumb-item>
              </el-breadcrumb>
            </div>
            <div class="user-info">
              <el-dropdown @command="handleCommand">
                <span class="el-dropdown-link">
                  <div class="user-avatar">
                    <el-icon><User /></el-icon>
                  </div>
                  <span class="user-name">{{ userStore.user?.username || 'User' }}</span>
                  <el-icon class="el-icon--right"><arrow-down /></el-icon>
                </span>
                <template #dropdown>
                  <el-dropdown-menu class="light-dropdown">
                    <el-dropdown-item command="logout">
                      <el-icon><SwitchButton /></el-icon>
                      <span>退出登录</span>
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </el-header>

        <!-- 页面内容 -->
        <el-main>
          <div class="page-content">
            <router-view />
          </div>
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const activeMenu = computed(() => route.path)
const currentRoute = computed(() => route)

const handleCommand = (command: string) => {
  if (command === 'logout') {
    ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      userStore.logout()
      ElMessage.success('已退出登录')
      router.push('/login')
    }).catch(() => {})
  }
}
</script>

<style scoped>
.layout-container {
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  position: relative;
  background: #f5f7fa;
}

/* ========== 动态背景 ========== */
.layout-bg {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
}

.bg-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0.08;
}

.orb-1 {
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, #6366f1 0%, transparent 70%);
  top: -10%;
  right: 10%;
  animation: orbFloat 25s ease-in-out infinite;
}

.orb-2 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, #8b5cf6 0%, transparent 70%);
  bottom: -5%;
  left: 20%;
  animation: orbFloat 20s ease-in-out infinite reverse;
}

@keyframes orbFloat {
  0%, 100% { transform: translate(0, 0) scale(1); }
  50% { transform: translate(30px, -20px) scale(1.1); }
}

.outer-container {
  height: 100%;
  position: relative;
  z-index: 1;
}

/* ========== 侧边栏 ========== */
.el-aside {
  background: #ffffff;
  border-right: 1px solid #e5e7eb;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.logo {
  height: 70px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 20px;
  border-bottom: 1px solid #f3f4f6;
  flex-shrink: 0;
}

.logo-icon {
  width: 36px;
  height: 36px;
}

.logo-icon svg {
  width: 100%;
  height: 100%;
}

.logo h2 {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  color: #1f2937;
  letter-spacing: 0.5px;
}

.side-menu {
  flex: 1;
  border-right: none;
  overflow-y: auto;
  background: transparent;
  padding: 12px 10px;
}

:deep(.side-menu .el-menu-item) {
  height: 48px;
  line-height: 48px;
  border-radius: 10px;
  margin-bottom: 4px;
  color: #6b7280;
  transition: all 0.3s ease;
}

:deep(.side-menu .el-menu-item:hover) {
  background: #f3f4f6;
  color: #374151;
}

:deep(.side-menu .el-menu-item.is-active) {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.1) 0%, rgba(139, 92, 246, 0.08) 100%);
  color: #6366f1;
  font-weight: 500;
}

:deep(.side-menu .el-menu-item.is-active::before) {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 20px;
  background: linear-gradient(180deg, #6366f1, #8b5cf6);
  border-radius: 0 3px 3px 0;
}

:deep(.side-menu .el-menu-item .el-icon) {
  color: inherit;
  font-size: 18px;
  margin-right: 10px;
}

:deep(.side-menu .el-menu-item span) {
  font-size: 14px;
}

/* ========== 顶部栏 ========== */
.el-header {
  background: #ffffff;
  border-bottom: 1px solid #e5e7eb;
  display: flex;
  align-items: center;
  padding: 0 24px;
  flex-shrink: 0;
  height: 60px;
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* 面包屑 */
:deep(.el-breadcrumb__item) {
  font-size: 14px;
}

:deep(.el-breadcrumb__inner) {
  color: #9ca3af !important;
  font-weight: 400;
}

:deep(.el-breadcrumb__inner.is-link:hover) {
  color: #6366f1 !important;
}

:deep(.el-breadcrumb__separator) {
  color: #d1d5db !important;
}

:deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
  color: #374151 !important;
}

/* 用户信息 */
.user-info {
  cursor: pointer;
}

.el-dropdown-link {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: 10px;
  transition: all 0.3s ease;
}

.el-dropdown-link:hover {
  background: #f3f4f6;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 16px;
}

.user-name {
  color: #374151;
  font-size: 14px;
  font-weight: 500;
}

:deep(.el-icon--right) {
  color: #9ca3af;
  font-size: 12px;
}

/* ========== 主内容区 ========== */
.el-main {
  flex: 1;
  overflow: hidden;
  padding: 0;
  background: transparent;
}

.page-content {
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* ========== 浅色下拉菜单 ========== */
:deep(.light-dropdown) {
  background: #ffffff !important;
  border: 1px solid #e5e7eb !important;
  border-radius: 12px !important;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1) !important;
  padding: 6px !important;
}

:deep(.light-dropdown .el-dropdown-menu__item) {
  color: #374151;
  border-radius: 8px;
  padding: 10px 14px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

:deep(.light-dropdown .el-dropdown-menu__item:hover) {
  background: #f3f4f6 !important;
  color: #1f2937 !important;
}

:deep(.light-dropdown .el-dropdown-menu__item .el-icon) {
  font-size: 16px;
}

:deep(.light-dropdown .el-dropdown-menu__item--divided) {
  border-top-color: #f3f4f6 !important;
}

:deep(.light-dropdown .el-dropdown-menu__item--divided::before) {
  background: transparent !important;
}

:deep(.el-popper__arrow::before) {
  background: #ffffff !important;
  border-color: #e5e7eb !important;
}
</style>
