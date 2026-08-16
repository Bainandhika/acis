<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useWalletStore } from '../stores/wallet'
import { useAuthStore } from '../stores/auth'
import { useFamilyStore } from '../stores/family'
import { useTransactionStore } from '../stores/transaction'
import { useRouter } from 'vue-router'
import type { CreateWalletPayload } from '../services/wallet'
import type { CreateTransactionPayload, CreateProposalPayload } from '../services/transaction'

// Component imports
import TopNavbar, { type NavTab } from '../components/TopNavbar.vue'
import CashflowChart from '../components/CashflowChart.vue'

const walletStore = useWalletStore()
const authStore = useAuthStore()
const familyStore = useFamilyStore()
const txStore = useTransactionStore()
const router = useRouter()

// UI State: Horizontal Navigation Tabs
const activeTab = ref<NavTab>('dashboard')
const selectedWalletId = ref('')

// Modal Controls
const isWalletModalOpen = ref(false)
const isTxModalOpen = ref(false)
const isProposalModalOpen = ref(false)
const isTelegramModalOpen = ref(false)
const isSubmitting = ref(false)

// Role Check
const isAdmin = computed(() => authStore.user?.role === 'admin')

// Form States (Strict 0 initial balances)
const newWallet = ref<CreateWalletPayload>({
  name: '',
  description: '',
  initial_balance: 0,
  minimum_limit: 0,
})

const newTx = ref<CreateTransactionPayload>({
  wallet_id: '',
  type: 'expense',
  amount: 0,
  category: 'General',
  description: '',
})

const newProposal = ref<CreateProposalPayload>({
  wallet_id: '',
  title: '',
  amount: 0,
  description: '',
})

// Metrics & Analytics Calculations (Strictly from real data, initialized to 0)
const totalBalance = computed(() => {
  return (walletStore.wallets || []).reduce((sum, w) => sum + (w?.current_balance || 0), 0)
})

const totalAllocation = computed(() => {
  return (walletStore.wallets || []).reduce((sum, w) => sum + (w?.initial_balance || 0), 0)
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

const sevenDaysSavings = computed(() => {
  return Math.max(0, sevenDaysIncome.value - sevenDaysExpense.value)
})

// Section 3: Expense Analysis by category (last 7 days strictly from real data)
const categoryPalette = [
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

  const categoryMap: Record<string, number> = {}
  expenseTxs.forEach(t => {
    const cat = t.category?.trim() || 'General'
    categoryMap[cat] = (categoryMap[cat] || 0) + (t.amount || 0)
  })

  const categories = Object.entries(categoryMap).map(([name, amount], index) => {
    const percentage = Math.round((amount / total) * 100)
    return {
      name,
      amount,
      percentage,
      color: categoryPalette[index % categoryPalette.length]
    }
  }).sort((a, b) => b.amount - a.amount)

  return { total, categories }
})

// Section 4: Real Family Members from database (NO dummy members)
const realFamilyMembers = computed(() => {
  return familyStore.family?.members || []
})

// Section 5: Recent History (Last 10 dated transactions)
const recentTenTransactions = computed(() => {
  return (txStore.transactions || []).slice(0, 10)
})

// Currency Formatter
const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount || 0)
}

