package payroll

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrAlreadyApproved     = errors.New("payroll run already approved")
	ErrIdempotencyConflict = errors.New("idempotency key reused with different approval")
	ErrInvalidPayrollRun   = errors.New("invalid payroll run")
	ErrInvalidApproval     = errors.New("invalid payroll run approval")
)

type Money struct {
	MinorUnits int64
	Currency   string
	Scale      int
}

type PayPeriod struct {
	StartsOn time.Time
	EndsOn   time.Time
}

type PaymentObligationInput struct {
	ID             string
	WorkerID       string
	WorkerName     string
	Amount         Money
	DestinationRef string
}

type NewPayrollRunInput struct {
	ID          string
	TenantID    string
	Period      PayPeriod
	PayDate     time.Time
	Obligations []PaymentObligationInput
}

type Approval struct {
	ActorID        string
	IdempotencyKey string
	ApprovedAt     time.Time
}

type PayrollRunStatus string

const (
	PayrollRunDraft    PayrollRunStatus = "draft"
	PayrollRunApproved PayrollRunStatus = "approved"
)

type WorkerPaymentStatus string

const WorkerPaymentPending WorkerPaymentStatus = "pending"

type WorkerPayment struct {
	ID       string
	WorkerID string
	Amount   Money
	Status   WorkerPaymentStatus
}

type ApprovalResult struct {
	Status   PayrollRunStatus
	Payments []WorkerPayment
}

type PayrollRun struct {
	status      PayrollRunStatus
	obligations []PaymentObligationInput
	payments    []WorkerPayment
	approval    *Approval
}

func NewPayrollRun(input NewPayrollRunInput) (*PayrollRun, error) {
	if !isValidPayrollRunInput(input) {
		return nil, ErrInvalidPayrollRun
	}

	return &PayrollRun{
		status:      PayrollRunDraft,
		obligations: append([]PaymentObligationInput(nil), input.Obligations...),
	}, nil
}

func isValidPayrollRunInput(input NewPayrollRunInput) bool {
	if strings.TrimSpace(input.ID) == "" || strings.TrimSpace(input.TenantID) == "" {
		return false
	}
	if input.Period.StartsOn.IsZero() || !input.Period.StartsOn.Before(input.Period.EndsOn) || input.PayDate.IsZero() {
		return false
	}
	if len(input.Obligations) == 0 {
		return false
	}
	for _, obligation := range input.Obligations {
		if strings.TrimSpace(obligation.ID) == "" || strings.TrimSpace(obligation.WorkerID) == "" || strings.TrimSpace(obligation.DestinationRef) == "" {
			return false
		}
		if obligation.Amount.MinorUnits <= 0 || strings.TrimSpace(obligation.Amount.Currency) == "" || obligation.Amount.Scale < 0 {
			return false
		}
	}
	return true
}

func (run *PayrollRun) Approve(approval Approval) (ApprovalResult, error) {
	if strings.TrimSpace(approval.ActorID) == "" || strings.TrimSpace(approval.IdempotencyKey) == "" || approval.ApprovedAt.IsZero() {
		return ApprovalResult{}, ErrInvalidApproval
	}

	if run.status != PayrollRunDraft {
		if run.approval != nil && run.approval.IdempotencyKey == approval.IdempotencyKey {
			if *run.approval == approval {
				return run.approvalResult(), nil
			}
			return ApprovalResult{}, ErrIdempotencyConflict
		}
		return ApprovalResult{}, ErrAlreadyApproved
	}

	run.payments = make([]WorkerPayment, 0, len(run.obligations))
	for _, obligation := range run.obligations {
		run.payments = append(run.payments, WorkerPayment{
			ID:       "payment-" + obligation.ID,
			WorkerID: obligation.WorkerID,
			Amount:   obligation.Amount,
			Status:   WorkerPaymentPending,
		})
	}
	run.status = PayrollRunApproved
	run.approval = &approval

	return run.approvalResult(), nil
}

func (run *PayrollRun) approvalResult() ApprovalResult {
	return ApprovalResult{
		Status:   run.status,
		Payments: append([]WorkerPayment(nil), run.payments...),
	}
}
