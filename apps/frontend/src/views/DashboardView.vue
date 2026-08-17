<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useWalletStore } from '../stores/wallet'
import { useAuthStore } from '../stores/auth'
import { useFamilyStore } from '../stores/family'
import { useTransactionStore } from '../stores/transaction'
import { useRouter } from 'vue-router'
import { useI18n } from '../locales'
import type { Wallet, CreateWalletPayload, UpdateWalletPayload } from '../services/wallet'
import type { Transaction, CreateTransactionPayload, CreateProposalPayload } from '../services/transaction'
import type { FamilyMember } from '../services/family'

// Component imports
import TopNavbar, { type NavTab } from '../components/TopNavbar.vue'
import CashflowChart from '../components/CashflowChart.vue'

const walletStore = useWalletStore()
const authStore = useAuthStore()
const familyStore = useFamilyStore()
const txStore = useTransactionStore()
const router = useRouter()
const { t } = useI18n()

// UI State: Horizontal Navigation Tabs
const activeTab = ref<NavTab>('dashboard')
const selectedWalletId = ref('')

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

// Independent Submitting States
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

// Role Check
const isAdmin = computed(() => authStore.user?.role === 'admin')

// Form States
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
  type: 'income' | 'expense'
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
  type: 'income' | 'expense'
  amount: number
  description: string
}>({
  type: 'expense',
  amount: 0,
  description: '',
})

const deleteRequestReason = ref('')

// Metrics & Analytics Calculations (Strictly from real data, initialized to 0)
const totalBalance = computed(() => {
  return (walletStore.wallets || []).reduce((sum, w) => sum + (w?.current_balance || 0), 0)
})

// Filter transactions for the last 7 days
const sevenDaysTransactions = computed(() => {
  const cutoff = new Date()
  cutoff.setDate(cutoff.getDate() - 7)
  return (txStore.transactions || []).filter(t => {
    if (!t?.created_at) return false
    return new Date(t.created_at) >= cutoff
  })
})

const sevenDaysIncome = computed(() => {
  return sevenDaysTransactions.value
    .filter(t => t?.type === 'income')
    .reduce((sum, t) => sum + (t?.amount || 0), 0)
})

const sevenDaysExpense = computed(() => {
  return sevenDaysTransactions.value
    .filter(t => t?.type === 'expense')
    .reduce((sum, t) => sum + (t?.amount || 0), 0)
})

// Expense Analysis by Wallet (last 7 days strictly from real data)
const walletPalette = [
  'bg-amber-400',
  'bg-emerald-400',
  'bg-sky-400',
  'bg-lime-400',
  'bg-rose-400',
  'bg-purple-400',
  'bg-slate-300'
]

const expenseAnalysis = computed(() => {
  const expenseTxs = sevenDaysTransactions.value.filter(t => t?.type === 'expense')
  const total = expenseTxs.reduce((sum, t) => sum + (t?.amount || 0), 0)
  
  if (total === 0 || expenseTxs.length === 0) {
    return {
      total: 0,
      categories: []
    }
  }

  const walletMap: Record<string, number> = {}
  expenseTxs.forEach(t => {
    const wName = getWalletName(t.wallet_id)
    walletMap[wName] = (walletMap[wName] || 0) + (t.amount || 0)
  })

  const categories = Object.entries(walletMap).map(([name, amount], index) => {
    const percentage = Math.round((amount / total) * 100)
    return {
      name,
      amount,
      percentage,
      color: walletPalette[index % walletPalette.length]
    }
  }).sort((a, b) => b.amount - a.amount)

  return { total, categories }
})

// Real Family Members from database (NO dummy members)
const realFamilyMembers = computed(() => {
  return familyStore.family?.members || []
})

// Recent History (Last 10 dated transactions)
const recentTenTransactions = computed(() => {
  return (txStore.transactions || []).slice(0, 10)
})

// Indonesian Rupiah Currency Formatter (e.g. Rp 1.500.000)
const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount || 0)
}

const getWalletName = (id?: string) => {
  if (!id) return t('dashboard.wallets.envelope')
  const w = (walletStore.wallets || []).find(item => item?.id === id)
  return w ? w.name : t('dashboard.wallets.envelope')
}

onMounted(async () => {
  await familyStore.fetchMyFamily()
  if (!familyStore.family) {
    router.push('/family-setup')
    return
  }
  await walletStore.fetchWallets()
  if (walletStore.wallets && walletStore.wallets.length > 0 && walletStore.wallets[0]) {
    selectedWalletId.value = walletStore.wallets[0].id
  }
  txStore.fetchTransactions()
  txStore.fetchProposals()
})

// Toast Notification System
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

// Copy Invite Code helper for Family Widget
const isInviteCodeCopied = ref(false)
const copyFamilyInviteCode = async () => {
  const code = familyStore.family?.invite_code
  if (!code) return
  try {
    await navigator.clipboard.writeText(code)
    isInviteCodeCopied.value = true
    setTimeout(() => {
      isInviteCodeCopied.value = false
    }, 2000)
  } catch {
    // Clipboard fallback
  }
}

// Telegram action
const handleDisconnectTelegram = async () => {
  if (confirm('Disconnect Telegram Bot?')) {
    try {
      await familyStore.handleDisconnectTelegram()
      isTelegramModalOpen.value = false
      showToast(t('toasts.botDisconnected'), 'info')
    } catch {
      showToast(t('toasts.botDisconnectFailed'), 'error')
    }
  }
}

// Modal actions - Create Wallet
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

// Modal actions - Edit Wallet
const openEditWalletModal = (w: Wallet) => {
  editWalletData.value = {
    id: w.id,
    name: w.name,
    description: w.description || '',
    minimum_limit: w.minimum_limit || 0,
  }
  isEditWalletModalOpen.value = true
}
const closeEditWalletModal = () => {
  isEditWalletModalOpen.value = false
  editWalletData.value = { id: '', name: '', description: '', minimum_limit: 0 }
}
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

// Modal actions - Delete Wallet
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
    await txStore.fetchTransactions()
    await txStore.fetchProposals()
    closeDeleteWalletModal()
    showToast(t('toasts.walletDeleted'), 'info')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to delete wallet', 'error')
  } finally {
    isDeleteWalletSubmitting.value = false
  }
}

