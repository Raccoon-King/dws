package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"dws/api"
)

// CreateRulesFile writes a minimal rules.yaml and returns its path.
func CreateRulesFile(t testing.TB) string {
	t.Helper()
	yamlContent := `rules:
- id: r1
  pattern: foo
  severity: high
- id: raccoon-mention
  pattern: "\\b(raccoon[s]?)\\b"
  severity: informational
`
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
	api.SetRulesFile(path)
	srv, err := NewServer(path)
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

func TestLoadRules(t *testing.T) {
	// Test with a valid rules file
	path := CreateRulesFile(t)
	os.Setenv("RULES_FILE", path)
	defer os.Unsetenv("RULES_FILE")
	if _, err := NewServer(path); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Test with a missing rules file
	if _, err := NewServer("missing.yaml"); err == nil {
		t.Fatalf("expected error for missing rules file")
	}
}

func TestRunError(t *testing.T) {
	os.Setenv("RULES_FILE", "missing.yaml")
	defer os.Unsetenv("RULES_FILE")

	// Test server creation with missing file
	_, err := NewServer("missing.yaml")
	if err == nil {
		t.Fatalf("expected error from NewServer with missing file")
	}
}