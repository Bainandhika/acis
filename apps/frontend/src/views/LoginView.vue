<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
import { useI18n } from '../locales'

const authStore = useAuthStore()
const router = useRouter()
const { t, locale, setLocale } = useI18n()

// Active Mode: 'login' (Sign In) or 'register' (Sign Up)
const authMode = ref<'login' | 'register'>('login')
const step = ref(1)

const username = ref('')
const phoneNumber = ref('')
const otp = ref('')
const isTestUser = ref(false)
const loading = ref(false)
const error = ref('')

// Indonesian phone format validation (+628... or 08...)
const isValidPhone = computed(() => {
  const cleaned = phoneNumber.value.trim().replace(/[\s-]/g, '')
  return /^(\+628|08|628)[0-9]{8,12}$/.test(cleaned)
})

const handleSwitchMode = (mode: 'login' | 'register') => {
  authMode.value = mode
  error.value = ''
}

const handleRequestOTP = async () => {
  const cleanedPhone = phoneNumber.value.trim().replace(/[\s-]/g, '')
  if (!isValidPhone.value) {
    error.value = t('login.errors.invalidPhone')
    return
  }

  if (authMode.value === 'register' && !username.value.trim()) {
    error.value = t('login.errors.usernameRequired')
    return
  }

  loading.value = true
  error.value = ''
  
  try {
    const res = await authStore.requestOTP(
      cleanedPhone, 
      authMode.value === 'register' ? username.value.trim() : undefined,
      authMode.value
    )
    isTestUser.value = !!res.is_test_user
    if (res.is_test_user && res.test_otp) {
      otp.value = res.test_otp
    }
    step.value = 2
  } catch (err: any) {
    const errMsg = err.response?.data?.error
    if (err.response?.status === 409) {
      error.value = t('login.errors.duplicatePhone')
    } else if (err.response?.status === 404) {
      error.value = t('login.errors.unregisteredPhone')
    } else {
      error.value = errMsg || t('login.errors.requestFailed')
    }
  } finally {
    loading.value = false
  }
}

