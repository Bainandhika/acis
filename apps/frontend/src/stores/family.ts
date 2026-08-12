import { defineStore } from 'pinia';
import { ref } from 'vue';
import { getMyFamily, createFamily, joinFamily, type Family } from '../services/family';

export const useFamilyStore = defineStore('family', () => {
  const family = ref<Family | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchMyFamily() {
    loading.value = true;
    error.value = null;
    try {
      const { data } = await getMyFamily();
      family.value = data.data;
    } catch (err: any) {
      family.value = null;
      error.value = err.response?.data?.error || 'Failed to fetch family';
    } finally {
      loading.value = false;
    }
  }

  async function handleCreateFamily(name: string) {
    loading.value = true;
    error.value = null;
    try {
      const { data } = await createFamily(name);
      family.value = data.data;
      await fetchMyFamily();
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to create family';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function handleJoinFamily(inviteCode: string) {
    loading.value = true;
    error.value = null;
    try {
      const { data } = await joinFamily(inviteCode);
      family.value = data.data;
      await fetchMyFamily();
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to join family';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  return { family, loading, error, fetchMyFamily, handleCreateFamily, handleJoinFamily };
});
