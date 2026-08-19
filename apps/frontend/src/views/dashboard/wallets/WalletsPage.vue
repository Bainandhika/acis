<script setup lang="ts">
import { ref, computed } from 'vue'
import { useWallets } from '../../../composables/useWallets'
import { useTransactions } from '../../../composables/useTransactions'
import { useAuthStore } from '../../../stores/auth'
import { useI18n } from '../../../locales'
import { formatRp } from '../../../utils/format'
import type { Wallet } from '../../../types'

// Components
import AllocatedWalletCard from '../../../components/common/AllocatedWalletCard.vue'
import Skeleton from '../../../components/ui/Skeleton.vue'
import Button from '../../../components/ui/Button.vue'

// Modals
import CreateWalletModal from './components/CreateWalletModal.vue'
import EditWalletModal from './components/EditWalletModal.vue'
import DeleteWalletModal from './components/DeleteWalletModal.vue'
import AllocateModal from './components/AllocateModal.vue'

const { wallets, totalBalance, isLoading } = useWallets()
const { transactions } = useTransactions()
const authStore = useAuthStore()
const { t } = useI18n()

const isAdmin = computed(() => authStore.user?.role === 'admin')

const isCreateModalOpen = ref(false)
const isEditModalOpen = ref(false)
const isDeleteModalOpen = ref(false)
const isAllocateModalOpen = ref(false)
const selectedWallet = ref<Wallet | null>(null)

const walletsWithSpent = computed(() => {
  return (wallets.value || []).map((w) => {
    const spent = (transactions.value || [])
      .filter((t) => t.wallet_id === w.id && t.type === 'expense')
      .reduce((sum, t) => sum + (t.amount || 0), 0)

    const limit = w.initial_balance > 0 ? w.initial_balance : (w.minimum_limit > 0 ? w.minimum_limit : 0)
    return {
      wallet: w,
      spent: spent > 0 ? spent : 0,
      limit: limit
    }
  })
})

const openAllocate = (wallet: Wallet) => {
  selectedWallet.value = wallet
  isAllocateModalOpen.value = true
}

const openEdit = (wallet: Wallet) => {
  selectedWallet.value = wallet
  isEditModalOpen.value = true
}

const openDelete = (wallet: Wallet) => {
  selectedWallet.value = wallet
  isDeleteModalOpen.value = true
}
</script>

<template>
  <div class="flex flex-col gap-8 pb-12">
    <!-- Top Header -->
    <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
      <div>
        <h1 class="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
          {{ t('walletsTab.title') || 'Dompet Virtual' }}
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          {{ t('walletsTab.subtitle') || 'Kelola alokasi dan batas anggaran keluarga' }}
        </p>
      </div>

      <Button
        v-if="isAdmin"
        variant="primary"
        size="sm"
        @click="isCreateModalOpen = true"
      >
        + {{ t('extra.addWallet') || 'Buat Dompet' }}
      </Button>
    </div>

    <!-- Total Balance Hero Card -->
    <div class="card-neo bg-gradient-to-br from-slate-900 via-slate-900 to-teal-950/40 p-6 sm:p-7 rounded-3xl border border-slate-800/90 shadow-card flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
      <div>
        <span class="text-[11px] font-bold uppercase tracking-wider text-teal-400">Total Saldo Terkelola</span>
        <h2 class="text-3xl sm:text-4xl font-black text-white tracking-tight mt-1 font-sans">
          {{ formatRp(totalBalance) }}
        </h2>
        <span class="text-xs text-slate-400 mt-1 block">
          {{ wallets.length }} {{ t('extra.activeWallets') || 'dompet aktif' }}
        </span>
      </div>

      <div class="flex items-center gap-3">
        <Button
          v-if="isAdmin && wallets.length > 0 && wallets[0]"
          variant="secondary"
          size="sm"
          @click="openAllocate(wallets[0])"
        >
          {{ t('extra.quickAllocate') || 'Alokasi Dana' }}
        </Button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <Skeleton v-for="i in 6" :key="i" type="wallet" />
    </div>

    <!-- Empty State -->
    <div
      v-else-if="wallets.length === 0"
      class="card-neo bg-slate-900/90 rounded-3xl p-12 border border-slate-800 text-center flex flex-col items-center justify-center gap-3"
    >
      <div class="w-12 h-12 rounded-2xl bg-slate-800 flex items-center justify-center text-teal-400 text-xl font-bold mb-1">
        💼
      </div>
      <h3 class="text-base font-bold text-white">
        {{ t('walletsTab.noWallets') || 'Belum ada dompet' }}
      </h3>
      <p class="text-xs text-slate-400 max-w-sm">
        Buat dompet virtual untuk mulai mengelompokkan anggaran dan melacak pengeluaran keluarga.
      </p>
      <Button
        v-if="isAdmin"
        variant="primary"
        size="sm"
        @click="isCreateModalOpen = true"
        class="mt-2"
      >
        {{ t('extra.createFirstWallet') || 'Buat Dompet Pertama' }}
      </Button>
    </div>

    <!-- Wallets Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <AllocatedWalletCard
        v-for="item in walletsWithSpent"
        :key="item.wallet.id"
        :wallet="item.wallet"
        :spent="item.spent"
        :limit="item.limit"
        :is-admin="isAdmin"
        @allocate="openAllocate(item.wallet)"
        @edit="openEdit(item.wallet)"
        @delete="openDelete(item.wallet)"
      />
    </div>

    <!-- Modals -->
    <CreateWalletModal v-model:isOpen="isCreateModalOpen" />
    <EditWalletModal v-model:isOpen="isEditModalOpen" :wallet="selectedWallet" />
    <DeleteWalletModal v-model:isOpen="isDeleteModalOpen" :wallet="selectedWallet" />
    <AllocateModal v-model:isOpen="isAllocateModalOpen" :wallet="selectedWallet" />
  </div>
</template>
