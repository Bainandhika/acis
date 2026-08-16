import { defineStore } from 'pinia';
import { ref } from 'vue';
import { getWallets, createWallet, type Wallet, type CreateWalletPayload } from '../services/wallet';

export const useWalletStore = defineStore('wallet', () => {
  const wallets = ref<Wallet[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchWallets() {
    loading.value = true;
    error.value = null;
    try {
      const response = await getWallets();
      wallets.value = response.data.data || [];
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to fetch wallets';
    } finally {
      loading.value = false;
    }
  }

  async function addWallet(payload: CreateWalletPayload) {
    loading.value = true;
    error.value = null;
    try {
      await createWallet(payload);
      await fetchWallets();
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to create wallet';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  function resetState() {
    wallets.value = [];
    loading.value = false;
    error.value = null;
  }

  return { wallets, loading, error, fetchWallets, addWallet, resetState };
});
