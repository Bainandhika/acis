<template>
    <div class="min-h-screen bg-slate-100 text-slate-800">
        <header class="bg-slate-900 text-white px-6 py-4 shadow">
            <div class="mx-auto flex max-w-6xl items-center justify-between">
                <div class="flex items-center gap-4">
                    <div
                        class="flex h-10 w-10 items-center justify-center overflow-hidden rounded-lg bg-white/5 ring-1 ring-white/10">
                        <img :src="logo" alt="ACIS logo" class="h-8 w-8 object-contain" />
                    </div>
                    <nav class="hidden gap-4 text-sm md:flex">
                        <router-link to="/">Ringkasan</router-link>
                        <router-link to="/transaksi">Transaksi</router-link>
                        <router-link to="/proposal">Proposal</router-link>
                        <router-link to="/keluarga">Keluarga</router-link>
                    </nav>
                </div>
                <button type="button"
                    class="rounded bg-red-500 px-3 py-2 text-sm font-medium text-white hover:bg-red-600"
                    @click="logout">
                    Keluar
                </button>
            </div>
        </header>

        <main class="mx-auto max-w-6xl p-6">
            <router-view />
        </main>
    </div>
</template>

<script setup>
import { useAuthStore } from '../../stores/useAuthStore'
import { useRouter } from 'vue-router'
import logo from '../../assets/logo-acis.png'

const router = useRouter()
const authStore = useAuthStore()

async function logout() {
    await authStore.logout()
    router.push({ name: 'Login' })
}
</script>
