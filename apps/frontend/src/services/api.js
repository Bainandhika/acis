import { useAuthStore } from '../stores/auth'
import { supabase } from '../lib/supabase'

export const API_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

export async function apiRequest(path, options = {}) {
  const authStore = useAuthStore()
  const headers = new Headers(options.headers || {})
  headers.set('Content-Type', 'application/json')

  // Retrieve current fresh access token from Supabase session
  let token = authStore.token
  if (!token) {
    const { data: { session } } = await supabase.auth.getSession()
    if (session?.access_token) {
      token = session.access_token
      authStore.session = session
      authStore.user = session.user
    }
  }

  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers
  })

  if (response.status === 401) {
    // If unauthorized, attempt to refresh session with Supabase
    const { data: { session } } = await supabase.auth.refreshSession()
    if (session?.access_token) {
      authStore.session = session
      authStore.user = session.user
      headers.set('Authorization', `Bearer ${session.access_token}`)
      const retryResponse = await fetch(`${API_URL}${path}`, {
        ...options,
        headers
      })
      const retryBody = await retryResponse.json().catch(() => ({}))
      if (!retryResponse.ok) {
        throw new Error(retryBody.error || 'Permintaan gagal diproses')
      }
      return retryBody
    }

    // Refresh failed or token completely invalid -> sign out & redirect to login
    await authStore.signOut()
    if (typeof window !== 'undefined' && window.location.pathname !== '/masuk') {
      window.location.href = '/masuk'
    }
    throw new Error('Sesi telah berakhir, silakan login kembali')
  }

  const body = await response.json().catch(() => ({}))
  if (!response.ok) {
    throw new Error(body.error || 'Permintaan gagal diproses')
  }
  return body
}

export const getFamily = () => apiRequest('/family/me')
export const createFamily = (payload) => apiRequest('/family', { method: 'POST', body: JSON.stringify(payload) })
export const joinFamily = (payload) => apiRequest('/family/join', { method: 'POST', body: JSON.stringify(payload) })
export const updateFamily = (payload) => apiRequest('/family', { method: 'PATCH', body: JSON.stringify(payload) })
export const updateFamilySettings = (payload) => apiRequest('/family/settings', { method: 'PATCH', body: JSON.stringify(payload) })
export const disconnectTelegram = () => apiRequest('/family/telegram/disconnect', { method: 'POST' })
export const getWallets = () => apiRequest('/family/wallets')
export const createWallet = (payload) => apiRequest('/family/wallets', { method: 'POST', body: JSON.stringify(payload) })
export const updateWallet = (id, payload) => apiRequest(`/family/wallets/${id}`, { method: 'PATCH', body: JSON.stringify(payload) })
export const deleteWallet = (id) => apiRequest(`/family/wallets/${id}`, { method: 'DELETE' })
export const removeMember = (id) => apiRequest(`/family/members/${id}`, { method: 'DELETE' })
export const getTransactions = (year, month) => apiRequest(`/transaction?year=${year}&month=${month}`)
export const createTransaction = (payload) => apiRequest('/transaction', { method: 'POST', body: JSON.stringify(payload) })
export const updateTransaction = (id, payload) => apiRequest(`/transaction/${id}`, { method: 'PATCH', body: JSON.stringify(payload) })
export const deleteTransaction = (id) => apiRequest(`/transaction/${id}`, { method: 'DELETE' })
export const getProposals = () => apiRequest('/transaction/proposals')
export const createProposal = (payload) => apiRequest('/transaction/proposals', { method: 'POST', body: JSON.stringify(payload) })
export const approveProposal = (id) => apiRequest(`/transaction/proposals/${id}/approve`, { method: 'POST' })
export const rejectProposal = (id) => apiRequest(`/transaction/proposals/${id}/reject`, { method: 'POST' })
export const requestTelegramLinkCode = () => apiRequest('/telegram/link-code', { method: 'POST' })
