<script setup lang="ts">
import { computed } from 'vue'
import { Badge } from '@/components/ui/badge'
import { paymentStatusLabel, runStatusLabel } from '@/format'
import type { PaymentStatus, PayrollRunStatus } from '@/types'

const props = withDefaults(
  defineProps<{ status: PaymentStatus | PayrollRunStatus; kind?: 'payment' | 'run' }>(),
  { kind: 'payment' },
)

const label = computed(() =>
  props.kind === 'run'
    ? runStatusLabel(props.status as PayrollRunStatus)
    : paymentStatusLabel(props.status as PaymentStatus),
)

const variant = computed<'default' | 'secondary' | 'destructive' | 'outline'>(() => {
  if (props.status === 'failed' || props.status === 'needs_review') return 'destructive'
  if (props.status === 'paid' || props.status === 'completed') return 'default'
  if (props.status === 'unknown' || props.status === 'compliance_hold') return 'outline'
  return 'secondary'
})
</script>

<template>
  <Badge
    :variant="variant"
    class="gap-1.5 capitalize"
  >
    <span
      class="size-1.5 rounded-full bg-current"
      :class="{ 'animate-pulse': status === 'processing' || status === 'unknown' }"
      aria-hidden="true"
    />
    {{ label }}
  </Badge>
</template>
