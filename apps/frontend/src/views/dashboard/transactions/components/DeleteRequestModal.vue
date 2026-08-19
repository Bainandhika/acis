<script setup lang="ts">
import { ref, watch } from 'vue'
import Modal from '../../../../components/ui/Modal.vue'
import Button from '../../../../components/ui/Button.vue'
import Input from '../../../../components/ui/Input.vue'
import { useProposals } from '../../../../composables/useProposals'
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
const { t } = useI18n()

const reason = ref('')
const reasonError = ref('')

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      reason.value = ''
      reasonError.value = ''
    }
  }
)

const handleClose = () => {
  emit('update:isOpen', false)
}

const handleSubmit = async () => {
  if (!props.transaction) return

  if (!reason.value.trim() || reason.value.trim().length < 3) {
    reasonError.value = 'Alasan harus diisi minimal 3 karakter'
    return
  }

  reasonError.value = ''

  const payload = {
    wallet_id: props.transaction.wallet_id,
    title: `Hapus Transaksi: ${props.transaction.description || formatRp(props.transaction.amount)}`,
    amount: props.transaction.amount,
    description: reason.value.trim(),
    request_type: 'delete_transaction' as const,
    target_transaction_id: props.transaction.id
  }

  const validation = validateForm(CreateProposalSchema, payload)
  if (!validation.success) {
    reasonError.value = validation.errors.description || 'Validasi gagal'
    return
  }

  const result = await createProposal(validation.data)
  if (result !== null) {
    handleClose()
  }
}
</script>

<template>
  <Modal
    :is-open="isOpen"
    :title="t('extra.requestDeleteTitle') || 'Ajukan Penghapusan Transaksi'"
    :description="t('extra.requestDeleteDesc') || 'Ajukan permintaan penghapusan transaksi ke Admin'"
    @close="handleClose"
  >
    <form @submit.prevent="handleSubmit" class="flex flex-col gap-4">
      <div v-if="transaction" class="p-4 rounded-2xl bg-slate-950 border border-slate-800 flex items-center justify-between">
        <div class="flex flex-col">
          <span class="text-xs font-bold text-white">{{ transaction.description || 'Transaksi' }}</span>
          <span class="text-[10px] text-slate-400 font-mono">{{ transaction.created_at }}</span>
        </div>
        <span class="text-sm font-black text-rose-400">
          {{ formatRp(transaction.amount) }}
        </span>
      </div>

      <Input
        v-model="reason"
        :label="t('extra.reasonLabel') || 'Alasan Penghapusan'"
        :placeholder="t('extra.reasonPlaceholder') || 'Contoh: Transaksi dobel atau salah input'"
        :error="reasonError"
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
          variant="danger"
          size="sm"
          :loading="isCreating"
        >
          {{ isCreating ? (t('extra.processing') || 'Mengirim...') : (t('extra.sendProposalBtn') || 'Kirim Pengajuan') }}
        </Button>
      </div>
    </form>
  </Modal>
</template>
