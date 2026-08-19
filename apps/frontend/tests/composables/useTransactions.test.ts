import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useTransactions } from '../../src/composables/useTransactions'
import { useTransactionStore } from '../../src/stores/transaction'

describe('useTransactions', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('filters transactions by search query', () => {
    const txStore = useTransactionStore()
    txStore.transactions = [
      {
        id: 't1',
        wallet_id: 'w1',
        type: 'expense',
        amount: 50000,
        description: 'Makan Siang',
        category: 'Food',
        created_at: '2026-08-19T10:00:00Z'
      },
      {
        id: 't2',
        wallet_id: 'w2',
        type: 'expense',
        amount: 20000,
        description: 'Bensin Motor',
        category: 'Transport',
        created_at: '2026-08-18T10:00:00Z'
      }
    ]

    const { filteredTransactions, setFilters, clearFilters } = useTransactions()

    expect(filteredTransactions.value.length).toBe(2)

    setFilters({ search: 'bensin' })
    expect(filteredTransactions.value.length).toBe(1)
    expect(filteredTransactions.value[0].id).toBe('t2')

    clearFilters()
    expect(filteredTransactions.value.length).toBe(2)
  })

  it('filters transactions by type', () => {
    const txStore = useTransactionStore()
    txStore.transactions = [
      {
        id: 't1',
        wallet_id: 'w1',
        type: 'expense',
        amount: 50000,
        description: 'Pengeluaran',
        created_at: '2026-08-19T10:00:00Z'
      },
      {
        id: 't2',
        wallet_id: 'w1',
        type: 'income',
        amount: 500000,
        description: 'Gaji',
        created_at: '2026-08-18T10:00:00Z'
      }
    ]

    const { filteredTransactions, setFilters } = useTransactions()
    setFilters({ type: 'income' })
    expect(filteredTransactions.value.length).toBe(1)
    expect(filteredTransactions.value[0].type).toBe('income')
  })
})
