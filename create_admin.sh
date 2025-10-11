#!/bin/bash

# Get database password from secret manager
DB_PASSWORD=$(gcloud secrets versions access latest --secret="db-password")
ADMIN_PASSWORD=$(gcloud secrets versions access latest --secret="admin-password")

# Create admin user via API
echo "Creating admin user..."

RESPONSE=$(curl -s -X POST https://civicweave-backend-162941711179.us-central1.run.app/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email":"admin@civicweave.com",
    "password":"'$ADMIN_PASSWORD'",
    "name":"System Administrator",
    "role":"admin",
    "consent_given":true
  }')

echo "Registration response: $RESPONSE"

# If user already exists, try to login with the secret password
echo "Testing login with secret password..."
LOGIN_RESPONSE=$(curl -s -X POST https://civicweave-backend-162941711179.us-central1.run.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@civicweave.com","password":"'$ADMIN_PASSWORD'"}')

echo "Login response: $LOGIN_RESPONSE"

# Also try with the hardcoded password from seed
echo "Testing login with seed password..."
LOGIN_RESPONSE2=$(curl -s -X POST https://civicweave-backend-162941711179.us-central1.run.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@civicweave.com","password":"admin123"}')

echo "Login with seed password response: $LOGIN_RESPONSE2"
