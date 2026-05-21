<template>
  <div class="login-page">
    <!-- 动态渐变背景 -->
    <div class="gradient-bg">
      <div class="gradient-orb orb-1"></div>
      <div class="gradient-orb orb-2"></div>
      <div class="gradient-orb orb-3"></div>
    </div>

    <!-- 网格装饰 -->
    <div class="grid-overlay"></div>

    <!-- 浮动粒子 -->
    <div class="particles">
      <div v-for="i in 20" :key="i" class="particle" :style="getParticleStyle(i)"></div>
    </div>

    <!-- 登录卡片 -->
    <div class="login-card" :class="{ 'card-enter': mounted }">
      <!-- 左侧装饰区 -->
      <div class="login-card-left">
        <div class="brand-section">
          <div class="brand-logo">
            <svg viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg">
              <rect x="8" y="8" width="20" height="20" rx="4" fill="#6366f1" />
              <rect x="36" y="8" width="20" height="20" rx="4" fill="#8b5cf6" />
              <rect x="8" y="36" width="20" height="20" rx="4" fill="#8b5cf6" />
              <rect x="36" y="36" width="20" height="20" rx="4" fill="#a78bfa" />
            </svg>
          </div>
          <h2 class="brand-title">RTM 初始化平台</h2>
          <p class="brand-desc">基础设施组件自动化安装与管理</p>
        </div>
        <div class="feature-list">
          <div class="feature-item">
            <el-icon class="feature-icon">
              <Connection />
            </el-icon>
            <span>自动化部署</span>
          </div>
          <div class="feature-item">
            <el-icon class="feature-icon">
              <Monitor />
            </el-icon>
            <span>实时监控</span>
          </div>
          <div class="feature-item">
            <el-icon class="feature-icon">
              <Setting />
            </el-icon>
            <span>配置管理</span>
          </div>
        </div>
      </div>

      <!-- 右侧表单区 -->
      <div class="login-card-right">
        <div class="form-header">
          <h3 class="form-title">欢迎回来</h3>
          <p class="form-subtitle">请登录您的账号以继续</p>
        </div>

        <el-form ref="formRef" :model="form" :rules="rules" class="login-form" @submit.prevent="handleLogin">
          <el-form-item prop="username">
            <div class="input-label">账号</div>
            <div class="input-wrapper">
              <el-icon class="input-icon">
                <User />
              </el-icon>
              <el-input style="width: 375px" v-model="form.username" placeholder="请输入用户名" size="large" class="custom-input"
                :class="{ 'is-focus': focusField === 'username' }" @focus="focusField = 'username'"
                @blur="focusField = ''" />
            </div>
          </el-form-item>

          <el-form-item prop="password">
            <div class="input-label">密码</div>
            <div class="input-wrapper">
              <el-icon class="input-icon">
                <Lock />
              </el-icon>
              <el-input style="width: 375px" v-model="form.password" type="password" placeholder="请输入密码" size="large" class="custom-input"
                :class="{ 'is-focus': focusField === 'password' }" show-password @focus="focusField = 'password'"
                @blur="focusField = ''" @keyup.enter="handleLogin" />
            </div>
          </el-form-item>

          <el-form-item class="submit-item">
            <button type="button" class="login-btn" :class="{ 'is-loading': loading }" :disabled="loading"
              @click="handleLogin">
              <span class="btn-text">{{ loading ? '登录中' : '登 录' }}</span>
            </button>
          </el-form-item>
        </el-form>
      </div>
    </div>

    <!-- 底部版权 -->
    <div class="login-footer">
      <p> RTM Initialization Platform </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { authApi } from '@/api/auth'
import type { LoginRequest } from '@/types'

const router = useRouter()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const mounted = ref(false)
const focusField = ref('')

const form = reactive<LoginRequest>({
  username: 'admin',
  password: 'admin123'
})

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

const getParticleStyle = (_index: number) => {
  const size = Math.random() * 4 + 2
  const left = Math.random() * 100
  const delay = Math.random() * 15
  const duration = Math.random() * 10 + 10
  return {
    width: `${size}px`,
    height: `${size}px`,
    left: `${left}%`,
    animationDelay: `${delay}s`,
    animationDuration: `${duration}s`
  }
}

const handleLogin = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      const response = await authApi.login(form)
      userStore.setToken(response.token)
      userStore.setUser(response.user)
      ElMessage.success('登录成功')
      router.push('/')
    } catch (error) {
      ElMessage.error('登录失败，请检查用户名和密码')
    } finally {
      loading.value = false
    }
  })
}

onMounted(() => {
  setTimeout(() => {
    mounted.value = true
  }, 100)
})
</script>

<style scoped>
.login-page {
  position: relative;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
}

/* ========== 动态渐变背景 ========== */
.gradient-bg {
  position: absolute;
  inset: 0;
  overflow: hidden;
}

