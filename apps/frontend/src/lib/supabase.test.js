import test from 'node:test'
import assert from 'node:assert/strict'

import { hasSupabaseConfig, supabase } from './supabase.js'

test('supabase client remains safe when env vars are missing', () => {
  assert.equal(hasSupabaseConfig, false)
  assert.equal(supabase, null)
})
