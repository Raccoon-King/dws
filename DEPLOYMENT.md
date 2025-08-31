# DWS Deployment Guide

This guide covers containerization and Kubernetes deployment for the Document Web Scanner (DWS) application.

## Table of Contents
- [Docker](#docker)
- [Kubernetes](#kubernetes)
- [Kubeconfig Setup](#kubeconfig-setup)
- [Deployment Commands](#deployment-commands)
- [Monitoring](#monitoring)

## Docker

### Building the Image

```bash
# Build the Docker image
docker build -t dws:latest .

# Build with specific tag
docker build -t dws:v1.0.0 .
```

### Running with Docker

```bash
# Run single container
docker run -d -p 8080:8080 --name dws-container dws:latest

# Run with custom rules file
docker run -d -p 8080:8080 -v ./rules.yaml:/root/rules.yaml dws:latest
```

### Docker Compose

For local development with optional nginx proxy:

```bash
# Start the application
docker-compose up -d

# Start with nginx proxy
docker-compose --profile proxy up -d

# View logs
docker-compose logs -f dws

# Stop services
docker-compose down
```

curl -X POST -F "file=@/path/to/your/document.pdf" http://localhost:8080/scan

## Kubernetes

### Prerequisites

- Kubernetes cluster (1.19+)
- kubectl configured
- NGINX Ingress Controller (optional)
- cert-manager for TLS (optional)

### Quick Deployment

```bash
# Apply all manifests
kubectl apply -f k8s/

# Or use kustomize
kubectl apply -k k8s/
```

### Step-by-Step Deployment

1. **Create Namespace**
   ```bash
   kubectl apply -f k8s/namespace.yaml
   ```

2. **Deploy ConfigMaps**
   ```bash
   kubectl apply -f k8s/configmap.yaml
   ```

3. **Deploy Application**
   ```bash
   kubectl apply -f k8s/deployment.yaml
   ```

4. **Create Services**
   ```bash
   kubectl apply -f k8s/service.yaml
   ```

5. **Setup Ingress (Optional)**
   ```bash
   # Update ingress.yaml with your domain
   kubectl apply -f k8s/ingress.yaml
   ```

6. **Enable Autoscaling**
   ```bash
   kubectl apply -f k8s/hpa.yaml
   ```

## Kubeconfig Setup

### Basic Kubeconfig Structure

```yaml
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: <BASE64_CA_CERT>
    server: https://your-cluster-api-server:6443
  name: dws-cluster
contexts:
- context:
    cluster: dws-cluster
    namespace: dws
    user: dws-user
  name: dws-context
current-context: dws-context
users:
- name: dws-user
  user:
    token: <YOUR_SERVICE_ACCOUNT_TOKEN>
```

### Service Account for DWS

```bash
# Create service account
kubectl create serviceaccount dws-sa -n dws

# Create cluster role
kubectl create clusterrole dws-role --verb=get,list,watch --resource=pods,services,deployments

# Bind role to service account
kubectl create clusterrolebinding dws-binding \
  --clusterrole=dws-role \
  --serviceaccount=dws:dws-sa

# Get service account token
kubectl get secret $(kubectl get serviceaccount dws-sa -n dws -o jsonpath='{.secrets[0].name}') -n dws -o jsonpath='{.data.token}' | base64 --decode
```

### Creating Kubeconfig for DWS

```bash
# Set cluster
kubectl config set-cluster dws-cluster \
  --server=https://your-cluster-api-server:6443 \
  --certificate-authority=/path/to/ca.crt \
  --embed-certs=true \
  --kubeconfig=dws-kubeconfig

# Set user credentials
kubectl config set-credentials dws-user \
  --token=<SERVICE_ACCOUNT_TOKEN> \
  --kubeconfig=dws-kubeconfig

# Set context
kubectl config set-context dws-context \
  --cluster=dws-cluster \
  --user=dws-user \
  --namespace=dws \
  --kubeconfig=dws-kubeconfig

# Use context
kubectl config use-context dws-context --kubeconfig=dws-kubeconfig
```

## Deployment Commands

### Using kubectl

```bash
# Deploy everything
kubectl apply -f k8s/

# Check deployment status
kubectl get pods -n dws
kubectl get services -n dws
kubectl get ingress -n dws

# Scale deployment
kubectl scale deployment dws-deployment --replicas=5 -n dws

# Update image
kubectl set image deployment/dws-deployment dws=dws:v1.1.0 -n dws

# Rollback deployment
kubectl rollout undo deployment/dws-deployment -n dws

# View logs
kubectl logs -f deployment/dws-deployment -n dws
```

### Using Kustomize

```bash
# Deploy with kustomize
kubectl apply -k k8s/

# Preview what will be applied
kubectl kustomize k8s/

# Deploy specific overlay (if you have overlays)
kubectl apply -k k8s/overlays/production/
```

### Helm Chart (Optional)

If you prefer Helm, create a basic chart structure:

```bash
# Create helm chart
helm create dws-chart

# Install with helm
helm install dws ./dws-chart -n dws --create-namespace

# Upgrade
helm upgrade dws ./dws-chart -n dws

# Uninstall
helm uninstall dws -n dws
```

## Monitoring

### Health Checks

```bash
# Check pod health
kubectl get pods -n dws

# Describe problematic pods
kubectl describe pod <pod-name> -n dws

# Check application logs
kubectl logs -f deployment/dws-deployment -n dws
```

### Metrics and Monitoring

```bash
# Check HPA status
kubectl get hpa -n dws

# View resource usage
kubectl top pods -n dws
kubectl top nodes
```

### Troubleshooting

```bash
# Port forward for local testing
kubectl port-forward service/dws-service 8080:8080 -n dws

# Execute commands in pod
kubectl exec -it deployment/dws-deployment -n dws -- /bin/sh

# Check service endpoints
kubectl get endpoints -n dws

# Describe services
kubectl describe service dws-service -n dws
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| RULES_FILE | rules.yaml | Path to rules configuration file |
| LOGGING | stdout | Logging destination (stdout/stderr/file) |
| LOG_LEVEL | info | Minimum log level (debug/info/warn/error) |
| LOG_FORMAT | text | Log output format (text/json) |
| MAX_UPLOAD_SIZE | 10485760 | Maximum upload size in bytes |

### ConfigMap Updates

```bash
# Update rules configuration
kubectl create configmap dws-rules --from-file=rules.yaml --dry-run=client -o yaml | kubectl apply -n dws -f -

# Restart deployment to pick up changes
kubectl rollout restart deployment/dws-deployment -n dws
```

### Secrets (if needed)

```bash
# Create TLS secret for ingress
kubectl create secret tls dws-tls-secret \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key \
  -n dws
```

## Production Considerations

1. **Security**
   - Use non-root user in containers
   - Enable Pod Security Standards
   - Network policies for traffic control
   - Regular security scanning

2. **Performance**
   - Configure resource requests/limits
   - Enable horizontal pod autoscaling
   - Use persistent volumes for logs if needed

3. **Reliability**
   - Configure liveness/readiness probes
   - Set up proper monitoring and alerting
   - Implement graceful shutdown

4. **Backup**
   - Backup configuration files
   - Document deployment procedures
   - Version control all manifests