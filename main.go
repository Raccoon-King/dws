package main

import (
	"log"
	"net/http"
	"os"

	"dws/api"
	"dws/engine"
)

func newServer() (*http.Server, error) {
	rules := os.Getenv("RULES_FILE")
	if rules == "" {
		rules = "/etc/dws/rules.yaml"
	}
	if err := engine.LoadRulesFromYAML(rules); err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/scan", api.ScanHandler)
	mux.HandleFunc("/report", api.ReportHandler)
	mux.HandleFunc("/rules/reload", api.ReloadRulesHandler)
	mux.HandleFunc("/rules/load", api.LoadRulesFromFileHandler)
	mux.HandleFunc("/health", api.HealthHandler)
	return &http.Server{Addr: ":8080", Handler: mux}, nil
}

func run() error {
	srv, err := newServer()
	if err != nil {
		return err
	}
	return srv.ListenAndServe()
}

func main() {
	log.Fatal(run())
}