const handleVerifyOTP = async () => {
  if (otp.value.length !== 6) return
  loading.value = true
  error.value = ''
  
  try {
    const cleanedPhone = phoneNumber.value.trim().replace(/[\s-]/g, '')
    await authStore.verifyOTP(
      cleanedPhone, 
      otp.value.trim(), 
      authMode.value === 'register' ? username.value.trim() : undefined,
      authMode.value
    )
    router.push('/')
  } catch (err: any) {
    error.value = err.response?.data?.error || t('login.errors.invalidOtp')
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
      <!-- Top Bar with Language Selector -->
      <div class="flex justify-end mb-4">
        <div class="inline-flex p-1 bg-slate-100/90 rounded-xl border border-slate-200/70 text-[11px] font-bold">
          <button 
            @click="setLocale('en')"
            class="px-2.5 py-1 rounded-lg transition-all"
            :class="locale === 'en' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-800'"
            type="button"
          >
            🇬🇧 EN
          </button>
          <button 
            @click="setLocale('id')"
            class="px-2.5 py-1 rounded-lg transition-all"
            :class="locale === 'id' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-800'"
            type="button"
          >
            🇮🇩 ID
          </button>
        </div>
      </div>

      <!-- Logo Header -->
      <div class="flex flex-col items-center text-center mb-6">
        <div class="w-14 h-14 rounded-3xl bg-gradient-to-tr from-brand-500 to-lime-300 flex items-center justify-center shadow-lg shadow-brand-500/30 text-white font-black text-2xl mb-4">
          A
        </div>
        <h2 class="text-2xl font-black text-slate-900 tracking-tight">
          {{ t('login.title') }}
        </h2>
        <p class="text-xs text-slate-400 font-medium mt-1">
          {{ t('login.subtitle') }}
        </p>
      </div>

      <!-- Sign In / Sign Up Segmented Tabs (Step 1 only) -->
      <div v-if="step === 1" class="flex p-1 bg-slate-100/90 rounded-2xl mb-6 text-xs font-black border border-slate-200/60">
        <button 
          class="flex-1 py-2.5 rounded-xl transition-all"
          :class="authMode === 'login' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'"
          @click="handleSwitchMode('login')"
          type="button"
        >
          {{ t('login.tabSignIn') }}
        </button>
        <button 
          class="flex-1 py-2.5 rounded-xl transition-all"
          :class="authMode === 'register' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'"
          @click="handleSwitchMode('register')"
          type="button"
        >
          {{ t('login.tabSignUp') }}
        </button>
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

      <!-- STEP 1: Phone & (Optional Username on Sign-Up) Inputs -->
      <div v-if="step === 1" class="flex flex-col gap-4">
        <!-- Username input shown only in Sign-Up mode -->
        <div v-if="authMode === 'register'">
          <label class="text-xs font-bold text-slate-700 block mb-1.5">{{ t('login.usernameLabel') }}</label>
          <div class="relative">
            <span class="absolute inset-y-0 left-0 flex items-center pl-3.5 pointer-events-none text-slate-400">
              <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                <circle cx="12" cy="7" r="4"></circle>
              </svg>
            </span>
            <input 
              type="text" 
              v-model="username" 
              :placeholder="t('login.usernamePlaceholder')" 
              class="w-full pl-10 pr-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-semibold text-slate-900 focus:outline-none focus:ring-2 focus:ring-brand-400 focus:bg-white transition" 
              @keyup.enter="phoneNumber && handleRequestOTP()"
            />
          </div>
          <p class="text-[10px] text-slate-400 mt-1.5">{{ t('login.usernameHint') }}</p>
        </div>

        <!-- Phone Number input (Required for both Sign-In and Sign-Up) -->
        <div>
          <label class="text-xs font-bold text-slate-700 block mb-1.5">{{ t('login.phoneLabel') }}</label>
          <div class="relative">
            <span class="absolute inset-y-0 left-0 flex items-center pl-3.5 pointer-events-none text-slate-400">
              <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/>
              </svg>
            </span>
            <input 
              type="tel" 
              v-model="phoneNumber" 
              :placeholder="t('login.phonePlaceholder')" 
              class="w-full pl-10 pr-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-semibold text-slate-900 focus:outline-none focus:ring-2 focus:ring-brand-400 focus:bg-white transition" 
              @keyup.enter="phoneNumber && handleRequestOTP()"
            />
          </div>
          <p class="text-[10px] text-slate-400 mt-1.5">{{ t('login.phoneHint') }}</p>
        </div>

        <button 
          class="w-full py-3.5 mt-2 rounded-2xl bg-slate-900 hover:bg-slate-800 text-white font-extrabold text-xs transition shadow-md active:scale-95 disabled:opacity-50 disabled:pointer-events-none flex items-center justify-center gap-2"
          :disabled="loading || !phoneNumber.trim() || (authMode === 'register' && !username.trim())"
          @click="handleRequestOTP"
        >
          <svg class="w-4 h-4 text-brand-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="22" y1="2" x2="11" y2="13"></line>
            <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
          </svg>
          <span>{{ loading ? t('login.sendingOtp') : (authMode === 'register' ? t('login.registerAndRequestOtp') : t('login.requestOtp')) }}</span>
        </button>
      </div>

      <!-- STEP 2: Input 6-Digit OTP -->
      <div v-else class="flex flex-col gap-4">
        <div class="p-3.5 bg-emerald-50 border border-emerald-200 rounded-2xl text-center">
          <p class="text-xs text-emerald-800 font-bold">
            {{ t('login.otpSentTitle') }}
          </p>
          <p class="text-[11px] text-emerald-700 mt-0.5">
            {{ t('login.otpSentSubtitle') }} <span class="font-bold font-mono">{{ phoneNumber }}</span>
          </p>
          <div v-if="isTestUser" class="mt-2 text-[11px] text-brand-700 bg-brand-100/80 px-2.5 py-1 rounded-xl font-mono font-bold inline-block">
            {{ t('login.testBypassHint') }}
          </div>
        </div>

        <div>
          <label class="text-xs font-bold text-slate-700 block mb-1.5 text-center">{{ t('login.enterOtpLabel') }}</label>
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
            {{ t('login.back') }}
          </button>
          <button 
            class="flex-1 py-3.5 rounded-2xl bg-slate-900 hover:bg-slate-800 text-white font-extrabold text-xs transition shadow-md active:scale-95 disabled:opacity-50"
            :disabled="loading || otp.length !== 6"
            @click="handleVerifyOTP"
          >
            {{ loading ? t('login.verifying') : t('login.verifyAndSignIn') }}
          </button>
        </div>
      </div>

      <!-- Footer Note -->
      <p class="text-[10px] text-slate-400 text-center mt-8">
        {{ t('login.footerSecurity') }}
      </p>
    </div>
  </div>
</template>