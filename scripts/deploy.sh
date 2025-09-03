#!/bin/bash

# DWS Kubernetes Deployment Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="dws"
IMAGE_NAME="dws"
IMAGE_TAG="latest"
KUBECONFIG_PATH=""

# Functions
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    print_info "Checking prerequisites..."
    
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v docker &> /dev/null; then
        print_error "docker is not installed or not in PATH"
        exit 1
    fi
    
    print_info "Prerequisites check passed"
}

build_image() {
    print_info "Building Docker image..."
    docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" .
    print_info "Docker image built successfully"
}

deploy_to_kubernetes() {
    print_info "Deploying to Kubernetes..."
    
    # Set kubeconfig if provided
    if [ -n "$KUBECONFIG_PATH" ]; then
        export KUBECONFIG="$KUBECONFIG_PATH"
        print_info "Using kubeconfig: $KUBECONFIG_PATH"
    fi
    
    # Check cluster connection
    if ! kubectl cluster-info &> /dev/null; then
        print_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    # Apply manifests
    print_info "Applying Kubernetes manifests..."
    kubectl apply -f k8s/namespace.yaml
    kubectl apply -f k8s/configmap.yaml
    kubectl apply -f k8s/deployment.yaml
    kubectl apply -f k8s/service.yaml
    kubectl apply -f k8s/hpa.yaml
    
    # Optionally apply ingress
    if kubectl get ingressclass nginx &> /dev/null; then
        print_info "NGINX Ingress Controller detected, applying ingress..."
        kubectl apply -f k8s/ingress.yaml
    else
        print_warn "NGINX Ingress Controller not found, skipping ingress deployment"
    fi
    
    print_info "Kubernetes deployment completed"
}

wait_for_deployment() {
    print_info "Waiting for deployment to be ready..."
    kubectl rollout status deployment/dws-deployment -n "$NAMESPACE" --timeout=300s
    print_info "Deployment is ready"
}

show_status() {
    print_info "Deployment Status:"
    echo
    kubectl get pods -n "$NAMESPACE" -l app=dws
    echo
    kubectl get services -n "$NAMESPACE" -l app=dws
    echo
    kubectl get ingress -n "$NAMESPACE" -l app=dws
    echo
    
    # Get external access information
    EXTERNAL_IP=$(kubectl get service dws-loadbalancer -n "$NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
    if [ -n "$EXTERNAL_IP" ]; then
        print_info "Application accessible at: http://$EXTERNAL_IP"
    fi
    
    INGRESS_HOST=$(kubectl get ingress dws-ingress -n "$NAMESPACE" -o jsonpath='{.spec.rules[0].host}' 2>/dev/null || echo "")
    if [ -n "$INGRESS_HOST" ]; then
        print_info "Application accessible via ingress at: https://$INGRESS_HOST"
    fi
}

cleanup() {
    print_info "Cleaning up resources..."
    kubectl delete -f k8s/ --ignore-not-found=true
    print_info "Cleanup completed"
}

# Main script
usage() {
    echo "Usage: $0 [OPTIONS] COMMAND"
    echo ""
    echo "Commands:"
    echo "  deploy     Build image and deploy to Kubernetes"
    echo "  build      Build Docker image only"
    echo "  k8s        Deploy to Kubernetes only (skip build)"
    echo "  status     Show deployment status"
    echo "  cleanup    Remove all resources"
    echo ""
    echo "Options:"
    echo "  -t TAG     Docker image tag (default: latest)"
    echo "  -k PATH    Path to kubeconfig file"
    echo "  -h         Show this help message"
}

# Parse command line arguments
while getopts "t:k:h" opt; do
    case $opt in
        t)
            IMAGE_TAG="$OPTARG"
            ;;
        k)
            KUBECONFIG_PATH="$OPTARG"
            ;;
        h)
            usage
            exit 0
            ;;
        \?)
            print_error "Invalid option: -$OPTARG"
            usage
            exit 1
            ;;
    esac
done

shift $((OPTIND-1))

COMMAND="$1"

case "$COMMAND" in
    deploy)
        check_prerequisites
        build_image
        deploy_to_kubernetes
        wait_for_deployment
        show_status
        ;;
    build)
        check_prerequisites
        build_image
        ;;
    k8s)
        check_prerequisites
        deploy_to_kubernetes
        wait_for_deployment
        show_status
        ;;
    status)
        show_status
        ;;
    cleanup)
        cleanup
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        usage
        exit 1
        ;;
esac

print_info "Script completed successfully"