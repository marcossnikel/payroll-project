package payroll_test

import (
	"testing"
	"time"

	"github.com/marcossnikel/payroll-project/internal/payroll"
)

func TestTimedOutPaymentCanBeReconciledAsPaid(t *testing.T) {
	run := newValidPayrollRun(t)
	approved, err := run.Approve(validApproval())
	if err != nil {
		t.Fatalf("approve payroll run: %v", err)
	}
	paymentID := approved.Payments[0].ID
	now := time.Date(2026, time.September, 25, 12, 1, 0, 0, time.UTC)

	if err := run.StartPaymentAttempt(paymentID, "attempt-1", now); err != nil {
		t.Fatalf("start payment attempt: %v", err)
	}
	if err := run.RecordProviderOutcome(paymentID, "attempt-1", payroll.ProviderOutcome{
		Kind:   payroll.ProviderTimedOut,
		Detail: "connection closed before response",
	}, now.Add(time.Second)); err != nil {
		t.Fatalf("record timeout: %v", err)
	}

	unknown := run.View().Payments[0]
	if unknown.Status != payroll.WorkerPaymentUnknown {
		t.Fatalf("payment status after timeout = %q, want %q", unknown.Status, payroll.WorkerPaymentUnknown)
	}

	if err := run.RecordReconciliationOutcome(paymentID, "reconciliation-1", payroll.ReconciliationOutcome{
		Kind:   payroll.ReconciliationPaid,
		Detail: "provider confirms settlement",
	}, now.Add(2*time.Second)); err != nil {
		t.Fatalf("record reconciliation: %v", err)
	}

	reconciled := run.View().Payments[0]
	if reconciled.Status != payroll.WorkerPaymentPaid {
		t.Fatalf("payment status after reconciliation = %q, want %q", reconciled.Status, payroll.WorkerPaymentPaid)
	}
	if len(reconciled.Attempts) != 1 || len(reconciled.Reconciliations) != 1 {
		t.Fatalf("history = %d payment attempts and %d reconciliations, want 1 and 1", len(reconciled.Attempts), len(reconciled.Reconciliations))
	}
}
