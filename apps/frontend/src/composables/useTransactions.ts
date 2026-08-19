import { computed, onMounted, getCurrentInstance, ref } from 'vue'
import { useTransactionStore } from '../stores/transaction'
import { useWalletStore } from '../stores/wallet'
import { useApi } from './useApi'
import { useUI } from './useUI'
import type { UpdateTransactionPayload } from '../services/transaction'
import type { TransactionFilters } from '../types'
import * as txService from '../services/transaction'

export function useTransactions() {
  const txStore = useTransactionStore()
  const walletStore = useWalletStore()
  const { showToast } = useUI()

  const transactions = computed(() => txStore.transactions)

  const filters = ref<TransactionFilters>({
    wallet_id: '',
    type: '',
    dateFrom: '',
    dateTo: '',
    search: ''
  })

  const filteredTransactions = computed(() => {
    let result = [...transactions.value]

    if (filters.value.wallet_id) {
      result = result.filter((t) => t.wallet_id === filters.value.wallet_id)
    }
    if (filters.value.type) {
      result = result.filter((t) => t.type === filters.value.type)
    }
    if (filters.value.dateFrom) {
      const from = new Date(filters.value.dateFrom)
      result = result.filter((t) => new Date(t.created_at) >= from)
    }
    if (filters.value.dateTo) {
      const to = new Date(filters.value.dateTo)
      to.setHours(23, 59, 59, 999)
      result = result.filter((t) => new Date(t.created_at) <= to)
    }
    if (filters.value.search) {
      const search = filters.value.search.toLowerCase()
      result = result.filter((t) =>
        (t.description || '').toLowerCase().includes(search) ||
        (t.category || '').toLowerCase().includes(search)
      )
    }

    return result.sort(
      (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    )
  })

  const fetchTransactionsApi = useApi(
    (year?: number, month?: number) => txService.getTransactions(year, month),
    {
      onSuccess: (data) => {
        const list = Array.isArray(data) ? data : (data as any)?.data || []
        txStore.transactions = list
      }
    }
  )

  const addTransactionApi = useApi(txService.createTransaction, {
    onSuccess: async () => {
      showToast('success', 'Transaction recorded successfully')
      await fetchTransactionsApi.execute(txStore.selectedYear, txStore.selectedMonth)
      await walletStore.fetchWallets()
    }
  })

  const editTransactionApi = useApi(
    (id: string, payload: UpdateTransactionPayload) => txService.updateTransaction(id, payload),
    {
      onSuccess: async () => {
        showToast('success', 'Transaction updated successfully')
        await fetchTransactionsApi.execute(txStore.selectedYear, txStore.selectedMonth)
        await walletStore.fetchWallets()
      }
    }
  )

  const removeTransactionApi = useApi(
    (id: string) => txService.deleteTransaction(id),
    {
      onSuccess: async () => {
        showToast('info', 'Transaction deleted successfully')
        await fetchTransactionsApi.execute(txStore.selectedYear, txStore.selectedMonth)
        await walletStore.fetchWallets()
      }
    }
  )

  const setPeriod = async (year: number, month: number) => {
    txStore.selectedYear = year
    txStore.selectedMonth = month
    await fetchTransactionsApi.execute(year, month)
  }

  const setFilters = (newFilters: Partial<TransactionFilters>) => {
    filters.value = { ...filters.value, ...newFilters }
  }

  const clearFilters = () => {
    filters.value = {
      wallet_id: '',
      type: '',
      dateFrom: '',
      dateTo: '',
      search: ''
    }
  }

  const exportCsv = (walletNameResolver?: (id?: string) => string) => {
    const headers = ['Date', 'Wallet', 'Type', 'Amount', 'Description']
    const rows = filteredTransactions.value.map((t) => [
      new Date(t.created_at).toLocaleDateString('id-ID'),
      walletNameResolver ? walletNameResolver(t.wallet_id) : t.wallet_id || 'General',
      t.type,
      t.amount,
      `"${(t.description || '').replace(/"/g, '""')}"`
    ])
    const csvContent = [headers.join(','), ...rows.map((r) => r.join(','))].join('\n')
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `acis-transactions-${txStore.selectedYear}-${txStore.selectedMonth}.csv`
    link.click()
    URL.revokeObjectURL(url)
    showToast('success', 'CSV exported successfully')
  }

  if (getCurrentInstance()) {
    onMounted(() => {
      if (transactions.value.length === 0) {
        fetchTransactionsApi.execute(txStore.selectedYear, txStore.selectedMonth)
      }
    })
  }

  return {
    transactions,
    filteredTransactions,
    filters,
    selectedYear: computed(() => txStore.selectedYear),
    selectedMonth: computed(() => txStore.selectedMonth),
    isLoading: computed(() => fetchTransactionsApi.isLoading.value || txStore.loading),
    isAdding: computed(() => addTransactionApi.isLoading.value),
    isEditing: computed(() => editTransactionApi.isLoading.value),
    isRemoving: computed(() => removeTransactionApi.isLoading.value),
    error: computed(() => fetchTransactionsApi.error.value || txStore.error),
    fetchTransactions: fetchTransactionsApi.execute,
    addTransaction: addTransactionApi.execute,
    editTransaction: editTransactionApi.execute,
    removeTransaction: removeTransactionApi.execute,
    setPeriod,
    setFilters,
    clearFilters,
    exportCsv
  }
}
