package engine

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

var debugMode bool

// SetDebugMode sets the debug mode for the engine package.
func SetDebugMode(mode bool) {
	debugMode = mode
}

// GetDebugMode returns the current debug mode for the engine package.
func GetDebugMode() bool {
	return debugMode
}

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
type RulesConfig struct {
	Rules []Rule `yaml:"rules"`
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
	if debugMode {
		log.Printf("ENGINE_DEBUG: Evaluating text for file %s", fileID)
		log.Printf("ENGINE_DEBUG: Text: %s", text)
		log.Printf("ENGINE_DEBUG: Number of rules: %d", len(rules))
	}
	var findings []Finding
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if debugMode {
			log.Printf("ENGINE_DEBUG: Processing line %d: %s", i+1, line)
		}
		for _, rule := range rules {
			if debugMode {
				log.Printf("ENGINE_DEBUG:   Checking rule ID: %s, Pattern: %s", rule.ID, rule.Pattern)
				log.Printf("ENGINE_DEBUG:   Compiled Pattern: %s", rule.CompiledPattern.String())
			}
			if rule.CompiledPattern == nil {
				if debugMode {
					log.Printf("ENGINE_DEBUG: CompiledPattern is NIL before MatchString for rule %s", rule.ID)
				}
				continue // Skip this rule if pattern is nil
			}
			if debugMode {
				log.Printf("ENGINE_DEBUG:   Attempting MatchString for rule %s on line: %s", rule.ID, line)
			}
			if rule.CompiledPattern.MatchString(line) {
				if debugMode {
					log.Printf("ENGINE_DEBUG:     MATCH FOUND for rule %s on line %d", rule.ID, i+1)
				}
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
	if debugMode {
		log.Printf("ENGINE_DEBUG: Findings before return: %+v", findings)
	}
	return findings
}
