# DWS Helm Chart

Document Scanner Service (DWS) - A Kubernetes-native document scanning service with rules engine and LLM integration.

## Features

- 🛡️ Iron Bank compliant container images
- 🔐 Security hardened with non-root execution
- 📈 Horizontal Pod Autoscaling support
- 🌐 Ingress configuration with TLS termination
- 📊 Prometheus monitoring integration
- 🔒 Network policies for security
- 🧪 Built-in Helm tests
- 🎛️ Flexible configuration management

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- Iron Bank registry access (for production images)

## Installation

### Quick Start

```bash
# Add helm repository (if published)
helm repo add dws https://charts.example.com
helm repo update

# Install with default values
helm install dws dws/dws -n dws-system --create-namespace
```

### Local Installation

```bash
# Clone repository
git clone https://github.com/example/dws
cd dws

# Install from local chart
helm install dws ./helm/dws -n dws-system --create-namespace
```

### Environment-Specific Deployments

#### Development Environment

```bash
helm install dws-dev ./helm/dws \
  -n dws-dev \
  --create-namespace \
  -f ./helm/dws/values-dev.yaml
```

#### Production Environment

```bash
helm install dws-prod ./helm/dws \
  -n dws-prod \
  --create-namespace \
  -f ./helm/dws/values-prod.yaml
```

## Configuration

The following table lists the configurable parameters and their default values.

### Global Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.imagePullSecrets` | Global image pull secrets | `[]` |
| `global.annotations` | Global annotations | `{}` |
| `global.labels` | Global labels | `{}` |

### Image Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.registry` | Container registry | `registry1.dso.mil` |
| `image.repository` | Container repository | `ironbank/opensource/dws` |
| `image.tag` | Container tag | `1.0.0` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |

### Application Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `app.name` | Application name | `dws` |
| `app.port` | Application port | `8080` |
| `app.debug` | Debug mode | `false` |
| `app.rulesFile` | Rules file path | `/etc/dws/rules.yaml` |

### Deployment Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `3` |
| `deployment.strategy` | Deployment strategy | `RollingUpdate` |
| `deployment.resources` | Resource limits and requests | See values.yaml |
| `deployment.nodeSelector` | Node selector | `{}` |
| `deployment.tolerations` | Tolerations | `[]` |
| `deployment.affinity` | Affinity rules | Anti-affinity enabled |

### Service Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | `80` |
| `service.targetPort` | Target port | `8080` |

### Ingress Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class | `nginx` |
| `ingress.hosts` | Ingress hosts | `[]` |
| `ingress.tls` | TLS configuration | `[]` |

### Autoscaling Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `autoscaling.enabled` | Enable HPA | `true` |
| `autoscaling.minReplicas` | Minimum replicas | `3` |
| `autoscaling.maxReplicas` | Maximum replicas | `20` |
| `autoscaling.targetCPUUtilizationPercentage` | CPU target | `70` |

### Monitoring Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `monitoring.serviceMonitor.enabled` | Enable ServiceMonitor | `false` |
| `monitoring.prometheusRule.enabled` | Enable PrometheusRule | `false` |

## Usage Examples

### Basic Deployment

```bash
helm install dws ./helm/dws \
  --set replicaCount=2 \
  --set app.debug=true
```

### Production Deployment with Ingress

```bash
helm install dws ./helm/dws \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=dws.example.com \
  --set ingress.hosts[0].paths[0].path=/ \
  --set ingress.hosts[0].paths[0].pathType=Prefix \
  --set ingress.tls[0].secretName=dws-tls \
  --set ingress.tls[0].hosts[0]=dws.example.com
```

### Custom Rules Configuration

```bash
# Create custom rules ConfigMap
kubectl create configmap dws-custom-rules \
  --from-file=rules.yaml=./my-rules.yaml

# Deploy with custom rules
helm install dws ./helm/dws \
  --set configMap.create=false \
  --set configMap.name=dws-custom-rules
```

### Enable Monitoring

```bash
helm install dws ./helm/dws \
  --set monitoring.serviceMonitor.enabled=true \
  --set monitoring.prometheusRule.enabled=true
```

## Testing

Run the built-in Helm tests:

```bash
helm test dws -n dws-system
```

## Upgrading

```bash
# Upgrade to new version
helm upgrade dws ./helm/dws -n dws-system

# Upgrade with new values
helm upgrade dws ./helm/dws -n dws-system -f new-values.yaml
```

## Uninstalling

```bash
helm uninstall dws -n dws-system
```

## Security

### Iron Bank Compliance

This chart uses Iron Bank hardened container images by default:
- Builder image: `registry1.dso.mil/ironbank/redhat/ubi/ubi8`
- Runtime image: `registry1.dso.mil/ironbank/google/distroless/static`

### Security Contexts

- Non-root user (UID: 65534)
- Read-only root filesystem
- All capabilities dropped
- Security profiles enforced

### Network Security

- Network policies enabled by default in production
- Ingress with rate limiting and security headers
- Admin endpoints with authentication

## Monitoring and Observability

### Prometheus Integration

The chart supports Prometheus monitoring through:
- ServiceMonitor for metrics scraping
- PrometheusRule for alerting
- Custom dashboards (via ConfigMap)

### Health Checks

- Liveness probe: `/health`
- Readiness probe: `/health`
- Startup probe: `/health`

## Troubleshooting

### Common Issues

1. **Pod CrashLoopBackOff**
   ```bash
   kubectl logs -l app.kubernetes.io/name=dws -n dws-system
   ```

2. **Ingress Not Working**
   ```bash
   kubectl describe ingress dws -n dws-system
   ```

3. **HPA Not Scaling**
   ```bash
   kubectl describe hpa dws -n dws-system
   ```

### Debug Mode

Enable debug mode for troubleshooting:
```bash
helm upgrade dws ./helm/dws --set app.debug=true
```

## Development

### Chart Structure

```
helm/dws/
├── Chart.yaml              # Chart metadata
├── values.yaml             # Default values
├── values-dev.yaml         # Development overrides
├── values-prod.yaml        # Production overrides
└── templates/
    ├── deployment.yaml     # Main application deployment
    ├── service.yaml        # Service definition
    ├── ingress.yaml        # Public ingress
    ├── admin-ingress.yaml  # Admin ingress
    ├── configmap.yaml      # Configuration
    ├── secret.yaml         # Secrets
    ├── serviceaccount.yaml # RBAC
    ├── hpa.yaml           # Horizontal Pod Autoscaler
    ├── poddisruptionbudget.yaml  # PDB
    ├── networkpolicy.yaml  # Network security
    ├── servicemonitor.yaml # Prometheus monitoring
    ├── prometheusrule.yaml # Alerting rules
    ├── tests/             # Helm tests
    ├── NOTES.txt          # Post-install notes
    └── _helpers.tpl       # Template helpers
```

### Testing Chart Changes

```bash
# Lint the chart
helm lint ./helm/dws

# Dry run installation
helm install dws ./helm/dws --dry-run --debug

# Template generation
helm template dws ./helm/dws > output.yaml
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test with `helm lint` and `helm template`
5. Submit a pull request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](../../LICENSE) file for details.