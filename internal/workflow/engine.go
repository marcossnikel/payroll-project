package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/marcossnikel/payroll-project/internal/payroll"
)

type Config struct {
	StepDelay                 time.Duration
	MaxPaymentAttempts        int
	MaxReconciliationAttempts int
}

type Engine struct {
	mu       sync.RWMutex
	run      *payroll.PayrollRun
	provider Provider
	config   Config
	started  bool
}

func NewEngine(run *payroll.PayrollRun, provider Provider, config Config) *Engine {
	if config.MaxPaymentAttempts <= 0 {
		config.MaxPaymentAttempts = 3
	}
	if config.MaxReconciliationAttempts <= 0 {
		config.MaxReconciliationAttempts = 5
	}
	return &Engine{run: run, provider: provider, config: config}
}

func (engine *Engine) Approve(_ context.Context, approval payroll.Approval) (payroll.PayrollRunView, error) {
	engine.mu.Lock()
	_, err := engine.run.Approve(approval)
	if err != nil {
		engine.mu.Unlock()
		return payroll.PayrollRunView{}, err
	}
	view := engine.run.View()
	shouldStart := !engine.started
	engine.started = true
	engine.mu.Unlock()

	if shouldStart {
		for _, payment := range view.Payments {
			go engine.processPayment(payment.ID)
		}
	}
	return view, nil
}

func (engine *Engine) View() payroll.PayrollRunView {
	engine.mu.RLock()
	defer engine.mu.RUnlock()
	return engine.run.View()
}

func (engine *Engine) WaitUntilSettled(ctx context.Context) (payroll.PayrollRunView, error) {
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		view := engine.View()
		if view.Status == payroll.PayrollRunCompleted {
			return view, nil
		}
		select {
		case <-ctx.Done():
			return view, ctx.Err()
		case <-ticker.C:
		}
	}
}

func (engine *Engine) ReleaseComplianceHold(paymentID string) error {
	engine.mu.Lock()
	err := engine.run.ReleaseComplianceHold(paymentID)
	engine.mu.Unlock()
	if err != nil {
		return err
	}
	go engine.processPayment(paymentID)
	return nil
}

func (engine *Engine) processPayment(paymentID string) {
	for {
		engine.mu.Lock()
		view, err := paymentFromView(engine.run.View(), paymentID)
		if err != nil {
			engine.mu.Unlock()
			return
		}
		attemptNumber := len(view.Attempts) + 1
		if attemptNumber > engine.config.MaxPaymentAttempts {
			_ = engine.run.MarkNeedsReview(paymentID)
			engine.mu.Unlock()
			return
		}
		attemptID := fmt.Sprintf("%s-attempt-%d", paymentID, attemptNumber)
		if err := engine.run.StartPaymentAttempt(paymentID, attemptID, time.Now().UTC()); err != nil {
			engine.mu.Unlock()
			return
		}
		view, _ = paymentFromView(engine.run.View(), paymentID)
		engine.mu.Unlock()

		if !sleep(context.Background(), engine.config.StepDelay) {
			return
		}
		outcome := engine.provider.Send(context.Background(), view, attemptNumber)

		engine.mu.Lock()
		err = engine.run.RecordProviderOutcome(paymentID, attemptID, outcome, time.Now().UTC())
		engine.mu.Unlock()
		if err != nil {
			return
		}

		switch outcome.Kind {
		case payroll.ProviderServerError:
			if !sleep(context.Background(), engine.config.StepDelay*time.Duration(attemptNumber)) {
				return
			}
			continue
		case payroll.ProviderTimedOut:
			engine.reconcilePayment(paymentID)
		}
		return
	}
}

func (engine *Engine) reconcilePayment(paymentID string) {
	for reconciliationNumber := 1; reconciliationNumber <= engine.config.MaxReconciliationAttempts; reconciliationNumber++ {
		if !sleep(context.Background(), engine.config.StepDelay*time.Duration(reconciliationNumber)) {
			return
		}

		engine.mu.RLock()
		payment, err := paymentFromView(engine.run.View(), paymentID)
		engine.mu.RUnlock()
		if err != nil || payment.Status != payroll.WorkerPaymentUnknown {
			return
		}

		outcome := engine.provider.Reconcile(context.Background(), payment, reconciliationNumber)
		reconciliationID := fmt.Sprintf("%s-reconciliation-%d", paymentID, reconciliationNumber)
		engine.mu.Lock()
		err = engine.run.RecordReconciliationOutcome(paymentID, reconciliationID, outcome, time.Now().UTC())
		engine.mu.Unlock()
		if err != nil {
			return
		}

		switch outcome.Kind {
		case payroll.ReconciliationPaid, payroll.ReconciliationRejected:
			return
		case payroll.ReconciliationNotFound:
			engine.processPayment(paymentID)
			return
		}
	}

	engine.mu.Lock()
	_ = engine.run.MarkNeedsReview(paymentID)
	engine.mu.Unlock()
}

func paymentFromView(view payroll.PayrollRunView, paymentID string) (payroll.WorkerPayment, error) {
	for _, payment := range view.Payments {
		if payment.ID == paymentID {
			return payment, nil
		}
	}
	return payroll.WorkerPayment{}, payroll.ErrPaymentNotFound
}

func sleep(ctx context.Context, duration time.Duration) bool {
	if duration <= 0 {
		return true
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
