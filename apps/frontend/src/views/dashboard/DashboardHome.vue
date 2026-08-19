<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import { useFamilyStore } from '../../stores/family'
import { useWallets } from '../../composables/useWallets'
import { useTransactions } from '../../composables/useTransactions'
import { useFamily } from '../../composables/useFamily'
import { useI18n } from '../../locales'
import type { Wallet, Transaction } from '../../types'

// Components
import TotalFundsCard from '../../components/common/TotalFundsCard.vue'
import FamilySpendingCard from '../../components/common/FamilySpendingCard.vue'
import AllocatedWalletCard from '../../components/common/AllocatedWalletCard.vue'
import RecentTransactionsTable from '../../components/common/RecentTransactionsTable.vue'
import CashflowChart from '../../components/common/CashflowChart.vue'
import Skeleton from '../../components/ui/Skeleton.vue'

// Modals
import CreateWalletModal from './wallets/components/CreateWalletModal.vue'
import EditWalletModal from './wallets/components/EditWalletModal.vue'
import DeleteWalletModal from './wallets/components/DeleteWalletModal.vue'
import AllocateModal from './wallets/components/AllocateModal.vue'
import CreateTransactionModal from './transactions/components/CreateTransactionModal.vue'
import EditTransactionModal from './transactions/components/EditTransactionModal.vue'
import ChangeRequestModal from './transactions/components/ChangeRequestModal.vue'
import DeleteRequestModal from './transactions/components/DeleteRequestModal.vue'

const authStore = useAuthStore()
const familyStore = useFamilyStore()
const router = useRouter()
const { t } = useI18n()

const { wallets, isLoading: isWalletsLoading } = useWallets()
const { transactions, filteredTransactions, isLoading: isTxLoading, exportCsv } = useTransactions()
const { members, primaryBalance } = useFamily()

const isAdmin = computed(() => authStore.user?.role === 'admin')

// Total funds = primary balance + sum of all wallet balances
const totalFunds = computed(() => {
  const primary = primaryBalance.value || 0
  const walletsSum = (wallets.value || []).reduce((sum, w) => sum + (w?.current_balance || 0), 0)
  return primary + walletsSum
})

// Wallets with calculated spent
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

// Member spending list
const familySpendingList = computed(() => {
  const mems = members.value || []
  return mems.map((m) => {
    const spent = (transactions.value || [])
      .filter((t) => t.user_id === m.user_id && t.type === 'expense')
      .reduce((sum, t) => sum + (t.amount || 0), 0)

    return {
      id: m.id,
      name: m.user_name || m.role,
      role: m.role,
      spent: spent > 0 ? spent : 0
    }
  })
})

// Modal states
const isCreateWalletOpen = ref(false)
const isEditWalletOpen = ref(false)
const isDeleteWalletOpen = ref(false)
const isAllocateModalOpen = ref(false)
const isCreateTxOpen = ref(false)
const isEditTxOpen = ref(false)
const isChangeRequestOpen = ref(false)
const isDeleteRequestOpen = ref(false)

const selectedWallet = ref<Wallet | null>(null)
const selectedTx = ref<Transaction | null>(null)

const openAllocate = (wallet?: Wallet) => {
  selectedWallet.value = wallet || wallets.value[0] || null
  isAllocateModalOpen.value = true
}

const openEditWallet = (wallet: Wallet) => {
  selectedWallet.value = wallet
  isEditWalletOpen.value = true
}

const openDeleteWallet = (wallet: Wallet) => {
  selectedWallet.value = wallet
  isDeleteWalletOpen.value = true
}

const handleSelectTransaction = (tx: Transaction) => {
  selectedTx.value = tx
  if (isAdmin.value) {
    isEditTxOpen.value = true
  } else {
    isChangeRequestOpen.value = true
  }
}

onMounted(async () => {
  if (!familyStore.family) {
    await familyStore.fetchMyFamily()
    if (!familyStore.family) {
      router.push('/family-setup')
    }
  }
})
</script>

