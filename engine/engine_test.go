package engine

import (
	"os"
	"testing"
)

func TestSetAndGetRules(t *testing.T) {
	rules := []Rule{{ID: "1", Pattern: "foo", Severity: "high"}}
	SetRules(rules)
	got := GetRules()
	if len(got) != 1 || got[0].ID != "1" {
		t.Fatalf("expected rule ID 1, got %+v", got)
	}
}

func TestEvaluate(t *testing.T) {
	rules := []Rule{{ID: "1", Pattern: "foo", Severity: "low", Description: "contains foo"}}
	text := "foo\nbar"
	findings := Evaluate(text, "file", rules)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 1 {
		t.Fatalf("expected line 1, got %d", findings[0].Line)
	}
	if findings[0].Description != "contains foo" {
		t.Fatalf("missing description")
	}
}

func TestEvaluateBadRegex(t *testing.T) {
	rules := []Rule{{ID: "1", Pattern: "[", Severity: "low"}}
	text := "foo"
	findings := Evaluate(text, "file", rules)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestLoadRulesFromYAML(t *testing.T) {
	yaml := "rules:\n- id: r1\n  pattern: foo\n  severity: high\n  description: test rule\n"
	f, err := os.CreateTemp(t.TempDir(), "rules*.yaml")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if _, err := f.WriteString(yaml); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()
	if err := LoadRulesFromYAML(f.Name()); err != nil {
		t.Fatalf("load: %v", err)
	}
	rules := GetRules()
	if len(rules) != 1 || rules[0].ID != "r1" || rules[0].Description != "test rule" {
		t.Fatalf("unexpected rules: %+v", rules)
	}

	if err := LoadRulesFromYAML("nonexistent.yaml"); err == nil {
		t.Fatalf("expected error for missing file")
	}
}

func TestLoadRulesFromYAMLAltFormat(t *testing.T) {
	yaml := "rules:\n- id: r1\n  pattern: foo\n  severity: low\n  id: r2\n  pattern: bar\n  severity: high\n"
	f, _ := os.CreateTemp(t.TempDir(), "r*.yaml")
	f.WriteString(yaml)
	f.Close()
	if err := LoadRulesFromYAML(f.Name()); err != nil {
		t.Fatalf("load: %v", err)
	}
	rules := GetRules()
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
}
