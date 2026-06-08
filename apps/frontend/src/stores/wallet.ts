import { defineStore } from 'pinia';
import { ref } from 'vue';
import { getWallets, createWallet, type Wallet, type CreateWalletPayload } from '../services/wallet';

export const useWalletStore = defineStore('wallet', () => {
  const wallets = ref<Wallet[]>([]);
  const loading = ref(false);
  // Hardcode family ID dulu buat MVP (Nanti dari Auth context)
  const currentFamilyId = '00000000-0000-0000-0000-000000000002';

  async function fetchWallets() {
    loading.value = true;
    try {
      const { data } = await getWallets(currentFamilyId);
      wallets.value = data.data;
    } catch (error) {
      console.error('Failed to fetch wallets', error);
    } finally {
      loading.value = false;
    }
  }

  async function addWallet(payload: CreateWalletPayload) {
    await createWallet({ ...payload, family_id: currentFamilyId });
    await fetchWallets(); // Refresh list
  }

  return { wallets, loading, fetchWallets, addWallet };
});
