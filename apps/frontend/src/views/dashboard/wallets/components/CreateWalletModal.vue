<script setup lang="ts">
import { ref } from 'vue'
import Modal from '../../../../components/ui/Modal.vue'
import Button from '../../../../components/ui/Button.vue'
import Input from '../../../../components/ui/Input.vue'
import { useWallets } from '../../../../composables/useWallets'
import { validateForm, CreateWalletSchema } from '../../../../utils/validate'
import { useI18n } from '../../../../locales'
import type { CreateWalletPayload } from '../../../../services/wallet'

defineProps<{
  isOpen: boolean
}>()

const emit = defineEmits<{
  (e: 'update:isOpen', value: boolean): void
}>()

const { addWallet, isAdding } = useWallets()
const { t } = useI18n()

const form = ref<CreateWalletPayload>({
  name: '',
  description: '',
  initial_balance: 0,
  minimum_limit: 0
})

const formErrors = ref<Record<string, string>>({})

const resetForm = () => {
  form.value = {
    name: '',
    description: '',
    initial_balance: 0,
    minimum_limit: 0
  }
  formErrors.value = {}
}

const handleClose = () => {
  emit('update:isOpen', false)
  resetForm()
}

const handleSubmit = async () => {
  const validation = validateForm(CreateWalletSchema, form.value)
  if (!validation.success) {
    formErrors.value = validation.errors
    return
  }

  formErrors.value = {}
  const result = await addWallet(validation.data)
  if (result !== null) {
    handleClose()
  }
}
</script>

<template>
  <Modal
    :is-open="isOpen"
    :title="t('extra.createWalletModalTitle') || 'Buat Dompet Baru'"
    :description="t('extra.createWalletModalDesc') || 'Buat dompet anggaran untuk kategori tertentu'"
    @close="handleClose"
  >
    <form @submit.prevent="handleSubmit" class="flex flex-col gap-4">
      <Input
        v-model="form.name"
        :label="t('extra.walletNameLabel') || 'Nama Dompet'"
        :placeholder="t('extra.walletNamePlaceholder') || 'Contoh: Belanja Bulanan'"
        :error="formErrors.name"
        required
      />

      <Input
        v-model="form.description"
        :label="t('extra.descLabel') || 'Keterangan (Opsional)'"
        :placeholder="t('extra.descPlaceholder') || 'Contoh: Anggaran belanja pasar & supermarket'"
        :error="formErrors.description"
      />

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <Input
          v-model="form.initial_balance"
          type="number"
          :label="t('extra.initialBalanceLabel') || 'Saldo Awal (Rp)'"
          :error="formErrors.initial_balance"
          :min="0"
          required
        />

        <Input
          v-model="form.minimum_limit"
          type="number"
          :label="t('extra.minLimitLabel') || 'Batas Minimum (Rp)'"
          :error="formErrors.minimum_limit"
          :min="0"
          required
        />
      </div>

      <div class="mt-4 pt-4 border-t border-slate-800 flex items-center justify-end gap-3">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          @click="handleClose"
          :disabled="isAdding"
        >
          {{ t('extra.cancel') || 'Batal' }}
        </Button>
        <Button
          type="submit"
          variant="primary"
          size="sm"
          :loading="isAdding"
        >
          {{ isAdding ? (t('extra.saving') || 'Menyimpan...') : (t('extra.saveWalletBtn') || 'Simpan Dompet') }}
        </Button>
      </div>
    </form>
  </Modal>
</template>
