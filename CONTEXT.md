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
A non-terminal operational condition in which automation cannot safely determine the next action.
_Avoid_: Failed, unknown
