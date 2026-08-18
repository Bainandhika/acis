export function formatRp(amount: number, fractions = 0): string {
  if (isNaN(amount) || amount === null || amount === undefined) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: fractions,
    maximumFractionDigits: fractions,
  }).format(amount)
}
