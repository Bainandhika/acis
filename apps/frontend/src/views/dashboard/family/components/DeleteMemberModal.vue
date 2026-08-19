<script setup lang="ts">
import Modal from '../../../../components/ui/Modal.vue'
import Button from '../../../../components/ui/Button.vue'
import { useFamily } from '../../../../composables/useFamily'
import { useI18n } from '../../../../locales'
import type { FamilyMember } from '../../../../types'

const props = defineProps<{
  isOpen: boolean
  member: FamilyMember | null
}>()

const emit = defineEmits<{
  (e: 'update:isOpen', value: boolean): void
}>()

const { removeMember, isRemovingMember } = useFamily()
const { t } = useI18n()

const handleClose = () => {
  emit('update:isOpen', false)
}

const handleConfirm = async () => {
  if (!props.member) return
  const result = await removeMember(props.member.id)
  if (result !== null) {
    handleClose()
  }
}
</script>

<template>
  <Modal
    :is-open="isOpen"
    :title="t('modals.familyManage.confirmRemoveTitle') || 'Keluarkan Anggota?'"
    :description="t('modals.familyManage.confirmRemoveDesc', { name: member?.user_name || member?.role || 'Anggota' }) || 'Anggota ini tidak akan lagi memiliki akses ke grup keluarga.'"
    @close="handleClose"
  >
    <div class="flex flex-col gap-4">
      <div v-if="member" class="p-4 rounded-2xl bg-slate-950 border border-slate-800 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 rounded-full bg-slate-800 flex items-center justify-center font-bold text-xs text-white">
            {{ (member.user_name || member.role).charAt(0).toUpperCase() }}
          </div>
          <div class="flex flex-col">
            <span class="text-xs font-bold text-white">{{ member.user_name || member.user_id }}</span>
            <span class="text-[10px] text-slate-400 capitalize">{{ member.role }}</span>
          </div>
        </div>
      </div>

      <div class="mt-4 pt-4 border-t border-slate-800 flex items-center justify-end gap-3">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          @click="handleClose"
          :disabled="isRemovingMember"
        >
          {{ t('extra.cancel') || 'Batal' }}
        </Button>
        <Button
          type="button"
          variant="danger"
          size="sm"
          :loading="isRemovingMember"
          @click="handleConfirm"
        >
          {{ isRemovingMember ? (t('extra.deleting') || 'Mengeluarkan...') : (t('extra.removeMemberBtn') || 'Keluarkan') }}
        </Button>
      </div>
    </div>
  </Modal>
</template>
