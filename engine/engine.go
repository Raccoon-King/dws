package engine

import (
	"io/ioutil"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/sirupsen/logrus"
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
	FileID      string `json:"file_id"`
	RuleID      string `json:"rule_id"`
	Severity    string `json:"severity"`
	Line        int    `json:"line"`
	Context     string `json:"context"`
	Description string `json:"description"`
}

var currentRules []Rule
var debugMode bool

// SetRules replaces the in-memory rule set.
func SetRules(rules []Rule) {
	currentRules = rules
}

// LoadRulesFromFile loads rules from a YAML file without setting them globally.
func LoadRulesFromFile(path string) ([]Rule, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file":  path,
			"error": err,
		}).Error("Failed to read rules file")
		return []Rule{}, err
	}
	var config RulesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		logrus.WithFields(logrus.Fields{
			"file":     path,
			"error":    err,
			"yaml_data": string(data),
		}).Error("Failed to unmarshal YAML rules file")
		return []Rule{}, err
	}
	return config.Rules, nil
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
				logrus.WithFields(logrus.Fields{
					"rule_id":  rule.ID,
					"pattern":  rule.Pattern,
					"error":    err,
				}).Warn("Failed to compile regex for rule")
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

// LoadRulesFromYAML loads rules from a YAML file and sets them globally.
// This is called during initialization.
func LoadRulesFromYAML(path string) error {
	rules, err := LoadRulesFromFile(path)
	if err != nil {
		return err
	}
	SetRules(rules)
	return nil
}

// SetDebugMode sets the debug mode for the engine.
func SetDebugMode(debug bool) {
	debugMode = debug
}

// GetDebugMode returns the current debug mode.
func GetDebugMode() bool {
	return debugMode
}
