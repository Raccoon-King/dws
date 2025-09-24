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

# Final stage: Iron Bank approved distroless image
FROM registry1.dso.mil/ironbank/google/distroless/static

# Copy the binary
COPY --from=builder /build/dws /dws

# Copy default configuration files (if they exist)
COPY --from=builder /build/config /etc/dws/

# Distroless images run as non-root by default (nobody user)
# No need to set USER as distroless handles this

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/dws", "-health-check"] || exit 1

# Run binary directly (distroless doesn't have shell for scripts)
ENTRYPOINT ["/dws"]