<script setup lang="ts">
import { ref } from 'vue'
import { useFamilyStore } from '../stores/family'
import { useRouter } from 'vue-router'

const familyStore = useFamilyStore()
const router = useRouter()

const activeTab = ref<'create' | 'join'>('create')
const familyName = ref('')
const inviteCode = ref('')
const loading = ref(false)
const errorMessage = ref('')

const handleCreate = async () => {
  if (!familyName.value.trim()) return
  loading.value = true
  errorMessage.value = ''
  try {
    await familyStore.handleCreateFamily(familyName.value.trim())
    router.push('/')
  } catch (err: any) {
    errorMessage.value = familyStore.error || 'Gagal membuat keluarga.'
  } finally {
    loading.value = false
  }
}

const handleJoin = async () => {
  if (!inviteCode.value.trim()) return
  loading.value = true
  errorMessage.value = ''
  try {
    await familyStore.handleJoinFamily(inviteCode.value.trim().toUpperCase())
    router.push('/')
  } catch (err: any) {
    errorMessage.value = familyStore.error || 'Kode invite tidak valid.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex items-center justify-center min-h-screen bg-base-200 p-4">
    <div class="card w-full max-w-md bg-base-100 shadow-xl">
      <div class="card-body">
        <h2 class="card-title justify-center text-2xl font-bold mb-2">
          Setup Keluarga ACIS 🏠
        </h2>
        <p class="text-sm text-center text-gray-500 mb-4">
          Buat grup keluarga baru atau bergabung dengan keluarga yang sudah ada.
        </p>

        <div v-if="errorMessage" class="alert alert-error text-sm py-2 shadow-lg mb-4">
          <span>{{ errorMessage }}</span>
        </div>

        <div class="tabs tabs-boxed mb-6 justify-center">
          <a 
            class="tab" 
            :class="{ 'tab-active': activeTab === 'create' }"
            @click="activeTab = 'create'"
          >
            Buat Keluarga
          </a>
          <a 
            class="tab" 
            :class="{ 'tab-active': activeTab === 'join' }"
            @click="activeTab = 'join'"
          >
            Gabung Keluarga
          </a>
        </div>

        <!-- Tab 1: Create Family -->
        <div v-if="activeTab === 'create'">
          <div class="form-control w-full">
            <label class="label"><span class="label-text">Nama Keluarga</span></label>
            <input 
              type="text" 
              v-model="familyName" 
              placeholder="Contoh: Keluarga Cemara" 
              class="input input-bordered w-full" 
            />
          </div>
          <div class="card-actions justify-end mt-6">
            <button 
              class="btn btn-primary w-full" 
              :disabled="loading || !familyName.trim()"
              @click="handleCreate"
            >
              {{ loading ? 'Memproses...' : 'Buat Sekarang' }}
            </button>
          </div>
        </div>

        <!-- Tab 2: Join Family -->
        <div v-else>
          <div class="form-control w-full">
            <label class="label"><span class="label-text">Kode Invite (6 Karakter)</span></label>
            <input 
              type="text" 
              v-model="inviteCode" 
              placeholder="Contoh: AB12CD" 
              maxlength="6"
              class="input input-bordered w-full text-center text-xl tracking-widest uppercase" 
            />
          </div>
          <div class="card-actions justify-end mt-6">
            <button 
              class="btn btn-primary w-full" 
              :disabled="loading || inviteCode.trim().length !== 6"
              @click="handleJoin"
            >
              {{ loading ? 'Memproses...' : 'Gabung Keluarga' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
