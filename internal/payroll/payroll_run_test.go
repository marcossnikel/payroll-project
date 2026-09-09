package payroll_test

import (
	"errors"
	"reflect"
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

func TestApprovingAnAlreadyApprovedPayrollRunWithAnotherKeyIsRejected(t *testing.T) {
	run := newValidPayrollRun(t)

	_, err := run.Approve(validApproval())
	if err != nil {
		t.Fatalf("first approval: %v", err)
	}

	conflictingApproval := validApproval()
	conflictingApproval.IdempotencyKey = "another-approval-command"
	_, err = run.Approve(conflictingApproval)
	if !errors.Is(err, payroll.ErrAlreadyApproved) {
		t.Fatalf("second approval error = %v, want %v", err, payroll.ErrAlreadyApproved)
	}
}

func TestReplayingTheSameApprovalReturnsTheExistingResult(t *testing.T) {
	run := newValidPayrollRun(t)
	approval := validApproval()

	first, err := run.Approve(approval)
	if err != nil {
		t.Fatalf("first approval: %v", err)
	}
	second, err := run.Approve(approval)
	if err != nil {
		t.Fatalf("replay approval: %v", err)
	}

	if !reflect.DeepEqual(second, first) {
		t.Fatalf("replayed result = %#v, want existing result %#v", second, first)
	}
}

func TestReusingApprovalKeyWithDifferentCommandIsRejectedAsConflict(t *testing.T) {
	run := newValidPayrollRun(t)
	approval := validApproval()

	_, err := run.Approve(approval)
	if err != nil {
		t.Fatalf("first approval: %v", err)
	}

	approval.ActorID = "operator-someone-else"
	_, err = run.Approve(approval)
	if !errors.Is(err, payroll.ErrIdempotencyConflict) {
		t.Fatalf("conflicting replay error = %v, want %v", err, payroll.ErrIdempotencyConflict)
	}
}

func TestPayrollRunWithoutObligationsCannotBeCreated(t *testing.T) {
	input := validPayrollRunInput()
	input.Obligations = nil

	_, err := payroll.NewPayrollRun(input)
	if !errors.Is(err, payroll.ErrInvalidPayrollRun) {
		t.Fatalf("create empty payroll run error = %v, want %v", err, payroll.ErrInvalidPayrollRun)
	}
}

func TestStructurallyInvalidPayrollRunCannotBeCreated(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*payroll.NewPayrollRunInput)
	}{
		{"missing run ID", func(input *payroll.NewPayrollRunInput) { input.ID = "" }},
		{"missing tenant", func(input *payroll.NewPayrollRunInput) { input.TenantID = "" }},
		{"inverted period", func(input *payroll.NewPayrollRunInput) {
			input.Period.StartsOn, input.Period.EndsOn = input.Period.EndsOn, input.Period.StartsOn
		}},
		{"missing pay date", func(input *payroll.NewPayrollRunInput) { input.PayDate = time.Time{} }},
		{"missing obligation ID", func(input *payroll.NewPayrollRunInput) { input.Obligations[0].ID = "" }},
		{"missing worker", func(input *payroll.NewPayrollRunInput) { input.Obligations[0].WorkerID = "" }},
		{"missing destination", func(input *payroll.NewPayrollRunInput) { input.Obligations[0].DestinationRef = "" }},
		{"non-positive amount", func(input *payroll.NewPayrollRunInput) { input.Obligations[0].Amount.MinorUnits = 0 }},
		{"missing currency", func(input *payroll.NewPayrollRunInput) { input.Obligations[0].Amount.Currency = "" }},
		{"negative scale", func(input *payroll.NewPayrollRunInput) { input.Obligations[0].Amount.Scale = -1 }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validPayrollRunInput()
			test.mutate(&input)

			_, err := payroll.NewPayrollRun(input)
			if !errors.Is(err, payroll.ErrInvalidPayrollRun) {
				t.Fatalf("create invalid payroll run error = %v, want %v", err, payroll.ErrInvalidPayrollRun)
			}
		})
	}
}

func TestApprovalWithoutAuditDataIsRejected(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*payroll.Approval)
	}{
		{"missing actor", func(approval *payroll.Approval) { approval.ActorID = "" }},
		{"missing idempotency key", func(approval *payroll.Approval) { approval.IdempotencyKey = "" }},
		{"missing approval time", func(approval *payroll.Approval) { approval.ApprovedAt = time.Time{} }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			run := newValidPayrollRun(t)
			approval := validApproval()
			test.mutate(&approval)

			_, err := run.Approve(approval)
			if !errors.Is(err, payroll.ErrInvalidApproval) {
				t.Fatalf("approve with missing audit data error = %v, want %v", err, payroll.ErrInvalidApproval)
			}
		})
	}
}

func newValidPayrollRun(t *testing.T) *payroll.PayrollRun {
	t.Helper()

	run, err := payroll.NewPayrollRun(validPayrollRunInput())
	if err != nil {
		t.Fatalf("create payroll run: %v", err)
	}
	return run
}

func validPayrollRunInput() payroll.NewPayrollRunInput {
	return payroll.NewPayrollRunInput{
		ID:       "run-2026-09",
		TenantID: "tenant-acme",
		Period: payroll.PayPeriod{
			StartsOn: time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
			EndsOn:   time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC),
		},
		PayDate: time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC),
		Obligations: []payroll.PaymentObligationInput{
			{
				ID:             "obligation-ana",
				WorkerID:       "worker-ana",
				WorkerName:     "Ana Silva",
				Amount:         payroll.Money{MinorUnits: 500_000, Currency: "BRL", Scale: 2},
				DestinationRef: "mock-bank-account-ana",
			},
		},
	}
}

func validApproval() payroll.Approval {
	return payroll.Approval{
		ActorID:        "operator-marcos",
		IdempotencyKey: "approve-run-2026-09",
		ApprovedAt:     time.Date(2026, time.September, 25, 12, 0, 0, 0, time.UTC),
	}
}
