<script setup lang="ts">
import { computed } from 'vue'
import type { Transaction } from '../services/transaction'
import type { Wallet } from '../services/wallet'
import { formatRp } from '../composables/useCurrency'
import { useI18n } from '../locales'

const props = defineProps<{
  transactions: Transaction[]
  wallets: Wallet[]
}>()

const emit = defineEmits<{
  (e: 'filter'): void
  (e: 'export-csv'): void
  (e: 'select-transaction', tx: Transaction): void
}>()

const { t } = useI18n()

const getWalletName = (walletId?: string) => {
  if (!walletId) return 'General'
  const w = props.wallets.find(item => item.id === walletId)
  return w ? w.name : 'General'
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('id-ID', { month: 'short', day: 'numeric' })
}

const formatSignedCurrency = (val: number, type: string) => {
  const formatted = formatRp(Math.abs(val || 0))
  return type === 'income' ? `+${formatted}` : `-${formatted}`
}

const displayTransactions = computed(() => {
  return props.transactions ? props.transactions.slice(0, 10) : []
})
</script>

<template>
  <div class="bg-white rounded-3xl p-6 sm:p-7 border border-slate-100/90 shadow-sm flex flex-col gap-6">
    <!-- Header Controls -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div class="flex items-center gap-2.5">
        <h3 class="text-base sm:text-lg font-bold text-slate-900">
          {{ t('dashboard.recentHistory.title') }}
        </h3>
        <span class="px-2 py-0.5 rounded-full text-xs font-bold bg-slate-100 text-slate-600">
          {{ transactions ? transactions.length : 0 }}
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
          <span>{{ t('extra.filterBtn') }}</span>
        </button>

        <button
          @click="emit('export-csv')"
          class="flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl border border-slate-200 text-xs font-semibold text-slate-700 hover:bg-slate-50 transition active:scale-95 cursor-pointer shadow-2xs"
        >
          <span>{{ t('extra.exportCsv') }}</span>
        </button>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="displayTransactions.length === 0" class="text-center py-8 text-xs text-slate-400 font-medium">
      {{ t('dashboard.recentHistory.noTransactions') }}
    </div>

    <!-- Transactions List Rows -->
    <div v-else class="divide-y divide-slate-100">
      <div
        v-for="tx in displayTransactions"
        :key="tx.id"
        class="py-3.5 flex items-center justify-between gap-4 hover:bg-slate-50/80 px-2 rounded-2xl transition cursor-pointer"
        @click="emit('select-transaction', tx)"
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
          </div>
        </div>

        <!-- Right: Category Tag & Signed Amount -->
        <div class="flex items-center gap-4 sm:gap-8 shrink-0">
          <span class="hidden sm:inline-block px-2.5 py-1 rounded-lg bg-slate-100 text-slate-600 text-[11px] font-semibold">
            {{ getWalletName(tx.wallet_id) }}
          </span>

          <span
            class="text-xs sm:text-sm font-black font-sans text-right min-w-[80px]"
            :class="tx.type === 'income' ? 'text-emerald-600' : 'text-slate-900'"
          >
            {{ formatSignedCurrency(tx.amount, tx.type) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
