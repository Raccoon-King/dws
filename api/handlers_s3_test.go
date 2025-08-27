package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"dws/api"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
)

type mockS3Client struct {
	GetObjectWithContextFunc func(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error)
}

func (m *mockS3Client) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	return m.GetObjectWithContextFunc(ctx, input, opts...)
}

func TestS3EventHandler(t *testing.T) {
	// Create a mock S3 event payload
	event := api.S3Event{
		Records: []struct {
			S3 struct {
				Bucket struct {
					Name string `json:"name"`
				}
				Object struct {
					Key string `json:"key"`
				}
			}
		}{
			{
				S3: struct {
					Bucket struct {
						Name string `json:"name"`
					}
					Object struct {
						Key string `json:"key"`
					}
				}{
					Bucket: struct {
						Name string `json:"name"`
					}{
						Name: "test-bucket",
					},
					Object: struct {
						Key string `json:"key"`
					}{
						Key: "test-object.pdf",
					},
				},
			},
		},
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal S3 event: %v", err)
	}

	// Create a mock S3 service
	mockSvc := &mockS3Client{
		GetObjectWithContextFunc: func(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
			// Simulate a successful S3 GetObject operation
			return &s3.GetObjectOutput{
				Body: io.NopCloser(bytes.NewReader([]byte("test content with badword"))),
			},
			nil
		},
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", "/s3-event", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler with the mock S3 service
	api.S3EventHandler(mockSvc).ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
