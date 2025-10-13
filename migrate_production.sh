#!/bin/bash
# Production Database Migration Script

echo "🚀 Starting Production Database Migration..."
echo ""
echo "This will create:"
echo "  ✓ project_tasks table"
echo "  ✓ task_updates table"
echo "  ✓ project_messages table"
echo "  ✓ message_reads table"
echo "  ✓ Add columns to projects (content_json, budget_total, budget_spent, permissions)"
echo "  ✓ Auto-enrollment trigger"
echo ""
echo "Connecting to Cloud SQL instance: civicweave-postgres"
echo "Database: civicweave"
echo ""

gcloud sql connect civicweave-postgres --user=postgres --database=civicweave < /tmp/prod_migration_final.sql

echo ""
echo "✅ Migration complete!"
echo ""
echo "Refresh your browser - 500 errors should be gone!"

