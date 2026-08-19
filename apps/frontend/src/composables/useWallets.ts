import { computed, onMounted, getCurrentInstance } from 'vue'
import { useWalletStore } from '../stores/wallet'
import { useApi } from './useApi'
import { useUI } from './useUI'
import type { UpdateWalletPayload } from '../services/wallet'
import * as walletService from '../services/wallet'

export function useWallets() {
  const walletStore = useWalletStore()
  const { showToast } = useUI()

  const wallets = computed(() => walletStore.wallets)
  const totalBalance = computed(() =>
    wallets.value.reduce((sum, w) => sum + (w?.current_balance || 0), 0)
  )

  const fetchWalletsApi = useApi(walletService.getWallets, {
    onSuccess: (data) => {
      const list = Array.isArray(data) ? data : (data as any)?.data || []
      walletStore.wallets = list
    }
  })

  const addWalletApi = useApi(walletService.createWallet, {
    onSuccess: async () => {
      showToast('success', 'Wallet created successfully')
      await fetchWalletsApi.execute()
    }
  })

  const editWalletApi = useApi(
    (id: string, payload: UpdateWalletPayload) => walletService.updateWallet(id, payload),
    {
      onSuccess: async () => {
        showToast('success', 'Wallet updated successfully')
        await fetchWalletsApi.execute()
      }
    }
  )

  const removeWalletApi = useApi(
    (id: string) => walletService.deleteWallet(id),
    {
      onSuccess: async () => {
        showToast('info', 'Wallet deleted successfully')
        await fetchWalletsApi.execute()
      }
    }
  )

  if (getCurrentInstance()) {
    onMounted(() => {
      if (wallets.value.length === 0) {
        fetchWalletsApi.execute()
      }
    })
  }

  return {
    wallets,
    totalBalance,
    isLoading: computed(() => fetchWalletsApi.isLoading.value || walletStore.loading),
    isAdding: computed(() => addWalletApi.isLoading.value),
    isEditing: computed(() => editWalletApi.isLoading.value),
    isRemoving: computed(() => removeWalletApi.isLoading.value),
    error: computed(() => fetchWalletsApi.error.value || walletStore.error),
    fetchWallets: fetchWalletsApi.execute,
    addWallet: addWalletApi.execute,
    editWallet: editWalletApi.execute,
    removeWallet: removeWalletApi.execute
  }
}
