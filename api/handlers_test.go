package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dws/engine"
)

func TestScanHandler(t *testing.T) {
	engine.SetRules([]engine.Rule{{ID: "1", Pattern: "foo", Severity: "low"}})
	body, _ := json.Marshal(map[string]string{"file_id": "f1", "text": "foo"})
	req := httptest.NewRequest(http.MethodPost, "/scan", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ScanHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestScanHandlerBadJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/scan", bytes.NewReader([]byte("{")))
	w := httptest.NewRecorder()
	ScanHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReloadRulesHandler(t *testing.T) {
	rules := []engine.Rule{{ID: "1", Pattern: "foo", Severity: "high"}}
	body, _ := json.Marshal(map[string][]engine.Rule{"rules": rules})
	req := httptest.NewRequest(http.MethodPost, "/rules/reload", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ReloadRulesHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if len(engine.GetRules()) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(engine.GetRules()))
	}
}

func TestReloadRulesHandlerBadJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/rules/reload", bytes.NewReader([]byte("{")))
	w := httptest.NewRecorder()
	ReloadRulesHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	HealthHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content type: %s", ct)
	}
}
