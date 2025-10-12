#!/bin/bash
# Script to run database migration on production Cloud SQL

echo "🚀 Running Migration on Production Cloud SQL..."
echo ""
echo "This will:"
echo "  1. Connect to civicweave-postgres (Cloud SQL)"
echo "  2. Run migration 006_project_enhancements.sql"
echo "  3. Create tables: project_tasks, task_updates, project_messages, message_reads"
echo "  4. Add columns to projects table"
echo ""

# Extract UP portion of migration
awk '/^-- UP/,/^-- DOWN/{if(!/^-- DOWN/)print}' backend/migrations/006_project_enhancements.sql > /tmp/prod_migration.sql

echo "📁 Migration file prepared"
echo ""
echo "Choose your method:"
echo ""
echo "Option 1: Use gcloud CLI (requires psql installed)"
echo "  Run: gcloud sql connect civicweave-postgres --user=postgres --database=civicweave"
echo "  Then: \\i /tmp/prod_migration.sql"
echo ""
echo "Option 2: Cloud Console Query Editor"
echo "  1. Go to: https://console.cloud.google.com/sql/instances/civicweave-postgres?project=civicweave-474622"
echo "  2. Click 'QUERY' button"
echo "  3. Copy and paste the SQL below:"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat /tmp/prod_migration.sql
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "After running the migration, the 500 errors will be fixed!"

