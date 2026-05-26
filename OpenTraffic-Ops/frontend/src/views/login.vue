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
            <svg-icon icon-class="dashboard" class="logo-svg" />
          </div>
          <h2 class="brand-title">OpenTraffic Ops 监控运维平台</h2>
          <p class="brand-desc">实时主机监控 · 智能告警 · 远程运维</p>
        </div>
        <div class="feature-list">
          <div class="feature-item">
            <div class="feature-icon-wrap">
              <svg-icon icon-class="server" class="feature-svg" />
            </div>
            <span>主机监控与告警</span>
          </div>
          <div class="feature-item">
            <div class="feature-icon-wrap">
              <svg-icon icon-class="tool" class="feature-svg" />
            </div>
            <span>远程终端与文件</span>
          </div>
          <div class="feature-item">
            <div class="feature-icon-wrap">
              <svg-icon icon-class="monitor" class="feature-svg" />
            </div>
            <span>智能 Agent 控制</span>
          </div>
        </div>
      </div>

      <!-- 右侧表单区 -->
      <div class="login-card-right">
        <div class="form-header">
          <h3 class="form-title">欢迎回来</h3>
          <p class="form-subtitle">请登录您的账号以继续使用</p>
        </div>

        <el-form ref="loginRef" :model="loginForm" :rules="loginRules" class="login-form">
          <el-form-item prop="username">
            <div class="input-label">账号</div>
            <div class="input-wrapper">
              <svg-icon icon-class="user" class="input-icon" />
              <el-input style="width: 375px;" v-model="loginForm.username" type="text" size="large" auto-complete="off"
                placeholder="请输入账号" class="custom-input" :class="{ 'is-focus': focusField === 'username' }"
                @focus="focusField = 'username'" @blur="focusField = ''" />
            </div>
          </el-form-item>

          <el-form-item prop="password">
            <div class="input-label">密码</div>
            <div class="input-wrapper">
              <svg-icon icon-class="password" class="input-icon" />
              <el-input style="width: 375px;" v-model="loginForm.password" type="password" size="large"
                auto-complete="off" placeholder="请输入密码" class="custom-input"
                :class="{ 'is-focus': focusField === 'password' }" show-password @focus="focusField = 'password'"
                @blur="focusField = ''" @keyup.enter="handleLogin" />
            </div>
          </el-form-item>

          <el-form-item prop="code" v-if="captchaEnabled">
            <div class="input-label">验证码</div>
            <div class="code-row">
              <div class="input-wrapper code-wrapper">
                <svg-icon icon-class="validCode" class="input-icon" />
                <el-input v-model="loginForm.code" size="large" auto-complete="off" placeholder="请输入验证码"
                  class="custom-input" :class="{ 'is-focus': focusField === 'code' }" @focus="focusField = 'code'"
                  @blur="focusField = ''" @keyup.enter="handleLogin" />
              </div>
              <div class="code-image" @click="getCode">
                <img :src="codeUrl" alt="验证码" />
              </div>
            </div>
          </el-form-item>

          <el-form-item class="submit-item">
            <button type="button" class="login-btn" :class="{ 'is-loading': loading }" :disabled="loading"
              @click.prevent="handleLogin">
              <span class="btn-text">{{ loading ? '登录中...' : '登 录' }}</span>
            </button>
          </el-form-item>
        </el-form>
      </div>
    </div>

    <!-- 底部版权 -->
    <div class="login-footer">
      <p> OpenTraffic Ops Monitor Platform </p>
    </div>
  </div>
</template>

<script setup>
import { getCodeImg } from "@/api/login";
import useUserStore from '@/store/modules/user'

const userStore = useUserStore()
const route = useRoute();
const router = useRouter();
const { proxy } = getCurrentInstance();

const loginForm = ref({
  username: "",
  password: "",
  code: "",
  uuid: ""
});

const loginRules = {
  username: [{ required: true, trigger: "blur", message: "请输入您的账号" }],
  password: [{ required: true, trigger: "blur", message: "请输入您的密码" }],
  code: [{ required: true, trigger: "change", message: "请输入验证码" }]
};

const codeUrl = ref("");
const loading = ref(false);
const captchaEnabled = ref(true);
const redirect = ref(undefined);
const mounted = ref(false);
const focusField = ref('');

watch(route, (newRoute) => {
  redirect.value = newRoute.query && newRoute.query.redirect;
}, { immediate: true });

