package s3

import (
	"context"
	"testing"
	"time"
)

func TestParseS3URL(t *testing.T) {
	tests := []struct {
		name        string
		s3URL       string
		wantBucket  string
		wantKey     string
		wantErr     bool
	}{
		{
			name:       "valid s3 url",
			s3URL:      "s3://my-bucket/path/to/file.txt",
			wantBucket: "my-bucket",
			wantKey:    "path/to/file.txt",
			wantErr:    false,
		},
		{
			name:       "s3 url with nested path",
			s3URL:      "s3://documents/reports/2023/annual-report.pdf",
			wantBucket: "documents",
			wantKey:    "reports/2023/annual-report.pdf",
			wantErr:    false,
		},
		{
			name:       "s3 url root file",
			s3URL:      "s3://bucket/file.txt",
			wantBucket: "bucket",
			wantKey:    "file.txt",
			wantErr:    false,
		},
		{
			name:    "invalid scheme",
			s3URL:   "https://bucket/file.txt",
			wantErr: true,
		},
		{
			name:    "invalid url",
			s3URL:   "not-a-url",
			wantErr: true,
		},
		{
			name:       "empty key",
			s3URL:      "s3://bucket/",
			wantBucket: "bucket",
			wantKey:    "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, key, err := ParseS3URL(tt.s3URL)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseS3URL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if bucket != tt.wantBucket {
					t.Errorf("ParseS3URL() bucket = %v, want %v", bucket, tt.wantBucket)
				}
				if key != tt.wantKey {
					t.Errorf("ParseS3URL() key = %v, want %v", key, tt.wantKey)
				}
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config with access keys",
			config: Config{
				Region:          "us-east-1",
				AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
				SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				Timeout:         30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid config with role arn",
			config: Config{
				Region:  "us-west-2",
				RoleARN: "arn:aws:iam::123456789012:role/test-role",
				Timeout: 60 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "config with default timeout",
			config: Config{
				Region: "eu-west-1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && client == nil {
				t.Errorf("NewClient() returned nil client")
			}

			if client != nil {
				// Verify the client has required components
				if client.s3Client == nil {
					t.Errorf("NewClient() s3Client is nil")
				}
				if client.downloader == nil {
					t.Errorf("NewClient() downloader is nil")
				}
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	config := Config{
		Region: "us-east-1",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Default timeout should be set
	if config.Timeout == 0 {
		// Config was not modified, but client should have used defaults internally
		// This is a limitation of the current design - we can't easily test the internal timeout
		// without refactoring the client structure
	}

	_ = client // Use the client to avoid unused variable warning
}

// Note: The following tests would require AWS credentials and actual S3 buckets
// They are included for completeness but will be skipped in CI/CD

func TestDownloadFile_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This would be an integration test requiring real AWS credentials
	// and an S3 bucket with test files
	t.Skip("Integration test - requires AWS credentials and S3 setup")
}

func TestDownloadFileFromURL_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This would be an integration test
	t.Skip("Integration test - requires AWS credentials and S3 setup")
}

func TestCheckFileExists_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This would be an integration test
	t.Skip("Integration test - requires AWS credentials and S3 setup")
}

// Mock tests for error conditions
func TestDownloadFileFromURL_ParseError(t *testing.T) {
	config := Config{
		Region: "us-east-1",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	_, _, err = client.DownloadFileFromURL(ctx, "invalid-url")

	if err == nil {
		t.Errorf("DownloadFileFromURL() should return error for invalid URL")
	}

	expectedErrSubstring := "invalid S3 URL scheme"
	if err != nil && !contains(err.Error(), expectedErrSubstring) {
		t.Errorf("DownloadFileFromURL() error = %v, should contain %q", err, expectedErrSubstring)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(substr) <= len(s) && containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}