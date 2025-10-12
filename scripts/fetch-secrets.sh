#!/bin/bash

# CivicWeave - Fetch Secrets from Google Cloud Secret Manager
# This script pulls production secrets and populates a local .env file

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🔐 CivicWeave Secret Manager - Fetch Secrets"
echo "=============================================="
echo ""

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}❌ gcloud CLI is not installed. Please install it first:${NC}"
    echo "   https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Get current project
PROJECT_ID=$(gcloud config get-value project 2>/dev/null)

if [ -z "$PROJECT_ID" ]; then
    echo -e "${RED}❌ No GCP project configured. Set one with:${NC}"
    echo "   gcloud config set project YOUR_PROJECT_ID"
    exit 1
fi

echo -e "${GREEN}📍 Project: $PROJECT_ID${NC}"
echo ""

# Determine target directory
read -p "Create .env in which directory? (backend/infrastructure/terraform/root) [backend]: " TARGET_DIR
TARGET_DIR=${TARGET_DIR:-backend}

case $TARGET_DIR in
    backend)
        ENV_FILE="backend/.env"
        ;;
    infrastructure|terraform)
        ENV_FILE="infrastructure/terraform/.env"
        ;;
    root|.)
        ENV_FILE=".env"
        ;;
    *)
        ENV_FILE="$TARGET_DIR/.env"
        ;;
esac

echo -e "${YELLOW}📝 Target file: $ENV_FILE${NC}"
echo ""

# Confirm before proceeding
if [ -f "$ENV_FILE" ]; then
    echo -e "${YELLOW}⚠️  Warning: $ENV_FILE already exists and will be overwritten!${NC}"
    read -p "Continue? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Cancelled."
        exit 0
    fi
    # Backup existing file
    cp "$ENV_FILE" "$ENV_FILE.backup"
    echo -e "${GREEN}✅ Backup created: $ENV_FILE.backup${NC}"
fi

# Secret names in GCP
SECRETS=(
    "jwt-secret"
    "mailgun-api-key"
    "mailgun-domain"
    "google-client-id"
    "google-client-secret"
    "db-password"
    "admin-password"
    "openai-api-key"
)

# Start writing .env file
echo "# CivicWeave Environment Variables" > "$ENV_FILE"
echo "# Generated on $(date)" >> "$ENV_FILE"
echo "# Project: $PROJECT_ID" >> "$ENV_FILE"
echo "" >> "$ENV_FILE"

echo "🔄 Fetching secrets from Google Cloud Secret Manager..."
echo ""

# Fetch each secret
for secret in "${SECRETS[@]}"; do
    echo -n "  Fetching $secret... "
    
    # Try to fetch the secret
    if value=$(gcloud secrets versions access latest --secret="$secret" 2>/dev/null); then
        # Convert secret name to env var format (jwt-secret -> JWT_SECRET)
        env_var=$(echo "$secret" | tr '[:lower:]' '[:upper:]' | tr '-' '_')
        
        # Write to .env file
        echo "${env_var}=${value}" >> "$ENV_FILE"
        
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${YELLOW}⚠️  not found (skipping)${NC}"
    fi
done

# Add additional static environment variables for backend
if [[ "$ENV_FILE" == "backend/.env" ]]; then
    echo "" >> "$ENV_FILE"
    echo "# Database Configuration (Local Development)" >> "$ENV_FILE"
    echo "DB_HOST=localhost" >> "$ENV_FILE"
    echo "DB_PORT=5432" >> "$ENV_FILE"
    echo "DB_NAME=civicweave" >> "$ENV_FILE"
    echo "DB_USER=civicweave" >> "$ENV_FILE"
    echo "" >> "$ENV_FILE"
    echo "# Redis Configuration (Local Development)" >> "$ENV_FILE"
    echo "REDIS_HOST=localhost" >> "$ENV_FILE"
    echo "REDIS_PORT=6379" >> "$ENV_FILE"
    echo "REDIS_PASSWORD=" >> "$ENV_FILE"
    echo "" >> "$ENV_FILE"
    echo "# Other Configuration" >> "$ENV_FILE"
    echo "NOMINATIM_BASE_URL=https://nominatim.openstreetmap.org" >> "$ENV_FILE"
    echo "ADMIN_EMAIL=admin@civicweave.com" >> "$ENV_FILE"
    echo "ADMIN_NAME=System Administrator" >> "$ENV_FILE"
    echo "OPENAI_EMBEDDING_MODEL=text-embedding-3-small" >> "$ENV_FILE"
fi

# Add Terraform variable format if terraform directory
if [[ "$ENV_FILE" == "infrastructure/terraform/.env" ]]; then
    # Convert to TF_VAR_ format
    sed -i 's/^JWT_SECRET=/TF_VAR_jwt_secret=/' "$ENV_FILE"
    sed -i 's/^MAILGUN_API_KEY=/TF_VAR_mailgun_api_key=/' "$ENV_FILE"
    sed -i 's/^MAILGUN_DOMAIN=/TF_VAR_mailgun_domain=/' "$ENV_FILE"
    sed -i 's/^GOOGLE_CLIENT_ID=/TF_VAR_google_client_id=/' "$ENV_FILE"
    sed -i 's/^GOOGLE_CLIENT_SECRET=/TF_VAR_google_client_secret=/' "$ENV_FILE"
    sed -i 's/^DB_PASSWORD=/TF_VAR_db_password=/' "$ENV_FILE"
    sed -i 's/^ADMIN_PASSWORD=/TF_VAR_admin_password=/' "$ENV_FILE"
    sed -i 's/^OPENAI_API_KEY=/TF_VAR_openai_api_key=/' "$ENV_FILE"
    
    echo "" >> "$ENV_FILE"
    echo "# Additional Terraform Variables" >> "$ENV_FILE"
    echo "TF_VAR_project_id=$PROJECT_ID" >> "$ENV_FILE"
fi

echo ""
echo -e "${GREEN}✅ Secrets successfully written to $ENV_FILE${NC}"
echo ""
echo "📋 Next steps:"
echo "   1. Review $ENV_FILE to ensure all values are correct"
echo "   2. Add any missing environment variables"
echo "   3. Never commit this file to version control!"
echo ""
echo "🔒 Security reminder:"
echo "   - Keep .env files secure and never share them"
echo "   - Add .env to .gitignore (should already be there)"
echo "   - Use different secrets for dev/staging/prod"
echo ""

