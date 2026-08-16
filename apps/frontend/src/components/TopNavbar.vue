<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useFamilyStore } from '../stores/family'
import { useWalletStore } from '../stores/wallet'
import { useTransactionStore } from '../stores/transaction'
import { useRouter } from 'vue-router'
import { useI18n } from '../locales'

export type NavTab = 'dashboard' | 'wallets' | 'transactions' | 'reports' | 'submissions'

const props = defineProps<{
  activeTab: NavTab
}>()

const emit = defineEmits<{
  (e: 'select-tab', tab: NavTab): void
  (e: 'open-telegram-modal'): void
  (e: 'open-family-manage'): void
}>()

const authStore = useAuthStore()
const familyStore = useFamilyStore()
const walletStore = useWalletStore()
const txStore = useTransactionStore()
const router = useRouter()
const { t, locale, setLocale } = useI18n()

const isUserMenuOpen = ref(false)
const isMobileMenuOpen = ref(false)
const isCodeCopied = ref(false)

onMounted(async () => {
  if (!familyStore.family) {
    await familyStore.fetchMyFamily()
  }
})

const copyInviteCode = async () => {
  const code = familyStore.family?.invite_code
  if (!code) return
  try {
    await navigator.clipboard.writeText(code)
    isCodeCopied.value = true
    setTimeout(() => {
      isCodeCopied.value = false
    }, 2000)
  } catch {
    // Clipboard fallback
  }
}

