package api

import (
	"encoding/json"
	"net/http"

	"dws/engine"
)

// ScanHandler ingests text and returns findings.
func ScanHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		FileID string `json:"file_id"`
		Text   string `json:"text"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	findings := engine.Evaluate(req.Text, req.FileID, engine.GetRules())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(findings)
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

// HealthHandler reports service health.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
