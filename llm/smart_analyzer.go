package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"dws/engine"
)

// SmartAnalyzer optimizes LLM usage by using rules as pre-filters
type SmartAnalyzer struct {
	analyzer *Analyzer
	config   SmartAnalysisConfig
}

// SmartAnalysisConfig controls when LLM analysis is triggered
type SmartAnalysisConfig struct {
	// Only use LLM if regex finds at least this many findings
	MinFindingsThreshold int `yaml:"min_findings_threshold"`

	// Only use LLM for these severity levels
	TriggerSeverities []string `yaml:"trigger_severities"`

	// Skip LLM for documents shorter than this
	MinDocumentLength int `yaml:"min_document_length"`

	// Maximum document length to send to LLM (token limit)
	MaxDocumentLength int `yaml:"max_document_length"`

	// Only analyze these rule categories with LLM
	AnalyzeRuleTypes []string `yaml:"analyze_rule_types"`

	// Skip LLM if confidence in regex results is high
	SkipIfHighConfidence bool `yaml:"skip_if_high_confidence"`
}

// SmartAnalysisResult combines regex pre-filtering with selective LLM usage
type SmartAnalysisResult struct {
	RegexFindings    []engine.Finding `json:"regex_findings"`
	LLMUsed          bool             `json:"llm_used"`
	LLMFindings      []LLMFinding     `json:"llm_findings,omitempty"`
	ValidatedFindings []engine.Finding `json:"validated_findings"`
	TokensUsed       int              `json:"tokens_used"`
	CostSavings      string           `json:"cost_savings,omitempty"`
	AnalysisReason   string           `json:"analysis_reason"`
}

// NewSmartAnalyzer creates an analyzer that uses rules to optimize LLM usage
func NewSmartAnalyzer(analyzer *Analyzer, config SmartAnalysisConfig) *SmartAnalyzer {
	// Set sensible defaults
	if config.MinFindingsThreshold == 0 {
		config.MinFindingsThreshold = 1
	}
	if len(config.TriggerSeverities) == 0 {
		config.TriggerSeverities = []string{"high", "medium"}
	}
	if config.MinDocumentLength == 0 {
		config.MinDocumentLength = 100
	}
	if config.MaxDocumentLength == 0 {
		config.MaxDocumentLength = 8000 // ~3000 tokens
	}

	return &SmartAnalyzer{
		analyzer: analyzer,
		config:   config,
	}
}

// AnalyzeWithPrefiltering performs intelligent analysis using rules as filters
func (s *SmartAnalyzer) AnalyzeWithPrefiltering(ctx context.Context, text, filename string, rules []engine.Rule) (*SmartAnalysisResult, error) {
	result := &SmartAnalysisResult{
		RegexFindings: []engine.Finding{},
		LLMUsed:       false,
		TokensUsed:    0,
	}

	// Step 1: Always run regex analysis first (fast and cheap)
	regexFindings := engine.Evaluate(text, filename, rules)
	result.RegexFindings = regexFindings

	logrus.WithFields(logrus.Fields{
		"filename":      filename,
		"regex_findings": len(regexFindings),
		"doc_length":    len(text),
	}).Debug("Regex pre-filtering complete")

	// Step 2: Decide if LLM analysis is warranted
	shouldUseLLM, reason := s.shouldUseLLM(text, regexFindings)
	result.AnalysisReason = reason

	if !shouldUseLLM {
		result.ValidatedFindings = regexFindings
		result.CostSavings = "100% - LLM not needed"
		logrus.WithFields(logrus.Fields{
			"filename": filename,
			"reason":   reason,
		}).Info("Skipping LLM analysis")
		return result, nil
	}

	// Step 3: Use LLM for validation/enhancement
	if s.analyzer != nil && s.analyzer.service != nil && s.analyzer.service.IsEnabled() {
		result.LLMUsed = true

		// Truncate document if too long
		analysisText := s.prepareTextForLLM(text)

		// Create focused analysis request based on regex findings
		analysisReq := s.createFocusedAnalysisRequest(analysisText, filename, regexFindings)

		llmResponse, err := s.analyzer.AnalyzeDocument(ctx, analysisReq)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"filename": filename,
				"error":    err,
			}).Warn("LLM analysis failed, using regex results")
			result.ValidatedFindings = regexFindings
			return result, nil
		}

		result.LLMFindings = llmResponse.Findings
		result.TokensUsed = llmResponse.TokensUsed

		// Validate regex findings with LLM
		validatedFindings, err := s.analyzer.ValidateFindings(ctx, regexFindings, analysisText, filename)
		if err != nil {
			result.ValidatedFindings = regexFindings
		} else {
			result.ValidatedFindings = validatedFindings
		}

		// Calculate approximate cost savings
		fullDocTokens := len(text) / 4 // Rough token estimate
		actualTokens := result.TokensUsed
		if fullDocTokens > actualTokens {
			savings := float64(fullDocTokens-actualTokens) / float64(fullDocTokens) * 100
			result.CostSavings = fmt.Sprintf("%.1f%% vs full document analysis", savings)
		}
	} else {
		result.ValidatedFindings = regexFindings
		result.CostSavings = "100% - LLM disabled"
	}

	return result, nil
}

