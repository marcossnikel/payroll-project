package workflow

import (
	"context"

	"github.com/marcossnikel/payroll-project/internal/payroll"
)

type Provider interface {
	Send(context.Context, payroll.WorkerPayment, int) payroll.ProviderOutcome
	Reconcile(context.Context, payroll.WorkerPayment, int) payroll.ReconciliationOutcome
}

type ScenarioProvider struct{}

func NewScenarioProvider() ScenarioProvider {
	return ScenarioProvider{}
}

func (ScenarioProvider) Send(_ context.Context, payment payroll.WorkerPayment, attemptNumber int) payroll.ProviderOutcome {
	switch payment.Scenario {
	case "validation-error":
		return payroll.ProviderOutcome{Kind: payroll.ProviderValidationFailed, Detail: "destination account failed validation"}
	case "server-error":
		if attemptNumber < 3 {
			return payroll.ProviderOutcome{Kind: payroll.ProviderServerError, Detail: "provider returned HTTP 500"}
		}
	case "timeout-paid", "timeout-not-found":
		if attemptNumber == 1 {
			return payroll.ProviderOutcome{Kind: payroll.ProviderTimedOut, Detail: "provider response timed out after submission"}
		}
	case "compliance-hold":
		if attemptNumber == 1 {
			return payroll.ProviderOutcome{Kind: payroll.ProviderComplianceHold, Detail: "synthetic beneficiary review required"}
		}
	}

	return payroll.ProviderOutcome{Kind: payroll.ProviderPaid, Detail: "provider confirmed settlement"}
}

func (ScenarioProvider) Reconcile(_ context.Context, payment payroll.WorkerPayment, reconciliationNumber int) payroll.ReconciliationOutcome {
	switch payment.Scenario {
	case "timeout-paid":
		if reconciliationNumber < 2 {
			return payroll.ReconciliationOutcome{Kind: payroll.ReconciliationPending, Detail: "provider is still processing"}
		}
		return payroll.ReconciliationOutcome{Kind: payroll.ReconciliationPaid, Detail: "provider confirms settlement"}
	case "timeout-not-found":
		if reconciliationNumber < 2 {
			return payroll.ReconciliationOutcome{Kind: payroll.ReconciliationPending, Detail: "provider lookup is not yet authoritative"}
		}
		return payroll.ReconciliationOutcome{Kind: payroll.ReconciliationNotFound, Detail: "provider confirms no payment was created"}
	default:
		return payroll.ReconciliationOutcome{Kind: payroll.ReconciliationInconclusive, Detail: "provider has no conclusive result"}
	}
}
