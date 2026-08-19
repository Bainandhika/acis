export function formatCurrency(amount: number, locale: string = 'id-ID', currency: string = 'IDR'): string {
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount || 0)
}

export function formatRp(amount: number): string {
  return formatCurrency(amount, 'id-ID', 'IDR')
}

export function formatDate(dateStr: string | Date, locale: string = 'id-ID'): string {
  if (!dateStr) return '-'
  const date = typeof dateStr === 'string' ? new Date(dateStr) : dateStr
  if (isNaN(date.getTime())) return '-'
  return date.toLocaleDateString(locale, {
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  })
}

export function formatDateShort(dateStr: string | Date, locale: string = 'id-ID'): string {
  if (!dateStr) return '-'
  const date = typeof dateStr === 'string' ? new Date(dateStr) : dateStr
  if (isNaN(date.getTime())) return '-'
  return date.toLocaleDateString(locale, {
    day: 'numeric',
    month: 'short'
  })
}

export function formatDateTime(dateStr: string | Date, locale: string = 'id-ID'): string {
  if (!dateStr) return '-'
  const date = typeof dateStr === 'string' ? new Date(dateStr) : dateStr
  if (isNaN(date.getTime())) return '-'
  return date.toLocaleDateString(locale, {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

export function formatPercent(value: number): string {
  return `${Math.round(value || 0)}%`
}