.gradient-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.3;
}

.orb-1 {
  width: 600px;
  height: 600px;
  background: radial-gradient(circle, #6366f1 0%, transparent 70%);
  top: -10%;
  left: -5%;
  animation-delay: 0s;
}

.orb-2 {
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, #8b5cf6 0%, transparent 70%);
  bottom: -10%;
  right: -5%;
  animation-delay: -7s;
  animation-duration: 25s;
}

.orb-3 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, #06b6d4 0%, transparent 70%);
  top: 40%;
  left: 50%;
  animation-delay: -14s;
  animation-duration: 22s;
}

@keyframes orbFloat {

  0%,
  100% {
    transform: translate(0, 0) scale(1);
  }

  25% {
    transform: translate(50px, -30px) scale(1.1);
  }

  50% {
    transform: translate(-30px, 50px) scale(0.95);
  }

  75% {
    transform: translate(30px, 30px) scale(1.05);
  }
}

/* ========== 网格装饰 ========== */
.grid-overlay {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(0, 0, 0, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 0, 0, 0.03) 1px, transparent 1px);
  background-size: 60px 60px;
  mask-image: radial-gradient(ellipse at center, black 30%, transparent 70%);
  -webkit-mask-image: radial-gradient(ellipse at center, black 30%, transparent 70%);
}

/* ========== 浮动粒子 ========== */
.particles {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.particle {
  position: absolute;
  bottom: -10px;
  background: rgba(99, 102, 241, 0.3);
  border-radius: 50%;
  animation: particleFloat linear infinite;
}

@keyframes particleFloat {
  0% {
    transform: translateY(0) scale(0);
    opacity: 0;
  }

  10% {
    opacity: 1;
    transform: translateY(-10vh) scale(1);
  }

  90% {
    opacity: 1;
  }

  100% {
    transform: translateY(-110vh) scale(0.5);
    opacity: 0;
  }
}

/* ========== 登录卡片 ========== */
.login-card {
  position: relative;
  z-index: 10;
  display: flex;
  width: 900px;
  min-height: 520px;
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 24px;
  overflow: hidden;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.1);
  opacity: 0;
  transform: translateY(30px) scale(0.98);
  transition: all 0.8s cubic-bezier(0.16, 1, 0.3, 1);
}

.login-card.card-enter {
  opacity: 1;
  transform: translateY(0) scale(1);
}

.input-label {
  font-size: 13px;
  font-weight: 500;
  color: #475569;
  margin-bottom: 10px;
  margin-left: 15px;
  margin-right: 10px;
}

/* 卡片发光边框效果 */
.login-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 24px;
  padding: 1.5px;
  background: linear-gradient(135deg,
      rgba(99, 102, 241, 0.3) 0%,
      transparent 30%,
      transparent 70%,
      rgba(139, 92, 246, 0.3) 100%);
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  pointer-events: none;
}

/* ========== 左侧装饰区 ========== */
.login-card-left {
  flex: 0 0 360px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 48px 36px;
  background: linear-gradient(135deg,
      rgba(99, 102, 241, 0.08) 0%,
      rgba(139, 92, 246, 0.05) 100%);
  border-right: 1px solid #f3f4f6;
  position: relative;
  overflow: hidden;
}

.login-card-left::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle at 30% 70%,
      rgba(99, 102, 241, 0.06) 0%,
      transparent 50%);
  pointer-events: none;
}

.brand-section {
  position: relative;
  z-index: 1;
}

.brand-logo {
  width: 56px;
  height: 56px;
  margin-bottom: 24px;
}

.brand-logo svg {
  width: 100%;
  height: 100%;
  filter: drop-shadow(0 0 20px rgba(99, 102, 241, 0.3));
}

.brand-title {
  font-size: 22px;
  font-weight: 700;
  color: #0f172a;
  margin-bottom: 8px;
  letter-spacing: -0.3px;
}

.brand-desc {
  font-size: 13px;
  color: #64748b;
  line-height: 1.5;
  margin: 0;
}

.feature-list {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #6b7280;
  font-size: 13px;
  transition: all 0.3s ease;
}

.feature-item:hover {
  color: #374151;
  transform: translateX(4px);
}

.feature-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f1f5f9;
  border-radius: 10px;
  transition: all 0.3s ease;
}

.feature-item:hover .feature-icon {
  background: rgba(99, 102, 241, 0.1);
  color: #6366f1;
}

/* ========== 右侧表单区 ========== */
.login-card-right {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 48px 44px;
}

.form-header {
  margin-bottom: 32px;
  font-size: 16px;
  color: #2563EB;
  border-bottom: 1px solid #E2E8F0;
  margin: 8px 0 24px 0;
  padding-bottom: 12px;
  font-weight: 600;
}

.form-title {
  font-size: 26px;
  font-weight: 700;
  color: #0f172a;
  margin-bottom: 6px;
  letter-spacing: -0.3px;
}

