package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"dws/engine"
	"dws/llm"
	"dws/scanner"
	"dws/s3"
)

var rulesFile string
var llmAnalyzer *llm.Analyzer

// SetRulesFile sets the rules file path for the api package.
func SetRulesFile(path string) {
	rulesFile = path
}

// SetLLMAnalyzer sets the LLM analyzer for the api package.
func SetLLMAnalyzer(analyzer *llm.Analyzer) {
	llmAnalyzer = analyzer
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
			Path:        "/scan/s3",
			Method:      "POST",
			Description: "Scan a document from S3 URL. Supports IAM roles and access key authentication.",
			DataShapes: []DataShape{
				{
					Name:        "Request",
					Description: "JSON object with S3 URL and optional authentication parameters",
					Shape:       `{"s3_url":"s3://bucket/path/file.pdf","region":"us-east-1","access_key_id":"optional","secret_access_key":"optional","session_token":"optional","role_arn":"optional"}`,
				},
				{
					Name:        "Response",
					Description: "A structured report of findings from the S3 file.",
					Shape:       `{"file_id":"file.pdf","findings":[{"rule_id":"rule-1","severity":"high","line":3,"context":"line containing match","description":"rule description"}]}`,
				},
			},
			CurlExample: `curl -X POST -H "Content-Type: application/json" -d '{"s3_url":"s3://my-bucket/document.pdf","region":"us-west-2"}' http://localhost:8080/scan/s3`,
		},
		{
			Path:        "/scan/llm",
			Method:      "POST",
			Description: "Upload a document for LLM-powered analysis with semantic understanding.",
			DataShapes: []DataShape{
				{
					Name:        "Request",
					Description: "multipart/form-data with optional analysis rules",
					Shape:       `{"file": "<file>", "rules": ["optional custom rules"]}`,
				},
				{
					Name:        "Response",
					Description: "LLM analysis results with confidence scores and reasoning.",
					Shape:       `{"file_id":"uploaded-filename","findings":[{"rule_id":"llm-finding-1","severity":"high","line":3,"context":"matching text","description":"finding description","confidence":0.9,"reasoning":"why this is a finding"}],"summary":"overall analysis","confidence":0.8,"tokens_used":150,"model":"gpt-3.5-turbo","provider":"openai"}`,
				},
			},
			CurlExample: `curl -X POST -F 'file=@/path/to/your/file.pdf' -F 'rules=["Look for API keys","Check for PII"]' http://localhost:8080/scan/llm`,
		},
		{
			Path:        "/scan/hybrid",
			Method:      "POST",
			Description: "Upload a document for hybrid analysis combining regex rules with LLM validation.",
			DataShapes: []DataShape{
				{
					Name:        "Request",
					Description: "multipart/form-data",
					Shape:       `{"file": "<file>"}`,
				},
				{
					Name:        "Response",
					Description: "Combined analysis with validated findings and LLM insights.",
					Shape:       `{"file_id":"uploaded-filename","regex_findings":[...],"llm_analysis":{...},"validated_findings":[...],"tokens_used":150}`,
				},
			},
			CurlExample: `curl -X POST -F 'file=@/path/to/your/file.pdf' http://localhost:8080/scan/hybrid`,
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
		logrus.WithFields(logrus.Fields{
			"file":  path,
			"error": err,
		}).Error("Failed to load ruleset from file")
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
		logrus.WithFields(logrus.Fields{
			"file_id":  header.Filename,
			"findings": findings,
		}).Debug("Findings before encoding")
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
		_, err := regexp.Compile(req.Rules[i].Pattern)
		if err != nil {
			ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to compile regex for rule %s: %v", req.Rules[i].ID, err))
			return
		}
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
		logrus.WithFields(logrus.Fields{
			"file":  path,
			"error": err,
		}).Error("Failed to load rules from YAML file")
		ErrorResponse(w, http.StatusInternalServerError, "failed to load rules file")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "rules loaded successfully"})
}

