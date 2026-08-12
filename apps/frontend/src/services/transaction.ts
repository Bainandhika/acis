import apiClient from './api';

export interface Transaction {
  id: string;
  wallet_id: string;
  user_id?: string;
  type: 'income' | 'expense';
  amount: number;
  category: string;
  description?: string;
  created_at: string;
}

export interface CreateTransactionPayload {
  wallet_id: string;
  type: 'income' | 'expense';
  amount: number;
  category: string;
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
  reviewed_by?: string;
  reviewed_at?: string;
  created_at: string;
}

export interface CreateProposalPayload {
  wallet_id: string;
  title: string;
  amount: number;
  description: string;
}

export const getTransactions = () =>
  apiClient.get<{ data: Transaction[] }>('/transaction');

export const createTransaction = (payload: CreateTransactionPayload) =>
  apiClient.post<{ data: Transaction }>('/transaction', payload);

export const getProposals = () =>
  apiClient.get<{ data: Proposal[] }>('/transaction/proposals');

export const createProposal = (payload: CreateProposalPayload) =>
  apiClient.post<{ data: Proposal }>('/transaction/proposals', payload);

export const approveProposal = (proposalId: string) =>
  apiClient.post(`/transaction/proposals/${proposalId}/approve`);

export const rejectProposal = (proposalId: string) =>
  apiClient.post(`/transaction/proposals/${proposalId}/reject`);
