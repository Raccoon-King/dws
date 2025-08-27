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

func createMultipart(body *bytes.Buffer, filename, content string) *http.Request {
	w := multipart.NewWriter(body)
	fw, _ := w.CreateFormFile("file", filename)
	fw.Write([]byte(content))
	w.Close()
	req := httptest.NewRequest(http.MethodPost, "/scan", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestScanHandler(t *testing.T) {
	engine.SetRules([]engine.Rule{{ID: "1", Pattern: "foo", Severity: "low"}})
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
	engine.SetRules([]engine.Rule{{ID: "1", Pattern: "foo", Severity: "low", Description: "contains foo"}})
	var b bytes.Buffer
	req := createMultipart(&b, "test.txt", "foo")
	req.URL.Path = "/report"
	w := httptest.NewRecorder()
	ReportHandler(w, req)
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
	yaml := "rules:\n- id: r1\n  pattern: foo\n  severity: high\n"
	f, _ := os.CreateTemp(t.TempDir(), "r*.yaml")
	f.WriteString(yaml)
	f.Close()
	body, _ := json.Marshal(map[string]string{"path": f.Name()})
	req := httptest.NewRequest(http.MethodPost, "/rules/load", bytes.NewReader(body))
	w := httptest.NewRecorder()
	LoadRulesFromFileHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if len(engine.GetRules()) != 1 {
		t.Fatalf("rules not loaded")
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
