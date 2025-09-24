package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"dws/engine"
)

// LLMService interface for dependency injection in tests
type LLMService interface {
	Complete(ctx context.Context, prompt string) (*CompletionResponse, error)
	IsEnabled() bool
}

// Analyzer provides LLM-powered document analysis capabilities
type Analyzer struct {
	service LLMService
}

// NewAnalyzer creates a new LLM analyzer
func NewAnalyzer(service LLMService) *Analyzer {
	return &Analyzer{
		service: service,
	}
}

// AnalysisRequest represents a request for LLM-based document analysis
type AnalysisRequest struct {
	Text     string   `json:"text"`
	Filename string   `json:"filename"`
	Rules    []string `json:"rules,omitempty"`    // Optional rule descriptions
	Context  string   `json:"context,omitempty"`  // Optional context about the document
}

// AnalysisResponse represents the response from LLM analysis
type AnalysisResponse struct {
	Findings     []LLMFinding `json:"findings"`
	Summary      string       `json:"summary,omitempty"`
	Confidence   float32      `json:"confidence"`
	TokensUsed   int          `json:"tokens_used"`
	Model        string       `json:"model"`
	Provider     Provider     `json:"provider"`
}

// LLMFinding represents a finding from LLM analysis
type LLMFinding struct {
	RuleID      string  `json:"rule_id"`
	Severity    string  `json:"severity"`
	Line        int     `json:"line"`
	Context     string  `json:"context"`
	Description string  `json:"description"`
	Confidence  float32 `json:"confidence"`
	Reasoning   string  `json:"reasoning,omitempty"`
}

// AnalyzeDocument performs comprehensive document analysis using LLM
func (a *Analyzer) AnalyzeDocument(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	if !a.service.IsEnabled() {
		return nil, fmt.Errorf("LLM service is not enabled")
	}

	prompt := a.buildAnalysisPrompt(req)

	response, err := a.service.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	// Parse the LLM response
	analysisResp, err := a.parseAnalysisResponse(response.Text, req.Filename)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"filename": req.Filename,
			"error":    err,
			"response": response.Text,
		}).Warn("Failed to parse LLM analysis response, returning raw analysis")

		// Fallback: return a single finding with the raw analysis
		analysisResp = &AnalysisResponse{
			Findings: []LLMFinding{
				{
					RuleID:      "llm-analysis",
					Severity:    "info",
					Line:        1,
					Context:     truncateString(req.Text, 200),
					Description: "LLM document analysis",
					Confidence:  0.5,
					Reasoning:   response.Text,
				},
			},
			Summary:    response.Text,
			Confidence: 0.5,
		}
	}

	analysisResp.TokensUsed = response.TokensUsed
	analysisResp.Model = response.Model
	analysisResp.Provider = response.Provider

	return analysisResp, nil
}

// ValidateFindings compares regex findings with LLM analysis to reduce false positives
func (a *Analyzer) ValidateFindings(ctx context.Context, findings []engine.Finding, text string, filename string) ([]engine.Finding, error) {
	if !a.service.IsEnabled() || len(findings) == 0 {
		return findings, nil
	}

	prompt := a.buildValidationPrompt(findings, text, filename)

	response, err := a.service.Complete(ctx, prompt)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"filename": filename,
			"error":    err,
		}).Warn("LLM validation failed, returning original findings")
		return findings, nil
	}

	// Parse validation response
	validatedFindings, err := a.parseValidationResponse(response.Text, findings)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"filename": filename,
			"error":    err,
		}).Warn("Failed to parse validation response, returning original findings")
		return findings, nil
	}

	return validatedFindings, nil
}

// buildAnalysisPrompt creates a prompt for document analysis
func (a *Analyzer) buildAnalysisPrompt(req AnalysisRequest) string {
	var sb strings.Builder

	sb.WriteString("You are a security analyst reviewing documents for sensitive information and policy violations.\n\n")

	if req.Context != "" {
		sb.WriteString(fmt.Sprintf("Document context: %s\n\n", req.Context))
	}

	sb.WriteString("Please analyze the following document and identify any findings. ")
	sb.WriteString("Return your analysis as a JSON object with this exact structure:\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"findings\": [\n")
	sb.WriteString("    {\n")
	sb.WriteString("      \"rule_id\": \"unique_identifier\",\n")
	sb.WriteString("      \"severity\": \"high|medium|low\",\n")
	sb.WriteString("      \"line\": line_number,\n")
	sb.WriteString("      \"context\": \"the_matching_text\",\n")
	sb.WriteString("      \"description\": \"what_was_found\",\n")
	sb.WriteString("      \"confidence\": 0.0_to_1.0,\n")
	sb.WriteString("      \"reasoning\": \"why_this_is_a_finding\"\n")
	sb.WriteString("    }\n")
	sb.WriteString("  ],\n")
	sb.WriteString("  \"summary\": \"brief_overall_assessment\",\n")
	sb.WriteString("  \"confidence\": 0.0_to_1.0\n")
	sb.WriteString("}\n\n")

	if len(req.Rules) > 0 {
		sb.WriteString("Focus on these specific areas:\n")
		for i, rule := range req.Rules {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, rule))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("Look for:\n")
		sb.WriteString("- Personally identifiable information (PII)\n")
		sb.WriteString("- Credit card numbers, SSNs, phone numbers\n")
		sb.WriteString("- API keys, passwords, tokens\n")
		sb.WriteString("- Confidential or sensitive business information\n")
		sb.WriteString("- Compliance violations\n\n")
	}

	sb.WriteString(fmt.Sprintf("Document to analyze (%s):\n", req.Filename))
	sb.WriteString("---\n")
	sb.WriteString(truncateString(req.Text, 8000)) // Limit text to avoid token limits
	sb.WriteString("\n---\n\n")
	sb.WriteString("Provide your analysis as valid JSON:")

	return sb.String()
}

