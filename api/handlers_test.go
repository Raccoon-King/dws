package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dws/engine"
)

// createTestRulesFile creates a temporary rules file for testing
func createTestRulesFile(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	rulesFile := filepath.Join(tempDir, "test_rules.yaml")

	rulesContent := `rules:
  - id: test-rule
    pattern: "test"
    severity: high
    description: "Test pattern"
  - id: foo-rule
    pattern: "foo"
    severity: medium
    description: "Foo pattern"
`

	err := os.WriteFile(rulesFile, []byte(rulesContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test rules file: %v", err)
	}

	SetRulesFile(rulesFile)
	return rulesFile
}

// createMultipartRequest creates a multipart form request with a file
func createMultipartRequest(t *testing.T, filename, content string) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = part.Write([]byte(content))
	if err != nil {
		t.Fatalf("Failed to write file content: %v", err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/scan", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

func TestScanHandler(t *testing.T) {
	// Setup rules
	engine.SetRules([]engine.Rule{
		{ID: "test-rule", Pattern: "test", Severity: "high", Description: "Test pattern"},
	})

	req := createMultipartRequest(t, "test.txt", "This is a test document")
	w := httptest.NewRecorder()

	ScanHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var response Report
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.FileID != "test.txt" {
		t.Errorf("expected file ID 'test.txt', got '%s'", response.FileID)
	}

	if len(response.Findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(response.Findings))
	}
}

func TestScanHandlerBadMultipart(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/scan", strings.NewReader("invalid"))
	w := httptest.NewRecorder()

	ScanHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestScanHandlerMissingFile(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/scan", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	ScanHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRulesetHandler(t *testing.T) {
	// Create test rules directory
	tempDir := t.TempDir()
	rulesDir := filepath.Join(tempDir, "rules")
	err := os.Mkdir(rulesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create rules directory: %v", err)
	}

	// Create test ruleset
	rulesetFile := filepath.Join(rulesDir, "test.yaml")
	rulesetContent := `rules:
  - id: specific-rule
    pattern: "specific"
    severity: high
    description: "Specific test rule"
`
	err = os.WriteFile(rulesetFile, []byte(rulesetContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create ruleset file: %v", err)
	}

	// Change to temp directory so relative path works
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	req := createMultipartRequest(t, "test.txt", "This contains specific content")
	req.URL.RawQuery = "rule=test"
	w := httptest.NewRecorder()

	RulesetHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRulesetHandlerMissingRule(t *testing.T) {
	req := createMultipartRequest(t, "test.txt", "content")
	w := httptest.NewRecorder()

	RulesetHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRulesetHandlerInvalidRule(t *testing.T) {
	req := createMultipartRequest(t, "test.txt", "content")
	req.URL.RawQuery = "rule=../../../etc/passwd"
	w := httptest.NewRecorder()

	RulesetHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReloadRulesHandler(t *testing.T) {
	rules := []engine.Rule{
		{ID: "new-rule", Pattern: "new", Severity: "low", Description: "New rule"},
	}

	reqBody := map[string]interface{}{
		"rules": rules,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/rules/reload", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ReloadRulesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify rules were set
	currentRules := engine.GetRules()
	if len(currentRules) != 1 || currentRules[0].ID != "new-rule" {
		t.Errorf("Rules were not properly reloaded")
	}
}

func TestReloadRulesHandlerBadRegex(t *testing.T) {
	rules := []engine.Rule{
		{ID: "bad-rule", Pattern: "[", Severity: "low", Description: "Bad regex"},
	}

	reqBody := map[string]interface{}{
		"rules": rules,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/rules/reload", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ReloadRulesHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestLoadRulesFromFileHandler(t *testing.T) {
	rulesFile := createTestRulesFile(t)

	reqBody := map[string]string{
		"path": rulesFile,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/rules/load", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoadRulesFromFileHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLoadRulesFromFileHandlerInvalidPath(t *testing.T) {
	reqBody := map[string]string{
		"path": "../../../etc/passwd",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/rules/load", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoadRulesFromFileHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHealthHandler(t *testing.T) {
	rulesFile := createTestRulesFile(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", response["status"])
	}

	// Clean up
	_ = rulesFile
}

func TestHealthHandlerMissingRulesFile(t *testing.T) {
	SetRulesFile("nonexistent.yaml")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	HealthHandler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestDocsHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	w := httptest.NewRecorder()

	DocsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var docs []EndpointDoc
	err := json.NewDecoder(w.Body).Decode(&docs)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(docs) == 0 {
		t.Errorf("expected at least one endpoint doc")
	}

	// Check for key endpoints
	foundScan := false
	foundHealth := false
	for _, doc := range docs {
		if doc.Path == "/scan" {
			foundScan = true
		}
		if doc.Path == "/health" {
			foundHealth = true
		}
	}

	if !foundScan {
		t.Errorf("missing /scan endpoint in docs")
	}
	if !foundHealth {
		t.Errorf("missing /health endpoint in docs")
	}
}

func TestErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()

	ErrorResponse(w, http.StatusBadRequest, "test error")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var errorResp Error
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	if err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errorResp.Code != http.StatusBadRequest {
		t.Errorf("expected error code 400, got %d", errorResp.Code)
	}

	if errorResp.Message != "test error" {
		t.Errorf("expected message 'test error', got '%s'", errorResp.Message)
	}
}