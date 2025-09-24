package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetAndGetRules(t *testing.T) {
	rules := []Rule{
		{ID: "1", Pattern: "test", Severity: "low", Description: "Test rule"},
		{ID: "2", Pattern: "foo", Severity: "high", Description: "Foo rule"},
	}

	SetRules(rules)
	retrieved := GetRules()

	if len(retrieved) != len(rules) {
		t.Errorf("Expected %d rules, got %d", len(rules), len(retrieved))
	}

	for i, rule := range retrieved {
		if rule.ID != rules[i].ID {
			t.Errorf("Rule %d ID mismatch: expected %s, got %s", i, rules[i].ID, rule.ID)
		}
		if rule.Pattern != rules[i].Pattern {
			t.Errorf("Rule %d Pattern mismatch: expected %s, got %s", i, rules[i].Pattern, rule.Pattern)
		}
		if rule.Severity != rules[i].Severity {
			t.Errorf("Rule %d Severity mismatch: expected %s, got %s", i, rules[i].Severity, rule.Severity)
		}
		if rule.Description != rules[i].Description {
			t.Errorf("Rule %d Description mismatch: expected %s, got %s", i, rules[i].Description, rule.Description)
		}
	}
}

func TestEvaluate(t *testing.T) {
	rules := []Rule{
		{ID: "test-rule", Pattern: "error", Severity: "high", Description: "Error pattern"},
		{ID: "info-rule", Pattern: "info", Severity: "low", Description: "Info pattern"},
	}

	text := "This is an error message\nThis is an info message\nThis is a normal message"
	findings := Evaluate(text, "test.txt", rules)

	expectedFindings := 2
	if len(findings) != expectedFindings {
		t.Errorf("Expected %d findings, got %d", expectedFindings, len(findings))
	}

	// Check first finding
	if findings[0].FileID != "test.txt" {
		t.Errorf("Expected file ID 'test.txt', got '%s'", findings[0].FileID)
	}
	if findings[0].RuleID != "test-rule" {
		t.Errorf("Expected rule ID 'test-rule', got '%s'", findings[0].RuleID)
	}
	if findings[0].Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", findings[0].Severity)
	}
	if findings[0].Line != 1 {
		t.Errorf("Expected line 1, got %d", findings[0].Line)
	}
	if findings[0].Description != "Error pattern" {
		t.Errorf("Expected description 'Error pattern', got '%s'", findings[0].Description)
	}

	// Check second finding
	if findings[1].RuleID != "info-rule" {
		t.Errorf("Expected rule ID 'info-rule', got '%s'", findings[1].RuleID)
	}
	if findings[1].Line != 2 {
		t.Errorf("Expected line 2, got %d", findings[1].Line)
	}
}

func TestEvaluateBadRegex(t *testing.T) {
	rules := []Rule{
		{ID: "1", Pattern: "[", Severity: "high", Description: "Bad regex"}, // Invalid regex
		{ID: "2", Pattern: "test", Severity: "low", Description: "Good regex"},
	}

	text := "This is a test"
	findings := Evaluate(text, "test.txt", rules)

	// Should only find the valid regex match
	if len(findings) != 1 {
		t.Errorf("Expected 1 finding (bad regex should be skipped), got %d", len(findings))
	}

	if findings[0].RuleID != "2" {
		t.Errorf("Expected rule ID '2', got '%s'", findings[0].RuleID)
	}
}

func TestLoadRulesFromFile(t *testing.T) {
	// Create temporary YAML file
	tempDir := t.TempDir()
	yamlFile := filepath.Join(tempDir, "test_rules.yaml")

	yamlContent := `rules:
  - id: yaml-rule-1
    pattern: "test"
    severity: medium
    description: "Test pattern from YAML"
  - id: yaml-rule-2
    pattern: "error"
    severity: high
    description: "Error pattern from YAML"
`

	err := os.WriteFile(yamlFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test YAML file: %v", err)
	}

	rules, err := LoadRulesFromFile(yamlFile)
	if err != nil {
		t.Fatalf("LoadRulesFromFile failed: %v", err)
	}

	if len(rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(rules))
	}

	// Check first rule
	if rules[0].ID != "yaml-rule-1" {
		t.Errorf("Expected rule ID 'yaml-rule-1', got '%s'", rules[0].ID)
	}
	if rules[0].Pattern != "test" {
		t.Errorf("Expected pattern 'test', got '%s'", rules[0].Pattern)
	}
	if rules[0].Severity != "medium" {
		t.Errorf("Expected severity 'medium', got '%s'", rules[0].Severity)
	}
	if rules[0].Description != "Test pattern from YAML" {
		t.Errorf("Expected description 'Test pattern from YAML', got '%s'", rules[0].Description)
	}
}

func TestLoadRulesFromFileNotFound(t *testing.T) {
	_, err := LoadRulesFromFile("nonexistent.yaml")
	if err == nil {
		t.Errorf("Expected error for nonexistent file")
	}
}

func TestLoadRulesFromFileInvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	yamlFile := filepath.Join(tempDir, "invalid.yaml")

	invalidContent := "invalid: yaml: content: ["
	err := os.WriteFile(yamlFile, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid YAML file: %v", err)
	}

	_, err = LoadRulesFromFile(yamlFile)
	if err == nil {
		t.Errorf("Expected error for invalid YAML")
	}
}

func TestLoadRulesFromYAML(t *testing.T) {
	// Create temporary YAML file
	tempDir := t.TempDir()
	yamlFile := filepath.Join(tempDir, "test_global_rules.yaml")

	yamlContent := `rules:
  - id: global-rule
    pattern: "critical"
    severity: critical
    description: "Critical pattern"
`

	err := os.WriteFile(yamlFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test YAML file: %v", err)
	}

	err = LoadRulesFromYAML(yamlFile)
	if err != nil {
		t.Fatalf("LoadRulesFromYAML failed: %v", err)
	}

	// Check that rules were set globally
	currentRules := GetRules()
	if len(currentRules) != 1 {
		t.Errorf("Expected 1 global rule, got %d", len(currentRules))
	}

	if currentRules[0].ID != "global-rule" {
		t.Errorf("Expected rule ID 'global-rule', got '%s'", currentRules[0].ID)
	}
}

func TestSetAndGetDebugMode(t *testing.T) {
	// Test setting debug mode to true
	SetDebugMode(true)
	if !GetDebugMode() {
		t.Errorf("Expected debug mode to be true")
	}

	// Test setting debug mode to false
	SetDebugMode(false)
	if GetDebugMode() {
		t.Errorf("Expected debug mode to be false")
	}
}
