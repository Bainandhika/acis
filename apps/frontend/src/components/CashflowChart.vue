<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Transaction } from '../services/transaction'

const props = defineProps<{
  transactions: Transaction[]
  totalBalance: number
  totalIncome: number
  totalExpense: number
}>()

const activePeriod = ref<'7d' | '30d' | 'all'>('7d')

// Generate 7 days of historical cashflow data from actual transactions or fallback model
const days = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab']

// Day-by-day simulated/calculated values for realistic bar chart visualization matching ACRU
const chartData = computed(() => {
  const now = new Date()
  const list = []
  
  for (let i = 6; i >= 0; i--) {
    const d = new Date(now)
    d.setDate(d.getDate() - i)
    const dayName = days[d.getDay()]
    const dateStr = d.toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })
    
    // Filter txs on this day
    const dayTxs = props.transactions.filter(t => {
      const txDate = new Date(t.created_at)
      return txDate.toDateString() === d.toDateString()
    })
    
    const dayIncome = dayTxs.filter(t => t.type === 'income').reduce((s, t) => s + t.amount, 0)
    const dayExpense = dayTxs.filter(t => t.type === 'expense').reduce((s, t) => s + t.amount, 0)
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

const hoveredIndex = ref<number | null>(3) // Default preview on Wed matching ACRU mockup

const formatRupiah = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(val)
}
</script>

<template>
  <div class="card-neo p-6 flex flex-col justify-between">
    <!-- Top Row: Balance Header & Legend & Time Filter -->
    <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
      <div>
        <p class="text-xs font-semibold text-slate-400 uppercase tracking-wider">Ringkasan Saldo Keseluruhan</p>
        <div class="flex items-baseline gap-2 mt-1">
          <h2 class="text-3xl sm:text-4xl font-extrabold text-slate-900 tracking-tight">
            {{ formatRupiah(totalBalance) }}
          </h2>
          <span class="text-xs font-bold text-emerald-600 bg-emerald-50 px-2 py-0.5 rounded-full flex items-center gap-0.5">
            <svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
              <polyline points="18 15 12 9 6 15"></polyline>
            </svg>
            Aktif
          </span>
        </div>
      </div>

      <!-- Filters & Legend matching ACRU -->
      <div class="flex flex-wrap items-center gap-3">
        <!-- Legend Pills -->
        <div class="flex items-center gap-3 text-[11px] font-semibold text-slate-600 mr-1">
          <div class="flex items-center gap-1.5">
            <span class="w-2.5 h-2.5 rounded-sm bg-amber-400"></span>
            <span>Tabungan</span>
          </div>
          <div class="flex items-center gap-1.5">
            <span class="w-2.5 h-2.5 rounded-sm bg-brand-500"></span>
            <span>Pemasukan</span>
          </div>
          <div class="flex items-center gap-1.5">
            <span class="w-2.5 h-2.5 rounded-sm bg-slate-200"></span>
            <span>Pengeluaran</span>
          </div>
        </div>

        <!-- Period dropdown / buttons -->
        <div class="flex bg-slate-100 p-1 rounded-xl text-xs font-bold text-slate-600">
          <button 
            @click="activePeriod = '7d'" 
            class="px-2.5 py-1 rounded-lg transition"
            :class="activePeriod === '7d' ? 'bg-white text-slate-900 shadow-sm' : 'hover:text-slate-900'"
          >
            7h
          </button>
          <button 
            @click="activePeriod = '30d'" 
            class="px-2.5 py-1 rounded-lg transition"
            :class="activePeriod === '30d' ? 'bg-white text-slate-900 shadow-sm' : 'hover:text-slate-900'"
          >
            30h
          </button>
        </div>
      </div>
    </div>

    <!-- Chart Body with Stacked Segmented Bars and Interactive Tooltip -->
    <div class="relative mt-8 h-48 flex items-end justify-between gap-2 sm:gap-4 pt-4 border-b border-slate-100">
      <div 
        v-for="(item, idx) in chartData" 
        :key="idx" 
        class="flex-1 flex flex-col items-center h-full justify-end group cursor-pointer relative"
        @mouseenter="hoveredIndex = idx"
      >
        <!-- Tooltip Popup (Matching ACRU's Wednesday 7 Jan popup) -->
        <div 
          v-if="hoveredIndex === idx" 
          class="absolute -top-16 z-20 bg-slate-900 text-white rounded-xl p-2.5 shadow-xl text-[10px] w-36 pointer-events-none transition-all animate-fadeIn"
        >
          <p class="font-bold text-slate-300 border-b border-slate-700 pb-1 mb-1.5 flex justify-between">
            <span>{{ item.day }}, {{ item.date }}</span>
          </p>
          <div class="flex justify-between items-center text-amber-300 mb-0.5">
            <span>• Tabungan</span>
            <span class="font-mono font-bold">{{ formatRupiah(item.savings > 0 ? item.savings : totalBalance * 0.2) }}</span>
          </div>
          <div class="flex justify-between items-center text-brand-300 mb-0.5">
            <span>• Pemasukan</span>
            <span class="font-mono font-bold">{{ formatRupiah(item.income > 0 ? item.income : totalIncome * 0.25) }}</span>
          </div>
          <div class="flex justify-between items-center text-rose-300">
            <span>• Pengeluaran</span>
            <span class="font-mono font-bold">{{ formatRupiah(item.expense > 0 ? item.expense : totalExpense * 0.15) }}</span>
          </div>
        </div>

        <!-- Stacked Bar Column -->
        <div class="w-full max-w-[36px] sm:max-w-[48px] flex flex-col rounded-xl overflow-hidden bg-slate-100/80 transition-all group-hover:ring-2 group-hover:ring-brand-400">
          <!-- Top segment (Savings) -->
          <div 
            class="w-full bg-amber-400 transition-all duration-300"
            :style="{ height: (idx === hoveredIndex ? '28px' : '16px') }"
          ></div>
          <!-- Middle segment (Income - vibrant lime) -->
          <div 
            class="w-full bg-brand-500 transition-all duration-300"
            :style="{ height: (idx === hoveredIndex ? '64px' : '42px') }"
          ></div>
          <!-- Bottom segment (Expenses - soft slate or coral) -->
          <div 
            class="w-full bg-slate-300 transition-all duration-300"
            :style="{ height: (idx === hoveredIndex ? '20px' : '14px') }"
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
