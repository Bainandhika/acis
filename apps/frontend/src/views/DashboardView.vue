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
import Sidebar from '../components/Sidebar.vue'
import Header from '../components/Header.vue'
import VirtualCard from '../components/VirtualCard.vue'
import CashflowChart from '../components/CashflowChart.vue'
import FinancialHealthGauge from '../components/FinancialHealthGauge.vue'

const walletStore = useWalletStore()
const authStore = useAuthStore()
const familyStore = useFamilyStore()
const txStore = useTransactionStore()
const router = useRouter()

// UI State
const activeTab = ref<'dashboard' | 'wallets' | 'transactions' | 'proposals'>('dashboard')
const isSidebarCollapsed = ref(false)
const searchQuery = ref('')
const selectedWalletId = ref('')

// Modal Controls
const isWalletModalOpen = ref(false)
const isTxModalOpen = ref(false)
const isProposalModalOpen = ref(false)
const isTelegramModalOpen = ref(false)
const isSubmitting = ref(false)

// Role Check
const isAdmin = computed(() => authStore.user?.role === 'admin')

// Form States
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
  category: 'Umum',
  description: '',
})

const newProposal = ref<CreateProposalPayload>({
  wallet_id: '',
  title: '',
  amount: 0,
  description: '',
})

// Metrics & Analytics Calculations
const totalBalance = computed(() => {
  return walletStore.wallets.reduce((sum, w) => sum + w.current_balance, 0)
})

const totalAllocation = computed(() => {
  return walletStore.wallets.reduce((sum, w) => sum + w.initial_balance, 0)
})

const monthlyIncome = computed(() => {
  return familyStore.family?.monthly_income || 0
})

const totalIncome = computed(() => {
  return txStore.transactions
    .filter(t => t.type === 'income')
    .reduce((sum, t) => sum + t.amount, 0) || (monthlyIncome.value > 0 ? monthlyIncome.value : 15000000)
})

const totalExpense = computed(() => {
  return txStore.transactions
    .filter(t => t.type === 'expense')
    .reduce((sum, t) => sum + t.amount, 0) || 6700000
})

const savedBalance = computed(() => {
  return Math.max(0, totalBalance.value)
})

const exceedsIncome = computed(() => {
  if (monthlyIncome.value <= 0) return false
  return (totalAllocation.value + newWallet.value.initial_balance) > monthlyIncome.value
})

const spendingLimitPercentage = computed(() => {
  if (monthlyIncome.value <= 0) return 65
  return Math.min(Math.round((totalExpense.value / monthlyIncome.value) * 100), 100)
})

// Category Breakdown for Cost Analysis (Segmented Bar matching ACRU)
const categoryBreakdown = computed(() => {
  const categories = [
    { name: 'Kebutuhan Rumah', amount: 3200000, color: 'bg-amber-400', percentage: 32 },
    { name: 'Cicilan & Tagihan', amount: 1500000, color: 'bg-lime-400', percentage: 22 },
    { name: 'Makanan & Belanja', amount: 1200000, color: 'bg-emerald-400', percentage: 18 },
    { name: 'Transportasi', amount: 800000, color: 'bg-sky-400', percentage: 12 },
    { name: 'Lain-lain', amount: 600000, color: 'bg-slate-300', percentage: 16 },
  ]
  return categories
})

// Filtered Transactions
const filteredTransactions = computed(() => {
  let list = txStore.transactions
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(t => 
      (t.description && t.description.toLowerCase().includes(q)) || 
      (t.category && t.category.toLowerCase().includes(q)) ||
      getWalletName(t.wallet_id).toLowerCase().includes(q)
    )
  }
  return list
})

// Recent 6 transactions for Right Column Feed
const recentTransactions = computed(() => {
  return filteredTransactions.value.slice(0, 6)
})

// Family members sample list matching ACRU avatar circles
const familyMembers = computed(() => {
  return [
    { name: authStore.user?.name || 'Anda', role: authStore.user?.role || 'Admin', avatar: '👨‍💼', color: 'bg-emerald-100 text-emerald-800' },
    { name: 'Pasangan', role: 'Member', avatar: '👩‍💼', color: 'bg-amber-100 text-amber-800' },
    { name: 'Anak 1', role: 'Member', avatar: '👦', color: 'bg-sky-100 text-sky-800' },
    { name: 'Anak 2', role: 'Member', avatar: '👧', color: 'bg-rose-100 text-rose-800' },
  ]
})

// Format currency
const formatRupiah = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(amount)
}

const getWalletName = (id: string) => {
  const w = walletStore.wallets.find(item => item.id === id)
  return w ? w.name : 'Dompet Utama'
}

