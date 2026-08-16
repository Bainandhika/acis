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
    await familyStore.handleCreateFamily(familyName.value.trim(), 0)
    router.push('/')
  } catch {
    errorMessage.value = familyStore.error || 'Failed to create family group.'
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
  } catch {
    errorMessage.value = familyStore.error || 'Invalid or expired invite code.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex items-center justify-center min-h-screen bg-[#F8FAFC] p-4 relative overflow-hidden">
    <!-- Ambient glow decorative circles -->
    <div class="absolute top-1/4 right-1/4 w-96 h-96 bg-brand-200/40 rounded-full blur-3xl pointer-events-none"></div>
    <div class="absolute bottom-1/4 left-1/4 w-80 h-80 bg-lime-200/30 rounded-full blur-3xl pointer-events-none"></div>

    <div class="card-neo w-full max-w-md p-8 relative z-10 border border-slate-200/80 shadow-2xl bg-white/95 backdrop-blur-xl">
      <!-- Header -->
      <div class="flex flex-col items-center text-center mb-6">
        <div class="w-14 h-14 rounded-3xl bg-gradient-to-tr from-brand-500 to-lime-300 flex items-center justify-center shadow-lg shadow-brand-500/30 text-white font-black text-2xl mb-4">
          🏠
        </div>
        <h2 class="text-2xl font-black text-slate-900 tracking-tight">
          Family Setup
        </h2>
        <p class="text-xs text-slate-400 font-medium mt-1">
          Create a new family workspace or join an existing one using an invite code
        </p>
      </div>

      <!-- Error Alert -->
      <div v-if="errorMessage" class="p-3.5 bg-rose-50 border border-rose-200 text-rose-700 text-xs rounded-2xl mb-6 font-semibold flex items-center gap-2">
        <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="8" x2="12" y2="12"></line>
          <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
        <span>{{ errorMessage }}</span>
      </div>

      <!-- Segmented Tabs -->
      <div class="flex p-1 bg-slate-100 rounded-2xl mb-6 text-xs font-extrabold">
        <button 
          class="flex-1 py-2.5 rounded-xl transition-all"
          :class="activeTab === 'create' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-800'"
          @click="activeTab = 'create'"
        >
          Create Family
        </button>
        <button 
          class="flex-1 py-2.5 rounded-xl transition-all"
          :class="activeTab === 'join' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-800'"
          @click="activeTab = 'join'"
        >
          Join Family
        </button>
      </div>

      <!-- Tab 1: Create Family (Family Name Only) -->
      <div v-if="activeTab === 'create'" class="flex flex-col gap-4">
        <div>
          <label class="text-xs font-bold text-slate-700 block mb-1.5">Family Name</label>
          <input 
            type="text" 
            v-model="familyName" 
            placeholder="e.g. Smith Family" 
            class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-semibold text-slate-900 focus:outline-none focus:ring-2 focus:ring-brand-400 focus:bg-white transition" 
            @keyup.enter="familyName.trim() && handleCreate()"
          />
        </div>

        <button 
          class="w-full py-3.5 mt-2 rounded-2xl bg-slate-900 hover:bg-slate-800 text-white font-extrabold text-xs transition shadow-md active:scale-95 disabled:opacity-50"
          :disabled="loading || !familyName.trim()"
          @click="handleCreate"
        >
          {{ loading ? 'Creating Workspace...' : 'Create Family Group' }}
        </button>
      </div>

      <!-- Tab 2: Join Family (Invite Code) -->
      <div v-else class="flex flex-col gap-4">
        <div>
          <label class="text-xs font-bold text-slate-700 block mb-1.5 text-center">Invite Code (6 Characters)</label>
          <input 
            type="text" 
            v-model="inviteCode" 
            placeholder="e.g. SMTH01" 
            maxlength="6"
            class="w-full py-3.5 bg-slate-50 border border-slate-200 rounded-2xl text-center text-2xl font-black tracking-[0.3em] uppercase text-slate-900 focus:outline-none focus:ring-2 focus:ring-brand-400 focus:bg-white transition font-mono" 
            @keyup.enter="inviteCode.trim().length === 6 && handleJoin()"
          />
          <p class="text-[10px] text-slate-400 text-center mt-1.5">Ask your Family Admin for the 6-character invitation code.</p>
        </div>

        <button 
          class="w-full py-3.5 mt-2 rounded-2xl bg-slate-900 hover:bg-slate-800 text-white font-extrabold text-xs transition shadow-md active:scale-95 disabled:opacity-50"
          :disabled="loading || inviteCode.trim().length !== 6"
          @click="handleJoin"
        >
          {{ loading ? 'Joining Family...' : 'Join Family Group' }}
        </button>
      </div>
    </div>
  </div>
</template>
