#!/bin/bash

# Kubernetes Deployment Script for Database Agent
# This script deploys the Database Agent to a Kubernetes cluster

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="db-agent"
REGISTRY=${REGISTRY:-"your-registry.com"}
IMAGE_TAG=${IMAGE_TAG:-"latest"}
DOMAIN=${DOMAIN:-"db-agent.yourdomain.com"}

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_kubectl() {
    log_info "Checking kubectl..."
    
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check if kubectl can connect to cluster
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
        exit 1
    fi
    
    log_success "kubectl is ready."
}

build_and_push_image() {
    log_info "Building and pushing Docker image..."
    
    # Build image
    docker build -f Dockerfile.agent -t ${REGISTRY}/db-agent:${IMAGE_TAG} .
    
    # Push image
    docker push ${REGISTRY}/db-agent:${IMAGE_TAG}
    
    log_success "Image built and pushed: ${REGISTRY}/db-agent:${IMAGE_TAG}"
}

update_image_references() {
    log_info "Updating image references in Kubernetes manifests..."
    
    # Update deployment.yaml with new image
    sed -i.bak "s|your-registry/db-agent:latest|${REGISTRY}/db-agent:${IMAGE_TAG}|g" k8s/deployment.yaml
    
    # Update ingress.yaml with domain
    sed -i.bak "s|db-agent.yourdomain.com|${DOMAIN}|g" k8s/ingress.yaml
    
    log_success "Image references updated."
}

create_namespace() {
    log_info "Creating namespace..."
    
    kubectl apply -f k8s/namespace.yaml
    
    log_success "Namespace created: ${NAMESPACE}"
}

create_secrets() {
    log_info "Creating secrets..."
    
    # Prompt for secrets if not provided via environment
    if [ -z "$METADB_PASSWORD" ]; then
        read -s -p "Enter metadata database password: " METADB_PASSWORD
        echo
    fi
    
    if [ -z "$API_KEY" ]; then
        read -s -p "Enter API key: " API_KEY
        echo
    fi
    
    # Update secret.yaml with actual values
    METADB_USER_B64=$(echo -n "db_agent_user" | base64)
    METADB_PASSWORD_B64=$(echo -n "$METADB_PASSWORD" | base64)
    API_KEY_B64=$(echo -n "$API_KEY" | base64)
    
    # Create temporary secret file
    cat > k8s/secret-temp.yaml << EOF
apiVersion: v1
kind: Secret
metadata:
  name: db-agent-secrets
  namespace: ${NAMESPACE}
  labels:
    app: db-agent
type: Opaque
data:
  METADB_USER: ${METADB_USER_B64}
  METADB_PASSWORD: ${METADB_PASSWORD_B64}
  API_KEY: ${API_KEY_B64}
EOF
    
    kubectl apply -f k8s/secret-temp.yaml
    rm k8s/secret-temp.yaml
    
    log_success "Secrets created."
}

setup_tls() {
    log_info "Setting up TLS..."
    
    if [ "$TLS_ENABLED" = "true" ]; then
        if [ -z "$TLS_CERT" ] || [ -z "$TLS_KEY" ]; then
            log_warning "TLS enabled but no certificate provided. Creating self-signed certificate..."
            
            # Generate self-signed certificate
            openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
                -keyout /tmp/tls.key \
                -out /tmp/tls.crt \
                -subj "/CN=${DOMAIN}"
            
            TLS_CERT=$(cat /tmp/tls.crt | base64 -w 0)
            TLS_KEY=$(cat /tmp/tls.key | base64 -w 0)
            
            rm /tmp/tls.crt /tmp/tls.key
        fi
        
        # Create TLS secret
        cat > k8s/tls-secret-temp.yaml << EOF
apiVersion: v1
kind: Secret
metadata:
  name: db-agent-tls
  namespace: ${NAMESPACE}
  labels:
    app: db-agent
type: kubernetes.io/tls
data:
  tls.crt: ${TLS_CERT}
  tls.key: ${TLS_KEY}
EOF
        
        kubectl apply -f k8s/tls-secret-temp.yaml
        rm k8s/tls-secret-temp.yaml
        
        log_success "TLS secret created."
    else
        log_info "TLS disabled. Using insecure connection."
    fi
}

deploy_config() {
    log_info "Deploying configuration..."
    
    kubectl apply -f k8s/configmap.yaml
    
    log_success "Configuration deployed."
}

deploy_application() {
    log_info "Deploying application..."
    
    kubectl apply -f k8s/deployment.yaml
    kubectl apply -f k8s/service.yaml
    kubectl apply -f k8s/hpa.yaml
    kubectl apply -f k8s/pdb.yaml
    
    log_success "Application deployed."
}

