package main

import (
	"fmt"
	"net/http"
	"os"

	"dws/api"
	"dws/engine"
	"dws/logging"
)

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
	logging.Init()
	if err := run(); err != nil {
		logging.Error("server failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
}
