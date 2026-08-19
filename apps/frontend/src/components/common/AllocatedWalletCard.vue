<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Wallet } from '../../types'
import { formatRp } from '../../utils/format'
import { useI18n } from '../../locales'

const props = defineProps<{
  wallet: Wallet
  spent: number
  limit: number
  isAdmin: boolean
}>()

const emit = defineEmits<{
  (e: 'edit', wallet: Wallet): void
  (e: 'delete', wallet: Wallet): void
  (e: 'allocate', wallet: Wallet): void
}>()

const { t } = useI18n()
const isMenuOpen = ref(false)

// Percentage Calculation
const percentageUsed = computed(() => {
  if (props.limit <= 0) return 0
  return Math.round((props.spent / props.limit) * 100)
})

// Progress bar width clamped to 100%
const progressWidth = computed(() => {
  return Math.min(Math.max(percentageUsed.value, 0), 100)
})

// Is Over Budget
const isOverBudget = computed(() => {
  return props.spent > props.limit && props.limit > 0
})

// Remaining or Over amount
const remainingAmount = computed(() => {
  return props.limit - props.spent
})

// Category Icon & Color mapping based on name / state
const categoryMeta = computed(() => {
  const name = (props.wallet.name || '').toLowerCase()

  if (name.includes('grocer') || name.includes('makan') || name.includes('belanja')) {
    return {
      icon: 'cart',
      bgClass: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
    }
  } else if (name.includes('rent') || name.includes('hous') || name.includes('rumah') || name.includes('tempat')) {
    return {
      icon: 'home',
      bgClass: 'bg-teal-500/10 text-teal-400 border-teal-500/20'
    }
  } else if (name.includes('educat') || name.includes('kid') || name.includes('sekolah') || name.includes('anak')) {
    return {
      icon: 'book',
      bgClass: 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20'
    }
  } else if (name.includes('emerg') || name.includes('darurat') || name.includes('tabung')) {
    return {
      icon: 'shield',
      bgClass: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
    }
  } else if (name.includes('vacat') || name.includes('holiday') || name.includes('liburan')) {
    return {
      icon: 'plane',
      bgClass: 'bg-orange-500/10 text-orange-400 border-orange-500/20'
    }
  } else {
    return {
      icon: 'zap',
      bgClass: isOverBudget.value
        ? 'bg-rose-500/10 text-rose-400 border-rose-500/20'
        : 'bg-teal-500/10 text-teal-400 border-teal-500/20'
    }
  }
})

// Progress Bar Color based on thresholds (<80% Teal, 80-99% Orange, >=100% Rose)
const progressBarColor = computed(() => {
  if (isOverBudget.value || percentageUsed.value >= 100) {
    return 'bg-rose-500'
  }
  if (percentageUsed.value >= 80) {
    return 'bg-orange-500'
  }
  return 'bg-teal-500'
})
</script>