function handleLogin() {
  proxy.$refs.loginRef.validate(valid => {
    if (valid) {
      loading.value = true;
      userStore.login(loginForm.value).then(() => {
        const query = route.query;
        const otherQueryParams = Object.keys(query).reduce((acc, cur) => {
          if (cur !== "redirect") {
            acc[cur] = query[cur];
          }
          return acc;
        }, {});
        router.push({ path: redirect.value || "/", query: otherQueryParams });
      }).catch(() => {
        loading.value = false;
        if (captchaEnabled.value) {
          getCode();
        }
      });
    }
  });
}

function getCode() {
  getCodeImg().then(res => {
    const data = res.data || res;
    captchaEnabled.value = data.captchaEnabled === undefined ? true : data.captchaEnabled;
    if (captchaEnabled.value) {
      codeUrl.value = data.img;
      loginForm.value.uuid = data.uuid;
    }
  });
}

function getParticleStyle(index) {
  const size = (index * 7 + 3) % 4 + 2;
  const left = (index * 17 + 5) % 100;
  const delay = (index * 1.3) % 15;
  const duration = (index * 2.1) % 10 + 10;
  return {
    width: `${size}px`,
    height: `${size}px`,
    left: `${left}%`,
    animationDelay: `${delay}s`,
    animationDuration: `${duration}s`
  };
}

getCode();

onMounted(() => {
  setTimeout(() => {
    mounted.value = true;
  }, 100);
});
</script>

<style lang="scss" scoped>
.login-page {
  position: relative;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f1f5f9;
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
  opacity: 0.25;
  animation: orbFloat 20s ease-in-out infinite;
}

.orb-1 {
  width: 600px;
  height: 600px;
  background: radial-gradient(circle, #3b82f6 0%, transparent 70%);
  top: -10%;
  left: -5%;
  animation-delay: 0s;
}

.orb-2 {
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, #6366f1 0%, transparent 70%);
  bottom: -10%;
  right: -5%;
  animation-delay: -7s;
  animation-duration: 25s;
}

.orb-3 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, #14b8a6 0%, transparent 70%);
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
    linear-gradient(rgba(0, 0, 0, 0.025) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 0, 0, 0.025) 1px, transparent 1px);
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
  background: rgba(59, 130, 246, 0.3);
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
  border: 1px solid #e2e8f0;
  border-radius: 24px;
  overflow: hidden;
  box-shadow:
    0 25px 50px -12px rgba(0, 0, 0, 0.08),
    0 0 0 1px rgba(255, 255, 255, 0.5) inset;
  opacity: 0;
  transform: translateY(30px) scale(0.98);
  transition: all 0.8s cubic-bezier(0.16, 1, 0.3, 1);
}

.login-card.card-enter {
  opacity: 1;
  transform: translateY(0) scale(1);
}

/* 卡片发光边框效果 */
.login-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 24px;
  padding: 1.5px;
  background: linear-gradient(135deg,
      rgba(59, 130, 246, 0.25) 0%,
      transparent 30%,
      transparent 70%,
      rgba(99, 102, 241, 0.25) 100%);
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
  background: linear-gradient(160deg,
      rgba(59, 130, 246, 0.06) 0%,
      rgba(99, 102, 241, 0.04) 50%,
      rgba(20, 184, 166, 0.03) 100%);
  border-right: 1px solid #f1f5f9;
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
      rgba(59, 130, 246, 0.05) 0%,
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
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #3b82f6 0%, #6366f1 100%);
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(59, 130, 246, 0.25);

  .logo-svg {
    font-size: 28px;
    color: #fff;
  }
}

.brand-title {
  font-size: 22px;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 8px;
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
  gap: 14px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #64748b;
  font-size: 13px;
  transition: all 0.3s ease;
  cursor: default;

  &:hover {
    color: #334155;
    transform: translateX(4px);
  }
}

.feature-icon-wrap {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f1f5f9;
  border-radius: 10px;
  transition: all 0.3s ease;

  .feature-svg {
    font-size: 16px;
    color: #3b82f6;
  }
}

.feature-item:hover .feature-icon-wrap {
  background: rgba(59, 130, 246, 0.1);
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
}

.form-title {
  font-size: 26px;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 6px;
  letter-spacing: -0.3px;
}

.form-subtitle {
  font-size: 14px;
  color: #94a3b8;
  margin: 0;
}

