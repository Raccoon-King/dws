package engine

import (
	"os"
	"regexp"
	"strings"
)

// Rule defines a pattern that will be searched in text.
type Rule struct {
	ID          string `json:"id"`
	Pattern     string `json:"pattern"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
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
	var r Rule
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "- id:"):
			if r.ID != "" && r.Pattern != "" {
				rules = append(rules, r)
				r = Rule{}
			}
			r.ID = strings.TrimSpace(strings.TrimPrefix(line, "- id:"))
		case strings.HasPrefix(line, "id:"):
			if r.ID != "" && r.Pattern != "" {
				rules = append(rules, r)
				r = Rule{}
			}
			r.ID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
		case strings.HasPrefix(line, "pattern:"):
			r.Pattern = strings.TrimSpace(strings.TrimPrefix(line, "pattern:"))
		case strings.HasPrefix(line, "severity:"):
			r.Severity = strings.TrimSpace(strings.TrimPrefix(line, "severity:"))
		case strings.HasPrefix(line, "description:"):
			r.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}
	}
	if r.ID != "" && r.Pattern != "" {
		rules = append(rules, r)
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
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				continue
			}
			if re.MatchString(line) {
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
