package llm

import (
	"context"
	"testing"
	"time"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "disabled service",
			config: Config{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "openai provider",
			config: Config{
				Enabled:  true,
				Provider: ProviderOpenAI,
				OpenAI: OpenAIConfig{
					APIKey: "test-key",
					Model:  "gpt-3.5-turbo",
				},
			},
			wantErr: false,
		},
		{
			name: "unsupported provider",
			config: Config{
				Enabled:  true,
				Provider: "unsupported",
			},
			wantErr: true,
		},
		{
			name: "invalid openai config",
			config: Config{
				Enabled:  true,
				Provider: ProviderOpenAI,
				OpenAI: OpenAIConfig{
					APIKey: "", // Missing API key
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewService(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && service == nil {
				t.Errorf("NewService() returned nil service without error")
			}

			if service != nil {
				if service.IsEnabled() != tt.config.Enabled {
					t.Errorf("IsEnabled() = %v, want %v", service.IsEnabled(), tt.config.Enabled)
				}
			}
		})
	}
}

func TestServiceDefaults(t *testing.T) {
	config := Config{
		Enabled:  true,
		Provider: ProviderOpenAI,
		OpenAI: OpenAIConfig{
			APIKey: "test-key",
			Model:  "gpt-3.5-turbo",
		},
	}

	service, err := NewService(config)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	// Check defaults were set
	if service.config.Timeout != 30*time.Second {
		t.Errorf("Default timeout = %v, want %v", service.config.Timeout, 30*time.Second)
	}

	if service.config.MaxTokens != 1000 {
		t.Errorf("Default MaxTokens = %v, want %v", service.config.MaxTokens, 1000)
	}

	if service.config.Temperature != 0.7 {
		t.Errorf("Default Temperature = %v, want %v", service.config.Temperature, 0.7)
	}
}

func TestGetConfig(t *testing.T) {
	config := Config{
		Enabled:  true,
		Provider: ProviderOpenAI,
		OpenAI: OpenAIConfig{
			APIKey: "secret-key-12345",
			Model:  "gpt-3.5-turbo",
		},
	}

	service, err := NewService(config)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	maskedConfig := service.GetConfig()

	// Check that sensitive data is masked
	if maskedConfig.OpenAI.APIKey == "secret-key-12345" {
		t.Errorf("API key was not masked: %s", maskedConfig.OpenAI.APIKey)
	}

	if maskedConfig.OpenAI.APIKey != "se****45" {
		t.Errorf("API key masking incorrect: got %s, want se****45", maskedConfig.OpenAI.APIKey)
	}

	// Non-sensitive data should remain
	if maskedConfig.OpenAI.Model != "gpt-3.5-turbo" {
		t.Errorf("Model should not be masked: %s", maskedConfig.OpenAI.Model)
	}
}

func TestMaskString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "****"},
		{"a", "****"},
		{"ab", "****"},
		{"abc", "****"},
		{"abcd", "****"},
		{"abcde", "ab****de"},
		{"secret-key-12345", "se****45"},
		{"sk-1234567890abcdef", "sk****ef"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := maskString(tt.input)
			if result != tt.expected {
				t.Errorf("maskString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetModelName(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected string
	}{
		{
			name: "openai model",
			config: Config{
				Provider: ProviderOpenAI,
				OpenAI:   OpenAIConfig{Model: "gpt-4"},
			},
			expected: "gpt-4",
		},
		{
			name: "bedrock model",
			config: Config{
				Provider: ProviderBedrock,
				Bedrock:  BedrockConfig{ModelID: "anthropic.claude-3-sonnet"},
			},
			expected: "anthropic.claude-3-sonnet",
		},
		{
			name: "unknown provider",
			config: Config{
				Provider: "unknown",
			},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getModelName(tt.config)
			if result != tt.expected {
				t.Errorf("getModelName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Mock provider for testing
type MockProvider struct {
	shouldError bool
	response    *CompletionResponse
}

func (m *MockProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if m.shouldError {
		return nil, context.DeadlineExceeded
	}
	if m.response != nil {
		return m.response, nil
	}
	return &CompletionResponse{
		Text:       "mock response",
		TokensUsed: 10,
		Model:      "mock-model",
		Provider:   ProviderOpenAI,
	}, nil
}

func (m *MockProvider) ValidateConfig() error {
	return nil
}

func (m *MockProvider) GetProviderName() Provider {
	return ProviderOpenAI
}

func TestServiceComplete(t *testing.T) {
	config := Config{
		Enabled:     true,
		Provider:    ProviderOpenAI,
		Timeout:     5 * time.Second,
		MaxTokens:   100,
		Temperature: 0.5,
	}

	service := &Service{
		config:   config,
		provider: &MockProvider{},
	}

	ctx := context.Background()
	response, err := service.Complete(ctx, "test prompt")

	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	if response.Text != "mock response" {
		t.Errorf("Complete() text = %q, want %q", response.Text, "mock response")
	}

	if response.TokensUsed != 10 {
		t.Errorf("Complete() tokens = %d, want %d", response.TokensUsed, 10)
	}
}

func TestServiceCompleteDisabled(t *testing.T) {
	service := &Service{
		config: Config{Enabled: false},
	}

	ctx := context.Background()
	_, err := service.Complete(ctx, "test prompt")

	if err == nil {
		t.Errorf("Complete() should return error when service is disabled")
	}

	expectedErr := "LLM service is disabled"
	if err.Error() != expectedErr {
		t.Errorf("Complete() error = %q, want %q", err.Error(), expectedErr)
	}
}

func TestServiceCompleteTimeout(t *testing.T) {
	config := Config{
		Enabled:  true,
		Provider: ProviderOpenAI,
		Timeout:  1 * time.Millisecond, // Very short timeout
	}

	service := &Service{
		config:   config,
		provider: &MockProvider{shouldError: true},
	}

	ctx := context.Background()
	_, err := service.Complete(ctx, "test prompt")

	if err == nil {
		t.Errorf("Complete() should return error on timeout")
	}
}