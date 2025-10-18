#!/bin/bash

# Database Agent Setup Script
# This script sets up the Database Agent for cloud deployment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
AGENT_PORT=${AGENT_PORT:-50051}
METADB_PORT=${METADB_PORT:-5433}
REDIS_PORT=${REDIS_PORT:-6379}
PROMETHEUS_PORT=${PROMETHEUS_PORT:-9090}
GRAFANA_PORT=${GRAFANA_PORT:-3000}

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

check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go first."
        exit 1
    fi
    
    log_success "All dependencies are installed."
}

generate_keys() {
    log_info "Generating API keys..."
    
    # Create keys directory
    mkdir -p keys/server keys/client
    
    # Generate server keys
    log_info "Generating server keys..."
    go run cmd/db-keygen/main.go -server -description="Production Server Keys" -output=keys/server/
    
    # Generate client keys
    log_info "Generating client keys..."
    go run cmd/db-keygen/main.go -client -agent-url=localhost:${AGENT_PORT} -description="Development Client Keys" -output=keys/client/
    
    log_success "API keys generated successfully."
}

create_config() {
    log_info "Creating configuration files..."
    
    # Create config directory
    mkdir -p config
    
    # Create environment file
    cat > config/.env << EOF
# Agent Configuration
AGENT_HOST=0.0.0.0
AGENT_PORT=${AGENT_PORT}
AGENT_TLS=false

# Metadata Database Configuration
METADB_HOST=metadata-db
METADB_PORT=5432
METADB_USER=db_agent_user
METADB_PASSWORD=secure_password
METADB_NAME=db_agent_metadata
METADB_SSL_MODE=disable

# Security Configuration
ENABLE_AUTH=true
ENABLE_RATE_LIMIT=true
RATE_LIMIT_RPS=100
LOG_LEVEL=info

# Redis Configuration (optional)
REDIS_HOST=redis
REDIS_PORT=6379

# Monitoring Configuration
PROMETHEUS_PORT=${PROMETHEUS_PORT}
GRAFANA_PORT=${GRAFANA_PORT}
EOF
    
    log_success "Configuration files created."
}

setup_monitoring() {
    log_info "Setting up monitoring configuration..."
    
    # Create monitoring directory
    mkdir -p monitoring/grafana/dashboards monitoring/grafana/datasources
    
    # Create Prometheus configuration
    cat > monitoring/prometheus.yml << EOF
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'db-agent'
    static_configs:
      - targets: ['db-agent:50051']
    metrics_path: /metrics
    scrape_interval: 5s

  - job_name: 'postgres'
    static_configs:
      - targets: ['metadata-db:5432']
    scrape_interval: 15s

  - job_name: 'redis'
    static_configs:
      - targets: ['redis:6379']
    scrape_interval: 15s
EOF
    
    # Create Grafana datasource configuration
    cat > monitoring/grafana/datasources/prometheus.yml << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
EOF
    
    # Create Grafana dashboard configuration
    cat > monitoring/grafana/dashboards/db-agent.yml << EOF
apiVersion: 1

providers:
  - name: 'db-agent'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
EOF
    
    log_success "Monitoring configuration created."
}

build_images() {
    log_info "Building Docker images..."
    
    # Build agent image
    docker build -f Dockerfile.agent -t db-agent:latest .
    
    log_success "Docker images built successfully."
}

start_services() {
    log_info "Starting services..."
    
    # Start services with docker-compose
    docker-compose -f docker-compose.agent.yml up -d
    
    log_success "Services started successfully."
}

wait_for_services() {
    log_info "Waiting for services to be ready..."
    
    # Wait for metadata database
    log_info "Waiting for metadata database..."
    until docker-compose -f docker-compose.agent.yml exec -T metadata-db pg_isready -U db_agent_user -d db_agent_metadata; do
        sleep 2
    done
    
    # Wait for agent
    log_info "Waiting for agent..."
    until docker-compose -f docker-compose.agent.yml exec -T db-agent ./db-client -command=ping -agent=localhost:50051 -headless; do
        sleep 2
    done
    
    log_success "All services are ready."
}

show_status() {
    log_info "Service Status:"
    echo ""
    
    # Show running containers
    docker-compose -f docker-compose.agent.yml ps
    
    echo ""
    log_info "Service URLs:"
    echo "  Agent gRPC:     localhost:${AGENT_PORT}"
    echo "  Metadata DB:    localhost:${METADB_PORT}"
    echo "  Redis:          localhost:${REDIS_PORT}"
    echo "  Prometheus:     http://localhost:${PROMETHEUS_PORT}"
    echo "  Grafana:        http://localhost:${GRAFANA_PORT} (admin/admin)"
    
    echo ""
    log_info "Test the agent:"
    echo "  go run cmd/db-client/main.go -command=ping -agent=localhost:${AGENT_PORT}"
}

cleanup() {
    log_info "Cleaning up..."
    
    # Stop and remove containers
    docker-compose -f docker-compose.agent.yml down -v
    
    # Remove images
    docker rmi db-agent:latest 2>/dev/null || true
    
    # Remove volumes
    docker volume rm $(docker volume ls -q | grep db-agent) 2>/dev/null || true
    
    log_success "Cleanup completed."
}

show_help() {
    echo "Database Agent Setup Script"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  setup     - Complete setup (default)"
    echo "  start     - Start services"
    echo "  stop      - Stop services"
    echo "  restart   - Restart services"
    echo "  status    - Show service status"
    echo "  logs      - Show logs"
    echo "  cleanup   - Remove all containers and volumes"
    echo "  help      - Show this help"
    echo ""
    echo "Environment Variables:"
    echo "  AGENT_PORT       - Agent gRPC port (default: 50051)"
    echo "  METADB_PORT      - Metadata database port (default: 5433)"
    echo "  REDIS_PORT       - Redis port (default: 6379)"
    echo "  PROMETHEUS_PORT  - Prometheus port (default: 9090)"
    echo "  GRAFANA_PORT     - Grafana port (default: 3000)"
}

show_logs() {
    docker-compose -f docker-compose.agent.yml logs -f
}

# Main script logic
case "${1:-setup}" in
    "setup")
        check_dependencies
        generate_keys
        create_config
        setup_monitoring
        build_images
        start_services
        wait_for_services
        show_status
        ;;
    "start")
        start_services
        wait_for_services
        show_status
        ;;
    "stop")
        log_info "Stopping services..."
        docker-compose -f docker-compose.agent.yml stop
        log_success "Services stopped."
        ;;
    "restart")
        log_info "Restarting services..."
        docker-compose -f docker-compose.agent.yml restart
        wait_for_services
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
