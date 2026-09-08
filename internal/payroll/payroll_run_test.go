package payroll_test

import (
	"testing"
	"time"

	"github.com/marcossnikel/payroll-project/internal/payroll"
)

func TestApprovingDraftPayrollRunCreatesPendingPaymentForEachObligation(t *testing.T) {
	payDate := time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
	run, err := payroll.NewPayrollRun(payroll.NewPayrollRunInput{
		ID:       "run-2026-09",
		TenantID: "tenant-acme",
		Period: payroll.PayPeriod{
			StartsOn: time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
			EndsOn:   time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC),
		},
		PayDate: payDate,
		Obligations: []payroll.PaymentObligationInput{
			{
				ID:             "obligation-ana",
				WorkerID:       "worker-ana",
				WorkerName:     "Ana Silva",
				Amount:         payroll.Money{MinorUnits: 500_000, Currency: "BRL", Scale: 2},
				DestinationRef: "mock-bank-account-ana",
			},
			{
				ID:             "obligation-bob",
				WorkerID:       "worker-bob",
				WorkerName:     "Bob Smith",
				Amount:         payroll.Money{MinorUnits: 200_000, Currency: "USD", Scale: 2},
				DestinationRef: "mock-bank-account-bob",
			},
		},
	})
	if err != nil {
		t.Fatalf("create payroll run: %v", err)
	}

	result, err := run.Approve(payroll.Approval{
		ActorID:        "operator-marcos",
		IdempotencyKey: "approve-run-2026-09",
		ApprovedAt:     time.Date(2026, time.September, 25, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("approve payroll run: %v", err)
	}

	if result.Status != payroll.PayrollRunApproved {
		t.Fatalf("status = %q, want %q", result.Status, payroll.PayrollRunApproved)
	}
	if len(result.Payments) != 2 {
		t.Fatalf("payment count = %d, want 2", len(result.Payments))
	}

	want := map[string]payroll.Money{
		"worker-ana": {MinorUnits: 500_000, Currency: "BRL", Scale: 2},
		"worker-bob": {MinorUnits: 200_000, Currency: "USD", Scale: 2},
	}
	for _, payment := range result.Payments {
		if payment.Status != payroll.WorkerPaymentPending {
			t.Errorf("payment %q status = %q, want %q", payment.ID, payment.Status, payroll.WorkerPaymentPending)
		}
		if payment.Amount != want[payment.WorkerID] {
			t.Errorf("payment %q amount = %#v, want %#v", payment.ID, payment.Amount, want[payment.WorkerID])
		}
	}
}
