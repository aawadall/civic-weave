#!/bin/bash

echo "🔐 Updating Google Secret Manager with real values..."
echo "⚠️  You'll need to provide the actual secret values"

# Function to update a secret
update_secret() {
    local secret_name=$1
    local secret_description=$2
    
    echo "Enter the real value for $secret_name ($secret_description):"
    read -s secret_value
    
    if [ -n "$secret_value" ]; then
        echo "$secret_value" | gcloud secrets versions add "$secret_name" --data-file=-
        echo "✅ Updated $secret_name"
    else
        echo "⏭️  Skipped $secret_name"
    fi
    echo
}

echo "=== Updating Secrets ==="
echo "Press Enter to skip any secret you don't want to update"
echo

update_secret "google-client-id" "Google OAuth Client ID"
update_secret "google-client-secret" "Google OAuth Client Secret"
update_secret "openai-api-key" "OpenAI API Key"
update_secret "mailgun-api-key" "Mailgun API Key"
update_secret "mailgun-domain" "Mailgun Domain"
update_secret "jwt-secret" "JWT Secret (generate a secure random string)"
update_secret "db-password" "Database Password"
update_secret "admin-password" "Admin Password"

echo "=== Updating local .env file ==="
./update_env.sh

echo "🎉 All secrets updated!"
echo "🔑 New admin password: $(gcloud secrets versions access latest --secret="admin-password")"
