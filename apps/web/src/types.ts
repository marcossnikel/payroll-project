export type PayrollRunStatus = 'draft' | 'approved' | 'processing' | 'completed'

export type PaymentStatus =
  | 'pending'
  | 'processing'
  | 'retry_scheduled'
  | 'compliance_hold'
  | 'unknown'
  | 'paid'
  | 'failed'
  | 'needs_review'

export interface Money {
  minorUnits: number
  currency: string
  scale: number
}

export interface PaymentAttempt {
  id: string
  startedAt: string
  completedAt: string
  outcome: string
  detail: string
}

export interface ReconciliationAttempt {
  id: string
  createdAt: string
  outcome: string
  detail: string
}

export interface WorkerPayment {
  id: string
  obligationId: string
  workerId: string
  workerName: string
  amount: Money
  destinationRef: string
  scenario: string
  status: PaymentStatus
  attempts: PaymentAttempt[] | null
  reconciliations: ReconciliationAttempt[] | null
}

export interface PaymentObligation {
  id: string
  workerId: string
  workerName: string
  amount: Money
  destinationRef: string
  scenario: string
}

export interface PayrollSummary {
  total: number
  pending: number
  processing: number
  paid: number
  failed: number
  unknown: number
  needsReview: number
  complianceHold?: number
  hasIssues: boolean
}

export interface PayrollRun {
  id: string
  tenantId: string
  period: { startsOn: string; endsOn: string }
  payDate: string
  status: PayrollRunStatus
  approvedBy?: string
  approvedAt?: string
  obligations: PaymentObligation[]
  payments: WorkerPayment[] | null
  summary: PayrollSummary
}
