<script setup lang="ts">
import { computed } from 'vue'
import { formatRp } from '../composables/useCurrency'
import { useI18n } from '../locales'

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
  return formatRp(props.totalFunds, 0)
})

const trendDisplay = computed(() => {
  const p = props.trendPercentage ?? 0
  return p >= 0 ? `+${p.toFixed(1)}%` : `${p.toFixed(1)}%`
})
</script>

<template>
  <div class="bg-white rounded-3xl p-6 sm:p-7 border border-slate-100/90 shadow-sm flex flex-col justify-between h-full min-h-[220px]">
    <!-- Top Info -->
    <div class="flex flex-col">
      <span class="text-[11px] font-bold uppercase tracking-wider text-slate-400 font-sans">
        {{ t('extra.totalFundsLabel') }}
      </span>
      <div class="flex flex-wrap items-baseline gap-3 mt-3">
        <h2 class="text-3xl sm:text-4xl font-black text-slate-900 tracking-tight font-sans">
          {{ formattedBalance }}
        </h2>
        <!-- Trend Badge -->
        <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-600 border border-emerald-100/80">
          <svg class="w-3 h-3 text-emerald-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="18 15 12 9 6 15"></polyline>
          </svg>
          <span>{{ trendDisplay }} {{ t('extra.vsLastMonth') }}</span>
        </span>
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="flex items-center gap-3 mt-6">
      <!-- Quick Allocate Button (Dark Teal) -->
      <button
        v-if="isAdmin"
        @click="emit('quick-allocate')"
        class="flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl bg-teal-700 hover:bg-teal-800 text-white text-xs font-bold transition shadow-sm hover:shadow active:scale-95 cursor-pointer"
        type="button"
      >
        <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <line x1="12" y1="5" x2="12" y2="19"></line>
          <line x1="5" y1="12" x2="19" y2="12"></line>
        </svg>
        <span>{{ t('extra.quickAllocate') }}</span>
      </button>

      <!-- Transfer Money Button (Subtle Neutral Pill) -->
      <button
        @click="emit('transfer-money')"
        class="flex items-center justify-center px-4 py-2.5 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-semibold transition active:scale-95 cursor-pointer"
        type="button"
      >
        <span>{{ t('extra.transferBtn') }}</span>
      </button>
    </div>
  </div>
</template>
