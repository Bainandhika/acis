<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from '../locales'

const props = defineProps<{
  familyName?: string
  selectedMonth: number
  selectedYear: number
  pendingProposalsCount: number
}>()

const emit = defineEmits<{
  (e: 'update-period', payload: { month: number; year: number }): void
  (e: 'open-notifications'): void
}>()

const { t } = useI18n()
const isPeriodDropdownOpen = ref(false)

const months = [
  { value: 1, label: 'Jan' },
  { value: 2, label: 'Feb' },
  { value: 3, label: 'Mar' },
  { value: 4, label: 'Apr' },
  { value: 5, label: 'May' },
  { value: 6, label: 'Jun' },
  { value: 7, label: 'Jul' },
  { value: 8, label: 'Aug' },
  { value: 9, label: 'Sep' },
  { value: 10, label: 'Oct' },
  { value: 11, label: 'Nov' },
  { value: 12, label: 'Dec' },
]

const years = computed(() => [props.selectedYear - 1, props.selectedYear, props.selectedYear + 1])

const currentMonthLabel = computed(() => {
  const m = months.find(item => item.value === props.selectedMonth)
  return m ? m.label : 'May'
})

const selectPeriod = (month: number, year: number) => {
  emit('update-period', { month, year })
  isPeriodDropdownOpen.value = false
}
</script>

<template>
  <header class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6">
    <!-- Title & Greeting -->
    <div>
      <h1 class="text-2xl sm:text-3xl font-extrabold text-slate-900 tracking-tight font-sans">
        {{ t('extra.headerTitle') }}
      </h1>
      <p class="text-xs sm:text-sm text-slate-500 font-medium mt-1">
        {{ t('extra.welcome', { name: familyName || t('extra.defaultFamily'), month: currentMonthLabel }) }}
      </p>
    </div>

    <!-- Actions: Date Period Pill & Notification Bell -->
    <div class="flex items-center gap-3 shrink-0">
      <!-- Date Period Selector Button & Dropdown -->
      <div class="relative">
        <button
          @click="isPeriodDropdownOpen = !isPeriodDropdownOpen"
          class="flex items-center gap-2 px-3.5 py-2 bg-white hover:bg-slate-50 border border-slate-200/90 rounded-xl text-xs font-semibold text-slate-700 shadow-sm transition active:scale-95 cursor-pointer"
        >
          <svg class="w-4 h-4 text-slate-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
            <line x1="16" y1="2" x2="16" y2="6"></line>
            <line x1="8" y1="2" x2="8" y2="6"></line>
            <line x1="3" y1="10" x2="21" y2="10"></line>
          </svg>
          <span>{{ currentMonthLabel }} {{ selectedYear }}</span>
          <svg class="w-3.5 h-3.5 text-slate-400 ml-0.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="6 9 12 15 18 9"></polyline>
          </svg>
        </button>

        <!-- Dropdown Popover -->
        <div
          v-if="isPeriodDropdownOpen"
          class="absolute right-0 mt-2 w-56 bg-white border border-slate-200 rounded-2xl shadow-xl p-3 z-40 animate-in fade-in zoom-in-95 duration-100"
        >
          <div class="text-[11px] font-bold uppercase text-slate-400 tracking-wider mb-2 px-1">
            {{ t('extra.selectPeriod') }}
          </div>
          <div class="grid grid-cols-3 gap-1.5 mb-3">
            <button
              v-for="m in months"
              :key="m.value"
              @click="selectPeriod(m.value, selectedYear)"
              class="px-2 py-1.5 rounded-lg text-xs font-semibold transition text-center cursor-pointer"
              :class="m.value === selectedMonth ? 'bg-teal-700 text-white shadow-sm' : 'text-slate-600 hover:bg-slate-100'"
            >
              {{ m.label }}
            </button>
          </div>
          <div class="flex items-center justify-between border-t border-slate-100 pt-2 px-1">
            <span class="text-xs text-slate-500 font-medium">{{ t('extra.year') }}:</span>
            <div class="flex gap-1">
              <button
                v-for="y in years"
                :key="y"
                @click="selectPeriod(selectedMonth, y)"
                class="px-2 py-1 rounded-md text-xs font-bold transition cursor-pointer"
                :class="y === selectedYear ? 'bg-slate-900 text-white' : 'text-slate-600 hover:bg-slate-100'"
              >
                {{ y }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Notification Bell Button with badge -->
      <button
        @click="emit('open-notifications')"
        class="relative w-9 h-9 flex items-center justify-center bg-white hover:bg-slate-50 border border-slate-200/90 rounded-xl text-slate-600 shadow-sm transition active:scale-95 cursor-pointer"
        title="Notifications & Proposals"
      >
        <svg class="w-4 h-4 text-slate-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
          <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
        </svg>
        <span
          v-if="pendingProposalsCount > 0"
          class="absolute -top-1 -right-1 w-3.5 h-3.5 bg-rose-500 text-white text-[9px] font-black rounded-full flex items-center justify-center ring-2 ring-white"
        >
          {{ pendingProposalsCount }}
        </span>
      </button>
    </div>
  </header>
</template>
