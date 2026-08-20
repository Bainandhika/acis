<template>
  <section>
    <div class="mb-6">
      <p class="text-sm font-semibold uppercase tracking-widest text-emerald-600">Ringkasan keluarga</p>
      <h2 class="mt-1 text-3xl font-bold">{{ family?.name || 'Memuat...' }}</h2>
    </div>
    <div v-if="!family" class="rounded-xl border border-blue-200 bg-blue-50 p-6 text-center shadow-sm">
      <h3 class="text-lg font-semibold text-blue-900">Anda belum bergabung dengan keluarga</h3>
      <p class="mt-1 text-sm text-blue-700">Buat keluarga baru atau gabung menggunakan kode undangan untuk mulai mencatat keuangan.</p>
      <router-link to="/keluarga" class="mt-4 inline-block rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-blue-700">
        Buka Menu Keluarga
      </router-link>
    </div>
    <template v-else>
      <div class="grid gap-4 md:grid-cols-3">
        <article v-for="card in cards" :key="card.label"
          class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
          <p class="text-sm text-slate-500">{{ card.label }}</p>
          <p class="mt-2 text-2xl font-bold">{{ money(card.value) }}</p>
        </article>
      </div>
      <div class="mt-6 rounded-lg bg-slate-900 p-6 text-white">
        <p class="text-sm text-slate-300">Kode undangan keluarga</p>
        <p class="mt-2 text-3xl font-bold tracking-[0.25em]">{{ family?.invite_code || '-' }}</p>
      </div>
    </template>
  </section>
</template>
<script setup>
import { computed, onMounted, ref } from 'vue'
import { getFamily, getTransactions, getWallets } from '../../services/api'
const family = ref(null); const transactions = ref([]); const wallets = ref([])
const money = (value) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(value || 0)
const cards = computed(() => [{
  label: 'Pendapatan bulan ini',
  value: transactions.value.filter(t => t.type === 'income').reduce((s, t) => s + t.amount, 0)
}, {
  label: 'Pengeluaran bulan ini',
  value: transactions.value.filter(t => t.type === 'expense').reduce((s, t) => s + t.amount, 0)
}, {
  label: 'Saldo dompet',
  value: wallets.value.reduce((s, w) => s + w.current_balance, 0)
}])
onMounted(async () => {
  try {
    const f = await getFamily()
    family.value = f.data
    if (!family.value) {
      return
    }
    const [t, w] = await Promise.all([
      getTransactions(new Date().getFullYear(), new Date().getMonth() + 1).catch(() => ({ data: [] })),
      getWallets().catch(() => ({ data: [] }))
    ])
    transactions.value = t.data || []
    wallets.value = w.data || []
  } catch (err) {
    console.error('Failed to load dashboard data:', err)
  }
})
</script>
