<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useFamily } from '../../../composables/useFamily'
import { useAuthStore } from '../../../stores/auth'
import { validateForm, UpdateFamilyNameSchema, FamilySettingsSchema } from '../../../utils/validate'
import { useI18n } from '../../../locales'
import type { FamilyMember } from '../../../types'

// Components
import Button from '../../../components/ui/Button.vue'
import Skeleton from '../../../components/ui/Skeleton.vue'
import DeleteMemberModal from './components/DeleteMemberModal.vue'
import TelegramModal from './components/TelegramModal.vue'

const {
  family,
  members,
  inviteCode,
  isLoading,
  updateFamilyName,
  isUpdatingName,
  updateMonthlyIncome,
  isUpdatingIncome,
  disconnectTelegram,
  isDisconnectingTelegram,
  copyInviteCode
} = useFamily()

const authStore = useAuthStore()
const { t } = useI18n()

const isAdmin = computed(() => authStore.user?.role === 'admin')

const editName = ref('')
const nameError = ref('')

const editIncome = ref(0)
const incomeError = ref('')

const isTelegramModalOpen = ref(false)
const isDeleteMemberOpen = ref(false)
const memberToDelete = ref<FamilyMember | null>(null)

watch(
  () => family.value,
  (f) => {
    if (f) {
      editName.value = f.name || ''
      editIncome.value = f.monthly_income || 0
    }
  },
  { immediate: true }
)

const handleSaveName = async () => {
  const validation = validateForm(UpdateFamilyNameSchema, { name: editName.value })
  if (!validation.success) {
    nameError.value = validation.errors.name || 'Nama keluarga tidak valid'
    return
  }
  nameError.value = ''
  await updateFamilyName(validation.data.name)
}

const handleSaveIncome = async () => {
  const validation = validateForm(FamilySettingsSchema, { monthly_income: editIncome.value })
  if (!validation.success) {
    incomeError.value = validation.errors.monthly_income || 'Nominal tidak valid'
    return
  }
  incomeError.value = ''
  await updateMonthlyIncome(validation.data.monthly_income)
}

const openDeleteMember = (member: FamilyMember) => {
  memberToDelete.value = member
  isDeleteMemberOpen.value = true
}
</script>

