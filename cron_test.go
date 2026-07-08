package stocks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleCron(t *testing.T) {
	req := httptest.NewRequest("GET", "/cron.json", nil)
	w := httptest.NewRecorder()
	HandleCron(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var spec struct {
		Jobs []struct {
			ID       string `json:"id"`
			Path     string `json:"path"`
			Method   string `json:"method"`
			Schedule string `json:"schedule"`
		} `json:"jobs"`
	}
	if err := json.NewDecoder(w.Body).Decode(&spec); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(spec.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(spec.Jobs))
	}

	if spec.Jobs[0].ID != "cron:stocks:refresh-quotes" {
		t.Errorf("expected id cron:stocks:refresh-quotes, got %s", spec.Jobs[0].ID)
	}
	if spec.Jobs[0].Path != "/api/refresh" {
		t.Errorf("expected path /api/refresh, got %s", spec.Jobs[0].Path)
	}
	if spec.Jobs[0].Schedule != "*/5 * * * *" {
		t.Errorf("expected schedule */5 * * * *, got %s", spec.Jobs[0].Schedule)
	}
}

func TestHandleCron_ContentType(t *testing.T) {
	req := httptest.NewRequest("GET", "/cron.json", nil)
	w := httptest.NewRecorder()
	HandleCron(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}

func TestHandleCron_JobHasRetry(t *testing.T) {
	req := httptest.NewRequest("GET", "/cron.json", nil)
	w := httptest.NewRecorder()
	HandleCron(w, req)

	var spec struct {
		Jobs []struct {
			Retry struct {
				MaxAttempts int `json:"max_attempts"`
			} `json:"retry"`
		} `json:"jobs"`
	}
	if err := json.NewDecoder(w.Body).Decode(&spec); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(spec.Jobs) == 0 {
		t.Fatal("expected at least 1 job")
	}

	if spec.Jobs[0].Retry.MaxAttempts != 1 {
		t.Errorf("expected retry max_attempts=1, got %d", spec.Jobs[0].Retry.MaxAttempts)
	}
}
