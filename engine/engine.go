package engine

import (
	"regexp"
	"strings"
)

// Rule defines a pattern that will be searched in text.
type Rule struct {
	ID       string `json:"id"`
	Pattern  string `json:"pattern"`
	Severity string `json:"severity"`
	Description string `json:"description"`
}

// RulesConfig represents the YAML structure for rules configuration
type RulesConfig struct {
	Rules []Rule `json:"rules" yaml:"rules"`
}

// Finding represents a rule match inside a document.
type Finding struct {
	FileID   string `json:"file_id"`
	RuleID   string `json:"rule_id"`
	Severity string `json:"severity"`
	Line     int    `json:"line"`
	Context  string `json:"context"`
}

var currentRules []Rule

// SetRules replaces the in-memory rule set.
func SetRules(rules []Rule) {
	currentRules = rules
}

// LoadRulesFromFile loads rules from a YAML file without setting them globally.
// TODO: Implement proper YAML parsing
func LoadRulesFromFile(path string) ([]Rule, error) {
	// Temporary implementation - returns empty rules for testing
	// In production, this would parse the YAML file
	return []Rule{}, nil
}

// GetRules returns the current in-memory rule set.
func GetRules() []Rule {
	return currentRules
}

// Evaluate scans the provided text and returns findings for the current rules.
func Evaluate(text, fileID string, rules []Rule) []Finding {
	var findings []Finding
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		for _, rule := range rules {
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				continue
			}
			if re.MatchString(line) {
				findings = append(findings, Finding{
					FileID:   fileID,
					RuleID:   rule.ID,
					Severity: rule.Severity,
					Line:     i + 1,
					Context:  line,
				})
			}
		}
	}
	return findings
}
