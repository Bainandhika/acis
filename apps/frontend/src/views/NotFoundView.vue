<script setup lang="ts">
import { useRouter } from 'vue-router'
import Button from '../components/ui/Button.vue'

const router = useRouter()
</script>

<template>
  <div class="min-h-screen bg-slate-950 text-slate-100 flex flex-col items-center justify-center p-6 text-center">
    <div class="w-20 h-20 rounded-3xl bg-slate-900 border border-slate-800 flex items-center justify-center text-4xl mb-6 shadow-2xl">
      🧭
    </div>
    <h1 class="text-4xl sm:text-5xl font-black text-white tracking-tight mb-2">404</h1>
    <h2 class="text-lg font-bold text-slate-300 mb-2">Halaman Tidak Ditemukan</h2>
    <p class="text-xs text-slate-500 max-w-sm mb-8">
      Halaman yang Anda cari tidak tersedia atau telah dipindahkan ke alamat lain.
    </p>
    <Button
      variant="primary"
      size="md"
      @click="router.push('/')"
    >
      Kembali ke Dashboard
    </Button>
  </div>
</template>
