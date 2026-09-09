import type { Money, PaymentStatus, PayrollRunStatus } from './types'

export function formatMoney(money: Money): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: money.currency,
    minimumFractionDigits: money.scale,
    maximumFractionDigits: money.scale,
  }).format(money.minorUnits / 10 ** money.scale)
}

export function formatDate(value: string): string {
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: '2-digit',
    year: 'numeric',
    timeZone: 'UTC',
  }).format(new Date(value))
}

export function formatTime(value: string): string {
  return new Intl.DateTimeFormat('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(new Date(value))
}

const paymentLabels: Record<PaymentStatus, string> = {
  pending: 'Pending',
  processing: 'Processing',
  retry_scheduled: 'Retry scheduled',
  compliance_hold: 'Compliance hold',
  unknown: 'Reconciling',
  paid: 'Paid',
  failed: 'Failed',
  needs_review: 'Needs review',
}

const runLabels: Record<PayrollRunStatus, string> = {
  draft: 'Draft',
  approved: 'Approved',
  processing: 'Processing',
  completed: 'Completed',
}

export const paymentStatusLabel = (status: PaymentStatus) => paymentLabels[status]
export const runStatusLabel = (status: PayrollRunStatus) => runLabels[status]
