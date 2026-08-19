export type Role = 'admin' | 'member'

export interface User {
  id: string
  username: string
  role?: Role
}

export interface FamilyMember {
  id: string
  user_id: string
  user_name?: string
  role: Role
  joined_at: string
}

export interface Family {
  id: string
  name: string
  invite_code: string
  telegram_chat_id?: number
  monthly_income: number
  primary_balance: number
  created_by?: string
  members?: FamilyMember[]
  created_at: string
}

export interface Wallet {
  id: string
  short_id?: string
  name: string
  description: string
  initial_balance: number
  current_balance: number
  minimum_limit: number
}

export type TransactionType = 'income' | 'expense' | 'allocation'

export interface Transaction {
  id: string
  wallet_id: string
  user_id?: string
  type: TransactionType
  amount: number
  description?: string
  category?: string
  created_at: string
}

export type ProposalStatus = 'pending' | 'approved' | 'rejected'
export type ProposalRequestType = 'add_transaction' | 'edit_transaction' | 'delete_transaction'

export interface Proposal {
  id: string
  wallet_id: string
  proposed_by?: string
  title: string
  amount: number
  description: string
  status: ProposalStatus
  request_type?: ProposalRequestType
  target_transaction_id?: string
  payload?: any
  reviewed_by?: string
  reviewed_at?: string
  created_at: string
}

export interface TransactionFilters {
  wallet_id: string
  type: 'income' | 'expense' | 'allocation' | ''
  dateFrom: string
  dateTo: string
  search: string
}

export type ToastType = 'success' | 'error' | 'info' | 'warning'

export interface Toast {
  id: string
  type: ToastType
  message: string
  duration: number
}
