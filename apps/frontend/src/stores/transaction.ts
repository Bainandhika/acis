import { defineStore } from 'pinia';
import { ref } from 'vue';
import { 
  getTransactions, 
  createTransaction, 
  updateTransaction,
  deleteTransaction,
  getProposals, 
  createProposal, 
  approveProposal, 
  rejectProposal,
  type Transaction, 
  type Proposal, 
  type CreateTransactionPayload, 
  type UpdateTransactionPayload,
  type CreateProposalPayload 
} from '../services/transaction';

export const useTransactionStore = defineStore('transaction', () => {
  const transactions = ref<Transaction[]>([]);
  const proposals = ref<Proposal[]>([]);
  const selectedYear = ref<number>(new Date().getFullYear());
  const selectedMonth = ref<number>(new Date().getMonth() + 1);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchTransactions(year?: number, month?: number) {
    loading.value = true;
    error.value = null;
    const y = year !== undefined ? year : selectedYear.value;
    const m = month !== undefined ? month : selectedMonth.value;
    try {
      const response = await getTransactions(y, m);
      transactions.value = response.data.data || [];
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to fetch transactions';
    } finally {
      loading.value = false;
    }
  }

  function setPeriod(year: number, month: number) {
    selectedYear.value = year;
    selectedMonth.value = month;
    return fetchTransactions(year, month);
  }

  async function addTransaction(payload: CreateTransactionPayload) {
    loading.value = true;
    error.value = null;
    try {
      await createTransaction(payload);
      await fetchTransactions();
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to record transaction';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function editTransaction(id: string, payload: UpdateTransactionPayload) {
    loading.value = true;
    error.value = null;
    try {
      await updateTransaction(id, payload);
      await fetchTransactions();
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to update transaction';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function fetchProposals() {
    loading.value = true;
    error.value = null;
    try {
      const response = await getProposals();
      proposals.value = response.data.data || [];
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to fetch proposals';
    } finally {
      loading.value = false;
    }
  }

  async function addProposal(payload: CreateProposalPayload) {
    loading.value = true;
    error.value = null;
    try {
      await createProposal(payload);
      await fetchProposals();
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to submit proposal';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function handleApprove(id: string) {
    loading.value = true;
    error.value = null;
    try {
      await approveProposal(id);
      await fetchProposals();
      await fetchTransactions();
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to approve proposal';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function handleReject(id: string) {
    loading.value = true;
    error.value = null;
    try {
      await rejectProposal(id);
      await fetchProposals();
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to reject proposal';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function removeTransaction(id: string) {
    loading.value = true;
    error.value = null;
    try {
      await deleteTransaction(id);
      await fetchTransactions();
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to delete transaction';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  function resetState() {
    transactions.value = [];
    proposals.value = [];
    loading.value = false;
    error.value = null;
  }

  return {
    transactions,
    proposals,
    selectedYear,
    selectedMonth,
    loading,
    error,
    fetchTransactions,
    setPeriod,
    addTransaction,
    editTransaction,
    removeTransaction,
    fetchProposals,
    addProposal,
    handleApprove,
    handleReject,
    resetState
  };
});
