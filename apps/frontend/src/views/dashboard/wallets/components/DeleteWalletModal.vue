<script setup lang="ts">
import Modal from '../../../../components/ui/Modal.vue'
import Button from '../../../../components/ui/Button.vue'
import { useWallets } from '../../../../composables/useWallets'
import { useI18n } from '../../../../locales'
import type { Wallet } from '../../../../services/wallet'

const props = defineProps<{
  isOpen: boolean
  wallet: Wallet | null
}>()

const emit = defineEmits<{
  (e: 'update:isOpen', value: boolean): void
}>()

const { removeWallet, isRemoving } = useWallets()
const { t } = useI18n()

const handleClose = () => {
  emit('update:isOpen', false)
}

const handleConfirm = async () => {
  if (!props.wallet) return
  const result = await removeWallet(props.wallet.id)
  if (result !== null) {
    handleClose()
  }
}
</script>

<template>
  <Modal
    :is-open="isOpen"
    :title="t('modals.confirmDelete.titleWallet') || 'Hapus Dompet?'"
    :description="t('modals.confirmDelete.descWallet') || 'Dompet ini beserta pengaturan alokasinya akan dihapus secara permanen.'"
    @close="handleClose"
  >
    <div class="flex flex-col gap-4">
      <div v-if="wallet" class="p-4 rounded-2xl bg-slate-950/80 border border-slate-800 flex items-center justify-between">
        <span class="text-xs font-bold text-white">{{ wallet.name }}</span>
        <span class="text-xs font-mono font-bold text-rose-400">ID: {{ wallet.id.slice(0, 8) }}...</span>
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
          {{ isRemoving ? (t('extra.deleting') || 'Menghapus...') : (t('modals.confirmDelete.confirm') || 'Hapus') }}
        </Button>
      </div>
    </div>
  </Modal>
</template>
