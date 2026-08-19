<script setup lang="ts">
import { ref, watch } from 'vue'
import Modal from '../../../../components/ui/Modal.vue'
import Button from '../../../../components/ui/Button.vue'
import Input from '../../../../components/ui/Input.vue'
import { useProposals } from '../../../../composables/useProposals'
import { useWallets } from '../../../../composables/useWallets'
import { validateForm, CreateProposalSchema } from '../../../../utils/validate'
import { formatRp } from '../../../../utils/format'
import { useI18n } from '../../../../locales'
import type { Transaction } from '../../../../services/transaction'

const props = defineProps<{
  isOpen: boolean
  transaction: Transaction | null
}>()

const emit = defineEmits<{
  (e: 'update:isOpen', value: boolean): void
}>()

const { createProposal, isCreating } = useProposals()
const { wallets } = useWallets()
const { t } = useI18n()

const form = ref<{
  wallet_id: string
  type: 'income' | 'expense' | 'allocation'
  amount: number
  description: string
}>({
  wallet_id: '',
  type: 'expense',
  amount: 0,
  description: ''
})

const formErrors = ref<Record<string, string>>({})

watch(
  () => props.transaction,
  (tx) => {
    if (tx) {
      form.value = {
        wallet_id: tx.wallet_id || '',
        type: tx.type === 'income' ? 'income' : 'expense',
        amount: tx.amount || 0,
        description: ''
      }
      formErrors.value = {}
    }
  },
  { immediate: true }
)

const handleClose = () => {
  emit('update:isOpen', false)
}

const handleSubmit = async () => {
  if (!props.transaction) return

  const payload = {
    wallet_id: form.value.wallet_id || props.transaction.wallet_id,
    title: `Revisi Transaksi: ${form.value.description || props.transaction.description || formatRp(form.value.amount)}`,
    amount: form.value.amount,
    description: form.value.description || 'Permintaan revisi data transaksi',
    request_type: 'edit_transaction' as const,
    target_transaction_id: props.transaction.id,
    payload: {
      wallet_id: form.value.wallet_id,
      type: form.value.type,
      amount: form.value.amount,
      description: form.value.description
    }
  }

  const validation = validateForm(CreateProposalSchema, payload)
  if (!validation.success) {
    formErrors.value = validation.errors
    return
  }

  formErrors.value = {}
  const result = await createProposal(validation.data)
  if (result !== null) {
    handleClose()
  }
}
</script>

<template>
  <Modal
    :is-open="isOpen"
    :title="t('extra.requestChangeTitle') || 'Ajukan Perubahan Transaksi'"
    :description="t('extra.requestChangeDesc') || 'Ajukan penyesuaian nominal atau detail transaksi ke Admin'"
    @close="handleClose"
  >
    <form @submit.prevent="handleSubmit" class="flex flex-col gap-4">
      <div>
        <label class="text-xs font-bold text-slate-300 block mb-1.5">
          {{ t('extra.selectWalletLabel') || 'Dompet' }}
        </label>
        <select
          v-model="form.wallet_id"
          class="w-full px-4 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-xs font-semibold text-slate-100 focus:outline-none focus:ring-2 focus:ring-teal-500/30 focus:border-teal-600"
        >
          <option v-for="w in wallets" :key="w.id" :value="w.id">
            {{ w.name }}
          </option>
        </select>
      </div>

      <Input
        v-model="form.amount"
        type="number"
        :label="t('extra.amountLabel') || 'Nominal Baru (Rp)'"
        :error="formErrors.amount"
        :min="1"
        required
      />

      <Input
        v-model="form.description"
        :label="t('extra.reasonLabel') || 'Alasan Perubahan'"
        :placeholder="t('extra.reasonPlaceholder') || 'Contoh: Koreksi nominal struk belanja'"
        :error="formErrors.description"
        required
      />

      <div class="mt-4 pt-4 border-t border-slate-800 flex items-center justify-end gap-3">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          @click="handleClose"
          :disabled="isCreating"
        >
          {{ t('extra.cancel') || 'Batal' }}
        </Button>
        <Button
          type="submit"
          variant="primary"
          size="sm"
          :loading="isCreating"
        >
          {{ isCreating ? (t('extra.processing') || 'Mengirim...') : (t('extra.sendProposalBtn') || 'Kirim Pengajuan') }}
        </Button>
      </div>
    </form>
  </Modal>
</template>
