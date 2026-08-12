import { defineStore } from 'pinia';
import { ref } from 'vue';
import {
  getTransactions,
  createTransaction,
  getProposals,
  createProposal,
  approveProposal,
  rejectProposal,
  type Transaction,
  type Proposal,
  type CreateTransactionPayload,
  type CreateProposalPayload
} from '../services/transaction';

export const useTransactionStore = defineStore('transaction', () => {
  const transactions = ref<Transaction[]>([]);
  const proposals = ref<Proposal[]>([]);
  const loading = ref(false);

  async function fetchTransactions() {
    loading.value = true;
    try {
      const { data } = await getTransactions();
      transactions.value = data.data || [];
    } catch (error) {
      console.error('Failed to fetch transactions', error);
    } finally {
      loading.value = false;
    }
  }

  async function addTransaction(payload: CreateTransactionPayload) {
    await createTransaction(payload);
    await fetchTransactions();
  }

  async function fetchProposals() {
    loading.value = true;
    try {
      const { data } = await getProposals();
      proposals.value = data.data || [];
    } catch (error) {
      console.error('Failed to fetch proposals', error);
    } finally {
      loading.value = false;
    }
  }

  async function addProposal(payload: CreateProposalPayload) {
    await createProposal(payload);
    await fetchProposals();
  }

  async function handleApprove(id: string) {
    await approveProposal(id);
    await fetchProposals();
    await fetchTransactions();
  }

  async function handleReject(id: string) {
    await rejectProposal(id);
    await fetchProposals();
  }

  return {
    transactions,
    proposals,
    loading,
    fetchTransactions,
    addTransaction,
    fetchProposals,
    addProposal,
    handleApprove,
    handleReject
  };
});
