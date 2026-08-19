import { useAuthStore } from '../stores/useAuthStore'

export const API_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

export async function apiRequest(path, options = {}, retry = true) {
  const authStore = useAuthStore()
  const headers = new Headers(options.headers || {})
  headers.set('Content-Type', 'application/json')
  if (authStore.token) headers.set('Authorization', `Bearer ${authStore.token}`)

  const response = await fetch(`${API_URL}${path}`, { ...options, headers, credentials: 'include' })
  if (response.status === 401 && retry && path !== '/authentication/refresh') {
    const refreshed = await refreshToken()
    if (refreshed) return apiRequest(path, options, false)
    authStore.clearToken()
  }

  const body = await response.json().catch(() => ({}))
  if (!response.ok) throw new Error(body.error || 'Permintaan gagal diproses')
  return body
}

export async function refreshToken() {
  try {
    const body = await apiRequest('/authentication/refresh', { method: 'POST' }, false)
    if (body.token) useAuthStore().setToken(body.token)
    return Boolean(body.token)
  } catch {
    return false
  }
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
