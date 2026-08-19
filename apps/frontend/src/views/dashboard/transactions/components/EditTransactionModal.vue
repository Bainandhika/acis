<script setup lang="ts">
import { ref, watch } from 'vue'
import Modal from '../../../../components/ui/Modal.vue'
import Button from '../../../../components/ui/Button.vue'
import Input from '../../../../components/ui/Input.vue'
import { useTransactions } from '../../../../composables/useTransactions'
import { useWallets } from '../../../../composables/useWallets'
import { validateForm, UpdateTransactionSchema } from '../../../../utils/validate'
import { formatRp } from '../../../../utils/format'
import { useI18n } from '../../../../locales'
import type { Transaction, UpdateTransactionPayload } from '../../../../services/transaction'

const props = defineProps<{
  isOpen: boolean
  transaction: Transaction | null
}>()

const emit = defineEmits<{
  (e: 'update:isOpen', value: boolean): void
}>()

const { editTransaction, isEditing } = useTransactions()
const { wallets } = useWallets()
const { t } = useI18n()

const form = ref<UpdateTransactionPayload>({
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
        type: tx.type,
        amount: tx.amount,
        description: tx.description || ''
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
  const validation = validateForm(UpdateTransactionSchema, form.value)
  if (!validation.success) {
    formErrors.value = validation.errors
    return
  }

  formErrors.value = {}
  const result = await editTransaction(props.transaction.id, validation.data)
  if (result !== null) {
    handleClose()
  }
}
</script>

<template>
  <Modal
    :is-open="isOpen"
    :title="t('extra.editTxTitle') || 'Edit Transaksi'"
    :description="t('extra.editTxDesc') || 'Perbarui detail transaksi yang telah tercatat'"
    @close="handleClose"
  >
    <form @submit.prevent="handleSubmit" class="flex flex-col gap-4">
      <div>
        <label class="text-xs font-bold text-slate-300 block mb-1.5">
          {{ t('extra.selectWalletLabel') || 'Pilih Dompet' }}
          <span class="text-rose-500">*</span>
        </label>
        <select
          v-model="form.wallet_id"
          class="w-full px-4 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-xs font-semibold text-slate-100 focus:outline-none focus:ring-2 focus:ring-teal-500/30 focus:border-teal-600"
        >
          <option value="">Saldo Utama</option>
          <option v-for="w in wallets" :key="w.id" :value="w.id">
            {{ w.name }} (Saldo: {{ formatRp(w.current_balance) }})
          </option>
        </select>
        <p v-if="formErrors.wallet_id" class="text-[11px] font-semibold text-rose-500 mt-1">
          {{ formErrors.wallet_id }}
        </p>
      </div>

      <div>
        <label class="text-xs font-bold text-slate-300 block mb-1.5">
          {{ t('extra.txTypeLabel') || 'Tipe Transaksi' }}
        </label>
        <select
          v-model="form.type"
          class="w-full px-4 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-xs font-semibold text-slate-100 focus:outline-none focus:ring-2 focus:ring-teal-500/30 focus:border-teal-600"
        >
          <option value="expense">{{ t('extra.expenseOption') || 'Pengeluaran' }}</option>
          <option value="income">{{ t('extra.incomeOption') || 'Pemasukan' }}</option>
        </select>
      </div>

      <Input
        v-model="form.amount"
        type="number"
        :label="t('extra.amountLabel') || 'Nominal (Rp)'"
        :error="formErrors.amount"
        :min="1"
        required
      />

      <Input
        v-model="form.description!"
        :label="t('extra.notesLabel') || 'Keterangan / Catatan'"
        :error="formErrors.description"
      />

      <div class="mt-4 pt-4 border-t border-slate-800 flex items-center justify-end gap-3">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          @click="handleClose"
          :disabled="isEditing"
        >
          {{ t('extra.cancel') || 'Batal' }}
        </Button>
        <Button
          type="submit"
          variant="primary"
          size="sm"
          :loading="isEditing"
        >
          {{ isEditing ? (t('extra.saving') || 'Menyimpan...') : (t('extra.saveBtn') || 'Simpan') }}
        </Button>
      </div>
    </form>
  </Modal>
</template>
