<script setup lang="ts">
import { useAuthStore } from '../stores/auth'
import { useFamilyStore } from '../stores/family'
import { useRouter } from 'vue-router'

const props = defineProps<{
  searchQuery?: string
}>()

const emit = defineEmits<{
  (e: 'update:searchQuery', val: string): void
  (e: 'open-tx-modal'): void
  (e: 'open-wallet-modal'): void
  (e: 'open-proposal-modal'): void
  (e: 'toggle-sidebar'): void
}>()

const authStore = useAuthStore()
const familyStore = useFamilyStore()
const router = useRouter()

const handleLogout = () => {
  if (confirm('Apakah Anda yakin ingin keluar?')) {
    authStore.logout()
    router.push('/login')
  }
}
</script>

<template>
  <header class="h-20 bg-white/80 backdrop-blur-md border-b border-slate-100/90 sticky top-0 z-20 px-6 flex items-center justify-between gap-4">
    <!-- Left: Mobile Menu Toggle & Search Bar -->
    <div class="flex items-center gap-4 flex-1 max-w-md">
      <!-- Mobile hamburger -->
      <button 
        @click="emit('toggle-sidebar')" 
        class="lg:hidden p-2 rounded-xl text-slate-500 hover:bg-slate-100 transition"
      >
        <svg class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="3" y1="12" x2="21" y2="12"></line>
          <line x1="3" y1="6" x2="21" y2="6"></line>
          <line x1="3" y1="18" x2="21" y2="18"></line>
        </svg>
      </button>

      <!-- Search Input (matching ACRU Quick search) -->
      <div class="relative w-full">
        <span class="absolute inset-y-0 left-0 flex items-center pl-3.5 pointer-events-none text-slate-400">
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2">
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
        </span>
        <input
          type="text"
          :value="searchQuery"
          @input="emit('update:searchQuery', ($event.target as HTMLInputElement).value)"
          placeholder="Cari transaksi, dompet, kategori..."
          class="w-full pl-10 pr-4 py-2.5 bg-slate-50 border border-slate-200/80 rounded-2xl text-xs font-medium text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-brand-400 focus:bg-white transition-all shadow-inner"
        />
      </div>
    </div>

    <!-- Right: Telegram status pill, Notification bell, User info, CTA Button -->
    <div class="flex items-center gap-3.5">
      <!-- Telegram Status Indicator Pill -->
      <div 
        class="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-semibold border transition"
        :class="familyStore.family?.telegram_chat_id 
          ? 'bg-emerald-50 border-emerald-200 text-emerald-700' 
          : 'bg-amber-50 border-amber-200 text-amber-700'"
      >
        <span 
          class="w-2 h-2 rounded-full"
          :class="familyStore.family?.telegram_chat_id ? 'bg-emerald-500 animate-pulse' : 'bg-amber-500'"
        ></span>
        <span class="text-[11px]">
          Bot {{ familyStore.family?.telegram_chat_id ? 'Terhubung' : 'Belum Terhubung' }}
        </span>
      </div>

      <!-- Notification Bell -->
      <div class="relative">
        <button class="w-10 h-10 rounded-2xl bg-slate-50 border border-slate-200/70 hover:bg-slate-100 flex items-center justify-center text-slate-600 transition">
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
            <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
          </svg>
        </button>
        <span class="absolute top-2 right-2 w-2 h-2 bg-brand-500 rounded-full ring-2 ring-white"></span>
      </div>

      <!-- User Profile Pill -->
      <div class="flex items-center gap-3 pl-2 py-1 pr-2 rounded-2xl bg-slate-50 border border-slate-200/70">
        <!-- User Avatar -->
        <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-slate-800 to-slate-600 text-white font-bold text-xs flex items-center justify-center shrink-0">
          {{ authStore.user?.name ? authStore.user.name.charAt(0).toUpperCase() : 'U' }}
        </div>
        <div class="hidden md:flex flex-col text-left">
          <div class="flex items-center gap-1.5 leading-tight">
            <span class="text-xs font-bold text-slate-900 truncate max-w-[100px]">{{ authStore.user?.name || 'Pengguna' }}</span>
            <span class="text-[9px] font-extrabold uppercase px-1 py-0.2 rounded bg-brand-200 text-brand-900">{{ authStore.user?.role }}</span>
          </div>
          <span class="text-[10px] text-slate-400 truncate max-w-[110px]">{{ authStore.user?.email }}</span>
        </div>
        
        <!-- Logout button -->
        <button 
          @click="handleLogout" 
          class="text-slate-400 hover:text-red-500 p-1 rounded-lg transition"
          title="Keluar"
        >
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
            <polyline points="16 17 21 12 16 7"></polyline>
            <line x1="21" y1="12" x2="9" y2="12"></line>
          </svg>
        </button>
      </div>

      <!-- Quick Action CTA Button (matching ACRU top right button) -->
      <button 
        @click="emit('open-tx-modal')" 
        class="hidden sm:flex items-center gap-2 px-4 py-2.5 rounded-2xl bg-slate-900 hover:bg-slate-800 text-white text-xs font-bold transition shadow-sm hover:shadow active:scale-95"
      >
        <svg class="w-4 h-4 text-brand-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <line x1="12" y1="5" x2="12" y2="19"></line>
          <line x1="5" y1="12" x2="19" y2="12"></line>
        </svg>
        <span>Catat Transaksi</span>
      </button>
    </div>
  </header>
</template>
