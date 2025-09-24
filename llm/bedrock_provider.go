package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/bedrockruntime"
	"github.com/sirupsen/logrus"
)

// BedrockProvider implements LLMProvider for Amazon Bedrock
type BedrockProvider struct {
	config         BedrockConfig
	bedrockClient  *bedrockruntime.BedrockRuntime
	modelHandler   BedrockModelHandler
}

// BedrockModelHandler interface for different model families
type BedrockModelHandler interface {
	PrepareRequest(prompt string, maxTokens int, temperature float32) ([]byte, error)
	ParseResponse(response []byte) (string, int, error)
	GetModelFamily() string
}

// Claude3Handler handles Anthropic Claude models
type Claude3Handler struct{}

// Titan/J2Handler handles Amazon Titan and AI21 Jurassic models
type TitanHandler struct{}

// LlamaHandler handles Meta Llama models
type LlamaHandler struct{}

// Claude3Request represents the request format for Claude models
type Claude3Request struct {
	Messages    []Claude3Message `json:"messages"`
	MaxTokens   int              `json:"max_tokens"`
	Temperature float32          `json:"temperature,omitempty"`
	System      string           `json:"system,omitempty"`
}

type Claude3Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Claude3Response represents the response from Claude models
type Claude3Response struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// TitanRequest represents the request format for Titan models
type TitanRequest struct {
	InputText       string              `json:"inputText"`
	TextGenerationConfig TitanGenConfig `json:"textGenerationConfig"`
}

type TitanGenConfig struct {
	MaxTokenCount int     `json:"maxTokenCount"`
	Temperature   float32 `json:"temperature"`
	TopP          float32 `json:"topP"`
}

// TitanResponse represents the response from Titan models
type TitanResponse struct {
	Results []struct {
		TokenCount       int    `json:"tokenCount"`
		OutputText       string `json:"outputText"`
		CompletionReason string `json:"completionReason"`
	} `json:"results"`
	InputTextTokenCount int `json:"inputTextTokenCount"`
}

// LlamaRequest represents the request format for Llama models
type LlamaRequest struct {
	Prompt      string  `json:"prompt"`
	MaxGenLen   int     `json:"max_gen_len"`
	Temperature float32 `json:"temperature"`
}

// LlamaResponse represents the response from Llama models
type LlamaResponse struct {
	Generation           string `json:"generation"`
	PromptTokenCount     int    `json:"prompt_token_count"`
	GenerationTokenCount int    `json:"generation_token_count"`
}

// NewBedrockProvider creates a new Amazon Bedrock provider
func NewBedrockProvider(config BedrockConfig) (*BedrockProvider, error) {
	if config.Region == "" {
		config.Region = "us-east-1" // Default region
	}

	if config.ModelID == "" {
		return nil, fmt.Errorf("model ID is required for Bedrock provider")
	}

	// Configure AWS session
	awsConfig := &aws.Config{
		Region:     aws.String(config.Region),
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
		MaxRetries: aws.Int(3),
	}

	var sess *session.Session
	var err error

	// Configure credentials
	if config.RoleARN != "" {
		// Use IAM role
		sess, err = session.NewSession(awsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS session: %w", err)
		}
		creds := stscreds.NewCredentials(sess, config.RoleARN)
		awsConfig.Credentials = creds
	} else if config.AccessKeyID != "" && config.SecretAccessKey != "" {
		// Use access keys
		creds := credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, config.SessionToken)
		awsConfig.Credentials = creds
	}
	// If neither is provided, it will use the default credential chain

	sess, err = session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	bedrockClient := bedrockruntime.New(sess)

	// Determine model handler based on model ID
	modelHandler, err := getModelHandler(config.ModelID)
	if err != nil {
		return nil, fmt.Errorf("unsupported model: %w", err)
	}

	return &BedrockProvider{
		config:         config,
		bedrockClient:  bedrockClient,
		modelHandler:   modelHandler,
	}, nil
}

