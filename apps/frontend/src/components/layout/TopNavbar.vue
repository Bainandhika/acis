<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import { useFamilyStore } from '../../stores/family'
import { useTransactionStore } from '../../stores/transaction'
import { useUI } from '../../composables/useUI'
import { useI18n } from '../../locales'

const authStore = useAuthStore()
const familyStore = useFamilyStore()
const txStore = useTransactionStore()
const { toggleMobileSidebar } = useUI()
const { t } = useI18n()
const router = useRouter()

const months = [
  'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
  'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
]

const currentYear = new Date().getFullYear()
const years = Array.from({ length: 5 }, (_, i) => currentYear - 2 + i)

const pendingProposalsCount = computed(() => {
  return (txStore.proposals || []).filter((p) => p.status === 'pending').length
})

const handlePeriodChange = async (month: number, year: number) => {
  txStore.selectedMonth = month
  txStore.selectedYear = year
  await txStore.fetchTransactions(year, month)
}
</script>

<template>
  <header class="sticky top-0 z-20 bg-slate-900/80 backdrop-blur-md border-b border-slate-800/80 px-4 sm:px-8 py-3.5 flex items-center justify-between gap-4">
    <!-- Left: Mobile Menu Button & Family Name -->
    <div class="flex items-center gap-3">
      <!-- Mobile hamburger toggle -->
      <button
        @click="toggleMobileSidebar"
        class="md:hidden p-2 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800 transition cursor-pointer"
        aria-label="Toggle navigation menu"
      >
        <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="3" y1="12" x2="21" y2="12"></line>
          <line x1="3" y1="6" x2="21" y2="6"></line>
          <line x1="3" y1="18" x2="21" y2="18"></line>
        </svg>
      </button>

      <div class="flex flex-col">
        <h2 class="text-sm font-extrabold text-white tracking-tight flex items-center gap-2">
          <span>{{ familyStore.family?.name || 'ACIS Family' }}</span>
        </h2>
        <span class="text-[11px] text-slate-400 font-medium hidden sm:inline">
          {{ t('header.subtitle') || 'Sistem Keuangan Keluarga' }}
        </span>
      </div>
    </div>

    <!-- Right: Period Selector, Proposals Alert, User Badge -->
    <div class="flex items-center gap-3">
      <!-- Month & Year Selector -->
      <div class="flex items-center gap-1.5 bg-slate-950/80 p-1 rounded-2xl border border-slate-800/90 text-xs">
        <select
          :value="txStore.selectedMonth"
          @change="handlePeriodChange(Number(($event.target as HTMLSelectElement).value), txStore.selectedYear)"
          class="bg-transparent text-slate-200 font-bold px-2 py-1 rounded-xl focus:outline-none cursor-pointer border-none"
        >
          <option
            v-for="(mName, idx) in months"
            :key="idx"
            :value="idx + 1"
            class="bg-slate-900 text-white"
          >
            {{ mName }}
          </option>
        </select>

        <select
          :value="txStore.selectedYear"
          @change="handlePeriodChange(txStore.selectedMonth, Number(($event.target as HTMLSelectElement).value))"
          class="bg-transparent text-slate-200 font-bold px-2 py-1 rounded-xl focus:outline-none cursor-pointer border-none"
        >
          <option
            v-for="y in years"
            :key="y"
            :value="y"
            class="bg-slate-900 text-white"
          >
            {{ y }}
          </option>
        </select>
      </div>

      <!-- Pending Proposals Badge Button -->
      <button
        v-if="pendingProposalsCount > 0"
        @click="router.push('/proposals')"
        class="relative flex items-center justify-center p-2 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-400 hover:bg-amber-500/20 transition cursor-pointer"
        :title="`${pendingProposalsCount} pending proposals`"
      >
        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
          <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
        </svg>
        <span class="absolute -top-1 -right-1 w-4 h-4 bg-amber-500 text-slate-950 font-black text-[9px] rounded-full flex items-center justify-center">
          {{ pendingProposalsCount }}
        </span>
      </button>

      <!-- User Role Pill -->
      <div class="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-2xl bg-slate-950/80 border border-slate-800/80">
        <div class="w-6 h-6 rounded-full bg-teal-500/20 text-teal-300 font-black text-xs flex items-center justify-center border border-teal-500/30">
          {{ (authStore.user?.username || 'U').charAt(0).toUpperCase() }}
        </div>
        <span class="text-xs font-bold text-slate-200">{{ authStore.user?.username || 'User' }}</span>
        <span class="text-[9px] uppercase font-extrabold px-1.5 py-0.5 rounded-full bg-teal-500/20 text-teal-300 border border-teal-500/30">
          {{ authStore.user?.role || 'member' }}
        </span>
      </div>
    </div>
  </header>
</template>
