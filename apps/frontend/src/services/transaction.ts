import apiClient from './api';

export type TransactionType = 'income' | 'expense' | 'allocation';

export interface Transaction {
  id: string;
  wallet_id: string;
  user_id?: string;
  type: TransactionType;
  amount: number;
  description?: string;
  created_at: string;
}

export interface CreateTransactionPayload {
  wallet_id?: string;
  type: TransactionType;
  amount: number;
  description?: string;
}

export interface UpdateTransactionPayload {
  wallet_id?: string;
  type: TransactionType;
  amount: number;
  description?: string;
}

export interface Proposal {
  id: string;
  wallet_id: string;
  proposed_by?: string;
  title: string;
  amount: number;
  description: string;
  status: 'pending' | 'approved' | 'rejected';
  request_type?: 'add_transaction' | 'edit_transaction' | 'delete_transaction';
  target_transaction_id?: string;
  payload?: any;
  reviewed_by?: string;
  reviewed_at?: string;
  created_at: string;
}

export interface CreateProposalPayload {
  wallet_id: string;
  title: string;
  amount: number;
  description: string;
  request_type: 'add_transaction' | 'edit_transaction' | 'delete_transaction';
  target_transaction_id?: string;
  payload?: any;
}

export const getTransactions = (year?: number, month?: number) => {
  const params: Record<string, number> = {};
  if (year && month) {
    params.year = year;
    params.month = month;
  }
  return apiClient.get<{ data: Transaction[] }>('/transaction', { params });
};

export const createTransaction = (payload: CreateTransactionPayload) =>
  apiClient.post<{ data: Transaction }>('/transaction', payload);

export const updateTransaction = (id: string, payload: UpdateTransactionPayload) =>
  apiClient.patch<{ data: Transaction }>(`/transaction/${id}`, payload);

export const getProposals = () =>
  apiClient.get<{ data: Proposal[] }>('/transaction/proposals');

export const createProposal = (payload: CreateProposalPayload) =>
  apiClient.post<{ data: Proposal }>('/transaction/proposals', payload);

export const approveProposal = (proposalId: string) =>
  apiClient.post(`/transaction/proposals/${proposalId}/approve`);

export const rejectProposal = (proposalId: string) =>
  apiClient.post(`/transaction/proposals/${proposalId}/reject`);

export const deleteTransaction = (transactionId: string) =>
  apiClient.delete(`/transaction/${transactionId}`);
