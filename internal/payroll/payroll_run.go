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
	ErrPaymentNotFound     = errors.New("worker payment not found")
	ErrInvalidTransition   = errors.New("invalid state transition")
)

type Money struct {
	MinorUnits int64  `json:"minorUnits"`
	Currency   string `json:"currency"`
	Scale      int    `json:"scale"`
}

type PayPeriod struct {
	StartsOn time.Time `json:"startsOn"`
	EndsOn   time.Time `json:"endsOn"`
}

type PaymentObligationInput struct {
	ID             string `json:"id"`
	WorkerID       string `json:"workerId"`
	WorkerName     string `json:"workerName"`
	Amount         Money  `json:"amount"`
	DestinationRef string `json:"destinationRef"`
	Scenario       string `json:"scenario"`
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
	PayrollRunDraft      PayrollRunStatus = "draft"
	PayrollRunApproved   PayrollRunStatus = "approved"
	PayrollRunProcessing PayrollRunStatus = "processing"
	PayrollRunCompleted  PayrollRunStatus = "completed"
)

type WorkerPaymentStatus string

const (
	WorkerPaymentPending        WorkerPaymentStatus = "pending"
	WorkerPaymentProcessing     WorkerPaymentStatus = "processing"
	WorkerPaymentRetryScheduled WorkerPaymentStatus = "retry_scheduled"
	WorkerPaymentComplianceHold WorkerPaymentStatus = "compliance_hold"
	WorkerPaymentUnknown        WorkerPaymentStatus = "unknown"
	WorkerPaymentPaid           WorkerPaymentStatus = "paid"
	WorkerPaymentFailed         WorkerPaymentStatus = "failed"
	WorkerPaymentNeedsReview    WorkerPaymentStatus = "needs_review"
)

type ProviderOutcomeKind string

const (
	ProviderPaid             ProviderOutcomeKind = "paid"
	ProviderValidationFailed ProviderOutcomeKind = "validation_failed"
	ProviderServerError      ProviderOutcomeKind = "server_error"
	ProviderTimedOut         ProviderOutcomeKind = "timed_out"
	ProviderComplianceHold   ProviderOutcomeKind = "compliance_hold"
)

type ProviderOutcome struct {
	Kind   ProviderOutcomeKind
	Detail string
}

type ReconciliationOutcomeKind string

const (
	ReconciliationPaid         ReconciliationOutcomeKind = "paid"
	ReconciliationRejected     ReconciliationOutcomeKind = "rejected"
	ReconciliationNotFound     ReconciliationOutcomeKind = "not_found"
	ReconciliationPending      ReconciliationOutcomeKind = "pending"
	ReconciliationInconclusive ReconciliationOutcomeKind = "inconclusive"
)

type ReconciliationOutcome struct {
	Kind   ReconciliationOutcomeKind
	Detail string
}

type PaymentAttempt struct {
	ID          string              `json:"id"`
	StartedAt   time.Time           `json:"startedAt"`
	CompletedAt time.Time           `json:"completedAt"`
	Outcome     ProviderOutcomeKind `json:"outcome"`
	Detail      string              `json:"detail"`
}

type ReconciliationAttempt struct {
	ID        string                    `json:"id"`
	CreatedAt time.Time                 `json:"createdAt"`
	Outcome   ReconciliationOutcomeKind `json:"outcome"`
	Detail    string                    `json:"detail"`
}

type WorkerPayment struct {
	ID              string                  `json:"id"`
	ObligationID    string                  `json:"obligationId"`
	WorkerID        string                  `json:"workerId"`
	WorkerName      string                  `json:"workerName"`
	Amount          Money                   `json:"amount"`
	DestinationRef  string                  `json:"destinationRef"`
	Scenario        string                  `json:"scenario"`
	Status          WorkerPaymentStatus     `json:"status"`
	Attempts        []PaymentAttempt        `json:"attempts"`
	Reconciliations []ReconciliationAttempt `json:"reconciliations"`
}

type ApprovalResult struct {
	Status   PayrollRunStatus
	Payments []WorkerPayment
}

