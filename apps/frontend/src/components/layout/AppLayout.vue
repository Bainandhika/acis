<template>
    <div class="min-h-screen bg-slate-100 text-slate-800">
        <header class="sticky top-0 z-50 bg-slate-900 px-4 sm:px-6 py-3.5 text-white shadow-md">
            <div class="mx-auto flex max-w-6xl items-center justify-between">
                <div class="flex items-center gap-6">
                    <router-link to="/" class="flex items-center gap-3">
                        <div
                            class="flex h-10 w-10 items-center justify-center overflow-hidden rounded-lg bg-white/5 ring-1 ring-white/10">
                            <img :src="logo" alt="ACIS logo" class="h-8 w-8 object-contain" />
                        </div>
                        <span class="hidden sm:inline-block font-bold text-lg tracking-tight text-white">ACIS</span>
                    </router-link>
                    
                    <!-- Desktop Nav -->
                    <nav class="hidden md:flex items-center gap-1 text-sm font-medium">
                        <router-link
                            v-for="item in navItems"
                            :key="item.to"
                            :to="item.to"
                            class="rounded-lg px-3 py-2 text-slate-300 transition-colors hover:bg-white/10 hover:text-white"
                            :class="{ 'bg-white/10 text-emerald-400 font-semibold': isRouteActive(item.to, item.exact) }"
                        >
                            {{ item.label }}
                        </router-link>
                    </nav>
                </div>

                <div class="flex items-center gap-3">
                    <button type="button"
                        class="rounded-lg bg-red-600 px-3.5 py-1.5 text-sm font-medium text-white transition hover:bg-red-700 shadow-sm"
                        @click="logout">
                        Keluar
                    </button>

                    <!-- Mobile Menu Button -->
                    <button
                        type="button"
                        class="flex md:hidden items-center justify-center rounded-lg p-2 text-slate-300 hover:bg-white/10 hover:text-white focus:outline-none"
                        @click="mobileMenuOpen = !mobileMenuOpen"
                        aria-label="Toggle navigation menu"
                    >
                        <svg v-if="!mobileMenuOpen" class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/>
                        </svg>
                        <svg v-else class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                        </svg>
                    </button>
                </div>
            </div>

            <!-- Mobile Nav Dropdown -->
            <div v-show="mobileMenuOpen" class="md:hidden mt-3 border-t border-slate-800 pt-3 pb-1 flex flex-col gap-1">
                <router-link
                    v-for="item in navItems"
                    :key="item.to"
                    :to="item.to"
                    class="rounded-lg px-3 py-2 text-sm font-medium text-slate-300 transition-colors hover:bg-white/10 hover:text-white"
                    :class="{ 'bg-white/10 text-emerald-400 font-semibold': isRouteActive(item.to, item.exact) }"
                    @click="mobileMenuOpen = false"
                >
                    {{ item.label }}
                </router-link>
            </div>
        </header>

        <main class="mx-auto max-w-6xl p-4 sm:p-6">
            <router-view />
        </main>
    </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import logo from '../../assets/logo-acis.png'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const mobileMenuOpen = ref(false)

const navItems = [
    { to: '/', label: 'Ringkasan', exact: true },
    { to: '/transaksi', label: 'Transaksi', exact: false },
    { to: '/proposal', label: 'Proposal', exact: false },
    { to: '/keluarga', label: 'Keluarga', exact: false }
]

function isRouteActive(path, exact) {
    if (exact) {
        return route.path === path
    }
    return route.path.startsWith(path)
}

async function logout() {
    await authStore.signOut()
    router.push({ name: 'Login' })
}
</script>