// Modal actions - Family Management & Member Deletion
const openFamilyManageModal = () => {
  editFamilyName.value = familyStore.family?.name || ''
  isFamilyManageModalOpen.value = true
}
const closeFamilyManageModal = () => {
  isFamilyManageModalOpen.value = false
  editFamilyName.value = ''
}
const handleSubmitEditFamilyName = async () => {
  if (!editFamilyName.value.trim()) return
  isFamilyManageSubmitting.value = true
  try {
    await familyStore.handleUpdateFamilyName(editFamilyName.value.trim())
    showToast(t('toasts.familyUpdated'), 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to rename family room', 'error')
  } finally {
    isFamilyManageSubmitting.value = false
  }
}

const openDeleteMemberModal = (member: FamilyMember) => {
  deleteMemberTarget.value = member
  isDeleteMemberModalOpen.value = true
}
const closeDeleteMemberModal = () => {
  isDeleteMemberModalOpen.value = false
  deleteMemberTarget.value = null
}
const handleConfirmDeleteMember = async () => {
  if (!deleteMemberTarget.value) return
  isDeleteMemberSubmitting.value = true
  try {
    await familyStore.handleRemoveMember(deleteMemberTarget.value.id)
    showToast(t('toasts.memberRemoved'), 'success')
    closeDeleteMemberModal()
  } catch (err: any) {
    showToast(err.response?.data?.error || t('toasts.memberRemoveFailed'), 'error')
  } finally {
    isDeleteMemberSubmitting.value = false
  }
}


// Modal actions - Create Transaction / Submit Proposal
const openTxModal = () => {
  if (walletStore.wallets.length > 0 && walletStore.wallets[0]) {
    newTx.value.wallet_id = selectedWalletId.value || walletStore.wallets[0].id
  } else {
    newTx.value.wallet_id = ''
  }
  newTx.value.amount = 0
  newTx.value.description = ''
  newTx.value.type = 'expense'
  isTxModalOpen.value = true
}

const closeTxModal = () => {
  isTxModalOpen.value = false
  newTx.value = { wallet_id: '', type: 'expense', amount: 0, description: '' }
}

const handleSubmitTx = async () => {
  isTxSubmitting.value = true
  try {
    if (isAdmin.value) {
      await txStore.addTransaction(newTx.value)
      await walletStore.fetchWallets()
      closeTxModal()
      showToast(t('toasts.txRecorded'), 'success')
    } else {
      await txStore.addProposal({
        wallet_id: newTx.value.wallet_id,
        title: newTx.value.description || (newTx.value.type === 'income' ? 'Pengajuan Pemasukan' : 'Pengajuan Pengeluaran'),
        amount: newTx.value.amount,
        description: newTx.value.description || '',
        request_type: 'add_transaction',
      })
      closeTxModal()
      showToast(t('toasts.txSubmitted'), 'success')
    }
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to submit transaction', 'error')
  } finally {
    isTxSubmitting.value = false
  }
}

// Direct Edit (Admin)
const openEditTxModal = (tx: Transaction) => {
  selectedTxForEdit.value = tx
  editTxForm.value = {
    id: tx.id,
    wallet_id: tx.wallet_id,
    type: tx.type,
    amount: tx.amount,
    description: tx.description || '',
  }
  isEditTxModalOpen.value = true
}

const closeEditTxModal = () => {
  isEditTxModalOpen.value = false
  selectedTxForEdit.value = null
}

const handleSubmitEditTx = async () => {
  if (!editTxForm.value.id) return
  isEditTxSubmitting.value = true
  try {
    await txStore.editTransaction(editTxForm.value.id, {
      wallet_id: editTxForm.value.wallet_id,
      type: editTxForm.value.type,
      amount: editTxForm.value.amount,
      description: editTxForm.value.description || undefined,
    })
    await walletStore.fetchWallets()
    closeEditTxModal()
    showToast(t('toasts.txUpdated'), 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to update transaction', 'error')
  } finally {
    isEditTxSubmitting.value = false
  }
}

// Change Request (Member)
const openChangeRequestModal = (tx: Transaction) => {
  selectedTxForChangeRequest.value = tx
  changeRequestForm.value = {
    type: tx.type,
    amount: tx.amount,
    description: '',
  }
  isChangeRequestModalOpen.value = true
}

const closeChangeRequestModal = () => {
  isChangeRequestModalOpen.value = false
  selectedTxForChangeRequest.value = null
}

const handleSubmitChangeRequest = async () => {
  if (!selectedTxForChangeRequest.value) return
  isChangeRequestSubmitting.value = true
  try {
    const orig = selectedTxForChangeRequest.value
    await txStore.addProposal({
      wallet_id: orig.wallet_id,
      title: `Ubah Transaksi: ${orig.description || formatCurrency(orig.amount)}`,
      amount: changeRequestForm.value.amount,
      description: changeRequestForm.value.description || 'Permintaan perubahan transaksi',
      request_type: 'edit_transaction',
      target_transaction_id: orig.id,
      payload: {
        type: changeRequestForm.value.type,
        amount: changeRequestForm.value.amount,
        description: changeRequestForm.value.description,
      },
    })
    closeChangeRequestModal()
    showToast(t('toasts.changeRequested'), 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to submit change request', 'error')
  } finally {
    isChangeRequestSubmitting.value = false
  }
}

// Delete Request (Member)
const openDeleteRequestModal = (tx: Transaction) => {
  selectedTxForDeleteRequest.value = tx
  deleteRequestReason.value = ''
  isDeleteRequestModalOpen.value = true
}

const closeDeleteRequestModal = () => {
  isDeleteRequestModalOpen.value = false
  selectedTxForDeleteRequest.value = null
}

const handleSubmitDeleteRequest = async () => {
  if (!selectedTxForDeleteRequest.value) return
  isDeleteRequestSubmitting.value = true
  try {
    const orig = selectedTxForDeleteRequest.value
    await txStore.addProposal({
      wallet_id: orig.wallet_id,
      title: `Hapus Transaksi: ${orig.description || formatCurrency(orig.amount)}`,
      amount: orig.amount,
      description: deleteRequestReason.value || 'Permintaan penghapusan transaksi',
      request_type: 'delete_transaction',
      target_transaction_id: orig.id,
    })
    closeDeleteRequestModal()
    showToast(t('toasts.deleteRequested'), 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to submit delete request', 'error')
  } finally {
    isDeleteRequestSubmitting.value = false
  }
}

// Direct Delete (Admin)
const openDeleteTxModal = (tx: Transaction) => {
  deleteTxTarget.value = tx
  isDeleteTxModalOpen.value = true
}

const closeDeleteTxModal = () => {
  isDeleteTxModalOpen.value = false
  deleteTxTarget.value = null
}

const handleConfirmDeleteTx = async () => {
  if (!deleteTxTarget.value) return
  isDeleteTxSubmitting.value = true
  try {
    await txStore.removeTransaction(deleteTxTarget.value.id)
    await walletStore.fetchWallets()
    closeDeleteTxModal()
    showToast(t('toasts.txDeleted'), 'info')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to delete transaction', 'error')
  } finally {
    isDeleteTxSubmitting.value = false
  }
}

const pendingProposals = computed(() => {
  return (txStore.proposals || []).filter(p => p?.status === 'pending')
})

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
</script>

<template>
  <div class="min-h-screen bg-[#F8FAFC] flex flex-col">
    <!-- Top Horizontal Navigation -->
    <TopNavbar 
      :active-tab="activeTab"
      @select-tab="activeTab = $event"
      @open-telegram-modal="isTelegramModalOpen = true"
      @open-family-manage="openFamilyManageModal"
    />

    <!-- Main Container -->
    <main class="flex-1 max-w-[1600px] w-full mx-auto p-4 sm:p-6 lg:p-8">

      <!-- ========================================================================= -->
      <!-- TAB 1: DASHBOARD TAB                                                      -->
      <!-- 1. Family                                                                 -->
      <!-- 2. Overall Financial Summary                                              -->
      <!-- 3. Wallets                                                                -->
      <!-- 4. Allocation                                           -->
      <!-- 5. Recent History                                                         -->
      <!-- ========================================================================= -->
      <div v-if="activeTab === 'dashboard'" class="grid grid-cols-1 lg:grid-cols-2 gap-8 items-start">

        <!-- LEFT COLUMN -->
        <div class="flex flex-col gap-8">

        <!-- 1. WIDGET: Family -->
        <section class="flex flex-col gap-4">
          <div class="card-neo p-6 sm:p-8 flex flex-col gap-6">
            <!-- Top: Large Family Name & Copyable Invite Code -->
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-slate-100">
              <div>
                <div class="flex items-center gap-2">
                  <span class="text-[10px] font-extrabold uppercase px-2.5 py-1 rounded-full bg-brand-100 text-brand-900 tracking-wider">
                    {{ t('dashboard.family.title') }}
                  </span>
                  <button 
                    type="button"
                    @click="openFamilyManageModal"
                    class="p-1 rounded-lg text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition text-xs flex items-center gap-1 font-bold cursor-pointer"
                    :title="t('modals.familyManage.title')"
                  >
                    <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M12 20h9"></path>
                      <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path>
                    </svg>
                    <span class="text-[10px]">{{ t('dashboard.family.edit') }}</span>
                  </button>
                </div>
                <h2 class="text-3xl sm:text-4xl font-black text-slate-900 tracking-tight mt-2">
                  {{ familyStore.family?.name || 'ACIS Family Room' }}
                </h2>
                <p class="text-xs text-slate-400 font-medium mt-1">
                  {{ t('dashboard.family.subtitle') }}
                </p>
              </div>

              <!-- Copyable Invitation Code with copy button beside it -->
              <div class="flex flex-col sm:items-end">
                <span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">
                  {{ t('dashboard.family.inviteCodeLabel') }}
                </span>
                <div class="inline-flex items-center gap-2 mt-1.5 px-3.5 py-2 rounded-2xl bg-slate-50 border border-slate-200 shadow-sm">
                  <span class="font-mono font-black text-base text-slate-900 tracking-wider">
                    {{ familyStore.family?.invite_code || '-' }}
                  </span>
                  <button 
                    v-if="familyStore.family?.invite_code"
                    @click="copyFamilyInviteCode"
                    class="p-1.5 rounded-xl text-slate-500 hover:text-slate-900 hover:bg-slate-200/70 transition active:scale-95 flex items-center justify-center"
                    :title="t('nav.copyCode')"
                    type="button"
                  >
                    <svg v-if="!isInviteCodeCopied" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                    </svg>
                    <span v-else class="text-xs text-emerald-600 font-black">✓</span>
                  </button>
                </div>
              </div>
            </div>

            <!-- Members List by Name and Family Status -->
            <div>
              <div class="flex items-center justify-between mb-4">
                <h3 class="text-xs font-black uppercase tracking-wider text-slate-500">
                  {{ t('dashboard.family.membersLabel') }} ({{ realFamilyMembers.length }})
                </h3>
              </div>

              <div v-if="realFamilyMembers.length === 0" class="py-6 text-center text-xs text-slate-400 font-medium">
                {{ t('dashboard.familyMembers.noMembers') }}
              </div>

              <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
                <div 
                  v-for="member in realFamilyMembers" 
                  :key="member.id"
                  class="p-4 rounded-2xl bg-slate-50/90 border border-slate-200/60 flex items-center justify-between gap-3.5 hover:bg-white hover:border-slate-300/80 hover:shadow-sm transition"
                >
                  <div class="flex items-center gap-3 min-w-0 flex-1">
                    <div class="w-10 h-10 rounded-2xl bg-gradient-to-tr from-slate-900 to-slate-700 text-white font-black text-sm flex items-center justify-center shrink-0 shadow-sm">
                      {{ (member.user_name || (member.user_id === authStore.user?.id ? authStore.user?.username : 'U') || 'U').charAt(0).toUpperCase() }}
                    </div>
                    <div class="min-w-0 flex-1">
                      <div class="flex items-center gap-1.5">
                        <h4 class="text-xs sm:text-sm font-black text-slate-900 truncate">
                          {{ member.user_name || (member.user_id === authStore.user?.id ? (authStore.user?.username || authStore.user?.name) : (`${t('dashboard.familyMembers.member')} #${member.user_id.slice(0, 6)}`)) }}
                        </h4>
                        <span 
                          v-if="member.user_id === authStore.user?.id" 
                          class="text-[9px] font-extrabold uppercase px-1.5 py-0.5 rounded bg-brand-400 text-slate-950 tracking-wider shrink-0"
                        >
                          {{ t('modals.familyManage.youBadge') }}
                        </span>
                      </div>
                      <span class="text-[10px] text-slate-400 font-medium block whitespace-nowrap mt-0.5">
                        {{ t('dashboard.familyMembers.joined') }} {{ new Date(member.joined_at).toLocaleDateString('id-ID') }}
                      </span>
                    </div>
                  </div>
                  <span 
                    class="text-[10px] font-extrabold uppercase px-2.5 py-1 rounded-full shrink-0 whitespace-nowrap"
                    :class="member.role === 'admin' ? 'bg-amber-100 text-amber-900 border border-amber-200' : 'bg-white text-slate-700 border border-slate-200/80'"
                  >
                    {{ member.role === 'admin' ? t('dashboard.familyMembers.adminRole') : t('dashboard.familyMembers.memberRole') }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 2. WIDGET: Overall Financial Summary -->
        <section class="flex flex-col gap-4">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-xl font-extrabold text-slate-900 tracking-tight">
                {{ t('dashboard.financialSummary.title') }}
              </h2>
              <p class="text-xs text-slate-400 font-medium">
                {{ t('dashboard.financialSummary.subtitle') }}
              </p>
            </div>
            <button 
              @click="openTxModal" 
              class="px-4 py-2.5 rounded-2xl bg-slate-900 hover:bg-slate-800 text-white font-bold text-xs transition shadow-sm hover:shadow flex items-center gap-1.5 shrink-0 active:scale-95"
            >
              <svg class="w-4 h-4 text-brand-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                <line x1="12" y1="5" x2="12" y2="19"></line>
                <line x1="5" y1="12" x2="19" y2="12"></line>
              </svg>
              <span>{{ t('dashboard.financialSummary.recordTxBtn') }}</span>
            </button>
          </div>

          <!-- Cashflow Chart with Red Expense Legend and Hover Color -->
          <CashflowChart 
            :transactions="txStore.transactions"
            :total-balance="totalBalance"
            :total-income="sevenDaysIncome"
            :total-expense="sevenDaysExpense"
          />

          <!-- 3 Metric Cards Row -->
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <!-- Metric 1: Total Balance -->
            <div class="card-neo p-4 flex flex-col justify-between">
              <span class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">{{ t('dashboard.financialSummary.totalBalance') }}</span>
              <h3 class="text-xl font-black text-slate-900 tracking-tight my-1 font-mono">
                {{ formatCurrency(totalBalance) }}
              </h3>
              <span class="text-[10px] font-semibold text-slate-400">{{ t('dashboard.financialSummary.totalBalanceDesc') }}</span>
            </div>

            <!-- Metric 2: 7-Day Income -->
            <div class="card-neo p-4 flex flex-col justify-between">
              <span class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">{{ t('dashboard.financialSummary.sevenDaysIncome') }}</span>
              <h3 class="text-xl font-black text-emerald-600 tracking-tight my-1 font-mono">
                +{{ formatCurrency(sevenDaysIncome) }}
              </h3>
              <span class="text-[10px] font-semibold text-slate-400">{{ t('dashboard.financialSummary.sevenDaysIncomeDesc') }}</span>
            </div>

            <!-- Metric 3: 7-Day Expenses -->
            <div class="card-neo p-4 flex flex-col justify-between">
              <span class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">{{ t('dashboard.financialSummary.sevenDaysExpense') }}</span>
              <h3 class="text-xl font-black text-rose-600 tracking-tight my-1 font-mono">
                -{{ formatCurrency(sevenDaysExpense) }}
              </h3>
              <span class="text-[10px] font-semibold text-slate-400">{{ t('dashboard.financialSummary.sevenDaysExpenseDesc') }}</span>
            </div>
          </div>
        </section>

        </div><!-- /LEFT COLUMN -->

        <!-- RIGHT COLUMN -->
        <div class="flex flex-col gap-8">

        <!-- 3. WIDGET: Wallets -->
        <section class="flex flex-col gap-4">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-xl font-extrabold text-slate-900 tracking-tight">
                {{ t('dashboard.wallets.title') }}
              </h2>
              <p class="text-xs text-slate-400 font-medium">
                {{ t('dashboard.wallets.subtitle') }}
              </p>
            </div>
            <button 
              v-if="isAdmin" 
              @click="openWalletModal"
              class="px-3.5 py-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-800 font-bold text-xs transition flex items-center gap-1.5"
            >
              <span>{{ t('dashboard.wallets.addWalletBtn') }}</span>
            </button>
          </div>

          <!-- Wallets Loading / Empty / Grid -->
          <div v-if="walletStore.loading && walletStore.wallets.length === 0" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            <div v-for="i in 3" :key="i" class="card-neo p-5 animate-pulse flex flex-col justify-between h-36">
              <div class="h-4 bg-slate-200 rounded w-1/3 mb-2"></div>
              <div class="h-6 bg-slate-200 rounded w-1/2"></div>
              <div class="h-4 bg-slate-100 rounded w-full mt-4"></div>
            </div>
          </div>

          <div v-else-if="walletStore.wallets.length === 0" class="card-neo p-8 text-center">
            <p class="text-xs text-slate-400 font-medium">{{ t('dashboard.wallets.noWallets') }}</p>
            <button v-if="isAdmin" @click="openWalletModal" class="mt-2 text-xs font-bold text-brand-600 hover:underline">
              {{ t('dashboard.wallets.createFirst') }}
            </button>
          </div>

          <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            <div 
              v-for="w in walletStore.wallets" 
              :key="w.id"
              class="card-neo p-5 flex flex-col justify-between hover:border-brand-300 transition group"
            >
              <div>
                <div class="flex items-start justify-between">
                  <div>
                    <span class="text-[10px] font-extrabold uppercase px-2 py-0.5 rounded-full bg-slate-100 text-slate-600">{{ t('dashboard.wallets.envelope') }}</span>
                    <h4 class="font-black text-slate-900 text-base mt-1">{{ w.name }}</h4>
                  </div>
                  <div class="w-8 h-8 rounded-xl bg-brand-50 text-brand-700 font-bold flex items-center justify-center text-sm">
                    💳
                  </div>
                </div>
                <p class="text-xs text-slate-400 mt-1 line-clamp-1">{{ w.description || t('walletsTab.noDescription') }}</p>
              </div>

              <div class="mt-4 pt-3 border-t border-slate-100 flex items-end justify-between">
                <div>
                  <span class="text-[9px] uppercase font-bold text-slate-400 block">{{ t('dashboard.wallets.currentBalance') }}</span>
                  <span class="text-lg font-black text-slate-900 font-mono">{{ formatCurrency(w.current_balance) }}</span>
                </div>
                <div class="text-right">
                  <span class="text-[9px] uppercase font-bold text-slate-400 block">{{ t('dashboard.wallets.minLimit') }}</span>
                  <span 
                    class="text-xs font-bold font-mono"
                    :class="w.current_balance <= w.minimum_limit ? 'text-rose-600 font-black' : 'text-slate-600'"
                  >
                    {{ formatCurrency(w.minimum_limit) }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 4. WIDGET: Allocation -->
        <section class="flex flex-col gap-4">
          <div>
            <h2 class="text-xl font-extrabold text-slate-900 tracking-tight">
              {{ t('dashboard.expenseAnalysis.title') }}
            </h2>
            <p class="text-xs text-slate-400 font-medium">
              {{ t('dashboard.expenseAnalysis.subtitle') }}
            </p>
          </div>

          <div class="card-neo p-6">
            <div v-if="expenseAnalysis.categories.length === 0" class="py-6 text-center text-xs text-slate-400 font-medium">
              {{ t('dashboard.expenseAnalysis.noExpenses') }}
            </div>

            <div v-else class="flex flex-col gap-4">
              <div class="flex items-baseline justify-between">
                <span class="text-xs font-bold text-slate-500 uppercase">{{ t('dashboard.expenseAnalysis.totalExpenses') }}</span>
                <span class="text-xl font-black text-slate-900 font-mono">{{ formatCurrency(expenseAnalysis.total) }}</span>
              </div>

              <!-- Segmented Multi-Color Progress Bar -->
              <div class="w-full flex h-3 rounded-full overflow-hidden gap-0.5 bg-slate-100 p-0.5">
                <div 
                  v-for="cat in expenseAnalysis.categories" 
                  :key="cat.name"
                  :class="cat.color"
                  :style="{ width: cat.percentage + '%' }"
                  class="h-full rounded-full transition-all"
                  :title="`${cat.name}: ${cat.percentage}%`"
                ></div>
              </div>

              <!-- Category breakdown items -->
              <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2.5 mt-1">
                <div 
                  v-for="cat in expenseAnalysis.categories" 
                  :key="cat.name"
                  class="p-2.5 rounded-xl bg-slate-50 border border-slate-100 flex items-center justify-between text-xs"
                >
                  <div class="flex items-center gap-2 min-w-0">
                    <span class="w-2.5 h-2.5 rounded-full shrink-0" :class="cat.color"></span>
                    <span class="font-bold text-slate-800 truncate text-[11px]">{{ cat.name }}</span>
                  </div>
                  <div class="text-right shrink-0">
                    <span class="font-bold text-slate-900 font-mono block text-xs">{{ formatCurrency(cat.amount) }}</span>
                    <span class="text-[10px] text-slate-400 font-semibold">{{ cat.percentage }}%</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 5. WIDGET: Recent History -->
        <section class="flex flex-col gap-4">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-xl font-extrabold text-slate-900 tracking-tight">
                {{ t('dashboard.recentHistory.title') }}
              </h2>
              <p class="text-xs text-slate-400 font-medium">
                {{ t('dashboard.recentHistory.subtitle') }}
              </p>
            </div>
            <button 
              @click="activeTab = 'transaksi'" 
              class="text-xs font-bold text-brand-600 hover:text-brand-700"
            >
              {{ t('dashboard.recentHistory.viewAll') }}
            </button>
          </div>

          <div class="card-neo overflow-hidden p-0 border border-slate-200/80">
            <div v-if="recentTenTransactions.length === 0" class="py-8 text-center text-xs text-slate-400 font-medium">
              {{ t('dashboard.recentHistory.noTransactions') }}
            </div>

            <div v-else class="overflow-x-auto max-h-80">
              <table class="w-full text-left text-xs">
                <thead class="bg-slate-50 border-b border-slate-200/80 text-slate-400 font-bold uppercase tracking-wider text-[10px] sticky top-0">
                  <tr>
                    <th class="py-3 px-4">{{ t('dashboard.recentHistory.dateCol') }}</th>
                    <th class="py-3 px-4">{{ t('dashboard.recentHistory.walletCol') }}</th>
                    <th class="py-3 px-4">{{ t('dashboard.recentHistory.descriptionCol') }}</th>
                    <th class="py-3 px-4 text-right">{{ t('dashboard.recentHistory.amountCol') }}</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                  <tr v-for="tx in recentTenTransactions" :key="tx.id" class="hover:bg-slate-50/80 transition font-medium">
                    <td class="py-2.5 px-4 text-slate-400 font-mono text-[11px] whitespace-nowrap">
                      {{ new Date(tx.created_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' }) }}
                    </td>
                    <td class="py-2.5 px-4 font-bold text-slate-800 text-[11px] truncate max-w-[100px]">{{ getWalletName(tx.wallet_id) }}</td>
                    <td class="py-2.5 px-4 text-slate-700 text-[11px] truncate max-w-[120px]">{{ tx.description || '-' }}</td>
                    <td class="py-2.5 px-4 text-right font-mono font-bold text-xs" :class="tx.type === 'income' ? 'text-emerald-600' : 'text-slate-900'">
                      {{ tx.type === 'income' ? '+' : '-' }}{{ formatCurrency(tx.amount) }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>

        </div><!-- /RIGHT COLUMN -->

      </div>

      <!-- ========================================================================= -->
      <!-- TAB 2: WALLETS TAB                                                        -->
      <!-- ========================================================================= -->
      <div v-else-if="activeTab === 'wallets'" class="flex flex-col gap-6">
        <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
          <div>
            <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">{{ t('walletsTab.title') }}</h2>
            <p class="text-xs text-slate-500 mt-0.5">{{ t('walletsTab.subtitle') }}</p>
          </div>
          <button 
            v-if="isAdmin" 
            class="px-4 py-2.5 bg-slate-900 hover:bg-slate-800 text-white rounded-2xl text-xs font-bold transition shadow-sm flex items-center gap-2"
            @click="openWalletModal"
          >
            {{ t('walletsTab.addBtn') }}
          </button>
        </div>

        <div v-if="walletStore.loading && walletStore.wallets.length === 0" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          <div v-for="i in 3" :key="i" class="card-neo p-6 animate-pulse h-48"></div>
        </div>

        <div v-else-if="walletStore.wallets.length === 0" class="card-neo p-12 text-center">
          <p class="text-sm font-semibold text-slate-500 mb-2">{{ t('walletsTab.noWallets') }}</p>
          <button v-if="isAdmin" @click="openWalletModal" class="text-xs font-bold text-brand-600 hover:underline">
            {{ t('walletsTab.createFirst') }}
          </button>
        </div>

        <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          <div 
            v-for="w in walletStore.wallets" 
            :key="w.id"
            class="card-neo p-6 flex flex-col justify-between group hover:border-brand-300 transition"
          >
            <div>
              <div class="flex items-start justify-between">
                <div>
                  <div class="flex items-center gap-1.5">
                    <span class="text-[10px] font-extrabold uppercase px-2 py-0.5 rounded-full bg-slate-100 text-slate-600">{{ t('dashboard.wallets.envelope') }}</span>
                    <span v-if="w.short_id" class="text-[10px] font-mono font-bold px-2 py-0.5 rounded-full bg-brand-100 text-brand-900">ID: {{ w.short_id }}</span>
                  </div>
                  <h3 class="text-lg font-black text-slate-900 mt-1">{{ w.name }}</h3>
                </div>
                <div class="flex items-center gap-1.5">
                  <button 
                    v-if="isAdmin" 
                    @click="openEditWalletModal(w)"
                    class="p-1.5 rounded-lg text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition text-xs"
                    :title="t('walletsTab.edit')"
                  >
                    <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M12 20h9"></path>
                      <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path>
                    </svg>
                  </button>
                  <button 
                    v-if="isAdmin" 
                    @click="openDeleteWalletModal(w)"
                    class="p-1.5 rounded-lg text-slate-400 hover:text-rose-600 hover:bg-rose-50 transition text-xs"
                    :title="t('walletsTab.delete')"
                  >
                    <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <polyline points="3 6 5 6 21 6"></polyline>
                      <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                    </svg>
                  </button>
                </div>
              </div>
              <p class="text-xs text-slate-400 mt-2">{{ w.description || t('walletsTab.noDescription') }}</p>
            </div>

            <div class="mt-6 pt-4 border-t border-slate-100 flex items-end justify-between">
              <div>
                <span class="text-[10px] font-bold uppercase text-slate-400 block">{{ t('dashboard.wallets.currentBalance') }}</span>
                <span class="text-xl font-black text-slate-900 font-mono">{{ formatCurrency(w.current_balance) }}</span>
              </div>
              <div class="text-right">
                <span class="text-[10px] font-bold uppercase text-slate-400 block">{{ t('dashboard.wallets.minLimit') }}</span>
                <span 
                  class="text-xs font-bold font-mono"
                  :class="w.current_balance <= w.minimum_limit ? 'text-rose-600 font-black' : 'text-slate-600'"
                >
                  {{ formatCurrency(w.minimum_limit) }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ========================================================================= -->
      <!-- TAB 3: TRANSAKSI TAB (Unified History, Action Buttons & Pending Widget)    -->
      <!-- ========================================================================= -->
      <div v-else-if="activeTab === 'transaksi'" class="flex flex-col gap-8">
        <!-- Top Title & Action Button -->
        <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
          <div>
            <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">{{ t('transaksiTab.title') }}</h2>
            <p class="text-xs text-slate-500 mt-0.5">{{ t('transaksiTab.subtitle') }}</p>
          </div>
          <button 
            @click="openTxModal" 
            class="px-4 py-2.5 bg-slate-900 hover:bg-slate-800 text-white rounded-2xl text-xs font-bold transition shadow-sm flex items-center gap-2"
          >
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="12" y1="5" x2="12" y2="19"></line>
              <line x1="5" y1="12" x2="19" y2="12"></line>
            </svg>
            <span>{{ isAdmin ? t('transaksiTab.recordBtn') : t('transaksiTab.submitBtn') }}</span>
          </button>
        </div>

        <!-- Transaction History Table -->
        <div class="card-neo overflow-hidden p-0 border border-slate-200/80">
          <div v-if="txStore.loading && txStore.transactions.length === 0" class="p-8 text-center text-xs text-slate-400">
            <span class="loading loading-spinner loading-md"></span>
          </div>

          <div v-else-if="txStore.transactions.length === 0" class="py-16 text-center text-xs text-slate-400">
            {{ t('transaksiTab.noTransactions') }}
          </div>

          <div v-else class="overflow-x-auto">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-50 border-b border-slate-200/80 text-slate-400 font-bold uppercase tracking-wider text-[10px]">
                <tr>
                  <th class="py-4 px-6">{{ t('transaksiTab.timestampCol') }}</th>
                  <th class="py-4 px-6">{{ t('transaksiTab.walletCol') }}</th>
                  <th class="py-4 px-6">{{ t('transaksiTab.typeCol') }}</th>
                  <th class="py-4 px-6">{{ t('transaksiTab.descriptionCol') }}</th>
                  <th class="py-4 px-6 text-right">{{ t('transaksiTab.amountCol') }}</th>
                  <th class="py-4 px-6 text-center">{{ t('transaksiTab.actionCol') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-for="tx in txStore.transactions" :key="tx.id" class="hover:bg-slate-50/80 transition font-medium">
                  <td class="py-4 px-6 text-slate-400 font-mono text-[11px] whitespace-nowrap">{{ new Date(tx.created_at).toLocaleString('id-ID') }}</td>
                  <td class="py-4 px-6 font-bold text-slate-800">{{ getWalletName(tx.wallet_id) }}</td>
                  <td class="py-4 px-6">
                    <span 
                      class="px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-wide inline-block"
                      :class="tx.type === 'income' ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-rose-50 text-rose-700 border border-rose-200'"
                    >
                      {{ tx.type }}
                    </span>
                  </td>
                  <td class="py-4 px-6 text-slate-800 font-semibold">{{ tx.description || '-' }}</td>
                  <td class="py-4 px-6 text-right font-mono font-black text-sm" :class="tx.type === 'income' ? 'text-emerald-600' : 'text-slate-900'">
                    {{ tx.type === 'income' ? '+' : '-' }}{{ formatCurrency(tx.amount) }}
                  </td>
                  <td class="py-4 px-6 text-center">
                    <div class="inline-flex items-center gap-1.5">
                      <!-- Pencil Icon: Admin = Direct Edit; Member = Change Request -->
                      <button 
                        @click="isAdmin ? openEditTxModal(tx) : openChangeRequestModal(tx)" 
                        class="p-1.5 rounded-lg text-slate-400 hover:text-brand-600 hover:bg-brand-50 transition"
                        :title="isAdmin ? t('transaksiTab.edit') : t('modals.changeRequest.title')"
                      >
                        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <path d="M12 20h9"></path>
                          <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path>
                        </svg>
                      </button>
                      <!-- Trash Icon: Admin = Direct Delete; Member = Delete Request -->
                      <button 
                        @click="isAdmin ? openDeleteTxModal(tx) : openDeleteRequestModal(tx)" 
                        class="p-1.5 rounded-lg text-slate-400 hover:text-rose-600 hover:bg-rose-50 transition"
                        :title="isAdmin ? t('transaksiTab.delete') : t('modals.deleteRequest.title')"
                      >
                        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <polyline points="3 6 5 6 21 6"></polyline>
                          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                        </svg>
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Admin-Only Pending Requests Widget (Appears after history table) -->
        <div v-if="isAdmin" class="flex flex-col gap-4 mt-2">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="font-extrabold text-lg text-slate-900 tracking-tight flex items-center gap-2">
                <span>{{ t('pendingRequests.title') }}</span>
                <span v-if="pendingProposals.length > 0" class="px-2 py-0.5 rounded-full bg-amber-100 text-amber-900 text-xs font-bold">
                  {{ pendingProposals.length }}
                </span>
              </h3>
              <p class="text-xs text-slate-500 mt-0.5">{{ t('pendingRequests.subtitle') }}</p>
            </div>
          </div>

          <div v-if="pendingProposals.length === 0" class="card-neo p-8 text-center text-xs text-slate-400">
            {{ t('pendingRequests.noRequests') }}
          </div>

          <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div 
              v-for="p in pendingProposals" 
              :key="p.id"
              class="card-neo p-5 flex flex-col justify-between border-amber-200/80 bg-amber-50/20"
            >
              <div>
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <span 
                      class="px-2 py-0.5 rounded-full text-[9px] font-black uppercase tracking-wider mb-1.5 inline-block"
                      :class="{
                        'bg-sky-100 text-sky-800': p.request_type === 'add_transaction' || !p.request_type,
                        'bg-amber-100 text-amber-800': p.request_type === 'edit_transaction',
                        'bg-rose-100 text-rose-800': p.request_type === 'delete_transaction',
                      }"
                    >
                      {{ p.request_type === 'edit_transaction' ? t('pendingRequests.typeEdit') : (p.request_type === 'delete_transaction' ? t('pendingRequests.typeDelete') : t('pendingRequests.typeAdd')) }}
                    </span>
                    <h4 class="font-black text-sm text-slate-900">{{ p.title }}</h4>
                  </div>
                  <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-100 text-amber-800 border border-amber-200">
                    {{ p.status }}
                  </span>
                </div>
                <p class="text-xs text-slate-600 mt-2 leading-relaxed">{{ p.description }}</p>
              </div>

              <div class="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between">
                <div>
                  <span class="text-[10px] text-slate-400 font-bold uppercase block">{{ t('pendingRequests.targetWallet') }}: {{ getWalletName(p.wallet_id) }}</span>
                  <span class="text-base font-black text-slate-900 font-mono">{{ formatCurrency(p.amount) }}</span>
                </div>

                <div class="flex gap-2">
                  <button 
                    @click="approveProp(p.id)" 
                    class="px-3 py-1.5 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs transition shadow-sm"
                  >
                    {{ t('pendingRequests.approveBtn') }}
                  </button>
                  <button 
                    @click="rejectProp(p.id)" 
                    class="px-3 py-1.5 rounded-xl bg-rose-600 hover:bg-rose-700 text-white font-bold text-xs transition shadow-sm"
                  >
                    {{ t('pendingRequests.rejectBtn') }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

    </main>

    <!-- ========================================================================= -->
    <!-- MODALS & POPUPS                                                           -->
    <!-- ========================================================================= -->

    <!-- 1. Modal Create Wallet -->
    <dialog :class="isWalletModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('modals.createWallet.title') }}</h3>
        <p class="text-xs text-slate-400 mb-4">{{ t('modals.createWallet.subtitle') }}</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.createWallet.nameLabel') }}</label>
            <input type="text" v-model="newWallet.name" :placeholder="t('modals.createWallet.namePlaceholder')" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.createWallet.descLabel') }}</label>
            <input type="text" v-model="newWallet.description" :placeholder="t('modals.createWallet.descPlaceholder')" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.createWallet.initialBalanceLabel') }}</label>
              <input type="number" v-model.number="newWallet.initial_balance" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
            </div>
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.createWallet.minLimitLabel') }}</label>
              <input type="number" v-model.number="newWallet.minimum_limit" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
            </div>
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeWalletModal" :disabled="isWalletSubmitting">{{ t('modals.createWallet.cancel') }}</button>
          <button class="btn bg-slate-900 hover:bg-slate-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitWallet" :disabled="isWalletSubmitting || !newWallet.name">
            {{ isWalletSubmitting ? t('modals.createWallet.saving') : t('modals.createWallet.save') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeWalletModal">close</button>
      </form>
    </dialog>

    <!-- 2. Modal Edit Wallet -->
    <dialog :class="isEditWalletModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('modals.editWallet.title') }}</h3>
        <p class="text-xs text-slate-400 mb-4">{{ t('modals.editWallet.subtitle') }}</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.editWallet.nameLabel') }}</label>
            <input type="text" v-model="editWalletData.name" :placeholder="t('modals.editWallet.namePlaceholder')" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.editWallet.descLabel') }}</label>
            <input type="text" v-model="editWalletData.description" :placeholder="t('modals.editWallet.descPlaceholder')" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.editWallet.minLimitLabel') }}</label>
            <input type="number" v-model.number="editWalletData.minimum_limit" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeEditWalletModal" :disabled="isEditWalletSubmitting">{{ t('modals.editWallet.cancel') }}</button>
          <button class="btn bg-slate-900 hover:bg-slate-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitEditWallet" :disabled="isEditWalletSubmitting || !editWalletData.name.trim()">
            {{ isEditWalletSubmitting ? t('modals.editWallet.saving') : t('modals.editWallet.save') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeEditWalletModal">close</button>
      </form>
    </dialog>

    <!-- 3. Modal Confirm Delete Wallet -->
    <dialog :class="isDeleteWalletModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-rose-100 max-w-md">
        <div class="w-12 h-12 rounded-2xl bg-rose-100 text-rose-600 flex items-center justify-center mb-3">
          <svg class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"></polyline>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
          </svg>
        </div>
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('modals.confirmDelete.titleWallet') }}</h3>
        <p class="text-xs text-slate-500 mb-2">{{ t('modals.confirmDelete.descWallet') }}</p>
        <p class="text-xs font-bold text-slate-800 bg-slate-50 p-2.5 rounded-xl border border-slate-200 font-mono">{{ deleteWalletTarget?.name }}</p>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeDeleteWalletModal" :disabled="isDeleteWalletSubmitting">{{ t('modals.confirmDelete.cancel') }}</button>
          <button class="btn bg-rose-600 hover:bg-rose-700 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleConfirmDeleteWallet" :disabled="isDeleteWalletSubmitting">
            {{ isDeleteWalletSubmitting ? t('modals.confirmDelete.deleting') : t('modals.confirmDelete.confirm') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeDeleteWalletModal">close</button>
      </form>
    </dialog>

    <!-- 4. Modal Family Management -->
    <dialog :class="isFamilyManageModalOpen ? 'modal modal-open' : 'modal'" :open="isFamilyManageModalOpen">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-xl">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('modals.familyManage.title') }}</h3>
        <p class="text-xs text-slate-400 mb-4">{{ t('modals.familyManage.subtitle') }}</p>

        <!-- Error State Banner -->
        <div v-if="familyStore.error" class="mb-4 p-3.5 bg-rose-50 border border-rose-100 rounded-2xl text-rose-800 text-xs font-semibold flex items-center gap-2">
          <span>❌</span>
          <span class="flex-1">{{ familyStore.error }}</span>
        </div>

        <!-- Permission Denied Banner if non-admin -->
        <div v-if="!isAdmin" class="mb-4 p-3.5 bg-amber-50 border border-amber-100 rounded-2xl text-amber-800 text-xs font-semibold flex items-center gap-2">
          <span>⚠️</span>
          <span class="flex-1">{{ t('modals.familyManage.permissionDenied') }}</span>
        </div>

        <!-- Family Room Name Editing (Admin only) -->
        <div v-if="isAdmin" class="mb-6 pb-6 border-b border-slate-100">
          <label class="text-xs font-bold text-slate-700 block mb-1.5">{{ t('modals.familyManage.nameLabel') }}</label>
          <div class="flex gap-2">
            <input 
              type="text" 
              v-model="editFamilyName" 
              :placeholder="t('modals.familyManage.namePlaceholder')" 
              class="flex-1 px-3.5 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" 
            />
            <button 
              class="px-4 py-2 bg-slate-900 hover:bg-slate-800 text-white rounded-xl text-xs font-bold transition disabled:opacity-50"
              @click="handleSubmitEditFamilyName"
              :disabled="isFamilyManageSubmitting || !editFamilyName.trim() || editFamilyName.trim() === familyStore.family?.name"
            >
              {{ isFamilyManageSubmitting ? t('modals.familyManage.savingName') : t('modals.familyManage.saveName') }}
            </button>
          </div>
        </div>

        <!-- Members list block -->
        <div>
          <h4 class="text-xs font-black uppercase tracking-wider text-slate-500 mb-3">
            {{ t('modals.familyManage.membersTitle') }} ({{ realFamilyMembers.length }})
          </h4>

          <!-- Loading State inside modal -->
          <div v-if="familyStore.loading && realFamilyMembers.length === 0" class="flex flex-col gap-2">
            <div v-for="i in 3" :key="i" class="h-14 bg-slate-50 animate-pulse rounded-2xl"></div>
          </div>

          <!-- Members Grid/List -->
          <div v-else class="flex flex-col gap-2.5 max-h-64 overflow-y-auto pr-1">
            <div 
              v-for="member in realFamilyMembers" 
              :key="member.id"
              class="p-3 rounded-2xl bg-slate-50 border border-slate-100 flex items-center justify-between gap-3 hover:bg-slate-100/70 transition"
            >
              <div class="flex items-center gap-3 min-w-0">
                <div class="w-9 h-9 rounded-xl bg-gradient-to-tr from-slate-900 to-slate-700 text-white font-black text-xs flex items-center justify-center shrink-0 shadow-sm">
                  {{ (member.user_name || (member.user_id === authStore.user?.id ? authStore.user?.username : 'U') || 'U').charAt(0).toUpperCase() }}
                </div>
                <div class="min-w-0">
                  <div class="flex items-center gap-1.5">
                    <h5 class="text-xs font-black text-slate-900 truncate">
                      {{ member.user_name || (member.user_id === authStore.user?.id ? (authStore.user?.username || authStore.user?.name) : (`${t('dashboard.familyMembers.member')} #${member.user_id.slice(0, 6)}`)) }}
                    </h5>
                    <!-- You Indicator -->
                    <span 
                      v-if="member.user_id === authStore.user?.id" 
                      class="text-[9px] font-extrabold uppercase px-1.5 py-0.5 rounded bg-brand-400 text-slate-950 tracking-wider shrink-0"
                    >
                      {{ t('modals.familyManage.youBadge') }}
                    </span>
                  </div>
                  <span class="text-[9px] text-slate-400 font-medium">
                    {{ t('dashboard.familyMembers.joined') }} {{ new Date(member.joined_at).toLocaleDateString('id-ID') }}
                  </span>
                </div>
              </div>

              <div class="flex items-center gap-2">
                <!-- Role Badge -->
                <span 
                  class="text-[9px] font-extrabold uppercase px-2 py-0.5 rounded-full shrink-0 border"
                  :class="member.role === 'admin' ? 'bg-amber-100 text-amber-900 border-amber-200' : 'bg-white text-slate-700 border-slate-200'"
                >
                  {{ member.role === 'admin' ? t('dashboard.familyMembers.adminRole') : t('dashboard.familyMembers.memberRole') }}
                </span>

                <!-- Member delete button (admin only, hidden for self) -->
                <button 
                  v-if="isAdmin && member.user_id !== authStore.user?.id"
                  @click="openDeleteMemberModal(member)"
                  class="p-1.5 rounded-lg text-slate-400 hover:text-rose-600 hover:bg-rose-50 transition"
                  :title="t('modals.familyManage.removeMember')"
                >
                  <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="3 6 5 6 21 6"></polyline>
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-action mt-6">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeFamilyManageModal">{{ t('modals.familyManage.cancel') }}</button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeFamilyManageModal">close</button>
      </form>
    </dialog>

    <!-- Modal Confirm Delete Member -->
    <dialog :class="isDeleteMemberModalOpen ? 'modal modal-open' : 'modal'" :open="isDeleteMemberModalOpen">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-rose-100 max-w-md">
        <div class="w-12 h-12 rounded-2xl bg-rose-100 text-rose-600 flex items-center justify-center mb-3">
          <svg class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
            <circle cx="8.5" cy="7" r="4"></circle>
            <line x1="18" y1="8" x2="23" y2="13"></line>
            <line x1="23" y1="8" x2="18" y2="13"></line>
          </svg>
        </div>
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('modals.familyManage.confirmRemoveTitle') }}</h3>
        <p class="text-xs text-slate-500 mb-4">
          {{ t('modals.familyManage.confirmRemoveDesc').replace('{name}', deleteMemberTarget?.user_name || '') }}
        </p>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeDeleteMemberModal" :disabled="isDeleteMemberSubmitting">{{ t('modals.confirmDelete.cancel') }}</button>
          <button class="btn bg-rose-600 hover:bg-rose-700 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleConfirmDeleteMember" :disabled="isDeleteMemberSubmitting">
            {{ isDeleteMemberSubmitting ? t('modals.familyManage.removingBtn') : t('modals.familyManage.confirmRemoveBtn') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeDeleteMemberModal">close</button>
      </form>
    </dialog>

    <!-- 5. Modal Create Transaction (Admin: Direct Record | Member: Submit Proposal) -->
    <dialog :class="isTxModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ isAdmin ? t('modals.createTx.titleAdmin') : t('modals.createTx.titleMember') }}</h3>
        <p class="text-xs text-slate-400 mb-4">{{ isAdmin ? t('modals.createTx.subtitleAdmin') : t('modals.createTx.subtitleMember') }}</p>

        <div v-if="walletStore.wallets.length === 0" class="p-4 bg-amber-50 border border-amber-200 rounded-2xl text-amber-800 text-xs mb-4 font-semibold">
          {{ t('transaksiTab.noWalletsWarning') }}
        </div>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.createTx.selectWallet') }}</label>
            <select v-model="newTx.wallet_id" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" :disabled="walletStore.wallets.length === 0">
              <option v-for="w in walletStore.wallets" :key="w.id" :value="w.id">{{ w.name }} ({{ t('dashboard.wallets.currentBalance') }}: {{ formatCurrency(w.current_balance) }})</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.createTx.txType') }}</label>
            <select v-model="newTx.type" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400">
              <option value="expense">{{ t('modals.createTx.expenseOption') }}</option>
              <option value="income">{{ t('modals.createTx.incomeOption') }}</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.createTx.amountLabel') }}</label>
            <input type="number" v-model.number="newTx.amount" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.createTx.descLabel') }}</label>
            <input type="text" v-model="newTx.description" :placeholder="t('modals.createTx.descPlaceholder')" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeTxModal" :disabled="isTxSubmitting">{{ t('modals.createTx.cancel') }}</button>
          <button class="btn bg-slate-900 hover:bg-slate-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitTx" :disabled="isTxSubmitting || !newTx.wallet_id || newTx.amount <= 0">
            {{ isTxSubmitting ? t('modals.createTx.saving') : (isAdmin ? t('modals.createTx.saveAdmin') : t('modals.createTx.saveMember')) }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeTxModal">close</button>
      </form>
    </dialog>

    <!-- 6. Modal Edit Transaction (Admin Direct Edit) -->
    <dialog :class="isEditTxModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('modals.editTx.title') }}</h3>
        <p class="text-xs text-slate-400 mb-4">{{ t('modals.editTx.subtitle') }}</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.editTx.selectWallet') }}</label>
            <select v-model="editTxForm.wallet_id" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400">
              <option v-for="w in walletStore.wallets" :key="w.id" :value="w.id">{{ w.name }} ({{ formatCurrency(w.current_balance) }})</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.editTx.txType') }}</label>
            <select v-model="editTxForm.type" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400">
              <option value="expense">{{ t('modals.editTx.expenseOption') }}</option>
              <option value="income">{{ t('modals.editTx.incomeOption') }}</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.editTx.amountLabel') }}</label>
            <input type="number" v-model.number="editTxForm.amount" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.editTx.descLabel') }}</label>
            <input type="text" v-model="editTxForm.description" :placeholder="t('modals.editTx.descPlaceholder')" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeEditTxModal" :disabled="isEditTxSubmitting">{{ t('modals.editTx.cancel') }}</button>
          <button class="btn bg-slate-900 hover:bg-slate-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitEditTx" :disabled="isEditTxSubmitting || editTxForm.amount <= 0">
            {{ isEditTxSubmitting ? t('modals.editTx.saving') : t('modals.editTx.save') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeEditTxModal">close</button>
      </form>
    </dialog>

    <!-- 7. Modal Change Request (Member) -->
    <dialog :class="isChangeRequestModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('modals.changeRequest.title') }}</h3>
        <p class="text-xs text-slate-400 mb-4">{{ t('modals.changeRequest.subtitle') }}</p>

        <div v-if="selectedTxForChangeRequest" class="p-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs mb-4">
          <div class="flex justify-between text-slate-500 font-bold mb-1">
            <span>{{ t('modals.changeRequest.origAmount') }}:</span>
            <span class="font-mono text-slate-900">{{ formatCurrency(selectedTxForChangeRequest.amount) }}</span>
          </div>
          <div class="flex justify-between text-slate-500 font-bold">
            <span>{{ t('modals.changeRequest.origDesc') }}:</span>
            <span class="text-slate-800">{{ selectedTxForChangeRequest.description || '-' }}</span>
          </div>
        </div>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.changeRequest.newTypeLabel') }}</label>
            <select v-model="changeRequestForm.type" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400">
              <option value="expense">{{ t('modals.editTx.expenseOption') }}</option>
              <option value="income">{{ t('modals.editTx.incomeOption') }}</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.changeRequest.newAmountLabel') }}</label>
            <input type="number" v-model.number="changeRequestForm.amount" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.changeRequest.newDescLabel') }}</label>
            <textarea v-model="changeRequestForm.description" rows="2" :placeholder="t('modals.changeRequest.newDescPlaceholder')" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400"></textarea>
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeChangeRequestModal" :disabled="isChangeRequestSubmitting">{{ t('modals.changeRequest.cancel') }}</button>
          <button class="btn bg-slate-900 hover:bg-slate-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitChangeRequest" :disabled="isChangeRequestSubmitting || changeRequestForm.amount <= 0">
            {{ isChangeRequestSubmitting ? t('modals.changeRequest.saving') : t('modals.changeRequest.save') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeChangeRequestModal">close</button>
      </form>
    </dialog>

    <!-- 8. Modal Delete Request (Member) -->
    <dialog :class="isDeleteRequestModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-rose-100 max-w-md">
        <div class="w-12 h-12 rounded-2xl bg-rose-100 text-rose-600 flex items-center justify-center mb-3">
          <svg class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"></polyline>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
          </svg>
        </div>
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('modals.deleteRequest.title') }}</h3>
        <p class="text-xs text-slate-500 mb-4">{{ t('modals.deleteRequest.confirmMsg').replace('{amount}', formatCurrency(selectedTxForDeleteRequest?.amount || 0)) }}</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">{{ t('modals.deleteRequest.reasonLabel') }}</label>
            <input type="text" v-model="deleteRequestReason" :placeholder="t('modals.deleteRequest.reasonPlaceholder')" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeDeleteRequestModal" :disabled="isDeleteRequestSubmitting">{{ t('modals.deleteRequest.cancel') }}</button>
          <button class="btn bg-rose-600 hover:bg-rose-700 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitDeleteRequest" :disabled="isDeleteRequestSubmitting">
            {{ isDeleteRequestSubmitting ? t('modals.deleteRequest.saving') : t('modals.deleteRequest.save') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeDeleteRequestModal">close</button>
      </form>
    </dialog>

    <!-- 9. Modal Confirm Delete Transaction (Admin Direct Delete) -->
    <dialog :class="isDeleteTxModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-rose-100 max-w-md">
        <div class="w-12 h-12 rounded-2xl bg-rose-100 text-rose-600 flex items-center justify-center mb-3">
          <svg class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"></polyline>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
          </svg>
        </div>
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('modals.confirmDelete.titleTx') }}</h3>
        <p class="text-xs text-slate-500 mb-2">{{ t('modals.confirmDelete.descTx') }}</p>
        <div v-if="deleteTxTarget" class="p-3 bg-slate-50 rounded-xl border border-slate-200 text-xs font-semibold">
          <div class="flex justify-between items-center">
            <span>{{ deleteTxTarget.description || 'Transaction' }}</span>
            <span class="font-mono font-bold">{{ formatCurrency(deleteTxTarget.amount) }}</span>
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeDeleteTxModal" :disabled="isDeleteTxSubmitting">{{ t('modals.confirmDelete.cancel') }}</button>
          <button class="btn bg-rose-600 hover:bg-rose-700 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleConfirmDeleteTx" :disabled="isDeleteTxSubmitting">
            {{ isDeleteTxSubmitting ? t('modals.confirmDelete.deleting') : t('modals.confirmDelete.confirm') }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="closeDeleteTxModal">close</button>
      </form>
    </dialog>

    <!-- 8. Modal Telegram Bot Info -->
    <dialog :class="isTelegramModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">{{ t('modals.telegram.title') }}</h3>
        <p class="text-xs text-slate-400 mb-4">{{ t('modals.telegram.subtitle') }}</p>

        <div v-if="familyStore.family?.telegram_chat_id" class="p-4 rounded-2xl bg-emerald-50 border border-emerald-200 text-xs">
          <div class="flex items-center gap-2 text-emerald-800 font-bold mb-1">
            <span class="w-2.5 h-2.5 rounded-full bg-emerald-500 animate-pulse"></span>
            <span>{{ t('modals.telegram.connected') }}</span>
          </div>
          <p class="text-emerald-700">{{ t('modals.telegram.chatId') }}: <span class="font-mono font-bold">{{ familyStore.family.telegram_chat_id }}</span></p>
        </div>

        <div v-else class="flex flex-col gap-3 text-xs text-slate-600 bg-slate-50 p-4 rounded-2xl border border-slate-200/80">
          <p class="font-bold text-slate-800">{{ t('modals.telegram.howToLink') }}</p>
          <ol class="list-decimal list-inside space-y-1.5">
            <li>{{ t('modals.telegram.step1') }}</li>
            <li>{{ t('modals.telegram.step2') }}</li>
          </ol>
          <div class="bg-white p-2.5 rounded-xl border border-slate-200 flex items-center justify-between font-mono font-bold text-slate-900">
            <code>/link {{ familyStore.family?.invite_code }}</code>
          </div>
        </div>

        <div class="modal-action mt-6 justify-between">
          <button 
            v-if="familyStore.family?.telegram_chat_id" 
            class="btn btn-error btn-outline btn-sm rounded-xl text-xs font-bold"
            @click="handleDisconnectTelegram"
          >
            {{ t('modals.telegram.disconnect') }}
          </button>
          <div v-else></div>
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="isTelegramModalOpen = false">{{ t('modals.telegram.close') }}</button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button type="button" @click="isTelegramModalOpen = false">close</button>
      </form>
    </dialog>

    <!-- Floating Toast Notifications -->
    <div class="fixed bottom-6 right-6 z-50 flex flex-col gap-2 pointer-events-none max-w-sm w-full">
      <transition-group 
        enter-active-class="transition duration-300 ease-out transform"
        enter-from-class="translate-y-4 opacity-0 scale-95"
        enter-to-class="translate-y-0 opacity-100 scale-100"
        leave-active-class="transition duration-200 ease-in transform"
        leave-from-class="translate-y-0 opacity-100 scale-100"
        leave-to-class="translate-y-4 opacity-0 scale-95"
      >
        <div 
          v-for="t in toasts" 
          :key="t.id"
          class="pointer-events-auto flex items-center gap-3 p-4 rounded-2xl shadow-xl border text-xs font-bold backdrop-blur-md transition-all"
          :class="{
            'bg-slate-900 text-white border-slate-700': t.type === 'info',
            'bg-emerald-950 text-emerald-200 border-emerald-800': t.type === 'success',
            'bg-rose-950 text-rose-200 border-rose-800': t.type === 'error'
          }"
        >
          <span v-if="t.type === 'success'" class="text-base">✅</span>
          <span v-else-if="t.type === 'error'" class="text-base">❌</span>
          <span v-else class="text-base">ℹ️</span>
          <span class="flex-1">{{ t.message }}</span>
        </div>
      </transition-group>
    </div>
  </div>
</template>

