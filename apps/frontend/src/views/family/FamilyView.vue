<template>
    <section>
        <div class="mb-6">
            <p class="text-sm font-semibold uppercase tracking-widest text-blue-600">Pengaturan</p>
            <h2 class="mt-1 text-3xl font-bold">Keluarga</h2>
        </div>
        <p v-if="error" class="mb-4 rounded bg-red-50 p-3 text-sm text-red-700">{{ error }}</p>
        <div v-if="loading"
            class="rounded-lg border border-dashed border-slate-300 bg-white p-6 text-sm text-slate-500">
            Memuat data keluarga...
        </div>
        <div v-else-if="!family" class="grid gap-5 md:grid-cols-2">
            <form class="rounded-lg bg-white p-5 shadow-sm" @submit.prevent="setup('create')">
                <h3 class="font-semibold">Buat keluarga</h3>
                <div class="mt-4">
                    <label for="create-family-name" class="mb-2 block text-sm font-medium text-slate-700">Nama
                        keluarga</label>
                    <input id="create-family-name" v-model="form.name" required class="w-full rounded border p-2"
                        placeholder="Nama keluarga" />
                </div>
                <button type="submit" class="mt-4 rounded bg-slate-900 px-4 py-2 text-sm font-semibold text-white">Buat
                    keluarga</button>
            </form>
            <form class="rounded-lg bg-white p-5 shadow-sm" @submit.prevent="setup('join')">
                <h3 class="font-semibold">Gabung keluarga</h3>
                <div class="mt-4">
                    <label for="join-family-code" class="mb-2 block text-sm font-medium text-slate-700">Kode
                        undangan</label>
                    <input id="join-family-code" v-model="inviteCode" required maxlength="6"
                        class="w-full rounded border p-2 uppercase" placeholder="Kode undangan" />
                </div>
                <button type="submit"
                    class="mt-4 rounded bg-emerald-600 px-4 py-2 text-sm font-semibold text-white">Gabung</button>
            </form>
        </div>
        <div v-else class="grid gap-5 lg:grid-cols-[1fr_1.3fr]">
            <form class="rounded-lg bg-white p-5 shadow-sm" @submit.prevent="saveFamily">
                <h3 class="font-semibold">Profil keluarga</h3>
                <div class="mt-4">
                    <label for="family-name" class="mb-2 block text-sm font-medium text-slate-700">Nama keluarga</label>
                    <input id="family-name" v-model="form.name" required class="w-full rounded border p-2" />
                </div>
                <div class="mt-3">
                    <label for="family-monthly-income" class="mb-2 block text-sm font-medium text-slate-700">Pendapatan
                        bulanan</label>
                    <input id="family-monthly-income" v-model.number="form.monthly_income" type="number" min="0"
                        class="w-full rounded border p-2" placeholder="Pendapatan bulanan" />
                </div>
                <button type="submit"
                    class="mt-4 rounded bg-slate-900 px-4 py-2 text-sm font-semibold text-white">Simpan
                    pengaturan</button>
            </form>
            <div class="rounded-lg bg-white p-5 shadow-sm">
                <div class="flex items-center justify-between">
                    <h3 class="font-semibold">Dompet</h3>
                    <button type="button" class="rounded bg-emerald-600 px-3 py-2 text-sm font-semibold text-white"
                        @click="walletOpen = !walletOpen">Tambah</button>
                </div>
                <form v-if="walletOpen" class="mt-4 grid gap-3 md:grid-cols-3" @submit.prevent="saveWallet">
                    <div>
                        <label for="wallet-name" class="mb-2 block text-sm font-medium text-slate-700">Nama
                            dompet</label>
                        <input id="wallet-name" v-model="wallet.name" required placeholder="Nama dompet"
                            class="w-full rounded border p-2" />
                    </div>
                    <div>
                        <label for="wallet-initial-balance" class="mb-2 block text-sm font-medium text-slate-700">Saldo
                            awal</label>
                        <input id="wallet-initial-balance" v-model.number="wallet.initial_balance" type="number"
                            placeholder="Saldo awal" class="w-full rounded border p-2" />
                    </div>
                    <div class="flex items-end">
                        <button type="submit" class="w-full rounded bg-slate-900 p-2 text-white">Simpan</button>
                    </div>
                </form>
                <div v-for="item in wallets" :key="item.id" class="mt-4 flex justify-between border-b pb-3">
                    <span><strong>{{ item.short_id }} · {{ item.name }}</strong><small
                            class="block text-slate-500">Batas minimum {{ money(item.minimum_limit) }}</small></span>
                    <div class="flex items-center gap-3"><strong>{{ money(item.current_balance) }}</strong><span
                            v-if="isAdmin" class="flex gap-1"><button type="button"
                                class="rounded border px-2 py-1 text-xs" @click="editWallet(item)">Edit</button><button
                                type="button" class="rounded border border-red-200 px-2 py-1 text-xs text-red-700"
                                @click="removeWallet(item.id)">Hapus</button></span></div>
                </div>
            </div>
        </div>
        <div v-if="isAdmin && family" class="mt-5 rounded-lg bg-white p-5 shadow-sm">
            <div class="flex items-center justify-between">
                <h3 class="font-semibold">Anggota keluarga</h3>
                <button v-if="family.telegram_chat_id" type="button"
                    class="rounded border border-red-200 px-3 py-2 text-sm text-red-700" @click="disconnect">Putuskan
                    Telegram</button>
            </div>
            <div v-for="member in family.members || []" :key="member.id"
                class="mt-3 flex items-center justify-between border-b pb-2">
                <span>{{ member.user_name || member.user_id }} <small class="text-slate-500">({{ member.role
                        }})</small></span>
                <button v-if="member.user_id !== authStore.user?.id" type="button"
                    class="rounded border border-red-200 px-2 py-1 text-xs text-red-700"
                    @click="removeMemberFromFamily(member.id)">Keluarkan</button>
            </div>
        </div>
    </section>