.login-form {
  :deep(.el-form-item) {
    margin-bottom: 24px;
  }

  :deep(.el-form-item__error) {
    padding-top: 4px;
    font-size: 12px;
    color: #ef4444;
  }

  :deep(.el-form-item.is-error .el-input__wrapper) {
    border-color: rgba(239, 68, 68, 0.5) !important;
    box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.08) !important;
  }
}

.input-label {
  font-size: 13px;
  font-weight: 500;
  color: #475569;
  margin-bottom: 10px;
  margin-left: 15px;
  margin-right: 10px;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;

  .input-icon {
    position: absolute;
    left: 14px;
    z-index: 1;
    font-size: 16px;
    color: #94a3b8;
    transition: color 0.3s ease;
    pointer-events: none;
  }
}

.code-wrapper {
  flex: 1;
}

.custom-input {
  width: 100%;

  :deep(.el-input__wrapper) {
    background: #f8fafc !important;
    border: 1px solid #e2e8f0 !important;
    border-radius: 12px !important;
    box-shadow: none !important;
    padding-left: 42px !important;
    padding-right: 14px !important;
    height: 46px;
    transition: all 0.25s ease;

    &:hover {
      border-color: #cbd5e1 !important;
    }
  }

  :deep(.el-input__inner) {
    color: #1e293b !important;
    font-size: 14px;

    &::placeholder {
      color: #cbd5e1 !important;
    }
  }

  :deep(.el-input__suffix-inner) {
    color: #94a3b8;
  }

  &.is-focus {
    :deep(.el-input__wrapper) {
      background: #ffffff !important;
      border-color: rgba(59, 130, 246, 0.5) !important;
      box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.08) !important;
    }

    &+.input-icon,
    &~.input-icon {
      color: #3b82f6;
    }
  }
}

/* 解决验证码区图标颜色 */
.input-wrapper:has(.is-focus) .input-icon {
  color: #3b82f6;
}

.code-row {
  display: flex;
  gap: 12px;
  align-items: center;
}

.code-image {
  width: 120px;
  height: 46px;
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  background: #f8fafc;
  box-shadow: 0 0 0 1px #e2e8f0 inset;
  transition: box-shadow 0.2s;
  flex-shrink: 0;

  &:hover {
    box-shadow: 0 0 0 1px #cbd5e1 inset;
  }

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.submit-item {
  margin-top: 8px;
  margin-bottom: 0 !important;
}

/* ========== 登录按钮 ========== */
.login-btn {
  position: relative;
  width: 100%;
  height: 48px;
  border: none;
  border-radius: 12px;
  background: linear-gradient(135deg, #2563eb 0%, #3b82f6 100%);
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 2px;
  cursor: pointer;
  overflow: hidden;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;

  &::before {
    content: '';
    position: absolute;
    inset: 0;
    background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%);
    opacity: 0;
    transition: opacity 0.3s ease;
  }

  &:hover:not(:disabled)::before {
    opacity: 1;
  }

  &:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 10px 30px -10px rgba(37, 99, 235, 0.4);
  }

  &:active:not(:disabled) {
    transform: translateY(0);
  }

  &:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }
}

.btn-text,
.btn-arrow,
.btn-spinner {
  position: relative;
  z-index: 1;
}

.btn-arrow {
  font-size: 16px;
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
  width: 5px;
  height: 5px;
  background: #fff;
  border-radius: 50%;
  animation: spinnerBounce 1.4s ease-in-out infinite both;

  &:nth-child(1) {
    animation-delay: -0.32s;
  }

  &:nth-child(2) {
    animation-delay: -0.16s;
  }
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
  padding: 10px 16px;
  background: #f8fafc;
  border-radius: 10px;
  border: 1px solid #f1f5f9;

  .tips-icon {
    font-size: 13px;
    color: #94a3b8;
  }

  span {
    font-size: 12px;
    color: #94a3b8;
  }
}

/* ========== 底部版权 ========== */
.login-footer {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10;

  p {
    font-size: 12px;
    color: #94a3b8;
    margin: 0;
  }
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
    padding: 28px 24px;
    border-right: none;
    border-bottom: 1px solid #f1f5f9;
  }

  .brand-logo {
    width: 48px;
    height: 48px;

    .logo-svg {
      font-size: 24px;
    }
  }

  .brand-title {
    font-size: 18px;
  }

  .feature-list {
    display: none;
  }

  .login-card-right {
    padding: 32px 28px;
  }

  .form-title {
    font-size: 22px;
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
    font-size: 20px;
  }
}
</style>
