package payroll

import "time"

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
}

func NewPayrollRun(input NewPayrollRunInput) (*PayrollRun, error) {
	return &PayrollRun{
		status:      PayrollRunDraft,
		obligations: append([]PaymentObligationInput(nil), input.Obligations...),
	}, nil
}

func (run *PayrollRun) Approve(_ Approval) (ApprovalResult, error) {
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

	return ApprovalResult{
		Status:   run.status,
		Payments: append([]WorkerPayment(nil), run.payments...),
	}, nil
}
