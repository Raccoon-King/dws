# LLM Integration Guide

This document explains how to integrate and use the LLM (Large Language Model) features in the Document Scanning Rules Engine.

## Overview

The LLM integration adds AI-powered document analysis capabilities to complement the existing regex-based scanning. It supports both **OpenAI-compatible APIs** and **Amazon Bedrock**.

## Supported Providers

### OpenAI-Compatible APIs
- **OpenAI** (ChatGPT, GPT-4, etc.)
- **Azure OpenAI Service**
- **Ollama** (local models)
- **Other OpenAI-compatible endpoints**

### Amazon Bedrock
- **Anthropic Claude** (Claude-3-Sonnet, Claude-3-Haiku, etc.)
- **Amazon Titan** (Titan Text Express, etc.)
- **Meta Llama** (Llama-2-70B-Chat, etc.)

## Configuration

### 1. Enable LLM Service

Set the environment variable:
```bash
export LLM_ENABLED=true
```

### 2. Configure Provider

#### OpenAI Configuration
```yaml
# config/llm.yaml
llm:
  enabled: true
  provider: "openai"
  timeout: "30s"
  max_tokens: 1000
  temperature: 0.7

openai:
  api_key: "${LLM_API_KEY}"
  model: "gpt-3.5-turbo"
  base_url: ""  # Optional, defaults to OpenAI
  org_id: ""    # Optional
```

Environment variables:
```bash
export LLM_API_KEY="sk-your-openai-api-key"
```

#### Azure OpenAI Configuration
```yaml
llm:
  provider: "azure"

openai:
  api_key: "${AZURE_OPENAI_KEY}"
  base_url: "https://your-resource.openai.azure.com"
  model: "gpt-35-turbo"  # Azure deployment name
```

#### Ollama Configuration (Local)
```yaml
llm:
  provider: "ollama"

openai:
  api_key: "not-needed"  # Ollama doesn't require API key
  base_url: "http://localhost:11434/v1"
  model: "llama2"
```

#### Amazon Bedrock Configuration
```yaml
llm:
  provider: "bedrock"

bedrock:
  region: "us-east-1"
  model_id: "anthropic.claude-3-sonnet-20240229-v1:0"
  # AWS credentials (optional, uses default credential chain)
  access_key_id: "${AWS_ACCESS_KEY_ID}"
  secret_access_key: "${AWS_SECRET_ACCESS_KEY}"
  session_token: "${AWS_SESSION_TOKEN}"
  role_arn: "${AWS_ROLE_ARN}"  # For assume role
```

Environment variables:
```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"
```

## API Endpoints

### 1. LLM-Only Analysis: `POST /scan/llm`

Performs semantic analysis using only the LLM.

**Request:**
```bash
curl -X POST \
  -F 'file=@document.pdf' \
  -F 'rules=["Look for API keys", "Check for personal information"]' \
  http://localhost:8080/scan/llm
```

**Response:**
```json
{
  "findings": [
    {
      "rule_id": "llm-finding-1",
      "severity": "high",
      "line": 15,
      "context": "API_KEY = sk-abc123def456",
      "description": "API key detected",
      "confidence": 0.95,
      "reasoning": "This appears to be an OpenAI API key based on the 'sk-' prefix"
    }
  ],
  "summary": "Document contains 1 high-severity finding related to API key exposure",
  "confidence": 0.9,
  "tokens_used": 245,
  "model": "gpt-3.5-turbo",
  "provider": "openai"
}
```

### 2. Hybrid Analysis: `POST /scan/hybrid`

Combines regex-based rules with LLM analysis and validation.

**Request:**
```bash
curl -X POST \
  -F 'file=@document.pdf' \
  http://localhost:8080/scan/hybrid
```

**Response:**
```json
{
  "file_id": "document.pdf",
  "regex_findings": [
    {
      "file_id": "document.pdf",
      "rule_id": "api-key-pattern",
      "severity": "high",
      "line": 15,
      "context": "API_KEY = sk-abc123def456"
    }
  ],
  "llm_analysis": {
    "findings": [...],
    "summary": "...",
    "confidence": 0.9
  },
  "validated_findings": [
    {
      "file_id": "document.pdf",
      "rule_id": "api-key-pattern",
      "severity": "high",
      "line": 15,
      "context": "API_KEY = sk-abc123def456"
    }
  ],
  "tokens_used": 320
}
```

