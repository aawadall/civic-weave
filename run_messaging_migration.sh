#!/bin/bash
# Script to run messaging expansion migration on production Cloud SQL

echo "🚀 Running Messaging Expansion Migration on Production Cloud SQL..."
echo ""
echo "This will add universal messaging columns to project_messages table:"
echo "  ✓ recipient_user_id (for user-to-user messages)"
echo "  ✓ recipient_team_id (for user-to-team messages)" 
echo "  ✓ subject (message subject)"
echo "  ✓ message_scope (message scope type)"
echo "  ✓ task_id (link to tasks)"
echo ""

# Extract UP portion of migration 012
awk '/^-- UP/,/^-- DOWN/{if(!/^-- DOWN/)print}' backend/migrations/012_messaging_expansion_and_resources.sql > /tmp/messaging_migration.sql

echo "📁 Migration file prepared"
echo ""
echo "Choose your method:"
echo ""
echo "Option 1: Use gcloud CLI (requires psql installed)"
echo "  Run: gcloud sql connect civicweave-postgres --user=postgres --database=civicweave"
echo "  Then: \\i /tmp/messaging_migration.sql"
echo ""
echo "Option 2: Cloud Console Query Editor"
echo "  1. Go to: https://console.cloud.google.com/sql/instances/civicweave-postgres?project=civicweave-474622"
echo "  2. Click 'QUERY' button"
echo "  3. Copy and paste the SQL below:"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat /tmp/messaging_migration.sql
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "After running the migration, the messaging autocomplete will work!"
