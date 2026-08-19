<script setup lang="ts">
import { useUI } from '../../composables/useUI'

const { toasts, removeToast } = useUI()

const getTypeClasses = (type: string) => {
  switch (type) {
    case 'success':
      return 'bg-emerald-950/90 text-emerald-100 border-emerald-500/40 shadow-emerald-950/50'
    case 'error':
      return 'bg-rose-950/90 text-rose-100 border-rose-500/40 shadow-rose-950/50'
    case 'warning':
      return 'bg-amber-950/90 text-amber-100 border-amber-500/40 shadow-amber-950/50'
    case 'info':
    default:
      return 'bg-slate-900/90 text-slate-100 border-slate-700/60 shadow-slate-950/50'
  }
}

const getIcon = (type: string) => {
  switch (type) {
    case 'success':
      return '✓'
    case 'error':
      return '✕'
    case 'warning':
      return '⚠'
    case 'info':
    default:
      return 'ℹ'
  }
}
</script>

<template>
  <div class="fixed bottom-6 right-6 z-50 flex flex-col gap-2.5 max-w-sm w-full pointer-events-none px-4 sm:px-0">
    <TransitionGroup
      enter-active-class="transform ease-out duration-300 transition"
      enter-from-class="translate-y-4 opacity-0 scale-95"
      enter-to-class="translate-y-0 opacity-100 scale-100"
      leave-active-class="transition ease-in duration-200"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="pointer-events-auto flex items-center justify-between gap-3 p-4 rounded-2xl border backdrop-blur-md shadow-xl text-xs font-semibold"
        :class="getTypeClasses(toast.type)"
        role="alert"
      >
        <div class="flex items-center gap-3 min-w-0">
          <span class="w-5 h-5 rounded-full flex items-center justify-center font-bold text-[11px] shrink-0 bg-white/10">
            {{ getIcon(toast.type) }}
          </span>
          <p class="truncate">{{ toast.message }}</p>
        </div>
        <button
          @click="removeToast(toast.id)"
          class="shrink-0 text-slate-400 hover:text-white transition p-1 cursor-pointer"
          aria-label="Dismiss notification"
        >
          ✕
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>
