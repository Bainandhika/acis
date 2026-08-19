import { describe, it, expect } from 'vitest'
import {
  validateForm,
  CreateWalletSchema,
  CreateTransactionSchema,
  CreateProposalSchema,
  AllocateFundsSchema
} from '../../src/utils/validate'

describe('validate utilities', () => {
  describe('CreateWalletSchema', () => {
    it('accepts valid wallet payload', () => {
      const payload = {
        name: 'Dompet Belanja',
        description: 'Untuk belanja pasar',
        initial_balance: 500000,
        minimum_limit: 100000
      }
      const result = validateForm(CreateWalletSchema, payload)
      expect(result.success).toBe(true)
    })

    it('rejects wallet where minimum_limit > initial_balance', () => {
      const payload = {
        name: 'Dompet Belanja',
        description: '',
        initial_balance: 100000,
        minimum_limit: 500000
      }
      const result = validateForm(CreateWalletSchema, payload)
      expect(result.success).toBe(false)
      if (!result.success) {
        expect(result.errors.minimum_limit).toBeDefined()
      }
    })

    it('rejects wallet with too short name', () => {
      const payload = {
        name: 'A',
        description: '',
        initial_balance: 100000,
        minimum_limit: 50000
      }
      const result = validateForm(CreateWalletSchema, payload)
      expect(result.success).toBe(false)
      if (!result.success) {
        expect(result.errors.name).toBeDefined()
      }
    })
  })

  describe('CreateTransactionSchema', () => {
    it('validates a valid transaction', () => {
      const payload = {
        wallet_id: 'wallet-123',
        type: 'expense',
        amount: 25000,
        description: 'Beli kopi'
      }
      const result = validateForm(CreateTransactionSchema, payload)
      expect(result.success).toBe(true)
    })

    it('rejects negative or 0 amount', () => {
      const payload = {
        wallet_id: 'wallet-123',
        type: 'expense',
        amount: 0,
        description: ''
      }
      const result = validateForm(CreateTransactionSchema, payload)
      expect(result.success).toBe(false)
      if (!result.success) {
        expect(result.errors.amount).toBeDefined()
      }
    })
  })

  describe('CreateProposalSchema', () => {
    it('validates a valid proposal', () => {
      const payload = {
        wallet_id: 'wallet-123',
        title: 'Beli Buku',
        amount: 120000,
        description: 'Buku pelajaran sekolah anak',
        request_type: 'add_transaction'
      }
      const result = validateForm(CreateProposalSchema, payload)
      expect(result.success).toBe(true)
    })

    it('rejects empty target wallet', () => {
      const payload = {
        wallet_id: '',
        title: 'Beli Buku',
        amount: 120000,
        description: 'Buku pelajaran sekolah anak',
        request_type: 'add_transaction'
      }
      const result = validateForm(CreateProposalSchema, payload)
      expect(result.success).toBe(false)
      if (!result.success) {
        expect(result.errors.wallet_id).toBeDefined()
      }
    })
  })

  describe('AllocateFundsSchema', () => {
    it('validates valid fund allocation', () => {
      const payload = {
        wallet_id: 'wallet-123',
        amount: 300000,
        description: 'Alokasi bulanan'
      }
      const result = validateForm(AllocateFundsSchema, payload)
      expect(result.success).toBe(true)
    })
  })
})
