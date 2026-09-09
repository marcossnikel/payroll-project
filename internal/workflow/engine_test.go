package workflow_test

import (
	"context"
	"testing"
	"time"

	"github.com/marcossnikel/payroll-project/internal/payroll"
	"github.com/marcossnikel/payroll-project/internal/workflow"
)

func TestEngineProcessesIndependentPaymentsAndReconcilesUnknownOutcomes(t *testing.T) {
	run, err := payroll.NewPayrollRun(payroll.NewPayrollRunInput{
		ID:       "run-demo",
		TenantID: "tenant-demo",
		Period: payroll.PayPeriod{
			StartsOn: time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
			EndsOn:   time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC),
		},
		PayDate: time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC),
		Obligations: []payroll.PaymentObligationInput{
			obligation("ana", "success"),
			obligation("bob", "server-error"),
			obligation("carla", "timeout-paid"),
			obligation("diego", "timeout-not-found"),
			obligation("eva", "validation-error"),
		},
	})
	if err != nil {
		t.Fatalf("create payroll run: %v", err)
	}
	engine := workflow.NewEngine(run, workflow.NewScenarioProvider(), workflow.Config{
		StepDelay:                 time.Millisecond,
		MaxPaymentAttempts:        3,
		MaxReconciliationAttempts: 5,
	})

	_, err = engine.Approve(context.Background(), payroll.Approval{
		ActorID:        "operator-marcos",
		IdempotencyKey: "approve-demo",
		ApprovedAt:     time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("approve payroll run: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	view, err := engine.WaitUntilSettled(ctx)
	if err != nil {
		t.Fatalf("wait until settled: %v", err)
	}
	if view.Status != payroll.PayrollRunCompleted {
		t.Fatalf("run status = %q, want %q", view.Status, payroll.PayrollRunCompleted)
	}
	if view.Summary.Paid != 4 || view.Summary.Failed != 1 {
		t.Fatalf("summary = %#v, want 4 paid and 1 failed", view.Summary)
	}
	if !view.Summary.HasIssues {
		t.Fatal("completed run with a failed payment should have issues")
	}
}

func TestComplianceHoldWaitsForOperatorRelease(t *testing.T) {
	run, err := payroll.NewPayrollRun(payroll.NewPayrollRunInput{
		ID:       "run-compliance",
		TenantID: "tenant-demo",
		Period: payroll.PayPeriod{
			StartsOn: time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
			EndsOn:   time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC),
		},
		PayDate:     time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC),
		Obligations: []payroll.PaymentObligationInput{obligation("ana", "compliance-hold")},
	})
	if err != nil {
		t.Fatalf("create payroll run: %v", err)
	}
	engine := workflow.NewEngine(run, workflow.NewScenarioProvider(), workflow.Config{
		StepDelay:                 time.Millisecond,
		MaxPaymentAttempts:        3,
		MaxReconciliationAttempts: 5,
	})
	_, err = engine.Approve(context.Background(), payroll.Approval{
		ActorID:        "operator-marcos",
		IdempotencyKey: "approve-compliance",
		ApprovedAt:     time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("approve payroll run: %v", err)
	}

	deadline := time.Now().Add(time.Second)
	for engine.View().Payments[0].Status != payroll.WorkerPaymentComplianceHold && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if status := engine.View().Payments[0].Status; status != payroll.WorkerPaymentComplianceHold {
		t.Fatalf("payment status = %q, want %q", status, payroll.WorkerPaymentComplianceHold)
	}
	held := engine.View()
	if held.Summary.ComplianceHold != 1 || !held.Summary.HasIssues {
		t.Fatalf("summary during compliance hold = %#v, want one visible issue", held.Summary)
	}

	if err := engine.ReleaseComplianceHold("payment-obligation-ana"); err != nil {
		t.Fatalf("release compliance hold: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	view, err := engine.WaitUntilSettled(ctx)
	if err != nil {
		t.Fatalf("wait until settled: %v", err)
	}
	if view.Summary.Paid != 1 {
		t.Fatalf("summary = %#v, want one paid payment", view.Summary)
	}
}

func obligation(workerID, scenario string) payroll.PaymentObligationInput {
	return payroll.PaymentObligationInput{
		ID:             "obligation-" + workerID,
		WorkerID:       "worker-" + workerID,
		WorkerName:     workerID,
		Amount:         payroll.Money{MinorUnits: 100_00, Currency: "USD", Scale: 2},
		DestinationRef: "mock-account-" + workerID,
		Scenario:       scenario,
	}
}
