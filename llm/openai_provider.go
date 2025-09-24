package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// OpenAIProvider implements LLMProvider for OpenAI-compatible APIs
type OpenAIProvider struct {
	config     OpenAIConfig
	httpClient *http.Client
	baseURL    string
}

// OpenAIRequest represents the request format for OpenAI-compatible APIs
type OpenAIRequest struct {
	Model       string            `json:"model"`
	Messages    []OpenAIMessage   `json:"messages"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float32           `json:"temperature,omitempty"`
	Stream      bool              `json:"stream"`
}

// OpenAIMessage represents a message in the OpenAI format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response from OpenAI-compatible APIs
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// OpenAIErrorResponse represents error responses from OpenAI-compatible APIs
type OpenAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// NewOpenAIProvider creates a new OpenAI-compatible provider
func NewOpenAIProvider(config OpenAIConfig) (*OpenAIProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required for OpenAI provider")
	}

	if config.Model == "" {
		config.Model = "gpt-3.5-turbo" // Default model
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1" // Default OpenAI URL
	}

	// Ensure baseURL doesn't end with slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	return &OpenAIProvider{
		config:     config,
		httpClient: httpClient,
		baseURL:    baseURL,
	}, nil
}

// Complete implements the LLMProvider interface
func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// Convert to OpenAI format
	openaiReq := OpenAIRequest{
		Model: p.config.Model,
		Messages: []OpenAIMessage{
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      false,
	}

	// Marshal request
	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	if p.config.OrgID != "" {
		httpReq.Header.Set("OpenAI-Organization", p.config.OrgID)
	}

	logrus.WithFields(logrus.Fields{
		"url":    httpReq.URL.String(),
		"model":  p.config.Model,
		"tokens": req.MaxTokens,
	}).Debug("Sending request to OpenAI-compatible API")

	// Send request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		var errorResp OpenAIErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
		}
		return nil, fmt.Errorf("API error (%s): %s", errorResp.Error.Type, errorResp.Error.Message)
	}

	// Parse successful response
	var openaiResp OpenAIResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Convert to our response format
	response := &CompletionResponse{
		Text:       openaiResp.Choices[0].Message.Content,
		TokensUsed: openaiResp.Usage.TotalTokens,
		Model:      openaiResp.Model,
		Provider:   ProviderOpenAI,
		Metadata: map[string]string{
			"finish_reason":      openaiResp.Choices[0].FinishReason,
			"prompt_tokens":      fmt.Sprintf("%d", openaiResp.Usage.PromptTokens),
			"completion_tokens":  fmt.Sprintf("%d", openaiResp.Usage.CompletionTokens),
			"id":                 openaiResp.ID,
		},
	}

	return response, nil
}

// ValidateConfig validates the OpenAI provider configuration
func (p *OpenAIProvider) ValidateConfig() error {
	if p.config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if p.config.Model == "" {
		return fmt.Errorf("model is required")
	}

	// Validate base URL format if provided
	if p.config.BaseURL != "" {
		if !strings.HasPrefix(p.config.BaseURL, "http://") && !strings.HasPrefix(p.config.BaseURL, "https://") {
			return fmt.Errorf("base URL must start with http:// or https://")
		}
	}

	return nil
}

// GetProviderName returns the provider name
func (p *OpenAIProvider) GetProviderName() Provider {
	// Determine actual provider based on base URL
	if strings.Contains(p.baseURL, "openai.com") {
		return ProviderOpenAI
	} else if strings.Contains(p.baseURL, "azure.com") {
		return ProviderAzure
	} else if strings.Contains(p.baseURL, "localhost") || strings.Contains(p.baseURL, "127.0.0.1") {
		return ProviderOllama
	}
	return ProviderOpenAI // Default
}