const handleLogout = async () => {
  if (confirm(t('nav.signOutConfirm'))) {
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
      
      <!-- Left: Logo & Brand: ACIS with Subtitle -->
      <div class="flex items-center gap-5">
        <div class="flex items-center gap-3">
          <img 
            src="/logo.png" 
            alt="ACIS Logo" 
            class="w-10 h-10 rounded-xl shadow-md shadow-brand-500/20 object-cover shrink-0"
          />
          
          <div class="flex flex-col">
            <span class="font-black text-xl tracking-tight text-slate-900 leading-tight">
              ACIS
            </span>
            <span class="text-[11px] text-slate-400 font-semibold truncate max-w-[220px]">
              {{ t('brand.subtagline') }}
            </span>
          </div>
        </div>

        <!-- Desktop Horizontal Top Navigation Bar -->
        <nav class="hidden xl:flex items-center gap-1 p-1 bg-slate-100/80 rounded-2xl border border-slate-200/60 ml-1">
          <button
            @click="emit('select-tab', 'dashboard')"
            class="px-3 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-1.5 whitespace-nowrap shrink-0"
            :class="activeTab === 'dashboard' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:text-slate-900 hover:bg-white/70'"
          >
            <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="7" height="7" rx="2"></rect>
              <rect x="14" y="3" width="7" height="7" rx="2"></rect>
              <rect x="14" y="14" width="7" height="7" rx="2"></rect>
              <rect x="3" y="14" width="7" height="7" rx="2"></rect>
            </svg>
            <span class="whitespace-nowrap">{{ t('nav.dashboard') }}</span>
          </button>

          <button
            @click="emit('select-tab', 'wallets')"
            class="px-3 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-1.5 whitespace-nowrap shrink-0"
            :class="activeTab === 'wallets' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:text-slate-900 hover:bg-white/70'"
          >
            <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="2" y="5" width="20" height="14" rx="3"></rect>
              <line x1="2" y1="10" x2="22" y2="10"></line>
            </svg>
            <span class="whitespace-nowrap">{{ t('nav.wallets') }}</span>
          </button>

          <button
            @click="emit('select-tab', 'transactions')"
            class="px-3 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-1.5 whitespace-nowrap shrink-0"
            :class="activeTab === 'transactions' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:text-slate-900 hover:bg-white/70'"
          >
            <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="17 1 21 5 17 9"></polyline>
              <path d="M3 11V9a4 4 0 0 1 4-4h14"></path>
              <polyline points="7 23 3 19 7 15"></polyline>
              <path d="M21 13v2a4 4 0 0 1-4 4H3"></path>
            </svg>
            <span class="whitespace-nowrap">{{ t('nav.transactions') }}</span>
          </button>

          <button
            @click="emit('select-tab', 'reports')"
            class="px-3 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-1.5 whitespace-nowrap shrink-0"
            :class="activeTab === 'reports' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:text-slate-900 hover:bg-white/70'"
          >
            <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 3v18h18"></path>
              <path d="M18 17V9"></path>
              <path d="M13 17V5"></path>
              <path d="M8 17v-3"></path>
            </svg>
            <span class="whitespace-nowrap">{{ t('nav.reports') }}</span>
          </button>

          <button
            @click="emit('select-tab', 'submissions')"
            class="px-3 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-1.5 whitespace-nowrap shrink-0"
            :class="activeTab === 'submissions' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:text-slate-900 hover:bg-white/70'"
          >
            <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 5v14"></path>
              <path d="M5 12h14"></path>
            </svg>
            <span class="whitespace-nowrap">{{ t('nav.submissions') }}</span>
          </button>
        </nav>
      </div>

      <!-- Right: Language Switcher, Telegram Status, User Profile Pill, Logout CTA, Mobile Toggle -->
      <div class="flex items-center gap-2.5">
        <!-- Language Switcher Selector -->
        <div class="flex items-center p-1 bg-slate-100/90 rounded-xl border border-slate-200/70 text-[11px] font-bold">
          <button 
            @click="setLocale('en')"
            class="px-2 py-0.5 rounded-lg transition-all"
            :class="locale === 'en' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-800'"
            type="button"
          >
            EN
          </button>
          <button 
            @click="setLocale('id')"
            class="px-2 py-0.5 rounded-lg transition-all"
            :class="locale === 'id' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-800'"
            type="button"
          >
            ID
          </button>
        </div>

        <!-- Telegram Bot Indicator -->
        <button
          @click="emit('open-telegram-modal')"
          class="hidden sm:flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-semibold border transition cursor-pointer hover:opacity-90"
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
            {{ familyStore.family?.telegram_chat_id ? t('nav.botConnected') : t('nav.botOffline') }}
          </span>
        </button>

        <!-- Interactive User Profile Box with Popup Dropdown Menu -->
        <div class="relative">
          <button
            @click="isUserMenuOpen = !isUserMenuOpen"
            type="button"
            class="flex items-center gap-2 pl-2 py-1 pr-3 rounded-2xl bg-slate-50 hover:bg-slate-100/80 border border-slate-200/80 transition cursor-pointer active:scale-95 text-left"
            :class="isUserMenuOpen ? 'ring-2 ring-brand-400/50 bg-slate-100' : ''"
            :title="t('nav.userMenu')"
          >
            <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-slate-900 to-slate-700 text-white font-black text-xs flex items-center justify-center shrink-0 shadow-sm">
              {{ (authStore.user?.username || authStore.user?.name || 'U').charAt(0).toUpperCase() }}
            </div>
            <div class="hidden md:flex flex-col text-left leading-tight">
              <div class="flex items-center gap-1.5">
                <span class="text-xs font-bold text-slate-900 truncate max-w-[100px]">{{ authStore.user?.username || authStore.user?.name || 'User' }}</span>
                <span class="text-[9px] font-extrabold uppercase px-1 py-0.2 rounded bg-brand-200 text-brand-900">{{ authStore.user?.role }}</span>
              </div>
              <span class="text-[10px] text-slate-400 truncate max-w-[110px]">{{ authStore.user?.phone_number }}</span>
            </div>
            <svg 
              class="w-3.5 h-3.5 text-slate-400 ml-0.5 transition-transform duration-200" 
              :class="isUserMenuOpen ? 'rotate-180' : ''" 
              viewBox="0 0 24 24" 
              fill="none" 
              stroke="currentColor" 
              stroke-width="2.5"
            >
              <polyline points="6 9 12 15 18 9"></polyline>
            </svg>
          </button>

          <!-- Backdrop to close dropdown on click outside -->
          <div 
            v-if="isUserMenuOpen" 
            class="fixed inset-0 z-40 bg-transparent" 
            @click="isUserMenuOpen = false"
          ></div>

          <!-- User Dropdown Popup -->
          <div 
            v-if="isUserMenuOpen" 
            class="absolute right-0 top-full mt-2 w-60 bg-white/95 backdrop-blur-md rounded-2xl shadow-xl border border-slate-200/80 p-2 z-50 flex flex-col gap-1 animate-in fade-in zoom-in-95 duration-100"
          >
            <!-- User Header Summary inside dropdown -->
            <div class="px-3 py-2.5 rounded-xl bg-slate-50 border border-slate-100 flex items-center gap-2.5 mb-1">
              <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-slate-900 to-slate-700 text-white font-black text-xs flex items-center justify-center shrink-0">
                {{ (authStore.user?.username || authStore.user?.name || 'U').charAt(0).toUpperCase() }}
              </div>
              <div class="flex flex-col min-w-0">
                <span class="text-xs font-black text-slate-900 truncate">{{ authStore.user?.username || authStore.user?.name || 'User' }}</span>
                <span class="text-[10px] text-slate-400 font-medium truncate">{{ authStore.user?.phone_number }}</span>
              </div>
            </div>

            <!-- Option 1: Edit Workspace / Family -->
            <button
              @click="emit('open-family-manage'); isUserMenuOpen = false"
              class="w-full px-3 py-2.5 rounded-xl text-xs font-bold text-slate-700 hover:text-slate-900 hover:bg-slate-100/80 transition flex items-center gap-2.5 text-left"
              type="button"
            >
              <svg class="w-4 h-4 text-brand-600 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 20h9"></path>
                <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path>
              </svg>
              <span>{{ t('nav.editProfile') }}</span>
            </button>

            <div class="h-px bg-slate-100 my-0.5"></div>

            <!-- Option 2: Sign Out Pop up -->
            <button
              @click="handleLogout(); isUserMenuOpen = false"
              class="w-full px-3 py-2.5 rounded-xl text-xs font-bold text-rose-600 hover:bg-rose-50 transition flex items-center gap-2.5 text-left"
              type="button"
            >
              <svg class="w-4 h-4 text-rose-500 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
                <polyline points="16 17 21 12 16 7"></polyline>
                <line x1="21" y1="12" x2="9" y2="12"></line>
              </svg>
              <span>{{ t('nav.signOut') }}</span>
            </button>
          </div>
        </div>

        <!-- Mobile Menu Toggle with Content-Width Popup Dropdown (No full-page overlay) -->
        <div class="relative xl:hidden">
          <button 
            @click="isMobileMenuOpen = !isMobileMenuOpen"
            class="p-2 rounded-xl text-slate-600 hover:bg-slate-100 transition"
            aria-label="Toggle navigation menu"
          >
            <svg class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="3" y1="12" x2="21" y2="12"></line>
              <line x1="3" y1="6" x2="21" y2="6"></line>
              <line x1="3" y1="18" x2="21" y2="18"></line>
            </svg>
          </button>

          <!-- Backdrop to close dropdown on click outside -->
          <div 
            v-if="isMobileMenuOpen" 
            class="fixed inset-0 z-40 bg-transparent" 
            @click="isMobileMenuOpen = false"
          ></div>

          <!-- Content-Width Popup Menu -->
          <div 
            v-if="isMobileMenuOpen" 
            class="absolute right-0 top-full mt-2 w-56 bg-white/95 backdrop-blur-md rounded-2xl shadow-xl border border-slate-200/80 p-2 z-50 flex flex-col gap-1"
          >
            <button
              @click="emit('select-tab', 'dashboard'); isMobileMenuOpen = false"
              class="px-3.5 py-2.5 rounded-xl text-xs font-bold text-left flex items-center gap-2.5 transition"
              :class="activeTab === 'dashboard' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'"
            >
              <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="3" width="7" height="7" rx="2"></rect>
                <rect x="14" y="3" width="7" height="7" rx="2"></rect>
                <rect x="14" y="14" width="7" height="7" rx="2"></rect>
                <rect x="3" y="14" width="7" height="7" rx="2"></rect>
              </svg>
              <span>{{ t('nav.dashboard') }}</span>
            </button>
            <button
              @click="emit('select-tab', 'wallets'); isMobileMenuOpen = false"
              class="px-3.5 py-2.5 rounded-xl text-xs font-bold text-left flex items-center gap-2.5 transition"
              :class="activeTab === 'wallets' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'"
            >
              <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="2" y="5" width="20" height="14" rx="3"></rect>
                <line x1="2" y1="10" x2="22" y2="10"></line>
              </svg>
              <span>{{ t('nav.wallets') }}</span>
            </button>
            <button
              @click="emit('select-tab', 'transactions'); isMobileMenuOpen = false"
              class="px-3.5 py-2.5 rounded-xl text-xs font-bold text-left flex items-center gap-2.5 transition"
              :class="activeTab === 'transactions' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'"
            >
              <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="17 1 21 5 17 9"></polyline>
                <path d="M3 11V9a4 4 0 0 1 4-4h14"></path>
                <polyline points="7 23 3 19 7 15"></polyline>
                <path d="M21 13v2a4 4 0 0 1-4 4H3"></path>
              </svg>
              <span>{{ t('nav.transactions') }}</span>
            </button>
            <button
              @click="emit('select-tab', 'reports'); isMobileMenuOpen = false"
              class="px-3.5 py-2.5 rounded-xl text-xs font-bold text-left flex items-center gap-2.5 transition"
              :class="activeTab === 'reports' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'"
            >
              <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M3 3v18h18"></path>
                <path d="M18 17V9"></path>
                <path d="M13 17V5"></path>
                <path d="M8 17v-3"></path>
              </svg>
              <span>{{ t('nav.reports') }}</span>
            </button>
            <button
              @click="emit('select-tab', 'submissions'); isMobileMenuOpen = false"
              class="px-3.5 py-2.5 rounded-xl text-xs font-bold text-left flex items-center gap-2.5 transition"
              :class="activeTab === 'submissions' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'"
            >
              <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 5v14"></path>
                <path d="M5 12h14"></path>
              </svg>
              <span>{{ t('nav.submissions') }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>