const getWalletName = (id?: string) => {
  if (!id) return 'Main Wallet'
  const w = (walletStore.wallets || []).find(item => item?.id === id)
  return w ? w.name : 'Main Wallet'
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

// Telegram action
const handleDisconnectTelegram = async () => {
  if (confirm('Disconnect Telegram Bot for your family?')) {
    try {
      await familyStore.handleDisconnectTelegram()
      isTelegramModalOpen.value = false
      showToast('Telegram bot disconnected successfully', 'info')
    } catch {
      showToast('Failed to disconnect Telegram bot', 'error')
    }
  }
}

// Modal actions
const openWalletModal = () => { isWalletModalOpen.value = true }
const closeWalletModal = () => {
  isWalletModalOpen.value = false
  newWallet.value = { name: '', description: '', initial_balance: 0, minimum_limit: 0 }
}
const handleSubmitWallet = async () => {
  isSubmitting.value = true
  try {
    await walletStore.addWallet(newWallet.value)
    closeWalletModal()
    showToast('New wallet envelope created!', 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to create wallet', 'error')
  } finally {
    isSubmitting.value = false
  }
}

const openTxModal = () => {
  if (walletStore.wallets.length > 0 && walletStore.wallets[0]) {
    newTx.value.wallet_id = selectedWalletId.value || walletStore.wallets[0].id
  }
  isTxModalOpen.value = true
}
const closeTxModal = () => {
  isTxModalOpen.value = false
  newTx.value = { wallet_id: '', type: 'expense', amount: 0, category: 'General', description: '' }
}
const handleSubmitTx = async () => {
  isSubmitting.value = true
  try {
    await txStore.addTransaction(newTx.value)
    await walletStore.fetchWallets()
    closeTxModal()
    showToast('Transaction recorded successfully!', 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to record transaction', 'error')
  } finally {
    isSubmitting.value = false
  }
}

const openProposalModal = () => {
  if (walletStore.wallets.length > 0 && walletStore.wallets[0]) {
    newProposal.value.wallet_id = selectedWalletId.value || walletStore.wallets[0].id
  }
  isProposalModalOpen.value = true
}
const closeProposalModal = () => {
  isProposalModalOpen.value = false
  newProposal.value = { wallet_id: '', title: '', amount: 0, description: '' }
}
const handleSubmitProposal = async () => {
  isSubmitting.value = true
  try {
    await txStore.addProposal(newProposal.value)
    closeProposalModal()
    showToast('Proposal submitted for Admin review!', 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to submit proposal', 'error')
  } finally {
    isSubmitting.value = false
  }
}

const approveProp = async (id: string) => {
  try {
    await txStore.handleApprove(id)
    await walletStore.fetchWallets()
    showToast('Proposal approved and wallet deducted!', 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to approve proposal', 'error')
  }
}

const rejectProp = async (id: string) => {
  try {
    await txStore.handleReject(id)
    showToast('Proposal rejected', 'info')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Failed to reject proposal', 'error')
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
    />

    <!-- Main Container -->
    <main class="flex-1 max-w-[1600px] w-full mx-auto p-4 sm:p-6 lg:p-8">

      <!-- ========================================================================= -->
      <!-- TAB 1: REBUILT DASHBOARD (STRICT 5 SECTIONS ONLY, NO DUMMY DATA)         -->
      <!-- ========================================================================= -->
      <div v-if="activeTab === 'dashboard'" class="flex flex-col gap-8">
        
        <!-- SECTION 1: Overall Financial Summary (last 7 days) -->
        <section class="flex flex-col gap-4">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-xl font-extrabold text-slate-900 tracking-tight">1. Overall Financial Summary</h2>
              <p class="text-xs text-slate-400 font-medium">Real cashflow performance for the last 7 days</p>
            </div>
            <button 
              @click="openTxModal" 
              class="px-4 py-2 rounded-xl bg-slate-900 hover:bg-slate-800 text-white font-bold text-xs transition shadow-sm flex items-center gap-1.5"
            >
              <span>+ Record Transaction</span>
            </button>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch">
            <!-- 7-Day Interactive Cashflow Chart (8 Cols) -->
            <div class="lg:col-span-8">
              <CashflowChart 
                :transactions="txStore.transactions"
                :total-balance="totalBalance"
                :total-income="sevenDaysIncome"
                :total-expense="sevenDaysExpense"
              />
            </div>

            <!-- 7-Day Metric Cards Stack (4 Cols) -->
            <div class="lg:col-span-4 flex flex-col gap-3 justify-between">
              <!-- Metric 1: Total Balance -->
              <div class="card-neo p-4 flex flex-col justify-between">
                <span class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Total Balance</span>
                <h3 class="text-2xl font-black text-slate-900 tracking-tight my-1">
                  {{ formatCurrency(totalBalance) }}
                </h3>
                <span class="text-[10px] font-semibold text-slate-400">Sum of all active wallet envelopes</span>
              </div>

              <!-- Metric 2: 7-Day Income -->
              <div class="card-neo p-4 flex flex-col justify-between">
                <span class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">7-Day Income</span>
                <h3 class="text-2xl font-black text-emerald-600 tracking-tight my-1">
                  +{{ formatCurrency(sevenDaysIncome) }}
                </h3>
                <span class="text-[10px] font-semibold text-slate-400">Total received in last 7 days</span>
              </div>

              <!-- Metric 3: 7-Day Expenses -->
              <div class="card-neo p-4 flex flex-col justify-between">
                <span class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">7-Day Expenses</span>
                <h3 class="text-2xl font-black text-rose-600 tracking-tight my-1">
                  -{{ formatCurrency(sevenDaysExpense) }}
                </h3>
                <span class="text-[10px] font-semibold text-slate-400">Total spent in last 7 days</span>
              </div>
            </div>
          </div>
        </section>

        <!-- SECTION 2: Wallets (Real Wallets e.g. Food, Shopping, Fuel) -->
        <section class="flex flex-col gap-4">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-xl font-extrabold text-slate-900 tracking-tight">2. Wallets</h2>
              <p class="text-xs text-slate-400 font-medium">Real active envelopes (e.g. Food &amp; Groceries, Shopping, Fuel)</p>
            </div>
            <button 
              v-if="isAdmin" 
              @click="openWalletModal"
              class="px-3.5 py-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-800 font-bold text-xs transition flex items-center gap-1.5"
            >
              <span>+ Add Wallet</span>
            </button>
          </div>

          <!-- Wallets Grid -->
          <div v-if="walletStore.wallets.length === 0" class="card-neo p-8 text-center">
            <p class="text-xs text-slate-400 font-medium">No wallets created yet.</p>
            <button v-if="isAdmin" @click="openWalletModal" class="mt-2 text-xs font-bold text-brand-600 hover:underline">
              + Create First Wallet Envelope
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
                    <span class="text-[10px] font-extrabold uppercase px-2 py-0.5 rounded-full bg-slate-100 text-slate-600">Envelope</span>
                    <h4 class="font-black text-slate-900 text-base mt-1">{{ w.name }}</h4>
                  </div>
                  <div class="w-8 h-8 rounded-xl bg-brand-50 text-brand-700 font-bold flex items-center justify-center text-sm">
                    💳
                  </div>
                </div>
                <p class="text-xs text-slate-400 mt-1 line-clamp-1">{{ w.description || 'Virtual envelope' }}</p>
              </div>

              <div class="mt-4 pt-3 border-t border-slate-100 flex items-end justify-between">
                <div>
                  <span class="text-[9px] uppercase font-bold text-slate-400 block">Current Balance</span>
                  <span class="text-lg font-black text-slate-900 font-mono">{{ formatCurrency(w.current_balance) }}</span>
                </div>
                <div class="text-right">
                  <span class="text-[9px] uppercase font-bold text-slate-400 block">Min Limit</span>
                  <span class="text-xs font-bold font-mono text-slate-600">{{ formatCurrency(w.minimum_limit) }}</span>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- SECTION 3: Expense Analysis by Category (last 7 days) -->
        <section class="flex flex-col gap-4">
          <div>
            <h2 class="text-xl font-extrabold text-slate-900 tracking-tight">3. Expense Analysis by Category (Last 7 Days)</h2>
            <p class="text-xs text-slate-400 font-medium">Distribution of actual expenses across categories</p>
          </div>

          <div class="card-neo p-6">
            <div v-if="expenseAnalysis.categories.length === 0" class="py-8 text-center text-xs text-slate-400 font-medium">
              No expense transactions recorded in the last 7 days.
            </div>

            <div v-else class="flex flex-col gap-4">
              <div class="flex items-baseline justify-between">
                <span class="text-xs font-bold text-slate-500 uppercase">Total 7-Day Expenses</span>
                <span class="text-2xl font-black text-slate-900">{{ formatCurrency(expenseAnalysis.total) }}</span>
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
              <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 mt-2">
                <div 
                  v-for="cat in expenseAnalysis.categories" 
                  :key="cat.name"
                  class="p-3 rounded-2xl bg-slate-50 border border-slate-100 flex items-center justify-between text-xs"
                >
                  <div class="flex items-center gap-2 min-w-0">
                    <span class="w-2.5 h-2.5 rounded-full shrink-0" :class="cat.color"></span>
                    <span class="font-bold text-slate-800 truncate">{{ cat.name }}</span>
                  </div>
                  <div class="text-right shrink-0">
                    <span class="font-bold text-slate-900 font-mono block">{{ formatCurrency(cat.amount) }}</span>
                    <span class="text-[10px] text-slate-400 font-semibold">{{ cat.percentage }}%</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- SECTION 4: Family Members (NO Dummy Members) -->
        <section class="flex flex-col gap-4">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-xl font-extrabold text-slate-900 tracking-tight">4. Family Members</h2>
              <p class="text-xs text-slate-400 font-medium">Registered members belonging to {{ familyStore.family?.name || 'this family' }}</p>
            </div>
            <div class="flex items-center gap-2 bg-slate-100 px-3 py-1 rounded-xl text-xs font-mono font-bold text-slate-700">
              <span>Invite Code:</span>
              <span class="text-brand-700 uppercase tracking-widest">{{ familyStore.family?.invite_code }}</span>
            </div>
          </div>

          <div v-if="realFamilyMembers.length === 0" class="card-neo p-8 text-center text-xs text-slate-400">
            No family members found.
          </div>

          <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            <div 
              v-for="member in realFamilyMembers" 
              :key="member.id"
              class="card-neo p-4 flex items-center justify-between gap-3"
            >
              <div class="flex items-center gap-3 min-w-0">
                <div class="w-10 h-10 rounded-2xl bg-gradient-to-tr from-slate-800 to-slate-600 text-white font-black text-sm flex items-center justify-center shrink-0">
                  {{ member.role === 'admin' ? '👑' : '👤' }}
                </div>
                <div class="min-w-0">
                  <h4 class="text-xs font-extrabold text-slate-900 truncate">Member #{{ member.user_id.slice(0, 6) }}</h4>
                  <span class="text-[10px] text-slate-400 font-medium">Joined {{ new Date(member.joined_at).toLocaleDateString('en-US') }}</span>
                </div>
              </div>
              <span 
                class="text-[10px] font-extrabold uppercase px-2.5 py-1 rounded-full shrink-0"
                :class="member.role === 'admin' ? 'bg-amber-100 text-amber-900 border border-amber-200' : 'bg-slate-100 text-slate-700'"
              >
                {{ member.role }}
              </span>
            </div>
          </div>
        </section>

        <!-- SECTION 5: Recent History (Last 10 Dated Transactions) -->
        <section class="flex flex-col gap-4">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-xl font-extrabold text-slate-900 tracking-tight">5. Recent History</h2>
              <p class="text-xs text-slate-400 font-medium">Last 10 recorded transactions</p>
            </div>
            <button 
              @click="activeTab = 'transactions'" 
              class="text-xs font-bold text-brand-600 hover:text-brand-700"
            >
              View Full History &rarr;
            </button>
          </div>

          <div class="card-neo overflow-hidden p-0 border border-slate-200/80">
            <div v-if="recentTenTransactions.length === 0" class="py-12 text-center text-xs text-slate-400 font-medium">
              No transactions recorded yet.
            </div>

            <div v-else class="overflow-x-auto">
              <table class="w-full text-left text-xs">
                <thead class="bg-slate-50 border-b border-slate-200/80 text-slate-400 font-bold uppercase tracking-wider text-[10px]">
                  <tr>
                    <th class="py-3.5 px-5">Date</th>
                    <th class="py-3.5 px-5">Wallet</th>
                    <th class="py-3.5 px-5">Category</th>
                    <th class="py-3.5 px-5">Description</th>
                    <th class="py-3.5 px-5 text-right">Amount</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                  <tr v-for="tx in recentTenTransactions" :key="tx.id" class="hover:bg-slate-50/80 transition font-medium">
                    <td class="py-3.5 px-5 text-slate-400 font-mono text-[11px] whitespace-nowrap">
                      {{ new Date(tx.created_at).toLocaleDateString('en-US', { day: 'numeric', month: 'short', year: 'numeric' }) }}
                    </td>
                    <td class="py-3.5 px-5 font-bold text-slate-800">{{ getWalletName(tx.wallet_id) }}</td>
                    <td class="py-3.5 px-5">
                      <span class="px-2 py-0.5 rounded-md bg-slate-100 text-slate-700 text-[10px] font-semibold">
                        {{ tx.category || 'General' }}
                      </span>
                    </td>
                    <td class="py-3.5 px-5 text-slate-800">{{ tx.description || '-' }}</td>
                    <td class="py-3.5 px-5 text-right font-mono font-black" :class="tx.type === 'income' ? 'text-emerald-600' : 'text-slate-900'">
                      {{ tx.type === 'income' ? '+' : '-' }}{{ formatCurrency(tx.amount) }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>

      </div>

      <!-- ========================================================================= -->
      <!-- TAB 2: WALLETS TAB                                                        -->
      <!-- ========================================================================= -->
      <div v-else-if="activeTab === 'wallets'" class="flex flex-col gap-6">
        <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
          <div>
            <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">Manage Wallet Envelopes</h2>
            <p class="text-xs text-slate-500 mt-0.5">Allocate and monitor family financial envelopes</p>
          </div>
          <button 
            v-if="isAdmin" 
            class="px-4 py-2.5 bg-slate-900 hover:bg-slate-800 text-white rounded-2xl text-xs font-bold transition shadow-sm flex items-center gap-2"
            @click="openWalletModal"
          >
            + Add New Wallet
          </button>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          <div 
            v-for="w in walletStore.wallets" 
            :key="w.id"
            class="card-neo p-6 flex flex-col justify-between group hover:border-brand-300"
          >
            <div>
              <div class="flex items-start justify-between">
                <div>
                  <span class="text-[10px] font-extrabold uppercase px-2 py-0.5 rounded-full bg-slate-100 text-slate-600">Envelope</span>
                  <h3 class="text-lg font-black text-slate-900 mt-1">{{ w.name }}</h3>
                </div>
                <div class="w-9 h-9 rounded-2xl bg-brand-50 text-brand-700 flex items-center justify-center font-bold">
                  💳
                </div>
              </div>
              <p class="text-xs text-slate-400 mt-2">{{ w.description || 'No description provided' }}</p>
            </div>

            <div class="mt-6 pt-4 border-t border-slate-100 flex items-end justify-between">
              <div>
                <span class="text-[10px] font-bold uppercase text-slate-400 block">Current Balance</span>
                <span class="text-xl font-black text-slate-900 font-mono">{{ formatCurrency(w.current_balance) }}</span>
              </div>
              <div class="text-right">
                <span class="text-[10px] font-bold uppercase text-slate-400 block">Minimum Limit</span>
                <span 
                  class="text-xs font-bold font-mono"
                  :class="w.current_balance <= w.minimum_limit ? 'text-rose-600' : 'text-slate-600'"
                >
                  {{ formatCurrency(w.minimum_limit) }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ========================================================================= -->
      <!-- TAB 3: TRANSACTION HISTORY TAB                                            -->
      <!-- ========================================================================= -->
      <div v-else-if="activeTab === 'transactions'" class="flex flex-col gap-6">
        <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
          <div>
            <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">Transaction History</h2>
            <p class="text-xs text-slate-500 mt-0.5">Complete log of all income and expenses</p>
          </div>
          <button 
            @click="openTxModal" 
            class="px-4 py-2.5 bg-slate-900 hover:bg-slate-800 text-white rounded-2xl text-xs font-bold transition shadow-sm"
          >
            + Record Transaction
          </button>
        </div>

        <div class="card-neo overflow-hidden p-0 border border-slate-200/80">
          <div v-if="txStore.transactions.length === 0" class="py-16 text-center text-xs text-slate-400">
            No transactions found.
          </div>

          <div v-else class="overflow-x-auto">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-50 border-b border-slate-200/80 text-slate-400 font-bold uppercase tracking-wider text-[10px]">
                <tr>
                  <th class="py-4 px-6">Timestamp</th>
                  <th class="py-4 px-6">Wallet</th>
                  <th class="py-4 px-6">Type</th>
                  <th class="py-4 px-6">Category</th>
                  <th class="py-4 px-6">Description</th>
                  <th class="py-4 px-6 text-right">Amount</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-for="tx in txStore.transactions" :key="tx.id" class="hover:bg-slate-50/80 transition font-medium">
                  <td class="py-4 px-6 text-slate-400 font-mono text-[11px] whitespace-nowrap">{{ new Date(tx.created_at).toLocaleString('en-US') }}</td>
                  <td class="py-4 px-6 font-bold text-slate-800">{{ getWalletName(tx.wallet_id) }}</td>
                  <td class="py-4 px-6">
                    <span 
                      class="px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-wide inline-block"
                      :class="tx.type === 'income' ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-rose-50 text-rose-700 border border-rose-200'"
                    >
                      {{ tx.type }}
                    </span>
                  </td>
                  <td class="py-4 px-6 text-slate-600">{{ tx.category || 'General' }}</td>
                  <td class="py-4 px-6 text-slate-800 font-semibold">{{ tx.description || '-' }}</td>
                  <td class="py-4 px-6 text-right font-mono font-black text-sm" :class="tx.type === 'income' ? 'text-emerald-600' : 'text-slate-900'">
                    {{ tx.type === 'income' ? '+' : '-' }}{{ formatCurrency(tx.amount) }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- ========================================================================= -->
      <!-- TAB 4: FINANCIAL REPORTS TAB                                              -->
      <!-- ========================================================================= -->
      <div v-else-if="activeTab === 'reports'" class="flex flex-col gap-6">
        <div>
          <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">Financial Reports</h2>
          <p class="text-xs text-slate-500 mt-0.5">Comprehensive analytics and spending reports</p>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div class="card-neo p-6 flex flex-col justify-between">
            <span class="text-xs font-bold text-slate-400 uppercase">Total Cumulative Inflows</span>
            <h3 class="text-3xl font-black text-emerald-600 my-2">
              +{{ formatCurrency((txStore.transactions || []).filter(t => t.type === 'income').reduce((s, t) => s + (t.amount || 0), 0)) }}
            </h3>
            <span class="text-xs text-slate-400">All historical income transactions</span>
          </div>

          <div class="card-neo p-6 flex flex-col justify-between">
            <span class="text-xs font-bold text-slate-400 uppercase">Total Cumulative Outflows</span>
            <h3 class="text-3xl font-black text-rose-600 my-2">
              -{{ formatCurrency((txStore.transactions || []).filter(t => t.type === 'expense').reduce((s, t) => s + (t.amount || 0), 0)) }}
            </h3>
            <span class="text-xs text-slate-400">All historical expense transactions</span>
          </div>

          <div class="card-neo p-6 flex flex-col justify-between">
            <span class="text-xs font-bold text-slate-400 uppercase">Net Wallet Balances</span>
            <h3 class="text-3xl font-black text-slate-900 my-2">
              {{ formatCurrency(totalBalance) }}
            </h3>
            <span class="text-xs text-slate-400">Current available funds</span>
          </div>
        </div>

        <!-- 7-Day Cashflow Performance Chart in Report -->
        <div class="card-neo p-6">
          <CashflowChart 
            :transactions="txStore.transactions"
            :total-balance="totalBalance"
            :total-income="sevenDaysIncome"
            :total-expense="sevenDaysExpense"
          />
        </div>
      </div>

      <!-- ========================================================================= -->
      <!-- TAB 5: TRANSACTION SUBMISSION & PROPOSALS TAB                             -->
      <!-- ========================================================================= -->
      <div v-else-if="activeTab === 'submissions'" class="flex flex-col gap-6">
        <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
          <div>
            <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">Transaction Submission &amp; Proposals</h2>
            <p class="text-xs text-slate-500 mt-0.5">Submit new transactions or expense proposals for Admin approval</p>
          </div>
          <div class="flex gap-2">
            <button 
              @click="openTxModal" 
              class="px-4 py-2.5 bg-slate-900 hover:bg-slate-800 text-white rounded-2xl text-xs font-bold transition shadow-sm"
            >
              + Record Direct Transaction
            </button>
            <button 
              @click="openProposalModal" 
              class="px-4 py-2.5 bg-brand-500 hover:bg-brand-600 text-slate-950 rounded-2xl text-xs font-bold transition shadow-sm"
            >
              + Submit Proposal
            </button>
          </div>
        </div>

        <div class="flex flex-col gap-4">
          <h3 class="font-extrabold text-base text-slate-900">Submitted Expense Proposals</h3>
          
          <div v-if="txStore.proposals.length === 0" class="card-neo p-12 text-center text-xs text-slate-400">
            No expense proposals submitted yet.
          </div>

          <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div 
              v-for="p in txStore.proposals" 
              :key="p.id"
              class="card-neo p-6 flex flex-col justify-between"
            >
              <div>
                <div class="flex items-start justify-between gap-3">
                  <h4 class="font-black text-base text-slate-900">{{ p.title }}</h4>
                  <span 
                    class="px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-wider"
                    :class="{
                      'bg-amber-50 text-amber-700 border border-amber-200': p.status === 'pending',
                      'bg-emerald-50 text-emerald-700 border border-emerald-200': p.status === 'approved',
                      'bg-rose-50 text-rose-700 border border-rose-200': p.status === 'rejected',
                    }"
                  >
                    {{ p.status }}
                  </span>
                </div>
                <p class="text-xs text-slate-500 mt-2 leading-relaxed">{{ p.description }}</p>
              </div>

              <div class="mt-6 pt-4 border-t border-slate-100 flex items-center justify-between">
                <div>
                  <span class="text-[10px] text-slate-400 font-bold uppercase block">Wallet: {{ getWalletName(p.wallet_id) }}</span>
                  <span class="text-xl font-black text-slate-900 font-mono">{{ formatCurrency(p.amount) }}</span>
                </div>

                <div v-if="isAdmin && p.status === 'pending'" class="flex gap-2">
                  <button 
                    @click="approveProp(p.id)" 
                    class="px-3.5 py-1.5 rounded-xl bg-emerald-500 hover:bg-emerald-600 text-white font-bold text-xs transition shadow-sm"
                  >
                    Approve
                  </button>
                  <button 
                    @click="rejectProp(p.id)" 
                    class="px-3.5 py-1.5 rounded-xl bg-rose-500 hover:bg-rose-600 text-white font-bold text-xs transition shadow-sm"
                  >
                    Reject
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
        <h3 class="font-black text-lg text-slate-900 mb-1">Create New Wallet</h3>
        <p class="text-xs text-slate-400 mb-4">Add a virtual envelope to organize family spending.</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Wallet Name</label>
            <input type="text" v-model="newWallet.name" placeholder="e.g. Food &amp; Groceries" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Description (Optional)</label>
            <input type="text" v-model="newWallet.description" placeholder="Monthly groceries &amp; dining" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1">Initial Balance ($)</label>
              <input type="number" v-model.number="newWallet.initial_balance" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
            </div>
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1">Min Limit ($)</label>
              <input type="number" v-model.number="newWallet.minimum_limit" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
            </div>
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeWalletModal" :disabled="isSubmitting">Cancel</button>
          <button class="btn bg-slate-900 hover:bg-slate-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitWallet" :disabled="isSubmitting || !newWallet.name">
            {{ isSubmitting ? 'Saving...' : 'Create Wallet' }}
          </button>
        </div>
      </div>
    </dialog>

    <!-- 2. Modal Create Transaction -->
    <dialog :class="isTxModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">Record Transaction</h3>
        <p class="text-xs text-slate-400 mb-4">Input direct expense or income transaction.</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Select Wallet</label>
            <select v-model="newTx.wallet_id" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400">
              <option v-for="w in walletStore.wallets" :key="w.id" :value="w.id">{{ w.name }} (Balance: {{ formatCurrency(w.current_balance) }})</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Transaction Type</label>
            <select v-model="newTx.type" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400">
              <option value="expense">Expense (-)</option>
              <option value="income">Income (+)</option>
            </select>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1">Amount ($)</label>
              <input type="number" v-model.number="newTx.amount" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
            </div>
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1">Category</label>
              <input type="text" v-model="newTx.category" placeholder="Food / Shopping / Fuel" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
            </div>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Description</label>
            <input type="text" v-model="newTx.description" placeholder="Description of item or receipt" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeTxModal" :disabled="isSubmitting">Cancel</button>
          <button class="btn bg-slate-900 hover:bg-slate-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitTx" :disabled="isSubmitting || !newTx.wallet_id || newTx.amount <= 0">
            {{ isSubmitting ? 'Saving...' : 'Record Transaction' }}
          </button>
        </div>
      </div>
    </dialog>

    <!-- 3. Modal Create Proposal -->
    <dialog :class="isProposalModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">Submit Expense Proposal</h3>
        <p class="text-xs text-slate-400 mb-4">Request approval from family Admin before spending.</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Target Wallet</label>
            <select v-model="newProposal.wallet_id" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400">
              <option v-for="w in walletStore.wallets" :key="w.id" :value="w.id">{{ w.name }}</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Proposal Title</label>
            <input type="text" v-model="newProposal.title" placeholder="e.g. Weekly Grocery Restock" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Amount ($)</label>
            <input type="number" v-model.number="newProposal.amount" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Reason / Notes</label>
            <input type="text" v-model="newProposal.description" placeholder="Reason for expenditure" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeProposalModal" :disabled="isSubmitting">Cancel</button>
          <button class="btn bg-slate-900 hover:bg-slate-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitProposal" :disabled="isSubmitting || !newProposal.title || newProposal.amount <= 0">
            {{ isSubmitting ? 'Submitting...' : 'Submit Proposal' }}
          </button>
        </div>
      </div>
    </dialog>

    <!-- 4. Modal Telegram Bot Info -->
    <dialog :class="isTelegramModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-black text-lg text-slate-900 mb-1">Telegram Bot Integration 🤖</h3>
        <p class="text-xs text-slate-400 mb-4">Send OTPs and record transactions directly from Telegram chat.</p>

        <div v-if="familyStore.family?.telegram_chat_id" class="p-4 rounded-2xl bg-emerald-50 border border-emerald-200 text-xs">
          <div class="flex items-center gap-2 text-emerald-800 font-bold mb-1">
            <span class="w-2.5 h-2.5 rounded-full bg-emerald-500 animate-pulse"></span>
            <span>Telegram Bot Connected</span>
          </div>
          <p class="text-emerald-700">Chat ID: <span class="font-mono font-bold">{{ familyStore.family.telegram_chat_id }}</span></p>
        </div>

        <div v-else class="flex flex-col gap-3 text-xs text-slate-600 bg-slate-50 p-4 rounded-2xl border border-slate-200/80">
          <p class="font-bold text-slate-800">How to link the bot:</p>
          <ol class="list-decimal list-inside space-y-1.5">
            <li>Open the ACIS Telegram Bot.</li>
            <li>Send the following command:</li>
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
            Disconnect Bot
          </button>
          <div v-else></div>
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="isTelegramModalOpen = false">Close</button>
        </div>
      </div>
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
