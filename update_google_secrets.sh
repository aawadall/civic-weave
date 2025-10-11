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
