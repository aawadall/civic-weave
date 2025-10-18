#!/bin/bash

# Google Cloud Run Deployment Script for Database Agent
# This script deploys the Database Agent to Google Cloud Run

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID=${PROJECT_ID:-"your-project-id"}
REGION=${REGION:-"us-central1"}
SERVICE_NAME=${SERVICE_NAME:-"db-agent"}
IMAGE_NAME=${IMAGE_NAME:-"db-agent"}
IMAGE_TAG=${IMAGE_TAG:-"latest"}

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

check_gcloud() {
    log_info "Checking gcloud CLI..."
    
    if ! command -v gcloud &> /dev/null; then
        log_error "gcloud CLI is not installed. Please install it first:"
        echo "curl https://sdk.cloud.google.com | bash"
        exit 1
    fi
    
    # Check if authenticated
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q .; then
        log_error "Not authenticated with gcloud. Please run: gcloud auth login"
        exit 1
    fi
    
    # Set project
    gcloud config set project $PROJECT_ID
    
    log_success "gcloud CLI is ready."
}

enable_apis() {
    log_info "Enabling required APIs..."
    
    gcloud services enable cloudbuild.googleapis.com
    gcloud services enable run.googleapis.com
    gcloud services enable sqladmin.googleapis.com
    
    log_success "APIs enabled."
}

setup_cloud_sql() {
    log_info "Setting up Cloud SQL metadata database..."
    
    # Check if instance already exists
    if gcloud sql instances describe db-agent-metadata --quiet &>/dev/null; then
        log_info "Cloud SQL instance already exists."
    else
        log_info "Creating Cloud SQL instance..."
        gcloud sql instances create db-agent-metadata \
            --database-version=POSTGRES_15 \
            --tier=db-f1-micro \
            --region=$REGION \
            --root-password=secure_password \
            --storage-type=SSD \
            --storage-size=10GB \
            --storage-auto-increase \
            --backup-start-time=03:00
    fi
    
    # Create database
    log_info "Creating metadata database..."
    gcloud sql databases create db_agent_metadata \
        --instance=db-agent-metadata \
        --quiet || log_info "Database may already exist."
    
    # Create user
    log_info "Creating database user..."
    gcloud sql users create db_agent_user \
        --instance=db-agent-metadata \
        --password=secure_password \
        --quiet || log_info "User may already exist."
    
    log_success "Cloud SQL setup completed."
}

build_and_push_image() {
    log_info "Skipping build - using existing image: gcr.io/$PROJECT_ID/$IMAGE_NAME:latest"
    log_success "Image ready: gcr.io/$PROJECT_ID/$IMAGE_NAME:latest"
}

create_secrets() {
    log_info "Creating secrets..."
    
    # Create secret for database password
    echo -n "secure_password" | gcloud secrets create db-agent-db-password --data-file=- || \
        echo -n "secure_password" | gcloud secrets versions add db-agent-db-password --data-file=-
    
    # Generate and store API key
    API_KEY=$(openssl rand -hex 32)
    echo -n "$API_KEY" | gcloud secrets create db-agent-api-key --data-file=- || \
        echo -n "$API_KEY" | gcloud secrets versions add db-agent-api-key --data-file=-
    
    log_success "Secrets created."
    log_info "API Key: $API_KEY"
    log_warning "Save this API key - you'll need it for client authentication!"
}

deploy_to_cloud_run() {
    log_info "Deploying to Cloud Run..."
    
    # Get Cloud SQL connection name
    CONNECTION_NAME=$(gcloud sql instances describe db-agent-metadata --format="value(connectionName)")
    
    # Deploy to Cloud Run
    gcloud run deploy $SERVICE_NAME \
        --image gcr.io/$PROJECT_ID/$IMAGE_NAME:$IMAGE_TAG \
        --platform managed \
        --region $REGION \
        --allow-unauthenticated \
        --port 50051 \
        --memory 1Gi \
        --cpu 1 \
        --min-instances 0 \
        --max-instances 10 \
        --timeout 300 \
        --concurrency 100 \
        --add-cloudsql-instances $CONNECTION_NAME \
        --set-env-vars AGENT_HOST=0.0.0.0,AGENT_PORT=50051,METADB_HOST=/cloudsql/$CONNECTION_NAME,METADB_PORT=5432,METADB_NAME=db_agent_metadata,METADB_USER=db_agent_user,METADB_SSL_MODE=disable,ENABLE_AUTH=true,ENABLE_RATE_LIMIT=true,RATE_LIMIT_RPS=100,LOG_LEVEL=info \
        --set-secrets METADB_PASSWORD=db-agent-db-password:latest,API_KEY=db-agent-api-key:latest
    
    log_success "Deployed to Cloud Run."
}

