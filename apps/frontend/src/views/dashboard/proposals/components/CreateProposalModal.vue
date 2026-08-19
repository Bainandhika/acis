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
import type { CreateProposalPayload } from '../../../../services/transaction'

const props = defineProps<{
  isOpen: boolean
}>()

const emit = defineEmits<{
  (e: 'update:isOpen', value: boolean): void
}>()

const { createProposal, isCreating } = useProposals()
const { wallets } = useWallets()
const { t } = useI18n()

const form = ref<CreateProposalPayload>({
  wallet_id: '',
  title: '',
  amount: 0,
  description: '',
  request_type: 'add_transaction'
})

const formErrors = ref<Record<string, string>>({})

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      form.value = {
        wallet_id: wallets.value[0]?.id || '',
        title: '',
        amount: 0,
        description: '',
        request_type: 'add_transaction'
      }
      formErrors.value = {}
    }
  }
)

const handleClose = () => {
  emit('update:isOpen', false)
}

const handleSubmit = async () => {
  const validation = validateForm(CreateProposalSchema, form.value)
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
    :title="t('extra.proposeTxModalTitle') || 'Ajukan Pengeluaran'"
    :description="t('extra.proposeTxModalDesc') || 'Ajukan pengeluaran baru untuk disetujui Admin'"
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
          <option v-for="w in wallets" :key="w.id" :value="w.id">
            {{ w.name }} (Saldo: {{ formatRp(w.current_balance) }})
          </option>
        </select>
        <p v-if="formErrors.wallet_id" class="text-[11px] font-semibold text-rose-500 mt-1">
          {{ formErrors.wallet_id }}
        </p>
      </div>

      <Input
        v-model="form.title"
        label="Judul Pengajuan"
        placeholder="Contoh: Pembelian buku pelajaran"
        :error="formErrors.title"
        required
      />

      <Input
        v-model="form.amount"
        type="number"
        :label="t('extra.amountLabel') || 'Nominal (Rp)'"
        :error="formErrors.amount"
        :min="1"
        required
      />

      <Input
        v-model="form.description"
        :label="t('extra.notesLabel') || 'Keterangan Rinci'"
        :placeholder="t('extra.notesPlaceholder') || 'Jelaskan keperluan dan detail pengeluaran'"
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
