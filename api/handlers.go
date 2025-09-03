package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"dws/engine"
	"dws/scanner"
)

// ErrorResponse sends an error response with the specified status code and message.
func ErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

var rulesFile string

// SetRulesFile sets the rules file path for the api package.
func SetRulesFile(path string) {
	rulesFile = path
}

type Report struct {
	FileID   string          `json:"fileID"`
	Findings []engine.Finding `json:"findings"`
}

// EndpointDoc represents the documentation for a single API endpoint.
type EndpointDoc struct {
	Path        string       `json:"path"`
	Method      string       `json:"method"`
	Description string       `json:"description"`
	DataShapes  []DataShape  `json:"data_shapes"`
	CurlExample string       `json:"curl_example"`
}

// DataShape represents the structure of a request or response body.
type DataShape struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Shape       string `json:"shape"`
}

// DocsHandler returns a JSON array of all available endpoints and their documentation.
func DocsHandler(w http.ResponseWriter, r *http.Request) {
	docs := []EndpointDoc{
		{
			Path:        "/scan",
			Method:      "POST",
			Description: "Upload a document to be scanned and receive a structured report of findings including rule descriptions.",
			DataShapes: []DataShape{
				{
					Name:        "Request",
					Description: "multipart/form-data",
					Shape:       `{"file": "<file>"}`,
				},
				{
					Name:        "Response",
					Description: "A structured report of findings.",
					Shape:       `{"file_id":"uploaded-filename","findings":[{"rule_id":"rule-1","severity":"high","line":3,"context":"line containing match","description":"rule description"}]}`,
				},
			},
			CurlExample: `curl -X POST -F 'file=@/path/to/your/file.pdf' http://localhost:8080/scan`,
		},
		{
			Path:        "/rules/reload",
			Method:      "POST",
			Description: "Replace the existing rules with a new set.",
			DataShapes: []DataShape{
				{
					Name:        "Request",
					Description: "A JSON object containing the new rules.",
					Shape:       `{"rules":[{"id":"rule-1","pattern":"secret","severity":"high"}]}`,
				},
			},
			CurlExample: `curl -X POST -H "Content-Type: application/json" -d '{\"rules\":[{\"id\":\"rule-1\",\"pattern\":\"secret\",\"severity\":\"high\"}]}' http://localhost:8080/rules/reload`,
		},
		{
			Path:        "/rules/load",
			Method:      "POST",
			Description: "Load rules from a YAML file on disk.",
			DataShapes: []DataShape{
				{
					Name:        "Request",
					Description: "A JSON object containing the path to the rules file.",
					Shape:       `{"path":"/etc/dws/rules.yaml"}`,
				},
			},
			CurlExample: `curl -X POST -H "Content-Type: application/json" -d '{\"path\":\"/etc/dws/rules.yaml\"}' http://localhost:8080/rules/load`,
		},
		{
			Path:        "/health",
			Method:      "GET",
			Description: "Health check endpoint.",
			DataShapes: []DataShape{
				{
					Name:        "Response",
					Description: "A JSON object indicating the status of the service.",
					Shape:       `{"status":"ok"}`,
				},
			},
			CurlExample: `curl http://localhost:8080/health`,
		},
		{
			Path:        "/ruleset?rule",
			Method:      "POST",
			Description: "Scan a document against a specific ruleset specified by the 'rule' query parameter (expects rules/{rule}.yaml).",
			DataShapes: []DataShape{
				{
					Name:        "Request",
					Description: "multipart/form-data with 'rule' query parameter",
					Shape:       `{"file": "<file>"}`,
				},
				{
					Name:        "Response",
					Description: "A structured report of findings for the specified ruleset.",
					Shape:       `{"file_id":"uploaded-filename","findings":[{"rule_id":"rule-1","severity":"high","line":3,"context":"line containing match","description":"rule description"}]}`,
				},
			},
			CurlExample: `curl -X POST -F 'file=@/path/to/your/file.pdf' 'http://localhost:8080/ruleset?rule=customrules'`,
		},
		{
			Path:        "/docs",
			Method:      "GET",
			Description: "Returns a JSON array of all available endpoints and their documentation.",
			CurlExample: `curl http://localhost:8080/docs`,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

// RulesetHandler handles scanning a document against a specific ruleset.
func RulesetHandler(w http.ResponseWriter, r *http.Request) {
	rule := r.URL.Query().Get("rule")
	if rule == "" {
		ErrorResponse(w, http.StatusBadRequest, "missing rule query parameter")
		return
	}
	// Prevent path traversal attacks by ensuring rule doesn't contain invalid characters
	if strings.ContainsAny(rule, "/\\..") {
		ErrorResponse(w, http.StatusBadRequest, "invalid rule name")
		return
	}

	path := "rules/" + rule + ".yaml"

	rules, err := engine.LoadRulesFromFile(path)
	if err != nil {
		log.Printf("Error loading rules from file %s: %v", path, err)
		ErrorResponse(w, http.StatusInternalServerError, "failed to load ruleset")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid multipart")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "missing file")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "read error")
		return
	}
	text, err := scanner.ExtractText(data, header.Filename)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "unsupported file")
		return
	}
	findings := engine.Evaluate(text, header.Filename, rules)
	// Debug mode is available via engine.GetDebugMode if implemented
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Report{FileID: header.Filename, Findings: findings})
}

// ScanHandler ingests text and returns findings.
func ScanHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid multipart")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "missing file")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "read error")
		return
	}
	text, err := scanner.ExtractText(data, header.Filename)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "unsupported file")
		return
	}
	findings := engine.Evaluate(text, header.Filename, engine.GetRules())
	if engine.GetDebugMode() {
		log.Printf("API_DEBUG: Findings before encoding: %+v", findings)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Report{FileID: header.Filename, Findings: findings})
}

// ReloadRulesHandler replaces the current rule set.
func ReloadRulesHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Rules []engine.Rule `json:"rules"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid request")
		return
	}

	for i := range req.Rules {
		compiled, err := regexp.Compile(req.Rules[i].Pattern)
		if err != nil {
			ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to compile regex for rule %s: %v", req.Rules[i].ID, err))
			return
		}
		req.Rules[i].CompiledPattern = compiled
	}

	engine.SetRules(req.Rules)
	w.WriteHeader(http.StatusOK)
}

// LoadRulesFromFileHandler loads rules from a file specified in the request body.
func LoadRulesFromFileHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Path string `json:"path"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Path == "" {
		ErrorResponse(w, http.StatusBadRequest, "missing path parameter")
		return
	}

	// Clean the path to prevent path traversal attacks.
	path := filepath.Clean(req.Path)
	if strings.HasPrefix(path, "..") {
		ErrorResponse(w, http.StatusBadRequest, "invalid path")
		return
	}

	if err := engine.LoadRulesFromYAML(path); err != nil {
		log.Printf("Error loading rules from file %s: %v", path, err)
		ErrorResponse(w, http.StatusInternalServerError, "failed to load rules file")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "rules loaded successfully"})
}

// HealthHandler reports service health.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the rules file is readable
	if _, err := os.Stat(rulesFile); err != nil {
		ErrorResponse(w, http.StatusServiceUnavailable, "rules file not readable")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
