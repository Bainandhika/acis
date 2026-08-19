<script setup lang="ts">
withDefaults(
  defineProps<{
    modelValue: string | number
    label?: string
    error?: string
    placeholder?: string
    type?: string
    disabled?: boolean
    required?: boolean
    hint?: string
    min?: number
    max?: number
  }>(),
  {
    label: '',
    error: '',
    placeholder: '',
    type: 'text',
    disabled: false,
    required: false,
    hint: ''
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: any): void
}>()

const handleInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  const value = target.type === 'number' ? (target.value === '' ? 0 : Number(target.value)) : target.value
  emit('update:modelValue', value)
}
</script>

<template>
  <div class="flex flex-col gap-1.5 w-full">
    <label v-if="label" class="text-xs font-bold text-slate-700 dark:text-slate-300">
      {{ label }}
      <span v-if="required" class="text-rose-500">*</span>
    </label>
    <div class="relative">
      <input
        :type="type"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :min="min"
        :max="max"
        @input="handleInput"
        class="w-full px-4 py-2.5 bg-slate-50 dark:bg-slate-950 border rounded-xl text-xs font-semibold text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 transition disabled:opacity-50 disabled:cursor-not-allowed"
        :class="[
          error
            ? 'border-rose-500/80 focus:ring-rose-500/30'
            : 'border-slate-200 dark:border-slate-800 focus:ring-teal-500/30 focus:border-teal-600'
        ]"
      />
    </div>
    <p v-if="error" class="text-[11px] font-semibold text-rose-500 animate-in fade-in duration-150">
      {{ error }}
    </p>
    <p v-else-if="hint" class="text-[11px] text-slate-400 dark:text-slate-500">
      {{ hint }}
    </p>
  </div>
</template>