// shouldUseLLM determines if LLM analysis is warranted based on regex results
func (s *SmartAnalyzer) shouldUseLLM(text string, findings []engine.Finding) (bool, string) {
	// Check document length
	if len(text) < s.config.MinDocumentLength {
		return false, fmt.Sprintf("Document too short (%d chars < %d min)", len(text), s.config.MinDocumentLength)
	}

	// Check if any findings meet threshold
	if len(findings) < s.config.MinFindingsThreshold {
		return false, fmt.Sprintf("Insufficient findings (%d < %d threshold)", len(findings), s.config.MinFindingsThreshold)
	}

	// Check severity levels
	hasTargetSeverity := false
	for _, finding := range findings {
		for _, targetSeverity := range s.config.TriggerSeverities {
			if strings.EqualFold(finding.Severity, targetSeverity) {
				hasTargetSeverity = true
				break
			}
		}
		if hasTargetSeverity {
			break
		}
	}

	if !hasTargetSeverity {
		return false, fmt.Sprintf("No findings match target severities: %v", s.config.TriggerSeverities)
	}

	// Check rule types (if configured)
	if len(s.config.AnalyzeRuleTypes) > 0 {
		hasTargetRuleType := false
		for _, finding := range findings {
			for _, targetType := range s.config.AnalyzeRuleTypes {
				if strings.Contains(finding.RuleID, targetType) {
					hasTargetRuleType = true
					break
				}
			}
			if hasTargetRuleType {
				break
			}
		}

		if !hasTargetRuleType {
			return false, fmt.Sprintf("No findings match target rule types: %v", s.config.AnalyzeRuleTypes)
		}
	}

	return true, fmt.Sprintf("LLM analysis triggered: %d findings with target severity", len(findings))
}

// prepareTextForLLM truncates or focuses text for LLM analysis
func (s *SmartAnalyzer) prepareTextForLLM(text string) string {
	if len(text) <= s.config.MaxDocumentLength {
		return text
	}

	// Truncate but try to keep complete sentences
	truncated := text[:s.config.MaxDocumentLength]

	// Find last complete sentence
	lastPeriod := strings.LastIndex(truncated, ".")
	lastNewline := strings.LastIndex(truncated, "\n")

	cutPoint := lastPeriod
	if lastNewline > lastPeriod {
		cutPoint = lastNewline
	}

	if cutPoint > s.config.MaxDocumentLength/2 {
		truncated = truncated[:cutPoint+1]
	}

	return truncated + "\n\n[Document truncated for analysis]"
}

// createFocusedAnalysisRequest creates an LLM request focused on regex findings
func (s *SmartAnalyzer) createFocusedAnalysisRequest(text, filename string, findings []engine.Finding) AnalysisRequest {
	// Create focused rules based on what regex found
	var focusedRules []string

	ruleCategories := make(map[string]bool)
	for _, finding := range findings {
		// Extract category from rule ID (e.g., "disease-rabies" -> "disease")
		parts := strings.Split(finding.RuleID, "-")
		if len(parts) > 0 {
			category := parts[0]
			if !ruleCategories[category] {
				ruleCategories[category] = true

				// Map categories to focused prompts
				switch category {
				case "disease":
					focusedRules = append(focusedRules, "Validate disease symptoms and assess public health risk severity")
				case "aggressive":
					focusedRules = append(focusedRules, "Analyze behavioral context - distinguish defensive vs truly aggressive behavior")
				case "property":
					focusedRules = append(focusedRules, "Assess property damage severity and determine intervention level needed")
				case "parasite":
					focusedRules = append(focusedRules, "Evaluate parasite exposure risk and contamination concerns")
				default:
					focusedRules = append(focusedRules, fmt.Sprintf("Analyze and validate %s-related findings", category))
				}
			}
		}
	}

	if len(focusedRules) == 0 {
		focusedRules = []string{"Validate and provide context for the flagged content"}
	}

	return AnalysisRequest{
		Text:     text,
		Filename: filename,
		Rules:    focusedRules,
		Context:  fmt.Sprintf("Focus on validating %d regex findings", len(findings)),
	}
}

// GetOptimizationStats returns statistics about LLM usage optimization
func (s *SmartAnalyzer) GetOptimizationStats() map[string]interface{} {
	return map[string]interface{}{
		"config": s.config,
		"optimization_strategies": []string{
			"Document length filtering",
			"Findings threshold gating",
			"Severity-based triggering",
			"Rule category focusing",
			"Text truncation for large documents",
		},
	}
}