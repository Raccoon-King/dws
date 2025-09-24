# Multi-stage build for Iron Bank compliance
# Using Iron Bank approved base image instead of golang:alpine
FROM registry1.dso.mil/ironbank/google/distroless/static AS runtime-base

# Builder stage - use approved UBI8 base for building
FROM registry1.dso.mil/ironbank/redhat/ubi/ubi8 AS builder

# Install required build tools
RUN dnf install -y go git ca-certificates \
    && dnf clean all \
    && rm -rf /var/cache/dnf

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Copy vendor directory (Iron Bank requires no internet downloads during build)
COPY vendor/ vendor/

# Copy source code
COPY . .

# Build the application with security flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -mod=vendor \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o dws .

# Run tests during build to ensure quality
RUN go test ./... -short

# Final stage: Iron Bank approved minimal image with shell for entrypoint script
FROM registry1.dso.mil/ironbank/redhat/ubi/ubi8-minimal

# Copy the binary
COPY --from=builder /build/dws /dws

# Copy default configuration files (if they exist)
COPY --from=builder /build/config /etc/dws/

# Copy entrypoint script from scripts folder
COPY --from=builder /build/scripts/entrypoint.sh /entrypoint.sh

# Make entrypoint executable and use existing nobody user (UID 65534)
RUN chmod +x /entrypoint.sh && \
    microdnf update -y && \
    microdnf clean all

# Use existing nobody user (UID 65534)
USER 65534

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/dws", "-health-check"] || exit 1

# Use entrypoint script for better startup handling
ENTRYPOINT ["/entrypoint.sh"]