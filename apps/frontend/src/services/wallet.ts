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
  family_id: string;
  name: string;
  description: string;
  initial_balance: number;
  minimum_limit: number;
}

export const getWallets = (familyId: string) =>
  apiClient.get<{ data: Wallet[] }>('/wallets', { params: { family_id: familyId } });

export const createWallet = (payload: CreateWalletPayload) =>
  apiClient.post('/wallets', payload);
