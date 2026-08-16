import apiClient from './api';

export interface Wallet {
  id: string;
  name: string;
  description: string;
  initial_balance: number;
  current_balance: number;
  minimum_limit: number;
}

export interface CreateWalletPayload {
  name: string;
  description: string;
  initial_balance: number;
  minimum_limit: number;
}

export interface UpdateWalletPayload {
  name: string;
  description?: string;
  minimum_limit: number;
}

export const getWallets = () =>
  apiClient.get<{ data: Wallet[] }>('/family/wallets');

export const createWallet = (payload: CreateWalletPayload) =>
  apiClient.post<{ message: string; data: Wallet }>('/family/wallets', payload);

export const updateWallet = (id: string, payload: UpdateWalletPayload) =>
  apiClient.patch<{ message: string; data: Wallet }>(`/family/wallets/${id}`, payload);

export const deleteWallet = (id: string) =>
  apiClient.delete(`/family/wallets/${id}`);


