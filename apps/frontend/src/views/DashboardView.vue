<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useWalletStore } from '../stores/wallet'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
import type { CreateWalletPayload } from '../services/wallet'

const walletStore = useWalletStore()
const authStore = useAuthStore()
const router = useRouter()

// State buat Modal
const isModalOpen = ref(false)
const newWallet = ref<CreateWalletPayload>({
  family_id: '',
  name: '',
  description: '',
  initial_balance: 0,
  minimum_limit: 0,
})
const isSubmitting = ref(false)

onMounted(() => {
  walletStore.fetchWallets()
})

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

const openModal = () => {
  isModalOpen.value = true
}
const closeModal = () => {
  isModalOpen.value = false
  // Reset form
  newWallet.value = {
    family_id: '',
    name: '',
    description: '',
    initial_balance: 0,
    minimum_limit: 0,
  }
}

const handleSubmitWallet = async () => {
  isSubmitting.value = true
  try {
    await walletStore.addWallet(newWallet.value)
    closeModal()
  } catch (error: unknown) {
    console.error(error)
    alert('Gagal membuat dompet!')
  } finally {
    isSubmitting.value = false
  }
}

// Helper format Rupiah
const formatRupiah = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(amount)
}
</script>

<template>
  <div class="min-h-screen bg-base-200">
    <!-- Navbar -->
    <div class="navbar bg-base-100 shadow-md px-4 md:px-10">
      <div class="flex-1">
        <a class="btn btn-ghost text-xl font-bold text-primary">ACIS 💰</a>
      </div>
      <div class="flex-none gap-4 items-center">
        <div class="text-right hidden sm:block">
          <p class="text-sm font-semibold">{{ authStore.user?.name }}</p>
          <p class="text-xs text-gray-500">{{ authStore.user?.email }}</p>
        </div>
        <button class="btn btn-outline btn-sm btn-error" @click="handleLogout">Logout</button>
      </div>
    </div>

    <!-- Main Content -->
    <div class="container mx-auto p-4 md:p-10">
      <div class="flex justify-between items-center mb-6">
        <h1 class="text-3xl font-bold">Dompet Keluarga</h1>
        <button class="btn btn-primary" @click="openModal">+ Tambah Dompet</button>
      </div>

      <!-- Loading State -->
      <div v-if="walletStore.loading" class="flex justify-center py-20">
        <span class="loading loading-spinner loading-lg text-primary"></span>
      </div>

      <!-- Empty State -->
      <div
        v-else-if="walletStore.wallets.length === 0"
        class="text-center py-20 bg-base-100 rounded-lg shadow"
      >
        <h2 class="text-2xl font-bold text-gray-400">Belum ada dompet</h2>
        <p class="text-gray-500 mt-2">Yuk buat dompet pertama buat mulai nyatat keuangan!</p>
      </div>

      <!-- Wallet Grid -->
      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div
          v-for="wallet in walletStore.wallets"
          :key="wallet.id"
          class="card bg-base-100 shadow-xl border border-base-300"
        >
          <div class="card-body">
            <h2 class="card-title text-lg">{{ wallet.name }}</h2>
            <p class="text-sm text-gray-500 mb-4">
              {{ wallet.description || 'Tidak ada deskripsi' }}
            </p>

            <div class="divider my-1"></div>

            <div class="flex justify-between items-end">
              <div>
                <p class="text-xs text-gray-400 uppercase">Saldo Saat Ini</p>
                <p class="text-2xl font-bold text-primary">
                  {{ formatRupiah(wallet.current_balance) }}
                </p>
              </div>
              <div class="text-right">
                <p class="text-xs text-gray-400">Limit Min</p>
                <p
                  class="text-sm font-semibold"
                  :class="
                    wallet.current_balance <= wallet.minimum_limit ? 'text-error' : 'text-success'
                  "
                >
                  {{ formatRupiah(wallet.minimum_limit) }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal Create Wallet -->
    <dialog :class="isModalOpen ? 'modal modal-open' : 'modal'">
      <div class="modal-box">
        <h3 class="font-bold text-lg mb-4">Buat Dompet Baru</h3>

        <div class="form-control w-full mb-3">
          <label class="label"><span class="label-text">Nama Dompet</span></label>
          <input
            type="text"
            v-model="newWallet.name"
            placeholder="Contoh: Makan Bulanan"
            class="input input-bordered w-full"
          />
        </div>

        <div class="form-control w-full mb-3">
          <label class="label"><span class="label-text">Deskripsi (Opsional)</span></label>
          <input
            type="text"
            v-model="newWallet.description"
            placeholder="Untuk belanja harian"
            class="input input-bordered w-full"
          />
        </div>

        <div class="grid grid-cols-2 gap-4 mb-4">
          <div class="form-control w-full">
            <label class="label"><span class="label-text">Saldo Awal (Rp)</span></label>
            <input
              type="number"
              v-model.number="newWallet.initial_balance"
              class="input input-bordered w-full"
            />
          </div>
          <div class="form-control w-full">
            <label class="label"><span class="label-text">Limit Minimum (Rp)</span></label>
            <input
              type="number"
              v-model.number="newWallet.minimum_limit"
              class="input input-bordered w-full"
            />
          </div>
        </div>

        <div class="modal-action">
          <button class="btn" @click="closeModal" :disabled="isSubmitting">Batal</button>
          <button
            class="btn btn-primary"
            @click="handleSubmitWallet"
            :class="{ loading: isSubmitting }"
            :disabled="isSubmitting || !newWallet.name"
          >
            Simpan
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop" @submit.prevent="closeModal">
        <button>close</button>
      </form>
    </dialog>
  </div>
</template>
