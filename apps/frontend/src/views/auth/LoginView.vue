<template>
    <div class="flex min-h-screen items-center justify-center bg-slate-200 px-4 py-8">
        <div
            class="w-full max-w-[560px] rounded-[20px] bg-white p-8 shadow-[0_10px_30px_rgba(15,23,42,0.12)] ring-1 ring-slate-200">
            <div class="mb-7 text-center">
                <p class="mb-2 text-[13px] font-semibold uppercase tracking-[0.22em] text-blue-600">ACIS</p>
                <h2 class="text-[42px] font-black leading-none text-slate-800">{{ authMode === 'login' ? 'Masuk' :
                    'Daftar' }}</h2>
            </div>

            <form v-if="step === 1" class="space-y-4" @submit.prevent="handleRequestOtp">
                <div v-if="authMode === 'register'" class="space-y-2">
                    <label class="block text-[20px] font-bold text-slate-800" for="name">Nama</label>
                    <input id="name" v-model="name" type="text" autocomplete="name"
                        class="w-full rounded-[10px] border border-slate-300 bg-white px-3 py-3 text-[18px] text-slate-800 outline-none transition focus:border-blue-500"
                        placeholder="Nama lengkap" />
                </div>
                <div class="space-y-2">
                    <label class="block text-[20px] font-bold text-slate-800" for="phone">Nomor Telepon</label>
                    <input id="phone" v-model="phoneNumber" type="tel" inputmode="tel" autocomplete="tel"
                        class="w-full rounded-[10px] border border-blue-400 bg-white px-3 py-3 text-[18px] text-slate-800 outline-none transition focus:border-blue-500"
                        placeholder="081234567890" />
                </div>

                <p v-if="error"
                    class="rounded-[8px] bg-red-100 px-4 py-3 text-[18px] font-medium leading-relaxed text-red-700">
                    {{ error }}
                </p>

                <button type="submit" :disabled="loading"
                    class="mt-2 w-full rounded-[10px] px-4 py-4 text-[22px] font-bold text-white transition disabled:cursor-not-allowed disabled:opacity-60"
                    :class="authMode === 'register' ? 'bg-emerald-600 hover:bg-emerald-700' : 'bg-blue-600 hover:bg-blue-700'">
                    {{ loading ? 'Mengirim...' : 'Minta Kode OTP' }}
                </button>

                <div v-if="authMode === 'login'" class="pt-1 text-center text-[18px] text-slate-600">
                    Belum punya akun?
                    <button type="button" class="font-semibold text-blue-600 hover:text-blue-700"
                        @click="switchMode('register')">
                        Daftar
                    </button>
                </div>
                <div v-else class="pt-1 text-center text-[18px] text-slate-600">
                    Sudah punya akun?
                    <button type="button" class="font-semibold text-blue-600 hover:text-blue-700"
                        @click="switchMode('login')">
                        Masuk
                    </button>
                </div>
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
                    <input id="otp" v-model="otp" type="text" inputmode="numeric" autocomplete="one-time-code"
                        maxlength="6"
                        class="w-full rounded border border-slate-300 px-3 py-2 text-center text-xl tracking-[0.4em] outline-none focus:border-blue-500"
                        placeholder="000000" />
                </div>
                <p v-if="error" class="rounded bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
                <button type="submit" :disabled="loading || otp.length !== 6"
                    class="w-full rounded bg-blue-600 px-4 py-2 font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60">
                    {{ loading ? 'Memverifikasi...' : authMode === 'register' ? 'Verifikasi & Daftar' : 'Verifikasi &
                    Masuk' }}
                </button>
                <button type="button" :disabled="loading" class="w-full text-sm text-slate-500 hover:text-slate-800"
                    @click="reset">
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
const authMode = ref('login')
const step = ref(1)
const name = ref('')
const phoneNumber = ref('')
const otp = ref('')
const loading = ref(false)
const error = ref('')
const directSent = ref(false)
const isTestUser = ref(false)
const testOtp = ref('')

function normalizedPhone() {
    const phone = phoneNumber.value.trim().replace(/[\s-]/g, '')

    if (phone.startsWith('08')) {
        return `+62${phone.slice(1)}`
    }

    if (phone.startsWith('628')) {
        return `+${phone}`
    }

    return phone
}

function switchMode(mode) {
    authMode.value = mode
    step.value = 1
    otp.value = ''
    error.value = ''
    if (mode === 'login') {
        name.value = ''
    }
}

async function handleRequestOtp() {
    error.value = ''

    if (authMode.value === 'register' && !name.value.trim()) {
        error.value = 'Nama wajib diisi untuk pendaftaran.'
        return
    }

    if (!/^(\+628|628|08)\d{8,12}$/.test(phoneNumber.value.trim().replace(/[\s-]/g, ''))) {
        error.value = 'Masukkan nomor telepon Indonesia yang valid.'
        return
    }

    loading.value = true
    try {
        const payload = {
            action: authMode.value,
            username: authMode.value === 'register' ? name.value.trim() : undefined
        }

        const data = await authStore.requestOtp(normalizedPhone(), payload)
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
        const payload = {
            action: authMode.value,
            username: authMode.value === 'register' ? name.value.trim() : undefined
        }

        await authStore.verifyOtp(normalizedPhone(), otp.value, payload)
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
