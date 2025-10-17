#!/bin/bash

# Simple Database Deployment Script
# This script uses gcloud sql connect to deploy migrations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Simple Database Deployment${NC}"
echo "=================================="
echo "Project: civicweave-474622"
echo "Instance: civicweave-postgres"
echo ""

# Check if gcloud is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | head -n1 > /dev/null; then
    echo -e "${RED}Error: Not authenticated with gcloud${NC}"
    echo "Run: gcloud auth login"
    exit 1
fi

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo -e "${YELLOW}Installing PostgreSQL client...${NC}"
    # Try to install without sudo
    if command -v apt-get &> /dev/null; then
        echo "Please install PostgreSQL client:"
        echo "sudo apt-get update && sudo apt-get install -y postgresql-client"
        exit 1
    fi
fi

# Get the migration file
MIGRATION_FILE="backend/migrations/011_task_enhancements.sql"
if [ ! -f "$MIGRATION_FILE" ]; then
    echo -e "${RED}Error: Migration file not found: $MIGRATION_FILE${NC}"
    exit 1
fi

echo -e "${YELLOW}📝 Migration file found: $MIGRATION_FILE${NC}"

# Extract the UP section from the migration file
echo -e "${YELLOW}🔍 Extracting migration content...${NC}"
UP_SECTION=$(sed -n '/^-- UP$/,/^-- DOWN$/p' "$MIGRATION_FILE" | sed '1d;$d')

# Create a temporary SQL file
TEMP_SQL="/tmp/migration_011.sql"
echo "$UP_SECTION" > "$TEMP_SQL"

echo -e "${YELLOW}📋 Migration content:${NC}"
echo "----------------------------------------"
head -20 "$TEMP_SQL"
echo "..."
echo "----------------------------------------"

echo ""
echo -e "${YELLOW}🚀 Ready to deploy migration 011_task_enhancements${NC}"
echo ""
echo "This will:"
echo "  - Create task_comments table"
echo "  - Create task_time_logs table"
echo "  - Add new status values to project_tasks"
echo "  - Add task_id and message_type to project_messages"
echo "  - Create indexes and helper functions"
echo ""

read -p "Do you want to proceed? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Deployment cancelled${NC}"
    rm -f "$TEMP_SQL"
    exit 0
fi

echo -e "${YELLOW}🔧 Connecting to database...${NC}"

# Connect to the database and run the migration
if gcloud sql connect civicweave-postgres --user=civicweave --database=civicweave --quiet < "$TEMP_SQL"; then
    echo ""
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
    echo "Check the error messages above for details."
    exit 1
fi

# Cleanup
rm -f "$TEMP_SQL"

echo ""
echo -e "${GREEN}🎉 Database deployment complete!${NC}"
