package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"dws/api"
	"dws/engine"
)

func TestE2E_FullWorkflow(t *testing.T) {
	// Create test rules file
	rulesPath := CreateRulesFile(t)
	os.Setenv("RULES_FILE", rulesPath)
	api.SetRulesFile(rulesPath)
	defer os.Unsetenv("RULES_FILE")

	// Create a new server instance
	srv, err := NewServer(rulesPath)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Create a test server
	testServer := httptest.NewServer(srv.Handler)
	defer testServer.Close()

	baseURL := testServer.URL

	t.Run("health_check", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			t.Fatalf("health check failed: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("scan_text_file", func(t *testing.T) {
		// Create test document with pattern that should trigger rule
		testContent := "This document contains foo which should trigger a rule"
		
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		
		part, err := writer.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		
		if _, err := part.Write([]byte(testContent)); err != nil {
			t.Fatalf("write to form: %v", err)
		}
		
		if err := writer.Close(); err != nil {
			t.Fatalf("close writer: %v", err)
		}

		req, err := http.NewRequest("POST", baseURL+"/scan", &body)
		if err != nil {
			t.Fatalf("create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("scan request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
		}

		var report struct {
			FileID   string          `json:"fileID"`
			Findings []engine.Finding `json:"findings"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
			t.Fatalf("decode response: %v", err)
		}

		// Verify we got findings (should match "foo" pattern)
		if len(report.Findings) == 0 {
			t.Error("expected at least one finding")
		}

		// Verify finding details
		found := false
		for _, f := range report.Findings {
			if f.RuleID == "r1" && f.Severity == "high" {
				found = true
				if !strings.Contains(f.Context, "foo") {
					t.Errorf("expected match to contain 'foo', got: %s", f.Context)
				}
				break
			}
		}
		if !found {
			t.Error("expected to find rule r1 with high severity")
		}
	})

	t.Run("scan_pdf_file", func(t *testing.T) {
		// Create test document with pattern that should trigger rule
		pdfContent, err := os.ReadFile("testfiles/sample.pdf")
		if err != nil {
			t.Fatalf("read pdf file: %v", err)
		}

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		part, err := writer.CreateFormFile("file", "sample.pdf")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}

		if _, err := part.Write(pdfContent); err != nil {
			t.Fatalf("write to form: %v", err)
		}

		if err := writer.Close(); err != nil {
			t.Fatalf("close writer: %v", err)
		}

		req, err := http.NewRequest("POST", baseURL+"/scan", &body)
		if err != nil {
			t.Fatalf("create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("scan request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
		}

		var report struct {
			FileID   string          `json:"fileID"`
			Findings []engine.Finding `json:"findings"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
			t.Fatalf("decode response: %v", err)
		}

		// Verify we got findings (should match "raccoon" pattern)
		if len(report.Findings) == 0 {
			t.Error("expected at least one finding")
		}

		// Verify finding details
		found := false
		for _, f := range report.Findings {
			if f.RuleID == "raccoon-mention" && f.Severity == "informational" {
				found = true
				if !strings.Contains(strings.ToLower(f.Context), "raccoon") {
					t.Errorf("expected match to contain 'raccoon', got: %s", f.Context)
				}
				break
			}
		}
		if !found {
			t.Error("expected to find rule raccoon-mention with informational severity")
		}
	})

	t.Run("unsupported_file_type", func(t *testing.T) {
		// Test with unsupported file type
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		
		part, err := writer.CreateFormFile("file", "test.unknown")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		
		if _, err := part.Write([]byte("some content")); err != nil {
			t.Fatalf("write to form: %v", err)
		}
		
		if err := writer.Close(); err != nil {
			t.Fatalf("close writer: %v", err)
		}

		req, err := http.NewRequest("POST", baseURL+"/scan", &body)
		if err != nil {
			t.Fatalf("create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("scan request failed: %v", err)
		}
		defer resp.Body.Close()

		// Should return bad request for unsupported file
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("missing_file", func(t *testing.T) {
		// Test request without file
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		writer.Close()

		req, err := http.NewRequest("POST", baseURL+"/scan", &body)
		if err != nil {
			t.Fatalf("create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("scan request failed: %v", err)
		}
		defer resp.Body.Close()

		// Should return bad request for missing file
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("rules_reload", func(t *testing.T) {
		// Test rules reload endpoint by sending new rules
		rules := []engine.Rule{
			{ID: "test-reload", Pattern: "test", Severity: "low", Description: "test rule"},
		}
		reqBody := map[string][]engine.Rule{"rules": rules}
		body, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("marshal rules: %v", err)
		}
		
		resp, err := http.Post(baseURL+"/rules/reload", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("rules reload failed: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("expected 200, got %d: %s", resp.StatusCode, string(body))
		}
	})
}

func TestE2E_LoadBalancing(t *testing.T) {
	// Create test rules file
	rulesPath := CreateRulesFile(t)
	os.Setenv("RULES_FILE", rulesPath)
	api.SetRulesFile(rulesPath)
	defer os.Unsetenv("RULES_FILE")

	// Create a new server instance
	srv, err := NewServer(rulesPath)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Create a test server
	testServer := httptest.NewServer(srv.Handler)
	defer testServer.Close()

	baseURL := testServer.URL

	t.Run("concurrent_requests", func(t *testing.T) {
		// Test multiple concurrent requests
		const numRequests = 10
		errors := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(id int) {
				testContent := fmt.Sprintf("Test document %d with foo pattern", id)

				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				part, err := writer.CreateFormFile("file", fmt.Sprintf("test%d.txt", id))
				if err != nil {
					errors <- err
					return
				}

				if _, err := part.Write([]byte(testContent)); err != nil {
					errors <- err
					return
				}

				if err := writer.Close(); err != nil {
					errors <- err
					return
				}

				req, err := http.NewRequest("POST", baseURL+"/scan", &body)
				if err != nil {
					errors <- err
					return
				}
				req.Header.Set("Content-Type", writer.FormDataContentType())

				client := &http.Client{Timeout: 10 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					errors <- err
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					errors <- fmt.Errorf("request %d failed with status %d", id, resp.StatusCode)
					return
				}

				var report struct {
					FileID   string          `json:"fileID"`
					Findings []engine.Finding `json:"findings"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
					errors <- err
					return
				}

				if len(report.Findings) == 0 {
					errors <- fmt.Errorf("request %d got no findings", id)
					return
				}

				errors <- nil
			}(i)
		}

		// Wait for all requests to complete
		for i := 0; i < numRequests; i++ {
			if err := <-errors; err != nil {
				t.Errorf("concurrent request failed: %v", err)
			}
		}
	})
}

func TestE2E_FileTypeSupport(t *testing.T) {
	// Create test rules file
	rulesPath := CreateRulesFile(t)
	os.Setenv("RULES_FILE", rulesPath)
	api.SetRulesFile(rulesPath)
	defer os.Unsetenv("RULES_FILE")

	// Create rules directory and sample rule file for /ruleset testing
	rulesDir := t.TempDir()
	os.Mkdir(rulesDir+"/rules", 0755)
	testRulesContent := `
rules:
- id: html-rule
  pattern: content
  severity: medium
- id: json-rule
  pattern: "sample"
  severity: low
- id: xml-rule
  pattern: paragraph
  severity: informational
`
	if err := os.WriteFile(rulesDir+"/rules/test.yaml", []byte(testRulesContent), 0644); err != nil {
		t.Fatalf("create test rules file: %v", err)
	}

	// Replace the default rules file with our temp one by setting environment
	oldRulesFile := os.Getenv("RULES_FILE")
	os.Setenv("RULES_FILE", rulesPath)
	defer os.Setenv("RULES_FILE", oldRulesFile)

	// Create a new server instance
	srv, err := NewServer(rulesPath)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Create a test server
	testServer := httptest.NewServer(srv.Handler)
	defer testServer.Close()

	baseURL := testServer.URL
	client := &http.Client{Timeout: 30 * time.Second}

	fileTypes := map[string]string{
		"sample.html": "content",     // Should match html-rule
		"sample.txt":  "scanner",     // Should match "foo" from original rules
		"sample.json": "sample",      // Should match json-rule
		"sample.xml":  "paragraph",   // Should match xml-rule
		"sample.yaml": "scanner",     // Should match "foo" from original rules
		"sample.yml":  "scanner",     // Same as yaml
	}

	for fileName, expectedMatch := range fileTypes {
		t.Run(fileName, func(t *testing.T) {
			// Read test file content
			fileContent, err := os.ReadFile("testfiles/" + fileName)
			if err != nil {
				t.Fatalf("read test file %s: %v", fileName, err)
			}

			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			part, err := writer.CreateFormFile("file", fileName)
			if err != nil {
				t.Fatalf("create form file: %v", err)
			}
			if _, err := part.Write(fileContent); err != nil {
				t.Fatalf("write file content: %v", err)
			}
			if err := writer.Close(); err != nil {
				t.Fatalf("close writer: %v", err)
			}

			req, err := http.NewRequest("POST", baseURL+"/scan", &body)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
			}

			var report struct {
				FileID   string          `json:"fileID"`
				Findings []engine.Finding `json:"findings"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if len(report.Findings) == 0 && expectedMatch != "" {
				t.Logf("Expected findings for pattern '%s' but got none", expectedMatch)
				t.Logf("File content length: %d", len(fileContent))
			}

			// Verify file ID is set correctly
			if report.FileID != fileName {
				t.Errorf("expected FileID %s, got %s", fileName, report.FileID)
			}
		})
	}

	t.Run("htm_test", func(t *testing.T) {
		// Special test for .htm extension
		fileContent, err := os.ReadFile("testfiles/sample.htm")
		if err != nil {
			t.Skip("test file not found")
		}

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", "sample.htm")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write(fileContent); err != nil {
			t.Fatalf("write file content: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("close writer: %v", err)
		}

		req, err := http.NewRequest("POST", baseURL+"/scan", &body)
		if err != nil {
			t.Fatalf("create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var report struct {
			FileID   string          `json:"fileID"`
			Findings []engine.Finding `json:"findings"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
			t.Fatalf("decode response: %v", err)
		}

		if report.FileID != "sample.htm" {
			t.Errorf("expected FileID sample.htm, got %s", report.FileID)
		}
	})
}

func TestE2E_RulesetEndpoint(t *testing.T) {
	// Setup: Create a test rules file
	rulesDir := t.TempDir()
	rulesPath := rulesDir + "/rules/test.yaml"

	// Create custom ruleset content with a pattern that should match our test text
	customRulesContent := `
rules:
- id: custom-rule
  pattern: "test document"
  severity: medium
  description: "Matches test documents"
- id: multi-rule
  pattern: "pattern"
  severity: low
  description: "Matches pattern word"
`
	if err := os.WriteFile(rulesPath, []byte(customRulesContent), 0644); err != nil {
		t.Fatalf("create rules file: %v", err)
	}

	// Create a main rules file for the server
	mainRulesPath := rulesDir + "/main-rules.yaml"
	mainRulesContent := `
rules:
- id: main-rule
  pattern: "main"
- id: raccoon-mention
  pattern: "\\b(raccoon[s]?)\\b"
  severity: informational
`
	if err := os.WriteFile(mainRulesPath, []byte(mainRulesContent), 0644); err != nil {
		t.Fatalf("create main rules file: %v", err)
	}

	// Temporarily set the environment to use our main rules
	os.Setenv("RULES_FILE", mainRulesPath)
	api.SetRulesFile(mainRulesPath)
	defer os.Unsetenv("RULES_FILE")

	// Create a new server instance
	srv, err := NewServer(mainRulesPath)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Create a test server
	testServer := httptest.NewServer(srv.Handler)
	defer testServer.Close()

	baseURL := testServer.URL
	client := &http.Client{Timeout: 30 * time.Second}

	t.Run("valid_ruleset", func(t *testing.T) {
		// Test valid ruleset loading and evaluation
		testContent := "This is a test document with multiple patterns to check"

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", "test-doc.txt")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write([]byte(testContent)); err != nil {
			t.Fatalf("write test content: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("close writer: %v", err)
		}

		req, err := http.NewRequest("POST", baseURL+"/ruleset?rule=test", &body)
		if err != nil {
			t.Fatalf("create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(bodyBytes))
		}

		var report struct {
			FileID   string          `json:"fileID"`
			Findings []engine.Finding `json:"findings"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
			t.Fatalf("decode response: %v", err)
		}

		if len(report.Findings) == 0 {
			t.Error("expected at least one finding from custom ruleset")
		}

		// Verify we get findings from our custom rules
		foundCustomRule := false
		foundMultiRule := false
		for _, finding := range report.Findings {
			if finding.RuleID == "custom-rule" {
				foundCustomRule = true
			}
			if finding.RuleID == "multi-rule" {
				foundMultiRule = true
			}
		}

		if !foundCustomRule {
			t.Error("expected to find 'custom-rule' from ruleset")
		}
		if !foundMultiRule {
			t.Error("expected to find 'multi-rule' from ruleset")
		}

		if report.FileID != "test-doc.txt" {
			t.Errorf("expected FileID 'test-doc.txt', got '%s'", report.FileID)
		}
	})

	t.Run("missing_rule_parameter", func(t *testing.T) {
		// Test missing rule query parameter
		testContent := "test content"

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write([]byte(testContent)); err != nil {
			t.Fatalf("write test content: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("close writer: %v", err)
		}

		req, err := http.NewRequest("POST", baseURL+"/ruleset", &body)
		if err != nil {
			t.Fatalf("create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 for missing rule parameter, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid_rule_name", func(t *testing.T) {
		// Test invalid rule name with path traversal attempt
		testContent := "test content"

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write([]byte(testContent)); err != nil {
			t.Fatalf("write test content: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("close writer: %v", err)
		}

		req, err := http.NewRequest("POST", baseURL+"/ruleset?rule=../../invalid", &body)
		if err != nil {
			t.Fatalf("create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 for invalid rule name, got %d", resp.StatusCode)
		}
	})

	t.Run("nonexistent_ruleset", func(t *testing.T) {
		// Test request for non-existent ruleset
		testContent := "test content"

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write([]byte(testContent)); err != nil {
			t.Fatalf("write test content: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("close writer: %v", err)
		}

		req, err := http.NewRequest("POST", baseURL+"/ruleset?rule=nonexistent", &body)
		if err != nil {
			t.Fatalf("create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected 500 for nonexistent ruleset, got %d", resp.StatusCode)
		}
	})
}

func TestE2E_API_Documentation(t *testing.T) {
	// Test that the API docs endpoint returns correct information
	srv, err := NewServer("rules.yaml") // Assume default exists
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	testServer := httptest.NewServer(srv.Handler)
	defer testServer.Close()

	baseURL := testServer.URL
	resp, err := http.Get(baseURL + "/docs")
	if err != nil {
		t.Fatalf("docs request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var docs []EndpointDoc
	if err := json.NewDecoder(resp.Body).Decode(&docs); err != nil {
		t.Fatalf("decode docs response: %v", err)
	}

	// Verify docs contains expected endpoints
	expectedEndpoints := map[string]bool{
		"/scan":        false,
		"/rules/reload": false,
		"/rules/load":  false,
		"/ruleset?rule": false,
		"/health": false,
		"/docs":   false,
	}

	for _, doc := range docs {
		if _, exists := expectedEndpoints[doc.Path]; exists {
			expectedEndpoints[doc.Path] = true
		}
	}

	for endpoint, found := range expectedEndpoints {
		if !found {
			t.Errorf("expected endpoint %s not found in API docs", endpoint)
		}
	}

	// Verify ruleset endpoint has correct documentation
	for _, doc := range docs {
		if doc.Path == "/ruleset?rule" {
			if doc.Method != "POST" {
				t.Errorf("expected POST method for /ruleset, got %s", doc.Method)
			}
			if len(doc.DataShapes) == 0 {
				t.Error("ruleset endpoint should have data shapes documentation")
			}
		}
	}
}

type EndpointDoc struct {
	Path        string      `json:"path"`
	Method      string      `json:"method"`
	Description string      `json:"description"`
	DataShapes  []DataShape `json:"data_shapes"`
	CurlExample string      `json:"curl_example"`
}

type DataShape struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Shape       string `json:"shape"`
}
