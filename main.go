package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"dws/api"
	"dws/engine"
)

var debugMode bool

func initLogging() {
	logOutput := os.Getenv("LOGGING")
	if logOutput == "stdout" {
		log.SetOutput(os.Stdout)
	} else if logOutput == "file" {
		logFile, err := os.OpenFile("dws.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		log.SetOutput(logFile)
	} else {
		// Default to stderr
		log.SetOutput(os.Stderr)
	}

	debugEnv := os.Getenv("DEBUG")
	if debugEnv == "true" {
		debugMode = true
	}
	log.Printf("DEBUG_MODE: %t", debugMode)
}

func NewServer(rulesFile string) (*http.Server, error) {
	if rulesFile != "" {
		if err := engine.LoadRulesFromYAML(rulesFile); err != nil {
			return nil, fmt.Errorf("failed to load rules from %s: %w", rulesFile, err)
		}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port to match Docker/K8s configs
	}
	
	mux := http.NewServeMux()
	mux.HandleFunc("/scan", api.ScanHandler)
	mux.HandleFunc("/rules/reload", api.ReloadRulesHandler)
	mux.HandleFunc("/rules/load", api.LoadRulesFromFileHandler)
	mux.HandleFunc("/ruleset", api.RulesetHandler)
	mux.HandleFunc("/health", api.HealthHandler)
	mux.HandleFunc("/docs", api.DocsHandler)
	return &http.Server{Addr: ":" + port, Handler: mux}, nil
}



func run() error {
	rulesFile := os.Getenv("RULES_FILE")
	if rulesFile == "" {
		rulesFile = "rules.yaml" // Default rules file
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
		log.Fatal(err)
	}
}
