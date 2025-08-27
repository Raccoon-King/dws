package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"dws/engine"
)

func createMultipart(body *bytes.Buffer, filename, content string) *http.Request {
	w := multipart.NewWriter(body)
	fw, _ := w.CreateFormFile("file", filename)
	fw.Write([]byte(content))
	w.Close()
	req := httptest.NewRequest(http.MethodPost, "/scan", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func createTestRules(t *testing.T, rules []engine.Rule) []engine.Rule {
	compiledRules := make([]engine.Rule, len(rules))
	for i, r := range rules {
		compiled, err := regexp.Compile(r.Pattern)
		if err != nil {
			t.Fatalf("failed to compile regex for rule %s: %v", r.ID, err)
		}
		compiledRules[i] = engine.Rule{
			ID:              r.ID,
			Pattern:         r.Pattern,
			Severity:        r.Severity,
			Description:     r.Description,
			CompiledPattern: compiled,
		}
	}
	return compiledRules
}

func TestScanHandler(t *testing.T) {
	rules := createTestRules(t, []engine.Rule{{ID: "1", Pattern: "foo", Severity: "low"}})
	engine.SetRules(rules)
	var b bytes.Buffer
	req := createMultipart(&b, "test.txt", "foo")
	w := httptest.NewRecorder()
	ScanHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestScanHandlerMissingFile(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/scan", nil)
	w := httptest.NewRecorder()
	ScanHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestScanHandlerUnsupported(t *testing.T) {
	var b bytes.Buffer
	req := createMultipart(&b, "file.bin", "data")
	w := httptest.NewRecorder()
	ScanHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReportHandler(t *testing.T) {
	rules := createTestRules(t, []engine.Rule{{ID: "1", Pattern: "foo", Severity: "low", Description: "contains foo"}})
	engine.SetRules(rules)
	var b bytes.Buffer
	req := createMultipart(&b, "test.txt", "foo")
	req.URL.Path = "/report"
	w := httptest.NewRecorder()
	ProcessDocumentHandler(w, req) // Changed ReportHandler to ProcessDocumentHandler
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		FileID   string           `json:"file_id"`
		Findings []engine.Finding `json:"findings"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.FileID != "test.txt" || len(resp.Findings) != 1 || resp.Findings[0].Description != "contains foo" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestScanHandlerBadMultipart(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/scan", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=foo")
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

func TestLoadRulesFromFileHandler(t *testing.T) {
	// Use the rules.yaml file created in the root directory for testing
	path := "C:\\Users\\jesse\\dws\\rules.yaml"

	body, _ := json.Marshal(map[string]string{"path": path})
	req := httptest.NewRequest(http.MethodPost, "/rules/load", bytes.NewReader(body))
	w := httptest.NewRecorder()
	LoadRulesFromFileHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if len(engine.GetRules()) != 2 { // Expect 2 rules from rules.yaml
		t.Fatalf("expected 2 rules, got %d", len(engine.GetRules()))
	}
}

func TestLoadRulesFromFileHandlerBadPath(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"path": "missing.yaml"})
	req := httptest.NewRequest(http.MethodPost, "/rules/load", bytes.NewReader(body))
	w := httptest.NewRecorder()
	LoadRulesFromFileHandler(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestLoadRulesFromFileHandlerBadJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/rules/load", bytes.NewReader([]byte("{")))
	w := httptest.NewRecorder()
	LoadRulesFromFileHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestLoadRulesFromFileHandlerEmptyPath(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"path": ""})
	req := httptest.NewRequest(http.MethodPost, "/rules/load", bytes.NewReader(body))
	w := httptest.NewRecorder()
	LoadRulesFromFileHandler(w, req)
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
