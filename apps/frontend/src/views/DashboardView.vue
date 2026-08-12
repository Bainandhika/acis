<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useWalletStore } from '../stores/wallet'
import { useAuthStore } from '../stores/auth'
import { useFamilyStore } from '../stores/family'
import { useTransactionStore } from '../stores/transaction'
import { useRouter } from 'vue-router'
import type { CreateWalletPayload } from '../services/wallet'
import type { CreateTransactionPayload, CreateProposalPayload } from '../services/transaction'

const walletStore = useWalletStore()
const authStore = useAuthStore()
const familyStore = useFamilyStore()
const txStore = useTransactionStore()
const router = useRouter()

const activeTab = ref<'wallets' | 'transactions' | 'proposals'>('wallets')

// Modal Controls
const isWalletModalOpen = ref(false)
const isTxModalOpen = ref(false)
const isProposalModalOpen = ref(false)
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

onMounted(async () => {
  await familyStore.fetchMyFamily()
  if (!familyStore.family) {
    router.push('/family-setup')
    return
  }
  walletStore.fetchWallets()
  txStore.fetchTransactions()
  txStore.fetchProposals()
})

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

// Wallet Modal Handlers
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
  } catch (error) {
    alert('Gagal membuat dompet!')
  } finally {
    isSubmitting.value = false
  }
}

// Transaction Modal Handlers
const openTxModal = () => {
  if (walletStore.wallets.length > 0) {
    newTx.value.wallet_id = walletStore.wallets[0].id
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
  } catch (error) {
    alert('Gagal membuat transaksi!')
  } finally {
    isSubmitting.value = false
  }
}

// Proposal Modal Handlers
const openProposalModal = () => {
  if (walletStore.wallets.length > 0) {
    newProposal.value.wallet_id = walletStore.wallets[0].id
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
  } catch (error) {
    alert('Gagal mengajukan pengeluaran!')
  } finally {
    isSubmitting.value = false
  }
}

// Approval Handlers
const approveProp = async (id: string) => {
  try {
    await txStore.handleApprove(id)
    await walletStore.fetchWallets()
  } catch (err) {
    alert('Gagal menyetujui pengajuan!')
  }
}

const rejectProp = async (id: string) => {
  try {
    await txStore.handleReject(id)
  } catch (err) {
    alert('Gagal menolak pengajuan!')
  }
}

// Helpers
const formatRupiah = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(amount)
}

const getWalletName = (id: string) => {
  const w = walletStore.wallets.find(item => item.id === id)
  return w ? w.name : 'Unknown'
}
</script>