onMounted(async () => {
  await familyStore.fetchMyFamily()
  if (!familyStore.family) {
    router.push('/family-setup')
    return
  }
  await walletStore.fetchWallets()
  if (walletStore.wallets.length > 0 && walletStore.wallets[0]) {
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
  if (confirm('Yakin ingin memutuskan koneksi bot Telegram?')) {
    try {
      await familyStore.handleDisconnectTelegram()
      isTelegramModalOpen.value = false
      showToast('Koneksi bot Telegram berhasil diputuskan', 'info')
    } catch {
      showToast('Gagal memutuskan koneksi Telegram', 'error')
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
    showToast('Dompet amplop berhasil dibuat!', 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal membuat dompet!', 'error')
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
  newTx.value = { wallet_id: '', type: 'expense', amount: 0, category: 'Umum', description: '' }
}
const handleSubmitTx = async () => {
  isSubmitting.value = true
  try {
    await txStore.addTransaction(newTx.value)
    await walletStore.fetchWallets()
    closeTxModal()
    showToast('Transaksi berhasil disimpan & saldo dompet terupdate!', 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal membuat transaksi!', 'error')
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
    showToast('Pengajuan pengeluaran berhasil dikirim ke Admin!', 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal mengajukan pengeluaran!', 'error')
  } finally {
    isSubmitting.value = false
  }
}

const approveProp = async (id: string) => {
  try {
    await txStore.handleApprove(id)
    await walletStore.fetchWallets()
    showToast('Pengajuan disetujui & saldo dompet dipotong!', 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal menyetujui pengajuan!', 'error')
  }
}

const rejectProp = async (id: string) => {
  try {
    await txStore.handleReject(id)
    showToast('Pengajuan ditolak', 'info')
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal menolak pengajuan!', 'error')
  }
}
</script>

<template>
  <div class="flex min-h-screen bg-[#F8FAFC]">
    <!-- 1. LEFT SIDEBAR (Matching ACRU Mockup) -->
    <Sidebar 
      :active-tab="activeTab"
      :is-collapsed="isSidebarCollapsed"
      @select-tab="activeTab = $event"
      @toggle-collapse="isSidebarCollapsed = !isSidebarCollapsed"
    />

    <!-- 2. MAIN APP CONTENT -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- Top Header Bar -->
      <Header 
        v-model:search-query="searchQuery"
        @open-tx-modal="openTxModal"
        @open-wallet-modal="openWalletModal"
        @open-proposal-modal="openProposalModal"
        @toggle-sidebar="isSidebarCollapsed = !isSidebarCollapsed"
      />

      <!-- Main View Body -->
      <main class="p-4 sm:p-6 lg:p-8 flex-1 max-w-[1600px] w-full mx-auto">
        
        <!-- ============================================== -->
        <!-- VIEW 1: DASHBOARD MAIN FINTECH OVERVIEW (ACRU) -->
        <!-- ============================================== -->
        <div v-if="activeTab === 'dashboard'" class="grid grid-cols-1 xl:grid-cols-12 gap-6 items-start">
          
          <!-- LEFT & CENTER COLUMN (8 of 12 Cols) -->
          <div class="xl:col-span-8 flex flex-col gap-6">
            
            <!-- TOP HERO ROW: Balance Chart + 3 Key Metrics Stack -->
            <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch">
              
              <!-- Cashflow Bar Chart (8 Cols on LG) -->
              <div class="lg:col-span-8">
                <CashflowChart 
                  :transactions="txStore.transactions"
                  :total-balance="totalBalance"
                  :total-income="totalIncome"
                  :total-expense="totalExpense"
                />
              </div>

              <!-- 3 Key Metric Cards Stack (4 Cols on LG, matching ACRU right stack) -->
              <div class="lg:col-span-4 flex flex-col gap-3.5 justify-between">
                
                <!-- Metric 1: Total Pemasukan -->
                <div class="card-neo p-4 flex flex-col justify-between">
                  <span class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Total Pemasukan</span>
                  <div class="my-1">
                    <h3 class="text-2xl font-black text-slate-900 tracking-tight">
                      {{ formatRupiah(totalIncome) }}
                    </h3>
                  </div>
                  <p class="text-[11px] font-bold text-emerald-600 flex items-center gap-1">
                    <svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                      <polyline points="18 15 12 9 6 15"></polyline>
                    </svg>
                    <span>+5.1% dari bulan lalu</span>
                  </p>
                </div>

                <!-- Metric 2: Total Pengeluaran -->
                <div class="card-neo p-4 flex flex-col justify-between">
                  <span class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Total Pengeluaran</span>
                  <div class="my-1">
                    <h3 class="text-2xl font-black text-slate-900 tracking-tight">
                      {{ formatRupiah(totalExpense) }}
                    </h3>
                  </div>
                  <p class="text-[11px] font-bold text-amber-500 flex items-center gap-1">
                    <svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                      <polyline points="6 9 12 15 18 9"></polyline>
                    </svg>
                    <span>15.5% dari batas limit</span>
                  </p>
                </div>

                <!-- Metric 3: Saldo Tersimpan / Alokasi -->
                <div class="card-neo p-4 flex flex-col justify-between">
                  <span class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Saldo Tersimpan</span>
                  <div class="my-1">
                    <h3 class="text-2xl font-black text-slate-900 tracking-tight">
                      {{ formatRupiah(savedBalance) }}
                    </h3>
                  </div>
                  <p class="text-[11px] font-bold text-emerald-600 flex items-center gap-1">
                    <svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                      <polyline points="18 15 12 9 6 15"></polyline>
                    </svg>
                    <span>+20.7% rasio tabungan</span>
                  </p>
                </div>
              </div>
            </div>

            <!-- MIDDLE ROW: Spending Limit Bar + Quick Tips Card -->
            <div class="grid grid-cols-1 md:grid-cols-12 gap-6">
              
              <!-- Monthly Spending Limit Progress (7 of 12) -->
              <div class="md:col-span-7 card-neo p-5 flex flex-col justify-between">
                <div class="flex items-center justify-between">
                  <div>
                    <h4 class="font-bold text-slate-900 text-sm">Limit Anggaran Bulanan</h4>
                    <p class="text-[11px] text-slate-400 font-medium">Batas belanja seluruh amplop keluarga</p>
                  </div>
                  <button 
                    v-if="isAdmin" 
                    @click="router.push('/family-setup')"
                    class="p-1.5 rounded-lg text-slate-400 hover:text-slate-700 hover:bg-slate-50 transition"
                    title="Ubah Target"
                  >
                    <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M12 20h9"></path>
                      <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path>
                    </svg>
                  </button>
                </div>

                <!-- Custom Green Segmented / Smooth Progress Bar -->
                <div class="my-4">
                  <div class="w-full bg-slate-100 rounded-full h-3 overflow-hidden p-0.5 border border-slate-200/50">
                    <div 
                      class="h-full rounded-full bg-gradient-to-r from-brand-400 to-brand-500 transition-all duration-700 shadow-sm"
                      :style="{ width: spendingLimitPercentage + '%' }"
                    ></div>
                  </div>
                  <div class="flex justify-between items-center text-xs font-extrabold mt-2 font-mono">
                    <span class="text-slate-800">{{ formatRupiah(totalExpense) }}</span>
                    <span class="text-slate-400">{{ formatRupiah(monthlyIncome > 0 ? monthlyIncome : 10000000) }}</span>
                  </div>
                </div>
              </div>

              <!-- Quick Insight / Telegram Assistant Card (5 of 12) -->
              <div class="md:col-span-5 card-neo p-5 bg-gradient-to-br from-white to-slate-50 flex flex-col justify-between relative overflow-hidden">
                <div class="flex items-start justify-between">
                  <div class="max-w-[190px]">
                    <h4 class="font-bold text-slate-900 text-sm">Optimalisasi Anggaran 💡</h4>
                    <p class="text-[11px] text-slate-500 mt-1 leading-relaxed">
                      {{ familyStore.family?.telegram_chat_id ? 'Telegram Bot aktif! Anggota keluarga bisa langsung catat belanja via chat.' : 'Hubungkan Telegram Bot untuk catat pengeluaran instan lewat chat grup!' }}
                    </p>
                  </div>
                  <!-- Mini graphic matching ACRU quick tips -->
                  <div class="w-12 h-12 rounded-2xl bg-brand-50 border border-brand-200 text-brand-600 flex items-center justify-center font-bold shrink-0">
                    🤖
                  </div>
                </div>

                <div class="mt-3 pt-2 border-t border-slate-100 flex items-center justify-between">
                  <button 
                    @click="isTelegramModalOpen = true" 
                    class="text-xs font-bold text-slate-900 hover:text-brand-600 flex items-center gap-1 transition"
                  >
                    <span>{{ familyStore.family?.telegram_chat_id ? 'Detail Koneksi Bot' : 'Hubungkan Bot Sekarang' }}</span>
                    <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                      <polyline points="9 18 15 12 9 6"></polyline>
                    </svg>
                  </button>
                </div>
              </div>
            </div>

            <!-- BOTTOM ROW: Cost Analysis + Financial Health + Goal Tracker -->
            <div class="grid grid-cols-1 md:grid-cols-12 gap-6">
              
              <!-- 1. Cost Analysis (4 of 12) -->
              <div class="md:col-span-4 card-neo p-5 flex flex-col justify-between">
                <div class="flex items-center justify-between">
                  <div>
                    <h4 class="font-bold text-slate-900 text-sm">Analisis Biaya</h4>
                    <p class="text-[11px] text-slate-400 font-medium">Berdasarkan kategori</p>
                  </div>
                  <span class="text-[11px] font-bold px-2 py-0.5 rounded-lg bg-slate-100 text-slate-600">Januari</span>
                </div>

                <div class="my-2">
                  <h3 class="text-2xl font-black text-slate-900 tracking-tight">
                    {{ formatRupiah(totalExpense) }}
                  </h3>
                </div>

                <!-- Segmented Multi-Color Horizontal Bar -->
                <div class="w-full flex h-2.5 rounded-full overflow-hidden gap-0.5 my-2">
                  <div class="bg-amber-400 h-full w-[30%]"></div>
                  <div class="bg-lime-400 h-full w-[25%]"></div>
                  <div class="bg-emerald-400 h-full w-[20%]"></div>
                  <div class="bg-sky-400 h-full w-[15%]"></div>
                  <div class="bg-slate-300 h-full w-[10%]"></div>
                </div>

                <!-- Categories Breakdown list -->
                <div class="flex flex-col gap-1.5 mt-2 text-[11px]">
                  <div v-for="cat in categoryBreakdown" :key="cat.name" class="flex items-center justify-between">
                    <div class="flex items-center gap-1.5">
                      <span class="w-2 h-2 rounded-full" :class="cat.color"></span>
                      <span class="text-slate-600 font-medium truncate max-w-[110px]">{{ cat.name }}</span>
                    </div>
                    <span class="font-bold text-slate-800 font-mono">{{ cat.percentage }}%</span>
                  </div>
                </div>
              </div>

              <!-- 2. Financial Health Meter (4 of 12) -->
              <div class="md:col-span-4">
                <FinancialHealthGauge 
                  :score-percentage="75"
                  :total-saved="savedBalance"
                />
              </div>

              <!-- 3. Goal & Envelope Tracker (4 of 12) -->
              <div class="md:col-span-4 card-neo p-5 flex flex-col justify-between">
                <div class="flex items-center justify-between">
                  <div>
                    <h4 class="font-bold text-slate-900 text-sm">Target Amplop</h4>
                    <p class="text-[11px] text-slate-400 font-medium">Batas minimal dompet</p>
                  </div>
                  <button 
                    v-if="isAdmin" 
                    @click="openWalletModal"
                    class="text-[11px] font-bold text-slate-700 bg-slate-100 hover:bg-slate-200 px-2 py-0.5 rounded-lg transition"
                  >
                    + Target
                  </button>
                </div>

                <!-- Envelopes items list -->
                <div class="flex flex-col gap-3.5 my-3">
                  <div v-for="w in walletStore.wallets.slice(0, 3)" :key="w.id" class="flex flex-col gap-1">
                    <div class="flex justify-between items-center text-[11px]">
                      <span class="font-bold text-slate-800 truncate max-w-[110px]">{{ w.name }}</span>
                      <span class="font-mono text-slate-500 font-semibold">{{ formatRupiah(w.current_balance) }} / {{ formatRupiah(w.minimum_limit || w.initial_balance) }}</span>
                    </div>
                    <div class="w-full bg-slate-100 rounded-full h-1.5 overflow-hidden">
                      <div 
                        class="h-full rounded-full transition-all"
                        :class="w.current_balance <= w.minimum_limit ? 'bg-rose-500' : 'bg-brand-500'"
                        :style="{ width: Math.min((w.current_balance / (w.minimum_limit || w.initial_balance || 1)) * 100, 100) + '%' }"
                      ></div>
                    </div>
                  </div>

                  <div v-if="walletStore.wallets.length === 0" class="text-center py-4 text-xs text-slate-400">
                    Belum ada data amplop.
                  </div>
                </div>

                <p class="text-[10px] text-slate-400 text-center border-t border-slate-100 pt-2">
                  Dipantau otomatis sistem amplop pintar.
                </p>
              </div>
            </div>

          </div>

          <!-- RIGHT COLUMN (4 of 12 Cols: Virtual Cards + Quick Members + Recent Feed) -->
          <div class="xl:col-span-4 flex flex-col gap-6">
            
            <!-- Widget 1: Virtual Card & Quick Actions -->
            <VirtualCard 
              :wallets="walletStore.wallets"
              :selected-wallet-id="selectedWalletId"
              @select-wallet="selectedWalletId = $event"
              @open-wallet-modal="openWalletModal"
              @open-tx-modal="openTxModal"
              @open-proposal-modal="openProposalModal"
              @link-telegram="isTelegramModalOpen = true"
            />

            <!-- Widget 2: Family Members (Quick Payee Avatars matching ACRU) -->
            <div class="card-neo p-5">
              <div class="flex items-center justify-between mb-3">
                <h4 class="font-bold text-slate-900 text-sm">Anggota Keluarga</h4>
                <span class="text-[11px] font-bold text-brand-600 bg-brand-50 px-2 py-0.5 rounded-full">
                  {{ familyMembers.length }} Aktif
                </span>
              </div>
              <div class="flex items-center justify-between gap-2 overflow-x-auto py-1">
                <div 
                  v-for="member in familyMembers" 
                  :key="member.name"
                  class="flex flex-col items-center gap-1.5 shrink-0 group cursor-pointer"
                >
                  <div class="w-11 h-11 rounded-2xl flex items-center justify-center text-lg border border-slate-200/80 shadow-sm transition group-hover:scale-105" :class="member.color">
                    {{ member.avatar }}
                  </div>
                  <span class="text-[10px] font-bold text-slate-700 truncate max-w-[55px] text-center">
                    {{ member.name }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Widget 3: Recent Transaction History Feed (Matching ACRU right column) -->
            <div class="card-neo p-5 flex flex-col gap-3">
              <div class="flex items-center justify-between border-b border-slate-100 pb-3">
                <div>
                  <h4 class="font-bold text-slate-900 text-sm">Riwayat Terkini</h4>
                  <p class="text-[11px] text-slate-400 font-medium">Transaksi terbaru</p>
                </div>
                <button 
                  @click="activeTab = 'transactions'" 
                  class="text-xs font-bold text-brand-600 hover:text-brand-700"
                >
                  Lihat Semua
                </button>
              </div>

              <!-- Transaction Feed Items -->
              <div class="flex flex-col divide-y divide-slate-100">
                <div 
                  v-for="tx in recentTransactions" 
                  :key="tx.id"
                  class="py-2.5 flex items-center justify-between gap-3 group hover:bg-slate-50/80 px-1.5 rounded-xl transition"
                >
                  <div class="flex items-center gap-3 min-w-0">
                    <!-- Icon Box -->
                    <div 
                      class="w-9 h-9 rounded-xl flex items-center justify-center shrink-0 text-xs font-bold"
                      :class="tx.type === 'income' ? 'bg-emerald-50 text-emerald-600 border border-emerald-100' : 'bg-slate-100 text-slate-700'"
                    >
                      <svg v-if="tx.type === 'income'" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                        <polyline points="18 15 12 9 6 15"></polyline>
                      </svg>
                      <svg v-else class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                        <polyline points="6 9 12 15 18 9"></polyline>
                      </svg>
                    </div>

                    <div class="min-w-0">
                      <p class="text-xs font-bold text-slate-800 truncate">
                        {{ tx.description || tx.category || 'Transaksi Tanpa Keterangan' }}
                      </p>
                      <p class="text-[10px] text-slate-400 flex items-center gap-1 font-medium">
                        <span>{{ getWalletName(tx.wallet_id) }}</span>
                        <span>•</span>
                        <span>{{ new Date(tx.created_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' }) }}</span>
                      </p>
                    </div>
                  </div>

                  <!-- Amount with + or - -->
                  <div class="text-right shrink-0">
                    <span 
                      class="text-xs font-extrabold font-mono"
                      :class="tx.type === 'income' ? 'text-emerald-600' : 'text-slate-900'"
                    >
                      {{ tx.type === 'income' ? '+' : '-' }}{{ formatRupiah(tx.amount) }}
                    </span>
                    <span class="text-[9px] block uppercase font-bold text-slate-400">Selesai</span>
                  </div>
                </div>

                <!-- Empty Fallback -->
                <div v-if="recentTransactions.length === 0" class="py-8 text-center text-xs text-slate-400">
                  Belum ada transaksi dicatat.
                </div>
              </div>
            </div>

          </div>
        </div>

        <!-- ============================================== -->
        <!-- VIEW 2: WALLETS TAB (DOMPET & AMPLOP)          -->
        <!-- ============================================== -->
        <div v-else-if="activeTab === 'wallets'" class="flex flex-col gap-6">
          <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
            <div>
              <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">Kelola Dompet & Amplop Virtual</h2>
              <p class="text-xs text-slate-500 mt-0.5">Alokasikan anggaran keluarga dengan sistem amplop terarah</p>
            </div>
            <button 
              v-if="isAdmin" 
              class="btn bg-slate-900 hover:bg-slate-800 text-white rounded-2xl px-5 text-xs font-bold border-none shadow-sm"
              @click="openWalletModal"
            >
              + Tambah Dompet Baru
            </button>
          </div>

          <!-- Summary Bar -->
          <div class="card-neo p-6 bg-gradient-to-r from-slate-900 to-slate-800 text-white flex flex-col md:flex-row justify-between items-start md:items-center gap-6">
            <div>
              <span class="text-xs font-bold uppercase tracking-wider text-brand-400">Ringkasan Anggaran</span>
              <h3 class="text-2xl font-black mt-1">Total Alokasi: {{ formatRupiah(totalAllocation) }}</h3>
              <p class="text-xs text-slate-300 mt-1">Estimasi Pendapatan: {{ monthlyIncome > 0 ? formatRupiah(monthlyIncome) : 'Belum diatur' }}</p>
            </div>
            <div class="w-full md:w-64">
              <div class="flex justify-between text-xs font-semibold mb-1">
                <span>Rasio Alokasi</span>
                <span :class="totalAllocation > monthlyIncome && monthlyIncome > 0 ? 'text-rose-400' : 'text-brand-400'">
                  {{ monthlyIncome > 0 ? Math.round((totalAllocation / monthlyIncome) * 100) : 0 }}%
                </span>
              </div>
              <div class="w-full bg-slate-700 rounded-full h-2.5 overflow-hidden">
                <div 
                  class="h-full rounded-full transition-all"
                  :class="totalAllocation > monthlyIncome && monthlyIncome > 0 ? 'bg-rose-500' : 'bg-brand-400'"
                  :style="{ width: Math.min((totalAllocation / (monthlyIncome || 1)) * 100, 100) + '%' }"
                ></div>
              </div>
            </div>
          </div>

          <!-- Wallets Grid -->
          <div v-if="walletStore.loading" class="flex justify-center py-20">
            <span class="loading loading-spinner loading-lg text-brand-500"></span>
          </div>

          <div v-else-if="walletStore.wallets.length === 0" class="card-neo p-12 text-center">
            <h3 class="text-lg font-bold text-slate-700">Belum Ada Dompet</h3>
            <p class="text-xs text-slate-400 mt-1">Mulai buat dompet pertama untuk mengelola alokasi belanja keluarga.</p>
            <button v-if="isAdmin" @click="openWalletModal" class="btn btn-primary btn-sm mt-4 rounded-xl">
              + Buat Dompet Pertama
            </button>
          </div>

          <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <div 
              v-for="wallet in walletStore.wallets" 
              :key="wallet.id"
              class="card-neo p-6 flex flex-col justify-between relative overflow-hidden group hover:border-brand-300"
            >
              <div>
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <span class="text-[10px] font-extrabold uppercase px-2 py-0.5 rounded-full bg-slate-100 text-slate-600">Amplop</span>
                    <h3 class="text-lg font-extrabold text-slate-900 mt-1.5">{{ wallet.name }}</h3>
                  </div>
                  <div class="w-8 h-8 rounded-xl bg-brand-50 text-brand-700 flex items-center justify-center font-bold">
                    💳
                  </div>
                </div>
                <p class="text-xs text-slate-400 mt-1">{{ wallet.description || 'Tidak ada deskripsi' }}</p>
              </div>

              <div class="mt-6 pt-4 border-t border-slate-100 flex items-end justify-between">
                <div>
                  <span class="text-[10px] font-bold uppercase text-slate-400 block">Saldo Saat Ini</span>
                  <span class="text-xl font-extrabold text-slate-900">{{ formatRupiah(wallet.current_balance) }}</span>
                </div>
                <div class="text-right">
                  <span class="text-[10px] font-bold uppercase text-slate-400 block">Limit Minimal</span>
                  <span 
                    class="text-xs font-bold font-mono"
                    :class="wallet.current_balance <= wallet.minimum_limit ? 'text-rose-600' : 'text-slate-600'"
                  >
                    {{ formatRupiah(wallet.minimum_limit) }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- ============================================== -->
        <!-- VIEW 3: TRANSACTIONS TAB (RIWAYAT TRANSAKSI)   -->
        <!-- ============================================== -->
        <div v-else-if="activeTab === 'transactions'" class="flex flex-col gap-6">
          <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
            <div>
              <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">Riwayat Seluruh Transaksi</h2>
              <p class="text-xs text-slate-500 mt-0.5">Catatan pengeluaran dan pemasukan keluarga secara realtime</p>
            </div>
            <button 
              v-if="isAdmin" 
              class="btn bg-slate-900 hover:bg-slate-800 text-white rounded-2xl px-5 text-xs font-bold border-none shadow-sm"
              @click="openTxModal"
            >
              + Catat Transaksi Baru
            </button>
          </div>

          <div v-if="txStore.loading" class="flex justify-center py-20">
            <span class="loading loading-spinner loading-lg text-brand-500"></span>
          </div>

          <div v-else-if="filteredTransactions.length === 0" class="card-neo p-12 text-center">
            <p class="text-sm font-semibold text-slate-400">Tidak ada transaksi yang cocok.</p>
          </div>

          <div v-else class="card-neo overflow-hidden p-0 border border-slate-200/70">
            <div class="overflow-x-auto">
              <table class="w-full text-left text-xs">
                <thead class="bg-slate-50 border-b border-slate-200/80 text-slate-400 font-bold uppercase tracking-wider text-[10px]">
                  <tr>
                    <th class="py-4 px-6">Waktu</th>
                    <th class="py-4 px-6">Dompet</th>
                    <th class="py-4 px-6">Tipe</th>
                    <th class="py-4 px-6">Kategori</th>
                    <th class="py-4 px-6">Keterangan</th>
                    <th class="py-4 px-6 text-right">Jumlah</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                  <tr v-for="tx in filteredTransactions" :key="tx.id" class="hover:bg-slate-50/80 transition font-medium">
                    <td class="py-4 px-6 text-slate-400 font-mono text-[11px]">{{ new Date(tx.created_at).toLocaleString('id-ID') }}</td>
                    <td class="py-4 px-6 font-bold text-slate-800">{{ getWalletName(tx.wallet_id) }}</td>
                    <td class="py-4 px-6">
                      <span 
                        class="px-2.5 py-1 rounded-full text-[10px] font-extrabold uppercase tracking-wide inline-block"
                        :class="tx.type === 'income' ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-rose-50 text-rose-700 border border-rose-200'"
                      >
                        {{ tx.type }}
                      </span>
                    </td>
                    <td class="py-4 px-6 text-slate-600">{{ tx.category || 'Umum' }}</td>
                    <td class="py-4 px-6 text-slate-800 font-semibold">{{ tx.description || '-' }}</td>
                    <td class="py-4 px-6 text-right font-mono font-extrabold text-sm" :class="tx.type === 'income' ? 'text-emerald-600' : 'text-slate-900'">
                      {{ tx.type === 'income' ? '+' : '-' }}{{ formatRupiah(tx.amount) }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- ============================================== -->
        <!-- VIEW 4: PROPOSALS TAB (PENGAJUAN DANA)         -->
        <!-- ============================================== -->
        <div v-else-if="activeTab === 'proposals'" class="flex flex-col gap-6">
          <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
            <div>
              <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">Pengajuan Anggaran Keluarga</h2>
              <p class="text-xs text-slate-500 mt-0.5">Semua anggota dapat mengajukan kebutuhan dana untuk persetujuan admin</p>
            </div>
            <button 
              class="btn bg-slate-900 hover:bg-slate-800 text-white rounded-2xl px-5 text-xs font-bold border-none shadow-sm"
              @click="openProposalModal"
            >
              + Ajukan Kebutuhan Baru
            </button>
          </div>

          <div v-if="txStore.loading" class="flex justify-center py-20">
            <span class="loading loading-spinner loading-lg text-brand-500"></span>
          </div>

          <div v-else-if="txStore.proposals.length === 0" class="card-neo p-12 text-center">
            <h3 class="text-base font-bold text-slate-700">Belum Ada Pengajuan</h3>
            <p class="text-xs text-slate-400 mt-1">Ajukan kebutuhan dana baru jika membutuhkan anggaran di luar rencana harian.</p>
          </div>

          <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div 
              v-for="p in txStore.proposals" 
              :key="p.id"
              class="card-neo p-6 flex flex-col justify-between"
            >
              <div>
                <div class="flex items-start justify-between gap-3">
                  <h3 class="font-extrabold text-base text-slate-900">{{ p.title }}</h3>
                  <span 
                    class="px-2.5 py-1 rounded-full text-[10px] font-extrabold uppercase tracking-wider"
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
                  <span class="text-[10px] text-slate-400 font-bold uppercase block">Target Dompet: {{ getWalletName(p.wallet_id) }}</span>
                  <span class="text-xl font-extrabold text-slate-900 font-mono">{{ formatRupiah(p.amount) }}</span>
                </div>

                <!-- Admin Action Buttons -->
                <div v-if="isAdmin && p.status === 'pending'" class="flex gap-2">
                  <button 
                    @click="approveProp(p.id)" 
                    class="px-3.5 py-1.5 rounded-xl bg-emerald-500 hover:bg-emerald-600 text-white font-bold text-xs transition shadow-sm"
                  >
                    Setujui
                  </button>
                  <button 
                    @click="rejectProp(p.id)" 
                    class="px-3.5 py-1.5 rounded-xl bg-rose-500 hover:bg-rose-600 text-white font-bold text-xs transition shadow-sm"
                  >
                    Tolak
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

      </main>
    </div>

    <!-- ============================================== -->
    <!-- MODALS & POPUPS                                -->
    <!-- ============================================== -->
    
    <!-- 1. Modal Create Wallet -->
    <dialog :class="isWalletModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-extrabold text-lg text-slate-900 mb-1">Buat Dompet Baru</h3>
        <p class="text-xs text-slate-400 mb-4">Tambahkan amplop virtual untuk mengkategorikan dana keluarga.</p>

        <div v-if="exceedsIncome" class="p-3 bg-rose-50 border border-rose-200 text-rose-700 text-xs rounded-2xl mb-4 font-medium">
          ⚠️ Total alokasi awal dompet melebihi pendapatan bulanan keluarga!
        </div>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Nama Dompet</label>
            <input type="text" v-model="newWallet.name" placeholder="Contoh: Makan Bulanan" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Deskripsi (Opsional)</label>
            <input type="text" v-model="newWallet.description" placeholder="Belanja bahan makanan & dapur" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1">Saldo Awal (Rp)</label>
              <input type="number" v-model.number="newWallet.initial_balance" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
            </div>
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1">Limit Min (Rp)</label>
              <input type="number" v-model.number="newWallet.minimum_limit" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
            </div>
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeWalletModal" :disabled="isSubmitting">Batal</button>
          <button class="btn bg-slate-900 hover:bg-slate-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitWallet" :disabled="isSubmitting || !newWallet.name || exceedsIncome">
            {{ isSubmitting ? 'Menyimpan...' : 'Simpan Dompet' }}
          </button>
        </div>
      </div>
    </dialog>

    <!-- 2. Modal Create Transaction -->
    <dialog :class="isTxModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-extrabold text-lg text-slate-900 mb-1">Catat Transaksi</h3>
        <p class="text-xs text-slate-400 mb-4">Input transaksi pengeluaran atau pemasukan langsung.</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Pilih Dompet</label>
            <select v-model="newTx.wallet_id" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400">
              <option v-for="w in walletStore.wallets" :key="w.id" :value="w.id">{{ w.name }} (Saldo: {{ formatRupiah(w.current_balance) }})</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Tipe Transaksi</label>
            <select v-model="newTx.type" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400">
              <option value="expense">Pengeluaran (Expense)</option>
              <option value="income">Pemasukan (Income)</option>
            </select>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1">Jumlah (Rp)</label>
              <input type="number" v-model.number="newTx.amount" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
            </div>
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1">Kategori</label>
              <input type="text" v-model="newTx.category" placeholder="Makanan / Utilitas" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
            </div>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Keterangan</label>
            <input type="text" v-model="newTx.description" placeholder="Beli galon & gas" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeTxModal" :disabled="isSubmitting">Batal</button>
          <button class="btn bg-slate-900 hover:bg-slate-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitTx" :disabled="isSubmitting || !newTx.wallet_id || newTx.amount <= 0">
            {{ isSubmitting ? 'Menyimpan...' : 'Catat Transaksi' }}
          </button>
        </div>
      </div>
    </dialog>

    <!-- 3. Modal Create Proposal -->
    <dialog :class="isProposalModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-extrabold text-lg text-slate-900 mb-1">Ajukan Pengeluaran Baru</h3>
        <p class="text-xs text-slate-400 mb-4">Pengajuan akan ditinjau oleh Admin keluarga sebelum disetujui.</p>

        <div class="flex flex-col gap-3.5">
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Pilih Dompet Target</label>
            <select v-model="newProposal.wallet_id" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400">
              <option v-for="w in walletStore.wallets" :key="w.id" :value="w.id">{{ w.name }}</option>
            </select>
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Judul Pengajuan</label>
            <input type="text" v-model="newProposal.title" placeholder="Contoh: Beli Popok Bayi" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Nominal (Rp)</label>
            <input type="number" v-model.number="newProposal.amount" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
          <div>
            <label class="text-xs font-bold text-slate-700 block mb-1">Alasan / Deskripsi</label>
            <input type="text" v-model="newProposal.description" placeholder="Popok stok mingguan habis" class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-brand-400" />
          </div>
        </div>

        <div class="modal-action mt-6 gap-2">
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="closeProposalModal" :disabled="isSubmitting">Batal</button>
          <button class="btn bg-slate-900 hover:bg-slate-800 text-white btn-sm rounded-xl text-xs font-bold border-none" @click="handleSubmitProposal" :disabled="isSubmitting || !newProposal.title || newProposal.amount <= 0">
            {{ isSubmitting ? 'Mengirim...' : 'Kirim Pengajuan' }}
          </button>
        </div>
      </div>
    </dialog>

    <!-- 4. Modal Telegram Bot Info -->
    <dialog :class="isTelegramModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box bg-white rounded-3xl p-6 shadow-2xl border border-slate-100 max-w-md">
        <h3 class="font-extrabold text-lg text-slate-900 mb-1">Integrasi Telegram Bot 🤖</h3>
        <p class="text-xs text-slate-400 mb-4">Catat transaksi langsung dari obrolan bot Telegram grup keluarga.</p>

        <div v-if="familyStore.family?.telegram_chat_id" class="p-4 rounded-2xl bg-emerald-50 border border-emerald-200 text-xs">
          <div class="flex items-center gap-2 text-emerald-800 font-bold mb-1">
            <span class="w-2.5 h-2.5 rounded-full bg-emerald-500 animate-pulse"></span>
            <span>Bot Telegram Terhubung</span>
          </div>
          <p class="text-emerald-700">Chat ID: <span class="font-mono font-bold">{{ familyStore.family.telegram_chat_id }}</span></p>
        </div>

        <div v-else class="flex flex-col gap-3 text-xs text-slate-600 bg-slate-50 p-4 rounded-2xl border border-slate-200/80">
          <p class="font-bold text-slate-800">Cara menghubungkan bot:</p>
          <ol class="list-decimal list-inside space-y-1.5">
            <li>Buka bot Telegram ACIS.</li>
            <li>Kirim perintah berikut ke bot:</li>
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
            Putuskan Koneksi
          </button>
          <div v-else></div>
          <button class="btn btn-ghost btn-sm rounded-xl text-xs font-bold" @click="isTelegramModalOpen = false">Tutup</button>
        </div>
      </div>
    </dialog>

    <!-- 5. Floating Toast Container -->
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

