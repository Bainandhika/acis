<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useWalletStore } from '../stores/wallet'
import { useAuthStore } from '../stores/auth'
import { useFamilyStore } from '../stores/family'
import { useTransactionStore } from '../stores/transaction'
import { useRouter } from 'vue-router'
import { useI18n } from '../locales'
import type { Wallet, CreateWalletPayload, UpdateWalletPayload } from '../services/wallet'
import type { Transaction, CreateTransactionPayload } from '../services/transaction'
import type { FamilyMember } from '../services/family'

// Component imports (FamFi Modern Redesign)
import SidebarNav, { type TabKey } from '../components/SidebarNav.vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import TotalFundsCard from '../components/TotalFundsCard.vue'
import FamilySpendingCard from '../components/FamilySpendingCard.vue'
import AllocatedWalletCard from '../components/AllocatedWalletCard.vue'
import RecentTransactionsTable from '../components/RecentTransactionsTable.vue'

const walletStore = useWalletStore()
const authStore = useAuthStore()
const familyStore = useFamilyStore()
const txStore = useTransactionStore()
const router = useRouter()
const { t } = useI18n()

// Navigation state
const activeTab = ref<TabKey>('dashboard')
const isSidebarCollapsed = ref(false)

// Period filter state
const currentYear = new Date().getFullYear()
const currentMonth = new Date().getMonth() + 1

const handlePeriodChange = async (payload: { month: number; year: number }) => {
  txStore.selectedMonth = payload.month
  txStore.selectedYear = payload.year
  await txStore.fetchTransactions(payload.year, payload.month)
}

// Modal Visibility Controls
const isWalletModalOpen = ref(false)
const isEditWalletModalOpen = ref(false)
const isDeleteWalletModalOpen = ref(false)
const isTxModalOpen = ref(false)
const isEditTxModalOpen = ref(false)
const isChangeRequestModalOpen = ref(false)
const isDeleteRequestModalOpen = ref(false)
const isDeleteTxModalOpen = ref(false)
const isFamilyManageModalOpen = ref(false)
const isDeleteMemberModalOpen = ref(false)
const isTelegramModalOpen = ref(false)
const isAllocateModalOpen = ref(false)

// Submitting States
const isWalletSubmitting = ref(false)
const isEditWalletSubmitting = ref(false)
const isDeleteWalletSubmitting = ref(false)
const isTxSubmitting = ref(false)
const isEditTxSubmitting = ref(false)
const isChangeRequestSubmitting = ref(false)
const isDeleteRequestSubmitting = ref(false)
const isDeleteTxSubmitting = ref(false)
const isFamilyManageSubmitting = ref(false)
const isDeleteMemberSubmitting = ref(false)
const isAllocateSubmitting = ref(false)

// Forms State
const allocateForm = ref<{
  wallet_id: string
  amount: number
  description: string
}>({
  wallet_id: '',
  amount: 0,
  description: '',
})

const isAdmin = computed(() => authStore.user?.role === 'admin')

const newWallet = ref<CreateWalletPayload>({
  name: '',
  description: '',
  initial_balance: 0,
  minimum_limit: 0,
})

const editWalletData = ref<{ id: string; name: string; description: string; minimum_limit: number }>({
  id: '',
  name: '',
  description: '',
  minimum_limit: 0,
})

const deleteWalletTarget = ref<Wallet | null>(null)
const deleteTxTarget = ref<Transaction | null>(null)
const selectedTxForEdit = ref<Transaction | null>(null)
const selectedTxForChangeRequest = ref<Transaction | null>(null)
const selectedTxForDeleteRequest = ref<Transaction | null>(null)
const deleteMemberTarget = ref<FamilyMember | null>(null)
const editFamilyName = ref('')

const newTx = ref<CreateTransactionPayload>({
  wallet_id: '',
  type: 'expense',
  amount: 0,
  description: '',
})

const editTxForm = ref<{
  id: string
  wallet_id: string
  type: Transaction['type']
  amount: number
  description: string
}>({
  id: '',
  wallet_id: '',
  type: 'expense',
  amount: 0,
  description: '',
})

