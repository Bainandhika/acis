<script setup lang="ts">
withDefaults(
  defineProps<{
    type?: 'card' | 'wallet' | 'table-row' | 'stat' | 'line'
    count?: number
  }>(),
  {
    type: 'card',
    count: 1
  }
)
</script>

<template>
  <div v-for="i in count" :key="i" class="animate-pulse">
    <!-- Wallet Skeleton -->
    <div
      v-if="type === 'wallet'"
      class="bg-slate-900/60 rounded-3xl p-6 border border-slate-800/60 shadow-card flex flex-col justify-between h-48"
    >
      <div class="flex items-start justify-between">
        <div class="space-y-2 w-2/3">
          <div class="h-4 bg-slate-800 rounded-lg w-3/4"></div>
          <div class="h-3 bg-slate-800/60 rounded w-1/2"></div>
        </div>
        <div class="w-8 h-8 rounded-xl bg-slate-800"></div>
      </div>
      <div class="space-y-2">
        <div class="h-6 bg-slate-800 rounded-lg w-1/2"></div>
        <div class="w-full h-2 bg-slate-800/80 rounded-full"></div>
      </div>
    </div>

    <!-- Stat Skeleton -->
    <div
      v-else-if="type === 'stat'"
      class="bg-slate-900/60 rounded-3xl p-6 border border-slate-800/60 shadow-card space-y-3"
    >
      <div class="h-3.5 bg-slate-800 rounded w-1/3"></div>
      <div class="h-8 bg-slate-800 rounded-xl w-3/5"></div>
      <div class="h-3 bg-slate-800/60 rounded w-1/4"></div>
    </div>

    <!-- Table Row Skeleton -->
    <div
      v-else-if="type === 'table-row'"
      class="flex items-center justify-between p-4 bg-slate-900/30 rounded-2xl border border-slate-800/40 my-2"
    >
      <div class="flex items-center gap-3 w-1/3">
        <div class="w-8 h-8 rounded-full bg-slate-800"></div>
        <div class="space-y-1 w-3/4">
          <div class="h-3.5 bg-slate-800 rounded"></div>
          <div class="h-2.5 bg-slate-800/60 rounded w-1/2"></div>
        </div>
      </div>
      <div class="h-4 bg-slate-800 rounded w-24"></div>
    </div>

    <!-- Line Skeleton -->
    <div v-else-if="type === 'line'" class="h-3.5 bg-slate-800 rounded w-full my-1.5"></div>

    <!-- Generic Card Skeleton -->
    <div
      v-else
      class="bg-slate-900/60 rounded-3xl p-6 border border-slate-800/60 shadow-card space-y-4"
    >
      <div class="h-4 bg-slate-800 rounded w-1/4"></div>
      <div class="h-20 bg-slate-800/50 rounded-2xl"></div>
    </div>
  </div>
</template>
