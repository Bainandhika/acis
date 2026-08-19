import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useProposals } from '../../src/composables/useProposals'
import { useTransactionStore } from '../../src/stores/transaction'

describe('useProposals', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('filters proposals by status correctly', () => {
    const txStore = useTransactionStore()
    txStore.proposals = [
      {
        id: 'p1',
        wallet_id: 'w1',
        title: 'Pengajuan 1',
        amount: 100000,
        description: 'Beli buku',
        status: 'pending',
        created_at: '2026-08-19T00:00:00Z'
      },
      {
        id: 'p2',
        wallet_id: 'w1',
        title: 'Pengajuan 2',
        amount: 200000,
        description: 'Beli obat',
        status: 'approved',
        created_at: '2026-08-18T00:00:00Z'
      }
    ]

    const { filteredProposals, pendingProposals, statusFilter } = useProposals()

    expect(filteredProposals.value.length).toBe(2)
    expect(pendingProposals.value.length).toBe(1)

    statusFilter.value = 'approved'
    expect(filteredProposals.value.length).toBe(1)
    expect(filteredProposals.value[0].id).toBe('p2')
  })
})
