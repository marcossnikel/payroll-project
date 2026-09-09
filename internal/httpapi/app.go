package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/marcossnikel/payroll-project/internal/payroll"
	"github.com/marcossnikel/payroll-project/internal/workflow"
)

const demoRunID = "demo-september-2026"

type Config struct {
	StepDelay time.Duration
}

type App struct {
	mu     sync.RWMutex
	config Config
	engine *workflow.Engine
}

func NewApp(config Config) *App {
	if config.StepDelay <= 0 {
		config.StepDelay = 650 * time.Millisecond
	}
	app := &App{config: config}
	app.engine = app.newDemoEngine()
	return app
}

func (app *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", app.health)
	mux.HandleFunc("GET /api/payroll-runs", app.listPayrollRuns)
	mux.HandleFunc("GET /api/payroll-runs/{runID}", app.getPayrollRun)
	mux.HandleFunc("POST /api/payroll-runs/{runID}/approve", app.approvePayrollRun)
	mux.HandleFunc("POST /api/payments/{paymentID}/release", app.releaseComplianceHold)
	mux.HandleFunc("POST /api/demo/reset", app.resetDemo)
	return cors(mux)
}

func (app *App) health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

func (app *App) listPayrollRuns(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, []payroll.PayrollRunView{app.currentEngine().View()})
}

func (app *App) getPayrollRun(writer http.ResponseWriter, request *http.Request) {
	if request.PathValue("runID") != demoRunID {
		writeError(writer, http.StatusNotFound, "payroll run not found")
		return
	}
	writeJSON(writer, http.StatusOK, app.currentEngine().View())
}

func (app *App) approvePayrollRun(writer http.ResponseWriter, request *http.Request) {
	if request.PathValue("runID") != demoRunID {
		writeError(writer, http.StatusNotFound, "payroll run not found")
		return
	}
	var input struct {
		ActorID        string `json:"actor_id"`
		IdempotencyKey string `json:"idempotency_key"`
	}
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid JSON body")
		return
	}
	view, err := app.currentEngine().Approve(request.Context(), payroll.Approval{
		ActorID:        input.ActorID,
		IdempotencyKey: input.IdempotencyKey,
		ApprovedAt:     time.Now().UTC(),
	})
	if err != nil {
		writeDomainError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, view)
}

func (app *App) releaseComplianceHold(writer http.ResponseWriter, request *http.Request) {
	paymentID := strings.TrimSpace(request.PathValue("paymentID"))
	if err := app.currentEngine().ReleaseComplianceHold(paymentID); err != nil {
		writeDomainError(writer, err)
		return
	}
	writeJSON(writer, http.StatusAccepted, app.currentEngine().View())
}

func (app *App) resetDemo(writer http.ResponseWriter, _ *http.Request) {
	app.mu.Lock()
	app.engine = app.newDemoEngine()
	view := app.engine.View()
	app.mu.Unlock()
	writeJSON(writer, http.StatusOK, view)
}

func (app *App) currentEngine() *workflow.Engine {
	app.mu.RLock()
	defer app.mu.RUnlock()
	return app.engine
}

func (app *App) newDemoEngine() *workflow.Engine {
	run, err := payroll.NewPayrollRun(demoPayrollRun())
	if err != nil {
		panic(err)
	}
	return workflow.NewEngine(run, workflow.NewScenarioProvider(), workflow.Config{
		StepDelay:                 app.config.StepDelay,
		MaxPaymentAttempts:        3,
		MaxReconciliationAttempts: 5,
	})
}

func demoPayrollRun() payroll.NewPayrollRunInput {
	return payroll.NewPayrollRunInput{
		ID:       demoRunID,
		TenantID: "northstar-labs",
		Period: payroll.PayPeriod{
			StartsOn: time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
			EndsOn:   time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC),
		},
		PayDate: time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC),
		Obligations: []payroll.PaymentObligationInput{
			demoObligation("ana-silva", "Ana Silva", 725_000, "BRL", "success"),
			demoObligation("bob-okafor", "Bob Okafor", 285_000, "USD", "server-error"),
			demoObligation("carla-mendes", "Carla Mendes", 412_000, "EUR", "timeout-paid"),
			demoObligation("diego-rossi", "Diego Rossi", 338_000, "GBP", "timeout-not-found"),
			demoObligation("eva-torres", "Eva Torres", 196_000, "MXN", "validation-error"),
			demoObligation("femi-adeyemi", "Femi Adeyemi", 540_000, "NGN", "compliance-hold"),
		},
	}
}

func demoObligation(id, name string, minorUnits int64, currency, scenario string) payroll.PaymentObligationInput {
	return payroll.PaymentObligationInput{
		ID:             "obligation-" + id,
		WorkerID:       "worker-" + id,
		WorkerName:     name,
		Amount:         payroll.Money{MinorUnits: minorUnits, Currency: currency, Scale: 2},
		DestinationRef: "mock-account-" + id,
		Scenario:       scenario,
	}
}

func writeDomainError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, payroll.ErrInvalidApproval), errors.Is(err, payroll.ErrInvalidPayrollRun):
		writeError(writer, http.StatusBadRequest, err.Error())
	case errors.Is(err, payroll.ErrPaymentNotFound):
		writeError(writer, http.StatusNotFound, err.Error())
	case errors.Is(err, payroll.ErrAlreadyApproved), errors.Is(err, payroll.ErrIdempotencyConflict), errors.Is(err, payroll.ErrInvalidTransition):
		writeError(writer, http.StatusConflict, err.Error())
	default:
		writeError(writer, http.StatusInternalServerError, "internal server error")
	}
}

func writeError(writer http.ResponseWriter, status int, message string) {
	writeJSON(writer, status, map[string]string{"error": message})
}

func writeJSON(writer http.ResponseWriter, status int, value interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