<template>
  <div class="flex flex-col gap-8 pb-12">
    <!-- Top Hero Metric Section: 2 Columns -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch">
      <!-- Left Card: Total Family Funds (7 cols) -->
      <div class="lg:col-span-7">
        <Skeleton v-if="isWalletsLoading" type="stat" />
        <TotalFundsCard
          v-else
          :total-funds="totalFunds"
          :trend-percentage="0"
          :is-admin="isAdmin"
          @quick-allocate="openAllocate()"
          @transfer-money="isCreateTxOpen = true"
        />
      </div>

      <!-- Right Card: Family Spending Breakdown (5 cols) -->
      <div class="lg:col-span-5">
        <Skeleton v-if="isTxLoading" type="card" />
        <FamilySpendingCard
          v-else
          :members="familySpendingList"
          @manage-family="router.push('/family')"
        />
      </div>
    </div>

    <!-- Middle Section: Allocated Wallets -->
    <section class="flex flex-col gap-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <h3 class="text-lg font-bold text-white tracking-tight">
            {{ t('dashboard.wallets.title') || 'Dompet Terkelola' }}
          </h3>
          <span class="text-xs font-semibold text-slate-400">
            {{ wallets.length }} {{ t('extra.activeWallets') || 'dompet aktif' }}
          </span>
        </div>

        <div class="flex items-center gap-3">
          <button
            v-if="isAdmin"
            @click="isCreateWalletOpen = true"
            class="text-xs font-bold text-teal-400 hover:text-teal-300 transition hover:underline cursor-pointer"
          >
            + {{ t('extra.addWallet') || 'Buat Dompet' }}
          </button>
          <router-link
            to="/wallets"
            class="text-xs font-bold text-slate-400 hover:text-white transition"
          >
            Lihat Semua →
          </router-link>
        </div>
      </div>

      <!-- Loading Skeletons -->
      <div v-if="isWalletsLoading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <Skeleton v-for="i in 3" :key="i" type="wallet" />
      </div>

      <!-- Empty State -->
      <div
        v-else-if="wallets.length === 0"
        class="card-neo bg-slate-900/90 rounded-3xl p-10 border border-slate-800 text-center flex flex-col items-center justify-center gap-3"
      >
        <p class="text-slate-400 text-xs font-semibold">
          {{ t('dashboard.wallets.noWallets') || 'Belum ada dompet teralokasi' }}
        </p>
        <button
          v-if="isAdmin"
          @click="isCreateWalletOpen = true"
          class="px-4 py-2 bg-teal-700 hover:bg-teal-800 text-white rounded-xl text-xs font-bold transition shadow-sm cursor-pointer"
        >
          {{ t('extra.createFirstWallet') || 'Buat Dompet Pertama' }}
        </button>
      </div>

      <!-- 3-Column Wallets Grid -->
      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <AllocatedWalletCard
          v-for="item in walletsWithSpent.slice(0, 6)"
          :key="item.wallet.id"
          :wallet="item.wallet"
          :spent="item.spent"
          :limit="item.limit"
          :is-admin="isAdmin"
          @allocate="openAllocate(item.wallet)"
          @edit="openEditWallet(item.wallet)"
          @delete="openDeleteWallet(item.wallet)"
        />
      </div>
    </section>

    <!-- Cashflow Analytics Chart Section -->
    <section class="card-neo bg-slate-900/90 rounded-3xl p-6 sm:p-7 border border-slate-800/90 shadow-card flex flex-col gap-4">
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-base sm:text-lg font-bold text-white tracking-tight">
            Cashflow 7 Hari Terakhir
          </h3>
          <p class="text-xs text-slate-400 mt-0.5">
            Perbandingan pemasukan dan pengeluaran per hari
          </p>
        </div>
      </div>
      <CashflowChart :transactions="filteredTransactions" />
    </section>

    <!-- Bottom Section: Recent Transactions -->
    <section class="flex flex-col gap-4">
      <RecentTransactionsTable
        :transactions="transactions"
        :wallets="wallets"
        @filter="router.push('/transactions')"
        @export-csv="exportCsv"
        @select-transaction="handleSelectTransaction"
      />
    </section>

    <!-- Modals -->
    <CreateWalletModal v-model:isOpen="isCreateWalletOpen" />
    <EditWalletModal v-model:isOpen="isEditWalletOpen" :wallet="selectedWallet" />
    <DeleteWalletModal v-model:isOpen="isDeleteWalletOpen" :wallet="selectedWallet" />
    <AllocateModal v-model:isOpen="isAllocateModalOpen" :wallet="selectedWallet" />
    <CreateTransactionModal v-model:isOpen="isCreateTxOpen" />
    <EditTransactionModal v-model:isOpen="isEditTxOpen" :transaction="selectedTx" />
    <ChangeRequestModal v-model:isOpen="isChangeRequestOpen" :transaction="selectedTx" />
    <DeleteRequestModal v-model:isOpen="isDeleteRequestOpen" :transaction="selectedTx" />
  </div>
</template>