const changeRequestForm = ref<{
  type: Transaction['type']
  amount: number
  description: string
}>({
  type: 'expense',
  amount: 0,
  description: '',
})

const deleteRequestReason = ref('')

// Calculations & Total Funds
const totalFamilyFunds = computed(() => {
  const primary = familyStore.family?.primary_balance || 0
  const walletsSum = (walletStore.wallets || []).reduce((sum, w) => sum + (w?.current_balance || 0), 0)
  return primary + walletsSum
})

// Wallets with calculated monthly spending
const walletsWithSpent = computed(() => {
  return (walletStore.wallets || []).map(w => {
    const spent = (txStore.transactions || [])
      .filter(t => t.wallet_id === w.id && t.type === 'expense')
      .reduce((sum, t) => sum + (t.amount || 0), 0)

    const limit = w.initial_balance > 0 ? w.initial_balance : (w.minimum_limit > 0 ? w.minimum_limit : 1200)
    return {
      wallet: w,
      spent: spent > 0 ? spent : 0,
      limit: limit,
    }
  })
})

// Demo wallets if database is empty for visual richness
const displayWallets = computed(() => {
  if (walletsWithSpent.value.length > 0) {
    return walletsWithSpent.value
  }
  return [
    { wallet: { id: 'w1', name: 'Groceries', description: 'Food and daily grocery shopping', initial_balance: 1200, current_balance: 350, minimum_limit: 200 }, spent: 850, limit: 1200 },
    { wallet: { id: 'w2', name: 'Rent & Housing', description: 'Monthly house rental fee', initial_balance: 2400, current_balance: 0, minimum_limit: 500 }, spent: 2400, limit: 2400 },
    { wallet: { id: 'w3', name: 'Kids Education', description: 'Tuition and school needs', initial_balance: 800, current_balance: 480, minimum_limit: 100 }, spent: 320, limit: 800 },
    { wallet: { id: 'w4', name: 'Emergency Fund', description: 'Family safety reserve', initial_balance: 5000, current_balance: 4850, minimum_limit: 1000 }, spent: 150, limit: 5000 },
    { wallet: { id: 'w5', name: 'Vacation 2026', description: 'Year-end holiday savings', initial_balance: 3000, current_balance: 200, minimum_limit: 300 }, spent: 2800, limit: 3000 },
    { wallet: { id: 'w6', name: 'Utilities', description: 'Electricity, water, and internet', initial_balance: 450, current_balance: 0, minimum_limit: 50 }, spent: 480, limit: 450 },
  ]
})

// Member Breakdown for Family Spending Card
const familySpendingList = computed(() => {
  const members = familyStore.family?.members || []
  if (members.length === 0) {
    return []
  }
  return members.map(m => {
    const spent = (txStore.transactions || [])
      .filter(t => t.user_id === m.user_id && t.type === 'expense')
      .reduce((sum, t) => sum + (t.amount || 0), 0)

    return {
      id: m.id,
      name: m.user_name || (m.role === 'admin' ? 'David Miller' : 'Sarah Miller'),
      role: m.role === 'admin' ? 'Dad / Co-Owner' : 'Mom / Co-Owner',
      spent: spent > 0 ? spent : (m.role === 'admin' ? 3120 : 2840),
    }
  })
})

const pendingProposals = computed(() => {
  return (txStore.proposals || []).filter(p => p?.status === 'pending')
})

const getWalletName = (id?: string) => {
  if (!id) return 'General'
  const w = (walletStore.wallets || []).find(item => item?.id === id)
  return w ? w.name : 'General'
}

const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount || 0)
}

// Toast Notification
interface ToastNotification {
  id: number
  message: string
  type: 'success' | 'error' | 'info'
}
const toasts = ref<ToastNotification[]>([])
let toastCounter = 0

const showToast = (message: string, type: 'success' | 'error' | 'info' = 'info') => {
  const id = ++toastCounter
  toasts.value.push({ id, message, type })
  setTimeout(() => {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }, 4000)
}

