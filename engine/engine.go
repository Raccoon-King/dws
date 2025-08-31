package engine

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"dws/logging"

	yaml "gopkg.in/yaml.v2"
)

// Rule defines a pattern that will be searched in text.
type Rule struct {
	ID              string         `json:"id" yaml:"id"`
	Pattern         string         `json:"pattern" yaml:"pattern"`
	Severity        string         `json:"severity" yaml:"severity"`
	Description     string         `json:"description" yaml:"description"`
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
type RulesConfig struct {
	Rules []Rule `yaml:"rules"`
}

// ValidateRules ensures each rule has a unique, non-empty ID.
func ValidateRules(rules []Rule) error {
	seen := make(map[string]struct{})
	validSev := map[string]struct{}{"low": {}, "medium": {}, "high": {}, "informational": {}}
	for _, r := range rules {
		if strings.TrimSpace(r.ID) == "" {
			return fmt.Errorf("rule ID cannot be empty")
		}
		if _, ok := seen[r.ID]; ok {
			return fmt.Errorf("duplicate rule ID: %s", r.ID)
		}
		if _, ok := validSev[strings.ToLower(r.Severity)]; !ok {
			return fmt.Errorf("invalid severity for rule %s: %s", r.ID, r.Severity)
		}
		seen[r.ID] = struct{}{}
	}
	return nil
}

func LoadRulesFromYAML(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var config RulesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	if err := ValidateRules(config.Rules); err != nil {
		return err
	}
	for i := range config.Rules {
		compiled, err := regexp.Compile(config.Rules[i].Pattern)
		if err != nil {
			return fmt.Errorf("failed to compile regex for rule %s: %w", config.Rules[i].ID, err)
		}
		config.Rules[i].CompiledPattern = compiled
	}

	SetRules(config.Rules)
	return nil
}

// Evaluate scans the provided text and returns findings for the current rules.
func Evaluate(text, fileID string, rules []Rule) []Finding {
	logging.Debug("evaluating text", map[string]any{"file": fileID, "rules": len(rules)})
	var findings []Finding
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		logging.Debug("processing line", map[string]any{"line_num": i + 1, "line": line})
		for _, rule := range rules {
			logging.Debug("checking rule", map[string]any{"rule_id": rule.ID, "pattern": rule.Pattern})
			if rule.CompiledPattern == nil {
				logging.Debug("compiled pattern nil", map[string]any{"rule_id": rule.ID})
				continue // Skip this rule if pattern is nil
			}
			if rule.CompiledPattern.MatchString(line) {
				logging.Debug("match found", map[string]any{"rule_id": rule.ID, "line_num": i + 1})
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
	logging.Debug("evaluation complete", map[string]any{"findings": findings})
	return findings
}
