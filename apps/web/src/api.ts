import type { PayrollRun } from './types'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...init?.headers },
  })
  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as { error?: string } | null
    throw new Error(body?.error ?? `Request failed with status ${response.status}`)
  }
  return response.json() as Promise<T>
}

export async function fetchDemoRun(): Promise<PayrollRun> {
  const runs = await request<PayrollRun[]>('/api/payroll-runs')
  if (!runs[0]) throw new Error('No demo payroll run was returned')
  return runs[0]
}

export function approveRun(runId: string): Promise<PayrollRun> {
  return request(`/api/payroll-runs/${runId}/approve`, {
    method: 'POST',
    body: JSON.stringify({
      actor_id: 'operator-marcos',
      idempotency_key: `approve-${runId}`,
    }),
  })
}

export function releaseComplianceHold(paymentId: string): Promise<PayrollRun> {
  return request(`/api/payments/${paymentId}/release`, { method: 'POST', body: '{}' })
}

export function resetDemo(): Promise<PayrollRun> {
  return request('/api/demo/reset', { method: 'POST', body: '{}' })
}
