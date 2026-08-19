<script setup lang="ts">
import { ref, watch } from 'vue'
import Modal from '../../../../components/ui/Modal.vue'
import Button from '../../../../components/ui/Button.vue'
import Input from '../../../../components/ui/Input.vue'
import { useWallets } from '../../../../composables/useWallets'
import { validateForm, UpdateWalletSchema } from '../../../../utils/validate'
import { useI18n } from '../../../../locales'
import type { Wallet, UpdateWalletPayload } from '../../../../services/wallet'

const props = defineProps<{
  isOpen: boolean
  wallet: Wallet | null
}>()

const emit = defineEmits<{
  (e: 'update:isOpen', value: boolean): void
}>()

const { editWallet, isEditing } = useWallets()
const { t } = useI18n()

const form = ref<UpdateWalletPayload>({
  name: '',
  description: '',
  minimum_limit: 0
})

const formErrors = ref<Record<string, string>>({})

watch(
  () => props.wallet,
  (w) => {
    if (w) {
      form.value = {
        name: w.name || '',
        description: w.description || '',
        minimum_limit: w.minimum_limit || 0
      }
      formErrors.value = {}
    }
  },
  { immediate: true }
)

const handleClose = () => {
  emit('update:isOpen', false)
  formErrors.value = {}
}

const handleSubmit = async () => {
  if (!props.wallet) return
  const validation = validateForm(UpdateWalletSchema, form.value)
  if (!validation.success) {
    formErrors.value = validation.errors
    return
  }

  formErrors.value = {}
  const result = await editWallet(props.wallet.id, validation.data)
  if (result !== null) {
    handleClose()
  }
}
</script>

<template>
  <Modal
    :is-open="isOpen"
    :title="t('modals.editWallet.title') || 'Edit Dompet'"
    :description="t('modals.editWallet.subtitle') || 'Perbarui detail nama, deskripsi, dan limit minimum dompet'"
    @close="handleClose"
  >
    <form @submit.prevent="handleSubmit" class="flex flex-col gap-4">
      <Input
        v-model="form.name"
        :label="t('extra.walletNameLabel') || 'Nama Dompet'"
        :placeholder="t('extra.walletNamePlaceholder') || 'Nama dompet'"
        :error="formErrors.name"
        required
      />

      <Input
        v-model="form.description!"
        :label="t('extra.descLabel') || 'Keterangan'"
        :placeholder="t('extra.descPlaceholder') || 'Keterangan'"
        :error="formErrors.description"
      />

      <Input
        v-model="form.minimum_limit"
        type="number"
        :label="t('extra.minLimitLabel') || 'Batas Minimum (Rp)'"
        :error="formErrors.minimum_limit"
        :min="0"
        required
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