setup_database_schema() {
    log_info "Setting up database schema..."
    
    # Get Cloud SQL connection name
    CONNECTION_NAME=$(gcloud sql instances describe db-agent-metadata --format="value(connectionName)")
    
    # Run schema setup
    gcloud sql connect db-agent-metadata \
        --user=postgres \
        --database=db_agent_metadata \
        --quiet << EOF
-- Create the metadata database schema
\i /dev/stdin
$(cat pkg/metadb/schema.sql)
EOF
    
    log_success "Database schema created."
}

get_service_url() {
    log_info "Getting service URL..."
    
    SERVICE_URL=$(gcloud run services describe $SERVICE_NAME --region=$REGION --format="value(status.url)")
    
    echo ""
    log_success "Deployment completed!"
    echo ""
    log_info "Service URL: $SERVICE_URL"
    echo ""
    log_info "Test the deployment:"
    echo "  go run cmd/db-client/main.go -command=ping -agent=$SERVICE_URL:443"
    echo ""
    log_info "Environment variables for client:"
    echo "  export DB_AGENT_URL=$SERVICE_URL:443"
    echo "  export DB_AGENT_API_KEY=your_api_key_here"
}

cleanup() {
    log_info "Cleaning up Cloud Run deployment..."
    
    # Delete Cloud Run service
    gcloud run services delete $SERVICE_NAME --region=$REGION --quiet || true
    
    # Delete Cloud SQL instance
    gcloud sql instances delete db-agent-metadata --quiet || true
    
    # Delete secrets
    gcloud secrets delete db-agent-db-password --quiet || true
    gcloud secrets delete db-agent-api-key --quiet || true
    
    # Delete images
    gcloud container images delete gcr.io/$PROJECT_ID/$IMAGE_NAME:$IMAGE_TAG --quiet || true
    
    log_success "Cleanup completed."
}

show_help() {
    echo "Google Cloud Run Deployment Script for Database Agent"
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
    echo "  PROJECT_ID    - Google Cloud project ID (required)"
    echo "  REGION        - Google Cloud region (default: us-central1)"
    echo "  SERVICE_NAME  - Cloud Run service name (default: db-agent)"
    echo "  IMAGE_NAME    - Container image name (default: db-agent)"
    echo "  IMAGE_TAG     - Container image tag (default: latest)"
    echo ""
    echo "Examples:"
    echo "  # Deploy with custom project"
    echo "  PROJECT_ID=my-project $0 deploy"
    echo ""
    echo "  # Deploy to different region"
    echo "  REGION=europe-west1 $0 deploy"
    echo ""
    echo "  # Update existing deployment"
    echo "  $0 update"
}

show_status() {
    if [ -z "$PROJECT_ID" ] || [ "$PROJECT_ID" = "your-project-id" ]; then
        log_error "Please set PROJECT_ID environment variable"
        echo "export PROJECT_ID=civicweave-474622"
        exit 1
    fi
    
    log_info "Deployment Status:"
    echo ""
    
    # Show Cloud Run service
    gcloud run services describe $SERVICE_NAME --region=$REGION
    
    echo ""
    # Show Cloud SQL instance
    gcloud sql instances describe db-agent-metadata
    
    echo ""
    # Show secrets
    gcloud secrets list --filter="name:db-agent"
}

show_logs() {
    gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=$SERVICE_NAME" --limit=50 --format="table(timestamp,severity,textPayload)"
}

# Main script logic
case "${1:-deploy}" in
    "deploy")
        if [ -z "$PROJECT_ID" ] || [ "$PROJECT_ID" = "your-project-id" ]; then
            log_error "Please set PROJECT_ID environment variable"
            echo "export PROJECT_ID=civicweave-474622"
            exit 1
        fi
        
        check_gcloud
        enable_apis
        setup_cloud_sql
        build_and_push_image
        create_secrets
        deploy_to_cloud_run
        setup_database_schema
        get_service_url
        ;;
    "update")
        check_gcloud
        build_and_push_image
        deploy_to_cloud_run
        get_service_url
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
