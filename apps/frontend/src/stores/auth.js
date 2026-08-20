import { defineStore } from 'pinia'
import { getSupabaseClient, supabase } from '../lib/supabase'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    session: null,
    user: null,
    profile: null,
    loading: true,
    initialized: false
  }),

  getters: {
    isAuthenticated: (state) => Boolean(state.session && state.session.access_token),
    token: (state) => state.session?.access_token || null,
    userID: (state) => state.session?.user?.id || null,
    userEmail: (state) => state.session?.user?.email || ''
  },

  actions: {
    async init() {
      if (this.initialized) return

      if (!supabase) {
        this.loading = false
        this.initialized = true
        return
      }

      try {
        this.loading = true
        const { data: { session } } = await supabase.auth.getSession()
        this.session = session
        this.user = session?.user || null

        if (session?.access_token) {
          await this.provisionProfile()
        }

        // Listen for auth state changes
        supabase.auth.onAuthStateChange(async (event, session) => {
          this.session = session
          this.user = session?.user || null

          if (event === 'SIGNED_IN' && session?.access_token) {
            await this.provisionProfile()
          } else if (event === 'SIGNED_OUT') {
            this.session = null
            this.user = null
            this.profile = null
          }
        })
      } catch (err) {
        console.error('Failed to initialize auth store:', err)
      } finally {
        this.loading = false
        this.initialized = true
      }
    },

    async signInWithGoogle() {
      const client = getSupabaseClient()
      const { data, error } = await client.auth.signInWithOAuth({
        provider: 'google',
        options: {
          redirectTo: window.location.origin
        }
      })
      if (error) {
        throw error
      }
      return data
    },

    async signOut() {
      try {
        if (supabase) {
          await supabase.auth.signOut()
        }
      } finally {
        this.session = null
        this.user = null
        this.profile = null
      }
    },

    async provisionProfile() {
      if (!this.session?.access_token) return null

      try {
        const metadata = this.user?.user_metadata || {}
        const name = metadata.full_name || metadata.name || this.user?.email?.split('@')[0] || 'User'
        const avatarUrl = metadata.avatar_url || metadata.picture || ''
        const username = metadata.user_name || metadata.preferred_username || ''

        const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'
        const response = await fetch(`${apiBaseUrl}/auth/provision`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.session.access_token}`
          },
          body: JSON.stringify({
            name,
            username,
            avatar_url: avatarUrl
          })
        })

        if (response.ok) {
          const res = await response.json()
          this.profile = res.data || res
          return this.profile
        } else {
          const errBody = await response.json().catch(() => ({}))
          console.error('Profile provisioning error response:', response.status, errBody)
        }
      } catch (err) {
        console.warn('Profile provisioning background sync failed:', err)
      }
      return null
    },

    async fetchMe() {
      if (!this.session?.access_token) return null

      try {
        const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'
        const response = await fetch(`${apiBaseUrl}/auth/me`, {
          headers: {
            'Authorization': `Bearer ${this.session.access_token}`
          }
        })
        if (response.ok) {
          const res = await response.json()
          this.profile = res.data || res
          return this.profile
        }
      } catch (err) {
        console.warn('Failed to fetch me profile:', err)
      }
      return null
    }
  }
})
