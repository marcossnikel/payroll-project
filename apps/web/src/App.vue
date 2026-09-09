<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  ArrowUpRightIcon,
  ChevronRightIcon,
  CircleAlertIcon,
  RadioIcon,
  RotateCcwIcon,
} from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import PaymentDrawer from '@/components/PaymentDrawer.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { usePayrollRun } from '@/composables/usePayrollRun'
import { formatDate, formatMoney } from '@/format'
import type { WorkerPayment } from '@/types'

const { run, payments, loading, action, error, approve, release, reset, load } = usePayrollRun()
const selectedPaymentID = ref<string | null>(null)

const selectedPayment = computed(() =>
  payments.value.find((payment) => payment.id === selectedPaymentID.value) ?? null,
)

const exceptionCount = computed(() => {
  if (!run.value) return 0
  return (
    run.value.summary.failed +
    run.value.summary.unknown +
    run.value.summary.needsReview +
    (run.value.summary.complianceHold ?? 0)
  )
})

async function releaseAndKeepOpen(payment: WorkerPayment) {
  await release(payment)
}
</script>

<template>
  <div class="min-h-screen">
    <header class="border-b bg-background/90 backdrop-blur">
      <div class="mx-auto flex max-w-7xl items-center justify-between gap-4 px-5 py-4 lg:px-8">
        <a
          class="flex items-center gap-3 text-sm font-semibold"
          href="#"
          aria-label="Payroll operations home"
        >
          <span class="grid size-8 place-items-center rounded-lg bg-primary font-mono text-xs text-primary-foreground">P/</span>
          <span class="hidden sm:inline">Payroll operations</span>
        </a>
        <Badge
          variant="outline"
          class="gap-2"
        >
          <RadioIcon class="text-primary" />
          Synthetic sandbox
        </Badge>
        <Button
          variant="ghost"
          :disabled="action === 'reset'"
          @click="reset"
        >
          <RotateCcwIcon />
          {{ action === 'reset' ? 'Resetting…' : 'Reset scenario' }}
        </Button>
      </div>
    </header>

    <main class="mx-auto flex max-w-7xl flex-col gap-6 px-5 py-8 lg:px-8 lg:py-12">
      <section
        v-if="loading"
        class="grid gap-6 lg:grid-cols-[1fr_22rem]"
      >
        <div class="flex flex-col gap-4">
          <Skeleton class="h-5 w-48" />
          <Skeleton class="h-32 w-full max-w-2xl" />
          <Skeleton class="h-5 w-72" />
        </div>
        <Skeleton class="h-52 w-full" />
      </section>

      <Alert
        v-else-if="error && !run"
        variant="destructive"
      >
        <CircleAlertIcon />
        <AlertTitle>The payroll API is unavailable</AlertTitle>
        <AlertDescription class="flex flex-wrap items-center justify-between gap-4">
          <span>{{ error }}</span>
          <Button
            variant="outline"
            @click="load()"
          >
            Try again
          </Button>
        </AlertDescription>
      </Alert>

      <template v-else-if="run">
        <Alert
          v-if="error"
          variant="destructive"
        >
          <CircleAlertIcon />
          <AlertTitle>Action interrupted</AlertTitle>
          <AlertDescription>{{ error }}</AlertDescription>
        </Alert>

        <section class="grid items-stretch gap-6 lg:grid-cols-[1fr_23rem]">
          <div class="flex min-h-64 flex-col justify-between rounded-xl border bg-card p-6 md:p-9">
            <div class="flex flex-wrap items-center gap-3 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">
              <span>{{ run.tenantId.replaceAll('-', ' ') }}</span>
              <Separator
                orientation="vertical"
                class="h-4"
              />
              <StatusBadge
                :status="run.status"
                kind="run"
              />
            </div>
            <div class="mt-12">
              <h1 class="max-w-3xl font-display text-6xl font-normal leading-[0.88] tracking-[-0.045em] sm:text-7xl lg:text-8xl">
                September<br><em class="text-primary">payroll run</em>
              </h1>
              <p class="mt-6 flex flex-wrap gap-x-5 gap-y-1 text-sm text-muted-foreground">
                <span>{{ formatDate(run.period.startsOn) }} — {{ formatDate(run.period.endsOn) }}</span>
                <span>Pay date {{ formatDate(run.payDate) }}</span>
              </p>
            </div>
          </div>

          <Card class="justify-between border-primary/20 bg-primary text-primary-foreground">
            <CardHeader>
              <Badge
                variant="secondary"
                class="w-fit"
              >
                Irreversible command
              </Badge>
              <CardTitle
                v-if="run.status === 'draft'"
                class="mt-8 font-display text-3xl font-normal"
              >
                Ready for operator approval
              </CardTitle>
              <CardTitle
                v-else
                class="mt-8 font-display text-3xl font-normal"
              >
                Approved snapshot
              </CardTitle>
              <CardDescription class="text-primary-foreground/65">
                <template v-if="run.status === 'draft'">
                  Approval freezes six obligations and starts independent execution.
                </template>
                <template v-else>
                  {{ run.approvedBy }} · {{ run.approvedAt ? formatDate(run.approvedAt) : '' }}
                </template>
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Button
                v-if="run.status === 'draft'"
                variant="secondary"
                size="lg"
                class="w-full"
                :disabled="action === 'approve'"
                @click="approve"
              >
                <Spinner v-if="action === 'approve'" />
                {{ action === 'approve' ? 'Approving…' : 'Approve payroll' }}
                <ArrowUpRightIcon v-if="action !== 'approve'" />
              </Button>
              <div
                v-else
                class="border-t border-primary-foreground/20 pt-5"
              >
                <p class="font-display text-2xl">
                  Immutable
                </p>
                <p class="mt-1 text-xs text-primary-foreground/60">
                  Each payment now evolves independently.
                </p>
              </div>
            </CardContent>
          </Card>
        </section>

        <section
          class="grid grid-cols-2 gap-3 lg:grid-cols-4"
          aria-label="Payroll run summary"
        >
          <Card>
            <CardHeader>
              <CardDescription>Obligations</CardDescription>
              <CardTitle class="font-display text-4xl font-normal">
                {{ run.status === 'draft' ? run.obligations.length : run.summary.total }}
              </CardTitle>
            </CardHeader>
            <CardContent class="text-xs text-muted-foreground">
              Across six currencies
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardDescription>Settled</CardDescription>
              <CardTitle class="font-display text-4xl font-normal">
                {{ run.summary.paid }}
              </CardTitle>
            </CardHeader>
            <CardContent class="text-xs text-muted-foreground">
              Provider-confirmed
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardDescription>In motion</CardDescription>
              <CardTitle class="font-display text-4xl font-normal">
                {{ run.summary.processing + run.summary.pending }}
              </CardTitle>
            </CardHeader>
            <CardContent class="text-xs text-muted-foreground">
              Independent effects
            </CardContent>
          </Card>
          <Card :class="exceptionCount ? 'border-destructive/40' : ''">
            <CardHeader>
              <CardDescription>Exceptions</CardDescription>
              <CardTitle
                class="font-display text-4xl font-normal"
                :class="exceptionCount ? 'text-destructive' : ''"
              >
                {{ exceptionCount }}
              </CardTitle>
            </CardHeader>
            <CardContent class="text-xs text-muted-foreground">
              {{ exceptionCount ? 'Visibility required' : 'None recorded' }}
            </CardContent>
          </Card>
        </section>

        <Card class="overflow-hidden">
          <CardHeader class="border-b">
            <div class="flex flex-col justify-between gap-3 sm:flex-row sm:items-end">
              <div>
                <CardDescription class="uppercase tracking-[0.16em]">
                  Execution ledger
                </CardDescription>
                <CardTitle class="mt-1 font-display text-3xl font-normal">
                  {{ run.status === 'draft' ? 'Payment instructions' : 'Worker payments' }}
                </CardTitle>
              </div>
              <p class="max-w-md text-sm leading-6 text-muted-foreground">
                Each row settles independently. Open a payment to inspect its history.
              </p>
            </div>
          </CardHeader>
          <CardContent class="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Worker</TableHead>
                  <TableHead>Destination</TableHead>
                  <TableHead class="hidden md:table-cell">
                    Scenario
                  </TableHead>
                  <TableHead class="text-right">
                    Amount
                  </TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead><span class="sr-only">Open</span></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody v-if="run.status === 'draft'">
                <TableRow
                  v-for="obligation in run.obligations"
                  :key="obligation.id"
                >
                  <TableCell>
                    <p class="font-medium">
                      {{ obligation.workerName }}
                    </p>
                    <p class="text-xs text-muted-foreground">
                      {{ obligation.workerId }}
                    </p>
                  </TableCell>
                  <TableCell class="font-mono text-xs">
                    {{ obligation.destinationRef }}
                  </TableCell>
                  <TableCell class="hidden capitalize md:table-cell">
                    {{ obligation.scenario.replaceAll('-', ' ') }}
                  </TableCell>
                  <TableCell class="text-right font-mono">
                    {{ formatMoney(obligation.amount) }}
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary">
                      Ready
                    </Badge>
                  </TableCell>
                  <TableCell class="text-muted-foreground">
                    —
                  </TableCell>
                </TableRow>
              </TableBody>
              <TableBody v-else>
                <TableRow
                  v-for="payment in payments"
                  :key="payment.id"
                  class="cursor-pointer"
                  tabindex="0"
                  @click="selectedPaymentID = payment.id"
                  @keydown.enter="selectedPaymentID = payment.id"
                >
                  <TableCell>
                    <p class="font-medium">
                      {{ payment.workerName }}
                    </p>
                    <p class="text-xs text-muted-foreground">
                      {{ payment.workerId }}
                    </p>
                  </TableCell>
                  <TableCell class="font-mono text-xs">
                    {{ payment.destinationRef }}
                  </TableCell>
                  <TableCell class="hidden capitalize md:table-cell">
                    {{ payment.scenario.replaceAll('-', ' ') }}
                  </TableCell>
                  <TableCell class="text-right font-mono">
                    {{ formatMoney(payment.amount) }}
                  </TableCell>
                  <TableCell><StatusBadge :status="payment.status" /></TableCell>
                  <TableCell><ChevronRightIcon class="size-4 text-muted-foreground" /></TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </CardContent>
          <div class="flex items-start gap-4 border-t bg-muted/35 px-6 py-5">
            <span class="font-mono text-xs text-primary">01</span>
            <p class="max-w-3xl text-sm leading-6 text-muted-foreground">
              Timeout is not failure. Ambiguous outcomes move to reconciliation before a new payment obligation is created.
            </p>
          </div>
        </Card>
      </template>
    </main>

    <PaymentDrawer
      v-if="selectedPayment"
      :payment="selectedPayment"
      :busy="action === selectedPayment.id"
      @close="selectedPaymentID = null"
      @release="releaseAndKeepOpen"
    />
  </div>
</template>
