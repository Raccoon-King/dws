# Environment Variables Reference

This document lists all environment variables used by the Document Scanner Service (DWS) and their Helm chart configuration.

## Core Application Variables

### Required Variables
| Variable | Default | Description | Helm Values Path |
|----------|---------|-------------|------------------|
| `PORT` | `8080` | HTTP server port | `app.port` |
| `RULES_FILE` | `/etc/dws/rules.yaml` | Path to rules configuration file | `app.rulesFile` |

### Optional Variables
| Variable | Default | Description | Helm Values Path |
|----------|---------|-------------|------------------|
| `DEBUG` | `false` | Enable debug logging | `app.debug` |
| `LOGGING` | `stdout` | Log output destination (`stdout`, `stderr`, `file`) | `app.logging` |

## LLM Service Variables

### LLM Configuration
| Variable | Default | Description | Helm Values Path |
|----------|---------|-------------|------------------|
| `LLM_ENABLED` | `false` | Enable/disable LLM functionality | `llm.enabled` |
| `LLM_CONFIG` | `/etc/dws/llm.yaml` | Path to LLM configuration file | `llm.configFile` |

### LLM API Keys (Sensitive - from Secrets)
| Variable | Default | Description | Helm Secret Key |
|----------|---------|-------------|-----------------|
| `LLM_API_KEY` | - | OpenAI/Azure API key | `LLM_API_KEY` |

## AWS Configuration Variables

### AWS Credentials (Sensitive - from Secrets or IAM)
| Variable | Default | Description | Helm Secret Key |
|----------|---------|-------------|-----------------|
| `AWS_ACCESS_KEY_ID` | - | AWS access key ID | `AWS_ACCESS_KEY_ID` |
| `AWS_SECRET_ACCESS_KEY` | - | AWS secret access key | `AWS_SECRET_ACCESS_KEY` |
| `AWS_SESSION_TOKEN` | - | AWS session token (temporary credentials) | `AWS_SESSION_TOKEN` |
| `AWS_ROLE_ARN` | - | AWS IAM role ARN for assume role | `AWS_ROLE_ARN` |

### AWS Configuration
| Variable | Default | Description | Helm Values Path |
|----------|---------|-------------|------------------|
| `AWS_REGION` | `us-east-1` | AWS region for S3/Bedrock | `aws.region` |

## Kubernetes/Helm Configuration

### Helm Values Mapping

#### Core Application (`app` section)
```yaml
app:
  port: 8080           # → PORT
  debug: false         # → DEBUG
  logging: "stdout"    # → LOGGING
  rulesFile: "/etc/dws/rules.yaml"  # → RULES_FILE
```

#### LLM Configuration (`llm` section)
```yaml
llm:
  enabled: false       # → LLM_ENABLED
  configFile: "/etc/dws/llm.yaml"  # → LLM_CONFIG
  provider: "openai"   # Used in llm.yaml ConfigMap
  timeout: "30s"       # Used in llm.yaml ConfigMap
  maxTokens: 1000      # Used in llm.yaml ConfigMap
  temperature: 0.7     # Used in llm.yaml ConfigMap

  openai:
    model: "gpt-3.5-turbo"  # Used in llm.yaml ConfigMap
    baseUrl: ""             # Used in llm.yaml ConfigMap
    orgId: ""               # Used in llm.yaml ConfigMap

  bedrock:
    region: "us-east-1"     # Used in llm.yaml ConfigMap
    modelId: "anthropic.claude-3-sonnet-20240229-v1:0"  # Used in llm.yaml ConfigMap
```

#### AWS Configuration (`aws` section)
```yaml
aws:
  region: "us-east-1"  # → AWS_REGION
  s3:
    timeout: "30s"     # Used in application logic
```

#### Secrets Configuration
```yaml
secrets:
  create: false        # Whether to create secret resource
  data:               # Secret data (base64 encoded)
    LLM_API_KEY: ""
    AWS_ACCESS_KEY_ID: ""
    AWS_SECRET_ACCESS_KEY: ""
    AWS_SESSION_TOKEN: ""
    AWS_ROLE_ARN: ""
```

## Deployment Examples

### Development Deployment
```bash
helm install dws-dev ./helm/dws -f values-dev.yaml \
  --set app.debug=true \
  --set llm.enabled=false
```

### Production with LLM Enabled
```bash
# Create secret first
kubectl create secret generic dws-secrets \
  --from-literal=LLM_API_KEY="sk-..." \
  --from-literal=AWS_ACCESS_KEY_ID="AKIA..." \
  --from-literal=AWS_SECRET_ACCESS_KEY="..."

# Install with LLM enabled
helm install dws-prod ./helm/dws -f values-prod.yaml \
  --set llm.enabled=true \
  --set secrets.create=false
```

### Using IAM Roles (Recommended for EKS)
```bash
# No AWS credential secrets needed when using IAM roles
helm install dws-prod ./helm/dws -f values-prod.yaml \
  --set llm.enabled=true \
  --set aws.region="us-west-2"
```

## Configuration File Templates

### LLM Configuration Template (`llm.yaml`)
The LLM configuration file is generated from Helm values:

```yaml
llm:
  enabled: {{ .Values.llm.enabled }}
  provider: "{{ .Values.llm.provider }}"
  timeout: "{{ .Values.llm.timeout }}"
  max_tokens: {{ .Values.llm.maxTokens }}
  temperature: {{ .Values.llm.temperature }}

openai:
  api_key: "${LLM_API_KEY}"  # Expanded from env var
  base_url: "{{ .Values.llm.openai.baseUrl }}"
  model: "{{ .Values.llm.openai.model }}"
  org_id: "{{ .Values.llm.openai.orgId }}"

bedrock:
  region: "{{ .Values.llm.bedrock.region }}"
  access_key_id: "${AWS_ACCESS_KEY_ID}"      # Expanded from env var
  secret_access_key: "${AWS_SECRET_ACCESS_KEY}"  # Expanded from env var
  session_token: "${AWS_SESSION_TOKEN}"      # Expanded from env var
  role_arn: "${AWS_ROLE_ARN}"               # Expanded from env var
  model_id: "{{ .Values.llm.bedrock.modelId }}"
```

## Security Best Practices

### Secret Management
1. **Never hardcode sensitive values** in Helm values files
2. **Use Kubernetes secrets** for API keys and credentials
3. **Prefer IAM roles** over static credentials in AWS environments
4. **Use external secret management** (Sealed Secrets, External Secrets Operator) in production

### Environment-Specific Configuration
- **Development**: Minimal configuration, LLM disabled by default
- **Staging**: LLM enabled with test credentials
- **Production**: Full LLM configuration with external secret management

### Container Security
- All sensitive environment variables are marked as `optional: true` in secret references
- Application gracefully handles missing environment variables
- Non-root execution enforced
- Read-only root filesystem