## Model Selection Guide

### OpenAI Models
- **gpt-3.5-turbo**: Fast, cost-effective for basic analysis
- **gpt-4**: More accurate, better reasoning for complex documents
- **gpt-4-turbo**: Good balance of speed and accuracy

### Azure OpenAI
Use your deployment names (e.g., "gpt-35-turbo", "gpt-4")

### Amazon Bedrock Models
- **Claude-3-Sonnet**: `anthropic.claude-3-sonnet-20240229-v1:0`
- **Claude-3-Haiku**: `anthropic.claude-3-haiku-20240307-v1:0`
- **Titan Text Express**: `amazon.titan-text-express-v1`
- **Llama-2-70B**: `meta.llama2-70b-chat-v1`

### Ollama Models
Common models: `llama2`, `codellama`, `mistral`, `phi`

## Usage Examples

### Example 1: Financial Document Analysis

```bash
# Configure for PII detection
curl -X POST \
  -F 'file=@financial_report.pdf' \
  -F 'rules=["Social Security Numbers", "Credit card numbers", "Bank account numbers", "Tax ID numbers"]' \
  http://localhost:8080/scan/llm
```

### Example 2: Code Repository Scan

```bash
# Look for secrets in code
curl -X POST \
  -F 'file=@config.yaml' \
  -F 'rules=["API keys and tokens", "Database passwords", "Private keys and certificates"]' \
  http://localhost:8080/scan/llm
```

### Example 3: Legal Document Review

```bash
# Hybrid analysis for comprehensive review
curl -X POST \
  -F 'file=@contract.pdf' \
  http://localhost:8080/scan/hybrid
```

## Performance Considerations

### Token Usage
- **Input**: Document length affects token consumption
- **Output**: Analysis depth affects response tokens
- **Cost**: Monitor token usage for cost control

### Timeout Settings
- Default: 30 seconds
- Large documents: Increase to 60-120 seconds
- Bedrock: May need longer timeouts

### Rate Limits
- **OpenAI**: Varies by tier (RPM/TPM limits)
- **Azure**: Based on deployment settings
- **Bedrock**: Region-specific limits
- **Ollama**: Local hardware limitations

## Security Considerations

### API Key Management
```bash
# Use environment variables
export LLM_API_KEY="sk-your-key-here"

# Or use secrets management
export LLM_API_KEY="$(vault kv get -field=api_key secret/openai)"
```

### AWS IAM Permissions
For Bedrock, ensure your IAM role has:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": "arn:aws:bedrock:*::foundation-model/*"
    }
  ]
}
```

### Data Privacy
- **OpenAI**: Data may be used for training (opt-out available)
- **Azure**: Enterprise privacy controls
- **Bedrock**: Data not used for training
- **Ollama**: Fully local processing

## Troubleshooting

### Common Issues

1. **"LLM service is not available"**
   - Check `LLM_ENABLED=true` environment variable
   - Verify configuration file exists and is valid

2. **Authentication errors**
   - Verify API keys are correct
   - Check AWS credentials and permissions

3. **Timeout errors**
   - Increase timeout in configuration
   - Try smaller documents or reduce max_tokens

4. **Model not found**
   - Verify model name/ID is correct for your provider
   - Check region availability for Bedrock models

### Debug Mode
Enable debug logging:
```bash
export DEBUG=true
```

## Migration Guide

### From Regex-Only to Hybrid

1. Start with hybrid mode (`/scan/hybrid`)
2. Compare regex vs LLM findings
3. Gradually replace regex rules with LLM analysis
4. Monitor accuracy and performance

### Provider Migration

OpenAI → Bedrock:
```yaml
# Change provider
llm:
  provider: "bedrock"  # was "openai"

# Update model configuration
bedrock:
  model_id: "anthropic.claude-3-sonnet-20240229-v1:0"
  region: "us-east-1"
```

## Cost Optimization

### Tips
1. **Use appropriate models**: GPT-3.5 for basic tasks, GPT-4 for complex
2. **Limit token usage**: Set reasonable max_tokens limits
3. **Batch processing**: Process multiple documents in batches
4. **Hybrid approach**: Use LLM only for validation, not initial scanning
5. **Local models**: Consider Ollama for high-volume use cases

### Monitoring
Track usage with logs:
```bash
grep "tokens_used" dws.log | awk '{sum+=$NF} END {print "Total tokens:", sum}'
```