// Complete implements the LLMProvider interface
func (p *BedrockProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// Prepare request body using appropriate model handler
	requestBody, err := p.modelHandler.PrepareRequest(req.Prompt, req.MaxTokens, req.Temperature)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"model_id":     p.config.ModelID,
		"model_family": p.modelHandler.GetModelFamily(),
		"region":       p.config.Region,
		"prompt_len":   len(req.Prompt),
	}).Debug("Sending request to Amazon Bedrock")

	// Call Bedrock
	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(p.config.ModelID),
		Body:        requestBody,
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	}

	result, err := p.bedrockClient.InvokeModelWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("Bedrock API call failed: %w", err)
	}

	// Parse response using appropriate model handler
	text, tokensUsed, err := p.modelHandler.ParseResponse(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	response := &CompletionResponse{
		Text:       text,
		TokensUsed: tokensUsed,
		Model:      p.config.ModelID,
		Provider:   ProviderBedrock,
		Metadata: map[string]string{
			"model_family": p.modelHandler.GetModelFamily(),
			"region":       p.config.Region,
		},
	}

	return response, nil
}

// ValidateConfig validates the Bedrock provider configuration
func (p *BedrockProvider) ValidateConfig() error {
	if p.config.ModelID == "" {
		return fmt.Errorf("model ID is required")
	}

	if p.config.Region == "" {
		return fmt.Errorf("region is required")
	}

	// Validate model ID format
	if !strings.Contains(p.config.ModelID, ".") {
		return fmt.Errorf("invalid model ID format: %s", p.config.ModelID)
	}

	return nil
}

// GetProviderName returns the provider name
func (p *BedrockProvider) GetProviderName() Provider {
	return ProviderBedrock
}

// getModelHandler returns the appropriate model handler based on model ID
func getModelHandler(modelID string) (BedrockModelHandler, error) {
	modelID = strings.ToLower(modelID)

	if strings.Contains(modelID, "claude") {
		return &Claude3Handler{}, nil
	} else if strings.Contains(modelID, "titan") || strings.Contains(modelID, "j2") {
		return &TitanHandler{}, nil
	} else if strings.Contains(modelID, "llama") {
		return &LlamaHandler{}, nil
	}

	return nil, fmt.Errorf("unsupported model family for model ID: %s", modelID)
}

// Claude3Handler implementation
func (h *Claude3Handler) PrepareRequest(prompt string, maxTokens int, temperature float32) ([]byte, error) {
	req := Claude3Request{
		Messages: []Claude3Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	return json.Marshal(req)
}

func (h *Claude3Handler) ParseResponse(response []byte) (string, int, error) {
	var resp Claude3Response
	if err := json.Unmarshal(response, &resp); err != nil {
		return "", 0, err
	}

	if len(resp.Content) == 0 {
		return "", 0, fmt.Errorf("no content in response")
	}

	text := resp.Content[0].Text
	tokens := resp.Usage.InputTokens + resp.Usage.OutputTokens

	return text, tokens, nil
}

func (h *Claude3Handler) GetModelFamily() string {
	return "claude"
}

// TitanHandler implementation
func (h *TitanHandler) PrepareRequest(prompt string, maxTokens int, temperature float32) ([]byte, error) {
	req := TitanRequest{
		InputText: prompt,
		TextGenerationConfig: TitanGenConfig{
			MaxTokenCount: maxTokens,
			Temperature:   temperature,
			TopP:          0.9,
		},
	}

	return json.Marshal(req)
}

func (h *TitanHandler) ParseResponse(response []byte) (string, int, error) {
	var resp TitanResponse
	if err := json.Unmarshal(response, &resp); err != nil {
		return "", 0, err
	}

	if len(resp.Results) == 0 {
		return "", 0, fmt.Errorf("no results in response")
	}

	text := resp.Results[0].OutputText
	tokens := resp.InputTextTokenCount + resp.Results[0].TokenCount

	return text, tokens, nil
}

func (h *TitanHandler) GetModelFamily() string {
	return "titan"
}

// LlamaHandler implementation
func (h *LlamaHandler) PrepareRequest(prompt string, maxTokens int, temperature float32) ([]byte, error) {
	req := LlamaRequest{
		Prompt:      prompt,
		MaxGenLen:   maxTokens,
		Temperature: temperature,
	}

	return json.Marshal(req)
}

func (h *LlamaHandler) ParseResponse(response []byte) (string, int, error) {
	var resp LlamaResponse
	if err := json.Unmarshal(response, &resp); err != nil {
		return "", 0, err
	}

	text := resp.Generation
	tokens := resp.PromptTokenCount + resp.GenerationTokenCount

	return text, tokens, nil
}

func (h *LlamaHandler) GetModelFamily() string {
	return "llama"
}