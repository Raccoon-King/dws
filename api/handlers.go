package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"

	"dws/engine"
	"dws/scanner"
)

type report struct {
	FileID   string          `json:"fileID"`
	Findings []engine.Finding `json:"findings"`
}
func ScanHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "invalid multipart", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}
	text, err := scanner.ExtractText(data, header.Filename)
	if err != nil {
		http.Error(w, "unsupported file", http.StatusBadRequest)
		return
	}
	findings := engine.Evaluate(text, header.Filename, engine.GetRules())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(findings)
}

// ProcessDocumentHandler accepts a document, scans it, and returns findings in a report structure.
func ProcessDocumentHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "invalid multipart", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}
	text, err := scanner.ExtractText(data, header.Filename)
	if err != nil {
		http.Error(w, "unsupported file", http.StatusBadRequest)
		return
	}
	findings := engine.Evaluate(text, header.Filename, engine.GetRules())
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report{FileID: header.Filename, Findings: findings})
}

// ReloadRulesHandler replaces the current rule set.
func ReloadRulesHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Rules []engine.Rule `json:"rules"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	for i := range req.Rules {
		compiled, err := regexp.Compile(".*" + req.Rules[i].Pattern + ".*")
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to compile regex for rule %s: %v", req.Rules[i].ID, err), http.StatusBadRequest)
			return
		}
		req.Rules[i].CompiledPattern = compiled
	}

	log.Printf("API_DEBUG: Calling engine.SetRules with %d rules. First rule CompiledPattern is nil: %t", len(req.Rules), req.Rules[0].CompiledPattern == nil)
	engine.SetRules(req.Rules)
	w.WriteHeader(http.StatusOK)
}

// LoadRulesFromFileHandler loads rules from a YAML file on disk.
func LoadRulesFromFileHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Path string `json:"path"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := engine.LoadRulesFromYAML(req.Path); err != nil {
		log.Printf("Error loading rules from file %s: %v", req.Path, err)
		http.Error(w, "load error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HealthHandler reports service health.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}