<template>
  <div class="flex flex-col gap-8 max-w-3xl pb-12">
    <!-- Header -->
    <div>
      <h1 class="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
        {{ t('modals.familyManage.title') || 'Pengaturan Keluarga' }}
      </h1>
      <p class="text-xs text-slate-400 mt-1">
        Kelola profil keluarga, target anggaran bulanan, integrasi bot Telegram, dan anggota grup.
      </p>
    </div>

    <!-- Loading Skeleton -->
    <div v-if="isLoading" class="space-y-6">
      <Skeleton v-for="i in 3" :key="i" type="card" />
    </div>

    <div v-else class="flex flex-col gap-6">
      <!-- General Details Card -->
      <div class="card-neo bg-slate-900/90 rounded-3xl p-6 sm:p-7 border border-slate-800/90 shadow-card flex flex-col gap-6">
        <h3 class="text-base font-bold text-white border-b border-slate-800/80 pb-3">
          Informasi Utama
        </h3>

        <!-- Family Name Field -->
        <div>
          <label class="text-xs font-bold text-slate-300 block mb-1.5">
            {{ t('extra.familyNameLabel') || 'Nama Keluarga' }}
          </label>
          <div class="flex flex-col sm:flex-row gap-2.5">
            <input
              type="text"
              v-model="editName"
              :disabled="!isAdmin"
              class="flex-1 px-4 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-xs font-semibold text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-teal-500/30 focus:border-teal-600 disabled:opacity-50"
            />
            <Button
              v-if="isAdmin"
              variant="primary"
              size="sm"
              :loading="isUpdatingName"
              @click="handleSaveName"
            >
              {{ isUpdatingName ? (t('extra.saving') || 'Menyimpan...') : (t('extra.saveBtn') || 'Simpan') }}
            </Button>
          </div>
          <p v-if="nameError" class="text-[11px] font-semibold text-rose-500 mt-1">
            {{ nameError }}
          </p>
        </div>

        <!-- Monthly Income Setting (Admin only) -->
        <div v-if="isAdmin">
          <label class="text-xs font-bold text-slate-300 block mb-1.5">
            {{ t('extra.monthlyIncomeLabel') || 'Target Pemasukan / Anggaran Bulanan (Rp)' }}
          </label>
          <div class="flex flex-col sm:flex-row gap-2.5">
            <input
              type="number"
              v-model.number="editIncome"
              :min="0"
              class="flex-1 px-4 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-xs font-semibold text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-teal-500/30 focus:border-teal-600"
            />
            <Button
              variant="primary"
              size="sm"
              :loading="isUpdatingIncome"
              @click="handleSaveIncome"
            >
              {{ isUpdatingIncome ? (t('extra.saving') || 'Menyimpan...') : (t('extra.saveBtn') || 'Simpan') }}
            </Button>
          </div>
          <p v-if="incomeError" class="text-[11px] font-semibold text-rose-500 mt-1">
            {{ incomeError }}
          </p>
        </div>

        <!-- Invite Code -->
        <div class="pt-4 border-t border-slate-800/80 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <span class="text-xs font-bold text-slate-300 block">
              {{ t('extra.inviteCodeLabel') || 'Kode Undangan Keluarga' }}
            </span>
            <span class="text-base font-mono font-black text-teal-400 select-all tracking-wider">
              {{ inviteCode }}
            </span>
          </div>
          <Button
            variant="secondary"
            size="sm"
            @click="copyInviteCode"
          >
            📋 {{ t('extra.copyCodeBtn') || 'Salin Kode' }}
          </Button>
        </div>
      </div>

      <!-- Telegram Integration Card -->
      <div class="card-neo bg-slate-900/90 rounded-3xl p-6 sm:p-7 border border-slate-800/90 shadow-card flex flex-col gap-4">
        <h3 class="text-base font-bold text-white border-b border-slate-800/80 pb-3">
          {{ t('extra.telegramIntegration') || 'Integrasi Telegram Bot' }}
        </h3>

        <!-- Connected state -->
        <div
          v-if="family?.telegram_chat_id"
          class="flex flex-col sm:flex-row sm:items-center justify-between p-4 rounded-2xl bg-emerald-950/30 border border-emerald-500/30 gap-3"
        >
          <div class="flex items-center gap-3">
            <span class="relative flex h-3 w-3">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span class="relative inline-flex rounded-full h-3 w-3 bg-emerald-500"></span>
            </span>
            <div class="flex flex-col">
              <span class="text-xs font-bold text-emerald-300">{{ t('extra.telegramConnected') || 'Bot Terhubung' }}</span>
              <span class="text-[11px] text-emerald-500 font-mono">Chat ID: {{ family.telegram_chat_id }}</span>
            </div>
          </div>
          <Button
            v-if="isAdmin"
            variant="danger"
            size="xs"
            :loading="isDisconnectingTelegram"
            @click="disconnectTelegram"
          >
            {{ t('extra.disconnectTelegramBtn') || 'Putuskan Bot' }}
          </Button>
        </div>

        <!-- Disconnected state -->
        <div
          v-else
          class="flex flex-col sm:flex-row sm:items-center justify-between p-4 rounded-2xl bg-slate-950 border border-slate-800 gap-3"
        >
          <div class="flex items-center gap-3">
            <span class="w-3 h-3 rounded-full bg-amber-400 shrink-0"></span>
            <span class="text-xs font-semibold text-slate-400">
              {{ t('extra.telegramNotConnected') || 'Bot Telegram belum terhubung' }}
            </span>
          </div>
          <Button
            variant="primary"
            size="xs"
            @click="isTelegramModalOpen = true"
          >
            {{ t('extra.connectTelegramBtn') || 'Hubungkan Bot' }}
          </Button>
        </div>
      </div>

      <!-- Members Roster Card -->
      <div class="card-neo bg-slate-900/90 rounded-3xl p-6 sm:p-7 border border-slate-800/90 shadow-card flex flex-col gap-4">
        <div class="flex items-center justify-between border-b border-slate-800/80 pb-3">
          <h3 class="text-base font-bold text-white">
            {{ t('dashboard.familyMembers.title') || 'Daftar Anggota Keluarga' }}
          </h3>
          <span class="px-2.5 py-0.5 rounded-full text-xs font-bold bg-slate-800 text-teal-300 border border-slate-700">
            {{ members.length }} Anggota
          </span>
        </div>

        <div class="divide-y divide-slate-800/80">
          <div
            v-for="m in members"
            :key="m.id"
            class="py-3.5 flex items-center justify-between gap-4"
          >
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-full bg-slate-800 border border-slate-700 flex items-center justify-center text-xs font-bold text-teal-300">
                {{ (m.user_name || m.role).charAt(0).toUpperCase() }}
              </div>
              <div class="flex flex-col">
                <span class="text-xs font-bold text-white">{{ m.user_name || m.user_id }}</span>
                <span class="text-[10px] text-slate-400 capitalize font-medium">{{ m.role }}</span>
              </div>
            </div>

            <!-- Remove Member button for Admin (cannot remove self) -->
            <Button
              v-if="isAdmin && m.user_id !== authStore.user?.id"
              variant="danger"
              size="xs"
              @click="openDeleteMember(m)"
            >
              {{ t('extra.removeMemberBtn') || 'Keluarkan' }}
            </Button>
          </div>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <DeleteMemberModal v-model:isOpen="isDeleteMemberOpen" :member="memberToDelete" />
    <TelegramModal v-model:isOpen="isTelegramModalOpen" />
  </div>
</template>
