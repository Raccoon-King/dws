package main

import (
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

func NewServer() (*http.Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scan", api.ScanHandler)
	mux.HandleFunc("/process-document", api.ProcessDocumentHandler)
	mux.HandleFunc("/rules/reload", api.ReloadRulesHandler)
	mux.HandleFunc("/rules/load", api.LoadRulesFromFileHandler)
	mux.HandleFunc("/health", api.HealthHandler)
	return &http.Server{Addr: ":8080", Handler: mux}, nil
}

func run() error {
	srv, err := NewServer()
	if err != nil {
		return err
	}
	return srv.ListenAndServe()
}

func main() {
	initLogging()
	engine.SetDebugMode(debugMode)
	log.Fatal(run())
}
