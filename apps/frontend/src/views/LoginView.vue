<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const step = ref(1)
const email = ref('')
const telegramIdentifier = ref('')
const otp = ref('')
const isTestUser = ref(false)
const loading = ref(false)
const error = ref('')

const testUsers = [
  { label: 'Admin User', email: 'admin@acis.test', identifier: '100000001' },
  { label: 'Sarah (Member)', email: 'member1@acis.test', identifier: '100000002' },
  { label: 'Alex (Member)', email: 'member2@acis.test', identifier: '100000003' },
]

const fillTestUser = (u: { email: string, identifier: string }) => {
  email.value = u.email
  telegramIdentifier.value = u.identifier
  error.value = ''
}

const handleRequestOTP = async () => {
  if (!email.value.trim() || !telegramIdentifier.value.trim()) {
    error.value = 'Please enter a valid email address and Telegram identifier.'
    return
  }

  loading.value = true
  error.value = ''
  
  try {
    const res = await authStore.requestOTP(email.value.trim().toLowerCase(), telegramIdentifier.value.trim())
    isTestUser.value = !!res.is_test_user
    if (res.is_test_user && res.test_otp) {
      otp.value = res.test_otp
    }
    step.value = 2
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to request Telegram OTP. Please check your credentials.'
  } finally {
    loading.value = false
  }
}

