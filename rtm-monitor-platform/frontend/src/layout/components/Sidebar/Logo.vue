<template>
  <div
    class="sidebar-logo-container"
    :class="{ collapse: collapse }"
    :style="{ backgroundColor: sideTheme === 'theme-dark' ? variables.menuBackground : variables.menuLightBackground }"
    @click="goHome"
  >
    <img :src="logo" class="sidebar-logo" />
    <span
      v-if="!collapse"
      class="sidebar-title"
      :style="{ color: sideTheme === 'theme-dark' ? variables.logoTitleColor : variables.logoLightTitleColor }"
    >
      {{ title }}
    </span>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import variables from '@/assets/styles/variables.module.scss'
import logo from '@/assets/logo/logo.png'
import useSettingsStore from '@/store/modules/settings'

const router = useRouter()
const settingsStore = useSettingsStore()

const props = defineProps({
  collapse: {
    type: Boolean,
    required: true
  }
})

const title = import.meta.env.VITE_APP_TITLE
const sideTheme = computed(() => settingsStore.sideTheme)

function goHome() {
  router.push('/')
}
</script>

<style scoped lang="scss">
.sidebar-logo-container {
  height: 56px;
  width: 100%;
  display: flex;
  align-items: center;
  padding: 0 14px;
  box-sizing: border-box;
  overflow: hidden;
  flex-shrink: 0;
  cursor: pointer;

  .sidebar-logo {
    width: 28px;
    height: 28px;
    flex-shrink: 0;
    display: block;
  }

  .sidebar-title {
    margin-left: 10px;
    font-size: 20px;
    font-weight: 600;
    line-height: 1;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  }

  &.collapse {
    justify-content: center;
    padding: 0;

    .sidebar-logo {
      margin: 0;
    }
  }
}
</style>
