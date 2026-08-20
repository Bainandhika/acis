import { createClient } from '@supabase/supabase-js'

const env = typeof import.meta !== 'undefined' && import.meta.env ? import.meta.env : {}
const supabaseUrl = env.VITE_SUPABASE_URL?.trim() || ''
const supabaseAnonKey = env.VITE_SUPABASE_ANON_KEY?.trim() || ''

export const hasSupabaseConfig = Boolean(supabaseUrl && supabaseAnonKey)

export const supabase = hasSupabaseConfig ? createClient(supabaseUrl, supabaseAnonKey) : null

export function getSupabaseClient() {
  if (!hasSupabaseConfig || !supabase) {
    throw new Error('Supabase is not configured. Create apps/frontend/.env from .env.example and set VITE_SUPABASE_URL and VITE_SUPABASE_ANON_KEY.')
  }

  return supabase
}
