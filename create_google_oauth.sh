#!/bin/bash

echo "🔐 Creating Google OAuth 2.0 Credentials for CivicWeave"
echo "Project: civicweave-474622"
echo

# Enable the Google+ API (required for OAuth)
echo "📋 Enabling Google+ API..."
gcloud services enable plus.googleapis.com

# Create OAuth consent screen (if not exists)
echo "📋 Setting up OAuth consent screen..."
gcloud auth application-default print-access-token > /dev/null 2>&1

echo "=== Manual Steps Required ==="
echo "1. Go to: https://console.cloud.google.com/apis/credentials?project=civicweave-474622"
echo "2. Click 'Create Credentials' → 'OAuth 2.0 Client ID'"
echo "3. Choose 'Web application'"
echo "4. Set these URIs:"
echo "   - Authorized JavaScript origins:"
echo "     * https://civicweave-frontend-162941711179.us-central1.run.app"
echo "     * https://civicweave-frontend-peedoie7va-uc.a.run.app"
echo "     * http://localhost:3000"
echo "     * http://localhost:5173"
echo "   - Authorized redirect URIs:"
echo "     * https://civicweave-backend-162941711179.us-central1.run.app/api/auth/google/callback"
echo "     * https://civicweave-backend-peedoie7va-uc.a.run.app/api/auth/google/callback"
echo "     * http://localhost:8080/api/auth/google/callback"
echo
echo "5. After creating, copy the Client ID and Client Secret"
echo "6. Run this script with the credentials:"
echo "   ./update_google_secrets.sh YOUR_CLIENT_ID YOUR_CLIENT_SECRET"
echo

# Create the update script
cat > update_google_secrets.sh << 'EOF'
#!/bin/bash

if [ $# -ne 2 ]; then
    echo "Usage: $0 <CLIENT_ID> <CLIENT_SECRET>"
    exit 1
fi

CLIENT_ID=$1
CLIENT_SECRET=$2

echo "🔐 Updating Google OAuth secrets..."

# Update Google Client ID
echo "$CLIENT_ID" | gcloud secrets versions add google-client-id --data-file=-
echo "✅ Updated google-client-id"

# Update Google Client Secret  
echo "$CLIENT_SECRET" | gcloud secrets versions add google-client-secret --data-file=-
echo "✅ Updated google-client-secret"

# Update local .env file
echo "📝 Updating local .env file..."
./update_env.sh

echo "🎉 Google OAuth secrets updated successfully!"
echo "🔑 Client ID: $CLIENT_ID"
echo "🔐 Client Secret: ${CLIENT_SECRET:0:10}..."
EOF

chmod +x update_google_secrets.sh
echo "✅ Created update_google_secrets.sh script"
