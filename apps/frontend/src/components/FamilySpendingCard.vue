<script setup lang="ts">
import { formatRp } from '../composables/useCurrency'
import { useI18n } from '../locales'

export interface MemberSpendItem {
  id: string
  name: string
  role: string
  spent: number
  avatar?: string
}

const props = defineProps<{
  members: MemberSpendItem[]
}>()

const emit = defineEmits<{
  (e: 'manage-family'): void
}>()

const { t } = useI18n()
</script>

<template>
  <div class="bg-white rounded-3xl p-6 sm:p-7 border border-slate-100/90 shadow-sm flex flex-col justify-between h-full min-h-[220px]">
    <!-- Header with Manage Link -->
    <div class="flex items-center justify-between pb-3">
      <h3 class="text-base font-bold text-slate-900">
        {{ t('dashboard.family.title') }}
      </h3>
      <button
        @click="emit('manage-family')"
        class="text-xs font-semibold text-teal-700 hover:text-teal-800 transition hover:underline cursor-pointer"
      >
        {{ t('dashboard.family.edit') }}
      </button>
    </div>

    <!-- Member List Breakdown -->
    <div class="flex flex-col gap-3 mt-2">
      <div v-if="members.length === 0" class="text-xs text-slate-400 text-center py-6">
        {{ t('dashboard.familyMembers.noMembers') }}
      </div>

      <div
        v-for="member in members"
        :key="member.id"
        class="flex items-center justify-between py-1"
      >
        <!-- Member Profile -->
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-full bg-gradient-to-tr from-slate-200 to-slate-100 border border-slate-200 flex items-center justify-center text-xs font-bold text-slate-700 shrink-0 overflow-hidden shadow-inner">
            <span v-if="!member.avatar">{{ member.name.charAt(0) }}</span>
            <img v-else :src="member.avatar" :alt="member.name" class="w-full h-full object-cover" />
          </div>
          <div class="flex flex-col text-left">
            <span class="text-xs font-bold text-slate-900 leading-tight">
              {{ member.name }}
            </span>
            <span class="text-[10px] text-slate-400 font-medium">
              {{ member.role }}
            </span>
          </div>
        </div>

        <!-- Monthly Spend -->
        <div class="flex flex-col items-end text-right">
          <span class="text-xs font-bold text-slate-900 font-sans">
            {{ formatRp(member.spent) }}
          </span>
          <span class="text-[10px] text-slate-400 font-medium">
            {{ t('extra.thisMonth') }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
