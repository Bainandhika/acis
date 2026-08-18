<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useWalletStore } from '../stores/wallet'
import { useAuthStore } from '../stores/auth'
import { useFamilyStore } from '../stores/family'
import { useTransactionStore } from '../stores/transaction'
import { useRouter } from 'vue-router'
import { useI18n } from '../locales'
import { formatRp } from '../composables/useCurrency'
import type { Wallet, CreateWalletPayload, UpdateWalletPayload } from '../services/wallet'
import type { Transaction, CreateTransactionPayload, UpdateTransactionPayload } from '../services/transaction'
import type { FamilyMember } from '../services/family'

// Component imports
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
const isDeleteTxModalOpen = ref(false)
const isChangeRequestModalOpen = ref(false)
const isDeleteRequestModalOpen = ref(false)
const isDeleteMemberModalOpen = ref(false)
const isTelegramModalOpen = ref(false)
const isAllocateModalOpen = ref(false)

// Submitting States
const isWalletSubmitting = ref(false)
const isEditWalletSubmitting = ref(false)
const isDeleteWalletSubmitting = ref(false)
const isTxSubmitting = ref(false)
const isEditTxSubmitting = ref(false)
const isDeleteTxSubmitting = ref(false)
const isChangeRequestSubmitting = ref(false)
const isDeleteRequestSubmitting = ref(false)
const isFamilyManageSubmitting = ref(false)
const isDeleteMemberSubmitting = ref(false)
const isAllocateSubmitting = ref(false)
const isIncomeSubmitting = ref(false)
const isTelegramSubmitting = ref(false)

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
const editMonthlyIncome = ref(0)

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
  wallet_id: string
  type: Transaction['type']
  amount: number
  description: string
}>({
  wallet_id: '',
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

    const limit = w.initial_balance > 0 ? w.initial_balance : (w.minimum_limit > 0 ? w.minimum_limit : 0)
    return {
      wallet: w,
      spent: spent > 0 ? spent : 0,
      limit: limit,
    }
  })
})

const displayWallets = computed(() => walletsWithSpent.value)

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
      name: m.user_name || m.role,
      role: m.role === 'admin' ? 'Admin' : 'Member',
      spent: spent > 0 ? spent : 0,
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
  editFamilyName.value = familyStore.family?.name || ''
  editMonthlyIncome.value = familyStore.family?.monthly_income || 0
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
    showToast(t('toasts.allocationSuccess'), 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || t('toasts.allocationFailed'), 'error')
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
    showToast(t('toasts.walletCreated'), 'success')
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
    showToast(t('toasts.walletUpdated'), 'success')
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
    showToast(t('toasts.walletDeleted'), 'info')
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
      showToast(t('toasts.txRecorded'), 'success')
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
      showToast(t('toasts.txSubmitted'), 'success')
    }
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal menyimpan transaksi', 'error')
  } finally {
    isTxSubmitting.value = false
  }
}

// Transaction Edit / Delete Modals (Admin)
const handleSelectTransaction = (tx: Transaction) => {
  if (isAdmin.value) {
    selectedTxForEdit.value = tx
    editTxForm.value = {
      id: tx.id,
      wallet_id: tx.wallet_id,
      type: tx.type,
      amount: tx.amount,
      description: tx.description || '',
    }
    isEditTxModalOpen.value = true
  } else {
    selectedTxForChangeRequest.value = tx
    changeRequestForm.value = {
      wallet_id: tx.wallet_id,
      type: tx.type,
      amount: tx.amount,
      description: '',
    }
    isChangeRequestModalOpen.value = true
  }
}

