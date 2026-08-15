<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const step = ref(1)
const email = ref('')
const otp = ref('')
const loading = ref(false)
const error = ref('')

const handleRequestOTP = async () => {
  loading.value = true
  error.value = ''
  
  try {
    await authStore.requestOTP(email.value)
    step.value = 2
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Gagal mengirim OTP. Coba lagi.'
  } finally {
    loading.value = false
  }
}

const handleVerifyOTP = async () => {
  loading.value = true
  error.value = ''
  
  try {
    await authStore.verifyOTP(email.value, otp.value)
    router.push('/')
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Kode OTP salah atau expired.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex items-center justify-center min-h-screen bg-[#F8FAFC] p-4 relative overflow-hidden">
    <!-- Ambient glow decorative circles matching modern fintech UI -->
    <div class="absolute top-1/4 left-1/3 w-96 h-96 bg-brand-200/40 rounded-full blur-3xl pointer-events-none"></div>
    <div class="absolute bottom-1/4 right-1/3 w-80 h-80 bg-lime-200/30 rounded-full blur-3xl pointer-events-none"></div>

    <div class="card-neo w-full max-w-md p-8 relative z-10 border border-slate-200/80 shadow-2xl bg-white/95 backdrop-blur-xl">
      <!-- Logo Header -->
      <div class="flex flex-col items-center text-center mb-8">
        <div class="w-14 h-14 rounded-3xl bg-gradient-to-tr from-brand-500 to-lime-300 flex items-center justify-center shadow-lg shadow-brand-500/30 text-white font-black text-2xl mb-4">
          A
        </div>
        <h2 class="text-2xl font-black text-slate-900 tracking-tight">
          Masuk ke ACIS
        </h2>
        <p class="text-xs text-slate-400 font-medium mt-1">
          Sistem manajemen dan alokasi finansial cerdas keluarga
        </p>
      </div>
      
      <!-- Error Alert -->
      <div v-if="error" class="p-3.5 bg-rose-50 border border-rose-200 text-rose-700 text-xs rounded-2xl mb-6 font-semibold flex items-center gap-2">
        <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="8" x2="12" y2="12"></line>
          <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
        <span>{{ error }}</span>
      </div>

      <!-- STEP 1: Input Email -->
      <div v-if="step === 1" class="flex flex-col gap-4">
        <div>
          <label class="text-xs font-bold text-slate-700 block mb-1.5">Alamat Email Keluarga</label>
          <div class="relative">
            <span class="absolute inset-y-0 left-0 flex items-center pl-3.5 pointer-events-none text-slate-400">
              <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="2" y="4" width="20" height="16" rx="2"></rect>
                <path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"></path>
              </svg>
            </span>
            <input 
              type="email" 
              v-model="email" 
              placeholder="nama@keluarga.com" 
              class="w-full pl-10 pr-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-semibold text-slate-900 focus:outline-none focus:ring-2 focus:ring-brand-400 focus:bg-white transition" 
              @keyup.enter="email && handleRequestOTP()"
            />
          </div>
        </div>

        <button 
          class="w-full py-3.5 mt-2 rounded-2xl bg-slate-900 hover:bg-slate-800 text-white font-extrabold text-xs transition shadow-md active:scale-95 disabled:opacity-50 disabled:pointer-events-none"
          :disabled="loading || !email"
          @click="handleRequestOTP"
        >
          {{ loading ? 'Mengirim OTP...' : 'Lanjutkan dengan Email' }}
        </button>
      </div>

      <!-- STEP 2: Input OTP -->
      <div v-else class="flex flex-col gap-4">
        <div class="p-3 bg-slate-50 border border-slate-200 rounded-2xl text-center">
          <p class="text-[11px] text-slate-500 font-medium">
            Kode OTP 6 digit telah dikirim ke <span class="font-bold text-slate-900">{{ email }}</span>
          </p>
        </div>

        <div>
          <label class="text-xs font-bold text-slate-700 block mb-1.5 text-center">Kode Verifikasi</label>
          <input 
            type="text" 
            v-model="otp" 
            placeholder="••••••" 
            maxlength="6"
            class="w-full py-3.5 bg-slate-50 border border-slate-200 rounded-2xl text-center text-2xl font-black tracking-[0.4em] text-slate-900 focus:outline-none focus:ring-2 focus:ring-brand-400 focus:bg-white transition font-mono" 
            @keyup.enter="otp.length === 6 && handleVerifyOTP()"
          />
        </div>

        <div class="flex gap-2.5 mt-2">
          <button 
            class="px-4 py-3.5 rounded-2xl bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs transition"
            @click="step = 1"
            :disabled="loading"
          >
            Kembali
          </button>
          <button 
            class="flex-1 py-3.5 rounded-2xl bg-slate-900 hover:bg-slate-800 text-white font-extrabold text-xs transition shadow-md active:scale-95 disabled:opacity-50"
            :disabled="loading || otp.length !== 6"
            @click="handleVerifyOTP"
          >
            {{ loading ? 'Memverifikasi...' : 'Verifikasi & Masuk' }}
          </button>
        </div>
      </div>

      <!-- Footer Note -->
      <p class="text-[10px] text-slate-400 text-center mt-8">
        Dilindungi enkripsi end-to-end ACIS Security.
      </p>
    </div>
  </div>
</template>