package llm

import (
	"context"
	"testing"

	"dws/engine"
)

func TestNewAnalyzer(t *testing.T) {
	service := &Service{config: Config{Enabled: true}}
	analyzer := NewAnalyzer(service)

	if analyzer == nil {
		t.Errorf("NewAnalyzer() returned nil")
	}

	if analyzer.service != service {
		t.Errorf("NewAnalyzer() service not set correctly")
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple json object",
			input:    `Here is the result: {"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "json array",
			input:    `Result: [{"item": 1}, {"item": 2}]`,
			expected: `[{"item": 1}, {"item": 2}]`,
		},
		{
			name:     "nested json",
			input:    `{"outer": {"inner": "value"}}`,
			expected: `{"outer": {"inner": "value"}}`,
		},
		{
			name:     "no json",
			input:    `This is just text`,
			expected: ``,
		},
		{
			name:     "incomplete json",
			input:    `{"incomplete": `,
			expected: ``,
		},
		{
			name:     "multiple json objects",
			input:    `{"first": 1} and {"second": 2}`,
			expected: `{"first": 1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSON(tt.input)
			if result != tt.expected {
				t.Errorf("extractJSON() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "long string",
			input:    "this is a very long string",
			maxLen:   10,
			expected: "this is a ... [truncated]",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   5,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestConvertLLMFindingsToEngine(t *testing.T) {
	llmFindings := []LLMFinding{
		{
			RuleID:      "test-rule-1",
			Severity:    "high",
			Line:        1,
			Context:     "test context 1",
			Description: "test description 1",
			Confidence:  0.9,
		},
		{
			RuleID:      "test-rule-2",
			Severity:    "medium",
			Line:        5,
			Context:     "test context 2",
			Description: "test description 2",
			Confidence:  0.7,
		},
	}

	fileID := "test.txt"
	engineFindings := ConvertLLMFindingsToEngine(llmFindings, fileID)

	if len(engineFindings) != len(llmFindings) {
		t.Errorf("ConvertLLMFindingsToEngine() returned %d findings, want %d", len(engineFindings), len(llmFindings))
	}

	for i, finding := range engineFindings {
		if finding.FileID != fileID {
			t.Errorf("Finding %d FileID = %q, want %q", i, finding.FileID, fileID)
		}

		if finding.RuleID != llmFindings[i].RuleID {
			t.Errorf("Finding %d RuleID = %q, want %q", i, finding.RuleID, llmFindings[i].RuleID)
		}

		if finding.Severity != llmFindings[i].Severity {
			t.Errorf("Finding %d Severity = %q, want %q", i, finding.Severity, llmFindings[i].Severity)
		}

		if finding.Line != llmFindings[i].Line {
			t.Errorf("Finding %d Line = %d, want %d", i, finding.Line, llmFindings[i].Line)
		}

		if finding.Context != llmFindings[i].Context {
			t.Errorf("Finding %d Context = %q, want %q", i, finding.Context, llmFindings[i].Context)
		}
	}
}

func TestBuildAnalysisPrompt(t *testing.T) {
	analyzer := &Analyzer{}

	req := AnalysisRequest{
		Text:     "This is test content with sensitive data",
		Filename: "test.txt",
		Rules:    []string{"Look for sensitive data", "Check for compliance"},
		Context:  "Security audit document",
	}

	prompt := analyzer.buildAnalysisPrompt(req)

	// Check that prompt contains expected elements
	expectedElements := []string{
		"security analyst",
		"test.txt",
		"Security audit document",
		"Look for sensitive data",
		"Check for compliance",
		"This is test content",
		"JSON",
	}

	for _, element := range expectedElements {
		if !contains(prompt, element) {
			t.Errorf("buildAnalysisPrompt() missing expected element: %q", element)
		}
	}
}

func TestBuildValidationPrompt(t *testing.T) {
	analyzer := &Analyzer{}

	findings := []engine.Finding{
		{
			RuleID:   "test-rule",
			Severity: "high",
			Line:     1,
			Context:  "sensitive data here",
		},
	}

	text := "This document contains sensitive data here"
	filename := "test.txt"

	prompt := analyzer.buildValidationPrompt(findings, text, filename)

	// Check that prompt contains expected elements
	expectedElements := []string{
		"validating security findings",
		"test-rule",
		"high",
		"sensitive data here",
		"test.txt",
		"JSON",
	}

	for _, element := range expectedElements {
		if !contains(prompt, element) {
			t.Errorf("buildValidationPrompt() missing expected element: %q", element)
		}
	}
}

// Mock service for testing analyzer
type MockAnalyzerService struct {
	enabled  bool
	response *CompletionResponse
	error    error
}

func (m *MockAnalyzerService) Complete(ctx context.Context, prompt string) (*CompletionResponse, error) {
	if m.error != nil {
		return nil, m.error
	}
	if m.response != nil {
		return m.response, nil
	}
	return &CompletionResponse{
		Text:       `{"findings": [], "summary": "No issues found", "confidence": 0.8}`,
		TokensUsed: 50,
		Model:      "mock-model",
		Provider:   ProviderOpenAI,
	}, nil
}

func (m *MockAnalyzerService) IsEnabled() bool {
	return m.enabled
}

func TestAnalyzeDocumentDisabled(t *testing.T) {
	service := &MockAnalyzerService{enabled: false}
	analyzer := NewAnalyzer(service)

	req := AnalysisRequest{
		Text:     "test content",
		Filename: "test.txt",
	}

	ctx := context.Background()
	_, err := analyzer.AnalyzeDocument(ctx, req)

	if err == nil {
		t.Errorf("AnalyzeDocument() should return error when service is disabled")
	}

	expectedErr := "LLM service is not enabled"
	if err.Error() != expectedErr {
		t.Errorf("AnalyzeDocument() error = %q, want %q", err.Error(), expectedErr)
	}
}

func TestAnalyzeDocumentSuccess(t *testing.T) {
	responseJSON := `{
		"findings": [
			{
				"rule_id": "test-finding",
				"severity": "high",
				"line": 1,
				"context": "sensitive data",
				"description": "Found sensitive data",
				"confidence": 0.9
			}
		],
		"summary": "Found 1 issue",
		"confidence": 0.85
	}`

	service := &MockAnalyzerService{
		enabled: true,
		response: &CompletionResponse{
			Text:       responseJSON,
			TokensUsed: 100,
			Model:      "gpt-3.5-turbo",
			Provider:   ProviderOpenAI,
		},
	}

	analyzer := NewAnalyzer(service)

	req := AnalysisRequest{
		Text:     "This contains sensitive data",
		Filename: "test.txt",
		Rules:    []string{"Check for sensitive data"},
	}

	ctx := context.Background()
	result, err := analyzer.AnalyzeDocument(ctx, req)

	if err != nil {
		t.Fatalf("AnalyzeDocument() error = %v", err)
	}

	if len(result.Findings) != 1 {
		t.Errorf("AnalyzeDocument() findings count = %d, want 1", len(result.Findings))
	}

	if result.Findings[0].RuleID != "test-finding" {
		t.Errorf("Finding RuleID = %q, want %q", result.Findings[0].RuleID, "test-finding")
	}

	if result.Summary != "Found 1 issue" {
		t.Errorf("Summary = %q, want %q", result.Summary, "Found 1 issue")
	}

	if result.TokensUsed != 100 {
		t.Errorf("TokensUsed = %d, want 100", result.TokensUsed)
	}
}

func TestValidateFindings(t *testing.T) {
	validationJSON := `{"valid_findings": ["finding_0"]}`

	service := &MockAnalyzerService{
		enabled: true,
		response: &CompletionResponse{
			Text: validationJSON,
		},
	}

	analyzer := NewAnalyzer(service)

	findings := []engine.Finding{
		{RuleID: "rule-1", Severity: "high", Line: 1, Context: "test"},
		{RuleID: "rule-2", Severity: "low", Line: 2, Context: "test2"},
	}

	ctx := context.Background()
	validated, err := analyzer.ValidateFindings(ctx, findings, "test text", "test.txt")

	if err != nil {
		t.Fatalf("ValidateFindings() error = %v", err)
	}

	// Should return only the first finding (finding_0)
	if len(validated) != 1 {
		t.Errorf("ValidateFindings() returned %d findings, want 1", len(validated))
	}

	if validated[0].RuleID != "rule-1" {
		t.Errorf("Validated finding RuleID = %q, want %q", validated[0].RuleID, "rule-1")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(substr) <= len(s) && containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}