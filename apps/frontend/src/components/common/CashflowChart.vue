<script setup lang="ts">
import { computed } from 'vue'
import type { Transaction } from '../../types'
import { formatRp } from '../../utils/format'

const props = defineProps<{
  transactions: Transaction[]
}>()

// Calculate last 7 days metrics
const chartData = computed(() => {
  const days: { label: string; dateStr: string; income: number; expense: number }[] = []
  const today = new Date()

  for (let i = 6; i >= 0; i--) {
    const d = new Date(today)
    d.setDate(d.getDate() - i)
    const dateStr = d.toISOString().split('T')[0] || ''
    const label = d.toLocaleDateString('id-ID', { weekday: 'short', day: 'numeric' })
    days.push({ label, dateStr, income: 0, expense: 0 })
  }

  (props.transactions || []).forEach((t) => {
    if (!t.created_at) return
    const txDateStr = t.created_at.split('T')[0] || ''
    const dayObj = days.find((d) => d.dateStr === txDateStr)
    if (dayObj) {
      if (t.type === 'income') {
        dayObj.income += t.amount || 0
      } else if (t.type === 'expense') {
        dayObj.expense += t.amount || 0
      }
    }
  })

  const maxVal = Math.max(
    ...days.map((d) => Math.max(d.income, d.expense)),
    100000
  )

  return days.map((d) => ({
    ...d,
    incomeHeight: Math.min((d.income / maxVal) * 100, 100),
    expenseHeight: Math.min((d.expense / maxVal) * 100, 100)
  }))
})

const totalIncome = computed(() =>
  (props.transactions || [])
    .filter((t) => t.type === 'income')
    .reduce((sum, t) => sum + (t.amount || 0), 0)
)

const totalExpense = computed(() =>
  (props.transactions || [])
    .filter((t) => t.type === 'expense')
    .reduce((sum, t) => sum + (t.amount || 0), 0)
)
</script>

<template>
  <div class="flex flex-col gap-6">
    <!-- Mini Metric Pills -->
    <div class="grid grid-cols-2 gap-4">
      <div class="p-4 rounded-2xl bg-emerald-950/30 border border-emerald-500/20 flex flex-col">
        <span class="text-[11px] font-bold text-emerald-400 uppercase tracking-wider">Total Income</span>
        <span class="text-xl font-black text-white mt-1">{{ formatRp(totalIncome) }}</span>
      </div>
      <div class="p-4 rounded-2xl bg-rose-950/30 border border-rose-500/20 flex flex-col">
        <span class="text-[11px] font-bold text-rose-400 uppercase tracking-wider">Total Expense</span>
        <span class="text-xl font-black text-white mt-1">{{ formatRp(totalExpense) }}</span>
      </div>
    </div>

    <!-- 7-Day Bar Chart -->
    <div class="h-48 flex items-end justify-between gap-2 sm:gap-4 pt-4 border-b border-slate-800 pb-2">
      <div
        v-for="d in chartData"
        :key="d.dateStr"
        class="flex-1 flex flex-col items-center gap-2 h-full justify-end group relative"
      >
        <!-- Tooltip Hover -->
        <div class="opacity-0 group-hover:opacity-100 transition-opacity absolute -top-12 z-20 bg-slate-900 text-[10px] text-white py-1 px-2 rounded-xl border border-slate-700 shadow-xl pointer-events-none whitespace-nowrap">
          <div class="text-emerald-400 font-bold">+{{ formatRp(d.income) }}</div>
          <div class="text-rose-400 font-bold">-{{ formatRp(d.expense) }}</div>
        </div>

        <!-- Double Bars Container -->
        <div class="w-full flex items-end justify-center gap-1 sm:gap-1.5 h-36">
          <!-- Income Bar -->
          <div
            class="w-2.5 sm:w-4 bg-gradient-to-t from-emerald-600 to-emerald-400 rounded-t-md transition-all duration-500"
            :style="{ height: `${Math.max(d.incomeHeight, 4)}%` }"
            :title="`Income: ${formatRp(d.income)}`"
          ></div>
          <!-- Expense Bar -->
          <div
            class="w-2.5 sm:w-4 bg-gradient-to-t from-rose-600 to-rose-400 rounded-t-md transition-all duration-500"
            :style="{ height: `${Math.max(d.expenseHeight, 4)}%` }"
            :title="`Expense: ${formatRp(d.expense)}`"
          ></div>
        </div>

        <!-- Day Label -->
        <span class="text-[10px] font-bold text-slate-400 truncate w-full text-center">
          {{ d.label }}
        </span>
      </div>
    </div>
  </div>
</template>
