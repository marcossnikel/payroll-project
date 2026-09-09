package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/marcossnikel/payroll-project/internal/httpapi"
	"github.com/marcossnikel/payroll-project/internal/payroll"
)

func TestOperatorCanReadAndApproveTheDemoPayrollRun(t *testing.T) {
	app := httpapi.NewApp(httpapi.Config{StepDelay: time.Millisecond})
	server := httptest.NewServer(app.Handler())
	defer server.Close()

	response, err := http.Get(server.URL + "/api/payroll-runs")
	if err != nil {
		t.Fatalf("list payroll runs: %v", err)
	}
	defer response.Body.Close()
	var runs []payroll.PayrollRunView
	if err := json.NewDecoder(response.Body).Decode(&runs); err != nil {
		t.Fatalf("decode payroll runs: %v", err)
	}
	if len(runs) != 1 || runs[0].Status != payroll.PayrollRunDraft {
		t.Fatalf("runs = %#v, want one draft run", runs)
	}

	body := `{"actor_id":"operator-marcos","idempotency_key":"approve-demo-run"}`
	response, err = http.Post(server.URL+"/api/payroll-runs/demo-september-2026/approve", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("approve payroll run: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("approve status = %d, want 200", response.StatusCode)
	}
}
