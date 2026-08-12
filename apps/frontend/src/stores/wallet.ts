import { defineStore } from 'pinia';
import { ref } from 'vue';
import { getWallets, createWallet, type Wallet, type CreateWalletPayload } from '../services/wallet';

export const useWalletStore = defineStore('wallet', () => {
  const wallets = ref<Wallet[]>([]);
  const loading = ref(false);

  async function fetchWallets() {
    loading.value = true;
    try {
      const { data } = await getWallets();
      wallets.value = data.data;
    } catch (error) {
      console.error('Failed to fetch wallets', error);
    } finally {
      loading.value = false;
    }
  }

  async function addWallet(payload: CreateWalletPayload) {
    await createWallet(payload);
    await fetchWallets(); // Refresh list
  }

  return { wallets, loading, fetchWallets, addWallet };
});

