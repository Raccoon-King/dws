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
		pdfContent, err := os.ReadFile("Raccoon_World_Domination_Report.pdf")
		if err != nil {
			t.Fatalf("read pdf file: %v", err)
		}

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		part, err := writer.CreateFormFile("file", "Raccoon_World_Domination_Report.pdf")
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