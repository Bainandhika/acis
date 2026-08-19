<script setup lang="ts">
import { ref, watch } from 'vue'
import Modal from '../../../../components/ui/Modal.vue'
import Button from '../../../../components/ui/Button.vue'
import Input from '../../../../components/ui/Input.vue'
import { useWallets } from '../../../../composables/useWallets'
import { useFamily } from '../../../../composables/useFamily'
import { validateForm, AllocateFundsSchema } from '../../../../utils/validate'
import { formatRp } from '../../../../utils/format'
import { useI18n } from '../../../../locales'
import type { Wallet } from '../../../../services/wallet'

const props = defineProps<{
  isOpen: boolean
  wallet: Wallet | null
}>()

const emit = defineEmits<{
  (e: 'update:isOpen', value: boolean): void
}>()

const { wallets } = useWallets()
const { allocateFunds, isAllocating, primaryBalance } = useFamily()
const { t } = useI18n()

const form = ref({
  wallet_id: '',
  amount: 0,
  description: ''
})

const formErrors = ref<Record<string, string>>({})

watch(
  () => [props.wallet, props.isOpen],
  ([w, open]) => {
    if (open) {
      form.value = {
        wallet_id: (w as Wallet)?.id || wallets.value[0]?.id || '',
        amount: 0,
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
  const validation = validateForm(AllocateFundsSchema, form.value)
  if (!validation.success) {
    formErrors.value = validation.errors
    return
  }

  formErrors.value = {}
  const result = await allocateFunds(validation.data)
  if (result !== null) {
    handleClose()
  }
}
</script>

<template>
  <Modal
    :is-open="isOpen"
    :title="t('extra.allocateModalTitle') || 'Alokasi Dana ke Dompet'"
    :description="t('extra.allocateModalDesc') || 'Pindahkan dana dari Saldo Utama Keluarga ke dompet spesifik'"
    @close="handleClose"
  >
    <form @submit.prevent="handleSubmit" class="flex flex-col gap-4">
      <!-- Primary Balance Info Box -->
      <div class="p-3.5 rounded-2xl bg-teal-950/40 border border-teal-500/20 flex items-center justify-between">
        <span class="text-xs font-bold text-slate-300">Saldo Utama Tersedia:</span>
        <span class="text-sm font-black text-teal-300">{{ formatRp(primaryBalance) }}</span>
      </div>

      <div>
        <label class="text-xs font-bold text-slate-300 block mb-1.5">
          {{ t('extra.selectTargetWallet') || 'Pilih Dompet Tujuan' }}
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
        v-model="form.amount"
        type="number"
        :label="t('extra.allocationAmount') || 'Nominal Alokasi (Rp)'"
        :error="formErrors.amount"
        :min="1"
        required
      />

      <Input
        v-model="form.description"
        :label="t('extra.memoOptional') || 'Catatan Alokasi (Opsional)'"
        :placeholder="t('extra.allocationMemoPlaceholder') || 'Contoh: Alokasi mingguan'"
        :error="formErrors.description"
      />

      <div class="mt-4 pt-4 border-t border-slate-800 flex items-center justify-end gap-3">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          @click="handleClose"
          :disabled="isAllocating"
        >
          {{ t('extra.cancel') || 'Batal' }}
        </Button>
        <Button
          type="submit"
          variant="primary"
          size="sm"
          :loading="isAllocating"
        >
          {{ isAllocating ? (t('extra.allocating') || 'Mengalokasikan...') : (t('extra.allocateFundsBtn') || 'Alokasikan Dana') }}
        </Button>
      </div>
    </form>
  </Modal>
</template>