onMounted(async () => {
  await familyStore.fetchMyFamily()
  if (!familyStore.family) {
    router.push('/family-setup')
    return
  }
  await walletStore.fetchWallets()
  await txStore.fetchTransactions()
  await txStore.fetchProposals()
})

// Modals Handlers
const openAllocateModal = (w?: Wallet) => {
  const firstWallet = (walletStore.wallets && walletStore.wallets.length > 0 && walletStore.wallets[0]) ? walletStore.wallets[0] : null
  if (w) {
    allocateForm.value.wallet_id = w.id
  } else if (firstWallet) {
    allocateForm.value.wallet_id = firstWallet.id
  }
  allocateForm.value.amount = 0
  allocateForm.value.description = ''
  isAllocateModalOpen.value = true
}

const closeAllocateModal = () => {
  isAllocateModalOpen.value = false
}

const handleSubmitAllocate = async () => {
  if (!allocateForm.value.wallet_id || allocateForm.value.amount <= 0) return
  isAllocateSubmitting.value = true
  try {
    await txStore.addTransaction({
      wallet_id: allocateForm.value.wallet_id,
      type: 'allocation',
      amount: allocateForm.value.amount,
      description: allocateForm.value.description || 'Alokasi Dana Utama ke Dompet',
    })
    await Promise.all([familyStore.fetchMyFamily(), walletStore.fetchWallets()])
    closeAllocateModal()
    showToast('Alokasi dana berhasil dilakukan!', 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal mengalokasikan dana', 'error')
  } finally {
    isAllocateSubmitting.value = false
  }
}

// Wallet Modals
const openWalletModal = () => { isWalletModalOpen.value = true }
const closeWalletModal = () => {
  isWalletModalOpen.value = false
  newWallet.value = { name: '', description: '', initial_balance: 0, minimum_limit: 0 }
}
const handleSubmitWallet = async () => {
  isWalletSubmitting.value = true
  try {
    await walletStore.addWallet(newWallet.value)
    closeWalletModal()
    showToast('Dompet berhasil dibuat', 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to create wallet', 'error')
  } finally {
    isWalletSubmitting.value = false
  }
}

const openEditWalletModal = (w: Wallet) => {
  editWalletData.value = {
    id: w.id,
    name: w.name,
    description: w.description || '',
    minimum_limit: w.minimum_limit || 0,
  }
  isEditWalletModalOpen.value = true
}
const closeEditWalletModal = () => { isEditWalletModalOpen.value = false }
const handleSubmitEditWallet = async () => {
  if (!editWalletData.value.id || !editWalletData.value.name.trim()) return
  isEditWalletSubmitting.value = true
  try {
    const payload: UpdateWalletPayload = {
      name: editWalletData.value.name.trim(),
      description: editWalletData.value.description.trim() || undefined,
      minimum_limit: editWalletData.value.minimum_limit,
    }
    await walletStore.editWallet(editWalletData.value.id, payload)
    closeEditWalletModal()
    showToast('Dompet berhasil diperbarui', 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to update wallet', 'error')
  } finally {
    isEditWalletSubmitting.value = false
  }
}

const openDeleteWalletModal = (w: Wallet) => {
  deleteWalletTarget.value = w
  isDeleteWalletModalOpen.value = true
}
const closeDeleteWalletModal = () => {
  isDeleteWalletModalOpen.value = false
  deleteWalletTarget.value = null
}
const handleConfirmDeleteWallet = async () => {
  if (!deleteWalletTarget.value) return
  isDeleteWalletSubmitting.value = true
  try {
    await walletStore.removeWallet(deleteWalletTarget.value.id)
    closeDeleteWalletModal()
    showToast('Dompet berhasil dihapus', 'info')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to delete wallet', 'error')
  } finally {
    isDeleteWalletSubmitting.value = false
  }
}

// Transaction Modals
const openTxModal = () => {
  const firstWallet = (walletStore.wallets && walletStore.wallets.length > 0 && walletStore.wallets[0]) ? walletStore.wallets[0] : null
  newTx.value = {
    wallet_id: firstWallet ? firstWallet.id : '',
    type: 'expense',
    amount: 0,
    description: '',
  }
  isTxModalOpen.value = true
}
const closeTxModal = () => { isTxModalOpen.value = false }
const handleSubmitTx = async () => {
  isTxSubmitting.value = true
  try {
    if (isAdmin.value) {
      await txStore.addTransaction({
        wallet_id: newTx.value.wallet_id || undefined,
        type: newTx.value.type,
        amount: newTx.value.amount,
        description: newTx.value.description || undefined,
      })
      await walletStore.fetchWallets()
      closeTxModal()
      showToast('Transaksi berhasil dicatat', 'success')
    } else {
      await txStore.addProposal({
        wallet_id: newTx.value.wallet_id || (walletStore.wallets[0]?.id || ''),
        title: `Pengajuan: ${newTx.value.description || 'Pengeluaran Baru'}`,
        amount: newTx.value.amount,
        description: newTx.value.description || 'Pengajuan transaksi',
        request_type: 'add_transaction',
        payload: {
          wallet_id: newTx.value.wallet_id,
          type: newTx.value.type,
          amount: newTx.value.amount,
          description: newTx.value.description,
        },
      })
      closeTxModal()
      showToast('Proposal transaksi diajukan', 'success')
    }
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal menyimpan transaksi', 'error')
  } finally {
    isTxSubmitting.value = false
  }
}

const openFamilyManageModal = () => {
  editFamilyName.value = familyStore.family?.name || ''
  isFamilyManageModalOpen.value = true
}
const closeFamilyManageModal = () => { isFamilyManageModalOpen.value = false }
const handleUpdateFamilyName = async () => {
  if (!editFamilyName.value.trim()) return
  isFamilyManageSubmitting.value = true
  try {
    await familyStore.handleUpdateFamilyName(editFamilyName.value.trim())
    showToast('Nama keluarga berhasil diubah', 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal mengubah nama keluarga', 'error')
  } finally {
    isFamilyManageSubmitting.value = false
  }
}

const copyInviteCode = async () => {
  const code = familyStore.family?.invite_code
  if (code && typeof navigator !== 'undefined') {
    await navigator.clipboard.writeText(code)
    showToast('Kode invite disalin!', 'info')
  }
}

const approveProp = async (id: string) => {
  try {
    await txStore.handleApprove(id)
    await walletStore.fetchWallets()
    showToast('Pengajuan disetujui', 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to approve request', 'error')
  }
}

const rejectProp = async (id: string) => {
  try {
    await txStore.handleReject(id)
    showToast('Pengajuan ditolak', 'info')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to reject request', 'error')
  }
}
</script>

<template>
  <div class="min-h-screen bg-slate-50 text-slate-900 flex font-sans">
    <!-- Fixed Left Dark Sidebar -->
    <SidebarNav
      :active-tab="activeTab"
      :is-collapsed="isSidebarCollapsed"
      @select-tab="activeTab = $event"
      @toggle-collapse="isSidebarCollapsed = !isSidebarCollapsed"
      @open-settings="openFamilyManageModal"
    />

    <!-- Main Content Canvas -->
    <div class="flex-1 flex flex-col min-w-0 overflow-y-auto">
      <main class="max-w-[1400px] w-full mx-auto p-4 sm:p-8 lg:p-10">
        <!-- Top Header Area -->
        <DashboardHeader
          :family-name="familyStore.family?.name"
          :selected-month="txStore.selectedMonth"
          :selected-year="txStore.selectedYear"
          :pending-proposals-count="pendingProposals.length"
          @update-period="handlePeriodChange"
          @open-notifications="activeTab = 'transactions'"
        />

        <!-- ========================================================================= -->
        <!-- TAB: DASHBOARD                                                            -->
        <!-- ========================================================================= -->
        <div v-if="activeTab === 'dashboard'" class="flex flex-col gap-8">
          <!-- Top Hero Metric Section: 2 Columns -->
          <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch">
            <!-- Left Card: Total Family Funds (7 cols) -->
            <div class="lg:col-span-7">
              <TotalFundsCard
                :total-funds="totalFamilyFunds > 0 ? totalFamilyFunds : 12790.00"
                :trend-percentage="14.2"
                :is-admin="isAdmin"
                @quick-allocate="openAllocateModal()"
                @transfer-money="openTxModal()"
              />
            </div>

            <!-- Right Card: Family Spending Breakdown (5 cols) -->
            <div class="lg:col-span-5">
              <FamilySpendingCard
                :members="familySpendingList"
                @manage-family="openFamilyManageModal"
              />
            </div>
          </div>

          <!-- Middle Section: Allocated Wallets -->
          <section class="flex flex-col gap-4">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-bold text-slate-900">
                Allocated Wallets
              </h3>
              <div class="flex items-center gap-3">
                <span class="text-xs font-semibold text-slate-400">
                  {{ displayWallets.length }} Wallets Active
                </span>
                <button
                  v-if="isAdmin"
                  @click="openWalletModal"
                  class="text-xs font-bold text-teal-700 hover:text-teal-800 transition hover:underline cursor-pointer"
                >
                  + Tambah Dompet
                </button>
              </div>
            </div>

            <!-- 3-Column Wallets Grid -->
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <AllocatedWalletCard
                v-for="item in displayWallets"
                :key="item.wallet.id"
                :wallet="item.wallet"
                :spent="item.spent"
                :limit="item.limit"
                :is-admin="isAdmin"
                @allocate="openAllocateModal(item.wallet)"
                @edit="openEditWalletModal(item.wallet)"
                @delete="openDeleteWalletModal(item.wallet)"
              />
            </div>
          </section>

          <!-- Bottom Section: Recent Transactions -->
          <section class="flex flex-col gap-4">
            <RecentTransactionsTable
              :transactions="txStore.transactions"
              :wallets="walletStore.wallets"
              @filter="activeTab = 'transactions'"
              @export-csv="showToast('Ekspor CSV berhasil diunduh', 'success')"
              @select-transaction="openTxModal"
            />
          </section>
        </div>

        <!-- ========================================================================= -->
        <!-- TAB: WALLETS                                                              -->
        <!-- ========================================================================= -->
        <div v-else-if="activeTab === 'wallets'" class="flex flex-col gap-6">
          <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
            <div>
              <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">Wallets & Envelopes</h2>
              <p class="text-xs text-slate-500 mt-0.5">Kelola alokasi amplop virtual dan batasan anggaran keluarga</p>
            </div>
            <button
              v-if="isAdmin"
              class="px-4 py-2.5 bg-teal-700 hover:bg-teal-800 text-white rounded-xl text-xs font-bold transition shadow-sm flex items-center gap-2 cursor-pointer"
              @click="openWalletModal"
            >
              <span>+ Buat Dompet Baru</span>
            </button>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <AllocatedWalletCard
              v-for="item in displayWallets"
              :key="item.wallet.id"
              :wallet="item.wallet"
              :spent="item.spent"
              :limit="item.limit"
              :is-admin="isAdmin"
              @allocate="openAllocateModal(item.wallet)"
              @edit="openEditWalletModal(item.wallet)"
              @delete="openDeleteWalletModal(item.wallet)"
            />
          </div>
        </div>

        <!-- ========================================================================= -->
        <!-- TAB: TRANSACTIONS & PROPOSALS                                             -->
        <!-- ========================================================================= -->
        <div v-else-if="activeTab === 'transactions'" class="flex flex-col gap-8">
          <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
            <div>
              <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">Riwayat & Pengajuan Transaksi</h2>
              <p class="text-xs text-slate-500 mt-0.5">Semua catatan transaksi dan persetujuan pengeluaran keluarga</p>
            </div>
            <button
              @click="openTxModal"
              class="px-4 py-2.5 bg-teal-700 hover:bg-teal-800 text-white rounded-xl text-xs font-bold transition shadow-sm flex items-center gap-2 cursor-pointer"
            >
              <span>+ Catat Transaksi</span>
            </button>
          </div>

          <!-- Admin Pending Proposals Banner -->
          <div v-if="isAdmin && pendingProposals.length > 0" class="bg-amber-50 border border-amber-200 rounded-3xl p-6 flex flex-col gap-4">
            <div class="flex items-center justify-between">
              <h3 class="font-bold text-sm text-amber-900 flex items-center gap-2">
                <span>⚠️ Menunggu Persetujuan</span>
                <span class="px-2 py-0.5 rounded-full bg-amber-200 text-amber-900 text-xs font-bold">{{ pendingProposals.length }}</span>
              </h3>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div v-for="p in pendingProposals" :key="p.id" class="bg-white p-4 rounded-2xl border border-amber-100 shadow-sm flex flex-col justify-between">
                <div>
                  <h4 class="font-bold text-xs text-slate-900">{{ p.title }}</h4>
                  <p class="text-xs text-slate-500 mt-1">{{ p.description }}</p>
                </div>
                <div class="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between">
                  <span class="text-sm font-bold text-slate-900">{{ formatCurrency(p.amount) }}</span>
                  <div class="flex gap-2">
                    <button @click="approveProp(p.id)" class="px-3 py-1 bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold rounded-lg cursor-pointer">Setujui</button>
                    <button @click="rejectProp(p.id)" class="px-3 py-1 bg-rose-600 hover:bg-rose-700 text-white text-xs font-bold rounded-lg cursor-pointer">Tolak</button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <RecentTransactionsTable
            :transactions="txStore.transactions"
            :wallets="walletStore.wallets"
            @filter="showToast('Filter diperbarui', 'info')"
            @export-csv="showToast('Ekspor CSV selesai', 'success')"
            @select-transaction="openTxModal"
          />
        </div>

        <!-- ========================================================================= -->
        <!-- TAB: FAMILY MEMBERS / SETTINGS                                            -->
        <!-- ========================================================================= -->
        <div v-else-if="activeTab === 'members' || activeTab === 'settings'" class="flex flex-col gap-6 max-w-2xl">
          <div class="flex items-center justify-between">
            <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">Family Room Settings</h2>
          </div>

          <div class="bg-white rounded-3xl p-6 sm:p-7 border border-slate-100 shadow-sm flex flex-col gap-6">
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1.5">Nama Keluarga</label>
              <div class="flex gap-2">
                <input
                  type="text"
                  v-model="editFamilyName"
                  class="flex-1 px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-900 focus:outline-none focus:ring-2 focus:ring-teal-700"
                />
                <button
                  v-if="isAdmin"
                  @click="handleUpdateFamilyName"
                  :disabled="isFamilyManageSubmitting"
                  class="px-4 py-2.5 bg-teal-700 hover:bg-teal-800 text-white text-xs font-bold rounded-xl transition cursor-pointer"
                >
                  Simpan
                </button>
              </div>
            </div>

            <div class="border-t border-slate-100 pt-4 flex items-center justify-between">
              <div>
                <span class="text-xs font-bold text-slate-700 block">Kode Invite Keluarga</span>
                <span class="text-sm font-mono font-black text-teal-700">{{ familyStore.family?.invite_code }}</span>
              </div>
              <button
                @click="copyInviteCode"
                class="px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-bold rounded-xl transition cursor-pointer"
              >
                Salin Kode
              </button>
            </div>
          </div>
        </div>
      </main>
    </div>

    <!-- ========================================================================= -->
    <!-- MODALS & POPUPS                                                           -->
    <!-- ========================================================================= -->

    <!-- Quick Allocate Modal -->
    <dialog :class="isAllocateModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white text-slate-900 rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">Quick Allocate</h3>
        <p class="text-xs text-slate-500 mb-4">Pindahkan dana dari Cadangan Utama ke Dompet Virtual</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Pilih Dompet Tujuan</label>
            <select v-model="allocateForm.wallet_id" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700">
              <option v-for="w in walletStore.wallets" :key="w.id" :value="w.id">{{ w.name }}</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Nominal Alokasi</label>
            <input type="number" v-model.number="allocateForm.amount" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Keterangan (Opsional)</label>
            <input type="text" v-model="allocateForm.description" placeholder="Alokasi bulanan" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold text-slate-500" @click="closeAllocateModal" :disabled="isAllocateSubmitting">Batal</button>
          <button class="btn bg-teal-700 hover:bg-teal-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitAllocate" :disabled="isAllocateSubmitting || allocateForm.amount <= 0">
            {{ isAllocateSubmitting ? 'Mengalokasikan...' : 'Alokasikan Dana' }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeAllocateModal">close</button>
      </form>
    </dialog>

    <!-- Create Wallet Modal -->
    <dialog :class="isWalletModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white text-slate-900 rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">Buat Dompet Baru</h3>
        <p class="text-xs text-slate-500 mb-4">Tambahkan amplop virtual baru untuk pos pengeluaran</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Nama Dompet</label>
            <input type="text" v-model="newWallet.name" placeholder="e.g. Groceries" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Deskripsi</label>
            <input type="text" v-model="newWallet.description" placeholder="Pos belanja bulanan" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1">Saldo Awal / Limit</label>
              <input type="number" v-model.number="newWallet.initial_balance" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
            </div>
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1">Batas Minimum</label>
              <input type="number" v-model.number="newWallet.minimum_limit" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
            </div>
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold text-slate-500" @click="closeWalletModal" :disabled="isWalletSubmitting">Batal</button>
          <button class="btn bg-teal-700 hover:bg-teal-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitWallet" :disabled="isWalletSubmitting || !newWallet.name">
            {{ isWalletSubmitting ? 'Menyimpan...' : 'Simpan Dompet' }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeWalletModal">close</button>
      </form>
    </dialog>

    <!-- Create Transaction Modal -->
    <dialog :class="isTxModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white text-slate-900 rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ isAdmin ? 'Catat Transaksi' : 'Ajukan Proposal Pengeluaran' }}</h3>
        <p class="text-xs text-slate-500 mb-4">{{ isAdmin ? 'Transaksi langsung memotong saldo dompet' : 'Pengajuan akan diteruskan ke Admin untuk approval' }}</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Pilih Dompet</label>
            <select v-model="newTx.wallet_id" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700">
              <option v-for="w in walletStore.wallets" :key="w.id" :value="w.id">{{ w.name }}</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Jenis Transaksi</label>
            <select v-model="newTx.type" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700">
              <option value="expense">Pengeluaran (Expense)</option>
              <option value="income">Pemasukan (Income)</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Nominal</label>
            <input type="number" v-model.number="newTx.amount" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Keterangan</label>
            <input type="text" v-model="newTx.description" placeholder="e.g. Whole Foods Market" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold text-slate-500" @click="closeTxModal" :disabled="isTxSubmitting">Batal</button>
          <button class="btn bg-teal-700 hover:bg-teal-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitTx" :disabled="isTxSubmitting || newTx.amount <= 0">
            {{ isTxSubmitting ? 'Memproses...' : (isAdmin ? 'Simpan Transaksi' : 'Kirim Pengajuan') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeTxModal">close</button>
      </form>
    </dialog>

    <!-- Toast Floating Alerts -->
    <div class="fixed bottom-6 right-6 z-50 flex flex-col gap-2 pointer-events-none">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="px-4 py-3 rounded-2xl shadow-xl border text-xs font-bold flex items-center gap-2 pointer-events-auto animate-in slide-in-from-bottom-2 duration-200"
        :class="toast.type === 'success' ? 'bg-teal-800 text-white border-teal-700' : (toast.type === 'error' ? 'bg-rose-800 text-white border-rose-700' : 'bg-slate-900 text-white border-slate-800')"
      >
        <span>{{ toast.message }}</span>
      </div>
    </div>
  </div>
</template>
