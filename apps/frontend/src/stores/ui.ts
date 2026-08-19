import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Toast } from '../types'

export const useUIStore = defineStore('ui', () => {
  const toasts = ref<Toast[]>([])
  const isGlobalLoading = ref(false)
  const modalStack = ref<string[]>([])
  const isMobileSidebarOpen = ref(false)
  const isSidebarCollapsed = ref(false)

  const addToast = (toast: Toast) => {
    toasts.value.push(toast)
    if (toasts.value.length > 5) {
      toasts.value.shift()
    }
  }

  const removeToast = (id: string) => {
    const index = toasts.value.findIndex((t) => t.id === id)
    if (index > -1) {
      toasts.value.splice(index, 1)
    }
  }

  const clearToasts = () => {
    toasts.value = []
  }

  const setGlobalLoading = (value: boolean) => {
    isGlobalLoading.value = value
  }

  const pushModal = (id: string) => {
    if (!modalStack.value.includes(id)) {
      modalStack.value.push(id)
    }
  }

  const popModal = (id?: string) => {
    if (id) {
      const idx = modalStack.value.indexOf(id)
      if (idx > -1) {
        modalStack.value.splice(idx, 1)
      }
    } else {
      modalStack.value.pop()
    }
  }

  const toggleMobileSidebar = () => {
    isMobileSidebarOpen.value = !isMobileSidebarOpen.value
  }

  const setMobileSidebar = (open: boolean) => {
    isMobileSidebarOpen.value = open
  }

  const toggleSidebarCollapse = () => {
    isSidebarCollapsed.value = !isSidebarCollapsed.value
  }

  return {
    toasts,
    isGlobalLoading,
    modalStack,
    isMobileSidebarOpen,
    isSidebarCollapsed,
    addToast,
    removeToast,
    clearToasts,
    setGlobalLoading,
    pushModal,
    popModal,
    toggleMobileSidebar,
    setMobileSidebar,
    toggleSidebarCollapse
  }
})
