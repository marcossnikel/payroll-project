<script setup lang="ts">
import { computed } from 'vue'
import { Clock3Icon, ShieldCheckIcon } from '@lucide/vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { formatMoney, formatTime, paymentStatusLabel } from '@/format'
import type { WorkerPayment } from '@/types'
import StatusBadge from './StatusBadge.vue'

const props = defineProps<{ payment: WorkerPayment; busy: boolean }>()
const emit = defineEmits<{ close: []; release: [payment: WorkerPayment] }>()

interface TimelineEvent {
  id: string
  at: string
  eyebrow: string
  title: string
  detail: string
}

const timeline = computed<TimelineEvent[]>(() => {
  const attempts = (props.payment.attempts ?? []).map((attempt): TimelineEvent => ({
    id: attempt.id,
    at: attempt.outcome ? attempt.completedAt : attempt.startedAt,
    eyebrow: 'Payment attempt',
    title: attempt.outcome ? attempt.outcome.replaceAll('_', ' ') : 'Processing',
    detail: attempt.detail || 'Request sent to the mock provider.',
  }))
  const reconciliations = (props.payment.reconciliations ?? []).map(
    (attempt): TimelineEvent => ({
      id: attempt.id,
      at: attempt.createdAt,
      eyebrow: 'Reconciliation',
      title: attempt.outcome.replaceAll('_', ' '),
      detail: attempt.detail,
    }),
  )
  return [...attempts, ...reconciliations].sort((left, right) => left.at.localeCompare(right.at))
})

function handleOpenChange(open: boolean) {
  if (!open) emit('close')
}
</script>

<template>
  <Sheet
    :open="true"
    @update:open="handleOpenChange"
  >
    <SheetContent class="w-full overflow-y-auto sm:max-w-xl">
      <SheetHeader class="border-b pb-5">
        <div class="flex items-center gap-2">
          <Badge variant="outline">
            Worker payment
          </Badge>
          <StatusBadge :status="payment.status" />
        </div>
        <SheetTitle class="font-display text-4xl font-normal tracking-tight">
          {{ payment.workerName }}
        </SheetTitle>
        <SheetDescription>
          One obligation, its provider effects, and every reconciliation decision.
        </SheetDescription>
      </SheetHeader>

      <div class="flex flex-col gap-6 px-4">
        <div class="flex items-end justify-between gap-4">
          <div>
            <p class="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">
              Net amount
            </p>
            <p class="font-display text-4xl tracking-tight">
              {{ formatMoney(payment.amount) }}
            </p>
          </div>
          <p class="font-mono text-xs text-muted-foreground">
            {{ payment.workerId }}
          </p>
        </div>

        <dl class="grid grid-cols-2 gap-px overflow-hidden rounded-lg border bg-border">
          <div class="bg-card p-4">
            <dt class="text-xs text-muted-foreground">
              Destination
            </dt>
            <dd class="mt-1 font-mono text-sm">
              {{ payment.destinationRef }}
            </dd>
          </div>
          <div class="bg-card p-4">
            <dt class="text-xs text-muted-foreground">
              Scenario
            </dt>
            <dd class="mt-1 text-sm capitalize">
              {{ payment.scenario.replaceAll('-', ' ') }}
            </dd>
          </div>
          <div class="bg-card p-4">
            <dt class="text-xs text-muted-foreground">
              Obligation
            </dt>
            <dd class="mt-1 truncate font-mono text-sm">
              {{ payment.obligationId }}
            </dd>
          </div>
          <div class="bg-card p-4">
            <dt class="text-xs text-muted-foreground">
              Current state
            </dt>
            <dd class="mt-1 text-sm">
              {{ paymentStatusLabel(payment.status) }}
            </dd>
          </div>
        </dl>

        <Button
          v-if="payment.status === 'compliance_hold'"
          class="w-full"
          size="lg"
          :disabled="busy"
          @click="emit('release', payment)"
        >
          <ShieldCheckIcon />
          {{ busy ? 'Releasing…' : 'Release synthetic hold' }}
        </Button>

        <Separator />

        <section class="flex flex-col gap-4">
          <div class="flex items-center justify-between gap-4">
            <div>
              <p class="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">
                Immutable history
              </p>
              <h3 class="mt-1 text-base font-semibold">
                Provider timeline
              </h3>
            </div>
            <Badge variant="secondary">
              {{ timeline.length }} events
            </Badge>
          </div>

          <ol
            v-if="timeline.length"
            class="flex flex-col"
          >
            <li
              v-for="event in timeline"
              :key="event.id"
              class="grid grid-cols-[1rem_1fr] gap-3"
            >
              <div class="flex flex-col items-center">
                <span
                  class="mt-1.5 size-2 rounded-full bg-primary"
                  aria-hidden="true"
                />
                <span
                  class="min-h-10 w-px flex-1 bg-border"
                  aria-hidden="true"
                />
              </div>
              <div class="pb-6">
                <div class="flex items-center justify-between gap-4 text-xs text-muted-foreground">
                  <span class="font-semibold uppercase tracking-[0.14em]">{{ event.eyebrow }}</span>
                  <time class="flex items-center gap-1"><Clock3Icon class="size-3" />{{ formatTime(event.at) }}</time>
                </div>
                <p class="mt-1 font-semibold capitalize">
                  {{ event.title }}
                </p>
                <p class="mt-1 text-sm leading-6 text-muted-foreground">
                  {{ event.detail }}
                </p>
              </div>
            </li>
          </ol>
          <p
            v-else
            class="rounded-lg border border-dashed p-6 text-sm text-muted-foreground"
          >
            No provider activity yet. Approve the payroll run to begin.
          </p>
        </section>
      </div>

      <SheetFooter class="border-t">
        <p class="text-xs leading-5 text-muted-foreground">
          This sandbox never moves real money. Provider responses are deterministic teaching scenarios.
        </p>
      </SheetFooter>
    </SheetContent>
  </Sheet>
</template>
