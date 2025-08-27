package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"dws/api"
	"dws/engine"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func newServer() (*http.Server, error) {
	rules := os.Getenv("RULES_FILE")
	if rules == "" {
		rules = "/etc/dws/rules.yaml"
	}
	if err := engine.LoadRulesFromYAML(rules); err != nil {
		return nil, err
	}

	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	if err != nil {
		return nil, fmt.Errorf("error creating S3 session: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/scan", api.ScanHandler)
	mux.HandleFunc("/report", api.ReportHandler)
	mux.HandleFunc("/rules/reload", api.ReloadRulesHandler)
	mux.HandleFunc("/rules/load", api.LoadRulesFromFileHandler)
	mux.HandleFunc("/s3-event", api.S3EventHandler(s3.New(sess)))
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