// S3ScanRequest represents a request to scan a file from S3
type S3ScanRequest struct {
	S3URL           string `json:"s3_url"`
	Region          string `json:"region,omitempty"`
	AccessKeyID     string `json:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secret_access_key,omitempty"`
	SessionToken    string `json:"session_token,omitempty"`
	RoleARN         string `json:"role_arn,omitempty"`
}

// S3ScanHandler processes documents from S3 URLs
func S3ScanHandler(w http.ResponseWriter, r *http.Request) {
	var req S3ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.S3URL == "" {
		ErrorResponse(w, http.StatusBadRequest, "missing s3_url parameter")
		return
	}

	// Set default region if not provided
	if req.Region == "" {
		req.Region = "us-east-1"
	}

	// Create S3 client configuration
	config := s3.Config{
		Region:          req.Region,
		AccessKeyID:     req.AccessKeyID,
		SecretAccessKey: req.SecretAccessKey,
		SessionToken:    req.SessionToken,
		RoleARN:         req.RoleARN,
		Timeout:         30 * time.Second,
	}

	client, err := s3.NewClient(config)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"s3_url": req.S3URL,
			"error":  err,
		}).Error("Failed to create S3 client")
		ErrorResponse(w, http.StatusInternalServerError, "failed to create S3 client")
		return
	}

	// Create context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Download file from S3 with detailed error handling
	data, filename, err := client.DownloadFileFromURL(ctx, req.S3URL)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"s3_url": req.S3URL,
			"error":  err,
		}).Error("Failed to download file from S3")

		// Check for specific error types
		if ctx.Err() == context.DeadlineExceeded {
			ErrorResponse(w, http.StatusRequestTimeout, "download timeout: file took too long to download from S3")
			return
		}

		// Check for AWS-specific errors
		if err.Error() == "NoSuchBucket" || strings.Contains(err.Error(), "NoSuchBucket") {
			ErrorResponse(w, http.StatusNotFound, "S3 bucket not found")
			return
		}
		if err.Error() == "NoSuchKey" || strings.Contains(err.Error(), "NoSuchKey") {
			ErrorResponse(w, http.StatusNotFound, "S3 file not found")
			return
		}
		if strings.Contains(err.Error(), "AccessDenied") {
			ErrorResponse(w, http.StatusForbidden, "access denied: check S3 permissions")
			return
		}
		if strings.Contains(err.Error(), "invalid S3 URL") {
			ErrorResponse(w, http.StatusBadRequest, "invalid S3 URL format")
			return
		}

		ErrorResponse(w, http.StatusInternalServerError, "failed to download file from S3")
		return
	}

	// Check file size limits (10MB max)
	const maxFileSize = 10 << 20 // 10 MB
	if len(data) > maxFileSize {
		logrus.WithFields(logrus.Fields{
			"s3_url":   req.S3URL,
			"filename": filename,
			"size":     len(data),
			"max_size": maxFileSize,
		}).Warn("File size exceeds maximum allowed")
		ErrorResponse(w, http.StatusRequestEntityTooLarge, "file size exceeds 10MB limit")
		return
	}

	// Extract text from the downloaded file with timeout protection
	text, err := scanner.ExtractText(data, filename)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"s3_url":   req.S3URL,
			"filename": filename,
			"error":    err,
		}).Error("Failed to extract text from S3 file")

		if strings.Contains(err.Error(), "unsupported file format") {
			ErrorResponse(w, http.StatusUnsupportedMediaType, fmt.Sprintf("unsupported file format: %s", err.Error()))
			return
		}

		ErrorResponse(w, http.StatusInternalServerError, "failed to extract text from file")
		return
	}

	// Process the text with the scanning engine
	findings := engine.Evaluate(text, filename, engine.GetRules())

	if engine.GetDebugMode() {
		logrus.WithFields(logrus.Fields{
			"s3_url":   req.S3URL,
			"filename": filename,
			"findings": findings,
		}).Debug("S3 scan findings before encoding")
	}

	// Return the results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Report{FileID: filename, Findings: findings})
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

