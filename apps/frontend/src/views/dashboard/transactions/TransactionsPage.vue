<script setup lang="ts">
import { ref, computed } from 'vue'
import { useTransactions } from '../../../composables/useTransactions'
import { useWallets } from '../../../composables/useWallets'
import { useAuthStore } from '../../../stores/auth'
import { formatRp, formatDate } from '../../../utils/format'
import { useI18n } from '../../../locales'
import type { Transaction } from '../../../types'

// Components
import TransactionFilters from './components/TransactionFilters.vue'
import Button from '../../../components/ui/Button.vue'
import Skeleton from '../../../components/ui/Skeleton.vue'

// Modals
import CreateTransactionModal from './components/CreateTransactionModal.vue'
import EditTransactionModal from './components/EditTransactionModal.vue'
import DeleteTransactionModal from './components/DeleteTransactionModal.vue'
import ChangeRequestModal from './components/ChangeRequestModal.vue'
import DeleteRequestModal from './components/DeleteRequestModal.vue'

const {
  filteredTransactions,
  filters,
  isLoading,
  setFilters,
  clearFilters,
  exportCsv
} = useTransactions()

const { wallets } = useWallets()
const authStore = useAuthStore()
const { t } = useI18n()

const isAdmin = computed(() => authStore.user?.role === 'admin')

const isCreateModalOpen = ref(false)
const isEditModalOpen = ref(false)
const isDeleteModalOpen = ref(false)
const isChangeRequestOpen = ref(false)
const isDeleteRequestOpen = ref(false)
const selectedTx = ref<Transaction | null>(null)

const getWalletName = (walletId?: string) => {
  if (!walletId) return 'Saldo Utama'
  const w = wallets.value.find((item) => item.id === walletId)
  return w ? w.name : 'General'
}

const openEdit = (tx: Transaction) => {
  selectedTx.value = tx
  isEditModalOpen.value = true
}

const openDelete = (tx: Transaction) => {
  selectedTx.value = tx
  isDeleteModalOpen.value = true
}

const openChangeRequest = (tx: Transaction) => {
  selectedTx.value = tx
  isChangeRequestOpen.value = true
}

const openDeleteRequest = (tx: Transaction) => {
  selectedTx.value = tx
  isDeleteRequestOpen.value = true
}

const handleExport = () => {
  exportCsv(getWalletName)
}
</script>