.form-subtitle {
  font-size: 14px;
  color: #94a3b8;
  margin: 0;
}

/* ========== 自定义输入框 ========== */
.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 16px;
  z-index: 1;
  font-size: 18px;
  color: #9ca3af;
  transition: color 0.3s ease;
  pointer-events: none;
}

:deep(.custom-input .el-input__wrapper) {
  background: #f9fafb !important;
  border: 1px solid #e5e7eb !important;
  border-radius: 12px !important;
  box-shadow: none !important;
  padding-left: 44px !important;
  padding-right: 16px !important;
  height: 48px;
  transition: all 0.3s ease;
}

:deep(.custom-input .el-input__inner) {
  color: #374151 !important;
  font-size: 15px;
}

:deep(.custom-input .el-input__inner::placeholder) {
  color: #9ca3af !important;
}

:deep(.custom-input.is-focus .el-input__wrapper) {
  background: #ffffff !important;
  border-color: rgba(99, 102, 241, 0.5) !important;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1) !important;
}

:deep(.custom-input.is-focus + .input-icon),
.input-wrapper:has(.is-focus) .input-icon {
  color: #6366f1;
}

/* 密码显示切换按钮 */
:deep(.custom-input .el-input__suffix-inner) {
  color: #9ca3af;
}

/* 确保所有输入框宽度一致 */
:deep(.custom-input) {
  width: 100%;
}

:deep(.custom-input .el-input__wrapper) {
  width: 100%;
}

/* 提交按钮间距 */
:deep(.submit-item) {
  margin-top: 8px;
}

/* ========== 登录按钮 ========== */
.login-btn {
  position: relative;
  width: 100%;
  height: 48px;
  border: none;
  border-radius: 12px;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  overflow: hidden;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.login-btn::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, #818cf8 0%, #a78bfa 100%);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.login-btn:hover:not(:disabled)::before {
  opacity: 1;
}

.login-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 10px 30px -10px rgba(99, 102, 241, 0.4);
}

.login-btn:active:not(:disabled) {
  transform: translateY(0);
}

.login-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn-text,
.btn-arrow,
.btn-spinner {
  position: relative;
  z-index: 1;
}

.btn-arrow {
  font-size: 18px;
  transition: transform 0.3s ease;
}

.login-btn:hover:not(:disabled) .btn-arrow {
  transform: translateX(4px);
}

/* 加载动画 */
.btn-spinner {
  display: flex;
  gap: 4px;
  align-items: center;
}

.spinner-dot {
  width: 6px;
  height: 6px;
  background: #fff;
  border-radius: 50%;
  animation: spinnerBounce 1.4s ease-in-out infinite both;
}

.spinner-dot:nth-child(1) {
  animation-delay: -0.32s;
}

.spinner-dot:nth-child(2) {
  animation-delay: -0.16s;
}

@keyframes spinnerBounce {

  0%,
  80%,
  100% {
    transform: scale(0);
    opacity: 0.5;
  }

  40% {
    transform: scale(1);
    opacity: 1;
  }
}

/* ========== 提示信息 ========== */
.login-tips {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-top: 24px;
  padding: 12px 16px;
  background: #f9fafb;
  border-radius: 10px;
  border: 1px solid #f3f4f6;
}

.login-tips .el-icon {
  font-size: 14px;
  color: #9ca3af;
}

.login-tips span {
  font-size: 13px;
  color: #9ca3af;
}

/* ========== 底部版权 ========== */
.login-footer {
  position: absolute;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10;
}

.login-footer p {
  font-size: 12px;
  color: #9ca3af;
}

/* ========== 表单错误提示 ========== */
:deep(.el-form-item__error) {
  color: #ef4444;
  font-size: 12px;
  padding-top: 4px;
}

:deep(.el-form-item.is-error .el-input__wrapper) {
  border-color: rgba(239, 68, 68, 0.5) !important;
}

/* ========== 响应式设计 ========== */
@media (max-width: 960px) {
  .login-card {
    width: 90%;
    max-width: 420px;
    flex-direction: column;
    min-height: auto;
  }

  .login-card-left {
    flex: none;
    padding: 32px 28px;
    border-right: none;
    border-bottom: 1px solid #f3f4f6;
  }

  .brand-logo {
    width: 44px;
    height: 44px;
  }

  .brand-title {
    font-size: 20px;
  }

  .feature-list {
    display: none;
  }

  .login-card-right {
    padding: 32px 28px;
  }

  .form-title {
    font-size: 24px;
  }
}

@media (max-width: 480px) {
  .login-card {
    width: 95%;
    border-radius: 20px;
  }

  .login-card::before {
    border-radius: 20px;
  }

  .login-card-left,
  .login-card-right {
    padding: 24px 20px;
  }

  .form-title {
    font-size: 22px;
  }
}
</style>
