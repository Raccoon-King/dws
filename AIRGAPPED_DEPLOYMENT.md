# Air-Gapped Deployment Guide

This guide explains how to deploy the Document Scanning Service (DWS) in air-gapped or isolated environments where internet access is restricted or unavailable.

## Prerequisites

### On a Networked System (Preparation Phase)

1. **Go 1.20+** installed
2. **Git** for cloning the repository
3. **Docker** (optional, for container builds)

### On the Air-Gapped System (Deployment Phase)

1. **Go 1.20+** (for binary builds) OR **Docker** (for container deployments)
2. **File transfer mechanism** to move artifacts from networked system

## Preparation (On Networked System)

### 1. Clone and Prepare Dependencies

```bash
# Clone the repository
git clone https://github.com/Raccoon-King/dws.git
cd dws

# Download and vendor all dependencies
go mod download
go mod vendor

# Verify all dependencies are vendored
go mod verify
```

### 2. Create Deployment Package

```bash
# Create deployment package
tar -czf dws-airgapped.tar.gz \
    --exclude='.git' \
    --exclude='*.md' \
    --exclude='scripts' \
    --exclude='k8s' \
    .

# Or create a minimal package with just essentials
mkdir -p dws-minimal
cp -r vendor/ dws-minimal/
cp go.mod go.sum *.go rules.yaml dws-minimal/
cp Dockerfile.airgapped build-airgapped.sh dws-minimal/
tar -czf dws-minimal.tar.gz dws-minimal/
```

### 3. Pre-build Docker Image (Optional)

```bash
# Build the air-gapped Docker image
docker build -f Dockerfile.airgapped -t dws:airgapped .

# Save image to file for transfer
docker save dws:airgapped -o dws-airgapped.tar
```

## Deployment (On Air-Gapped System)

### Method 1: Binary Deployment

1. **Transfer and Extract**
   ```bash
   # On air-gapped system
   tar -xzf dws-airgapped.tar.gz
   cd dws
   ```

2. **Build and Run**
   ```bash
   # Build using vendored dependencies
   ./build-airgapped.sh
   
   # Configure environment
   export RULES_FILE=rules.yaml
   export PORT=8080
   export DEBUG_MODE=false
   
   # Run the service
   ./dws
   ```

### Method 2: Docker Deployment

1. **Transfer and Load Image**
   ```bash
   # Load pre-built image
   docker load -i dws-airgapped.tar
   ```

2. **Run Container**
   ```bash
   # Create data directory
   mkdir -p ./data
   cp rules.yaml ./data/
   
   # Run container
   docker run -d \
     --name dws-service \
     -p 8080:8080 \
     -v ./data/rules.yaml:/app/rules.yaml:ro \
     -e RULES_FILE=/app/rules.yaml \
     dws:airgapped
   ```

### Method 3: Build on Air-Gapped System

1. **Transfer Source and Build**
   ```bash
   # Extract source
   tar -xzf dws-minimal.tar.gz
   cd dws-minimal
   
   # Build using vendored dependencies
   CGO_ENABLED=0 go build -mod=vendor -o dws .
   
   # Run
   export RULES_FILE=rules.yaml
   ./dws
   ```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `RULES_FILE` | `rules.yaml` | Path to rules configuration file |
| `PORT` | `8080` | HTTP server port |
| `DEBUG_MODE` | `false` | Enable debug logging |
| `LOGGING` | `stdout` | Log output destination |

### Rules Configuration

The `rules.yaml` file contains the scanning rules. You can customize it for your environment:

```yaml
rules:
  - id: custom-rule-1
    pattern: "sensitive-keyword"
    severity: high
    description: "Detects sensitive keywords"
```

## Health Monitoring

Since air-gapped environments may not have `wget` or `curl`, monitor the service using:

### 1. HTTP Health Check
```bash
# Use any HTTP client available
curl http://localhost:8080/health
# OR
wget -qO- http://localhost:8080/health
```

### 2. Process Monitoring
```bash
# Check if process is running
pgrep -f dws
ps aux | grep dws
```

### 3. Log Monitoring
```bash
# Monitor application logs
./dws 2>&1 | tee dws.log
```

## Security Considerations

### File Permissions
```bash
# Secure the binary
chmod 755 dws
chmod 644 rules.yaml

# Restrict access to rules file
chmod 600 rules.yaml  # If contains sensitive patterns
```

### Network Isolation
The service only requires:
- **Inbound**: HTTP connections on configured port (default 8080)
- **Outbound**: None (fully self-contained)

### Data Handling
- All processing is done locally
- No data leaves the air-gapped environment
- No external API calls or dependencies

## Troubleshooting

### Build Issues
```bash
# If vendor directory is missing
echo "ERROR: Run 'go mod vendor' on networked system first"

# If Go modules fail verification
go mod verify
```

### Runtime Issues
```bash
# Check if rules file exists
ls -la rules.yaml

# Verify rules file syntax
go run -c 'import "gopkg.in/yaml.v2"; yaml.Unmarshal(...)'

# Test service endpoints
curl -f http://localhost:8080/health
```

### Docker Issues
```bash
# Check container logs
docker logs dws-service

# Verify image
docker images | grep dws

# Check container status
docker ps -a | grep dws
```

## File Checksums

For security verification, generate checksums before transfer:

```bash
# On networked system
sha256sum dws-airgapped.tar.gz > dws-airgapped.sha256
sha256sum dws-airgapped.tar > dws-docker.sha256

# On air-gapped system
sha256sum -c dws-airgapped.sha256
sha256sum -c dws-docker.sha256
```

## Support

For air-gapped environments, ensure you have:
1. Complete source code with vendored dependencies
2. Build scripts and documentation
3. Checksum verification files
4. Emergency contact procedures for your organization

The service is designed to be completely self-contained once deployed.