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
  created_by?: string;
  members?: FamilyMember[];
  created_at: string;
}

export const getMyFamily = () =>
  apiClient.get<{ data: Family }>('/family/me');

export const createFamily = (name: string) =>
  apiClient.post<{ data: Family }>('/family', { name });

export const joinFamily = (invite_code: string) =>
  apiClient.post<{ data: Family }>('/family/join', { invite_code });
