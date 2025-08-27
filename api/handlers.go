package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"dws/engine"
	"dws/scanner"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Event represents a simplified S3 event notification structure.
type S3Event struct {
	Records []struct {
		S3 struct {
			Bucket struct {
				Name string `json:"name"`
			}
			Object struct {
				Key string `json:"key"`
			}
		}
	} `json:"Records"`
}

type S3Client interface {
	GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error)
}

// S3EventHandler processes S3 event notifications.
func S3EventHandler(s3Svc S3Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var event S3Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, "invalid S3 event payload", http.StatusBadRequest)
			log.Printf("Error decoding S3 event: %v", err)
			return
		}

		for _, record := range event.Records {
			bucketName := record.S3.Bucket.Name
			objectKey := record.S3.Object.Key
			log.Printf("Received S3 event for bucket: %s, key: %s", bucketName, objectKey)

			// Download the object from S3
			input := &s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(objectKey),
			}

			result, err := s3Svc.GetObjectWithContext(context.Background(), input)
			if err != nil {
				log.Printf("Error getting object %s from bucket %s: %v", objectKey, bucketName, err)
				continue // Continue to next record if there's an error with this one
			}
			defer result.Body.Close()

			data, err := io.ReadAll(result.Body)
			if err != nil {
				log.Printf("Error reading object body: %v", err)
				continue
			}

			// Extract text and evaluate
			text, err := scanner.ExtractText(data, objectKey)
			if err != nil {
				log.Printf("Error extracting text from %s: %v", objectKey, err)
				continue
			}

			findings := engine.Evaluate(text, objectKey, engine.GetRules())
			log.Printf("Findings for %s: %+v", objectKey, findings)

			// TODO: Store findings in MySQL (Task 2.3)
		}

		w.WriteHeader(http.StatusOK)
	}
}

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
