#!/bin/bash

# GCP Database Deployment Script
# This script uses Cloud SQL proxy to deploy migrations to production

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID="civicweave-474622"
INSTANCE_NAME="civicweave-postgres"
CONNECTION_NAME="civicweave-474622:us-central1:civicweave-postgres"
DB_NAME="civicweave"
DB_USER="civicweave"
DB_PASSWORD="changeme-db-password"
LOCAL_PORT="5433"

echo -e "${BLUE}🚀 GCP Database Deployment Utility${NC}"
echo "=================================="
echo "Project: $PROJECT_ID"
echo "Instance: $INSTANCE_NAME"
echo "Connection: $CONNECTION_NAME"
echo ""

# Check if gcloud is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | head -n1 > /dev/null; then
    echo -e "${RED}Error: Not authenticated with gcloud${NC}"
    echo "Run: gcloud auth login"
    exit 1
fi

# Check if Cloud SQL proxy is available
if ! command -v cloud_sql_proxy &> /dev/null; then
    echo -e "${YELLOW}Installing Cloud SQL proxy...${NC}"
    wget https://dl.google.com/cloudsql/cloud_sql_proxy.linux.amd64 -O cloud_sql_proxy
    chmod +x cloud_sql_proxy
    sudo mv cloud_sql_proxy /usr/local/bin/
fi

# Start Cloud SQL proxy in background
echo -e "${YELLOW}🔧 Starting Cloud SQL proxy...${NC}"
cloud_sql_proxy -instances=$CONNECTION_NAME=tcp:$LOCAL_PORT &
PROXY_PID=$!

# Wait for proxy to start
sleep 3

# Function to cleanup
cleanup() {
    echo -e "${YELLOW}🧹 Cleaning up...${NC}"
    kill $PROXY_PID 2>/dev/null || true
    wait $PROXY_PID 2>/dev/null || true
}
trap cleanup EXIT

# Test connection
echo -e "${YELLOW}🔍 Testing database connection...${NC}"
if ! PGPASSWORD=$DB_PASSWORD psql -h localhost -p $LOCAL_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}Error: Failed to connect to database${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Connected to database successfully!${NC}"

# Check current migration status
echo -e "${YELLOW}📊 Checking migration status...${NC}"
MIGRATION_TABLE_EXISTS=$(PGPASSWORD=$DB_PASSWORD psql -h localhost -p $LOCAL_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations');" 2>/dev/null | xargs)

if [ "$MIGRATION_TABLE_EXISTS" = "f" ]; then
    echo -e "${YELLOW}📝 Creating migrations table...${NC}"
    PGPASSWORD=$DB_PASSWORD psql -h localhost -p $LOCAL_PORT -U $DB_USER -d $DB_NAME -c "
        CREATE TABLE schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            checksum VARCHAR(255)
        );
    "
fi

# Get applied migrations
echo -e "${YELLOW}📋 Checking applied migrations...${NC}"
APPLIED_MIGRATIONS=$(PGPASSWORD=$DB_PASSWORD psql -h localhost -p $LOCAL_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT version FROM schema_migrations ORDER BY version;" 2>/dev/null | xargs)

echo "Applied migrations: $APPLIED_MIGRATIONS"

# Check if 011_task_enhancements is already applied
if echo "$APPLIED_MIGRATIONS" | grep -q "011"; then
    echo -e "${GREEN}✅ Migration 011_task_enhancements is already applied!${NC}"
    exit 0
fi

# Run the migration
echo -e "${YELLOW}🚀 Running migration 011_task_enhancements...${NC}"

# Read and execute the migration
MIGRATION_FILE="backend/migrations/011_task_enhancements.sql"
if [ ! -f "$MIGRATION_FILE" ]; then
    echo -e "${RED}Error: Migration file not found: $MIGRATION_FILE${NC}"
    exit 1
fi

# Extract UP section from migration file
UP_SECTION=$(sed -n '/^-- UP$/,/^-- DOWN$/p' "$MIGRATION_FILE" | sed '1d;$d')

echo -e "${YELLOW}📝 Executing migration...${NC}"
if PGPASSWORD=$DB_PASSWORD psql -h localhost -p $LOCAL_PORT -U $DB_USER -d $DB_NAME -c "$UP_SECTION"; then
    # Record migration as applied
    PGPASSWORD=$DB_PASSWORD psql -h localhost -p $LOCAL_PORT -U $DB_USER -d $DB_NAME -c "
        INSERT INTO schema_migrations (version, checksum) 
        VALUES ('011', '$(wc -c < "$MIGRATION_FILE")')
        ON CONFLICT (version) DO NOTHING;
    "
    
    echo -e "${GREEN}✅ Migration 011_task_enhancements completed successfully!${NC}"
    echo ""
    echo -e "${BLUE}🎉 Task management enhancements deployed!${NC}"
    echo ""
    echo "New features available:"
    echo "  - Task comments and progress updates"
    echo "  - Volunteer time logging with automatic tallying"
    echo "  - Task status transitions (blocked, takeover requested)"
    echo "  - Automated Team Lead notifications"
    echo ""
    echo "Next steps:"
    echo "  1. Deploy the application: make build-push && make deploy-app"
    echo "  2. Test the new task management features"
else
    echo -e "${RED}❌ Migration failed!${NC}"
    exit 1
fi
