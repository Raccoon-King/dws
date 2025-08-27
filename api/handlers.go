package api

import (
	"encoding/json"
	"io"
	"net/http"

	"dws/engine"
	"dws/scanner"
)

// ScanHandler ingests a document file and returns findings.
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

// ReportHandler scans a document and wraps findings in a report structure.
func ReportHandler(w http.ResponseWriter, r *http.Request) {
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
	type report struct {
		FileID   string           `json:"file_id"`
		Findings []engine.Finding `json:"findings"`
	}
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
