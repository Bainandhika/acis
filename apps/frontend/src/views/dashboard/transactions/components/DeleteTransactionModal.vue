<script setup lang="ts">
import Modal from '../../../../components/ui/Modal.vue'
import Button from '../../../../components/ui/Button.vue'
import { useTransactions } from '../../../../composables/useTransactions'
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

const { removeTransaction, isRemoving } = useTransactions()
const { t } = useI18n()

const handleClose = () => {
  emit('update:isOpen', false)
}

const handleConfirm = async () => {
  if (!props.transaction) return
  const result = await removeTransaction(props.transaction.id)
  if (result !== null) {
    handleClose()
  }
}
</script>

<template>
  <Modal
    :is-open="isOpen"
    :title="t('extra.deleteTxTitle') || 'Hapus Transaksi?'"
    :description="t('extra.confirmDeleteTx') || 'Transaksi ini akan dihapus dan saldo dompet terkait akan disesuaikan kembali.'"
    @close="handleClose"
  >
    <div class="flex flex-col gap-4">
      <div v-if="transaction" class="p-4 rounded-2xl bg-slate-950 border border-slate-800 flex items-center justify-between">
        <div class="flex flex-col">
          <span class="text-xs font-bold text-white">{{ transaction.description || 'Transaksi' }}</span>
          <span class="text-[10px] text-slate-400 font-mono">{{ transaction.created_at }}</span>
        </div>
        <span class="text-sm font-black" :class="transaction.type === 'income' ? 'text-emerald-400' : 'text-rose-400'">
          {{ transaction.type === 'income' ? '+' : '-' }}{{ formatRp(transaction.amount) }}
        </span>
      </div>

      <div class="mt-4 pt-4 border-t border-slate-800 flex items-center justify-end gap-3">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          @click="handleClose"
          :disabled="isRemoving"
        >
          {{ t('extra.cancel') || 'Batal' }}
        </Button>
        <Button
          type="button"
          variant="danger"
          size="sm"
          :loading="isRemoving"
          @click="handleConfirm"
        >
          {{ isRemoving ? (t('extra.deleting') || 'Menghapus...') : (t('extra.deleteBtn') || 'Hapus') }}
        </Button>
      </div>
    </div>
  </Modal>
</template>
