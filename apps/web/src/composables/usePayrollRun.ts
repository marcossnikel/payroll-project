import { computed, onMounted, onUnmounted, ref } from 'vue'
import { approveRun, fetchDemoRun, releaseComplianceHold, resetDemo } from '../api'
import type { PayrollRun, WorkerPayment } from '../types'

export function usePayrollRun() {
  const run = ref<PayrollRun | null>(null)
  const loading = ref(true)
  const action = ref<string | null>(null)
  const error = ref<string | null>(null)
  let pollTimer: number | undefined

  const payments = computed(() => run.value?.payments ?? [])

  async function load(silent = false) {
    if (!silent) loading.value = true
    try {
      run.value = await fetchDemoRun()
      error.value = null
    } catch (cause) {
      error.value = cause instanceof Error ? cause.message : 'Unable to load payroll run'
    } finally {
      if (!silent) loading.value = false
    }
  }

  async function approve() {
    if (!run.value) return
    action.value = 'approve'
    try {
      run.value = await approveRun(run.value.id)
      error.value = null
    } catch (cause) {
      error.value = cause instanceof Error ? cause.message : 'Unable to approve payroll run'
    } finally {
      action.value = null
    }
  }

  async function release(payment: WorkerPayment) {
    action.value = payment.id
    try {
      run.value = await releaseComplianceHold(payment.id)
      error.value = null
    } catch (cause) {
      error.value = cause instanceof Error ? cause.message : 'Unable to release compliance hold'
    } finally {
      action.value = null
    }
  }

  async function reset() {
    action.value = 'reset'
    try {
      run.value = await resetDemo()
      error.value = null
    } catch (cause) {
      error.value = cause instanceof Error ? cause.message : 'Unable to reset demo'
    } finally {
      action.value = null
    }
  }

  onMounted(() => {
    void load()
    pollTimer = window.setInterval(() => void load(true), 700)
  })
  onUnmounted(() => window.clearInterval(pollTimer))

  return { run, payments, loading, action, error, load, approve, release, reset }
}
