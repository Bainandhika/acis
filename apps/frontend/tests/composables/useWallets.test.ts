import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useWallets } from '../../src/composables/useWallets'
import { useWalletStore } from '../../src/stores/wallet'

describe('useWallets', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('initializes with empty wallets', () => {
    const { wallets, totalBalance } = useWallets()
    expect(wallets.value).toEqual([])
    expect(totalBalance.value).toBe(0)
  })

  it('calculates total balance accurately from store wallets', () => {
    const walletStore = useWalletStore()
    walletStore.wallets = [
      {
        id: 'w1',
        name: 'Groceries',
        description: '',
        initial_balance: 1000000,
        current_balance: 750000,
        minimum_limit: 100000
      },
      {
        id: 'w2',
        name: 'Transport',
        description: '',
        initial_balance: 500000,
        current_balance: 250000,
        minimum_limit: 50000
      }
    ]

    const { totalBalance } = useWallets()
    expect(totalBalance.value).toBe(1000000)
  })
})
