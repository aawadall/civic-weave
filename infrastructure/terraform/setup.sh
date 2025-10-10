#!/bin/bash

# CivicWeave GCP Project Setup Script
# This script helps you create a new GCP project and set up the infrastructure

set -e

echo "🚀 CivicWeave GCP Project Setup"
echo "================================"

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    echo "❌ gcloud CLI is not installed. Please install it first:"
    echo "   https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Check if terraform is installed
if ! command -v terraform &> /dev/null; then
    echo "❌ Terraform is not installed. Please install it first:"
    echo "   https://terraform.io/downloads"
    exit 1
fi

# Get project details
read -p "Enter your GCP project ID (e.g., civicweave-2024): " PROJECT_ID
read -p "Enter your organization ID (optional, press Enter to skip): " ORG_ID

echo ""
echo "📋 Project Configuration:"
echo "   Project ID: $PROJECT_ID"
echo "   Organization ID: ${ORG_ID:-'None'}"
echo ""

read -p "Continue? (y/N): " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Setup cancelled."
    exit 1
fi

# Create new GCP project
echo "🏗️  Creating GCP project..."
if gcloud projects describe $PROJECT_ID &> /dev/null; then
    echo "   Project $PROJECT_ID already exists"
else
    gcloud projects create $PROJECT_ID \
        --name="CivicWeave" \
        ${ORG_ID:+--organization=$ORG_ID}
    echo "   ✅ Project created successfully"
fi

# Set project as default
echo "🔧 Setting project as default..."
gcloud config set project $PROJECT_ID

# Enable billing (user needs to do this manually)
echo ""
echo "💰 IMPORTANT: Enable billing for your project:"
echo "   https://console.cloud.google.com/billing/linkedaccount?project=$PROJECT_ID"
echo ""
read -p "Press Enter after enabling billing..."

# Enable required APIs
echo "🔌 Enabling required APIs..."
APIS=(
    "run.googleapis.com"
    "sqladmin.googleapis.com"
    "redis.googleapis.com"
    "secretmanager.googleapis.com"
    "oauth2.googleapis.com"
    "cloudbuild.googleapis.com"
    "artifactregistry.googleapis.com"
)

for api in "${APIS[@]}"; do
    echo "   Enabling $api..."
    if gcloud services enable $api --project=$PROJECT_ID 2>/dev/null; then
        echo "   ✅ $api enabled"
    else
        echo "   ⚠️  $api failed to enable (may require special permissions or already enabled)"
    fi
done

# Create Artifact Registry repository
echo "📦 Creating Artifact Registry repository..."
gcloud artifacts repositories create civicweave \
    --repository-format=docker \
    --location=us-central1 \
    --description="CivicWeave container repository" \
    --project=$PROJECT_ID || echo "   Repository already exists"

# Configure Docker authentication
echo "🐳 Configuring Docker authentication..."
gcloud auth configure-docker us-central1-docker.pkg.dev

# Create terraform.tfvars
echo "📝 Creating terraform.tfvars..."
cat > terraform.tfvars << EOF
# GCP Project Configuration
project_id   = "$PROJECT_ID"
project_name = "civicweave"
region       = "us-central1"
zone         = "us-central1-a"

# Security (CHANGE THESE VALUES!)
jwt_secret = "$(openssl rand -base64 32)"

# Email Service (UPDATE WITH YOUR VALUES)
mailgun_api_key = "your-mailgun-api-key"
mailgun_domain  = "your-mailgun-domain.com"

# Google OAuth (UPDATE WITH YOUR VALUES)
google_client_id     = "your-google-oauth-client-id"
google_client_secret = "your-google-oauth-client-secret"

# Database (CHANGE THIS PASSWORD!)
db_password = "$(openssl rand -base64 16)"
EOF

echo "   ✅ terraform.tfvars created"
echo ""
echo "⚠️  IMPORTANT: Update the following values in terraform.tfvars:"
echo "   - mailgun_api_key"
echo "   - mailgun_domain"
echo "   - google_client_id"
echo "   - google_client_secret"
echo ""

# Initialize Terraform
echo "🏗️  Initializing Terraform..."
terraform init

echo ""
echo "🎉 GCP Project Setup Complete!"
echo ""
echo "Next steps:"
echo "1. Update terraform.tfvars with your API keys"
echo "2. Run: terraform plan"
echo "3. Run: terraform apply"
echo "4. Build and deploy your applications"
echo ""
echo "For detailed deployment instructions, see:"
echo "   ../deployment/README.md"
