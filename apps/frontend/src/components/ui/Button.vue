<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'success'
    size?: 'xs' | 'sm' | 'md' | 'lg'
    loading?: boolean
    disabled?: boolean
    type?: 'button' | 'submit' | 'reset'
  }>(),
  {
    variant: 'primary',
    size: 'md',
    loading: false,
    disabled: false,
    type: 'button'
  }
)

const emit = defineEmits<{
  (e: 'click', event: MouseEvent): void
}>()

const variantClasses = {
  primary:
    'bg-teal-700 hover:bg-teal-800 text-white shadow-sm border border-teal-600/30 hover:border-teal-600/60 focus:ring-teal-500',
  secondary:
    'bg-slate-800 hover:bg-slate-700 text-slate-100 border border-slate-700 hover:border-slate-600 focus:ring-slate-500',
  danger:
    'bg-rose-700 hover:bg-rose-800 text-white shadow-sm border border-rose-600/40 focus:ring-rose-500',
  success:
    'bg-emerald-700 hover:bg-emerald-800 text-white shadow-sm border border-emerald-600/40 focus:ring-emerald-500',
  ghost:
    'bg-transparent hover:bg-slate-800/50 text-slate-400 hover:text-slate-100 border border-transparent focus:ring-slate-500'
}

const sizeClasses = {
  xs: 'px-2.5 py-1 text-xs rounded-lg gap-1.5',
  sm: 'px-3.5 py-1.5 text-xs rounded-xl gap-2',
  md: 'px-4 py-2.5 text-sm rounded-xl gap-2',
  lg: 'px-6 py-3 text-base rounded-2xl gap-2.5'
}
</script>

<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    @click="emit('click', $event)"
    class="inline-flex items-center justify-center font-bold tracking-tight transition-all duration-150 cursor-pointer focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed select-none"
    :class="[variantClasses[variant], sizeClasses[size]]"
  >
    <svg
      v-if="loading"
      class="animate-spin -ml-1 mr-2 h-4 w-4 text-current"
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
    >
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
      <path
        class="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
      ></path>
    </svg>
    <slot />
  </button>
</template>
