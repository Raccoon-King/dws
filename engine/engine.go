package engine

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// Rule defines a pattern that will be searched in text.
type Rule struct {
	ID          string `json:"id" yaml:"id"`
	Pattern     string `json:"pattern" yaml:"pattern"`
	Severity    string `json:"severity" yaml:"severity"`
	Description string `json:"description" yaml:"description"`
	CompiledPattern *regexp.Regexp `json:"-" yaml:"-"` // Compiled regex for internal use
}

// Finding represents a rule match inside a document.
type Finding struct {
	FileID      string `json:"file_id"`
	RuleID      string `json:"rule_id"`
	Severity    string `json:"severity"`
	Line        int    `json:"line"`
	Context     string `json:"context"`
	Description string `json:"description"`
}

var currentRules []Rule

// SetRules replaces the in-memory rule set.
func SetRules(rules []Rule) {
	currentRules = rules
}

// GetRules returns the current in-memory rule set.
func GetRules() []Rule {
	return currentRules
}

// LoadRulesFromYAML reads rule definitions from a YAML file and replaces the current rule set.
func LoadRulesFromYAML(path string) error {
		data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var rules []Rule
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return err
	}

	for i := range rules {
		compiled, err := regexp.Compile(rules[i].Pattern)
		if err != nil {
			return fmt.Errorf("failed to compile regex for rule %s: %w", rules[i].ID, err)
		}
		rules[i].CompiledPattern = compiled
	}

	SetRules(rules)
	return nil
}

// Evaluate scans the provided text and returns findings for the current rules.
func Evaluate(text, fileID string, rules []Rule) []Finding {
	var findings []Finding
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		for _, rule := range rules {
						if rule.CompiledPattern != nil && rule.CompiledPattern.MatchString(line) {
				findings = append(findings, Finding{
					FileID:      fileID,
					RuleID:      rule.ID,
					Severity:    rule.Severity,
					Line:        i + 1,
					Context:     line,
					Description: rule.Description,
				})
			}
		}
	}
	return findings
}
