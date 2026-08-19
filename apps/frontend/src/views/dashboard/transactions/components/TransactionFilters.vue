<script setup lang="ts">
import { ref, watch } from 'vue'
import type { TransactionFilters } from '../../../../types'
import { useWalletStore } from '../../../../stores/wallet'
import Button from '../../../../components/ui/Button.vue'

const props = defineProps<{
  filters: TransactionFilters
}>()

const emit = defineEmits<{
  (e: 'update', filters: Partial<TransactionFilters>): void
  (e: 'clear'): void
}>()

const walletStore = useWalletStore()
const localFilters = ref<TransactionFilters>({ ...props.filters })

watch(
  () => props.filters,
  (newF) => {
    localFilters.value = { ...newF }
  },
  { deep: true }
)

const handleUpdate = (key: keyof TransactionFilters, value: any) => {
  localFilters.value[key] = value
  emit('update', { [key]: value })
}

const handleClear = () => {
  localFilters.value = {
    wallet_id: '',
    type: '',
    dateFrom: '',
    dateTo: '',
    search: ''
  }
  emit('clear')
}
</script>

<template>
  <div class="card-neo bg-slate-900/90 p-5 sm:p-6 rounded-3xl border border-slate-800/90 shadow-card flex flex-col gap-4">
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3.5">
      <!-- Search Input -->
      <div>
        <label class="text-[11px] font-bold text-slate-300 block mb-1">Cari Keterangan</label>
        <input
          type="text"
          :value="localFilters.search"
          @input="handleUpdate('search', ($event.target as HTMLInputElement).value)"
          placeholder="Cari transaksi..."
          class="w-full px-3.5 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs font-semibold text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-teal-500/30 focus:border-teal-600"
        />
      </div>

      <!-- Wallet Selector -->
      <div>
        <label class="text-[11px] font-bold text-slate-300 block mb-1">Dompet</label>
        <select
          :value="localFilters.wallet_id"
          @change="handleUpdate('wallet_id', ($event.target as HTMLSelectElement).value)"
          class="w-full px-3.5 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs font-semibold text-slate-100 focus:outline-none focus:ring-2 focus:ring-teal-500/30 focus:border-teal-600"
        >
          <option value="">Semua Dompet</option>
          <option v-for="w in walletStore.wallets" :key="w.id" :value="w.id">
            {{ w.name }}
          </option>
        </select>
      </div>

      <!-- Type Selector -->
      <div>
        <label class="text-[11px] font-bold text-slate-300 block mb-1">Tipe Transaksi</label>
        <select
          :value="localFilters.type"
          @change="handleUpdate('type', ($event.target as HTMLSelectElement).value)"
          class="w-full px-3.5 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs font-semibold text-slate-100 focus:outline-none focus:ring-2 focus:ring-teal-500/30 focus:border-teal-600"
        >
          <option value="">Semua Tipe</option>
          <option value="expense">Pengeluaran</option>
          <option value="income">Pemasukan</option>
          <option value="allocation">Alokasi</option>
        </select>
      </div>

      <!-- From Date -->
      <div>
        <label class="text-[11px] font-bold text-slate-300 block mb-1">Dari Tanggal</label>
        <input
          type="date"
          :value="localFilters.dateFrom"
          @input="handleUpdate('dateFrom', ($event.target as HTMLInputElement).value)"
          class="w-full px-3.5 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs font-semibold text-slate-100 focus:outline-none focus:ring-2 focus:ring-teal-500/30 focus:border-teal-600"
        />
      </div>

      <!-- To Date -->
      <div>
        <label class="text-[11px] font-bold text-slate-300 block mb-1">Sampai Tanggal</label>
        <input
          type="date"
          :value="localFilters.dateTo"
          @input="handleUpdate('dateTo', ($event.target as HTMLInputElement).value)"
          class="w-full px-3.5 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs font-semibold text-slate-100 focus:outline-none focus:ring-2 focus:ring-teal-500/30 focus:border-teal-600"
        />
      </div>
    </div>

    <!-- Clear filters row -->
    <div class="flex items-center justify-end pt-2 border-t border-slate-800/60">
      <Button
        variant="ghost"
        size="xs"
        @click="handleClear"
      >
        ✕ Reset Filter
      </Button>
    </div>
  </div>
</template>