const handleVerifyOTP = async () => {
  if (otp.value.length !== 6) return
  loading.value = true
  error.value = ''
  
  try {
    await authStore.verifyOTP(email.value.trim().toLowerCase(), telegramIdentifier.value.trim(), otp.value.trim())
    router.push('/')
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Invalid or expired OTP.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex items-center justify-center min-h-screen bg-[#F8FAFC] p-4 relative overflow-hidden">
    <!-- Ambient glow decorative backgrounds -->
    <div class="absolute top-1/4 left-1/3 w-96 h-96 bg-brand-200/40 rounded-full blur-3xl pointer-events-none"></div>
    <div class="absolute bottom-1/4 right-1/3 w-80 h-80 bg-lime-200/30 rounded-full blur-3xl pointer-events-none"></div>

    <div class="card-neo w-full max-w-md p-8 relative z-10 border border-slate-200/80 shadow-2xl bg-white/95 backdrop-blur-xl">
      <!-- Logo Header -->
      <div class="flex flex-col items-center text-center mb-6">
        <div class="w-14 h-14 rounded-3xl bg-gradient-to-tr from-brand-500 to-lime-300 flex items-center justify-center shadow-lg shadow-brand-500/30 text-white font-black text-2xl mb-4">
          A
        </div>
        <h2 class="text-2xl font-black text-slate-900 tracking-tight">
          Sign In to ACIS
        </h2>
        <p class="text-xs text-slate-400 font-medium mt-1">
          Smart Family Finance Management &amp; Automated Telegram OTP
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

      <!-- Quick Test User Seeds & Bypass Selector -->
      <div class="mb-5 p-3 rounded-2xl bg-slate-50 border border-slate-200 text-xs">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[10px] font-extrabold uppercase text-slate-400 tracking-wider">Test Users (OTP Bypass: 123456)</span>
        </div>
        <div class="grid grid-cols-3 gap-1.5">
          <button 
            v-for="u in testUsers" 
            :key="u.email"
            @click="fillTestUser(u)"
            class="py-1.5 px-2 rounded-xl bg-white hover:bg-brand-50 hover:text-brand-800 border border-slate-200 text-[11px] font-bold text-slate-700 transition text-center shadow-sm"
          >
            {{ u.label }}
          </button>
        </div>
      </div>

      <!-- STEP 1: Input Email & Telegram Identifier -->
      <div v-if="step === 1" class="flex flex-col gap-4">
        <div>
          <label class="text-xs font-bold text-slate-700 block mb-1.5">Email Address</label>
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
              placeholder="user@family.com" 
              class="w-full pl-10 pr-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-semibold text-slate-900 focus:outline-none focus:ring-2 focus:ring-brand-400 focus:bg-white transition" 
              @keyup.enter="telegramIdentifier && email && handleRequestOTP()"
            />
          </div>
        </div>

        <div>
          <label class="text-xs font-bold text-slate-700 block mb-1.5">Telegram Identifier</label>
          <div class="relative">
            <span class="absolute inset-y-0 left-0 flex items-center pl-3.5 pointer-events-none text-slate-400">
              <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69a.2.2 0 00-.05-.18c-.06-.05-.14-.03-.21-.02-.09.02-1.49.95-4.22 2.79-.4.27-.76.41-1.08.4-.36-.01-1.04-.2-1.55-.37-.63-.2-1.12-.31-1.08-.66.02-.18.27-.36.75-.55 2.92-1.27 4.86-2.11 5.83-2.51 2.78-1.16 3.35-1.36 3.73-1.36.08 0 .27.02.39.12.1.08.13.19.14.27-.01.06.01.24 0 .38z"/>
              </svg>
            </span>
            <input 
              type="text" 
              v-model="telegramIdentifier" 
              placeholder="Telegram Chat ID (e.g. 100000001) or @username" 
              class="w-full pl-10 pr-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-semibold text-slate-900 focus:outline-none focus:ring-2 focus:ring-brand-400 focus:bg-white transition" 
              @keyup.enter="email && telegramIdentifier && handleRequestOTP()"
            />
          </div>
          <p class="text-[10px] text-slate-400 mt-1">OTP is sent automatically to your Telegram without needing to run /start.</p>
        </div>

        <button 
          class="w-full py-3.5 mt-2 rounded-2xl bg-slate-900 hover:bg-slate-800 text-white font-extrabold text-xs transition shadow-md active:scale-95 disabled:opacity-50 disabled:pointer-events-none flex items-center justify-center gap-2"
          :disabled="loading || !email || !telegramIdentifier"
          @click="handleRequestOTP"
        >
          <svg class="w-4 h-4 text-brand-300" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69a.2.2 0 00-.05-.18c-.06-.05-.14-.03-.21-.02-.09.02-1.49.95-4.22 2.79-.4.27-.76.41-1.08.4-.36-.01-1.04-.2-1.55-.37-.63-.2-1.12-.31-1.08-.66.02-.18.27-.36.75-.55 2.92-1.27 4.86-2.11 5.83-2.51 2.78-1.16 3.35-1.36 3.73-1.36.08 0 .27.02.39.12.1.08.13.19.14.27-.01.06.01.24 0 .38z"/>
          </svg>
          <span>{{ loading ? 'Sending OTP...' : 'Request Telegram OTP' }}</span>
        </button>
      </div>

      <!-- STEP 2: Input Telegram OTP -->
      <div v-else class="flex flex-col gap-4">
        <div class="p-3.5 bg-emerald-50 border border-emerald-200 rounded-2xl text-center">
          <p class="text-xs text-emerald-800 font-bold">
            OTP Sent Automatically!
          </p>
          <p class="text-[11px] text-emerald-700 mt-0.5">
            Verification code sent for <span class="font-bold">{{ email }}</span> ({{ telegramIdentifier }})
          </p>
          <div v-if="isTestUser" class="mt-2 text-[11px] text-brand-700 bg-brand-100/80 px-2.5 py-1 rounded-xl font-mono font-bold inline-block">
            Test Bypass OTP: 123456
          </div>
        </div>

        <div>
          <label class="text-xs font-bold text-slate-700 block mb-1.5 text-center">Enter 6-Digit Telegram OTP</label>
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
            Back
          </button>
          <button 
            class="flex-1 py-3.5 rounded-2xl bg-slate-900 hover:bg-slate-800 text-white font-extrabold text-xs transition shadow-md active:scale-95 disabled:opacity-50"
            :disabled="loading || otp.length !== 6"
            @click="handleVerifyOTP"
          >
            {{ loading ? 'Verifying...' : 'Verify & Sign In' }}
          </button>
        </div>
      </div>

      <!-- Footer Note -->
      <p class="text-[10px] text-slate-400 text-center mt-8">
        Protected with encrypted end-to-end security &amp; ACIS Bot Token.
      </p>
    </div>
  </div>
</template>