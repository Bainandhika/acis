<script setup lang="ts">
import { computed } from 'vue'
import type { Transaction } from '../services/transaction'
import type { Wallet } from '../services/wallet'

const props = defineProps<{
  transactions: Transaction[]
  wallets: Wallet[]
}>()

const emit = defineEmits<{
  (e: 'filter'): void
  (e: 'export-csv'): void
  (e: 'select-transaction', tx: Transaction): void
}>()

const getWalletName = (walletId?: string) => {
  if (!walletId) return 'General'
  const w = props.wallets.find(item => item.id === walletId)
  return w ? w.name : 'General'
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return 'May 18'
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}

const formatCurrency = (val: number, type: string) => {
  const formatted = new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(Math.abs(val || 0))

  return type === 'income' ? `+${formatted}` : `-${formatted}`
}

// Fallback demo transactions if store is empty
const displayTransactions = computed(() => {
  if (props.transactions && props.transactions.length > 0) {
    return props.transactions.slice(0, 6)
  }
  return [
    { id: '1', description: 'Whole Foods Market', author: 'Sarah (Mom)', wallet_name: 'Groceries', amount: 164.50, type: 'expense', created_at: '2026-05-18T10:00:00Z' },
    { id: '2', description: 'Monthly Rent Deposit', author: 'David (Dad)', wallet_name: 'Rent & Housing', amount: 2400.00, type: 'expense', created_at: '2026-05-17T10:00:00Z' },
    { id: '3', description: 'Weekly Allowance', author: 'Emma (Daughter)', wallet_name: 'Kids Allowance', amount: 40.00, type: 'expense', created_at: '2026-05-15T10:00:00Z' },
    { id: '4', description: 'Interest Payoff Reward', author: 'System Interest', wallet_name: 'Emergency Fund', amount: 25.00, type: 'income', created_at: '2026-05-14T10:00:00Z' },
    { id: '5', description: 'Kumon Math Center', author: 'Sarah (Mom)', wallet_name: 'Kids Education', amount: 180.00, type: 'expense', created_at: '2026-05-12T10:00:00Z' },
    { id: '6', description: 'State Park Parking', author: 'Leo (Son)', wallet_name: 'Vacation 2026', amount: 15.00, type: 'expense', created_at: '2026-05-10T10:00:00Z' },
  ]
})
</script>

<template>
  <div class="bg-white rounded-3xl p-6 sm:p-7 border border-slate-100/90 shadow-sm flex flex-col gap-6">
    <!-- Header Controls -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div class="flex items-center gap-2.5">
        <h3 class="text-base sm:text-lg font-bold text-slate-900">
          Recent Transactions
        </h3>
        <span class="px-2 py-0.5 rounded-full text-xs font-bold bg-slate-100 text-slate-600">
          {{ transactions.length > 0 ? transactions.length : '6' }}
        </span>
      </div>

      <!-- Action Buttons: Filter & Export CSV -->
      <div class="flex items-center gap-2.5">
        <button
          @click="emit('filter')"
          class="flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl border border-slate-200 text-xs font-semibold text-slate-700 hover:bg-slate-50 transition active:scale-95 cursor-pointer shadow-2xs"
        >
          <svg class="w-3.5 h-3.5 text-slate-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"></polygon>
          </svg>
          <span>Filter</span>
        </button>

        <button
          @click="emit('export-csv')"
          class="flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl border border-slate-200 text-xs font-semibold text-slate-700 hover:bg-slate-50 transition active:scale-95 cursor-pointer shadow-2xs"
        >
          <span>Export CSV</span>
        </button>
      </div>
    </div>

    <!-- Transactions List Rows -->
    <div class="divide-y divide-slate-100">
      <div
        v-for="tx in displayTransactions"
        :key="tx.id"
        class="py-3.5 flex items-center justify-between gap-4 hover:bg-slate-50/80 px-2 rounded-2xl transition cursor-pointer"
        @click="emit('select-transaction', tx as any)"
      >
        <!-- Left: Date & Description -->
        <div class="flex items-center gap-4 sm:gap-6 min-w-0">
          <span class="text-xs font-semibold text-slate-400 w-14 shrink-0">
            {{ formatDate(tx.created_at) }}
          </span>

          <div class="flex flex-col min-w-0">
            <span class="text-xs sm:text-sm font-bold text-slate-900 truncate">
              {{ tx.description || 'Transaction' }}
            </span>
            <span class="text-[11px] text-slate-400 font-medium truncate">
              by {{ (tx as any).author || 'Member' }}
            </span>
          </div>
        </div>

        <!-- Right: Category Tag & Signed Amount -->
        <div class="flex items-center gap-4 sm:gap-8 shrink-0">
          <span class="hidden sm:inline-block px-2.5 py-1 rounded-lg bg-slate-100 text-slate-600 text-[11px] font-semibold">
            {{ (tx as any).wallet_name || getWalletName((tx as any).wallet_id) }}
          </span>

          <span
            class="text-xs sm:text-sm font-black font-sans text-right min-w-[80px]"
            :class="tx.type === 'income' ? 'text-emerald-600' : 'text-slate-900'"
          >
            {{ formatCurrency(tx.amount, tx.type) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