// buildValidationPrompt creates a prompt for validating existing findings
func (a *Analyzer) buildValidationPrompt(findings []engine.Finding, text string, filename string) string {
	var sb strings.Builder

	sb.WriteString("You are validating security findings from an automated scanner. ")
	sb.WriteString("Review each finding and determine if it's a true positive or false positive.\n\n")

	sb.WriteString("Return a JSON array of finding IDs that should be KEPT (true positives):\n")
	sb.WriteString("{ \"valid_findings\": [\"finding_1\", \"finding_2\"] }\n\n")

	sb.WriteString("Original findings to validate:\n")
	for i, finding := range findings {
		sb.WriteString(fmt.Sprintf("%d. ID: finding_%d, Rule: %s, Severity: %s, Line: %d, Context: %s\n",
			i+1, i, finding.RuleID, finding.Severity, finding.Line, finding.Context))
	}

	sb.WriteString(fmt.Sprintf("\nDocument context (%s):\n", filename))
	sb.WriteString("---\n")
	sb.WriteString(truncateString(text, 4000))
	sb.WriteString("\n---\n\n")
	sb.WriteString("Return JSON with valid finding IDs:")

	return sb.String()
}

// parseAnalysisResponse parses the LLM analysis response
func (a *Analyzer) parseAnalysisResponse(responseText, filename string) (*AnalysisResponse, error) {
	// Try to extract JSON from the response
	jsonStr := extractJSON(responseText)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var resp AnalysisResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Validate and set defaults
	for i := range resp.Findings {
		if resp.Findings[i].RuleID == "" {
			resp.Findings[i].RuleID = fmt.Sprintf("llm-finding-%d", i+1)
		}
		if resp.Findings[i].Severity == "" {
			resp.Findings[i].Severity = "medium"
		}
		if resp.Findings[i].Line <= 0 {
			resp.Findings[i].Line = 1
		}
		if resp.Findings[i].Confidence <= 0 {
			resp.Findings[i].Confidence = 0.7
		}
	}

	if resp.Confidence <= 0 {
		resp.Confidence = 0.7
	}

	return &resp, nil
}

// parseValidationResponse parses the validation response
func (a *Analyzer) parseValidationResponse(responseText string, originalFindings []engine.Finding) ([]engine.Finding, error) {
	jsonStr := extractJSON(responseText)
	if jsonStr == "" {
		return originalFindings, fmt.Errorf("no JSON found in validation response")
	}

	var validationResp struct {
		ValidFindings []string `json:"valid_findings"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &validationResp); err != nil {
		return originalFindings, fmt.Errorf("failed to unmarshal validation JSON: %w", err)
	}

	// Create map of valid finding IDs
	validMap := make(map[string]bool)
	for _, id := range validationResp.ValidFindings {
		validMap[id] = true
	}

	// Filter findings
	var validatedFindings []engine.Finding
	for i, finding := range originalFindings {
		findingID := fmt.Sprintf("finding_%d", i)
		if validMap[findingID] {
			validatedFindings = append(validatedFindings, finding)
		}
	}

	return validatedFindings, nil
}

// extractJSON attempts to extract JSON from a text response
func extractJSON(text string) string {
	// Look for JSON object or array
	start := strings.Index(text, "{")
	if start == -1 {
		start = strings.Index(text, "[")
	}
	if start == -1 {
		return ""
	}

	// Find the matching closing brace/bracket
	var end int
	openChar := text[start]
	var closeChar byte
	if openChar == '{' {
		closeChar = '}'
	} else {
		closeChar = ']'
	}

	depth := 0
	for i := start; i < len(text); i++ {
		if text[i] == openChar {
			depth++
		} else if text[i] == closeChar {
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
	}

	if end == 0 {
		return ""
	}

	return text[start:end]
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "... [truncated]"
}

// ConvertLLMFindingsToEngine converts LLM findings to engine findings
func ConvertLLMFindingsToEngine(llmFindings []LLMFinding, fileID string) []engine.Finding {
	var findings []engine.Finding

	for _, llmFinding := range llmFindings {
		finding := engine.Finding{
			FileID:   fileID,
			RuleID:   llmFinding.RuleID,
			Severity: llmFinding.Severity,
			Line:     llmFinding.Line,
			Context:  llmFinding.Context,
		}
		findings = append(findings, finding)
	}

	return findings
}