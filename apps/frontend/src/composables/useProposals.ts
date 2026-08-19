import { computed, ref, onMounted, getCurrentInstance } from 'vue'
import { useTransactionStore } from '../stores/transaction'
import { useWalletStore } from '../stores/wallet'
import { useApi } from './useApi'
import { useUI } from './useUI'
import type { ProposalStatus } from '../types'
import * as txService from '../services/transaction'

export function useProposals() {
  const txStore = useTransactionStore()
  const walletStore = useWalletStore()
  const { showToast } = useUI()

  const proposals = computed(() => txStore.proposals || [])
  const statusFilter = ref<'all' | ProposalStatus>('all')

  const filteredProposals = computed(() => {
    if (statusFilter.value === 'all') {
      return proposals.value
    }
    return proposals.value.filter((p) => p.status === statusFilter.value)
  })

  const pendingProposals = computed(() =>
    proposals.value.filter((p) => p.status === 'pending')
  )

  const fetchProposalsApi = useApi(txService.getProposals, {
    onSuccess: (data) => {
      const list = Array.isArray(data) ? data : (data as any)?.data || []
      txStore.proposals = list
    }
  })

  const createProposalApi = useApi(txService.createProposal, {
    onSuccess: async () => {
      showToast('success', 'Proposal submitted successfully')
      await fetchProposalsApi.execute()
    }
  })

  const approveProposalApi = useApi(
    (id: string) => txService.approveProposal(id),
    {
      onSuccess: async () => {
        showToast('success', 'Proposal approved')
        await Promise.all([
          fetchProposalsApi.execute(),
          txStore.fetchTransactions(),
          walletStore.fetchWallets()
        ])
      }
    }
  )

  const rejectProposalApi = useApi(
    (id: string) => txService.rejectProposal(id),
    {
      onSuccess: async () => {
        showToast('info', 'Proposal rejected')
        await fetchProposalsApi.execute()
      }
    }
  )

  if (getCurrentInstance()) {
    onMounted(() => {
      if (proposals.value.length === 0) {
        fetchProposalsApi.execute()
      }
    })
  }

  return {
    proposals,
    filteredProposals,
    pendingProposals,
    statusFilter,
    isLoading: computed(() => fetchProposalsApi.isLoading.value || txStore.loading),
    isCreating: computed(() => createProposalApi.isLoading.value),
    isApproving: computed(() => approveProposalApi.isLoading.value),
    isRejecting: computed(() => rejectProposalApi.isLoading.value),
    fetchProposals: fetchProposalsApi.execute,
    createProposal: createProposalApi.execute,
    approveProposal: approveProposalApi.execute,
    rejectProposal: rejectProposalApi.execute
  }
}