<template>
  <div class="flex flex-col gap-6 pb-12">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
      <div>
        <h1 class="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
          {{ t('transaksiTab.title') || 'Daftar Transaksi' }}
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          {{ t('transaksiTab.subtitle') || 'Riwayat lengkap pencatatan arus kas dan pengeluaran' }}
        </p>
      </div>

      <div class="flex items-center gap-3">
        <Button
          variant="secondary"
          size="sm"
          @click="handleExport"
        >
          {{ t('extra.exportCsv') || 'Ekspor CSV' }}
        </Button>
        <Button
          variant="primary"
          size="sm"
          @click="isCreateModalOpen = true"
        >
          + {{ isAdmin ? (t('transaksiTab.recordBtn') || 'Catat Transaksi') : (t('extra.sendProposalBtn') || 'Ajukan Pengeluaran') }}
        </Button>
      </div>
    </div>

    <!-- Filter Bar -->
    <TransactionFilters
      :filters="filters"
      @update="setFilters"
      @clear="clearFilters"
    />

    <!-- Transactions Data Card -->
    <div class="card-neo bg-slate-900/90 rounded-3xl p-6 border border-slate-800/90 shadow-card overflow-hidden">
      <!-- Loading Skeleton -->
      <div v-if="isLoading" class="space-y-3">
        <Skeleton v-for="i in 5" :key="i" type="table-row" />
      </div>

      <!-- Empty State -->
      <div v-else-if="filteredTransactions.length === 0" class="text-center py-16 flex flex-col items-center gap-3">
        <div class="w-12 h-12 rounded-2xl bg-slate-800 flex items-center justify-center text-xl text-slate-400">
          🔍
        </div>
        <h4 class="text-sm font-bold text-white">Tidak ada transaksi ditemukan</h4>
        <p class="text-xs text-slate-500 max-w-xs">
          Coba sesuaikan filter pencarian atau tanggal untuk melihat data transaksi lainnya.
        </p>
        <Button variant="ghost" size="xs" @click="clearFilters">
          Reset Semua Filter
        </Button>
      </div>

      <!-- Transactions Table -->
      <div v-else class="overflow-x-auto">
        <table class="w-full text-left">
          <thead>
            <tr class="border-b border-slate-800 text-[11px] font-bold text-slate-400 uppercase tracking-wider">
              <th class="pb-3 px-3">Tanggal</th>
              <th class="pb-3 px-3">Keterangan</th>
              <th class="pb-3 px-3">Dompet</th>
              <th class="pb-3 px-3">Tipe</th>
              <th class="pb-3 px-3 text-right">Nominal</th>
              <th class="pb-3 px-3 text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60 text-xs font-semibold">
            <tr
              v-for="tx in filteredTransactions"
              :key="tx.id"
              class="hover:bg-slate-800/40 transition group"
            >
              <!-- Date -->
              <td class="py-4 px-3 text-slate-400 font-mono whitespace-nowrap">
                {{ formatDate(tx.created_at) }}
              </td>

              <!-- Description -->
              <td class="py-4 px-3 text-white max-w-xs truncate">
                {{ tx.description || 'Transaksi' }}
              </td>

              <!-- Wallet -->
              <td class="py-4 px-3 whitespace-nowrap">
                <span class="px-2.5 py-1 rounded-lg bg-slate-800 text-teal-300 text-[11px] border border-slate-700/60 font-medium">
                  {{ getWalletName(tx.wallet_id) }}
                </span>
              </td>

              <!-- Type Badge -->
              <td class="py-4 px-3 whitespace-nowrap">
                <span
                  class="px-2 py-0.5 rounded-md text-[10px] uppercase font-bold"
                  :class="[
                    tx.type === 'income' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' :
                    tx.type === 'allocation' ? 'bg-cyan-500/10 text-cyan-400 border border-cyan-500/20' :
                    'bg-rose-500/10 text-rose-400 border border-rose-500/20'
                  ]"
                >
                  {{ tx.type }}
                </span>
              </td>

              <!-- Amount -->
              <td class="py-4 px-3 text-right whitespace-nowrap font-mono font-bold" :class="tx.type === 'income' ? 'text-emerald-400' : 'text-slate-100'">
                {{ tx.type === 'income' ? '+' : '-' }}{{ formatRp(tx.amount) }}
              </td>

              <!-- Actions -->
              <td class="py-4 px-3 text-center whitespace-nowrap">
                <!-- Admin Actions -->
                <div v-if="isAdmin" class="flex items-center justify-center gap-2">
                  <button
                    @click="openEdit(tx)"
                    class="px-2.5 py-1 rounded-lg text-xs text-slate-300 hover:text-white hover:bg-slate-800 transition cursor-pointer"
                  >
                    Edit
                  </button>
                  <button
                    @click="openDelete(tx)"
                    class="px-2.5 py-1 rounded-lg text-xs text-rose-400 hover:text-rose-300 hover:bg-rose-950/40 transition cursor-pointer"
                  >
                    Hapus
                  </button>
                </div>

                <!-- Member Actions -->
                <div v-else class="flex items-center justify-center gap-2">
                  <button
                    @click="openChangeRequest(tx)"
                    class="px-2.5 py-1 rounded-lg text-xs text-teal-400 hover:text-teal-300 hover:bg-teal-950/40 transition cursor-pointer"
                  >
                    Revisi
                  </button>
                  <button
                    @click="openDeleteRequest(tx)"
                    class="px-2.5 py-1 rounded-lg text-xs text-rose-400 hover:text-rose-300 hover:bg-rose-950/40 transition cursor-pointer"
                  >
                    Minta Hapus
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modals -->
    <CreateTransactionModal v-model:isOpen="isCreateModalOpen" />
    <EditTransactionModal v-model:isOpen="isEditModalOpen" :transaction="selectedTx" />
    <DeleteTransactionModal v-model:isOpen="isDeleteModalOpen" :transaction="selectedTx" />
    <ChangeRequestModal v-model:isOpen="isChangeRequestOpen" :transaction="selectedTx" />
    <DeleteRequestModal v-model:isOpen="isDeleteRequestOpen" :transaction="selectedTx" />
  </div>
</template>
