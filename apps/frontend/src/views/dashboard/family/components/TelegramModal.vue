<script setup lang="ts">
import Modal from '../../../../components/ui/Modal.vue'
import Button from '../../../../components/ui/Button.vue'
import { useFamily } from '../../../../composables/useFamily'
import { useI18n } from '../../../../locales'

defineProps<{
  isOpen: boolean
}>()

const emit = defineEmits<{
  (e: 'update:isOpen', value: boolean): void
}>()

const { inviteCode } = useFamily()
const { t } = useI18n()

const handleClose = () => {
  emit('update:isOpen', false)
}
</script>

<template>
  <Modal
    :is-open="isOpen"
    :title="t('modals.telegram.title') || 'Hubungkan Bot Telegram'"
    :description="t('modals.telegram.subtitle') || 'Integrasikan bot Telegram untuk menerima notifikasi dan mencatat transaksi via chat'"
    @close="handleClose"
  >
    <div class="flex flex-col gap-4">
      <div class="flex flex-col gap-2.5 text-xs text-slate-300 bg-slate-950 p-4 rounded-2xl border border-slate-800">
        <p class="font-bold text-white">{{ t('modals.telegram.howToLink') || 'Cara Menghubungkan:' }}</p>
        <p>1. {{ t('modals.telegram.step1') || 'Buka bot Telegram ACIS di aplikasi Anda.' }}</p>
        <p>
          2. {{ t('modals.telegram.step2') || 'Kirim perintah:' }}
          <span class="font-mono font-bold bg-slate-900 px-2 py-0.5 rounded border border-slate-700 text-teal-400 select-all">
            /link {{ inviteCode }}
          </span>
        </p>
        <p>3. {{ t('modals.telegram.step3') || 'Bot akan terhubung secara otomatis ke grup keluarga Anda.' }}</p>
      </div>

      <div class="mt-4 pt-4 border-t border-slate-800 flex items-center justify-end">
        <Button
          type="button"
          variant="primary"
          size="sm"
          @click="handleClose"
        >
          {{ t('modals.telegram.close') || 'Tutup' }}
        </Button>
      </div>
    </div>
  </Modal>
</template>