deploy_ingress() {
    log_info "Deploying ingress..."
    
    if [ "$INGRESS_ENABLED" = "true" ]; then
        kubectl apply -f k8s/ingress.yaml
        log_success "Ingress deployed."
    else
        log_info "Ingress disabled. Service will be accessible via ClusterIP or NodePort."
    fi
}

wait_for_deployment() {
    log_info "Waiting for deployment to be ready..."
    
    kubectl wait --for=condition=available --timeout=300s deployment/db-agent -n ${NAMESPACE}
    
    log_success "Deployment is ready."
}

show_status() {
    log_info "Deployment Status:"
    echo ""
    
    # Show pods
    kubectl get pods -n ${NAMESPACE}
    
    echo ""
    # Show services
    kubectl get services -n ${NAMESPACE}
    
    echo ""
    # Show ingress
    if [ "$INGRESS_ENABLED" = "true" ]; then
        kubectl get ingress -n ${NAMESPACE}
    fi
    
    echo ""
    log_info "Access Information:"
    
    if [ "$INGRESS_ENABLED" = "true" ]; then
        echo "  External gRPC: ${DOMAIN}:443"
        echo "  Internal gRPC: db-agent-service.${NAMESPACE}.svc.cluster.local:50051"
    else
        echo "  Internal gRPC: db-agent-service.${NAMESPACE}.svc.cluster.local:50051"
        echo "  Port forward: kubectl port-forward -n ${NAMESPACE} service/db-agent-service 50051:50051"
    fi
    
    echo ""
    log_info "Test the deployment:"
    echo "  kubectl run db-client --image=${REGISTRY}/db-agent:${IMAGE_TAG} --rm -it -- ./db-client -command=ping -agent=db-agent-service.${NAMESPACE}.svc.cluster.local:50051"
}

cleanup() {
    log_info "Cleaning up deployment..."
    
    # Delete resources in reverse order
    kubectl delete -f k8s/ingress.yaml --ignore-not-found=true
    kubectl delete -f k8s/pdb.yaml --ignore-not-found=true
    kubectl delete -f k8s/hpa.yaml --ignore-not-found=true
    kubectl delete -f k8s/service.yaml --ignore-not-found=true
    kubectl delete -f k8s/deployment.yaml --ignore-not-found=true
    kubectl delete -f k8s/configmap.yaml --ignore-not-found=true
    kubectl delete -f k8s/tls-secret.yaml --ignore-not-found=true
    kubectl delete -f k8s/secret.yaml --ignore-not-found=true
    kubectl delete -f k8s/namespace.yaml --ignore-not-found=true
    
    log_success "Cleanup completed."
}

show_help() {
    echo "Kubernetes Deployment Script for Database Agent"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  deploy     - Complete deployment (default)"
    echo "  update     - Update existing deployment"
    echo "  status     - Show deployment status"
    echo "  logs       - Show application logs"
    echo "  cleanup    - Remove all resources"
    echo "  help       - Show this help"
    echo ""
    echo "Environment Variables:"
    echo "  REGISTRY         - Docker registry URL (default: your-registry.com)"
    echo "  IMAGE_TAG        - Image tag (default: latest)"
    echo "  DOMAIN           - Domain for ingress (default: db-agent.yourdomain.com)"
    echo "  TLS_ENABLED      - Enable TLS (default: false)"
    echo "  INGRESS_ENABLED  - Enable ingress (default: true)"
    echo "  METADB_PASSWORD  - Metadata database password"
    echo "  API_KEY          - API key for authentication"
    echo "  TLS_CERT         - Base64 encoded TLS certificate"
    echo "  TLS_KEY          - Base64 encoded TLS private key"
    echo ""
    echo "Examples:"
    echo "  # Deploy with custom registry and tag"
    echo "  REGISTRY=my-registry.com IMAGE_TAG=v1.0.0 $0 deploy"
    echo ""
    echo "  # Deploy with TLS"
    echo "  TLS_ENABLED=true DOMAIN=db-agent.mydomain.com $0 deploy"
    echo ""
    echo "  # Deploy without ingress"
    echo "  INGRESS_ENABLED=false $0 deploy"
}

show_logs() {
    kubectl logs -f deployment/db-agent -n ${NAMESPACE}
}

# Main script logic
case "${1:-deploy}" in
    "deploy")
        check_kubectl
        build_and_push_image
        update_image_references
        create_namespace
        create_secrets
        setup_tls
        deploy_config
        deploy_application
        deploy_ingress
        wait_for_deployment
        show_status
        ;;
    "update")
        check_kubectl
        build_and_push_image
        update_image_references
        deploy_application
        wait_for_deployment
        show_status
        ;;
    "status")
        show_status
        ;;
    "logs")
        show_logs
        ;;
    "cleanup")
        cleanup
        ;;
    "help")
        show_help
        ;;
    *)
        log_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
