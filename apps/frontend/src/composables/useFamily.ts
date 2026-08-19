import { computed, onMounted, getCurrentInstance } from 'vue'
import { useFamilyStore } from '../stores/family'
import { useWalletStore } from '../stores/wallet'
import { useTransactionStore } from '../stores/transaction'
import { useApi } from './useApi'
import { useUI } from './useUI'
import * as familyService from '../services/family'
import * as txService from '../services/transaction'

export function useFamily() {
  const familyStore = useFamilyStore()
  const walletStore = useWalletStore()
  const txStore = useTransactionStore()
  const { showToast } = useUI()

  const family = computed(() => familyStore.family)
  const members = computed(() => familyStore.family?.members || [])
  const inviteCode = computed(() => familyStore.family?.invite_code || '')
  const monthlyIncome = computed(() => familyStore.family?.monthly_income || 0)
  const primaryBalance = computed(() => familyStore.family?.primary_balance || 0)

  const fetchFamilyApi = useApi(familyService.getMyFamily, {
    onSuccess: (data) => {
      const f = (data as any)?.data || data
      familyStore.family = f
    }
  })

  const updateFamilyNameApi = useApi(
    (name: string) => familyService.updateFamilyName(name),
    {
      onSuccess: async (_data, name: string) => {
        if (familyStore.family) {
          familyStore.family.name = name
        }
        showToast('success', 'Family name updated successfully')
      }
    }
  )

  const updateMonthlyIncomeApi = useApi(
    (income: number) => familyService.updateFamilySettings(income),
    {
      onSuccess: async (_data, income: number) => {
        if (familyStore.family) {
          familyStore.family.monthly_income = income
        }
        showToast('success', 'Monthly income updated successfully')
      }
    }
  )

  const disconnectTelegramApi = useApi(familyService.disconnectTelegram, {
    onSuccess: () => {
      if (familyStore.family) {
        familyStore.family.telegram_chat_id = undefined
      }
      showToast('info', 'Telegram bot disconnected')
    }
  })

  const removeMemberApi = useApi(
    (memberId: string) => familyService.removeMember(memberId),
    {
      onSuccess: async (_data, memberId: string) => {
        if (familyStore.family && familyStore.family.members) {
          familyStore.family.members = familyStore.family.members.filter((m) => m.id !== memberId)
        }
        showToast('success', 'Family member removed')
      }
    }
  )

  const allocateFundsApi = useApi(
    (payload: { wallet_id: string; amount: number; description?: string }) =>
      txService.createTransaction({
        wallet_id: payload.wallet_id,
        type: 'allocation',
        amount: payload.amount,
        description: payload.description || 'Alokasi Dana Utama ke Dompet'
      }),
    {
      onSuccess: async () => {
        showToast('success', 'Funds allocated successfully')
        await Promise.all([
          fetchFamilyApi.execute(),
          walletStore.fetchWallets(),
          txStore.fetchTransactions()
        ])
      }
    }
  )

  const copyInviteCode = async () => {
    if (inviteCode.value && typeof navigator !== 'undefined') {
      await navigator.clipboard.writeText(inviteCode.value)
      showToast('info', 'Invite code copied to clipboard')
    }
  }

  if (getCurrentInstance()) {
    onMounted(() => {
      if (!family.value) {
        fetchFamilyApi.execute()
      }
    })
  }

  return {
    family,
    members,
    inviteCode,
    monthlyIncome,
    primaryBalance,
    isLoading: computed(() => fetchFamilyApi.isLoading.value || familyStore.loading),
    isUpdatingName: computed(() => updateFamilyNameApi.isLoading.value),
    isUpdatingIncome: computed(() => updateMonthlyIncomeApi.isLoading.value),
    isDisconnectingTelegram: computed(() => disconnectTelegramApi.isLoading.value),
    isRemovingMember: computed(() => removeMemberApi.isLoading.value),
    isAllocating: computed(() => allocateFundsApi.isLoading.value),
    fetchFamily: fetchFamilyApi.execute,
    updateFamilyName: updateFamilyNameApi.execute,
    updateMonthlyIncome: updateMonthlyIncomeApi.execute,
    disconnectTelegram: disconnectTelegramApi.execute,
    removeMember: removeMemberApi.execute,
    allocateFunds: allocateFundsApi.execute,
    copyInviteCode
  }
}