type PayrollRunSummary struct {
	Total          int  `json:"total"`
	Pending        int  `json:"pending"`
	Processing     int  `json:"processing"`
	Paid           int  `json:"paid"`
	Failed         int  `json:"failed"`
	Unknown        int  `json:"unknown"`
	NeedsReview    int  `json:"needsReview"`
	ComplianceHold int  `json:"complianceHold"`
	HasIssues      bool `json:"hasIssues"`
}

type PayrollRunView struct {
	ID          string                   `json:"id"`
	TenantID    string                   `json:"tenantId"`
	Period      PayPeriod                `json:"period"`
	PayDate     time.Time                `json:"payDate"`
	Status      PayrollRunStatus         `json:"status"`
	ApprovedBy  string                   `json:"approvedBy,omitempty"`
	ApprovedAt  time.Time                `json:"approvedAt,omitempty"`
	Obligations []PaymentObligationInput `json:"obligations"`
	Payments    []WorkerPayment          `json:"payments"`
	Summary     PayrollRunSummary        `json:"summary"`
}

type PayrollRun struct {
	id          string
	tenantID    string
	period      PayPeriod
	payDate     time.Time
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
		id:          input.ID,
		tenantID:    input.TenantID,
		period:      input.Period,
		payDate:     input.PayDate,
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
			if run.approval.ActorID == approval.ActorID {
				return run.approvalResult(), nil
			}
			return ApprovalResult{}, ErrIdempotencyConflict
		}
		return ApprovalResult{}, ErrAlreadyApproved
	}

	run.payments = make([]WorkerPayment, 0, len(run.obligations))
	for _, obligation := range run.obligations {
		run.payments = append(run.payments, WorkerPayment{
			ID:             "payment-" + obligation.ID,
			ObligationID:   obligation.ID,
			WorkerID:       obligation.WorkerID,
			WorkerName:     obligation.WorkerName,
			Amount:         obligation.Amount,
			DestinationRef: obligation.DestinationRef,
			Scenario:       obligation.Scenario,
			Status:         WorkerPaymentPending,
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

func (run *PayrollRun) StartPaymentAttempt(paymentID, attemptID string, startedAt time.Time) error {
	payment, err := run.payment(paymentID)
	if err != nil {
		return err
	}
	if payment.Status != WorkerPaymentPending && payment.Status != WorkerPaymentRetryScheduled {
		return ErrInvalidTransition
	}

	payment.Status = WorkerPaymentProcessing
	payment.Attempts = append(payment.Attempts, PaymentAttempt{ID: attemptID, StartedAt: startedAt})
	run.status = PayrollRunProcessing
	return nil
}

func (run *PayrollRun) RecordProviderOutcome(paymentID, attemptID string, outcome ProviderOutcome, occurredAt time.Time) error {
	payment, err := run.payment(paymentID)
	if err != nil {
		return err
	}
	if payment.Status != WorkerPaymentProcessing || len(payment.Attempts) == 0 {
		return ErrInvalidTransition
	}
	attempt := &payment.Attempts[len(payment.Attempts)-1]
	if attempt.ID != attemptID || attempt.Outcome != "" {
		return ErrInvalidTransition
	}

	attempt.Outcome = outcome.Kind
	attempt.Detail = outcome.Detail
	attempt.CompletedAt = occurredAt
	switch outcome.Kind {
	case ProviderPaid:
		payment.Status = WorkerPaymentPaid
	case ProviderValidationFailed:
		payment.Status = WorkerPaymentFailed
	case ProviderServerError:
		payment.Status = WorkerPaymentRetryScheduled
	case ProviderTimedOut:
		payment.Status = WorkerPaymentUnknown
	case ProviderComplianceHold:
		payment.Status = WorkerPaymentComplianceHold
	default:
		return ErrInvalidTransition
	}
	run.deriveStatus()
	return nil
}

func (run *PayrollRun) RecordReconciliationOutcome(paymentID, reconciliationID string, outcome ReconciliationOutcome, occurredAt time.Time) error {
	payment, err := run.payment(paymentID)
	if err != nil {
		return err
	}
	if payment.Status != WorkerPaymentUnknown {
		return ErrInvalidTransition
	}

	payment.Reconciliations = append(payment.Reconciliations, ReconciliationAttempt{
		ID:        reconciliationID,
		CreatedAt: occurredAt,
		Outcome:   outcome.Kind,
		Detail:    outcome.Detail,
	})
	switch outcome.Kind {
	case ReconciliationPaid:
		payment.Status = WorkerPaymentPaid
	case ReconciliationRejected:
		payment.Status = WorkerPaymentFailed
	case ReconciliationNotFound:
		payment.Status = WorkerPaymentRetryScheduled
	case ReconciliationPending, ReconciliationInconclusive:
		payment.Status = WorkerPaymentUnknown
	default:
		return ErrInvalidTransition
	}
	run.deriveStatus()
	return nil
}

func (run *PayrollRun) MarkNeedsReview(paymentID string) error {
	payment, err := run.payment(paymentID)
	if err != nil {
		return err
	}
	if payment.Status != WorkerPaymentUnknown && payment.Status != WorkerPaymentRetryScheduled && payment.Status != WorkerPaymentComplianceHold {
		return ErrInvalidTransition
	}
	payment.Status = WorkerPaymentNeedsReview
	run.deriveStatus()
	return nil
}

func (run *PayrollRun) ReleaseComplianceHold(paymentID string) error {
	payment, err := run.payment(paymentID)
	if err != nil {
		return err
	}
	if payment.Status != WorkerPaymentComplianceHold {
		return ErrInvalidTransition
	}
	payment.Status = WorkerPaymentRetryScheduled
	run.deriveStatus()
	return nil
}

func (run *PayrollRun) View() PayrollRunView {
	view := PayrollRunView{
		ID:          run.id,
		TenantID:    run.tenantID,
		Period:      run.period,
		PayDate:     run.payDate,
		Status:      run.status,
		Obligations: append([]PaymentObligationInput(nil), run.obligations...),
		Payments:    clonePayments(run.payments),
		Summary:     summarize(run.payments),
	}
	if run.approval != nil {
		view.ApprovedBy = run.approval.ActorID
		view.ApprovedAt = run.approval.ApprovedAt
	}
	return view
}

func (run *PayrollRun) payment(paymentID string) (*WorkerPayment, error) {
	for i := range run.payments {
		if run.payments[i].ID == paymentID {
			return &run.payments[i], nil
		}
	}
	return nil, ErrPaymentNotFound
}

func (run *PayrollRun) deriveStatus() {
	if len(run.payments) == 0 {
		return
	}
	allTerminal := true
	for _, payment := range run.payments {
		if payment.Status != WorkerPaymentPaid && payment.Status != WorkerPaymentFailed && payment.Status != WorkerPaymentNeedsReview {
			allTerminal = false
			break
		}
	}
	if allTerminal {
		run.status = PayrollRunCompleted
	} else {
		run.status = PayrollRunProcessing
	}
}

func clonePayments(payments []WorkerPayment) []WorkerPayment {
	result := make([]WorkerPayment, len(payments))
	for i, payment := range payments {
		result[i] = payment
		result[i].Attempts = append([]PaymentAttempt(nil), payment.Attempts...)
		result[i].Reconciliations = append([]ReconciliationAttempt(nil), payment.Reconciliations...)
	}
	return result
}

func summarize(payments []WorkerPayment) PayrollRunSummary {
	summary := PayrollRunSummary{Total: len(payments)}
	for _, payment := range payments {
		switch payment.Status {
		case WorkerPaymentPending:
			summary.Pending++
		case WorkerPaymentProcessing, WorkerPaymentRetryScheduled:
			summary.Processing++
		case WorkerPaymentComplianceHold:
			summary.ComplianceHold++
		case WorkerPaymentPaid:
			summary.Paid++
		case WorkerPaymentFailed:
			summary.Failed++
		case WorkerPaymentUnknown:
			summary.Unknown++
		case WorkerPaymentNeedsReview:
			summary.NeedsReview++
		}
	}
	summary.HasIssues = summary.Failed > 0 || summary.Unknown > 0 || summary.NeedsReview > 0 || summary.ComplianceHold > 0
	return summary
}
