<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  scorePercentage: number
  totalSaved: number
  statusText?: string
}>()

const formatRupiah = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(val)
}

const clampedScore = computed(() => {
  return Math.min(Math.max(props.scorePercentage, 0), 100)
})

// Calculate stroke-dashoffset for semi-circular SVG meter (circumference for radius 70 = ~220 half)
const circumference = Math.PI * 70 // ~220
const strokeOffset = computed(() => {
  return circumference - (clampedScore.value / 100) * circumference
})
</script>

<template>
  <div class="card-neo p-5 flex flex-col justify-between">
    <!-- Header with 30d dropdown -->
    <div class="flex items-center justify-between">
      <div>
        <h3 class="font-bold text-slate-900 text-sm">Kesehatan Finansial</h3>
        <p class="text-[11px] text-slate-400 font-medium">Status keuangan keluarga</p>
      </div>
      <span class="text-[11px] font-bold px-2 py-0.5 rounded-lg bg-slate-100 text-slate-600">30h</span>
    </div>

    <!-- Metric Value -->
    <div class="my-2">
      <h4 class="text-2xl font-black text-slate-900 tracking-tight">
        {{ formatRupiah(totalSaved) }}
      </h4>
      <p class="text-[11px] text-emerald-600 font-semibold flex items-center gap-1 mt-0.5">
        <svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <polyline points="18 15 12 9 6 15"></polyline>
        </svg>
        <span>+17.5% dari bulan lalu</span>
      </p>
    </div>

    <!-- Semi-Circle Gauge Visualization (Matching ACRU Financial health meter) -->
    <div class="relative flex flex-col items-center justify-center my-1">
      <svg class="w-44 h-24 overflow-visible" viewBox="0 0 160 85">
        <!-- Background track (Gray) -->
        <path
          d="M 10 80 A 70 70 0 0 1 150 80"
          fill="none"
          stroke="#F1F5F9"
          stroke-width="16"
          stroke-linecap="round"
        />
        <!-- Active gradient progress path -->
        <path
          d="M 10 80 A 70 70 0 0 1 150 80"
          fill="none"
          stroke="url(#limeGradient)"
          stroke-width="16"
          stroke-linecap="round"
          :stroke-dasharray="circumference"
          :stroke-dashoffset="strokeOffset"
          class="transition-all duration-1000 ease-out"
        />
        <defs>
          <linearGradient id="limeGradient" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stop-color="#FACC15" />
            <stop offset="60%" stop-color="#84CC16" />
            <stop offset="100%" stop-color="#22C55E" />
          </linearGradient>
        </defs>
      </svg>

      <!-- Center percentage text -->
      <div class="absolute bottom-0 text-center">
        <span class="text-2xl font-black text-slate-900 tracking-tight leading-none block">
          {{ clampedScore }}%
        </span>
        <span class="text-[9px] text-slate-400 font-medium block mt-0.5">
          pendapatan tersimpan
        </span>
      </div>
    </div>

    <p class="text-[10px] text-slate-400 text-center mt-2 leading-relaxed">
      Berdasarkan metrik transaksi dan alokasi dompet dalam 30 hari terakhir.
    </p>
  </div>
</template>
