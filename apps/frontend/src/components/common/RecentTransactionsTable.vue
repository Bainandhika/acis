<script setup lang="ts">
import { computed } from 'vue'
import type { Transaction, Wallet } from '../../types'
import { formatRp, formatDateShort } from '../../utils/format'
import { useI18n } from '../../locales'

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
  if (!walletId) return 'Saldo Utama'
  const w = props.wallets.find((item) => item.id === walletId)
  return w ? w.name : 'General'
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
  <div class="card-neo bg-slate-900/90 rounded-3xl p-6 sm:p-7 border border-slate-800/90 shadow-card flex flex-col gap-6">
    <!-- Header Controls -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div class="flex items-center gap-2.5">
        <h3 class="text-base sm:text-lg font-bold text-white">
          {{ t('dashboard.recentHistory.title') || 'Riwayat Transaksi Terkini' }}
        </h3>
        <span class="px-2.5 py-0.5 rounded-full text-xs font-bold bg-slate-800 text-teal-300 border border-slate-700">
          {{ transactions ? transactions.length : 0 }}
        </span>
      </div>

      <!-- Action Buttons: Filter & Export CSV -->
      <div class="flex items-center gap-2.5">
        <button
          @click="emit('filter')"
          class="flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl border border-slate-700 bg-slate-800/60 text-xs font-semibold text-slate-300 hover:text-white hover:bg-slate-800 transition active:scale-95 cursor-pointer"
        >
          <svg class="w-3.5 h-3.5 text-slate-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"></polygon>
          </svg>
          <span>{{ t('extra.filterBtn') || 'Lihat Semua' }}</span>
        </button>

        <button
          @click="emit('export-csv')"
          class="flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl border border-slate-700 bg-slate-800/60 text-xs font-semibold text-slate-300 hover:text-white hover:bg-slate-800 transition active:scale-95 cursor-pointer"
        >
          <span>{{ t('extra.exportCsv') || 'Ekspor CSV' }}</span>
        </button>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="displayTransactions.length === 0" class="text-center py-8 text-xs text-slate-500 font-medium">
      {{ t('dashboard.recentHistory.noTransactions') || 'Belum ada transaksi di periode ini' }}
    </div>

    <!-- Transactions List Rows -->
    <div v-else class="divide-y divide-slate-800/80">
      <div
        v-for="tx in displayTransactions"
        :key="tx.id"
        class="py-3.5 flex items-center justify-between gap-4 hover:bg-slate-800/40 px-2 rounded-2xl transition cursor-pointer"
        @click="emit('select-transaction', tx)"
      >
        <!-- Left: Date & Description -->
        <div class="flex items-center gap-4 sm:gap-6 min-w-0">
          <span class="text-xs font-semibold text-slate-400 w-14 shrink-0 font-mono">
            {{ formatDateShort(tx.created_at) }}
          </span>

          <div class="flex flex-col min-w-0">
            <span class="text-xs sm:text-sm font-bold text-slate-100 truncate">
              {{ tx.description || 'Transaksi' }}
            </span>
            <span class="text-[11px] text-slate-400 sm:hidden">
              {{ getWalletName(tx.wallet_id) }}
            </span>
          </div>
        </div>

        <!-- Right: Category Tag & Signed Amount -->
        <div class="flex items-center gap-4 sm:gap-8 shrink-0">
          <span class="hidden sm:inline-block px-2.5 py-1 rounded-lg bg-slate-800 text-slate-300 text-[11px] font-semibold border border-slate-700/60">
            {{ getWalletName(tx.wallet_id) }}
          </span>

          <span
            class="text-xs sm:text-sm font-black font-sans text-right min-w-[90px]"
            :class="tx.type === 'income' ? 'text-emerald-400' : 'text-slate-100'"
          >
            {{ formatSignedCurrency(tx.amount, tx.type) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