</template>
<script setup>
import { computed, onMounted, ref } from 'vue'
import { createFamily, createWallet, deleteWallet, disconnectTelegram, getFamily, getWallets, joinFamily, removeMember, updateFamily, updateFamilySettings, updateWallet } from '../../services/api'
import { useAuthStore } from '../../stores/useAuthStore'
const authStore = useAuthStore(); const family = ref(null); const wallets = ref([]); const error = ref(''); const loading = ref(true); const walletOpen = ref(false); const inviteCode = ref(''); const form = ref({ name: '', monthly_income: 0 }); const wallet = ref({ name: '', initial_balance: 0, minimum_limit: 0 })
const isAdmin = computed(() => authStore.user?.role === 'admin' || family.value?.members?.some(member => member.user_id === authStore.user?.id && member.role === 'admin'))
const money = value => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(value || 0)
async function load() {
    loading.value = true
    try {
        const f = await getFamily();
        family.value = f.data

        if (family.value) {
            const w = await getWallets();
            form.value = { name: f.data.name, monthly_income: f.data.monthly_income }
            wallets.value = w.data || []
        }
    } catch (loadError) {
        error.value = loadError.message
    } finally {
        loading.value = false
    }
}
async function setup(mode) { try { await (mode === 'create' ? createFamily({ name: form.value.name, monthly_income: form.value.monthly_income }) : joinFamily({ invite_code: inviteCode.value.toUpperCase() })); await load() } catch (setupError) { error.value = setupError.message } }
async function saveFamily() { try { await updateFamily({ name: form.value.name }); await updateFamilySettings({ monthly_income: form.value.monthly_income }); await load() } catch (saveError) { error.value = saveError.message } }
async function saveWallet() { try { await createWallet(wallet.value); wallet.value = { name: '', initial_balance: 0, minimum_limit: 0 }; walletOpen.value = false; await load() } catch (saveError) { error.value = saveError.message } }
async function editWallet(item) {
    const name = window.prompt('Nama dompet', item.name)
    if (name && name !== item.name) {
        try {
            await updateWallet(item.id, { name, minimum_limit: item.minimum_limit })
            await load()
        } catch (saveError) {
            error.value = saveError.message
        }
    }
}
async function removeWallet(id) {
    if (window.confirm('Hapus dompet ini?')) {
        try {
            await deleteWallet(id)
            await load()
        } catch (deleteError) {
            error.value = deleteError.message
        }
    }
}
async function removeMemberFromFamily(id) {
    if (window.confirm('Keluarkan anggota ini?')) {
        try {
            await removeMember(id)
            await load()
        } catch (removeError) {
            error.value = removeError.message
        }
    }
}
async function disconnect() {
    if (window.confirm('Putuskan koneksi Telegram?')) {
        try {
            await disconnectTelegram()
            await load()
        } catch (disconnectError) {
            error.value = disconnectError.message
        }
    }
}
onMounted(load)
</script>
