<script setup lang="ts">
import { formatRp } from '../../utils/format'
import { useI18n } from '../../locales'

export interface MemberSpendItem {
  id: string
  name: string
  role: string
  spent: number
  avatar?: string
}

defineProps<{
  members: MemberSpendItem[]
}>()

const emit = defineEmits<{
  (e: 'manage-family'): void
}>()

const { t } = useI18n()
</script>

<template>
  <div class="card-neo bg-slate-900/90 rounded-3xl p-6 sm:p-7 border border-slate-800/90 shadow-card flex flex-col justify-between h-full min-h-[220px]">
    <!-- Header with Manage Link -->
    <div class="flex items-center justify-between pb-3 border-b border-slate-800/60">
      <h3 class="text-base font-bold text-white">
        {{ t('dashboard.family.title') || 'Pengeluaran Keluarga' }}
      </h3>
      <button
        @click="emit('manage-family')"
        class="text-xs font-semibold text-teal-400 hover:text-teal-300 transition hover:underline cursor-pointer"
      >
        {{ t('dashboard.family.edit') || 'Kelola' }}
      </button>
    </div>

    <!-- Member List Breakdown -->
    <div class="flex flex-col gap-3 mt-3">
      <div v-if="members.length === 0" class="text-xs text-slate-500 text-center py-6">
        {{ t('dashboard.familyMembers.noMembers') || 'Belum ada data anggota' }}
      </div>

      <div
        v-for="member in members"
        :key="member.id"
        class="flex items-center justify-between py-1"
      >
        <!-- Member Profile -->
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-full bg-slate-800 border border-slate-700 flex items-center justify-center text-xs font-bold text-teal-300 shrink-0 overflow-hidden">
            <span v-if="!member.avatar">{{ member.name.charAt(0).toUpperCase() }}</span>
            <img v-else :src="member.avatar" :alt="member.name" class="w-full h-full object-cover" />
          </div>
          <div class="flex flex-col text-left">
            <span class="text-xs font-bold text-slate-100 leading-tight">
              {{ member.name }}
            </span>
            <span class="text-[10px] text-slate-400 font-medium capitalize">
              {{ member.role }}
            </span>
          </div>
        </div>

        <!-- Monthly Spend -->
        <div class="flex flex-col items-end text-right">
          <span class="text-xs font-bold text-slate-100 font-sans">
            {{ formatRp(member.spent) }}
          </span>
          <span class="text-[10px] text-slate-400 font-medium">
            {{ t('extra.thisMonth') || 'Bulan ini' }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