<template>
  <div class="min-h-screen bg-base-200">
    <!-- Navbar -->
    <div class="navbar bg-base-100 shadow-md px-4 md:px-10">
      <div class="flex-1">
        <a class="btn btn-ghost text-xl font-bold text-primary">ACIS 💰</a>
        <span v-if="familyStore.family" class="badge badge-secondary ml-2 font-semibold">
          {{ familyStore.family.name }} (Invite: {{ familyStore.family.invite_code }})
        </span>
      </div>
      <div class="flex-none gap-4 items-center">
        <div class="text-right hidden sm:block">
          <p class="text-sm font-semibold">{{ authStore.user?.name }} ({{ authStore.user?.role }})</p>
          <p class="text-xs text-gray-500">{{ authStore.user?.email }}</p>
        </div>
        <button class="btn btn-outline btn-sm btn-error" @click="handleLogout">Logout</button>
      </div>
    </div>

    <!-- Main Content -->
    <div class="container mx-auto p-4 md:p-10">
      <!-- Tabs Navigation -->
      <div class="tabs tabs-boxed mb-6 bg-base-100 p-2 shadow-sm">
        <a class="tab" :class="{ 'tab-active': activeTab === 'wallets' }" @click="activeTab = 'wallets'">Dompet</a>
        <a class="tab" :class="{ 'tab-active': activeTab === 'transactions' }" @click="activeTab = 'transactions'">Transaksi</a>
        <a class="tab" :class="{ 'tab-active': activeTab === 'proposals' }" @click="activeTab = 'proposals'">Pengajuan Proposal</a>
      </div>

      <!-- TAB 1: DOMPET -->
      <div v-if="activeTab === 'wallets'">
        <div class="flex justify-between items-center mb-6">
          <h1 class="text-3xl font-bold">Dompet Keluarga</h1>
          <!-- Task 3.4: Role Guard - Admin Only -->
          <button v-if="isAdmin" class="btn btn-primary" @click="openWalletModal">+ Tambah Dompet</button>
        </div>

        <div v-if="walletStore.loading" class="flex justify-center py-20">
          <span class="loading loading-spinner loading-lg text-primary"></span>
        </div>

        <div v-else-if="walletStore.wallets.length === 0" class="text-center py-20 bg-base-100 rounded-lg shadow">
          <h2 class="text-2xl font-bold text-gray-400">Belum ada dompet</h2>
          <p class="text-gray-500 mt-2">Yuk buat dompet pertama buat mulai nyatat keuangan!</p>
        </div>

        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div v-for="wallet in walletStore.wallets" :key="wallet.id" class="card bg-base-100 shadow-xl border border-base-300">
            <div class="card-body">
              <h2 class="card-title text-lg">{{ wallet.name }}</h2>
              <p class="text-sm text-gray-500 mb-4">{{ wallet.description || 'Tidak ada deskripsi' }}</p>
              <div class="divider my-1"></div>
              <div class="flex justify-between items-end">
                <div>
                  <p class="text-xs text-gray-400 uppercase">Saldo Saat Ini</p>
                  <p class="text-2xl font-bold text-primary">{{ formatRupiah(wallet.current_balance) }}</p>
                </div>
                <div class="text-right">
                  <p class="text-xs text-gray-400">Limit Min</p>
                  <p class="text-sm font-semibold" :class="wallet.current_balance <= wallet.minimum_limit ? 'text-error' : 'text-success'">
                    {{ formatRupiah(wallet.minimum_limit) }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- TAB 2: TRANSAKSI (Task 3.2) -->
      <div v-else-if="activeTab === 'transactions'">
        <div class="flex justify-between items-center mb-6">
          <h1 class="text-3xl font-bold">Riwayat Transaksi</h1>
          <!-- Task 3.4: Role Guard - Admin Only -->
          <button v-if="isAdmin" class="btn btn-primary" @click="openTxModal">+ Catat Transaksi</button>
        </div>

        <div v-if="txStore.loading" class="flex justify-center py-20">
          <span class="loading loading-spinner loading-lg text-primary"></span>
        </div>

        <div v-else-if="txStore.transactions.length === 0" class="text-center py-20 bg-base-100 rounded-lg shadow">
          <h2 class="text-2xl font-bold text-gray-400">Belum ada transaksi</h2>
        </div>

        <div v-else class="overflow-x-auto bg-base-100 rounded-lg shadow border border-base-300">
          <table class="table w-full">
            <thead>
              <tr>
                <th>Waktu</th>
                <th>Dompet</th>
                <th>Tipe</th>
                <th>Kategori</th>
                <th>Keterangan</th>
                <th>Jumlah</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="tx in txStore.transactions" :key="tx.id">
                <td class="text-xs">{{ new Date(tx.created_at).toLocaleString('id-ID') }}</td>
                <td class="font-semibold">{{ getWalletName(tx.wallet_id) }}</td>
                <td>
                  <span class="badge" :class="tx.type === 'income' ? 'badge-success text-white' : 'badge-error text-white'">
                    {{ tx.type.toUpperCase() }}
                  </span>
                </td>
                <td>{{ tx.category || '-' }}</td>
                <td>{{ tx.description || '-' }}</td>
                <td class="font-bold" :class="tx.type === 'income' ? 'text-success' : 'text-error'">
                  {{ tx.type === 'income' ? '+' : '-' }} {{ formatRupiah(tx.amount) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- TAB 3: PROPOSAL / PENGAJUAN (Task 3.3) -->
      <div v-else-if="activeTab === 'proposals'">
        <div class="flex justify-between items-center mb-6">
          <h1 class="text-3xl font-bold">Pengajuan Pengeluaran</h1>
          <button class="btn btn-primary" @click="openProposalModal">+ Ajukan Pengeluaran</button>
        </div>

        <div v-if="txStore.loading" class="flex justify-center py-20">
          <span class="loading loading-spinner loading-lg text-primary"></span>
        </div>

        <div v-else-if="txStore.proposals.length === 0" class="text-center py-20 bg-base-100 rounded-lg shadow">
          <h2 class="text-2xl font-bold text-gray-400">Belum ada proposal pengeluaran</h2>
        </div>

        <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div v-for="p in txStore.proposals" :key="p.id" class="card bg-base-100 shadow-xl border border-base-300">
            <div class="card-body">
              <div class="flex justify-between items-start">
                <h2 class="card-title text-lg">{{ p.title }}</h2>
                <span class="badge" :class="{
                  'badge-warning': p.status === 'pending',
                  'badge-success text-white': p.status === 'approved',
                  'badge-error text-white': p.status === 'rejected'
                }">
                  {{ p.status.toUpperCase() }}
                </span>
              </div>
              <p class="text-sm text-gray-500">{{ p.description }}</p>
              <div class="divider my-1"></div>
              <div class="flex justify-between items-center">
                <div>
                  <p class="text-xs text-gray-400">Target Dompet: <span class="font-semibold text-base-content">{{ getWalletName(p.wallet_id) }}</span></p>
                  <p class="text-xl font-bold text-primary mt-1">{{ formatRupiah(p.amount) }}</p>
                </div>
                <!-- Task 3.4: Role Guard - Admin Actions Only -->
                <div v-if="isAdmin && p.status === 'pending'" class="flex gap-2">
                  <button class="btn btn-sm btn-success text-white" @click="approveProp(p.id)">Approve</button>
                  <button class="btn btn-sm btn-error text-white" @click="rejectProp(p.id)">Reject</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal Create Wallet -->
    <dialog :class="isWalletModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box">
        <h3 class="font-bold text-lg mb-4">Buat Dompet Baru</h3>
        <div class="form-control w-full mb-3">
          <label class="label"><span class="label-text">Nama Dompet</span></label>
          <input type="text" v-model="newWallet.name" placeholder="Contoh: Makan Bulanan" class="input input-bordered w-full" />
        </div>
        <div class="form-control w-full mb-3">
          <label class="label"><span class="label-text">Deskripsi (Opsional)</span></label>
          <input type="text" v-model="newWallet.description" placeholder="Belanja harian" class="input input-bordered w-full" />
        </div>
        <div class="grid grid-cols-2 gap-4 mb-4">
          <div class="form-control w-full">
            <label class="label"><span class="label-text">Saldo Awal (Rp)</span></label>
            <input type="number" v-model.number="newWallet.initial_balance" class="input input-bordered w-full" />
          </div>
          <div class="form-control w-full">
            <label class="label"><span class="label-text">Limit Minimum (Rp)</span></label>
            <input type="number" v-model.number="newWallet.minimum_limit" class="input input-bordered w-full" />
          </div>
        </div>
        <div class="modal-action">
          <button class="btn" @click="closeWalletModal" :disabled="isSubmitting">Batal</button>
          <button class="btn btn-primary" @click="handleSubmitWallet" :class="{ loading: isSubmitting }" :disabled="isSubmitting || !newWallet.name">Simpan</button>
        </div>
      </div>
    </dialog>

    <!-- Modal Create Transaction -->
    <dialog :class="isTxModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box">
        <h3 class="font-bold text-lg mb-4">Catat Transaksi Langsung</h3>
        <div class="form-control w-full mb-3">
          <label class="label"><span class="label-text">Pilih Dompet</span></label>
          <select v-model="newTx.wallet_id" class="select select-bordered w-full">
            <option v-for="w in walletStore.wallets" :key="w.id" :value="w.id">{{ w.name }} (Saldo: {{ formatRupiah(w.current_balance) }})</option>
          </select>
        </div>
        <div class="form-control w-full mb-3">
          <label class="label"><span class="label-text">Tipe Transaksi</span></label>
          <select v-model="newTx.type" class="select select-bordered w-full">
            <option value="expense">Pengeluaran (Expense)</option>
            <option value="income">Pemasukan (Income)</option>
          </select>
        </div>
        <div class="grid grid-cols-2 gap-4 mb-3">
          <div class="form-control w-full">
            <label class="label"><span class="label-text">Jumlah (Rp)</span></label>
            <input type="number" v-model.number="newTx.amount" class="input input-bordered w-full" />
          </div>
          <div class="form-control w-full">
            <label class="label"><span class="label-text">Kategori</span></label>
            <input type="text" v-model="newTx.category" placeholder="Makanan / Utilitas" class="input input-bordered w-full" />
          </div>
        </div>
        <div class="form-control w-full mb-4">
          <label class="label"><span class="label-text">Keterangan</span></label>
          <input type="text" v-model="newTx.description" placeholder="Keterangan transaksi" class="input input-bordered w-full" />
        </div>
        <div class="modal-action">
          <button class="btn" @click="closeTxModal" :disabled="isSubmitting">Batal</button>
          <button class="btn btn-primary" @click="handleSubmitTx" :class="{ loading: isSubmitting }" :disabled="isSubmitting || !newTx.wallet_id || newTx.amount <= 0">Simpan</button>
        </div>
      </div>
    </dialog>

    <!-- Modal Create Proposal -->
    <dialog :class="isProposalModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box">
        <h3 class="font-bold text-lg mb-4">Ajukan Pengeluaran Baru</h3>
        <div class="form-control w-full mb-3">
          <label class="label"><span class="label-text">Pilih Dompet Target</span></label>
          <select v-model="newProposal.wallet_id" class="select select-bordered w-full">
            <option v-for="w in walletStore.wallets" :key="w.id" :value="w.id">{{ w.name }}</option>
          </select>
        </div>
        <div class="form-control w-full mb-3">
          <label class="label"><span class="label-text">Judul Pengajuan</span></label>
          <input type="text" v-model="newProposal.title" placeholder="Beli Popok Bayi" class="input input-bordered w-full" />
        </div>
        <div class="form-control w-full mb-3">
          <label class="label"><span class="label-text">Nominal (Rp)</span></label>
          <input type="number" v-model.number="newProposal.amount" class="input input-bordered w-full" />
        </div>
        <div class="form-control w-full mb-4">
          <label class="label"><span class="label-text">Alasan / Deskripsi</span></label>
          <input type="text" v-model="newProposal.description" placeholder="Popok stok mingguan habis" class="input input-bordered w-full" />
        </div>
        <div class="modal-action">
          <button class="btn" @click="closeProposalModal" :disabled="isSubmitting">Batal</button>
          <button class="btn btn-primary" @click="handleSubmitProposal" :class="{ loading: isSubmitting }" :disabled="isSubmitting || !newProposal.title || newProposal.amount <= 0">Kirim Pengajuan</button>
        </div>
      </div>
    </dialog>
  </div>
</template>
