import apiClient from './api';

export interface FamilyMember {
  id: string;
  user_id: string;
  user_name?: string;
  role: 'admin' | 'member';
  joined_at: string;
}

export interface Family {
  id: string;
  name: string;
  invite_code: string;
  telegram_chat_id?: number;
  monthly_income: number;
  created_by?: string;
  members?: FamilyMember[];
  created_at: string;
}

export const getMyFamily = () =>
  apiClient.get<{ data: Family }>('/family/me');

export const createFamily = (name: string, monthly_income: number = 0) =>
  apiClient.post<{ data: Family }>('/family', { name, monthly_income });

export const joinFamily = (invite_code: string) =>
  apiClient.post<{ data: Family }>('/family/join', { invite_code });

export const updateFamilySettings = (monthly_income: number) =>
  apiClient.patch('/family/settings', { monthly_income });

export const disconnectTelegram = () =>
  apiClient.post('/family/telegram/disconnect');
