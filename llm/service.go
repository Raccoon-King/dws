package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// Provider represents different LLM providers
type Provider string

const (
	ProviderOpenAI   Provider = "openai"
	ProviderBedrock  Provider = "bedrock"
	ProviderOllama   Provider = "ollama"
	ProviderAzure    Provider = "azure"
)

// Config holds LLM service configuration
type Config struct {
	Provider    Provider      `yaml:"provider" json:"provider"`
	Enabled     bool          `yaml:"enabled" json:"enabled"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`
	MaxTokens   int           `yaml:"max_tokens" json:"max_tokens"`
	Temperature float32       `yaml:"temperature" json:"temperature"`

	// OpenAI-compatible configuration
	OpenAI OpenAIConfig `yaml:"openai" json:"openai"`

	// Bedrock configuration
	Bedrock BedrockConfig `yaml:"bedrock" json:"bedrock"`
}

// OpenAIConfig holds OpenAI-compatible API configuration
type OpenAIConfig struct {
	APIKey   string `yaml:"api_key" json:"api_key"`
	BaseURL  string `yaml:"base_url" json:"base_url"` // For OpenAI-compatible APIs
	Model    string `yaml:"model" json:"model"`
	OrgID    string `yaml:"org_id" json:"org_id"`
}

// BedrockConfig holds Amazon Bedrock configuration
type BedrockConfig struct {
	Region          string `yaml:"region" json:"region"`
	AccessKeyID     string `yaml:"access_key_id" json:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key" json:"secret_access_key"`
	SessionToken    string `yaml:"session_token" json:"session_token"`
	RoleARN         string `yaml:"role_arn" json:"role_arn"`
	ModelID         string `yaml:"model_id" json:"model_id"`
}

// CompletionRequest represents a request for text completion
type CompletionRequest struct {
	Prompt      string            `json:"prompt"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float32           `json:"temperature,omitempty"`
	Context     map[string]string `json:"context,omitempty"`
}

// CompletionResponse represents the response from LLM
type CompletionResponse struct {
	Text         string            `json:"text"`
	TokensUsed   int               `json:"tokens_used"`
	Model        string            `json:"model"`
	Provider     Provider          `json:"provider"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// LLMProvider interface that all providers must implement
type LLMProvider interface {
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	ValidateConfig() error
	GetProviderName() Provider
}

// Service manages LLM operations
type Service struct {
	config   Config
	provider LLMProvider
}

// NewService creates a new LLM service with the specified configuration
func NewService(config Config) (*Service, error) {
	if !config.Enabled {
		return &Service{config: config}, nil
	}

	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 1000
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}

	var provider LLMProvider
	var err error

	switch config.Provider {
	case ProviderOpenAI, ProviderOllama, ProviderAzure:
		provider, err = NewOpenAIProvider(config.OpenAI)
	case ProviderBedrock:
		provider, err = NewBedrockProvider(config.Bedrock)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", config.Provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create %s provider: %w", config.Provider, err)
	}

	if err := provider.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid %s configuration: %w", config.Provider, err)
	}

	logrus.WithFields(logrus.Fields{
		"provider": config.Provider,
		"model":    getModelName(config),
		"timeout":  config.Timeout,
	}).Info("LLM service initialized")

	return &Service{
		config:   config,
		provider: provider,
	}, nil
}

// Complete performs text completion using the configured provider
func (s *Service) Complete(ctx context.Context, prompt string) (*CompletionResponse, error) {
	if !s.config.Enabled || s.provider == nil {
		return nil, fmt.Errorf("LLM service is disabled")
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	req := CompletionRequest{
		Prompt:      prompt,
		MaxTokens:   s.config.MaxTokens,
		Temperature: s.config.Temperature,
	}

	logrus.WithFields(logrus.Fields{
		"provider":    s.provider.GetProviderName(),
		"prompt_len":  len(prompt),
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
	}).Debug("Sending completion request to LLM")

	response, err := s.provider.Complete(timeoutCtx, req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"provider": s.provider.GetProviderName(),
			"error":    err,
		}).Error("LLM completion failed")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"provider":    response.Provider,
		"tokens_used": response.TokensUsed,
		"model":       response.Model,
		"response_len": len(response.Text),
	}).Debug("LLM completion successful")

	return response, nil
}

// IsEnabled returns whether the LLM service is enabled
func (s *Service) IsEnabled() bool {
	return s.config.Enabled && s.provider != nil
}

// GetConfig returns the current configuration (with sensitive data masked)
func (s *Service) GetConfig() Config {
	config := s.config
	// Mask sensitive information
	config.OpenAI.APIKey = maskString(config.OpenAI.APIKey)
	config.Bedrock.AccessKeyID = maskString(config.Bedrock.AccessKeyID)
	config.Bedrock.SecretAccessKey = maskString(config.Bedrock.SecretAccessKey)
	config.Bedrock.SessionToken = maskString(config.Bedrock.SessionToken)
	return config
}

// getModelName returns the model name based on provider
func getModelName(config Config) string {
	switch config.Provider {
	case ProviderOpenAI, ProviderOllama, ProviderAzure:
		return config.OpenAI.Model
	case ProviderBedrock:
		return config.Bedrock.ModelID
	default:
		return "unknown"
	}
}

// maskString masks sensitive strings for logging
func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}