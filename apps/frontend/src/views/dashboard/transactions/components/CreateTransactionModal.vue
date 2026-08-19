<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import Modal from '../../../../components/ui/Modal.vue'
import Button from '../../../../components/ui/Button.vue'
import Input from '../../../../components/ui/Input.vue'
import { useTransactions } from '../../../../composables/useTransactions'
import { useProposals } from '../../../../composables/useProposals'
import { useWallets } from '../../../../composables/useWallets'
import { useAuthStore } from '../../../../stores/auth'
import { validateForm, CreateTransactionSchema, CreateProposalSchema } from '../../../../utils/validate'
import { formatRp } from '../../../../utils/format'
import { useI18n } from '../../../../locales'
import type { CreateTransactionPayload } from '../../../../services/transaction'

const props = defineProps<{
  isOpen: boolean
}>()

const emit = defineEmits<{
  (e: 'update:isOpen', value: boolean): void
}>()

const { addTransaction, isAdding: isTxAdding } = useTransactions()
const { createProposal, isCreating: isProposalCreating } = useProposals()
const { wallets } = useWallets()
const authStore = useAuthStore()
const { t } = useI18n()

const isAdmin = computed(() => authStore.user?.role === 'admin')
const isLoading = computed(() => isTxAdding.value || isProposalCreating.value)

const form = ref<CreateTransactionPayload>({
  wallet_id: '',
  type: 'expense',
  amount: 0,
  description: ''
})

const formErrors = ref<Record<string, string>>({})

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      form.value = {
        wallet_id: wallets.value[0]?.id || '',
        type: 'expense',
        amount: 0,
        description: ''
      }
      formErrors.value = {}
    }
  }
)

const handleClose = () => {
  emit('update:isOpen', false)
}

const handleSubmit = async () => {
  if (isAdmin.value) {
    const validation = validateForm(CreateTransactionSchema, form.value)
    if (!validation.success) {
      formErrors.value = validation.errors
      return
    }

    formErrors.value = {}
    const result = await addTransaction(validation.data)
    if (result !== null) {
      handleClose()
    }
  } else {
    const proposalPayload = {
      wallet_id: form.value.wallet_id || (wallets.value[0]?.id || ''),
      title: `Pengajuan: ${form.value.description || 'Pengeluaran Baru'}`,
      amount: form.value.amount,
      description: form.value.description || 'Pengajuan transaksi',
      request_type: 'add_transaction' as const,
      payload: {
        wallet_id: form.value.wallet_id,
        type: form.value.type,
        amount: form.value.amount,
        description: form.value.description
      }
    }

    const validation = validateForm(CreateProposalSchema, proposalPayload)
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
}
</script>

<template>
  <Modal
    :is-open="isOpen"
    :title="isAdmin ? (t('extra.recordTxModalTitle') || 'Catat Transaksi Baru') : (t('extra.proposeTxModalTitle') || 'Ajukan Transaksi Baru')"
    :description="isAdmin ? (t('extra.recordTxModalDesc') || 'Catat arus kas masuk atau keluar langsung ke saldo dompet') : (t('extra.proposeTxModalDesc') || 'Ajukan pengeluaran untuk ditinjau oleh Admin keluarga')"
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
          <option v-if="form.type === 'income'" value="">
            {{ t('extra.primaryBalance') || 'Saldo Utama Keluarga' }}
          </option>
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
        :placeholder="t('extra.notesPlaceholder') || 'Contoh: Belanja mingguan di pasar'"
        :error="formErrors.description"
      />

      <div class="mt-4 pt-4 border-t border-slate-800 flex items-center justify-end gap-3">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          @click="handleClose"
          :disabled="isLoading"
        >
          {{ t('extra.cancel') || 'Batal' }}
        </Button>
        <Button
          type="submit"
          variant="primary"
          size="sm"
          :loading="isLoading"
        >
          {{ isLoading ? (t('extra.processing') || 'Memproses...') : (isAdmin ? (t('extra.saveTxBtn') || 'Simpan Transaksi') : (t('extra.sendProposalBtn') || 'Kirim Pengajuan')) }}
        </Button>
      </div>
    </form>
  </Modal>
</template>
