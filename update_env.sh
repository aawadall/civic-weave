#!/bin/bash

# Script to update .env file with current values from Google Secret Manager

echo "Updating .env file with current secrets from Google Secret Manager..."

cat > .env << EOF
# Database Configuration
DB_HOST=107.178.212.19
DB_PORT=5432
DB_NAME=civicweave
DB_USER=civicweave
DB_PASSWORD=$(gcloud secrets versions access latest --secret="db-password")

# Admin User Configuration (for seeding)
ADMIN_EMAIL=admin@civicweave.com
ADMIN_PASSWORD=$(gcloud secrets versions access latest --secret="admin-password")
ADMIN_NAME=System Administrator

# JWT Configuration
JWT_SECRET=$(gcloud secrets versions access latest --secret="jwt-secret")

# Mailgun Configuration
MAILGUN_API_KEY=$(gcloud secrets versions access latest --secret="mailgun-api-key")
MAILGUN_DOMAIN=$(gcloud secrets versions access latest --secret="mailgun-domain")

# Google OAuth Configuration
GOOGLE_CLIENT_ID=$(gcloud secrets versions access latest --secret="google-client-id")
GOOGLE_CLIENT_SECRET=$(gcloud secrets versions access latest --secret="google-client-secret")

# Geocoding Configuration
NOMINATIM_BASE_URL=https://nominatim.openstreetmap.org

# OpenAI Configuration
OPENAI_API_KEY=$(gcloud secrets versions access latest --secret="openai-api-key")
OPENAI_EMBEDDING_MODEL=text-embedding-3-small

# Redis Configuration
REDIS_HOST=10.63.105.59
REDIS_PORT=6379
REDIS_PASSWORD=
EOF

echo "✅ .env file updated successfully with current secrets!"
echo "🔐 Admin password is now: $(gcloud secrets versions access latest --secret="admin-password")"
