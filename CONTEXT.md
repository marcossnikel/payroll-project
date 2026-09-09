# Payroll Operations

This context receives already-calculated payroll obligations and tracks their approval, payment execution, and reconciliation.

## Language

**Tenant**:
A business whose payroll data and payment activity are isolated from every other business.
_Avoid_: Account, customer

**Payroll Run**:
An immutable-after-approval collection of payment obligations for one tenant and pay period.
_Avoid_: Batch, payroll calculation

**Payment Obligation**:
The amount and currency a tenant owes one worker in a payroll run.
_Avoid_: Payment attempt, transaction

**Worker Payment**:
The tracked execution of one payment obligation, potentially spanning multiple attempts and reconciliation checks.
_Avoid_: Obligation, provider call

**Payment Attempt**:
One request to an external payment provider made to satisfy a worker payment.
_Avoid_: Retry, obligation

**Reconciliation Attempt**:
One query to an external provider to resolve an ambiguous payment attempt.
_Avoid_: Payment attempt

**Needs Review**:
A terminal state for automation in which the system cannot safely determine the next action and must preserve the full history for an operator.
_Avoid_: Failed, unknown

## Decisions

- A payment obligation is immutable after payroll approval.
- Retrying creates a new payment attempt for the same worker payment; it does not replace or mutate the obligation.
- Provider server errors retry automatically with a bounded attempt budget.
- Invalid destination data fails without retry because repetition cannot repair the input.
- A timeout moves the payment to unknown and reconciliation must run before another send.
- A provider-confirmed not-found result permits a new automatic payment attempt.
- Compliance holds pause automation until an operator releases the hold.
- A payroll run is completed when every payment is terminal, even when its summary carries warnings.
