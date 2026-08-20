<template>
  <div class="flex min-h-screen items-center justify-center bg-slate-100 px-4 py-8">
    <div class="w-full max-w-[440px] rounded-2xl bg-white p-8 shadow-xl ring-1 ring-slate-200">
      <div class="mb-8 text-center">
        <img
          src="../../assets/logo-acis.png"
          alt="ACIS Logo"
          class="mx-auto mb-4 h-16 w-auto"
          fetchpriority="high"
          width="64"
          height="64"
        />
        <p class="mb-1 text-xs font-bold uppercase tracking-widest text-blue-600">ACIS Family Finance</p>
        <h2 class="text-3xl font-extrabold text-slate-800">Masuk ke ACIS</h2>
        <p class="mt-2 text-sm text-slate-500">Kelola keuangan keluarga secara aman & transparan</p>
      </div>

      <div class="space-y-4">
        <p
          v-if="error"
          class="rounded-lg bg-red-50 p-3 text-sm font-medium text-red-700 ring-1 ring-red-200"
        >
          {{ error }}
        </p>

        <button
          type="button"
          :disabled="loading"
          class="flex w-full items-center justify-center gap-3 rounded-xl border border-slate-300 bg-white px-4 py-3.5 text-base font-semibold text-slate-700 shadow-sm transition hover:bg-slate-50 active:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-60"
          @click="handleGoogleLogin"
        >
          <svg class="h-5 w-5" viewBox="0 0 24 24">
            <path
              fill="#4285F4"
              d="M23.745 12.27c0-.7-.06-1.4-.19-2.07H12v4.51h6.6c-.29 1.52-1.14 2.82-2.4 3.68v3.05h3.88c2.27-2.09 3.665-5.17 3.665-9.17z"
            />
            <path
              fill="#34A853"
              d="M12 24c3.24 0 5.95-1.08 7.93-2.91l-3.88-3.05c-1.08.72-2.45 1.16-4.05 1.16-3.12 0-5.77-2.1-6.72-4.93H1.25v3.15C3.26 21.36 7.34 24 12 24z"
            />
            <path
              fill="#FBBC05"
              d="M5.28 14.27c-.25-.72-.38-1.49-.38-2.27s.13-1.55.38-2.27V6.58H1.25C.45 8.18 0 10.03 0 12s.45 3.82 1.25 5.42l4.03-3.15z"
            />
            <path
              fill="#EA4335"
              d="M12 4.75c1.77 0 3.35.61 4.6 1.8l3.42-3.42C17.95 1.19 15.24 0 12 0 7.34 0 3.26 2.64 1.25 6.58l4.03 3.15c.95-2.83 3.6-4.98 6.72-4.98z"
            />
          </svg>
          <span>{{ loading ? 'Menghubungkan...' : 'Lanjutkan dengan Google' }}</span>
        </button>

        <div class="pt-4 text-center text-xs text-slate-400">
          Dengan masuk, Anda menyetujui kebijakan privasi dan keamanan ACIS.
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const error = ref('')

onMounted(async () => {
  await authStore.init()
  if (authStore.isAuthenticated) {
    router.replace('/')
  }
})

async function handleGoogleLogin() {
  error.value = ''
  loading.value = true
  try {
    await authStore.signInWithGoogle()
  } catch (err) {
    error.value = err.message || 'Gagal memulai login dengan Google'
    loading.value = false
  }
}
</script>
