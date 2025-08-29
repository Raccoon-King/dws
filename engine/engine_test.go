package engine

import (
	"os"
	"regexp"
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
	compiledPattern := regexp.MustCompile("foo")
	rules := []Rule{{ID: "1", Pattern: "foo", Severity: "low", Description: "contains foo", CompiledPattern: compiledPattern}}
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
	rules := []Rule{{ID: "1", Pattern: "[", Severity: "low", CompiledPattern: nil}}
	text := "foo"
	findings := Evaluate(text, "file", rules)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestLoadRulesFromYAML(t *testing.T) {
	// Create a temporary test rules file
	rulesContent := `rules:
- id: profanity-1
  pattern: badword
  severity: high
  description: Detects common profanity
- id: sensitive-phrase-1
  pattern: confidential information
  severity: medium
  description: Detects sensitive phrases
`
	tmpFile := t.TempDir() + "/test_rules.yaml"
	if err := os.WriteFile(tmpFile, []byte(rulesContent), 0644); err != nil {
		t.Fatalf("failed to create temp rules file: %v", err)
	}

	if err := LoadRulesFromYAML(tmpFile); err != nil {
		t.Fatalf("load: %v", err)
	}
	rules := GetRules()

	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}

	// Verify the first rule
	if rules[0].ID != "profanity-1" || rules[0].Pattern != "badword" || rules[0].Severity != "high" || rules[0].Description != "Detects common profanity" {
		t.Fatalf("unexpected first rule: %+v", rules[0])
	}

	// Verify the second rule
	if rules[1].ID != "sensitive-phrase-1" || rules[1].Pattern != "confidential information" || rules[1].Severity != "medium" || rules[1].Description != "Detects sensitive phrases" {
		t.Fatalf("unexpected second rule: %+v", rules[1])
	}

	// Test for non-existent file
	if err := LoadRulesFromYAML("nonexistent.yaml"); err == nil {
		t.Fatalf("expected error for missing file")
	}
}



