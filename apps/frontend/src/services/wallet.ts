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

export const getWallets = () =>
  apiClient.get<{ data: Wallet[] }>('/family/wallets');

export const createWallet = (payload: CreateWalletPayload) =>
  apiClient.post('/family/wallets', payload);

