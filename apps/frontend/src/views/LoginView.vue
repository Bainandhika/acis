<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

// State Management (Mirip variabel pointer di Go)
const step = ref(1)
const email = ref('')
const otp = ref('')
const loading = ref(false)
const error = ref('')

// Action: Request OTP
const handleRequestOTP = async () => {
  loading.value = true
  error.value = ''
  
  try {
    await authStore.requestOTP(email.value)
    step.value = 2 // Pindah ke step input OTP kalau sukses
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Gagal mengirim OTP. Coba lagi.'
  } finally {
    loading.value = false
  }
}

// Action: Verify OTP
const handleVerifyOTP = async () => {
  loading.value = true
  error.value = ''
  
  try {
    await authStore.verifyOTP(email.value, otp.value)
    router.push('/') // Redirect ke Dashboard kalau sukses
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Kode OTP salah atau expired.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex items-center justify-center min-h-screen">
    <div class="card w-96 bg-base-100 shadow-xl">
      <div class="card-body">
        <h2 class="card-title justify-center text-2xl font-bold mb-4">
          Login ACIS 💰
        </h2>
        
        <!-- Error Alert -->
        <div v-if="error" class="alert alert-error shadow-lg mb-4">
          <span>{{ error }}</span>
        </div>

        <!-- STEP 1: Input Email -->
        <div v-if="step === 1">
          <div class="form-control w-full">
            <label class="label">
              <span class="label-text">Email Keluarga</span>
            </label>
            <!-- v-model: Two-way binding (mirip pointer di Go) -->
            <input 
              type="email" 
              v-model="email" 
              placeholder="bagas@acis.app" 
              class="input input-bordered w-full" 
            />
          </div>
          <div class="card-actions justify-end mt-6">
            <button 
              class="btn btn-primary w-full" 
              :class="{ 'loading': loading }"
              :disabled="loading || !email"
              @click="handleRequestOTP"
            >
              {{ loading ? 'Mengirim...' : 'Kirim OTP' }}
            </button>
          </div>
        </div>

        <!-- STEP 2: Input OTP -->
        <div v-else>
          <p class="text-sm text-center mb-4 text-base-content/70">
            Kode OTP dikirim ke <b>{{ email }}</b>
          </p>
          <div class="form-control w-full">
            <label class="label">
              <span class="label-text">Masukkan 6 Digit Kode</span>
            </label>
            <input 
              type="text" 
              v-model="otp" 
              placeholder="123456" 
              maxlength="6"
              class="input input-bordered w-full text-center text-xl tracking-widest" 
            />
          </div>
          <div class="card-actions justify-end mt-6 gap-2">
            <button 
              class="btn btn-ghost" 
              @click="step = 1"
              :disabled="loading"
            >
              Ganti Email
            </button>
            <button 
              class="btn btn-primary flex-1" 
              :class="{ 'loading': loading }"
              :disabled="loading || otp.length !== 6"
              @click="handleVerifyOTP"
            >
              {{ loading ? 'Memverifikasi...' : 'Masuk' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>