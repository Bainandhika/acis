<script setup lang="ts">
import { ref, computed } from 'vue'
import { useProposals } from '../../../composables/useProposals'
import { useWallets } from '../../../composables/useWallets'
import { useAuthStore } from '../../../stores/auth'
import { formatRp, formatDate } from '../../../utils/format'
import { useI18n } from '../../../locales'
import type { ProposalStatus } from '../../../types'

// Components
import Button from '../../../components/ui/Button.vue'
import Skeleton from '../../../components/ui/Skeleton.vue'
import CreateProposalModal from './components/CreateProposalModal.vue'

const {
  filteredProposals,
  statusFilter,
  isLoading,
  approveProposal,
  rejectProposal,
  isApproving,
  isRejecting
} = useProposals()

const { wallets } = useWallets()
const authStore = useAuthStore()
const { t } = useI18n()

const isAdmin = computed(() => authStore.user?.role === 'admin')
const isCreateModalOpen = ref(false)
const processingProposalId = ref<string | null>(null)

const getWalletName = (walletId?: string) => {
  if (!walletId) return 'Saldo Utama'
  const w = wallets.value.find((item) => item.id === walletId)
  return w ? w.name : 'General'
}

const handleApprove = async (id: string) => {
  processingProposalId.value = id
  await approveProposal(id)
  processingProposalId.value = null
}

const handleReject = async (id: string) => {
  processingProposalId.value = id
  await rejectProposal(id)
  processingProposalId.value = null
}

const getStatusBadge = (status: ProposalStatus) => {
  switch (status) {
    case 'approved':
      return 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30'
    case 'rejected':
      return 'bg-rose-500/10 text-rose-400 border border-rose-500/30'
    case 'pending':
    default:
      return 'bg-amber-500/10 text-amber-400 border border-amber-500/30'
  }
}
</script>

<template>
  <div class="flex flex-col gap-6 pb-12">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
      <div>
        <h1 class="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
          {{ t('proposals.title') || 'Pengajuan Anggaran' }}
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          {{ isAdmin ? 'Tinjau dan setujui pengajuan pengeluaran dari anggota keluarga' : 'Kirim dan pantau status pengajuan pengeluaran Anda' }}
        </p>
      </div>

      <Button
        variant="primary"
        size="sm"
        @click="isCreateModalOpen = true"
      >
        + {{ t('extra.sendProposalBtn') || 'Ajukan Pengeluaran' }}
      </Button>
    </div>

    <!-- Status Tabs Filter -->
    <div class="card-neo bg-slate-900/90 p-2 rounded-2xl border border-slate-800 flex items-center gap-1.5 overflow-x-auto">
      <button
        v-for="status in ['all', 'pending', 'approved', 'rejected']"
        :key="status"
        @click="statusFilter = status as any"
        class="px-4 py-2 rounded-xl text-xs font-bold transition capitalize cursor-pointer whitespace-nowrap"
        :class="[
          statusFilter === status
            ? 'bg-slate-800 text-white shadow-sm ring-1 ring-slate-700'
            : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/40'
        ]"
      >
        {{ status }}
      </button>
    </div>

    <!-- Loading Skeleton -->
    <div v-if="isLoading" class="space-y-4">
      <Skeleton v-for="i in 3" :key="i" type="card" />
    </div>

    <!-- Empty State -->
    <div
      v-else-if="filteredProposals.length === 0"
      class="card-neo bg-slate-900/90 rounded-3xl p-16 border border-slate-800 text-center flex flex-col items-center justify-center gap-3"
    >
      <div class="w-12 h-12 rounded-2xl bg-slate-800 flex items-center justify-center text-xl text-teal-400">
        📝
      </div>
      <h3 class="text-base font-bold text-white">Tidak ada pengajuan</h3>
      <p class="text-xs text-slate-400 max-w-sm">
        Belum ada pengajuan pengeluaran dengan filter status yang dipilih.
      </p>
    </div>

    <!-- Proposals List -->
    <div v-else class="flex flex-col gap-4">
      <div
        v-for="proposal in filteredProposals"
        :key="proposal.id"
        class="card-neo bg-slate-900/90 rounded-3xl p-6 border border-slate-800/90 shadow-card flex flex-col gap-4"
      >
        <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2.5 flex-wrap mb-1.5">
              <h3 class="text-base font-bold text-white tracking-tight">
                {{ proposal.title }}
              </h3>
              <span
                class="px-2.5 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider"
                :class="getStatusBadge(proposal.status)"
              >
                {{ proposal.status }}
              </span>
              <span v-if="proposal.request_type" class="px-2 py-0.5 rounded-md bg-slate-800 text-slate-300 text-[10px] font-mono border border-slate-700">
                {{ proposal.request_type }}
              </span>
            </div>

            <p class="text-xs text-slate-300 mb-3 whitespace-pre-line">
              {{ proposal.description }}
            </p>

            <div class="flex items-center gap-3 text-xs text-slate-400">
              <span class="px-2.5 py-1 rounded-lg bg-slate-950 border border-slate-800 font-semibold text-teal-300">
                {{ getWalletName(proposal.wallet_id) }}
              </span>
              <span class="font-mono text-[11px]">
                {{ formatDate(proposal.created_at) }}
              </span>
            </div>
          </div>

          <div class="text-right sm:text-right flex sm:flex-col justify-between items-end">
            <span class="text-xl sm:text-2xl font-black text-white font-sans">
              {{ formatRp(proposal.amount) }}
            </span>
          </div>
        </div>

        <!-- Admin Actions for Pending Proposals -->
        <div
          v-if="isAdmin && proposal.status === 'pending'"
          class="pt-4 border-t border-slate-800 flex items-center justify-end gap-3"
        >
          <Button
            variant="danger"
            size="xs"
            :loading="isRejecting && processingProposalId === proposal.id"
            :disabled="isApproving || isRejecting"
            @click="handleReject(proposal.id)"
          >
            ✕ Tolak
          </Button>
          <Button
            variant="success"
            size="xs"
            :loading="isApproving && processingProposalId === proposal.id"
            :disabled="isApproving || isRejecting"
            @click="handleApprove(proposal.id)"
          >
            ✓ Setujui
          </Button>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <CreateProposalModal v-model:isOpen="isCreateModalOpen" />
  </div>
</template>
