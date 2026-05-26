import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Component, ComponentCatalogItem } from '@/types'
import { componentApi } from '@/api/component'

export const useComponentStore = defineStore('component', () => {
  const components = ref<Component[]>([])
  const catalog = ref<ComponentCatalogItem[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchComponents() {
    loading.value = true
    error.value = null
    try {
      components.value = await componentApi.list()
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取组件列表失败'
    } finally {
      loading.value = false
    }
  }

  async function fetchCatalog() {
    loading.value = true
    error.value = null
    try {
      catalog.value = await componentApi.getCatalog()
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取组件目录失败'
    } finally {
      loading.value = false
    }
  }

  async function installComponent(data: any) {
    loading.value = true
    error.value = null
    try {
      const component = await componentApi.install(data)
      components.value.push(component)
      return component
    } catch (err) {
      error.value = err instanceof Error ? err.message : '安装组件失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function uninstallComponent(id: string) {
    loading.value = true
    error.value = null
    try {
      await componentApi.uninstall(id)
      components.value = components.value.filter(c => c.id !== id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '卸载组件失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function controlComponent(id: string, action: 'start' | 'stop' | 'restart') {
    loading.value = true
    error.value = null
    try {
      await componentApi.control(id, action)
      // 更新组件状态
      const component = components.value.find(c => c.id === id)
      if (component) {
        if (action === 'start') component.status = 'running'
        if (action === 'stop') component.status = 'stopped'
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '操作失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    components,
    catalog,
    loading,
    error,
    fetchComponents,
    fetchCatalog,
    installComponent,
    uninstallComponent,
    controlComponent
  }
})
