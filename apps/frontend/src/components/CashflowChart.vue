<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Transaction } from '../services/transaction'
import { useI18n } from '../locales'

const props = defineProps<{
  transactions: Transaction[]
  totalBalance: number
  totalIncome: number
  totalExpense: number
}>()

const { t, locale } = useI18n()

// Generate 7 days of historical cashflow data from actual transactions strictly (Mon - Sun / last 7 days)
const dayLabelsEn = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
const dayLabelsId = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab']

// Day-by-day calculation from real transactions
const chartData = computed(() => {
  const now = new Date()
  const list = []
  const labels = locale.value === 'id' ? dayLabelsId : dayLabelsEn
  
  for (let i = 6; i >= 0; i--) {
    const d = new Date(now)
    d.setDate(d.getDate() - i)
    const dayName = labels[d.getDay()]
    const dateStr = d.toLocaleDateString(locale.value === 'id' ? 'id-ID' : 'en-US', { day: 'numeric', month: 'short' })
    
    // Filter txs on this specific date
    const dayTxs = (props.transactions || []).filter(t => {
      if (!t?.created_at) return false
      const txDate = new Date(t.created_at)
      return txDate.toDateString() === d.toDateString()
    })
    
    const dayIncome = dayTxs.filter(t => t.type === 'income').reduce((s, t) => s + (t.amount || 0), 0)
    const dayExpense = dayTxs.filter(t => t.type === 'expense').reduce((s, t) => s + (t.amount || 0), 0)
    const daySavings = Math.max(0, dayIncome - dayExpense)
    
    list.push({
      day: dayName,
      date: dateStr,
      income: dayIncome,
      expense: dayExpense,
      savings: daySavings,
      hasData: dayTxs.length > 0
    })
  }
  return list
})

// Tooltip / hover state: null by default (strictly disabled when cursor is outside the chart)
const hoveredIndex = ref<number | null>(null)

// Max height computation for clean proportional bar rendering
const maxAmountInPeriod = computed(() => {
  let max = 0
  chartData.value.forEach(d => {
    if (d.income > max) max = d.income
    if (d.expense > max) max = d.expense
    if (d.savings > max) max = d.savings
  })
  return max > 0 ? max : 1
})

const getBarHeight = (amount: number, isHovered: boolean) => {
  if (amount <= 0) return '0px'
  const maxHeightPx = 80
  const ratio = Math.min(amount / maxAmountInPeriod.value, 1)
  const basePx = Math.max(Math.round(ratio * maxHeightPx), 6)
  return `${isHovered ? basePx + 4 : basePx}px`
}

// Rupiah Currency Formatter (e.g. Rp 1.500.000)
const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(val || 0)
}
</script>

<template>
  <div class="card-neo p-6 flex flex-col justify-between" id="seven-day-cashflow-card">
    <!-- Top Row: Balance Header & Legend -->
    <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
      <div>
        <p class="text-xs font-semibold text-slate-400 uppercase tracking-wider">{{ t('dashboard.financialSummary.title') }}</p>
        <div class="flex items-baseline gap-2 mt-1">
          <h2 class="text-2xl sm:text-3xl font-extrabold text-slate-900 tracking-tight font-mono">
            {{ formatCurrency(totalBalance) }}
          </h2>
          <span class="text-xs font-bold text-emerald-600 bg-emerald-50 px-2.5 py-0.5 rounded-full flex items-center gap-1 border border-emerald-200">
            <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
            {{ t('dashboard.financialSummary.sevenDaysView') }}
          </span>
        </div>
      </div>

      <!-- Legend Pills -->
      <div class="flex items-center gap-3 text-[11px] font-semibold text-slate-600">
        <div class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-sm bg-amber-400"></span>
          <span>{{ t('dashboard.financialSummary.savings') }}</span>
        </div>
        <div class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-sm bg-brand-500"></span>
          <span>{{ t('dashboard.financialSummary.income') }}</span>
        </div>
        <div class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-sm bg-slate-300"></span>
          <span>{{ t('dashboard.financialSummary.expense') }}</span>
        </div>
      </div>
    </div>

    <!-- Chart Body with Stacked Segmented Bars and Interactive Tooltip (Strictly disabled on mouseleave) -->
    <div 
      class="relative mt-8 h-48 flex items-end justify-between gap-2 sm:gap-4 pt-4 border-b border-slate-100"
      @mouseleave="hoveredIndex = null"
    >
      <div 
        v-for="(item, idx) in chartData" 
        :key="idx" 
        class="flex-1 flex flex-col items-center h-full justify-end group cursor-pointer relative"
        @mouseenter="hoveredIndex = idx"
      >
        <!-- Tooltip Popup: ONLY rendered when hoveredIndex === idx -->
        <div 
          v-if="hoveredIndex === idx" 
          class="absolute -top-20 z-30 bg-slate-900 text-white rounded-xl p-2.5 shadow-xl text-[10px] w-44 pointer-events-none transition-all animate-fadeIn"
        >
          <p class="font-bold text-slate-300 border-b border-slate-700 pb-1 mb-1.5 flex justify-between">
            <span>{{ item.day }}, {{ item.date }}</span>
          </p>
          <div class="flex justify-between items-center text-amber-300 mb-0.5">
            <span>• {{ t('dashboard.financialSummary.savings') }}</span>
            <span class="font-mono font-bold">{{ formatCurrency(item.savings) }}</span>
          </div>
          <div class="flex justify-between items-center text-brand-300 mb-0.5">
            <span>• {{ t('dashboard.financialSummary.income') }}</span>
            <span class="font-mono font-bold">{{ formatCurrency(item.income) }}</span>
          </div>
          <div class="flex justify-between items-center text-rose-300">
            <span>• {{ t('dashboard.financialSummary.expense') }}</span>
            <span class="font-mono font-bold">{{ formatCurrency(item.expense) }}</span>
          </div>
        </div>

        <!-- Stacked Bar Column with actual proportional height -->
        <div class="w-full max-w-[36px] sm:max-w-[48px] flex flex-col rounded-xl overflow-hidden bg-slate-100/80 transition-all group-hover:ring-2 group-hover:ring-brand-400">
          <!-- Top segment (Savings) -->
          <div 
            v-if="item.savings > 0"
            class="w-full bg-amber-400 transition-all duration-300"
            :style="{ height: getBarHeight(item.savings, idx === hoveredIndex) }"
          ></div>
          <!-- Middle segment (Income) -->
          <div 
            v-if="item.income > 0"
            class="w-full bg-brand-500 transition-all duration-300"
            :style="{ height: getBarHeight(item.income, idx === hoveredIndex) }"
          ></div>
          <!-- Bottom segment (Expenses) -->
          <div 
            v-if="item.expense > 0"
            class="w-full bg-slate-300 transition-all duration-300"
            :style="{ height: getBarHeight(item.expense, idx === hoveredIndex) }"
          ></div>
          <!-- Empty Base Indicator when 0 -->
          <div 
            v-if="!item.hasData"
            class="w-full h-1.5 bg-slate-200"
          ></div>
        </div>

        <!-- Day Label -->
        <span 
          class="text-[11px] font-semibold mt-3 transition-colors"
          :class="hoveredIndex === idx ? 'text-slate-900 font-bold' : 'text-slate-400'"
        >
          {{ item.day }}
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}
.animate-fadeIn {
  animation: fadeIn 0.15s ease-out forwards;
}
</style>
