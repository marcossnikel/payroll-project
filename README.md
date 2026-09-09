# Payroll Project

A synthetic global payroll operations sandbox for learning reliable payment orchestration with Go, Vue, and AWS serverless services.

The app receives an already-calculated payroll run, freezes its obligations on approval, processes each worker payment independently, and makes retries, ambiguous outcomes, reconciliation, failures, and compliance holds visible to a payroll operator.

## What is included

- A dependency-light Go domain model with explicit state transitions.
- An asynchronous workflow engine with automatic bounded retries.
- A deterministic mock payment provider with six teaching scenarios.
- A `net/http` JSON API and an AWS Lambda adapter.
- A Vue 3 operations UI built with TypeScript, shadcn-vue, Reka UI, and Tailwind CSS.
- Strict TypeScript and an ESLint rule that rejects explicit `any` types.
- A Go AWS CDK stack for API Gateway, Lambda, DynamoDB, streams, SNS, SQS, DLQs, CloudWatch, and EventBridge.
- Domain, workflow, HTTP, and frontend tests.
- CI for Go and Vue.

## Safety and scope

- Synthetic data only.
- No real money movement, bank credentials, tax calculation, or identity documents.
- No Cadana proprietary code, data, or inferred private architecture.
- No AWS resources are deployed by setup or test commands.

The CDK stack can create billable AWS resources if you explicitly deploy it. Review the synthesized template and your AWS account before doing that.

## Run locally

Requirements: Go 1.26+, Node.js 24+, and npm.

Install the frontend dependencies:

```bash
npm --prefix apps/web install
```

Start the Go API in one terminal:

```bash
make api
```

Start Vue in another:

```bash
make web
```

Open [http://localhost:5173](http://localhost:5173). The frontend proxies `/api` to the Go server at `http://localhost:8080`.

## The six scenarios

| Worker | Scenario | Expected behavior |
| --- | --- | --- |
| Ana Silva | Success | Provider confirms the first attempt. |
| Bob Okafor | Server error | Two HTTP 500-style failures retry automatically; attempt three succeeds. |
| Carla Mendes | Timeout, paid | The first request becomes unknown; reconciliation later confirms settlement. |
| Diego Rossi | Timeout, not found | Reconciliation confirms no payment exists; a new attempt is sent and succeeds. |
| Eva Torres | Validation error | Bad destination data fails without retry. |
| Femi Adeyemi | Compliance hold | Automation pauses until the operator releases the synthetic hold. |

After Femi is released, the payroll run becomes `completed` with `hasIssues: true` because Eva's payment failed. A completed run means processing is over, not that every payment succeeded.

## Commands

```bash
make test            # race-tested Go suite + frontend lint/types/tests + infra compile
make lint            # Go vet + ESLint
make build           # Go binaries + production Vue bundle + infra compile
make build-lambdas   # Linux ARM64 Lambda bootstrap binaries
make synth           # build Lambda assets and synthesize CloudFormation locally
```

`make synth` downloads a pinned AWS CDK CLI through `npx`; it does not deploy anything.

## API

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/health` | Health check |
| `GET` | `/api/payroll-runs` | List the synthetic run |
| `GET` | `/api/payroll-runs/{id}` | Read the run, summary, payments, and histories |
| `POST` | `/api/payroll-runs/{id}/approve` | Approve idempotently and start execution |
| `POST` | `/api/payments/{id}/release` | Release a synthetic compliance hold |
| `POST` | `/api/demo/reset` | Restore the initial draft |

Approval body:

```json
{
  "actor_id": "operator-marcos",
  "idempotency_key": "approve-demo-september-2026"
}
```

## Repository map

```text
apps/web/                 Vue operations interface
cmd/server/               local HTTP server
cmd/lambda/               Go Lambda entrypoints
internal/payroll/         aggregate, invariants, and state transitions
internal/workflow/        retries, reconciliation, and mock provider
internal/httpapi/         transport and synthetic fixture
infra/                    AWS CDK stack in Go
docs/architecture.md      domain flow and serverless topology
docs/learning-guide.md    exercises and questions without a schedule
```

Start with [the architecture guide](docs/architecture.md), then use [the learning guide](docs/learning-guide.md) while exploring the running app.

## Important architecture boundary

The local application is a complete in-memory sandbox. The AWS stack is production-shaped and synthesizes successfully, but its API Lambda still hosts that same ephemeral demo. Durable DynamoDB repositories, conditional writes, and a real transactional outbox are the next infrastructure exercise; the README does not pretend those adapters already exist.
