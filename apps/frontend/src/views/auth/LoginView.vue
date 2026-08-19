<template>
  <div class="flex min-h-screen items-center justify-center bg-slate-200 px-4 py-8">
    <div class="w-full max-w-md rounded-xl bg-white p-8 shadow-lg">
      <div class="mb-7 text-center">
        <p class="mb-2 text-xs font-semibold uppercase tracking-[0.2em] text-blue-600">ACIS</p>
        <h2 class="text-2xl font-bold text-slate-800">Masuk</h2>
        <p class="mt-2 text-sm text-slate-500">Masuk dengan kode OTP dari Telegram</p>
      </div>

      <form v-if="step === 1" class="space-y-4" @submit.prevent="handleRequestOtp">
        <div>
          <label class="mb-2 block text-sm font-medium text-slate-700" for="phone">Nomor Telepon</label>
          <input id="phone" v-model="phoneNumber" type="tel" inputmode="tel" autocomplete="tel" class="w-full rounded border border-slate-300 px-3 py-2 outline-none focus:border-blue-500" placeholder="081234567890" />
          <p class="mt-2 text-xs text-slate-500">Gunakan nomor yang terhubung ke akun ACIS.</p>
        </div>
        <p v-if="error" class="rounded bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
        <button type="submit" :disabled="loading" class="w-full rounded bg-blue-600 px-4 py-2 font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60">
          {{ loading ? 'Mengirim...' : 'Minta Kode OTP' }}
        </button>
      </form>

      <form v-else class="space-y-4" @submit.prevent="handleVerifyOtp">
        <div class="rounded bg-blue-50 px-4 py-3 text-sm text-blue-800">
          Kode OTP dikirim ke Telegram untuk <strong>{{ phoneNumber }}</strong>.
          <span v-if="directSent"> Periksa pesan dari bot ACIS.</span>
          <span v-else> Buka bot ACIS untuk mengambil kode Anda.</span>
        </div>
        <div v-if="isTestUser" class="rounded bg-amber-50 px-4 py-3 text-sm text-amber-800">
          Akun uji coba terdeteksi. Kode bypass: <strong>{{ testOtp }}</strong>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-slate-700" for="otp">Kode OTP</label>
          <input id="otp" v-model="otp" type="text" inputmode="numeric" autocomplete="one-time-code" maxlength="6" class="w-full rounded border border-slate-300 px-3 py-2 text-center text-xl tracking-[0.4em] outline-none focus:border-blue-500" placeholder="000000" />
        </div>
        <p v-if="error" class="rounded bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
        <button type="submit" :disabled="loading || otp.length !== 6" class="w-full rounded bg-blue-600 px-4 py-2 font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60">
          {{ loading ? 'Memverifikasi...' : 'Verifikasi & Masuk' }}
        </button>
        <button type="button" :disabled="loading" class="w-full text-sm text-slate-500 hover:text-slate-800" @click="reset">
          Ganti nomor telepon
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/useAuthStore'

const router = useRouter()
const authStore = useAuthStore()
const step = ref(1)
const phoneNumber = ref('')
const otp = ref('')
const loading = ref(false)
const error = ref('')
const directSent = ref(false)
const isTestUser = ref(false)
const testOtp = ref('')

function normalizedPhone() {
  const phone = phoneNumber.value.trim().replace(/[\s-]/g, '')
  return phone.startsWith('08') ? `+62${phone.slice(1)}` : phone.startsWith('628') ? `+${phone}` : phone
}

async function handleRequestOtp() {
  error.value = ''
  if (!/^(\+628|628|08)[0-9]{8,12}$/.test(phoneNumber.value.trim().replace(/[\s-]/g, ''))) {
    error.value = 'Masukkan nomor telepon Indonesia yang valid.'
    return
  }

  loading.value = true
  try {
    const data = await authStore.requestOtp(normalizedPhone())
    directSent.value = data.direct_sent
    isTestUser.value = Boolean(data.is_test_user)
    testOtp.value = data.test_otp || ''
    step.value = 2
  } catch (requestError) {
    error.value = requestError.message
  } finally {
    loading.value = false
  }
}

async function handleVerifyOtp() {
  error.value = ''
  loading.value = true
  try {
    await authStore.verifyOtp(normalizedPhone(), otp.value)
    router.push({ path: '/' })
  } catch (verifyError) {
    error.value = verifyError.message
  } finally {
    loading.value = false
  }
}

function reset() {
  step.value = 1
  otp.value = ''
  error.value = ''
}
</script>
