package engine

import "testing"

func TestSetAndGetRules(t *testing.T) {
	rules := []Rule{{ID: "1", Pattern: "foo", Severity: "high"}}
	SetRules(rules)
	got := GetRules()
	if len(got) != 1 || got[0].ID != "1" {
		t.Fatalf("expected rule ID 1, got %+v", got)
	}
}

func TestEvaluate(t *testing.T) {
	rules := []Rule{{ID: "1", Pattern: "foo", Severity: "low"}}
	text := "foo\nbar"
	findings := Evaluate(text, "file", rules)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 1 {
		t.Fatalf("expected line 1, got %d", findings[0].Line)
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
