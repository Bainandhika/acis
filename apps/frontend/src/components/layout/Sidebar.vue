<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import { useFamilyStore } from '../../stores/family'
import { useTransactionStore } from '../../stores/transaction'
import { useUI } from '../../composables/useUI'
import { useI18n } from '../../locales'

const authStore = useAuthStore()
const familyStore = useFamilyStore()
const txStore = useTransactionStore()
const { isMobileSidebarOpen, setMobileSidebar, isSidebarCollapsed, toggleSidebarCollapse } = useUI()
const { t, locale, setLocale } = useI18n()
const router = useRouter()
const route = useRoute()

const pendingProposalsCount = computed(() => {
  return (txStore.proposals || []).filter((p) => p.status === 'pending').length
})

const showTelegramModal = ref(false)

const openTelegramModal = () => {
  showTelegramModal.value = true
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

const navItems = computed(() => {
  const items = [
    {
      to: '/',
      exact: true,
      name: 'dashboard',
      label: t('nav.dashboard'),
      icon: 'dashboard'
    },
    {
      to: '/wallets',
      exact: false,
      name: 'wallets',
      label: t('nav.wallets'),
      icon: 'wallets'
    },
    {
      to: '/transactions',
      exact: false,
      name: 'transactions',
      label: t('nav.transactions'),
      icon: 'transactions'
    },
    {
      to: '/proposals',
      exact: false,
      name: 'proposals',
      label: t('proposals.title') || 'Proposals',
      icon: 'proposals',
      badge: pendingProposalsCount.value
    }
  ]

  if (authStore.user?.role === 'admin') {
    items.push({
      to: '/family',
      exact: false,
      name: 'family',
      label: t('nav.settings') || 'Family Settings',
      icon: 'family',
      badge: 0
    })
  }

  return items
})

const isActive = (item: { to: string; exact?: boolean }) => {
  if (item.exact) {
    return route.path === item.to
  }
  return route.path.startsWith(item.to)
}
</script>

<template>
  <div>
    <!-- Mobile Backdrop -->
    <div
      v-if="isMobileSidebarOpen"
      @click="setMobileSidebar(false)"
      class="fixed inset-0 bg-slate-950/70 backdrop-blur-xs z-40 md:hidden"
    ></div>

    <!-- Sidebar Container -->
    <aside
      class="fixed md:sticky top-0 left-0 h-screen bg-slate-900 text-slate-300 py-6 px-4 flex flex-col justify-between transition-all duration-300 z-50 shrink-0 border-r border-slate-800"
      :class="[
        isSidebarCollapsed ? 'md:w-20' : 'md:w-64',
        isMobileSidebarOpen ? 'translate-x-0 w-64' : '-translate-x-full md:translate-x-0'
      ]"
    >
      <!-- Brand & Top Menu -->
      <div class="flex flex-col gap-6">
        <!-- Logo Header -->
        <div class="flex items-center gap-3 px-2">
          <img
            src="/logo.png"
            alt="ACIS Logo"
            class="w-10 h-10 rounded-2xl object-cover shrink-0 shadow-lg shadow-teal-950/40 border border-teal-500/30"
          />
          <div v-if="!isSidebarCollapsed || isMobileSidebarOpen" class="flex flex-col min-w-0">
            <div class="flex items-center gap-1.5">
              <span class="font-extrabold text-xl tracking-tight text-white font-sans">ACIS</span>
              <span class="text-[9px] uppercase font-bold tracking-wider px-1.5 py-0.5 rounded-full bg-teal-500/20 text-teal-300 border border-teal-500/30">
                {{ (authStore.user?.role || 'MEMBER').toUpperCase() }}
              </span>
            </div>
            <span class="text-xs text-slate-400 font-medium truncate">
              {{ familyStore.family?.name || t('extra.defaultFamily') }}
            </span>
          </div>
        </div>

        <!-- Navigation Links -->
        <nav class="flex flex-col gap-1.5">
          <router-link
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            @click="setMobileSidebar(false)"
            class="flex items-center justify-between px-3.5 py-3 rounded-2xl text-xs font-bold transition-all group"
            :class="[
              isActive(item)
                ? 'bg-slate-800 text-white shadow-sm ring-1 ring-slate-700/60'
                : 'text-slate-400 hover:text-slate-100 hover:bg-slate-800/50'
            ]"
          >
            <div class="flex items-center gap-3.5 min-w-0">
              <div class="w-5 h-5 flex items-center justify-center shrink-0">
                <!-- Dashboard Icon -->
                <svg
                  v-if="item.icon === 'dashboard'"
                  class="w-5 h-5 transition-colors"
                  :class="isActive(item) ? 'text-teal-400' : 'text-slate-400 group-hover:text-slate-200'"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <rect x="3" y="3" width="7" height="7" rx="2"></rect>
                  <rect x="14" y="3" width="7" height="7" rx="2"></rect>
                  <rect x="14" y="14" width="7" height="7" rx="2"></rect>
                  <rect x="3" y="14" width="7" height="7" rx="2"></rect>
                </svg>

                <!-- Wallets Icon -->
                <svg
                  v-else-if="item.icon === 'wallets'"
                  class="w-5 h-5 transition-colors"
                  :class="isActive(item) ? 'text-teal-400' : 'text-slate-400 group-hover:text-slate-200'"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <path d="M20 12V8H6a2 2 0 0 1-2-2c0-1.1.9-2 2-2h12v4"></path>
                  <path d="M4 6v12c0 1.1.9 2 2 2h14v-4"></path>
                  <circle cx="18" cy="12" r="2"></circle>
                </svg>

                <!-- Transactions Icon -->
                <svg
                  v-else-if="item.icon === 'transactions'"
                  class="w-5 h-5 transition-colors"
                  :class="isActive(item) ? 'text-teal-400' : 'text-slate-400 group-hover:text-slate-200'"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <line x1="12" y1="1" x2="12" y2="23"></line>
                  <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
                </svg>

                <!-- Proposals Icon -->
                <svg
                  v-else-if="item.icon === 'proposals'"
                  class="w-5 h-5 transition-colors"
                  :class="isActive(item) ? 'text-teal-400' : 'text-slate-400 group-hover:text-slate-200'"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                  <polyline points="14 2 14 8 20 8"></polyline>
                  <line x1="16" y1="13" x2="8" y2="13"></line>
                  <line x1="16" y1="17" x2="8" y2="17"></line>
                  <polyline points="10 9 9 9 8 9"></polyline>
                </svg>

                <!-- Family / Settings Icon -->
                <svg
                  v-else
                  class="w-5 h-5 transition-colors"
                  :class="isActive(item) ? 'text-teal-400' : 'text-slate-400 group-hover:text-slate-200'"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <circle cx="12" cy="12" r="3"></circle>
                  <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
                </svg>
              </div>
              <span v-if="!isSidebarCollapsed || isMobileSidebarOpen">{{ item.label }}</span>
            </div>

            <!-- Badge -->
            <span
              v-if="item.badge && item.badge > 0 && (!isSidebarCollapsed || isMobileSidebarOpen)"
              class="px-2 py-0.5 text-[10px] font-black rounded-full bg-amber-500 text-slate-950 shrink-0"
            >
              {{ item.badge }}
            </span>
          </router-link>
        </nav>
      </div>

      <!-- Bottom Section: Language, Telegram Status, Logout -->
      <div class="flex flex-col gap-3 pt-4 border-t border-slate-800/80">
        <!-- Language Selector & Logout -->
        <div v-if="!isSidebarCollapsed || isMobileSidebarOpen" class="flex items-center justify-between px-1">
          <div class="inline-flex p-1 bg-slate-950 rounded-xl border border-slate-800 text-[11px] font-bold">
            <button
              @click="setLocale('en')"
              class="px-2.5 py-1 rounded-lg transition-all cursor-pointer"
              :class="locale === 'en' ? 'bg-slate-800 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200'"
              type="button"
            >
              EN
            </button>
            <button
              @click="setLocale('id')"
              class="px-2.5 py-1 rounded-lg transition-all cursor-pointer"
              :class="locale === 'id' ? 'bg-slate-800 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200'"
              type="button"
            >
              ID
            </button>
          </div>

          <button
            @click="handleLogout"
            class="text-xs font-semibold text-rose-400 hover:text-rose-300 hover:bg-rose-950/40 transition cursor-pointer px-2.5 py-1 rounded-xl border border-rose-500/30 hover:border-rose-500/50 bg-rose-950/20 shadow-sm"
            :title="t('nav.signOut')"
          >
            {{ t('nav.signOut') }}
          </button>
        </div>

        <!-- Telegram Connection Status -->
        <button
          v-if="!isSidebarCollapsed || isMobileSidebarOpen"
          type="button"
          @click="openTelegramModal"
          class="w-full flex items-center justify-between px-3 py-2 rounded-xl transition-all border text-left cursor-pointer"
          :class="[
            familyStore.family?.telegram_chat_id
              ? 'bg-slate-950/60 border-slate-800 text-slate-400'
              : 'bg-teal-950/30 border-teal-500/30 text-teal-300 hover:bg-teal-900/40 hover:border-teal-500/50 group'
          ]"
        >
          <div class="flex items-center gap-2.5 min-w-0">
            <span class="relative flex h-2 w-2 shrink-0">
              <span
                v-if="familyStore.family?.telegram_chat_id"
                class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"
              ></span>
              <span
                class="relative inline-flex rounded-full h-2 w-2"
                :class="familyStore.family?.telegram_chat_id ? 'bg-emerald-500' : 'bg-amber-400 animate-pulse'"
              ></span>
            </span>
            <div class="flex flex-col min-w-0">
              <span class="text-xs font-medium truncate">
                {{ familyStore.family?.telegram_chat_id ? t('extra.telegramConnected') : t('extra.telegramNotConnected') }}
              </span>
            </div>
          </div>
        </button>

        <!-- Collapsed Sidebar Telegram icon -->
        <button
          v-else
          @click="openTelegramModal"
          type="button"
          class="flex items-center justify-center p-2 rounded-xl border transition-all cursor-pointer"
          :class="[
            familyStore.family?.telegram_chat_id
              ? 'bg-slate-950/60 border-slate-800 text-emerald-400'
              : 'bg-teal-950/30 border-teal-500/30 text-amber-400'
          ]"
          :title="familyStore.family?.telegram_chat_id ? t('extra.telegramConnected') : t('extra.telegramNotConnected')"
        >
          <span class="relative flex h-2.5 w-2.5">
            <span
              v-if="familyStore.family?.telegram_chat_id"
              class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"
            ></span>
            <span
              class="relative inline-flex rounded-full h-2.5 w-2.5"
              :class="familyStore.family?.telegram_chat_id ? 'bg-emerald-500' : 'bg-amber-400'"
            ></span>
          </span>
        </button>

        <!-- Desktop Collapse Button -->
        <button
          @click="toggleSidebarCollapse"
          class="hidden md:flex items-center justify-center p-2 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800/60 transition cursor-pointer"
          :title="isSidebarCollapsed ? 'Expand Sidebar' : 'Collapse Sidebar'"
        >
          <svg
            class="w-4 h-4 transition-transform duration-200"
            :class="isSidebarCollapsed ? 'rotate-180' : ''"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <polyline points="11 19 4 12 11 5"></polyline>
            <polyline points="18 19 11 12 18 5"></polyline>
          </svg>
        </button>
      </div>

      <!-- Telegram Link Modal Dialog -->
      <Teleport to="body">
        <dialog :open="showTelegramModal" :class="showTelegramModal ? 'modal modal-open' : 'modal'" class="z-50">
          <div class="modal-box bg-slate-900 text-slate-100 rounded-3xl p-6 shadow-2xl border border-slate-800 max-w-md">
            <div class="flex items-center justify-between mb-2">
              <h3 class="font-black text-lg text-white">{{ t('modals.telegram.title') }}</h3>
              <button
                @click="showTelegramModal = false"
                class="btn btn-sm btn-ghost btn-circle text-slate-400 hover:text-white"
                type="button"
              >
                ✕
              </button>
            </div>
            <p class="text-xs text-slate-400 mb-4">{{ t('modals.telegram.subtitle') }}</p>
            <div class="flex flex-col gap-2.5 text-xs text-slate-300 bg-slate-950 p-4 rounded-2xl border border-slate-800">
              <p class="font-bold text-white">{{ t('modals.telegram.howToLink') }}</p>
              <p>1. {{ t('modals.telegram.step1') }}</p>
              <p>
                2. {{ t('modals.telegram.step2') }}
                <span class="font-mono font-bold bg-slate-900 px-2 py-0.5 rounded border border-slate-700 text-teal-400 select-all">
                  /link {{ familyStore.family?.invite_code }}
                </span>
              </p>
              <p>3. {{ t('modals.telegram.step3') }}</p>
            </div>
            <div class="modal-action mt-6">
              <button
                class="btn bg-teal-700 hover:bg-teal-800 text-white btn-sm rounded-xl text-xs font-bold border-none"
                @click="showTelegramModal = false"
                type="button"
              >
                {{ t('modals.telegram.close') }}
              </button>
            </div>
          </div>
          <form method="dialog" class="modal-backdrop bg-slate-950/60 backdrop-blur-xs">
            <button type="button" @click="showTelegramModal = false">close</button>
          </form>
        </dialog>
      </Teleport>
    </aside>
  </div>
</template>
