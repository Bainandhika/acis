<script setup lang="ts">
import { computed } from 'vue'
import { formatRp } from '../../utils/format'
import { useI18n } from '../../locales'

const props = defineProps<{
  totalFunds: number
  trendPercentage?: number
  isAdmin: boolean
}>()

const emit = defineEmits<{
  (e: 'quick-allocate'): void
  (e: 'transfer-money'): void
}>()

const { t } = useI18n()

const formattedBalance = computed(() => {
  return formatRp(props.totalFunds)
})

const trendDisplay = computed(() => {
  const p = props.trendPercentage ?? 0
  return p >= 0 ? `+${p.toFixed(1)}%` : `${p.toFixed(1)}%`
})
</script>

<template>
  <div class="card-neo bg-gradient-to-br from-slate-900 via-slate-900 to-teal-950/40 p-6 sm:p-7 rounded-3xl border border-slate-800/90 shadow-card flex flex-col justify-between h-full min-h-[220px]">
    <!-- Top Info -->
    <div class="flex flex-col">
      <span class="text-[11px] font-bold uppercase tracking-wider text-teal-400 font-sans">
        {{ t('extra.totalFundsLabel') || 'Total Dana Keluarga' }}
      </span>
      <div class="flex flex-wrap items-baseline gap-3 mt-3">
        <h2 class="text-3xl sm:text-4xl font-black text-white tracking-tight font-sans">
          {{ formattedBalance }}
        </h2>
        <!-- Trend Badge -->
        <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
          <svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="18 15 12 9 6 15"></polyline>
          </svg>
          <span>{{ trendDisplay }} {{ t('extra.vsLastMonth') || 'bulan lalu' }}</span>
        </span>
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="flex items-center gap-3 mt-6">
      <!-- Quick Allocate Button (Teal) -->
      <button
        v-if="isAdmin"
        @click="emit('quick-allocate')"
        class="flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl bg-teal-700 hover:bg-teal-800 text-white text-xs font-bold transition shadow-sm hover:shadow active:scale-95 cursor-pointer border border-teal-600/30"
        type="button"
      >
        <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <line x1="12" y1="5" x2="12" y2="19"></line>
          <line x1="5" y1="12" x2="19" y2="12"></line>
        </svg>
        <span>{{ t('extra.quickAllocate') || 'Alokasi Dana' }}</span>
      </button>

      <!-- Transfer Money Button -->
      <button
        @click="emit('transfer-money')"
        class="flex items-center justify-center px-4 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold transition active:scale-95 cursor-pointer border border-slate-700"
        type="button"
      >
        <span>{{ t('extra.transferBtn') || 'Catat Transaksi' }}</span>
      </button>
    </div>
  </div>
</template>
