import { defineStore } from 'pinia';
import { ref } from 'vue';
import { getMyFamily, createFamily, joinFamily, updateFamilyName, updateFamilySettings, disconnectTelegram, type Family } from '../services/family';

export const useFamilyStore = defineStore('family', () => {
  const family = ref<Family | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchMyFamily() {
    loading.value = true;
    error.value = null;
    try {
      const response = await getMyFamily();
      family.value = response.data.data;
    } catch (err: any) {
      if (err.response?.status !== 404) {
        error.value = err.response?.data?.error || 'Failed to fetch family profile';
      }
      family.value = null;
    } finally {
      loading.value = false;
    }
  }

  async function handleUpdateFamilyName(name: string) {
    loading.value = true;
    error.value = null;
    try {
      await updateFamilyName(name);
      if (family.value) {
        family.value.name = name;
      }
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to update family name';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function handleCreateFamily(name: string, monthlyIncome: number = 0) {
    loading.value = true;
    error.value = null;
    try {
      const response = await createFamily(name, monthlyIncome);
      family.value = response.data.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to create family group';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function handleJoinFamily(inviteCode: string) {
    loading.value = true;
    error.value = null;
    try {
      const response = await joinFamily(inviteCode);
      family.value = response.data.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Invalid or expired invite code';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function handleUpdateSettings(monthlyIncome: number) {
    loading.value = true;
    error.value = null;
    try {
      await updateFamilySettings(monthlyIncome);
      if (family.value) {
        family.value.monthly_income = monthlyIncome;
      }
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to update family settings';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function handleDisconnectTelegram() {
    loading.value = true;
    error.value = null;
    try {
      await disconnectTelegram();
      if (family.value) {
        family.value.telegram_chat_id = undefined;
      }
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to disconnect Telegram bot';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  function resetState() {
    family.value = null;
    loading.value = false;
    error.value = null;
  }

  return { 
    family, 
    loading, 
    error, 
    fetchMyFamily, 
    handleUpdateFamilyName,
    handleCreateFamily, 
    handleJoinFamily, 
    handleUpdateSettings, 
    handleDisconnectTelegram,
    resetState
  };
});
