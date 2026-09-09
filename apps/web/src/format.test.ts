import { describe, expect, it } from 'vitest'
import { formatMoney, paymentStatusLabel } from './format'

describe('operations formatting', () => {
  it('formats authoritative minor units using the explicit currency scale', () => {
    expect(formatMoney({ minorUnits: 725_000, currency: 'BRL', scale: 2 })).toBe('R$7,250.00')
  })

  it('names unknown payment state as an active reconciliation operation', () => {
    expect(paymentStatusLabel('unknown')).toBe('Reconciling')
  })
})
