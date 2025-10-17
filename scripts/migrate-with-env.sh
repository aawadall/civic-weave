#!/bin/bash
# Simple Cloud Migration Script using Environment Variables
# Usage: Set environment variables and run ./migrate-with-env.sh

set -e

# Load environment variables
if [ -f .env.production ]; then
    source .env.production
fi

# Default values
SOURCE_HOST=${SOURCE_DB_HOST:-"localhost"}
SOURCE_USER=${SOURCE_DB_USER:-"postgres"}
SOURCE_DB=${SOURCE_DB_NAME:-"civicweave"}
SOURCE_PASSWORD=${SOURCE_DB_PASSWORD:-""}

TARGET_HOST=${NEW_DB_HOST:-""}
TARGET_USER=${NEW_DB_USER:-"postgres"}
TARGET_DB=${NEW_DB_NAME:-"civicweave"}
TARGET_PASSWORD=${NEW_DB_PASSWORD:-""}

# Check if target is specified
if [ -z "$TARGET_HOST" ]; then
    echo "❌ Error: NEW_DB_HOST environment variable is required"
    echo "Set NEW_DB_HOST, NEW_DB_USER, NEW_DB_NAME, NEW_DB_PASSWORD"
    exit 1
fi

echo "🚀 Starting database migration..."
echo "Source: $SOURCE_HOST/$SOURCE_DB"
echo "Target: $TARGET_HOST/$TARGET_DB"

# Create timestamp for files
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="backup_${TIMESTAMP}.sql"
EXPORT_FILE="civicweave_export_${TIMESTAMP}.sql"

# Step 1: Create backup
echo "📦 Creating backup..."
PGPASSWORD=$SOURCE_PASSWORD pg_dump -h $SOURCE_HOST -U $SOURCE_USER -d $SOURCE_DB \
    --no-owner --no-privileges --clean --if-exists \
    --file="$BACKUP_FILE"

# Step 2: Export database
echo "📤 Exporting database..."
PGPASSWORD=$SOURCE_PASSWORD pg_dump -h $SOURCE_HOST -U $SOURCE_USER -d $SOURCE_DB \
    --no-owner --no-privileges --clean --if-exists \
    --file="$EXPORT_FILE"

# Step 3: Import to new cloud
echo "📥 Importing to new cloud..."
PGPASSWORD=$TARGET_PASSWORD psql -h $TARGET_HOST -U $TARGET_USER -d $TARGET_DB \
    -f "$EXPORT_FILE"

# Step 4: Validate migration
echo "🔍 Validating migration..."
PGPASSWORD=$TARGET_PASSWORD psql -h $TARGET_HOST -U $TARGET_USER -d $TARGET_DB \
    -c "SELECT COUNT(*) as user_count FROM users;"
PGPASSWORD=$TARGET_PASSWORD psql -h $TARGET_HOST -U $TARGET_USER -d $TARGET_DB \
    -c "SELECT COUNT(*) as project_count FROM projects;"

echo "✅ Migration completed successfully!"
echo "Backup: $BACKUP_FILE"
echo "Export: $EXPORT_FILE"
