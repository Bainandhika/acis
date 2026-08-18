<script setup lang="ts">
import { ref } from 'vue'
import { useFamilyStore } from '../stores/family'
import { useRouter } from 'vue-router'
import { useI18n } from '../locales'

const familyStore = useFamilyStore()
const router = useRouter()
const { t, locale, setLocale } = useI18n()

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
    errorMessage.value = familyStore.error || t('familySetup.errors.createFailed')
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
    errorMessage.value = familyStore.error || t('familySetup.errors.joinFailed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex items-center justify-center min-h-screen bg-slate-950 text-slate-100 p-4 relative overflow-hidden">
    <!-- Ambient glow decorative circles -->
    <div class="absolute top-1/4 right-1/4 w-96 h-96 bg-brand-500/10 rounded-full blur-3xl pointer-events-none"></div>
    <div class="absolute bottom-1/4 left-1/4 w-80 h-80 bg-emerald-900/15 rounded-full blur-3xl pointer-events-none"></div>

    <div class="card-neo w-full max-w-md p-8 relative z-10 border border-slate-800 shadow-2xl bg-slate-900/95 backdrop-blur-xl">
      <!-- Top Bar: Back to Dashboard (if already in family) + Language Switcher -->
      <div class="flex items-center justify-between mb-4">
        <div>
          <button 
            v-if="familyStore.family"
            @click="router.push('/')"
            class="px-2.5 py-1 rounded-xl bg-slate-800 hover:bg-slate-700 border border-slate-700 text-[11px] font-bold text-slate-300 hover:text-white transition-all flex items-center gap-1.5 cursor-pointer"
            type="button"
          >
            <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M19 12H5M12 19l-7-7 7-7"/>
            </svg>
            <span>{{ t('familySetup.backToDashboard') }}</span>
          </button>
        </div>

        <div class="inline-flex p-1 bg-slate-950 rounded-xl border border-slate-800 text-[11px] font-bold">
          <button 
            @click="setLocale('en')"
            class="px-2.5 py-1 rounded-lg transition-all"
            :class="locale === 'en' ? 'bg-slate-800 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200'"
            type="button"
          >
            EN
          </button>
          <button 
            @click="setLocale('id')"
            class="px-2.5 py-1 rounded-lg transition-all"
            :class="locale === 'id' ? 'bg-slate-800 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200'"
            type="button"
          >
            ID
          </button>
        </div>
      </div>

      <!-- Header -->
      <div class="flex flex-col items-center text-center mb-6">
        <!-- Grayed out decorative icon -->
        <div class="w-14 h-14 rounded-3xl bg-slate-800 border border-slate-700 text-slate-300 flex items-center justify-center shadow-lg text-2xl mb-4 font-black">
          🏠
        </div>
        <h2 class="text-2xl font-black text-white tracking-tight">
          {{ t('familySetup.title') }}
        </h2>
        <p class="text-xs text-slate-400 font-medium mt-1">
          {{ t('familySetup.subtitle') }}
        </p>
      </div>

      <!-- Error Alert -->
      <div v-if="errorMessage" class="p-3.5 bg-rose-950/40 border border-rose-800/40 text-rose-300 text-xs rounded-2xl mb-6 font-semibold flex items-center gap-2">
        <svg class="w-4 h-4 shrink-0 text-rose-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="8" x2="12" y2="12"></line>
          <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
        <span>{{ errorMessage }}</span>
      </div>

      <!-- Segmented Tabs -->
      <div class="flex p-1 bg-slate-950 rounded-2xl mb-6 text-xs font-bold border border-slate-800">
        <button 
          class="flex-1 py-2.5 rounded-xl transition-all cursor-pointer"
          :class="activeTab === 'create' ? 'bg-slate-800 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200'"
          @click="activeTab === 'create'"
        >
          {{ t('familySetup.createTab') }}
        </button>
        <button 
          class="flex-1 py-2.5 rounded-xl transition-all cursor-pointer"
          :class="activeTab === 'join' ? 'bg-slate-800 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200'"
          @click="activeTab === 'join'"
        >
          {{ t('familySetup.joinTab') }}
        </button>
      </div>

      <!-- Tab 1: Create Family (Family Name Only) -->
      <div v-if="activeTab === 'create'" class="flex flex-col gap-4">
        <div>
          <label class="text-xs font-bold text-slate-300 block mb-1.5">{{ t('familySetup.familyNameLabel') }}</label>
          <input 
            type="text" 
            v-model="familyName" 
            :placeholder="t('familySetup.familyNamePlaceholder')" 
            class="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-2xl text-xs font-semibold text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-600 transition" 
            @keyup.enter="familyName.trim() && handleCreate()"
          />
        </div>

        <button 
          class="w-full py-3.5 mt-2 rounded-2xl bg-slate-800 hover:bg-slate-700 text-white font-bold text-xs transition border border-slate-700 active:scale-95 disabled:opacity-50 cursor-pointer"
          :disabled="loading || !familyName.trim()"
          @click="handleCreate"
        >
          {{ loading ? t('familySetup.creatingBtn') : t('familySetup.createBtn') }}
        </button>
      </div>

      <!-- Tab 2: Join Family (Invite Code) -->
      <div v-else class="flex flex-col gap-4">
        <div>
          <label class="text-xs font-bold text-slate-300 block mb-1.5 text-center">{{ t('familySetup.inviteCodeLabel') }}</label>
          <input 
            type="text" 
            v-model="inviteCode" 
            :placeholder="t('familySetup.inviteCodePlaceholder')" 
            maxlength="6"
            class="w-full py-3.5 bg-slate-800/80 border border-slate-700 rounded-2xl text-center text-2xl font-black tracking-[0.3em] uppercase text-white focus:outline-none focus:ring-2 focus:ring-slate-600 transition font-mono" 
            @keyup.enter="inviteCode.trim().length === 6 && handleJoin()"
          />
          <p class="text-[10px] text-slate-400 text-center mt-1.5">{{ t('familySetup.inviteCodeHint') }}</p>
        </div>

        <button 
          class="w-full py-3.5 mt-2 rounded-2xl bg-slate-800 hover:bg-slate-700 text-white font-bold text-xs transition border border-slate-700 active:scale-95 disabled:opacity-50 cursor-pointer"
          :disabled="loading || inviteCode.trim().length !== 6"
          @click="handleJoin"
        >
          {{ loading ? t('familySetup.joiningBtn') : t('familySetup.joinBtn') }}
        </button>
      </div>
    </div>
  </div>
</template>
