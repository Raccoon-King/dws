package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"dws/api"
	"dws/engine"
	"dws/llm"
)

var debugMode bool

func initLogging() {
	logOutput := os.Getenv("LOGGING")
	if logOutput == "stdout" {
		logrus.SetOutput(os.Stdout)
	} else if logOutput == "file" {
		logFile, err := os.OpenFile("dws.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logrus.Fatalf("Failed to open log file: %v", err)
		}
		logrus.SetOutput(logFile)
	} else {
		// Default to stderr
		logrus.SetOutput(os.Stderr)
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})

	debugEnv := os.Getenv("DEBUG")
	if debugEnv == "true" {
		debugMode = true
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.Printf("DEBUG_MODE: %t", debugMode)
}

func NewServer(rulesFile string) (*http.Server, error) {
	if rulesFile != "" {
		if err := engine.LoadRulesFromYAML(rulesFile); err != nil {
			return nil, fmt.Errorf("failed to load rules from %s: %w", rulesFile, err)
		}
	}

	// Initialize LLM service
	llmService, err := initLLMService()
	if err != nil {
		logrus.WithError(err).Warn("Failed to initialize LLM service, LLM features will be disabled")
	} else if llmService != nil && llmService.IsEnabled() {
		analyzer := llm.NewAnalyzer(llmService)
		api.SetLLMAnalyzer(analyzer)
		logrus.Info("LLM service initialized successfully")
	} else {
		logrus.Info("LLM service disabled")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port to match Docker/K8s configs
	}

	recoveryMiddleware := func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logrus.WithFields(logrus.Fields{
						"error":      err,
						"url":        r.URL.Path,
						"method":     r.Method,
						"user_agent": r.UserAgent(),
					}).Error("HTTP handler panic recovered")
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			handler.ServeHTTP(w, r)
		})
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/scan", api.ScanHandler)
	mux.HandleFunc("/scan/s3", api.S3ScanHandler)
	mux.HandleFunc("/scan/llm", api.LLMScanHandler)
	mux.HandleFunc("/scan/hybrid", api.HybridScanHandler)
	mux.HandleFunc("/scan/smart", api.SmartScanHandler)
	mux.HandleFunc("/rules/reload", api.ReloadRulesHandler)
	mux.HandleFunc("/rules/load", api.LoadRulesFromFileHandler)
	mux.HandleFunc("/ruleset", api.RulesetHandler)
	mux.HandleFunc("/health", api.HealthHandler)
	mux.HandleFunc("/docs", api.DocsHandler)
	return &http.Server{Addr: ":" + port, Handler: recoveryMiddleware(mux)}, nil
}

// initLLMService initializes the LLM service from configuration
func initLLMService() (*llm.Service, error) {
	// Check if LLM is enabled via environment variable
	if enabled := os.Getenv("LLM_ENABLED"); enabled != "true" {
		return nil, nil // LLM disabled
	}

	// Load LLM configuration
	configFile := os.Getenv("LLM_CONFIG")
	if configFile == "" {
		configFile = "config/llm.yaml" // Default LLM config
	}

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("LLM config file not found: %s", configFile)
	}

	// Read and parse config
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read LLM config: %w", err)
	}

	var config struct {
		LLM     llm.Config        `yaml:"llm"`
		OpenAI  llm.OpenAIConfig  `yaml:"openai"`
		Bedrock llm.BedrockConfig `yaml:"bedrock"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse LLM config: %w", err)
	}

	// Set provider-specific configs
	config.LLM.OpenAI = config.OpenAI
	config.LLM.Bedrock = config.Bedrock

	// Expand environment variables in sensitive fields
	config.LLM.OpenAI.APIKey = os.ExpandEnv(config.LLM.OpenAI.APIKey)
	config.LLM.Bedrock.AccessKeyID = os.ExpandEnv(config.LLM.Bedrock.AccessKeyID)
	config.LLM.Bedrock.SecretAccessKey = os.ExpandEnv(config.LLM.Bedrock.SecretAccessKey)
	config.LLM.Bedrock.SessionToken = os.ExpandEnv(config.LLM.Bedrock.SessionToken)
	config.LLM.Bedrock.RoleARN = os.ExpandEnv(config.LLM.Bedrock.RoleARN)

	// Parse timeout
	if config.LLM.Timeout == 0 {
		config.LLM.Timeout = 30 * time.Second
	}

	// Create LLM service
	service, err := llm.NewService(config.LLM)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM service: %w", err)
	}

	return service, nil
}

func run() error {
	rulesFile := os.Getenv("RULES_FILE")
	if rulesFile == "" {
		rulesFile = "config/default.yaml" // Default rules file
	}
	api.SetRulesFile(rulesFile)
	srv, err := NewServer(rulesFile)
	if err != nil {
		return err
	}
	return srv.ListenAndServe()
}

func main() {
	initLogging()
	engine.SetDebugMode(debugMode)

	if err := run(); err != nil {
		logrus.Fatal(err)
	}
}
