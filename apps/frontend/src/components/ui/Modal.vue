<script setup lang="ts">
import { watch, onMounted, onUnmounted } from 'vue'

const props = withDefaults(
  defineProps<{
    isOpen: boolean
    title?: string
    description?: string
    maxWidth?: 'sm' | 'md' | 'lg' | 'xl'
  }>(),
  {
    title: '',
    description: '',
    maxWidth: 'md'
  }
)

const emit = defineEmits<{
  (e: 'update:isOpen', value: boolean): void
  (e: 'close'): void
}>()

const close = () => {
  emit('update:isOpen', false)
  emit('close')
}

const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && props.isOpen) {
    close()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})

watch(
  () => props.isOpen,
  (val) => {
    if (val) {
      document.body.classList.add('overflow-hidden')
    } else {
      document.body.classList.remove('overflow-hidden')
    }
  }
)

const sizeClasses = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-xl'
}
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="ease-out duration-200"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="ease-in duration-150"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="isOpen"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 overflow-y-auto bg-slate-950/60 backdrop-blur-sm"
        @click.self="close"
      >
        <Transition
          enter-active-class="ease-out duration-200"
          enter-from-class="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
          enter-to-class="opacity-100 translate-y-0 sm:scale-100"
          leave-active-class="ease-in duration-150"
          leave-from-class="opacity-100 translate-y-0 sm:scale-100"
          leave-to-class="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
        >
          <div
            v-if="isOpen"
            class="relative w-full bg-white dark:bg-slate-900 rounded-3xl p-6 sm:p-7 shadow-2xl border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-slate-100 max-h-[90vh] flex flex-col justify-between"
            :class="sizeClasses[maxWidth]"
          >
            <!-- Header -->
            <div class="flex items-start justify-between gap-4 mb-4">
              <div>
                <slot name="header">
                  <h3 v-if="title" class="font-extrabold text-lg text-slate-900 dark:text-white tracking-tight">
                    {{ title }}
                  </h3>
                  <p v-if="description" class="text-xs text-slate-500 dark:text-slate-400 mt-1">
                    {{ description }}
                  </p>
                </slot>
              </div>
              <button
                @click="close"
                class="p-1 rounded-xl text-slate-400 hover:text-slate-600 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-800 transition cursor-pointer shrink-0"
                aria-label="Close modal"
              >
                ✕
              </button>
            </div>

            <!-- Content Body -->
            <div class="overflow-y-auto pr-1">
              <slot />
            </div>

            <!-- Footer / Action bar -->
            <div v-if="$slots.footer" class="mt-6 pt-4 border-t border-slate-100 dark:border-slate-800 flex items-center justify-end gap-3">
              <slot name="footer" />
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>
