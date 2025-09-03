package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"

	"dws/api"
	"dws/engine"
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port to match Docker/K8s configs
	}
	
	recoveryMiddleware := func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err,
						"url":   r.URL.Path,
						"method": r.Method,
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
	mux.HandleFunc("/rules/reload", api.ReloadRulesHandler)
	mux.HandleFunc("/rules/load", api.LoadRulesFromFileHandler)
	mux.HandleFunc("/ruleset", api.RulesetHandler)
	mux.HandleFunc("/health", api.HealthHandler)
	mux.HandleFunc("/docs", api.DocsHandler)
	return &http.Server{Addr: ":" + port, Handler: recoveryMiddleware(mux)}, nil
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
