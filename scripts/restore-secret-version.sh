#!/bin/bash

# Script to restore a previous version of a GCP secret

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "🔄 GCP Secret Manager - Restore Secret Version"
echo "==============================================="
echo ""

# Get secret name
read -p "Enter secret name (e.g., mailgun-api-key): " SECRET_NAME

# Check if secret exists
if ! gcloud secrets describe "$SECRET_NAME" &>/dev/null; then
    echo -e "${RED}❌ Secret '$SECRET_NAME' not found${NC}"
    exit 1
fi

# List all versions
echo -e "${BLUE}📋 Available versions for '$SECRET_NAME':${NC}"
echo ""
gcloud secrets versions list "$SECRET_NAME" --format="table(name,state,createTime)"
echo ""

# Get version to restore
read -p "Enter version number to restore (e.g., 3): " VERSION_NUMBER

# Confirm the value
echo ""
echo -e "${YELLOW}⚠️  Preview of version $VERSION_NUMBER:${NC}"
VALUE=$(gcloud secrets versions access "$VERSION_NUMBER" --secret="$SECRET_NAME")
echo "First 50 chars: ${VALUE:0:50}..."
echo ""

read -p "Restore this version as the latest? (y/N): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# Create new version with the old value
echo ""
echo "🔄 Creating new version with value from version $VERSION_NUMBER..."
echo -n "$VALUE" | gcloud secrets versions add "$SECRET_NAME" --data-file=-

echo ""
echo -e "${GREEN}✅ Successfully restored version $VERSION_NUMBER as the latest version!${NC}"
echo ""
echo "📋 Updated version list:"
gcloud secrets versions list "$SECRET_NAME" --limit=5 --format="table(name,state,createTime)"
echo ""
echo "To verify: gcloud secrets versions access latest --secret=\"$SECRET_NAME\""

