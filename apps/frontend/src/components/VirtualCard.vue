<script setup lang="ts">
import { computed } from 'vue'
import type { Wallet } from '../services/wallet'
import { useAuthStore } from '../stores/auth'

const props = defineProps<{
  wallets: Wallet[]
  selectedWalletId: string
}>()

const emit = defineEmits<{
  (e: 'select-wallet', id: string): void
  (e: 'open-wallet-modal'): void
  (e: 'open-tx-modal'): void
  (e: 'open-proposal-modal'): void
  (e: 'link-telegram'): void
}>()

const authStore = useAuthStore()
const isAdmin = computed(() => authStore.user?.role === 'admin')

const activeWallet = computed(() => {
  return props.wallets.find(w => w.id === props.selectedWalletId) || props.wallets[0] || null
})

const formatRupiah = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(val)
}
</script>

<template>
  <div class="card-neo p-5 flex flex-col gap-5">
    <!-- Header with quick actions title and + Add card button -->
    <div class="flex items-center justify-between">
      <div>
        <h3 class="font-bold text-slate-900 text-base">Kartu Dompet</h3>
        <p class="text-[11px] text-slate-400 font-medium">Akses cepat dompet keluarga</p>
      </div>
      <button 
        v-if="isAdmin" 
        @click="emit('open-wallet-modal')" 
        class="text-xs font-bold text-slate-700 hover:text-slate-900 bg-slate-100 hover:bg-slate-200 px-3 py-1.5 rounded-xl flex items-center gap-1.5 transition"
      >
        <span>+ Tambah</span>
      </button>
    </div>

    <!-- Virtual Cards Carousel / Stack (Matching ACRU's green debit card aesthetic) -->
    <div v-if="activeWallet" class="relative overflow-hidden group">
      <!-- Realistic Lime-Green Virtual Card -->
      <div class="w-full aspect-[1.58/1] rounded-3xl p-5 bg-gradient-to-br from-brand-400 via-brand-500 to-lime-600 text-slate-900 shadow-xl shadow-brand-500/20 relative flex flex-col justify-between overflow-hidden">
        <!-- Background subtle decorative pattern -->
        <div class="absolute -right-8 -bottom-8 w-40 h-40 bg-white/10 rounded-full blur-xl pointer-events-none"></div>
        <div class="absolute right-4 top-4 opacity-20 font-black text-4xl italic tracking-tighter select-none">
          ACIS
        </div>

        <!-- Top Card Info -->
        <div class="flex items-start justify-between relative z-10">
          <div>
            <span class="text-[10px] font-bold uppercase tracking-wider bg-slate-900/10 px-2 py-0.5 rounded-full backdrop-blur-sm text-slate-900">
              Virtual Envelope
            </span>
            <h4 class="font-black text-lg text-slate-900 tracking-tight mt-1 truncate max-w-[170px]">
              {{ activeWallet.name }}
            </h4>
          </div>

          <!-- EMV Chip & Contactless Icon -->
          <div class="flex items-center gap-1.5">
            <div class="w-8 h-6 rounded-md bg-amber-200/90 border border-amber-300/80 flex items-center justify-center shadow-inner">
              <div class="w-4 h-3 border border-amber-400/50 rounded-sm"></div>
            </div>
            <!-- Contactless SVG -->
            <svg class="w-4 h-4 text-slate-900/60" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M8.5 16.5a5 5 0 0 1 0-9"></path>
              <path d="M12 19a8.5 8.5 0 0 0 0-14"></path>
            </svg>
          </div>
        </div>

        <!-- Card Number / Masked ID & Balance -->
        <div class="relative z-10 my-1">
          <p class="text-[11px] font-semibold text-slate-900/70 tracking-widest font-mono">
            •••• •••• •••• {{ activeWallet.id.slice(0, 4).toUpperCase() }}
          </p>
          <div class="mt-1">
            <span class="text-[10px] uppercase font-bold text-slate-900/60 block">Saldo Berjalan</span>
            <span class="text-2xl font-extrabold text-slate-950 tracking-tight">
              {{ formatRupiah(activeWallet.current_balance) }}
            </span>
          </div>
        </div>

        <!-- Card Bottom (Holder name & Min limit) -->
        <div class="flex items-end justify-between relative z-10 text-[11px]">
          <div>
            <span class="text-[9px] uppercase font-bold text-slate-900/50 block">Limit Min</span>
            <span class="font-bold text-slate-900 font-mono">{{ formatRupiah(activeWallet.minimum_limit) }}</span>
          </div>
          <div class="text-right">
            <span class="text-[9px] uppercase font-bold text-slate-900/50 block">Status</span>
            <span 
              class="font-extrabold"
              :class="activeWallet.current_balance <= activeWallet.minimum_limit ? 'text-red-950' : 'text-slate-950'"
            >
              {{ activeWallet.current_balance <= activeWallet.minimum_limit ? '⚠️ Low Balance' : '● Normal' }}
            </span>
          </div>
        </div>
      </div>

      <!-- Multiple Wallets Switcher Dots/Pills -->
      <div v-if="wallets.length > 1" class="flex items-center justify-center gap-1.5 mt-3 overflow-x-auto py-1">
        <button
          v-for="w in wallets"
          :key="w.id"
          @click="emit('select-wallet', w.id)"
          class="px-2.5 py-1 rounded-full text-[10px] font-bold transition-all shrink-0"
          :class="w.id === (activeWallet?.id) 
            ? 'bg-slate-900 text-white shadow-sm' 
            : 'bg-slate-100 text-slate-500 hover:bg-slate-200'"
        >
          {{ w.name }}
        </button>
      </div>
    </div>

    <!-- Empty State for Cards -->
    <div v-else class="p-6 bg-slate-50 rounded-2xl text-center border border-dashed border-slate-200">
      <p class="text-xs text-slate-400 font-medium">Belum ada dompet terdaftar.</p>
      <button 
        v-if="isAdmin" 
        @click="emit('open-wallet-modal')" 
        class="mt-2 text-xs font-bold text-brand-600 hover:underline"
      >
        + Buat Dompet Sekarang
      </button>
    </div>

    <!-- Quick Action Icons (Matching ACRU bottom card actions: Top up, Send, Request, History, More) -->
    <div class="grid grid-cols-4 gap-2 pt-1 border-t border-slate-100">
      <!-- 1. Catat Transaksi -->
      <button 
        @click="emit('open-tx-modal')" 
        class="flex flex-col items-center gap-1.5 p-2 rounded-2xl hover:bg-slate-50 transition group"
      >
        <div class="w-10 h-10 rounded-2xl bg-brand-50 border border-brand-200 text-brand-700 flex items-center justify-center group-hover:scale-105 transition">
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2">
            <line x1="12" y1="5" x2="12" y2="19"></line>
            <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
        </div>
        <span class="text-[10px] font-bold text-slate-600 group-hover:text-slate-900">Transaksi</span>
      </button>

      <!-- 2. Ajukan Proposal -->
      <button 
        @click="emit('open-proposal-modal')" 
        class="flex flex-col items-center gap-1.5 p-2 rounded-2xl hover:bg-slate-50 transition group"
      >
        <div class="w-10 h-10 rounded-2xl bg-amber-50 border border-amber-200 text-amber-600 flex items-center justify-center group-hover:scale-105 transition">
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2">
            <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"></path>
            <rect x="8" y="2" width="8" height="4" rx="1" ry="1"></rect>
          </svg>
        </div>
        <span class="text-[10px] font-bold text-slate-600 group-hover:text-slate-900">Proposal</span>
      </button>

      <!-- 3. Tambah Dompet (Admin) -->
      <button 
        @click="isAdmin ? emit('open-wallet-modal') : null" 
        :class="isAdmin ? 'cursor-pointer hover:bg-slate-50' : 'opacity-40 cursor-not-allowed'"
        class="flex flex-col items-center gap-1.5 p-2 rounded-2xl transition group"
      >
        <div class="w-10 h-10 rounded-2xl bg-blue-50 border border-blue-200 text-blue-600 flex items-center justify-center group-hover:scale-105 transition">
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2">
            <rect x="2" y="5" width="20" height="14" rx="3"></rect>
            <line x1="2" y1="10" x2="22" y2="10"></line>
          </svg>
        </div>
        <span class="text-[10px] font-bold text-slate-600 group-hover:text-slate-900">Dompet</span>
      </button>

      <!-- 4. Telegram Bot -->
      <button 
        @click="emit('link-telegram')" 
        class="flex flex-col items-center gap-1.5 p-2 rounded-2xl hover:bg-slate-50 transition group"
      >
        <div class="w-10 h-10 rounded-2xl bg-sky-50 border border-sky-200 text-sky-600 flex items-center justify-center group-hover:scale-105 transition">
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2">
            <line x1="22" y1="2" x2="11" y2="13"></line>
            <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
          </svg>
        </div>
        <span class="text-[10px] font-bold text-slate-600 group-hover:text-slate-900">Bot Tel</span>
      </button>
    </div>
  </div>
</template>
