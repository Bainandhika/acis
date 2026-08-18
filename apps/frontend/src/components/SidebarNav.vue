<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useFamilyStore } from '../stores/family'
import { useTransactionStore } from '../stores/transaction'

export type TabKey = 'dashboard' | 'wallets' | 'transactions' | 'members' | 'settings'

const props = defineProps<{
  activeTab: TabKey
  isCollapsed?: boolean
}>()

const emit = defineEmits<{
  (e: 'select-tab', tab: TabKey): void
  (e: 'toggle-collapse'): void
  (e: 'open-settings'): void
}>()

const authStore = useAuthStore()
const familyStore = useFamilyStore()
const txStore = useTransactionStore()

const pendingProposalsCount = computed(() => {
  return txStore.proposals.filter(p => p.status === 'pending').length
})
</script>

<template>
  <aside
    class="flex flex-col justify-between h-screen sticky top-0 bg-slate-900 text-slate-300 py-6 px-4 transition-all duration-300 z-30 shrink-0 border-r border-slate-800"
    :class="isCollapsed ? 'w-20' : 'w-64'"
  >
    <!-- Brand & Top Menu -->
    <div class="flex flex-col gap-8">
      <!-- Logo Header -->
      <div class="flex items-center gap-3 px-2">
        <div class="w-10 h-10 rounded-2xl bg-teal-500/20 border border-teal-500/30 flex items-center justify-center shrink-0 shadow-lg shadow-teal-950/40">
          <svg class="w-6 h-6 text-teal-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M19 7V4a1 1 0 0 0-1-1H5a2 2 0 0 0 0 4h15a1 1 0 0 1 1 1v4h-3a2 2 0 0 0 0 4h3a8 8 0 0 1-16 0V6"></path>
            <circle cx="18" cy="14" r="1"></circle>
          </svg>
        </div>
        <div v-if="!isCollapsed" class="flex flex-col">
          <div class="flex items-center gap-1.5">
            <span class="font-extrabold text-xl tracking-tight text-white font-sans">FamFi</span>
            <span class="text-[9px] uppercase font-bold tracking-wider px-1.5 py-0.5 rounded-full bg-teal-500/20 text-teal-300 border border-teal-500/30">ACIS</span>
          </div>
          <span class="text-xs text-slate-400 font-medium truncate max-w-[130px]">
            {{ familyStore.family?.name || 'Miller Family' }}
          </span>
        </div>
      </div>

      <!-- Navigation List -->
      <nav class="flex flex-col gap-1.5">
        <!-- 1. Dashboard -->
        <button
          @click="emit('select-tab', 'dashboard')"
          class="flex items-center gap-3.5 px-3.5 py-3 rounded-2xl text-sm font-semibold transition-all group text-left cursor-pointer"
          :class="activeTab === 'dashboard'
            ? 'bg-slate-800 text-white shadow-sm ring-1 ring-slate-700/60'
            : 'text-slate-400 hover:text-slate-100 hover:bg-slate-800/50'"
        >
          <div class="w-5 h-5 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 transition-colors" :class="activeTab === 'dashboard' ? 'text-teal-400' : 'text-slate-400 group-hover:text-slate-200'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="3" width="7" height="7" rx="2"></rect>
              <rect x="14" y="3" width="7" height="7" rx="2"></rect>
              <rect x="14" y="14" width="7" height="7" rx="2"></rect>
              <rect x="3" y="14" width="7" height="7" rx="2"></rect>
            </svg>
          </div>
          <span v-if="!isCollapsed">Dashboard</span>
        </button>

        <!-- 2. Wallets -->
        <button
          @click="emit('select-tab', 'wallets')"
          class="flex items-center gap-3.5 px-3.5 py-3 rounded-2xl text-sm font-semibold transition-all group text-left cursor-pointer"
          :class="activeTab === 'wallets'
            ? 'bg-slate-800 text-white shadow-sm ring-1 ring-slate-700/60'
            : 'text-slate-400 hover:text-slate-100 hover:bg-slate-800/50'"
        >
          <div class="w-5 h-5 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 transition-colors" :class="activeTab === 'wallets' ? 'text-teal-400' : 'text-slate-400 group-hover:text-slate-200'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M20 12V8H6a2 2 0 0 1-2-2c0-1.1.9-2 2-2h12v4"></path>
              <path d="M4 6v12c0 1.1.9 2 2 2h14v-4"></path>
              <circle cx="18" cy="12" r="2"></circle>
            </svg>
          </div>
          <span v-if="!isCollapsed">Wallets</span>
        </button>

        <!-- 3. Transactions -->
        <button
          @click="emit('select-tab', 'transactions')"
          class="flex items-center justify-between px-3.5 py-3 rounded-2xl text-sm font-semibold transition-all group text-left cursor-pointer"
          :class="activeTab === 'transactions'
            ? 'bg-slate-800 text-white shadow-sm ring-1 ring-slate-700/60'
            : 'text-slate-400 hover:text-slate-100 hover:bg-slate-800/50'"
        >
          <div class="flex items-center gap-3.5">
            <div class="w-5 h-5 flex items-center justify-center shrink-0">
              <svg class="w-5 h-5 transition-colors" :class="activeTab === 'transactions' ? 'text-teal-400' : 'text-slate-400 group-hover:text-slate-200'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="8" y1="6" x2="21" y2="6"></line>
                <line x1="8" y1="12" x2="21" y2="12"></line>
                <line x1="8" y1="18" x2="21" y2="18"></line>
                <line x1="3" y1="6" x2="3.01" y2="6"></line>
                <line x1="3" y1="12" x2="3.01" y2="12"></line>
                <line x1="3" y1="18" x2="3.01" y2="18"></line>
              </svg>
            </div>
            <span v-if="!isCollapsed">Transactions</span>
          </div>
          <span
            v-if="!isCollapsed && pendingProposalsCount > 0"
            class="text-[10px] font-bold px-2 py-0.5 rounded-full bg-amber-500/20 text-amber-300 border border-amber-500/30"
          >
            {{ pendingProposalsCount }}
          </span>
        </button>

        <!-- 4. Family Members -->
        <button
          @click="emit('select-tab', 'members')"
          class="flex items-center gap-3.5 px-3.5 py-3 rounded-2xl text-sm font-semibold transition-all group text-left cursor-pointer"
          :class="activeTab === 'members'
            ? 'bg-slate-800 text-white shadow-sm ring-1 ring-slate-700/60'
            : 'text-slate-400 hover:text-slate-100 hover:bg-slate-800/50'"
        >
          <div class="w-5 h-5 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 transition-colors" :class="activeTab === 'members' ? 'text-teal-400' : 'text-slate-400 group-hover:text-slate-200'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
              <circle cx="9" cy="7" r="4"></circle>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
            </svg>
          </div>
          <span v-if="!isCollapsed">Family Members</span>
        </button>

        <!-- 5. Settings -->
        <button
          @click="emit('select-tab', 'settings')"
          class="flex items-center gap-3.5 px-3.5 py-3 rounded-2xl text-sm font-semibold transition-all group text-left cursor-pointer"
          :class="activeTab === 'settings'
            ? 'bg-slate-800 text-white shadow-sm ring-1 ring-slate-700/60'
            : 'text-slate-400 hover:text-slate-100 hover:bg-slate-800/50'"
        >
          <div class="w-5 h-5 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 transition-colors" :class="activeTab === 'settings' ? 'text-teal-400' : 'text-slate-400 group-hover:text-slate-200'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="3"></circle>
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
            </svg>
          </div>
          <span v-if="!isCollapsed">Settings</span>
        </button>
      </nav>
    </div>

    <!-- Bottom Section: Live Sync Status -->
    <div class="flex flex-col gap-3 pt-4 border-t border-slate-800/80">
      <div v-if="!isCollapsed" class="flex items-center gap-2.5 px-3 py-2 rounded-xl bg-slate-950/60 border border-slate-800">
        <span class="relative flex h-2 w-2">
          <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
          <span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
        </span>
        <span class="text-xs font-medium text-slate-400">All accounts synced</span>
      </div>

      <!-- Collapse Toggle Button -->
      <button
        @click="emit('toggle-collapse')"
        class="hidden lg:flex items-center justify-center p-2 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800/60 transition cursor-pointer"
        :title="isCollapsed ? 'Expand Sidebar' : 'Collapse Sidebar'"
      >
        <svg
          class="w-4 h-4 transition-transform duration-200"
          :class="isCollapsed ? 'rotate-180' : ''"
          viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
        >
          <polyline points="11 19 4 12 11 5"></polyline>
          <polyline points="18 19 11 12 18 5"></polyline>
        </svg>
      </button>
    </div>
  </aside>
</template>