const handleSubmitEditTx = async () => {
  if (!editTxForm.value.id || editTxForm.value.amount <= 0) return
  isEditTxSubmitting.value = true
  try {
    const payload: UpdateTransactionPayload = {
      wallet_id: editTxForm.value.wallet_id || undefined,
      type: editTxForm.value.type,
      amount: editTxForm.value.amount,
      description: editTxForm.value.description || undefined,
    }
    await txStore.editTransaction(editTxForm.value.id, payload)
    await walletStore.fetchWallets()
    isEditTxModalOpen.value = false
    showToast(t('toasts.txUpdated'), 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to update transaction', 'error')
  } finally {
    isEditTxSubmitting.value = false
  }
}

const openDeleteTxModal = (tx: Transaction) => {
  deleteTxTarget.value = tx
  isDeleteTxModalOpen.value = true
}

const handleConfirmDeleteTx = async () => {
  if (!deleteTxTarget.value) return
  isDeleteTxSubmitting.value = true
  try {
    await txStore.removeTransaction(deleteTxTarget.value.id)
    await walletStore.fetchWallets()
    isDeleteTxModalOpen.value = false
    deleteTxTarget.value = null
    showToast(t('toasts.txDeleted'), 'info')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to delete transaction', 'error')
  } finally {
    isDeleteTxSubmitting.value = false
  }
}

// Member Proposals: Change & Delete Request
const handleSubmitChangeRequest = async () => {
  if (!selectedTxForChangeRequest.value || changeRequestForm.value.amount <= 0) return
  isChangeRequestSubmitting.value = true
  try {
    await txStore.addProposal({
      wallet_id: changeRequestForm.value.wallet_id || selectedTxForChangeRequest.value.wallet_id,
      title: `Revisi: ${changeRequestForm.value.description || selectedTxForChangeRequest.value.description || 'Transaksi'}`,
      amount: changeRequestForm.value.amount,
      description: changeRequestForm.value.description || 'Permintaan perubahan transaksi',
      request_type: 'edit_transaction',
      target_transaction_id: selectedTxForChangeRequest.value.id,
      payload: {
        wallet_id: changeRequestForm.value.wallet_id,
        type: changeRequestForm.value.type,
        amount: changeRequestForm.value.amount,
        description: changeRequestForm.value.description,
      },
    })
    isChangeRequestModalOpen.value = false
    showToast(t('toasts.changeRequested'), 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal mengajukan perubahan', 'error')
  } finally {
    isChangeRequestSubmitting.value = false
  }
}

const openDeleteRequestModal = (tx: Transaction) => {
  selectedTxForDeleteRequest.value = tx
  deleteRequestReason.value = ''
  isDeleteRequestModalOpen.value = true
}

const handleSubmitDeleteRequest = async () => {
  if (!selectedTxForDeleteRequest.value) return
  isDeleteRequestSubmitting.value = true
  try {
    await txStore.addProposal({
      wallet_id: selectedTxForDeleteRequest.value.wallet_id,
      title: `Hapus Transaksi: ${selectedTxForDeleteRequest.value.description || formatRp(selectedTxForDeleteRequest.value.amount)}`,
      amount: selectedTxForDeleteRequest.value.amount,
      description: deleteRequestReason.value || 'Permintaan penghapusan transaksi',
      request_type: 'delete_transaction',
      target_transaction_id: selectedTxForDeleteRequest.value.id,
    })
    isDeleteRequestModalOpen.value = false
    selectedTxForDeleteRequest.value = null
    showToast(t('toasts.deleteRequested'), 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal mengajukan penghapusan', 'error')
  } finally {
    isDeleteRequestSubmitting.value = false
  }
}

// Settings & Family Handlers
const handleUpdateFamilyName = async () => {
  if (!editFamilyName.value.trim()) return
  isFamilyManageSubmitting.value = true
  try {
    await familyStore.handleUpdateFamilyName(editFamilyName.value.trim())
    showToast(t('toasts.familyUpdated'), 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal mengubah nama keluarga', 'error')
  } finally {
    isFamilyManageSubmitting.value = false
  }
}

const handleUpdateMonthlyIncome = async () => {
  if (editMonthlyIncome.value < 0) return
  isIncomeSubmitting.value = true
  try {
    await familyStore.handleUpdateSettings(editMonthlyIncome.value)
    showToast(t('toasts.incomeSaved'), 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || t('toasts.incomeFailed'), 'error')
  } finally {
    isIncomeSubmitting.value = false
  }
}

const handleDisconnectTelegram = async () => {
  isTelegramSubmitting.value = true
  try {
    await familyStore.handleDisconnectTelegram()
    showToast(t('toasts.botDisconnected'), 'info')
  } catch (err: any) {
    showToast(err.response?.data?.error || t('toasts.botDisconnectFailed'), 'error')
  } finally {
    isTelegramSubmitting.value = false
  }
}

const openDeleteMemberModal = (member: FamilyMember) => {
  deleteMemberTarget.value = member
  isDeleteMemberModalOpen.value = true
}

const handleConfirmDeleteMember = async () => {
  if (!deleteMemberTarget.value) return
  isDeleteMemberSubmitting.value = true
  try {
    await familyStore.handleRemoveMember(deleteMemberTarget.value.id)
    isDeleteMemberModalOpen.value = false
    deleteMemberTarget.value = null
    showToast(t('toasts.memberRemoved'), 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || t('toasts.memberRemoveFailed'), 'error')
  } finally {
    isDeleteMemberSubmitting.value = false
  }
}

const copyInviteCode = async () => {
  const code = familyStore.family?.invite_code
  if (code && typeof navigator !== 'undefined') {
    await navigator.clipboard.writeText(code)
    showToast(t('toasts.codeCopied'), 'info')
  }
}

const approveProp = async (id: string) => {
  try {
    await txStore.handleApprove(id)
    await walletStore.fetchWallets()
    showToast(t('toasts.requestApproved'), 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to approve request', 'error')
  }
}

const rejectProp = async (id: string) => {
  try {
    await txStore.handleReject(id)
    showToast(t('toasts.requestRejected'), 'info')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to reject request', 'error')
  }
}

const handleExportCsv = () => {
  const headers = ['Tanggal', 'Dompet', 'Tipe', 'Nominal', 'Keterangan']
  const rows = (txStore.transactions || []).map(t => [
    new Date(t.created_at).toLocaleDateString('id-ID'),
    getWalletName(t.wallet_id),
    t.type,
    t.amount,
    `"${(t.description || '').replace(/"/g, '""')}"`
  ])
  const csv = [headers.join(','), ...rows.map(r => r.join(','))].join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `acis-transactions-${txStore.selectedYear}-${txStore.selectedMonth}.csv`
  a.click()
  URL.revokeObjectURL(url)
  showToast(t('extra.exportCsv') + ' selesai', 'success')
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
      @open-settings="activeTab = 'settings'"
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
                :total-funds="totalFamilyFunds"
                :trend-percentage="0"
                :is-admin="isAdmin"
                @quick-allocate="openAllocateModal()"
                @transfer-money="openTxModal()"
              />
            </div>

            <!-- Right Card: Family Spending Breakdown (5 cols) -->
            <div class="lg:col-span-5">
              <FamilySpendingCard
                :members="familySpendingList"
                @manage-family="activeTab = 'members'"
              />
            </div>
          </div>

          <!-- Middle Section: Allocated Wallets -->
          <section class="flex flex-col gap-4">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-bold text-slate-900">
                {{ t('dashboard.wallets.title') }}
              </h3>
              <div class="flex items-center gap-3">
                <span class="text-xs font-semibold text-slate-400">
                  {{ displayWallets.length }} {{ t('extra.activeWallets') }}
                </span>
                <button
                  v-if="isAdmin"
                  @click="openWalletModal"
                  class="text-xs font-bold text-teal-700 hover:text-teal-800 transition hover:underline cursor-pointer"
                >
                  {{ t('extra.addWallet') }}
                </button>
              </div>
            </div>

            <!-- Empty State Wallets -->
            <div v-if="displayWallets.length === 0" class="bg-white rounded-3xl p-8 border border-slate-100 shadow-sm text-center flex flex-col items-center justify-center gap-3">
              <p class="text-slate-400 text-xs font-semibold">{{ t('dashboard.wallets.noWallets') }}</p>
              <button
                v-if="isAdmin"
                @click="openWalletModal"
                class="px-4 py-2 bg-teal-700 hover:bg-teal-800 text-white rounded-xl text-xs font-bold transition shadow-sm cursor-pointer"
              >
                {{ t('extra.createFirstWallet') }}
              </button>
            </div>

            <!-- 3-Column Wallets Grid -->
            <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
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
              @export-csv="handleExportCsv"
              @select-transaction="handleSelectTransaction"
            />
          </section>
        </div>

        <!-- ========================================================================= -->
        <!-- TAB: WALLETS                                                              -->
        <!-- ========================================================================= -->
        <div v-else-if="activeTab === 'wallets'" class="flex flex-col gap-6">
          <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
            <div>
              <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">{{ t('walletsTab.title') }}</h2>
              <p class="text-xs text-slate-500 mt-0.5">{{ t('walletsTab.subtitle') }}</p>
            </div>
            <button
              v-if="isAdmin"
              class="px-4 py-2.5 bg-teal-700 hover:bg-teal-800 text-white rounded-xl text-xs font-bold transition shadow-sm flex items-center gap-2 cursor-pointer"
              @click="openWalletModal"
            >
              <span>{{ t('extra.addWallet') }}</span>
            </button>
          </div>

          <!-- Empty State Wallets -->
          <div v-if="displayWallets.length === 0" class="bg-white rounded-3xl p-12 border border-slate-100 shadow-sm text-center flex flex-col items-center justify-center gap-3">
            <p class="text-slate-400 text-xs font-semibold">{{ t('walletsTab.noWallets') }}</p>
            <button
              v-if="isAdmin"
              @click="openWalletModal"
              class="px-4 py-2 bg-teal-700 hover:bg-teal-800 text-white rounded-xl text-xs font-bold transition shadow-sm cursor-pointer"
            >
              {{ t('extra.createFirstWallet') }}
            </button>
          </div>

          <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
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
              <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">{{ t('transaksiTab.title') }}</h2>
              <p class="text-xs text-slate-500 mt-0.5">{{ t('transaksiTab.subtitle') }}</p>
            </div>
            <button
              @click="openTxModal"
              class="px-4 py-2.5 bg-teal-700 hover:bg-teal-800 text-white rounded-xl text-xs font-bold transition shadow-sm flex items-center gap-2 cursor-pointer"
            >
              <span>+ {{ isAdmin ? t('transaksiTab.recordBtn') : t('extra.sendProposalBtn') }}</span>
            </button>
          </div>

          <!-- Admin Pending Proposals Banner -->
          <div v-if="isAdmin && pendingProposals.length > 0" class="bg-amber-50 border border-amber-200 rounded-3xl p-6 flex flex-col gap-4">
            <div class="flex items-center justify-between">
              <h3 class="font-bold text-sm text-amber-900 flex items-center gap-2">
                <span>{{ t('extra.pendingApproval') }}</span>
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
                  <span class="text-sm font-bold text-slate-900">{{ formatRp(p.amount) }}</span>
                  <div class="flex gap-2">
                    <button @click="approveProp(p.id)" class="px-3 py-1 bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold rounded-lg cursor-pointer">{{ t('extra.approve') }}</button>
                    <button @click="rejectProp(p.id)" class="px-3 py-1 bg-rose-600 hover:bg-rose-700 text-white text-xs font-bold rounded-lg cursor-pointer">{{ t('extra.reject') }}</button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Transactions List with Admin Action Menu -->
          <div class="bg-white rounded-3xl p-6 sm:p-7 border border-slate-100/90 shadow-sm flex flex-col gap-6">
            <div class="flex items-center justify-between">
              <h3 class="text-base sm:text-lg font-bold text-slate-900">{{ t('transaksiTab.title') }}</h3>
              <button @click="handleExportCsv" class="px-3.5 py-1.5 rounded-xl border border-slate-200 text-xs font-semibold text-slate-700 hover:bg-slate-50 transition cursor-pointer">
                {{ t('extra.exportCsv') }}
              </button>
            </div>

            <div v-if="txStore.transactions.length === 0" class="text-center py-10 text-xs text-slate-400 font-medium">
              {{ t('transaksiTab.noTransactions') }}
            </div>

            <div v-else class="divide-y divide-slate-100">
              <div
                v-for="tx in txStore.transactions"
                :key="tx.id"
                class="py-3.5 flex items-center justify-between gap-4 hover:bg-slate-50/80 px-2 rounded-2xl transition"
              >
                <div class="flex items-center gap-4 min-w-0">
                  <span class="text-xs font-semibold text-slate-400 w-14 shrink-0">
                    {{ new Date(tx.created_at).toLocaleDateString('id-ID', { month: 'short', day: 'numeric' }) }}
                  </span>
                  <div class="flex flex-col min-w-0">
                    <span class="text-xs sm:text-sm font-bold text-slate-900 truncate">
                      {{ tx.description || 'Transaction' }}
                    </span>
                    <span class="text-[11px] text-slate-400 font-medium">
                      {{ getWalletName(tx.wallet_id) }}
                    </span>
                  </div>
                </div>

                <div class="flex items-center gap-4 shrink-0">
                  <span
                    class="text-xs sm:text-sm font-black font-sans text-right"
                    :class="tx.type === 'income' ? 'text-emerald-600' : 'text-slate-900'"
                  >
                    {{ tx.type === 'income' ? '+' : '-' }}{{ formatRp(tx.amount) }}
                  </span>

                  <!-- Actions Dropdown / Buttons -->
                  <div class="flex items-center gap-1.5">
                    <button
                      v-if="isAdmin"
                      @click="handleSelectTransaction(tx)"
                      class="px-2.5 py-1 text-xs font-bold text-slate-600 hover:bg-slate-100 rounded-lg transition cursor-pointer"
                    >
                      {{ t('extra.contextEdit') }}
                    </button>
                    <button
                      v-if="isAdmin"
                      @click="openDeleteTxModal(tx)"
                      class="px-2.5 py-1 text-xs font-bold text-rose-600 hover:bg-rose-50 rounded-lg transition cursor-pointer"
                    >
                      {{ t('extra.deleteBtn') }}
                    </button>

                    <!-- Member Action Buttons -->
                    <button
                      v-if="!isAdmin"
                      @click="handleSelectTransaction(tx)"
                      class="px-2.5 py-1 text-xs font-bold text-slate-600 hover:bg-slate-100 rounded-lg transition cursor-pointer"
                    >
                      {{ t('toasts.changeRequested') }}
                    </button>
                    <button
                      v-if="!isAdmin"
                      @click="openDeleteRequestModal(tx)"
                      class="px-2.5 py-1 text-xs font-bold text-rose-600 hover:bg-rose-50 rounded-lg transition cursor-pointer"
                    >
                      {{ t('toasts.deleteRequested') }}
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- ========================================================================= -->
        <!-- TAB: FAMILY MEMBERS / SETTINGS                                            -->
        <!-- ========================================================================= -->
        <div v-else-if="activeTab === 'members' || activeTab === 'settings'" class="flex flex-col gap-6 max-w-2xl">
          <div class="flex items-center justify-between">
            <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">{{ t('modals.familyManage.title') }}</h2>
          </div>

          <!-- Family Room General Details -->
          <div class="bg-white rounded-3xl p-6 sm:p-7 border border-slate-100 shadow-sm flex flex-col gap-6">
            <!-- Family Name -->
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1.5">{{ t('extra.familyNameLabel') }}</label>
              <div class="flex gap-2">
                <input
                  type="text"
                  v-model="editFamilyName"
                  :disabled="!isAdmin"
                  class="flex-1 px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-900 focus:outline-none focus:ring-2 focus:ring-teal-700"
                />
                <button
                  v-if="isAdmin"
                  @click="handleUpdateFamilyName"
                  :disabled="isFamilyManageSubmitting"
                  class="px-4 py-2.5 bg-teal-700 hover:bg-teal-800 text-white text-xs font-bold rounded-xl transition cursor-pointer"
                >
                  {{ t('extra.saveBtn') }}
                </button>
              </div>
            </div>

            <!-- Monthly Income Setting (Admin) -->
            <div v-if="isAdmin">
              <label class="text-xs font-bold text-slate-700 block mb-1.5">{{ t('extra.monthlyIncomeLabel') }}</label>
              <div class="flex gap-2">
                <input
                  type="number"
                  v-model.number="editMonthlyIncome"
                  class="flex-1 px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-900 focus:outline-none focus:ring-2 focus:ring-teal-700"
                />
                <button
                  @click="handleUpdateMonthlyIncome"
                  :disabled="isIncomeSubmitting"
                  class="px-4 py-2.5 bg-teal-700 hover:bg-teal-800 text-white text-xs font-bold rounded-xl transition cursor-pointer"
                >
                  {{ isIncomeSubmitting ? t('extra.saving') : t('extra.saveBtn') }}
                </button>
              </div>
            </div>

            <!-- Invite Code -->
            <div class="border-t border-slate-100 pt-4 flex items-center justify-between">
              <div>
                <span class="text-xs font-bold text-slate-700 block">{{ t('extra.inviteCodeLabel') }}</span>
                <span class="text-sm font-mono font-black text-teal-700">{{ familyStore.family?.invite_code }}</span>
              </div>
              <button
                @click="copyInviteCode"
                class="px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-bold rounded-xl transition cursor-pointer"
              >
                {{ t('extra.copyCodeBtn') }}
              </button>
            </div>

            <!-- Telegram Integration Section -->
            <div class="border-t border-slate-100 pt-4 flex flex-col gap-3">
              <span class="text-xs font-bold text-slate-700 block">{{ t('extra.telegramIntegration') }}</span>
              <div v-if="familyStore.family?.telegram_chat_id" class="flex items-center justify-between p-3 rounded-2xl bg-emerald-50 border border-emerald-200">
                <div class="flex flex-col">
                  <span class="text-xs font-bold text-emerald-900">{{ t('extra.telegramConnected') }}</span>
                  <span class="text-[11px] text-emerald-700 font-mono">Chat ID: {{ familyStore.family.telegram_chat_id }}</span>
                </div>
                <button
                  v-if="isAdmin"
                  @click="handleDisconnectTelegram"
                  :disabled="isTelegramSubmitting"
                  class="px-3 py-1.5 bg-rose-600 hover:bg-rose-700 text-white text-xs font-bold rounded-xl transition cursor-pointer"
                >
                  {{ isTelegramSubmitting ? '...' : t('extra.disconnectTelegramBtn') }}
                </button>
              </div>
              <div v-else class="flex items-center justify-between p-3 rounded-2xl bg-slate-50 border border-slate-200">
                <span class="text-xs text-slate-500 font-medium">{{ t('extra.telegramNotConnected') }}</span>
                <button
                  @click="isTelegramModalOpen = true"
                  class="px-3 py-1.5 bg-slate-800 hover:bg-slate-900 text-white text-xs font-bold rounded-xl transition cursor-pointer"
                >
                  {{ t('extra.connectTelegramBtn') }}
                </button>
              </div>
            </div>
          </div>

          <!-- Members List & Management -->
          <div class="bg-white rounded-3xl p-6 sm:p-7 border border-slate-100 shadow-sm flex flex-col gap-4">
            <h3 class="font-bold text-base text-slate-900">{{ t('dashboard.familyMembers.title') }}</h3>
            <div class="divide-y divide-slate-100">
              <div
                v-for="m in familyStore.family?.members || []"
                :key="m.id"
                class="py-3 flex items-center justify-between"
              >
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-full bg-slate-200 text-slate-700 font-bold text-xs flex items-center justify-center">
                    {{ (m.user_name || m.role).charAt(0).toUpperCase() }}
                  </div>
                  <div class="flex flex-col">
                    <span class="text-xs font-bold text-slate-900">{{ m.user_name || m.user_id }}</span>
                    <span class="text-[10px] text-slate-400 capitalize font-medium">{{ m.role }}</span>
                  </div>
                </div>
                <button
                  v-if="isAdmin && m.user_id !== authStore.user?.id"
                  @click="openDeleteMemberModal(m)"
                  class="px-3 py-1 bg-rose-50 hover:bg-rose-100 text-rose-600 text-xs font-bold rounded-lg transition cursor-pointer"
                >
                  {{ t('extra.removeMemberBtn') }}
                </button>
              </div>
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
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('extra.allocateModalTitle') }}</h3>
        <p class="text-xs text-slate-500 mb-4">{{ t('extra.allocateModalDesc') }}</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.selectTargetWallet') }}</label>
            <select v-model="allocateForm.wallet_id" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700">
              <option v-for="w in walletStore.wallets" :key="w.id" :value="w.id">{{ w.name }}</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.allocationAmount') }}</label>
            <input type="number" v-model.number="allocateForm.amount" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.memoOptional') }}</label>
            <input type="text" v-model="allocateForm.description" :placeholder="t('extra.allocationMemoPlaceholder')" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold text-slate-500" @click="closeAllocateModal" :disabled="isAllocateSubmitting">{{ t('extra.cancel') }}</button>
          <button class="btn bg-teal-700 hover:bg-teal-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitAllocate" :disabled="isAllocateSubmitting || allocateForm.amount <= 0">
            {{ isAllocateSubmitting ? t('extra.allocating') : t('extra.allocateFundsBtn') }}
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
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('extra.createWalletModalTitle') }}</h3>
        <p class="text-xs text-slate-500 mb-4">{{ t('extra.createWalletModalDesc') }}</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.walletNameLabel') }}</label>
            <input type="text" v-model="newWallet.name" :placeholder="t('extra.walletNamePlaceholder')" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.descLabel') }}</label>
            <input type="text" v-model="newWallet.description" :placeholder="t('extra.descPlaceholder')" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.initialBalanceLabel') }}</label>
              <input type="number" v-model.number="newWallet.initial_balance" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
            </div>
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.minLimitLabel') }}</label>
              <input type="number" v-model.number="newWallet.minimum_limit" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
            </div>
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold text-slate-500" @click="closeWalletModal" :disabled="isWalletSubmitting">{{ t('extra.cancel') }}</button>
          <button class="btn bg-teal-700 hover:bg-teal-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitWallet" :disabled="isWalletSubmitting || !newWallet.name">
            {{ isWalletSubmitting ? t('extra.saving') : t('extra.saveWalletBtn') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeWalletModal">close</button>
      </form>
    </dialog>

    <!-- Edit Wallet Modal -->
    <dialog :class="isEditWalletModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white text-slate-900 rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('modals.editWallet.title') }}</h3>
        <p class="text-xs text-slate-500 mb-4">{{ t('modals.editWallet.subtitle') }}</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.walletNameLabel') }}</label>
            <input type="text" v-model="editWalletData.name" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.descLabel') }}</label>
            <input type="text" v-model="editWalletData.description" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.minLimitLabel') }}</label>
            <input type="number" v-model.number="editWalletData.minimum_limit" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold text-slate-500" @click="closeEditWalletModal" :disabled="isEditWalletSubmitting">{{ t('extra.cancel') }}</button>
          <button class="btn bg-teal-700 hover:bg-teal-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitEditWallet" :disabled="isEditWalletSubmitting || !editWalletData.name">
            {{ isEditWalletSubmitting ? t('extra.saving') : t('extra.saveBtn') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeEditWalletModal">close</button>
      </form>
    </dialog>

    <!-- Delete Wallet Confirm Modal -->
    <dialog :class="isDeleteWalletModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white text-slate-900 rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('modals.confirmDelete.titleWallet') }}</h3>
        <p class="text-xs text-slate-500 mb-4">{{ t('modals.confirmDelete.descWallet') }}</p>
        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold text-slate-500" @click="closeDeleteWalletModal" :disabled="isDeleteWalletSubmitting">{{ t('extra.cancel') }}</button>
          <button class="btn bg-rose-600 hover:bg-rose-700 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleConfirmDeleteWallet" :disabled="isDeleteWalletSubmitting">
            {{ isDeleteWalletSubmitting ? t('extra.deleting') : t('modals.confirmDelete.confirm') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeDeleteWalletModal">close</button>
      </form>
    </dialog>

    <!-- Create Transaction Modal -->
    <dialog :class="isTxModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white text-slate-900 rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ isAdmin ? t('extra.recordTxModalTitle') : t('extra.proposeTxModalTitle') }}</h3>
        <p class="text-xs text-slate-500 mb-4">{{ isAdmin ? t('extra.recordTxModalDesc') : t('extra.proposeTxModalDesc') }}</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.selectWalletLabel') }}</label>
            <select v-model="newTx.wallet_id" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700">
              <option v-for="w in walletStore.wallets" :key="w.id" :value="w.id">{{ w.name }}</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.txTypeLabel') }}</label>
            <select v-model="newTx.type" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700">
              <option value="expense">{{ t('extra.expenseOption') }}</option>
              <option value="income">{{ t('extra.incomeOption') }}</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.amountLabel') }}</label>
            <input type="number" v-model.number="newTx.amount" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.notesLabel') }}</label>
            <input type="text" v-model="newTx.description" :placeholder="t('extra.notesPlaceholder')" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold text-slate-500" @click="closeTxModal" :disabled="isTxSubmitting">{{ t('extra.cancel') }}</button>
          <button class="btn bg-teal-700 hover:bg-teal-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitTx" :disabled="isTxSubmitting || newTx.amount <= 0">
            {{ isTxSubmitting ? t('extra.processing') : (isAdmin ? t('extra.saveTxBtn') : t('extra.sendProposalBtn')) }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeTxModal">close</button>
      </form>
    </dialog>

    <!-- Edit Transaction Modal (Admin) -->
    <dialog :class="isEditTxModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white text-slate-900 rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('extra.editTxTitle') }}</h3>
        <div class="flex flex-col gap-3.5 mt-4">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.selectWalletLabel') }}</label>
            <select v-model="editTxForm.wallet_id" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700">
              <option v-for="w in walletStore.wallets" :key="w.id" :value="w.id">{{ w.name }}</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.txTypeLabel') }}</label>
            <select v-model="editTxForm.type" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700">
              <option value="expense">{{ t('extra.expenseOption') }}</option>
              <option value="income">{{ t('extra.incomeOption') }}</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.amountLabel') }}</label>
            <input type="number" v-model.number="editTxForm.amount" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.notesLabel') }}</label>
            <input type="text" v-model="editTxForm.description" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold text-slate-500" @click="isEditTxModalOpen = false" :disabled="isEditTxSubmitting">{{ t('extra.cancel') }}</button>
          <button class="btn bg-teal-700 hover:bg-teal-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitEditTx" :disabled="isEditTxSubmitting || editTxForm.amount <= 0">
            {{ isEditTxSubmitting ? t('extra.saving') : t('extra.saveBtn') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="isEditTxModalOpen = false">close</button>
      </form>
    </dialog>

    <!-- Delete Transaction Confirm Modal -->
    <dialog :class="isDeleteTxModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white text-slate-900 rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('extra.deleteTxTitle') }}</h3>
        <p class="text-xs text-slate-500 mb-4">{{ t('extra.confirmDeleteTx') }}</p>
        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold text-slate-500" @click="isDeleteTxModalOpen = false" :disabled="isDeleteTxSubmitting">{{ t('extra.cancel') }}</button>
          <button class="btn bg-rose-600 hover:bg-rose-700 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleConfirmDeleteTx" :disabled="isDeleteTxSubmitting">
            {{ isDeleteTxSubmitting ? t('extra.deleting') : t('extra.deleteBtn') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="isDeleteTxModalOpen = false">close</button>
      </form>
    </dialog>

    <!-- Change Request Modal (Member) -->
    <dialog :class="isChangeRequestModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white text-slate-900 rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('extra.requestChangeTitle') }}</h3>
        <div class="flex flex-col gap-3.5 mt-4">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.amountLabel') }}</label>
            <input type="number" v-model.number="changeRequestForm.amount" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.reasonLabel') }}</label>
            <input type="text" v-model="changeRequestForm.description" :placeholder="t('extra.reasonPlaceholder')" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold text-slate-500" @click="isChangeRequestModalOpen = false" :disabled="isChangeRequestSubmitting">{{ t('extra.cancel') }}</button>
          <button class="btn bg-teal-700 hover:bg-teal-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitChangeRequest" :disabled="isChangeRequestSubmitting || changeRequestForm.amount <= 0">
            {{ isChangeRequestSubmitting ? t('extra.processing') : t('extra.sendProposalBtn') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="isChangeRequestModalOpen = false">close</button>
      </form>
    </dialog>

    <!-- Delete Request Modal (Member) -->
    <dialog :class="isDeleteRequestModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white text-slate-900 rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('extra.requestDeleteTitle') }}</h3>
        <div class="flex flex-col gap-3.5 mt-4">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('extra.reasonLabel') }}</label>
            <input type="text" v-model="deleteRequestReason" :placeholder="t('extra.reasonPlaceholder')" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-teal-700" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold text-slate-500" @click="isDeleteRequestModalOpen = false" :disabled="isDeleteRequestSubmitting">{{ t('extra.cancel') }}</button>
          <button class="btn bg-rose-600 hover:bg-rose-700 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitDeleteRequest" :disabled="isDeleteRequestSubmitting">
            {{ isDeleteRequestSubmitting ? t('extra.processing') : t('extra.sendProposalBtn') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="isDeleteRequestModalOpen = false">close</button>
      </form>
    </dialog>

    <!-- Delete Member Confirm Modal -->
    <dialog :class="isDeleteMemberModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white text-slate-900 rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('modals.familyManage.confirmRemoveTitle') }}</h3>
        <p class="text-xs text-slate-500 mb-4">{{ t('modals.familyManage.confirmRemoveDesc', { name: deleteMemberTarget?.user_name || 'Member' }) }}</p>
        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold text-slate-500" @click="isDeleteMemberModalOpen = false" :disabled="isDeleteMemberSubmitting">{{ t('extra.cancel') }}</button>
          <button class="btn bg-rose-600 hover:bg-rose-700 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleConfirmDeleteMember" :disabled="isDeleteMemberSubmitting">
            {{ isDeleteMemberSubmitting ? t('extra.deleting') : t('extra.removeMemberBtn') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="isDeleteMemberModalOpen = false">close</button>
      </form>
    </dialog>

    <!-- Telegram Link Modal -->
    <dialog :class="isTelegramModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white text-slate-900 rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('modals.telegram.title') }}</h3>
        <p class="text-xs text-slate-500 mb-4">{{ t('modals.telegram.subtitle') }}</p>
        <div class="flex flex-col gap-2.5 text-xs text-slate-700 bg-slate-50 p-4 rounded-2xl border border-slate-200">
          <p class="font-bold">{{ t('modals.telegram.howToLink') }}</p>
          <p>1. {{ t('modals.telegram.step1') }}</p>
          <p>2. {{ t('modals.telegram.step2') }} <span class="font-mono font-bold bg-white px-2 py-0.5 rounded border border-slate-200 text-teal-700">/link {{ familyStore.family?.invite_code }}</span></p>
          <p>3. {{ t('modals.telegram.step3') }}</p>
        </div>
        <div class="modal-action mt-6">
          <button class="btn bg-teal-700 hover:bg-teal-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="isTelegramModalOpen = false">{{ t('modals.telegram.close') }}</button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="isTelegramModalOpen = false">close</button>
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
