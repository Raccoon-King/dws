package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// CreateRulesFile writes a minimal rules.yaml and returns its path.
func CreateRulesFile(t *testing.T) string {
	t.Helper()
	yamlContent := "rules:\n- id: r1\n  pattern: foo\n  severity: high\n"
	f, err := os.CreateTemp(t.TempDir(), "rules*.yaml")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if _, err := f.WriteString(yamlContent); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestNewServer(t *testing.T) {
	path := CreateRulesFile(t)
	os.Setenv("RULES_FILE", path)
	defer os.Unsetenv("RULES_FILE")
	srv, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestNewServerMissingRules(t *testing.T) {
	os.Setenv("RULES_FILE", "missing.yaml")
	defer os.Unsetenv("RULES_FILE")
	if _, err := NewServer(); err == nil {
		t.Fatalf("expected error for missing rules file")
	}
}

func TestRunError(t *testing.T) {
	os.Setenv("RULES_FILE", "missing.yaml")
	defer os.Unsetenv("RULES_FILE")
	if err := run(); err == nil {
		t.Fatalf("expected error from run")
	}
}
