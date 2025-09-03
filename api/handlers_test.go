package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestRulesetHandler(t *testing.T) {
	// Create a temp rules directory
	rulesDir := t.TempDir()

	// Create a rules file
	rulesFile := rulesDir + "/test.yaml"
	if err := os.WriteFile(rulesFile, []byte("rules:\n  - id: r1\n    pattern: foo\n    severity: high\n    description: test rule"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a test file to upload
	testContent := "This document contains foo which should trigger a rule"
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(testContent)); err != nil {
		t.Fatalf("write to form: %v", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", "/ruleset?rule=test", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RulesetHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var report struct {
		FileID   string          `json:"fileID"`
		Findings []engine.Finding `json:"findings"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&report); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(report.Findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(report.Findings))
	}
}

func TestRulesetHandler_InvalidRule(t *testing.T) {
	// Test with invalid rule name
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", "/ruleset?rule=../../invalid", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RulesetHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestRulesetHandler_MissingRule(t *testing.T) {
	// Test without rule query parameter
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", "/ruleset", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RulesetHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestLoadRulesFromFileHandler(t *testing.T) {
	// Create a dummy rules file
	_, err := os.CreateTemp(t.TempDir(), "rules*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	// TODO: Implement actual test body once LoadRulesFromFileHandler is implemented
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
