# Learning guide

Use this as a playground, not a calendar. Pick a thread when you have energy and stop when you do not.

## A useful first session

1. Start the API and UI.
2. Open the six payment instructions before approval.
3. Approve the payroll run once.
4. Watch the scenarios diverge.
5. Open Carla and Diego to compare two identical timeouts with different reconciliation results.
6. Release Femi's synthetic compliance hold.
7. Notice that the run completes with a warning because Eva's validation failure is terminal.

## Questions the code should help answer

### Payroll and fintech

- Why is a payment obligation different from a payment attempt?
- Why can a completed payroll run contain failed payments?
- Which data must be immutable after approval?
- Why is a provider timeout more dangerous than an HTTP 500?
- When is human review safer than another retry?

### Go

- Which invariants belong inside `PayrollRun` rather than the HTTP handler?
- Why does the workflow engine lock around mutations but call the provider outside the lock?
- How do interfaces make the provider deterministic in tests?
- Which transitions reject duplicate or out-of-order events?

### Serverless infrastructure

- Why fan out commands through SNS before SQS?
- What does the queue visibility timeout protect?
- What belongs in a dead-letter queue alarm?
- Why does a transactional outbox close the gap between a database write and message publication?
- How would DynamoDB conditional writes enforce approval idempotency?

### Vue

- Which state is server state and which is local UI state?
- Why does the screen poll instead of pretending the approval response contains final results?
- How does the details sheet preserve the aggregate list context?
- Where do discriminated unions prevent impossible status handling?

## Suggested extensions

- Add an `inconclusive` scenario that exhausts five reconciliations and reaches `needs_review`.
- Persist the aggregate in DynamoDB with optimistic concurrency.
- Write an outbox record atomically with approval.
- Replace polling with Server-Sent Events.
- Add tenant authentication and authorization.
- Model FX quotes separately from payment obligations.
- Add an operator action to acknowledge, but never silently erase, a terminal failure.
