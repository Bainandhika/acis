<template>
  <section>
    <div class="mb-6 flex items-end justify-between"><div><p class="text-sm font-semibold uppercase tracking-widest text-amber-600">Persetujuan</p><h2 class="mt-1 text-3xl font-bold">Proposal keluarga</h2></div><button v-if="isAdmin" class="rounded bg-slate-900 px-4 py-2 text-sm font-semibold text-white" @click="reviewing = null">{{ pending.length }} menunggu</button></div>
    <form v-if="family" class="mb-5 grid gap-3 rounded-lg bg-white p-5 shadow-sm md:grid-cols-4" @submit.prevent="submitProposal"><select v-model="proposal.wallet_id" required class="rounded border p-2"><option value="" disabled>Pilih dompet</option><option v-for="wallet in wallets" :key="wallet.id" :value="wallet.id">{{ wallet.name }}</option></select><input v-model="proposal.title" required minlength="3" placeholder="Judul proposal" class="rounded border p-2" /><select v-model="proposal.request_type" class="rounded border p-2"><option value="add_transaction">Tambah transaksi</option><option value="edit_transaction">Edit transaksi</option><option value="delete_transaction">Hapus transaksi</option></select><input v-model.number="proposal.amount" type="number" min="0" placeholder="Jumlah" class="rounded border p-2" /><input v-model="proposal.description" placeholder="Deskripsi" class="rounded border p-2 md:col-span-3" /><button class="rounded bg-amber-500 p-2 font-semibold text-white">Ajukan proposal</button></form>
    <p v-if="error" class="mb-4 rounded bg-red-50 p-3 text-sm text-red-700">{{ error }}</p>
    <div v-if="!proposals.length" class="rounded-lg border border-dashed border-slate-300 bg-white p-10 text-center text-slate-500">Belum ada proposal.</div>
    <div v-else class="space-y-3"><article v-for="proposal in proposals" :key="proposal.id" class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm"><div class="flex flex-wrap items-start justify-between gap-3"><div><h3 class="font-semibold">{{ proposal.title }}</h3><p class="mt-1 text-sm text-slate-500">{{ proposal.description || 'Tidak ada deskripsi' }}</p></div><span class="rounded-full px-3 py-1 text-xs font-semibold" :class="statusClass(proposal.status)">{{ proposal.status }}</span></div><div class="mt-4 flex items-center justify-between"><strong>{{ money(proposal.amount) }}</strong><div v-if="isAdmin && proposal.status === 'pending'" class="flex gap-2"><button class="rounded bg-emerald-600 px-3 py-2 text-sm font-medium text-white" @click="decide(proposal.id, true)">Setujui</button><button class="rounded border border-red-200 px-3 py-2 text-sm font-medium text-red-700" @click="decide(proposal.id, false)">Tolak</button></div></div></article></div>
  </section>
</template>
<script setup>
import { computed, onMounted, ref } from 'vue'
import { approveProposal, createProposal, getFamily, getProposals, getWallets, rejectProposal } from '../../services/api'
import { useAuthStore } from '../../stores/useAuthStore'
const proposals = ref([]); const wallets = ref([]); const family = ref(null); const error = ref(''); const reviewing = ref(null); const authStore = useAuthStore(); const proposal = ref({ wallet_id: '', title: '', amount: 0, description: '', request_type: 'add_transaction' })
const isAdmin = computed(() => authStore.user?.role === 'admin' || family.value?.members?.some(member => member.user_id === authStore.user?.id && member.role === 'admin'))
const pending = computed(() => proposals.value.filter(proposal => proposal.status === 'pending'))
const money = value => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(value || 0)
const statusClass = status => status === 'approved' ? 'bg-emerald-100 text-emerald-700' : status === 'rejected' ? 'bg-red-100 text-red-700' : 'bg-amber-100 text-amber-700'
async function load() { try { const [f, p, w] = await Promise.all([getFamily(), getProposals(), getWallets()]); family.value = f.data; proposals.value = p.data || []; wallets.value = w.data || [] } catch (loadError) { error.value = loadError.message } }
async function submitProposal() { try { await createProposal(proposal.value); proposal.value = { wallet_id: '', title: '', amount: 0, description: '', request_type: 'add_transaction' }; await load() } catch (submitError) { error.value = submitError.message } }
async function decide(id, approve) { try { await (approve ? approveProposal(id) : rejectProposal(id)); await load() } catch (actionError) { error.value = actionError.message } }
onMounted(load)
</script>
