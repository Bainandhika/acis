<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useFamilyStore } from '../stores/family'
import { useWalletStore } from '../stores/wallet'
import { useTransactionStore } from '../stores/transaction'
import { useRouter } from 'vue-router'

export type NavTab = 'dashboard' | 'wallets' | 'transactions' | 'reports' | 'submissions'

const props = defineProps<{
  activeTab: NavTab
}>()

const emit = defineEmits<{
  (e: 'select-tab', tab: NavTab): void
  (e: 'open-telegram-modal'): void
}>()

const authStore = useAuthStore()
const familyStore = useFamilyStore()
const walletStore = useWalletStore()
const txStore = useTransactionStore()
const router = useRouter()

const isMobileMenuOpen = ref(false)

const handleLogout = async () => {
  if (confirm('Are you sure you want to sign out?')) {
    await authStore.logout()
    familyStore.resetState()
    walletStore.resetState()
    txStore.resetState()
    router.push('/login')
  }
}
</script>

<template>
  <header class="bg-white/90 backdrop-blur-md border-b border-slate-200/80 sticky top-0 z-40 px-4 sm:px-6 lg:px-8">
    <div class="max-w-[1600px] mx-auto flex items-center justify-between h-20 gap-4">
      
      <!-- Left: Logo & Brand + Family Name Badge -->
      <div class="flex items-center gap-6">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-2xl bg-gradient-to-tr from-brand-500 to-lime-300 flex items-center justify-center shadow-md shadow-brand-500/20 text-white font-black text-xl shrink-0">
            A
          </div>
          <div class="flex flex-col">
            <div class="flex items-center gap-1.5 leading-tight">
              <span class="font-black text-xl tracking-tight text-slate-900">ACIS</span>
              <span class="text-[9px] uppercase font-bold tracking-wider px-1.5 py-0.5 rounded-full bg-brand-100 text-brand-800">Family</span>
            </div>
            <span class="text-xs text-slate-400 font-semibold truncate max-w-[140px]">
              {{ familyStore.family?.name || 'My Family' }}
            </span>
          </div>
        </div>

        <!-- Desktop Horizontal Top Navigation Bar -->
        <nav class="hidden lg:flex items-center gap-1 p-1 bg-slate-100/80 rounded-2xl border border-slate-200/60">
          <button
            @click="emit('select-tab', 'dashboard')"
            class="px-3.5 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2"
            :class="activeTab === 'dashboard' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:text-slate-900 hover:bg-white/70'"
          >
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="7" height="7" rx="2"></rect>
              <rect x="14" y="3" width="7" height="7" rx="2"></rect>
              <rect x="14" y="14" width="7" height="7" rx="2"></rect>
              <rect x="3" y="14" width="7" height="7" rx="2"></rect>
            </svg>
            <span>Dashboard</span>
          </button>

          <button
            @click="emit('select-tab', 'wallets')"
            class="px-3.5 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2"
            :class="activeTab === 'wallets' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:text-slate-900 hover:bg-white/70'"
          >
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="2" y="5" width="20" height="14" rx="3"></rect>
              <line x1="2" y1="10" x2="22" y2="10"></line>
            </svg>
            <span>Wallets</span>
          </button>

          <button
            @click="emit('select-tab', 'transactions')"
            class="px-3.5 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2"
            :class="activeTab === 'transactions' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:text-slate-900 hover:bg-white/70'"
          >
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="17 1 21 5 17 9"></polyline>
              <path d="M3 11V9a4 4 0 0 1 4-4h14"></path>
              <polyline points="7 23 3 19 7 15"></polyline>
              <path d="M21 13v2a4 4 0 0 1-4 4H3"></path>
            </svg>
            <span>Transaction History</span>
          </button>

          <button
            @click="emit('select-tab', 'reports')"
            class="px-3.5 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2"
            :class="activeTab === 'reports' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:text-slate-900 hover:bg-white/70'"
          >
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 3v18h18"></path>
              <path d="M18 17V9"></path>
              <path d="M13 17V5"></path>
              <path d="M8 17v-3"></path>
            </svg>
            <span>Financial Reports</span>
          </button>

          <button
            @click="emit('select-tab', 'submissions')"
            class="px-3.5 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2"
            :class="activeTab === 'submissions' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:text-slate-900 hover:bg-white/70'"
          >
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 5v14"></path>
              <path d="M5 12h14"></path>
            </svg>
            <span>Transaction Submission</span>
          </button>
        </nav>
      </div>

      <!-- Right: Telegram Bot Status, User Profile Pill, Logout CTA, Mobile Toggle -->
      <div class="flex items-center gap-3">
        <!-- Telegram Bot Indicator -->
        <button
          @click="emit('open-telegram-modal')"
          class="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-semibold border transition cursor-pointer hover:opacity-90"
          :class="familyStore.family?.telegram_chat_id 
            ? 'bg-emerald-50 border-emerald-200 text-emerald-700' 
            : 'bg-slate-100 border-slate-200 text-slate-600'"
          title="Telegram Bot Integration"
        >
          <span 
            class="w-2 h-2 rounded-full"
            :class="familyStore.family?.telegram_chat_id ? 'bg-emerald-500 animate-pulse' : 'bg-slate-400'"
          ></span>
          <span class="text-[11px] font-bold">
            Bot {{ familyStore.family?.telegram_chat_id ? 'Connected' : 'Offline' }}
          </span>
        </button>

        <!-- User Profile Pill -->
        <div class="flex items-center gap-2.5 pl-2 py-1 pr-3 rounded-2xl bg-slate-50 border border-slate-200/80">
          <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-slate-900 to-slate-700 text-white font-black text-xs flex items-center justify-center shrink-0">
            {{ authStore.user?.name ? authStore.user.name.charAt(0).toUpperCase() : 'U' }}
          </div>
          <div class="hidden md:flex flex-col text-left leading-tight">
            <div class="flex items-center gap-1.5">
              <span class="text-xs font-bold text-slate-900 truncate max-w-[100px]">{{ authStore.user?.name || 'User' }}</span>
              <span class="text-[9px] font-extrabold uppercase px-1 py-0.2 rounded bg-brand-200 text-brand-900">{{ authStore.user?.role }}</span>
            </div>
            <span class="text-[10px] text-slate-400 truncate max-w-[110px]">{{ authStore.user?.email }}</span>
          </div>
        </div>

        <!-- Logout Button -->
        <button 
          @click="handleLogout"
          class="flex items-center gap-1.5 px-3 py-2 rounded-2xl bg-slate-100 hover:bg-rose-50 text-slate-600 hover:text-rose-600 font-bold text-xs transition border border-slate-200/70"
          title="Sign Out"
        >
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
            <polyline points="16 17 21 12 16 7"></polyline>
            <line x1="21" y1="12" x2="9" y2="12"></line>
          </svg>
          <span class="hidden sm:inline">Sign Out</span>
        </button>

        <!-- Mobile Menu Toggle -->
        <button 
          @click="isMobileMenuOpen = !isMobileMenuOpen"
          class="lg:hidden p-2 rounded-xl text-slate-600 hover:bg-slate-100 transition"
        >
          <svg class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="3" y1="12" x2="21" y2="12"></line>
            <line x1="3" y1="6" x2="21" y2="6"></line>
            <line x1="3" y1="18" x2="21" y2="18"></line>
          </svg>
        </button>
      </div>
    </div>

    <!-- Mobile Dropdown Navigation Menu -->
    <div v-if="isMobileMenuOpen" class="lg:hidden border-t border-slate-100 py-3 flex flex-col gap-1">
      <button
        @click="emit('select-tab', 'dashboard'); isMobileMenuOpen = false"
        class="px-4 py-2.5 rounded-xl text-xs font-bold text-left flex items-center gap-2"
        :class="activeTab === 'dashboard' ? 'bg-slate-900 text-white' : 'text-slate-600 hover:bg-slate-50'"
      >
        <span>Dashboard</span>
      </button>
      <button
        @click="emit('select-tab', 'wallets'); isMobileMenuOpen = false"
        class="px-4 py-2.5 rounded-xl text-xs font-bold text-left flex items-center gap-2"
        :class="activeTab === 'wallets' ? 'bg-slate-900 text-white' : 'text-slate-600 hover:bg-slate-50'"
      >
        <span>Wallets</span>
      </button>
      <button
        @click="emit('select-tab', 'transactions'); isMobileMenuOpen = false"
        class="px-4 py-2.5 rounded-xl text-xs font-bold text-left flex items-center gap-2"
        :class="activeTab === 'transactions' ? 'bg-slate-900 text-white' : 'text-slate-600 hover:bg-slate-50'"
      >
        <span>Transaction History</span>
      </button>
      <button
        @click="emit('select-tab', 'reports'); isMobileMenuOpen = false"
        class="px-4 py-2.5 rounded-xl text-xs font-bold text-left flex items-center gap-2"
        :class="activeTab === 'reports' ? 'bg-slate-900 text-white' : 'text-slate-600 hover:bg-slate-50'"
      >
        <span>Financial Reports</span>
      </button>
      <button
        @click="emit('select-tab', 'submissions'); isMobileMenuOpen = false"
        class="px-4 py-2.5 rounded-xl text-xs font-bold text-left flex items-center gap-2"
        :class="activeTab === 'submissions' ? 'bg-slate-900 text-white' : 'text-slate-600 hover:bg-slate-50'"
      >
        <span>Transaction Submission</span>
      </button>
    </div>
  </header>
</template>
