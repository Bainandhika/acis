<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useFamilyStore } from '../stores/family'
import { useTransactionStore } from '../stores/transaction'

const props = defineProps<{
  activeTab: 'dashboard' | 'wallets' | 'transactions' | 'proposals'
  isCollapsed?: boolean
}>()

const emit = defineEmits<{
  (e: 'select-tab', tab: 'dashboard' | 'wallets' | 'transactions' | 'proposals'): void
  (e: 'toggle-collapse'): void
}>()

const authStore = useAuthStore()
const familyStore = useFamilyStore()
const txStore = useTransactionStore()

const pendingProposalsCount = computed(() => {
  return txStore.proposals.filter(p => p.status === 'pending').length
})

const copyInviteCode = () => {
  if (familyStore.family?.invite_code) {
    navigator.clipboard.writeText(familyStore.family.invite_code)
    alert('Kode invite berhasil disalin!')
  }
}
</script>

<template>
  <aside 
    class="flex flex-col justify-between h-screen sticky top-0 bg-white border-r border-slate-100/90 py-6 px-4 transition-all duration-300 z-30"
    :class="isCollapsed ? 'w-20' : 'w-64'"
  >
    <!-- Brand & Top Menu -->
    <div class="flex flex-col gap-8">
      <!-- Logo Header -->
      <div class="flex items-center gap-3 px-2">
        <div class="w-10 h-10 rounded-2xl bg-gradient-to-tr from-brand-500 to-lime-300 flex items-center justify-center shadow-md shadow-brand-500/20 text-white font-black text-xl shrink-0">
          A
        </div>
        <div v-if="!isCollapsed" class="flex flex-col">
          <div class="flex items-center gap-1.5">
            <span class="font-extrabold text-xl tracking-tight text-slate-900">ACIS</span>
            <span class="text-[10px] uppercase font-bold tracking-wider px-1.5 py-0.5 rounded-full bg-brand-100 text-brand-700">v2.0</span>
          </div>
          <span class="text-xs text-slate-400 font-medium truncate max-w-[140px]">
            {{ familyStore.family?.name || 'Smart Family' }}
          </span>
        </div>
      </div>

      <!-- Navigation List -->
      <nav class="flex flex-col gap-1.5">
        <button
          @click="emit('select-tab', 'dashboard')"
          class="flex items-center gap-3.5 px-3.5 py-3 rounded-2xl text-sm font-semibold transition-all group text-left relative"
          :class="activeTab === 'dashboard' 
            ? 'bg-slate-900 text-white shadow-sm' 
            : 'text-slate-500 hover:text-slate-900 hover:bg-slate-50'"
        >
          <!-- Active bar indicator on left -->
          <div 
            v-if="activeTab === 'dashboard'" 
            class="absolute left-0 w-1.5 h-6 bg-brand-400 rounded-r-full"
          ></div>
          <!-- Dashboard Grid Icon -->
          <svg class="w-5 h-5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="3" y="3" width="7" height="7" rx="2"></rect>
            <rect x="14" y="3" width="7" height="7" rx="2"></rect>
            <rect x="14" y="14" width="7" height="7" rx="2"></rect>
            <rect x="3" y="14" width="7" height="7" rx="2"></rect>
          </svg>
          <span v-if="!isCollapsed">Dashboard</span>
        </button>

        <button
          @click="emit('select-tab', 'wallets')"
          class="flex items-center gap-3.5 px-3.5 py-3 rounded-2xl text-sm font-semibold transition-all group text-left relative"
          :class="activeTab === 'wallets' 
            ? 'bg-slate-900 text-white shadow-sm' 
            : 'text-slate-500 hover:text-slate-900 hover:bg-slate-50'"
        >
          <div 
            v-if="activeTab === 'wallets'" 
            class="absolute left-0 w-1.5 h-6 bg-brand-400 rounded-r-full"
          ></div>
          <!-- Wallet / Cards Icon -->
          <svg class="w-5 h-5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="2" y="5" width="20" height="14" rx="3"></rect>
            <line x1="2" y1="10" x2="22" y2="10"></line>
          </svg>
          <span v-if="!isCollapsed">Dompet & Amplop</span>
        </button>

        <button
          @click="emit('select-tab', 'transactions')"
          class="flex items-center justify-between px-3.5 py-3 rounded-2xl text-sm font-semibold transition-all group text-left relative"
          :class="activeTab === 'transactions' 
            ? 'bg-slate-900 text-white shadow-sm' 
            : 'text-slate-500 hover:text-slate-900 hover:bg-slate-50'"
        >
          <div class="flex items-center gap-3.5">
            <div 
              v-if="activeTab === 'transactions'" 
              class="absolute left-0 w-1.5 h-6 bg-brand-400 rounded-r-full"
            ></div>
            <!-- Transactions Icon -->
            <svg class="w-5 h-5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="17 1 21 5 17 9"></polyline>
              <path d="M3 11V9a4 4 0 0 1 4-4h14"></path>
              <polyline points="7 23 3 19 7 15"></polyline>
              <path d="M21 13v2a4 4 0 0 1-4 4H3"></path>
            </svg>
            <span v-if="!isCollapsed">Riwayat Transaksi</span>
          </div>
          <span 
            v-if="!isCollapsed && txStore.transactions.length > 0" 
            class="text-[11px] font-bold px-2 py-0.5 rounded-full"
            :class="activeTab === 'transactions' ? 'bg-brand-400 text-slate-900' : 'bg-slate-100 text-slate-600'"
          >
            {{ txStore.transactions.length }}
          </span>
        </button>

        <button
          @click="emit('select-tab', 'proposals')"
          class="flex items-center justify-between px-3.5 py-3 rounded-2xl text-sm font-semibold transition-all group text-left relative"
          :class="activeTab === 'proposals' 
            ? 'bg-slate-900 text-white shadow-sm' 
            : 'text-slate-500 hover:text-slate-900 hover:bg-slate-50'"
        >
          <div class="flex items-center gap-3.5">
            <div 
              v-if="activeTab === 'proposals'" 
              class="absolute left-0 w-1.5 h-6 bg-brand-400 rounded-r-full"
            ></div>
            <!-- Proposal / Clipboard Icon -->
            <svg class="w-5 h-5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"></path>
              <rect x="8" y="2" width="8" height="4" rx="1" ry="1"></rect>
            </svg>
            <span v-if="!isCollapsed">Pengajuan Dana</span>
          </div>
          <span 
            v-if="!isCollapsed && pendingProposalsCount > 0" 
            class="text-[11px] font-bold px-2 py-0.5 rounded-full bg-amber-500 text-white animate-pulse"
          >
            {{ pendingProposalsCount }}
          </span>
        </button>
      </nav>
    </div>

    <!-- Bottom Section (Invite Code Card & Collapse Toggle) -->
    <div class="flex flex-col gap-4">
      <!-- Family Invite Card -->
      <div 
        v-if="!isCollapsed && familyStore.family" 
        class="p-3.5 rounded-2xl bg-gradient-to-b from-slate-50 to-slate-100/80 border border-slate-200/70 text-xs flex flex-col gap-2"
      >
        <div class="flex items-center justify-between">
          <span class="text-slate-400 font-semibold uppercase text-[10px] tracking-wider">Kode Keluarga</span>
          <span class="badge badge-xs bg-slate-900 text-white font-mono font-bold">{{ authStore.user?.role }}</span>
        </div>
        <div class="flex items-center justify-between bg-white px-2.5 py-1.5 rounded-xl border border-slate-200 shadow-sm">
          <code class="font-mono font-bold text-slate-800 text-sm tracking-wider">{{ familyStore.family.invite_code }}</code>
          <button 
            @click="copyInviteCode" 
            class="text-brand-600 hover:text-brand-700 p-1 rounded-md transition hover:bg-brand-50"
            title="Salin Kode Invite"
          >
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
              <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
            </svg>
          </button>
        </div>
      </div>

      <!-- Collapse Sidebar Toggle Button -->
      <button 
        @click="emit('toggle-collapse')" 
        class="flex items-center gap-3 px-3 py-2 rounded-xl text-slate-400 hover:text-slate-700 hover:bg-slate-50 transition text-xs font-semibold"
      >
        <svg 
          class="w-4 h-4 transition-transform" 
          :class="isCollapsed ? 'rotate-180' : ''"
          viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
        >
          <polyline points="11 19 4 12 11 5"></polyline>
          <polyline points="18 19 11 12 18 5"></polyline>
        </svg>
        <span v-if="!isCollapsed">Sembunyikan Menu</span>
      </button>
    </div>
  </aside>
</template>
