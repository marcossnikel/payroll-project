# Payroll Project

A synthetic global payroll operations sandbox for learning reliable payment orchestration with Go, Vue, and AWS serverless services.

The system receives an already-calculated payroll run, approves it once, processes each worker payment independently, and exposes failures and ambiguous outcomes for reconciliation.

## Safety and scope

- Synthetic data only.
- No real money movement, bank credentials, tax calculation, or identity documents.
- No Cadana proprietary code, data, or inferred private architecture.

## Current slice

The first slice is the Go domain: approving a draft payroll run creates one pending worker payment for each payment obligation.
