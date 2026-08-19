import { describe, it, expect } from 'vitest'
import { formatCurrency, formatRp, formatDate, formatPercent } from '../../src/utils/format'

describe('format utilities', () => {
  it('formats Indonesian Rupiah currency correctly', () => {
    const formatted = formatRp(150000)
    expect(formatted).toContain('150.000')
    const custom = formatCurrency(100, 'en-US', 'USD')
    expect(custom).toBeDefined()
  })

  it('handles 0 and falsy values in currency formatting', () => {
    const formatted = formatRp(0)
    expect(formatted).toContain('0')
  })

  it('formats dates properly', () => {
    const dateStr = '2026-08-19T00:00:00Z'
    const formatted = formatDate(dateStr)
    expect(formatted).toBeDefined()
    expect(formatted).not.toBe('-')
  })

  it('handles invalid dates gracefully', () => {
    expect(formatDate('')).toBe('-')
    expect(formatDate('invalid-date')).toBe('-')
  })

  it('formats percentages correctly', () => {
    expect(formatPercent(75.4)).toBe('75%')
    expect(formatPercent(100)).toBe('100%')
  })
})
