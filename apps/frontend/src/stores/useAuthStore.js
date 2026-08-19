import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('auth_token') || null,
    user: JSON.parse(localStorage.getItem('auth_user') || 'null')
  }),
  actions: {
    async requestOtp(phoneNumber) {
      const response = await fetch(`${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'}/authentication/request-otp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ phone_number: phoneNumber, action: 'login' })
      })

      if (!response.ok) {
        const body = await response.json().catch(() => ({}))
        throw new Error(body.error || 'Gagal meminta kode OTP')
      }

      return response.json()
    },
    async verifyOtp(phoneNumber, otp) {
      const response = await fetch(`${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'}/authentication/verify-otp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ phone_number: phoneNumber, otp })
      })

      if (!response.ok) {
        const body = await response.json().catch(() => ({}))
        throw new Error(body.error || 'Kode OTP tidak valid')
      }

      const data = await response.json()
      this.setToken(data.token)
      this.user = data.user
      localStorage.setItem('auth_user', JSON.stringify(data.user))
      return data
    },
    async refresh() {
      const response = await fetch(`${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'}/authentication/refresh`, { method: 'POST', credentials: 'include' })
      if (!response.ok) return false
      const data = await response.json()
      this.setToken(data.token)
      this.user = data.user
      localStorage.setItem('auth_user', JSON.stringify(data.user))
      return true
    },
    async logout() {
      await fetch(`${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'}/authentication/logout`, { method: 'POST', credentials: 'include', headers: this.token ? { Authorization: `Bearer ${this.token}` } : {} })
      this.clearToken()
    },
    setToken(token) {
      this.token = token
      if (token) {
        localStorage.setItem('auth_token', token)
      } else {
        localStorage.removeItem('auth_token')
      }
    },
    clearToken() {
      this.token = null
      this.user = null
      localStorage.removeItem('auth_user')
      localStorage.removeItem('auth_token')
    }
  }
})
