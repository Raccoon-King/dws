package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"dws/engine"
)

func TestE2E_FullWorkflow(t *testing.T) {
	// Create test rules file
	rulesPath := createRulesFile(t)
	os.Setenv("RULES_FILE", rulesPath)
	defer os.Unsetenv("RULES_FILE")

	// Start server in background
	srv, err := newServer()
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- srv.ListenAndServe()
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Ensure server is running
	defer func() {
		if err := srv.Shutdown(context.Background()); err != nil {
			t.Logf("server shutdown error: %v", err)
		}
	}()

	baseURL := "http://localhost:8080"

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

		var findings []engine.Finding
		if err := json.NewDecoder(resp.Body).Decode(&findings); err != nil {
			t.Fatalf("decode response: %v", err)
		}

		// Verify we got findings (should match "foo" pattern)
		if len(findings) == 0 {
			t.Error("expected at least one finding")
		}

		// Verify finding details
		found := false
		for _, f := range findings {
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

	t.Run("process_document", func(t *testing.T) {
		// Test the process-document endpoint
		testContent := "This is another test with foo pattern"
		
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		
		part, err := writer.CreateFormFile("file", "document.txt")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		
		if _, err := part.Write([]byte(testContent)); err != nil {
			t.Fatalf("write to form: %v", err)
		}
		
		if err := writer.Close(); err != nil {
			t.Fatalf("close writer: %v", err)
		}

		req, err := http.NewRequest("POST", baseURL+"/process-document", &body)
		if err != nil {
			t.Fatalf("create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("process-document request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
		}

		// Decode the report structure
		var report map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
			t.Fatalf("decode response: %v", err)
		}

		// Verify report structure
		if _, exists := report["fileID"]; !exists {
			t.Error("expected fileID in report")
		}
		if _, exists := report["findings"]; !exists {
			t.Error("expected findings in report")
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
		// Test rules reload endpoint
		resp, err := http.Post(baseURL+"/rules/reload", "application/json", nil)
		if err != nil {
			t.Fatalf("rules reload failed: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestE2E_LoadBalancing(t *testing.T) {
	// Create test rules file
	rulesPath := createRulesFile(t)
	os.Setenv("RULES_FILE", rulesPath)
	defer os.Unsetenv("RULES_FILE")

	// Start server
	srv, err := newServer()
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	go func() {
		srv.ListenAndServe()
	}()

	time.Sleep(100 * time.Millisecond)
	defer srv.Shutdown(context.Background())

	baseURL := "http://localhost:8080"

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

				var findings []engine.Finding
				if err := json.NewDecoder(resp.Body).Decode(&findings); err != nil {
					errors <- err
					return
				}

				if len(findings) == 0 {
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
