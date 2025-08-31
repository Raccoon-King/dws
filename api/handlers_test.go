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

func TestHealthHandler(t *testing.T) {
	engine.SetRules(nil)
	rulesFile, err := os.CreateTemp(t.TempDir(), "rules*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rulesFile.WriteString("rules:\n  - id: r1\n    pattern: foo\n    severity: high\n"); err != nil {
		t.Fatal(err)
	}
	rulesFile.Close()
	defer os.Remove(rulesFile.Name())

	SetRulesFile(rulesFile.Name())
	if err := engine.LoadRulesFromYAML(rulesFile.Name()); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"status":"ok"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestHealthHandler_NoRulesFile(t *testing.T) {
	engine.SetRules(nil)
	SetRulesFile("nonexistent.yaml")

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusServiceUnavailable {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusServiceUnavailable)
	}
}

func TestHealthHandler_NoRulesLoaded(t *testing.T) {
	engine.SetRules(nil)
	rulesFile, err := os.CreateTemp(t.TempDir(), "rules*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	rulesFile.Close()
	defer os.Remove(rulesFile.Name())

	SetRulesFile(rulesFile.Name())

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusServiceUnavailable {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusServiceUnavailable)
	}
}

func TestScanHandler(t *testing.T) {
	// Create a dummy rules file
	rulesFile, err := os.CreateTemp(t.TempDir(), "rules*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(rulesFile.Name())

	if _, err := rulesFile.WriteString("rules:\n  - id: r1\n    pattern: foo\n    severity: high"); err != nil {
		t.Fatal(err)
	}
	rulesFile.Close()

	if err := engine.LoadRulesFromYAML(rulesFile.Name()); err != nil {
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

	req, err := http.NewRequest("POST", "/scan", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ScanHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var report struct {
		FileID   string           `json:"fileID"`
		Findings []engine.Finding `json:"findings"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&report); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(report.Findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(report.Findings))
	}
}

func TestScanHandler_FileTooLarge(t *testing.T) {
	engine.SetRules(nil)
	// Create a dummy rules file
	rulesFile, err := os.CreateTemp(t.TempDir(), "rules*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rulesFile.WriteString("rules:\n  - id: r1\n    pattern: foo\n    severity: high\n"); err != nil {
		t.Fatal(err)
	}
	rulesFile.Close()
	defer os.Remove(rulesFile.Name())

	if err := engine.LoadRulesFromYAML(rulesFile.Name()); err != nil {
		t.Fatal(err)
	}

	// Create a large test file exceeding maxUploadSize
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "big.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	big := bytes.Repeat([]byte("a"), int(maxUploadSize)+1)
	if _, err := part.Write(big); err != nil {
		t.Fatalf("write to form: %v", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", "/scan", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ScanHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestReloadRulesHandler(t *testing.T) {
	rules := []engine.Rule{
		{ID: "test-reload", Pattern: "test", Severity: "low", Description: "test rule"},
	}
	reqBody := map[string][]engine.Rule{"rules": rules}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal rules: %v", err)
	}

	req, err := http.NewRequest("POST", "/rules/reload", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ReloadRulesHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestLoadRulesFromFileHandler(t *testing.T) {
	// Create a dummy rules file
	rulesFile, err := os.CreateTemp(t.TempDir(), "rules*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(rulesFile.Name())

	if _, err := rulesFile.WriteString("rules:\n  - id: r1\n    pattern: foo\n    severity: high"); err != nil {
		t.Fatal(err)
	}
	rulesFile.Close()

	reqBody := map[string]string{"path": rulesFile.Name()}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "/rules/load", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(LoadRulesFromFileHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