// LLMScanHandler performs document analysis using LLM
func LLMScanHandler(w http.ResponseWriter, r *http.Request) {
	if llmAnalyzer == nil {
		ErrorResponse(w, http.StatusServiceUnavailable, "LLM service is not available")
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

	// Parse optional custom rules
	var customRules []string
	if rulesParam := r.FormValue("rules"); rulesParam != "" {
		if err := json.Unmarshal([]byte(rulesParam), &customRules); err != nil {
			logrus.WithFields(logrus.Fields{
				"rules_param": rulesParam,
				"error":       err,
			}).Warn("Failed to parse custom rules, using defaults")
		}
	}

	// Create analysis request
	analysisReq := llm.AnalysisRequest{
		Text:     text,
		Filename: header.Filename,
		Rules:    customRules,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Perform LLM analysis
	analysisResp, err := llmAnalyzer.AnalyzeDocument(ctx, analysisReq)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"filename": header.Filename,
			"error":    err,
		}).Error("LLM analysis failed")
		ErrorResponse(w, http.StatusInternalServerError, "LLM analysis failed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysisResp)
}

// HybridScanHandler performs both regex and LLM analysis
func HybridScanHandler(w http.ResponseWriter, r *http.Request) {
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

	// Perform regex analysis first
	regexFindings := engine.Evaluate(text, header.Filename, engine.GetRules())

	// Create response object
	response := map[string]interface{}{
		"file_id":        header.Filename,
		"regex_findings": regexFindings,
	}

	// Perform LLM analysis if available
	if llmAnalyzer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		// LLM analysis
		analysisReq := llm.AnalysisRequest{
			Text:     text,
			Filename: header.Filename,
		}

		llmAnalysis, err := llmAnalyzer.AnalyzeDocument(ctx, analysisReq)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"filename": header.Filename,
				"error":    err,
			}).Warn("LLM analysis failed in hybrid mode")
		} else {
			response["llm_analysis"] = llmAnalysis
		}

		// Validate regex findings with LLM
		validatedFindings, err := llmAnalyzer.ValidateFindings(ctx, regexFindings, text, header.Filename)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"filename": header.Filename,
				"error":    err,
			}).Warn("LLM validation failed in hybrid mode")
			response["validated_findings"] = regexFindings // Use original if validation fails
		} else {
			response["validated_findings"] = validatedFindings
		}

		if llmAnalysis != nil {
			response["tokens_used"] = llmAnalysis.TokensUsed
		}
	} else {
		response["validated_findings"] = regexFindings
		response["llm_analysis"] = nil
		response["tokens_used"] = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SmartScanHandler performs optimized analysis using rules as pre-filters
func SmartScanHandler(w http.ResponseWriter, r *http.Request) {
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

	// Smart pre-filtering analysis
	if llmAnalyzer != nil {
		// Create smart analyzer with cost optimization
		smartConfig := llm.SmartAnalysisConfig{
			MinFindingsThreshold: 2,
			TriggerSeverities:    []string{"high", "medium"},
			MinDocumentLength:    200,
			MaxDocumentLength:    4000,
			AnalyzeRuleTypes:     []string{"disease", "aggressive", "property"},
		}

		smartAnalyzer := llm.NewSmartAnalyzer(llmAnalyzer, smartConfig)

		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		result, err := smartAnalyzer.AnalyzeWithPrefiltering(ctx, text, header.Filename, engine.GetRules())
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"filename": header.Filename,
				"error":    err,
			}).Error("Smart analysis failed")
			ErrorResponse(w, http.StatusInternalServerError, "smart analysis failed")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	} else {
		// Fallback to regex-only
		regexFindings := engine.Evaluate(text, header.Filename, engine.GetRules())
		response := map[string]interface{}{
			"regex_findings":     regexFindings,
			"llm_used":          false,
			"validated_findings": regexFindings,
			"tokens_used":       0,
			"cost_savings":      "100% - LLM disabled",
			"analysis_reason":   "LLM service not available",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
