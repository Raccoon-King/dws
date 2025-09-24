# Project Structure

## Clean Project Organization

The DWS project has been restructured for optimal organization, Iron Bank compliance, and production readiness.

```
dws/
├── .github/                      # GitHub workflows and CI/CD
│   └── workflows/
│       └── go.yml               # Go 1.24 CI pipeline
├── api/                         # HTTP API handlers and routes
│   ├── handlers.go
│   ├── handlers_test.go
│   └── s3_handlers.go
├── config/                      # Configuration files
│   ├── default.yaml            # Default rules
│   ├── llm.yaml               # LLM service configuration
│   └── smart_llm.yaml         # Smart LLM analysis rules
├── engine/                      # Core scanning engine
│   ├── engine.go              # Rule engine implementation
│   └── engine_test.go         # Comprehensive engine tests
├── helm/                        # Helm charts for Kubernetes deployment
│   └── dws/                    # Main Helm chart
│       ├── Chart.yaml          # Chart metadata
│       ├── values.yaml         # Default values
│       ├── values-dev.yaml     # Development overrides
│       ├── values-prod.yaml    # Production overrides
│       ├── README.md           # Chart documentation
│       ├── Makefile            # Chart management commands
│       ├── .helmignore         # Files to exclude
│       └── templates/          # Kubernetes resource templates
│           ├── _helpers.tpl    # Template helpers
│           ├── deployment.yaml # Application deployment
│           ├── service.yaml    # Service definition
│           ├── configmap.yaml  # Configuration management
│           ├── llm-configmap.yaml # LLM configuration
│           ├── secret.yaml     # Secret management
│           ├── serviceaccount.yaml # RBAC
│           ├── ingress.yaml    # Public ingress
│           ├── admin-ingress.yaml # Admin ingress
│           ├── hpa.yaml        # Autoscaling
│           ├── poddisruptionbudget.yaml # Availability
│           ├── networkpolicy.yaml # Security
│           ├── servicemonitor.yaml # Monitoring
│           ├── prometheusrule.yaml # Alerting
│           ├── tests/          # Helm tests
│           │   └── test-connection.yaml
│           └── NOTES.txt       # Post-install instructions
├── llm/                         # Large Language Model integration
│   ├── service.go              # LLM service interface
│   ├── service_test.go         # Service tests
│   ├── openai_provider.go      # OpenAI API provider
│   ├── bedrock_provider.go     # AWS Bedrock provider
│   ├── analyzer.go             # Document analysis
│   ├── smart_analyzer.go       # Cost-optimized analysis
│   └── smart_analyzer_test.go  # Smart analyzer tests
├── s3/                          # AWS S3 integration
│   ├── client.go               # S3 client implementation
│   └── client_test.go          # S3 client tests
├── scanner/                     # Document text extraction
│   ├── extract.go              # Text extraction logic
│   └── extract_test.go         # Extraction tests
├── scripts/                     # Iron Bank compliant scripts
│   └── entrypoint.sh          # Container entrypoint
├── testfiles/                   # Test data files
├── vendor/                      # Go module dependencies (Iron Bank requirement)
├── Dockerfile                   # Iron Bank compliant container build
├── hardening_manifest.yaml     # Iron Bank hardening specification
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── main.go                      # Application entry point
├── main_test.go                 # Main application tests
├── e2e_test.go                  # End-to-end tests
├── .gitignore                   # Git ignore rules
├── LICENSE                      # Apache 2.0 license
├── README.md                    # Project documentation
├── ENVIRONMENT_VARIABLES.md    # Environment variable reference
├── IRON_BANK_COMPLIANCE.md     # Iron Bank compliance report
├── LLM_INTEGRATION.md           # LLM integration guide
└── PROJECT_STRUCTURE.md        # This file
```

## Key Structural Improvements

### ✅ **Removed Obsolete Files/Directories:**
- `docker/` - Replaced by root Dockerfile and Helm charts
- `k8s/` - Replaced by comprehensive Helm charts
- `docs/` - Consolidated into root documentation files
- `tests/` - Tests moved inline with modules
- `web-bundles/` - Unused web assets
- `documentation/` - Iron Bank docs now in root
- Binary artifacts (`dws`, `dws-airgapped.exe`, etc.)
- Generated files (`coverage.html`, `coverage.out`)
- IDE-specific directories (`.bmad-core`, `.gemini`)

### ✅ **Organized by Functionality:**
- **Core Logic**: `main.go`, `engine/`, `api/`
- **Integrations**: `llm/`, `s3/`, `scanner/`
- **Deployment**: `helm/`, `scripts/`, `Dockerfile`
- **Configuration**: `config/`, `hardening_manifest.yaml`
- **Documentation**: Root-level `.md` files
- **Testing**: `*_test.go` files alongside source

### ✅ **Production Ready Structure:**
- Iron Bank compliant Dockerfile
- Comprehensive Helm charts with dev/prod variants
- Security hardening manifest
- Environment variable documentation
- CI/CD pipeline updated to Go 1.24
- Clean `.gitignore` for all artifacts

### ✅ **Modular Architecture:**
- Each major component in its own directory
- Clear separation of concerns
- Testable modules with comprehensive test coverage
- Configuration externalized

## Usage

### Development
```bash
# Run locally
go run main.go

# Run tests
go test ./...

# Deploy to dev environment
helm install dws-dev ./helm/dws -f ./helm/dws/values-dev.yaml
```

### Production
```bash
# Deploy to production
helm install dws-prod ./helm/dws -f ./helm/dws/values-prod.yaml

# Build Iron Bank compliant image
docker build -t dws:1.0.0 .
```

This clean structure supports scalable development, secure deployment, and maintainable operations.