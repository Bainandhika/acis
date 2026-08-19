import { computed } from 'vue'
import { useUIStore } from '../stores/ui'
import type { ToastType } from '../types'

export function useUI() {
  const uiStore = useUIStore()

  const showToast = (type: ToastType, message: string, duration: number = 4000): string => {
    const id = Math.random().toString(36).substring(2, 11)
    uiStore.addToast({ id, type, message, duration })
    if (duration > 0) {
      setTimeout(() => {
        uiStore.removeToast(id)
      }, duration)
    }
    return id
  }

  return {
    toasts: computed(() => uiStore.toasts),
    isGlobalLoading: computed(() => uiStore.isGlobalLoading),
    modalStack: computed(() => uiStore.modalStack),
    isMobileSidebarOpen: computed(() => uiStore.isMobileSidebarOpen),
    isSidebarCollapsed: computed(() => uiStore.isSidebarCollapsed),
    showToast,
    removeToast: (id: string) => uiStore.removeToast(id),
    showGlobalLoading: () => uiStore.setGlobalLoading(true),
    hideGlobalLoading: () => uiStore.setGlobalLoading(false),
    pushModal: (id: string) => uiStore.pushModal(id),
    popModal: (id?: string) => uiStore.popModal(id),
    isModalOpen: (id: string) => uiStore.modalStack.includes(id),
    toggleMobileSidebar: () => uiStore.toggleMobileSidebar(),
    setMobileSidebar: (open: boolean) => uiStore.setMobileSidebar(open),
    toggleSidebarCollapse: () => uiStore.toggleSidebarCollapse()
  }
}