<template>
  <div
    class="card-neo bg-slate-900/90 rounded-3xl p-5 sm:p-6 border border-slate-800/90 shadow-card flex flex-col justify-between transition-all hover:border-slate-700 relative"
    :class="{ 'ring-1 ring-rose-500/50': isOverBudget }"
  >
    <!-- Top Row: Icon + Name + Context Menu -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3 min-w-0">
        <!-- Category Icon Box -->
        <div class="w-10 h-10 rounded-2xl flex items-center justify-center border shrink-0" :class="categoryMeta.bgClass">
          <!-- Cart Icon -->
          <svg v-if="categoryMeta.icon === 'cart'" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="9" cy="21" r="1"></circle>
            <circle cx="20" cy="21" r="1"></circle>
            <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"></path>
          </svg>
          <!-- Home Icon -->
          <svg v-else-if="categoryMeta.icon === 'home'" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
            <polyline points="9 22 9 12 15 12 15 22"></polyline>
          </svg>
          <!-- Book Icon -->
          <svg v-else-if="categoryMeta.icon === 'book'" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path>
            <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path>
          </svg>
          <!-- Shield Icon -->
          <svg v-else-if="categoryMeta.icon === 'shield'" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
          </svg>
          <!-- Plane Icon -->
          <svg v-else-if="categoryMeta.icon === 'plane'" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M17.8 19.2L16 11l3.5-3.5C21 6 21.5 4 21 3c-1-.5-3 0-4.5 1.5L13 8 4.8 6.2c-.5-.1-.9.1-1.1.5l-.3.5c-.2.5-.1 1 .3 1.3L9 12l-2 3H4l-1 1 3 2 2 3 1-1v-3l3-2 3.5 5.3c.3.4.8.5 1.3.3l.5-.3c.4-.2.6-.6.5-1.1z"></path>
          </svg>
          <!-- Default Zap Icon -->
          <svg v-else class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon>
          </svg>
        </div>

        <div class="flex flex-col min-w-0">
          <h4 class="font-extrabold text-sm text-white truncate">
            {{ wallet.name }}
          </h4>
          <span class="text-[11px] text-slate-400 truncate">
            {{ wallet.description || 'Virtual Wallet' }}
          </span>
        </div>
      </div>

      <!-- 3-Dots Action Menu -->
      <div v-if="isAdmin" class="relative">
        <button
          @click="isMenuOpen = !isMenuOpen"
          class="w-8 h-8 flex items-center justify-center rounded-xl text-slate-400 hover:text-white hover:bg-slate-800 transition cursor-pointer"
        >
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="1"></circle>
            <circle cx="19" cy="12" r="1"></circle>
            <circle cx="5" cy="12" r="1"></circle>
          </svg>
        </button>

        <!-- Dropdown Context Menu -->
        <div
          v-if="isMenuOpen"
          class="absolute right-0 mt-1 w-36 bg-slate-900 border border-slate-700/80 rounded-2xl shadow-2xl py-1.5 z-30"
          @mouseleave="isMenuOpen = false"
        >
          <button
            @click="emit('allocate', wallet); isMenuOpen = false"
            class="w-full text-left px-3.5 py-1.5 text-xs text-slate-300 hover:text-white hover:bg-slate-800 font-medium cursor-pointer"
          >
            {{ t('extra.contextAllocate') || 'Alokasi Dana' }}
          </button>
          <button
            @click="emit('edit', wallet); isMenuOpen = false"
            class="w-full text-left px-3.5 py-1.5 text-xs text-slate-300 hover:text-white hover:bg-slate-800 font-medium cursor-pointer"
          >
            {{ t('extra.contextEdit') || 'Edit Dompet' }}
          </button>
          <button
            @click="emit('delete', wallet); isMenuOpen = false"
            class="w-full text-left px-3.5 py-1.5 text-xs text-rose-400 hover:bg-rose-950/40 font-medium cursor-pointer"
          >
            {{ t('extra.contextDelete') || 'Hapus Dompet' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Balance & Spent Numbers -->
    <div class="flex items-baseline justify-between mt-5 mb-2">
      <div>
        <span class="text-[10px] font-bold uppercase tracking-wider text-slate-400 block">{{ t('extra.spent') || 'Terpakai' }}</span>
        <span class="text-xl font-extrabold text-white font-sans">
          {{ formatRp(spent) }}
        </span>
      </div>
      <div class="text-right">
        <span class="text-[10px] font-bold uppercase tracking-wider text-slate-400 block">{{ t('extra.limit') || 'Saldo / Limit' }}</span>
        <span class="text-xs font-bold text-teal-400 font-sans">
          {{ formatRp(wallet.current_balance || limit) }}
        </span>
      </div>
    </div>

    <!-- Progress Bar -->
    <div class="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden my-1">
      <div
        class="h-full rounded-full transition-all duration-500"
        :class="progressBarColor"
        :style="{ width: `${progressWidth}%` }"
      ></div>
    </div>

    <!-- Subtext Details -->
    <div class="flex items-center justify-between text-xs mt-2 font-medium">
      <span class="text-slate-400 font-semibold">
        {{ percentageUsed }}% {{ t('extra.used') || 'terpakai' }}
      </span>

      <!-- Right Status (Left or Over) -->
      <span
        v-if="isOverBudget"
        class="text-rose-400 font-bold"
      >
        {{ formatRp(Math.abs(remainingAmount)) }} {{ t('extra.over') || 'melebihi' }}
      </span>
      <span
        v-else-if="remainingAmount === 0"
        class="text-slate-400 font-bold"
      >
        {{ t('extra.zeroLeft') || 'Habis' }}
      </span>
      <span
        v-else
        class="text-emerald-400 font-bold"
      >
        {{ formatRp(remainingAmount) }} {{ t('extra.left') || 'tersisa' }}
      </span>
    </div>
  </div>
</template>
